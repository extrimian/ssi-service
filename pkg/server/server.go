// Package server contains the full set of handler functions and routes
// supported by the http api
package server

import (
	"fmt"
	"os"

	sdkutil "github.com/TBD54566975/ssi-sdk/util"
	"github.com/gin-gonic/gin"
	swaggerfiles "github.com/swaggo/files"
	ginswagger "github.com/swaggo/gin-swagger"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"

	"github.com/tbd54566975/ssi-service/config"
	"github.com/tbd54566975/ssi-service/pkg/server/framework"
	"github.com/tbd54566975/ssi-service/pkg/server/middleware"
	"github.com/tbd54566975/ssi-service/pkg/server/router"
	"github.com/tbd54566975/ssi-service/pkg/service"
	didsvc "github.com/tbd54566975/ssi-service/pkg/service/did"
	svcframework "github.com/tbd54566975/ssi-service/pkg/service/framework"
	"github.com/tbd54566975/ssi-service/pkg/service/webhook"
)

const (
	HealthPrefix            = "/health"
	ReadinessPrefix         = "/readiness"
	SwaggerPrefix           = "/swagger/*any"
	V1Prefix                = "/v1"
	OperationPrefix         = "/operations"
	DIDsPrefix              = "/dids"
	ResolverPrefix          = "/resolver"
	SchemasPrefix           = "/schemas"
	CredentialsPrefix       = "/credentials"
	StatusPrefix            = "/status"
	PresentationsPrefix     = "/presentations"
	DefinitionsPrefix       = "/definitions"
	SubmissionsPrefix       = "/submissions"
	IssuanceTemplatePrefix  = "/issuancetemplates"
	RequestsPrefix          = "/requests"
	ManifestsPrefix         = "/manifests"
	ApplicationsPrefix      = "/applications"
	ResponsesPrefix         = "/responses"
	KeyStorePrefix          = "/keys"
	VerificationPath        = "/verification"
	WebhookPrefix           = "/webhooks"
	DIDConfigurationsPrefix = "/did-configurations"

	batchSuffix = "/batch"
)

// SSIServer exposes all dependencies needed to run a http server and all its services
type SSIServer struct {
	*config.ServerConfig
	*service.SSIService
	*framework.Server
}

// NewSSIServer does two things: instantiates all service and registers their HTTP bindings
func NewSSIServer(shutdown chan os.Signal, cfg config.SSIServiceConfig) (*SSIServer, error) {
	// creates an HTTP server from the framework, and wrap it to extend it for the SSIS
	engine := setUpEngine(cfg.Server, shutdown)
	httpServer := framework.NewServer(cfg.Server, engine, shutdown)
	ssi, err := service.InstantiateSSIService(cfg.Services)
	if err != nil {
		return nil, sdkutil.LoggingErrorMsg(err, "unable to instantiate ssi service")
	}

	// make sure to set the api base in our service info
	config.SetAPIBase(cfg.Services.ServiceEndpoint)

	// service-level routers
	engine.GET(HealthPrefix, router.Health)
	engine.GET(ReadinessPrefix, router.Readiness(ssi.GetServices()))
	engine.StaticFile("swagger.yaml", "./doc/swagger.yaml")
	engine.GET(SwaggerPrefix, ginswagger.WrapHandler(swaggerfiles.Handler, ginswagger.URL("/swagger.yaml")))

	// register all v1 routers
	v1 := engine.Group(V1Prefix)
	if err = KeyStoreAPI(v1, ssi.KeyStore); err != nil {
		return nil, sdkutil.LoggingErrorMsg(err, "unable to instantiate KeyStore API")
	}
	if err = DecentralizedIdentityAPI(v1, ssi.DID, ssi.BatchDID, ssi.Webhook); err != nil {
		return nil, sdkutil.LoggingErrorMsg(err, "unable to instantiate DID API")
	}
	if err = SchemaAPI(v1, ssi.Schema, ssi.Webhook); err != nil {
		return nil, sdkutil.LoggingErrorMsg(err, "unable to instantiate Schema API")
	}
	if err = CredentialAPI(v1, ssi.Credential, ssi.Webhook, cfg.Services.StatusEndpoint); err != nil {
		return nil, sdkutil.LoggingErrorMsg(err, "unable to instantiate Credential API")
	}
	if err = OperationAPI(v1, ssi.Operation); err != nil {
		return nil, sdkutil.LoggingErrorMsg(err, "unable to instantiate Operation API")
	}
	if err = PresentationAPI(v1, ssi.Presentation, ssi.Webhook); err != nil {
		return nil, sdkutil.LoggingErrorMsg(err, "unable to instantiate Presentation API")
	}
	if err = ManifestAPI(v1, ssi.Manifest, ssi.Webhook); err != nil {
		return nil, sdkutil.LoggingErrorMsg(err, "unable to instantiate Manifest API")
	}
	if err = IssuanceAPI(v1, ssi.Issuance); err != nil {
		return nil, sdkutil.LoggingErrorMsg(err, "unable to instantiate Issuance API")
	}
	if err = WebhookAPI(v1, ssi.Webhook); err != nil {
		return nil, sdkutil.LoggingErrorMsg(err, "unable to instantiate Webhook API")
	}
	if err = DIDConfigurationAPI(v1, ssi.DIDConfiguration); err != nil {
		return nil, sdkutil.LoggingErrorMsg(err, "unable to instantiate DIDConfiguration API")
	}

	return &SSIServer{
		Server:       httpServer,
		SSIService:   ssi,
		ServerConfig: &cfg.Server,
	}, nil
}

