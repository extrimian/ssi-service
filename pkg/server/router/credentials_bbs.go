package router

import (
	"fmt"
	"net/http"

	"github.com/extrimian/ssi-service/pkg/server/framework"
	credentials_bbs "github.com/extrimian/ssi-service/pkg/service/credentials-bbs"
	svcframework "github.com/extrimian/ssi-service/pkg/service/framework"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

type CredentialsBBSRouter struct {
	service *credentials_bbs.Service
}

func NewCredentialsBBSRouter(s svcframework.Service) (*CredentialsBBSRouter, error) {
	if s == nil {
		return nil, errors.New("service cannot be nil")
	}
	credentialsBBSService, ok := s.(*credentials_bbs.Service)
	if !ok {
		return nil, fmt.Errorf("could not create DID router with service type: %s", s.Type())
	}
	return &CredentialsBBSRouter{service: credentialsBBSService}, nil
}

func (cbr CredentialsBBSRouter) PackDIDComm(c *gin.Context) {
	var packDIDCommRequest credentials_bbs.PackDIDCommRequest
	if err := framework.Decode(c.Request, &packDIDCommRequest); err != nil {
		framework.LoggingRespondErrMsg(c, err.Error(), http.StatusBadRequest)
		return
	}

	packDIDCommResponse, err := cbr.service.PackDIDComm(packDIDCommRequest)
	if err != nil {
		framework.LoggingRespondErrMsg(c, err.Error(), http.StatusNotFound)
		return
	}
	framework.Respond(c, packDIDCommResponse, http.StatusOK)
}

func (cbr CredentialsBBSRouter) CreateCredential(c *gin.Context) {
	var vcRequest credentials_bbs.CredentialRequest
	if err := framework.Decode(c.Request, &vcRequest); err != nil {
		framework.LoggingRespondErrMsg(c, err.Error(), http.StatusBadRequest)
		return
	}

	vcResponse, err := cbr.service.CreateCredential(vcRequest)
	if err != nil {
		framework.LoggingRespondErrMsg(c, err.Error(), http.StatusNotFound)
		return
	}
	framework.Respond(c, vcResponse, http.StatusOK)
}

func (cbr CredentialsBBSRouter) VerifyCredential(c *gin.Context) {
	var vcRequest credentials_bbs.SignedCredential
	if err := framework.Decode(c.Request, &vcRequest); err != nil {
		framework.LoggingRespondErrMsg(c, err.Error(), http.StatusBadRequest)
		return
	}

	vcResponse, err := cbr.service.VerifyCredential(vcRequest)
	if err != nil {
		framework.LoggingRespondErrMsg(c, err.Error(), http.StatusNotFound)
		return
	}
	framework.Respond(c, vcResponse, http.StatusOK)
}

func (cbr CredentialsBBSRouter) CreateOOBFromVC(c *gin.Context) {
	var oobRequest credentials_bbs.OOBRequest
	if err := framework.Decode(c.Request, &oobRequest); err != nil {
		framework.LoggingRespondErrMsg(c, err.Error(), http.StatusBadRequest)
		return
	}

	oobResponse, err := cbr.service.CreateOOBFromVC(oobRequest)
	if err != nil {
		framework.LoggingRespondErrMsg(c, err.Error(), http.StatusNotFound)
		return
	}
	framework.Respond(c, oobResponse, http.StatusOK)
}

func (cbr CredentialsBBSRouter) ListCredentials(c *gin.Context) {
	did := framework.GetQueryValue(c, "did")
	if did == nil {
		errMsg := fmt.Sprintf("get list of credentials request missing did query parameter")
		framework.LoggingRespondErrMsg(c, errMsg, http.StatusBadRequest)
		return
	}

	listResponse, err := cbr.service.ListCredentials(*did)
	if err != nil {
		framework.LoggingRespondErrMsg(c, err.Error(), http.StatusNotFound)
		return
	}
	framework.Respond(c, listResponse, http.StatusOK)
}

func (cbr CredentialsBBSRouter) ListCredentialsWithRender(c *gin.Context) {
	did := framework.GetQueryValue(c, "did")
	if did == nil {
		errMsg := fmt.Sprintf("get list of credentials with render request missing did query parameter")
		framework.LoggingRespondErrMsg(c, errMsg, http.StatusBadRequest)
		return
	}

	listResponse, err := cbr.service.ListCredentialsWithRender(*did)
	if err != nil {
		framework.LoggingRespondErrMsg(c, err.Error(), http.StatusNotFound)
		return
	}
	framework.Respond(c, listResponse, http.StatusOK)
}

func (cbr CredentialsBBSRouter) ProcessMessage(c *gin.Context) {
	var Request credentials_bbs.DIDCommMessage
	if err := framework.Decode(c.Request, &Request); err != nil {
		framework.LoggingRespondErrMsg(c, err.Error(), http.StatusBadRequest)
		return
	}

	err := cbr.service.ProcessMessage(Request)
	if err != nil {
		framework.LoggingRespondErrMsg(c, err.Error(), http.StatusNotFound)
		return
	}
	framework.Respond(c, nil, http.StatusOK)
}
