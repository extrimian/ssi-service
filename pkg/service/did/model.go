package did

import (
	gocrypto "crypto"

	"github.com/extrimian/ssi-sdk/crypto"
	"github.com/extrimian/ssi-sdk/crypto/jwx"
	didsdk "github.com/extrimian/ssi-sdk/did"
	"github.com/extrimian/ssi-sdk/did/ion"
	"github.com/extrimian/ssi-sdk/did/resolution"
	"github.com/extrimian/ssi-service/pkg/service/common"
)

type GetSupportedMethodsResponse struct {
	Methods []didsdk.Method `json:"method"`
}

type ResolveDIDRequest struct {
	DID string `json:"did" validate:"required"`
}

type ResolveDIDResponse struct {
	ResolutionMetadata  *resolution.Metadata         `json:"didResolutionMetadata,omitempty"`
	DIDDocument         *didsdk.Document             `json:"didDocument"`
	DIDDocumentMetadata *resolution.DocumentMetadata `json:"didDocumentMetadata,omitempty"`
}

type CreateDIDRequestOptions interface {
	Method() didsdk.Method
}

// CreateDIDRequest is the JSON-serializable request for creating a DID across DID method
type CreateDIDRequest struct {
	Method  didsdk.Method           `json:"method" validate:"required"`
	KeyType crypto.KeyType          `validate:"required"`
	Options CreateDIDRequestOptions `json:"options"`
}

// CreateDIDResponse is the JSON-serializable response for creating a DID
type CreateDIDResponse struct {
	DID didsdk.Document `json:"did"`
}

type BatchCreateDIDsRequest struct {
	Requests []CreateDIDRequest `json:"requests"`
}

type BatchCreateDIDsResponse struct {
	DIDs []didsdk.Document `json:"dids"`
}

type GetDIDRequest struct {
	Method didsdk.Method `json:"method" validate:"required"`
	ID     string        `json:"id" validate:"required"`
}

// GetDIDResponse is the JSON-serializable response for getting a DID
type GetDIDResponse struct {
	DID didsdk.Document `json:"did"`
}

type GetKeyFromDIDRequest struct {
	ID    string `json:"id" validate:"required"`
	KeyID string `json:"keyId,omitempty"`
}

type GetKeyFromDIDResponse struct {
	KeyID     string             `json:"keyId"`
	PublicKey gocrypto.PublicKey `json:"publicKey"`
}

type ListDIDsRequest struct {
	Method  didsdk.Method `json:"method" validate:"required"`
	Deleted bool          `json:"deleted"`

	PageRequest *common.Page
}

// ListDIDsResponse is the JSON-serializable response for getting all DIDs for a given method
type ListDIDsResponse struct {
	DIDs          []didsdk.Document `json:"dids"`
	NextPageToken string
}

type DeleteDIDRequest struct {
	Method didsdk.Method `json:"method" validate:"required"`
	ID     string        `json:"id" validate:"required"`
}

type UpdateIONDIDRequest struct {
	DID ion.ION `json:"did"`

	StateChange ion.StateChange `json:"stateChange"`
}

type UpdateIONDIDResponse struct {
	DID didsdk.Document `json:"did"`
}

type UpdateRequestStatus string

func (s UpdateRequestStatus) Bytes() []byte {
	return []byte(s)
}

const (
	PreAnchorStatus   UpdateRequestStatus = "pre-anchor"
	AnchorErrorStatus UpdateRequestStatus = "anchor-error"
	AnchoredStatus    UpdateRequestStatus = "anchored"
	DoneStatus        UpdateRequestStatus = "done"
)

type identityRequest struct {
	PublicKeys []publicKeyIdentityRequest `json:"publicKeys,omitempty"`
	Services   []DIDDocService            `json:"services,omitempty"`
	DidMethod  string                     `json:"didMethod,omitempty"`
}

type DIDDocService struct {
	ID              string `json:"id"`
	Type            string `json:"type"`
	ServiceEndpoint string `json:"serviceEndpoint"`
}

type publicKeyIdentityRequest struct {
	VmId                     string           `json:"vmId,omitempty"`
	PublicKeyJWK             jwx.PublicKeyJWK `json:"publicKeyJWK,omitempty"`
	VerificationRelationship string           `json:"verificationRelationship,omitempty"`
}

type PublicKeyResponse struct {
	PublicKeyJWK jwx.PublicKeyJWK `json:"publicKeyJWK,omitempty"`
}

type QuarkidIdentityResponse struct {
	DID string `json:"did,omitempty"`
}

type QuarkidIdentity struct {
	Context            []interface{}        `json:"@context"`
	ID                 string               `json:"id"`
	VerificationMethod []VerificationMethod `json:"verificationMethod"`
	KeyAgreement       []string             `json:"keyAgreement"`
	Service            []DIDDocService      `json:"service,omitempty"`
	AssertionMethod    []string             `json:"assertionMethod"`
}

type VerificationMethod struct {
	ID           string           `json:"id"`
	Controller   string           `json:"controller"`
	Type         string           `json:"type"`
	PublicKeyJWK jwx.PublicKeyJWK `json:"publicKeyJwk"`
}