// setUpEngine creates the gin engine and sets up the middleware based on config
func setUpEngine(cfg config.ServerConfig, shutdown chan os.Signal) *gin.Engine {
	gin.ForceConsoleColor()
	middlewares := gin.HandlersChain{
		gin.Recovery(),
		gin.Logger(),
		middleware.Errors(shutdown),
	}
	if cfg.JagerEnabled {
		middlewares = append(middlewares, otelgin.Middleware(config.ServiceName))
	}
	if cfg.EnableAllowAllCORS {
		middlewares = append(middlewares, middleware.CORS())
	}

	// set up engine and middleware
	engine := gin.New()
	engine.Use(middlewares...)
	switch cfg.Environment {
	case config.EnvironmentDev:
		gin.SetMode(gin.DebugMode)
	case config.EnvironmentTest:
		gin.SetMode(gin.TestMode)
	case config.EnvironmentProd:
		gin.SetMode(gin.ReleaseMode)
	}
	return engine
}

// KeyStoreAPI registers all HTTP handlers for the Key Store Service
func KeyStoreAPI(rg *gin.RouterGroup, service svcframework.Service) (err error) {
	keyStoreRouter, err := router.NewKeyStoreRouter(service)
	if err != nil {
		return sdkutil.LoggingErrorMsg(err, "creating key store router")
	}

	// make sure the keystore service is configured to use the correct path
	config.SetServicePath(svcframework.KeyStore, KeyStorePrefix)
	keyStoreAPI := rg.Group(KeyStorePrefix)
	keyStoreAPI.PUT("", keyStoreRouter.StoreKey)
	keyStoreAPI.GET("/:id", keyStoreRouter.GetKeyDetails)
	keyStoreAPI.DELETE("/:id", keyStoreRouter.RevokeKey)
	return
}

// DecentralizedIdentityAPI registers all HTTP handlers for the DID Service
func DecentralizedIdentityAPI(rg *gin.RouterGroup, service *didsvc.Service, did *didsvc.BatchService, webhookService *webhook.Service) (err error) {
	didRouter, err := router.NewDIDRouter(service)
	if err != nil {
		return sdkutil.LoggingErrorMsg(err, "creating DID router")
	}
	batchDIDRouter := router.NewBatchDIDRouter(did)

	// make sure the DID service is configured to use the correct path
	config.SetServicePath(svcframework.DID, DIDsPrefix)
	didAPI := rg.Group(DIDsPrefix)
	didAPI.GET("", didRouter.ListDIDMethods)
	didAPI.PUT("/:method", middleware.Webhook(webhookService, webhook.DID, webhook.Create), didRouter.CreateDIDByMethod)
	didAPI.PUT("/:method/:id", didRouter.UpdateDIDByMethod)
	didAPI.PUT("/:method/batch", middleware.Webhook(webhookService, webhook.DID, webhook.BatchCreate), batchDIDRouter.BatchCreateDIDs)
	didAPI.GET("/:method", didRouter.ListDIDsByMethod)
	didAPI.GET("/:method/:id", didRouter.GetDIDByMethod)
	didAPI.DELETE("/:method/:id", didRouter.SoftDeleteDIDByMethod)
	didAPI.GET(ResolverPrefix+"/:id", didRouter.ResolveDID)
	return
}

// SchemaAPI registers all HTTP handlers for the Schema Service
func SchemaAPI(rg *gin.RouterGroup, service svcframework.Service, webhookService *webhook.Service) (err error) {
	schemaRouter, err := router.NewSchemaRouter(service)
	if err != nil {
		return sdkutil.LoggingErrorMsg(err, "creating schema router")
	}

	// make sure the schema service is configured to use the correct path
	config.SetServicePath(svcframework.Schema, SchemasPrefix)
	schemaAPI := rg.Group(SchemasPrefix)
	schemaAPI.PUT("", middleware.Webhook(webhookService, webhook.Schema, webhook.Create), schemaRouter.CreateSchema)
	schemaAPI.GET("/:id", schemaRouter.GetSchema)
	schemaAPI.GET("", schemaRouter.ListSchemas)
	schemaAPI.DELETE("/:id", middleware.Webhook(webhookService, webhook.Schema, webhook.Delete), schemaRouter.DeleteSchema)
	return
}

// CredentialAPI registers all HTTP handlers for the Credentials Service
func CredentialAPI(rg *gin.RouterGroup, service svcframework.Service, webhookService *webhook.Service, statusEndpoint string) (err error) {
	credRouter, err := router.NewCredentialRouter(service)
	if err != nil {
		return sdkutil.LoggingErrorMsg(err, "creating credential router")
	}

	// make sure the credential service is configured to use the correct path
	config.SetServicePath(svcframework.Credential, CredentialsPrefix)

	// allows for a custom URI to be used for status list credentials, if not set, we use the default path
	if statusEndpoint != "" {
		config.SetStatusBase(statusEndpoint)
	} else {
		config.SetStatusBase(fmt.Sprintf("%s/status", config.GetServicePath(svcframework.Credential)))
	}

	// Credentials
	credentialAPI := rg.Group(CredentialsPrefix)
	credentialAPI.PUT("", middleware.Webhook(webhookService, webhook.Credential, webhook.Create), credRouter.CreateCredential)
	credentialAPI.PUT(batchSuffix, middleware.Webhook(webhookService, webhook.Credential, webhook.BatchCreate), credRouter.BatchCreateCredentials)
	credentialAPI.GET("", credRouter.ListCredentials)
	credentialAPI.GET("/:id", credRouter.GetCredential)
	credentialAPI.PUT(VerificationPath, credRouter.VerifyCredential)
	credentialAPI.DELETE("/:id", middleware.Webhook(webhookService, webhook.Credential, webhook.Delete), credRouter.DeleteCredential)

	// Credential Status
	credentialAPI.GET("/:id"+StatusPrefix, credRouter.GetCredentialStatus)
	credentialAPI.PUT("/:id"+StatusPrefix, credRouter.UpdateCredentialStatus)
	credentialAPI.PUT(StatusPrefix+batchSuffix, credRouter.BatchUpdateCredentialStatus)
	credentialAPI.GET(StatusPrefix+"/:id", credRouter.GetCredentialStatusList)
	return
}

// PresentationAPI registers all HTTP handlers for the Presentation Service
func PresentationAPI(rg *gin.RouterGroup, service svcframework.Service, webhookService *webhook.Service) (err error) {
	presRouter, err := router.NewPresentationRouter(service)
	if err != nil {
		return sdkutil.LoggingErrorMsg(err, "creating credential router")
	}

	// make sure the presentation service is configured to use the correct path
	config.SetServicePath(svcframework.Presentation, PresentationsPrefix)

	presAPI := rg.Group(PresentationsPrefix)
	presAPI.PUT(VerificationPath, presRouter.VerifyPresentation)

	presDefAPI := rg.Group(PresentationsPrefix + DefinitionsPrefix)
	presDefAPI.PUT("", presRouter.CreateDefinition)
	presDefAPI.GET("/:id", presRouter.GetDefinition)
	presDefAPI.GET("", presRouter.ListDefinitions)
	presDefAPI.DELETE("/:id", presRouter.DeleteDefinition)

	presReqAPI := rg.Group(PresentationsPrefix + RequestsPrefix)
	presReqAPI.PUT("", presRouter.CreateRequest)
	presReqAPI.GET("/:id", presRouter.GetRequest)
	presReqAPI.GET("", presRouter.ListRequests)
	presReqAPI.PUT("/:id", presRouter.DeleteRequest)

	presSubAPI := rg.Group(PresentationsPrefix + SubmissionsPrefix)
	presSubAPI.PUT("", middleware.Webhook(webhookService, webhook.Submission, webhook.Create), presRouter.CreateSubmission)
	presSubAPI.GET("/:id", presRouter.GetSubmission)
	presSubAPI.GET("", presRouter.ListSubmissions)
	presSubAPI.PUT("/:id/review", presRouter.ReviewSubmission)
	return
}

// OperationAPI registers all HTTP handlers for the Operations Service
func OperationAPI(rg *gin.RouterGroup, service svcframework.Service) (err error) {
	operationRouter, err := router.NewOperationRouter(service)
	if err != nil {
		return sdkutil.LoggingErrorMsg(err, "creating operation router")
	}

	// make sure the operation service is configured to use the correct path
	config.SetServicePath(svcframework.Operation, OperationPrefix)

	operationAPI := rg.Group(OperationPrefix)
	operationAPI.GET("", operationRouter.ListOperations)
	// In this case, it's used so that the operation id matches `presentations/submissions/{submission_id}` for the DIDWebID
	// path	`/v1/operations/cancel/presentations/submissions/{id}`
	operationAPI.PUT("/cancel/*id", operationRouter.CancelOperation)
	operationAPI.GET("/*id", operationRouter.GetOperation)
	return
}

// ManifestAPI registers all HTTP handlers for the Manifest Service
func ManifestAPI(rg *gin.RouterGroup, service svcframework.Service, webhookService *webhook.Service) (err error) {
	manifestRouter, err := router.NewManifestRouter(service)
	if err != nil {
		return sdkutil.LoggingErrorMsg(err, "creating manifest router")
	}

	// make sure the manifest service is configured to use the correct path
	config.SetServicePath(svcframework.Manifest, ManifestsPrefix)

	manifestAPI := rg.Group(ManifestsPrefix)
	manifestAPI.PUT("", middleware.Webhook(webhookService, webhook.Manifest, webhook.Create), manifestRouter.CreateManifest)
	manifestAPI.GET("", manifestRouter.ListManifests)
	manifestAPI.GET("/:id", manifestRouter.GetManifest)
	manifestAPI.DELETE("/:id", middleware.Webhook(webhookService, webhook.Manifest, webhook.Delete), manifestRouter.DeleteManifest)

	applicationAPI := manifestAPI.Group(ApplicationsPrefix)
	applicationAPI.PUT("", middleware.Webhook(webhookService, webhook.Application, webhook.Create), manifestRouter.SubmitApplication)
	applicationAPI.GET("", manifestRouter.ListApplications)
	applicationAPI.GET("/:id", manifestRouter.GetApplication)
	applicationAPI.DELETE("/:id", middleware.Webhook(webhookService, webhook.Application, webhook.Delete), manifestRouter.DeleteApplication)
	applicationAPI.PUT("/:id/review", manifestRouter.ReviewApplication)

	manifestReqAPI := manifestAPI.Group(RequestsPrefix)
	manifestReqAPI.PUT("", manifestRouter.CreateRequest)
	manifestReqAPI.GET("", manifestRouter.ListRequests)
	manifestReqAPI.GET("/:id", manifestRouter.GetRequest)
	manifestReqAPI.PUT("/:id", manifestRouter.DeleteRequest)

	responseAPI := manifestAPI.Group(ResponsesPrefix)
	responseAPI.GET("", manifestRouter.ListResponses)
	responseAPI.GET("/:id", manifestRouter.GetResponse)
	responseAPI.DELETE("/:id", manifestRouter.DeleteResponse)
	return
}

// IssuanceAPI registers all HTTP handlers for the Issuance Service
func IssuanceAPI(rg *gin.RouterGroup, service svcframework.Service) error {
	issuanceRouter, err := router.NewIssuanceRouter(service)
	if err != nil {
		return sdkutil.LoggingErrorMsg(err, "creating issuing router")
	}

	// make sure the issuance service is configured to use the correct path
	config.SetServicePath(svcframework.Issuance, IssuanceTemplatePrefix)

	issuanceAPI := rg.Group(IssuanceTemplatePrefix)
	issuanceAPI.PUT("", issuanceRouter.CreateIssuanceTemplate)
	issuanceAPI.GET("", issuanceRouter.ListIssuanceTemplates)
	issuanceAPI.GET("/:id", issuanceRouter.GetIssuanceTemplate)
	issuanceAPI.DELETE("/:id", issuanceRouter.DeleteIssuanceTemplate)
	return nil
}

// WebhookAPI registers all HTTP handlers for the Webhook Service
func WebhookAPI(rg *gin.RouterGroup, service svcframework.Service) (err error) {
	webhookRouter, err := router.NewWebhookRouter(service)
	if err != nil {
		return sdkutil.LoggingErrorMsg(err, "creating webhook router")
	}

	// make sure the webhook service is configured to use the correct path
	config.SetServicePath(svcframework.Webhook, WebhookPrefix)

	webhookAPI := rg.Group(WebhookPrefix)
	webhookAPI.PUT("", webhookRouter.CreateWebhook)
	webhookAPI.GET("", webhookRouter.ListWebhooks)
	webhookAPI.GET("/:noun/:verb", webhookRouter.GetWebhook)
	webhookAPI.DELETE("/:noun/:verb", webhookRouter.DeleteWebhook)

	// TODO(gabe): consider refactoring this to a single get on /webhooks/info or similar
	webhookAPI.GET("nouns", webhookRouter.GetSupportedNouns)
	webhookAPI.GET("verbs", webhookRouter.GetSupportedVerbs)
	return
}

func DIDConfigurationAPI(rg *gin.RouterGroup, service svcframework.Service) error {
	didConfigurationsRouter, err := router.NewDIDConfigurationsRouter(service)
	if err != nil {
		return sdkutil.LoggingErrorMsg(err, "creating webhook router")
	}

	// make sure the did configuration service is configured to use the correct path
	config.SetServicePath(svcframework.DIDConfiguration, DIDConfigurationsPrefix)

	webhookAPI := rg.Group(DIDConfigurationsPrefix)
	webhookAPI.PUT("", didConfigurationsRouter.CreateDIDConfiguration)
	webhookAPI.PUT(VerificationPath, didConfigurationsRouter.VerifyDIDConfiguration)

	return nil
}
