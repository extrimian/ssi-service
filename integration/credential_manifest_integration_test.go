package integration

import (
	"strings"
	"testing"

	"github.com/extrimian/ssi-sdk/credential/parsing"
	"github.com/extrimian/ssi-sdk/crypto"
	"github.com/extrimian/ssi-sdk/did/key"
	"github.com/stretchr/testify/assert"

	"github.com/extrimian/ssi-service/pkg/service/operation/storage"
)

var credentialManifestContext = NewTestContext("CredentialManifest")

func TestCreateIssuerDIDKeyIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	didKeyOutput, err := CreateDIDKey()
	assert.NoError(t, err)

	issuerDID, err := getJSONElement(didKeyOutput, "$.did.id")
	assert.NoError(t, err)
	assert.Contains(t, issuerDID, "did:key")
	SetValue(credentialManifestContext, "issuerDID", issuerDID)

	verificationMethodID, err := getJSONElement(didKeyOutput, "$.did.verificationMethod[0].id")
	assert.NoError(t, err)
	assert.NotEmpty(t, verificationMethodID)
	SetValue(credentialManifestContext, "verificationMethodID", verificationMethodID)
}

func TestCreateAliceDIDKeyIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	applicantPrivKey, applicantDIDKey, err := key.GenerateDIDKey(crypto.Ed25519)
	assert.NoError(t, err)
	assert.NotEmpty(t, applicantPrivKey)
	assert.NotEmpty(t, applicantDIDKey)

	applicantDID, err := applicantDIDKey.Expand()
	assert.NoError(t, err)
	assert.NotEmpty(t, applicantDID)

	aliceDID := applicantDID.ID
	assert.Contains(t, aliceDID, "did:key")
	SetValue(credentialManifestContext, "aliceDID", aliceDID)

	aliceKID := applicantDID.VerificationMethod[0].ID
	assert.NotEmpty(t, aliceKID)
	SetValue(credentialManifestContext, "aliceKID", aliceKID)
	SetValue(credentialManifestContext, "aliceDIDPrivateKey", applicantPrivKey)
}

func TestCreateSchemaIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	output, err := CreateKYCSchema()
	assert.NoError(t, err)

	schemaID, err := getJSONElement(output, "$.id")
	assert.NoError(t, err)
	assert.NotEmpty(t, schemaID)
	SetValue(credentialManifestContext, "schemaID", schemaID)
}

func TestCreateVerifiableCredentialIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	issuerDID, err := GetValue(credentialManifestContext, "issuerDID")
	assert.NoError(t, err)
	assert.NotEmpty(t, issuerDID)

	verificationMethodID, err := GetValue(credentialManifestContext, "verificationMethodID")
	assert.NoError(t, err)
	assert.NotEmpty(t, verificationMethodID)

	schemaID, err := GetValue(credentialManifestContext, "schemaID")
	assert.NoError(t, err)
	assert.NotEmpty(t, schemaID)

	vcOutput, err := CreateVerifiableCredential(credInputParams{
		IssuerID:             issuerDID.(string),
		VerificationMethodID: verificationMethodID.(string),
		SchemaID:             schemaID.(string),
		SubjectID:            issuerDID.(string),
	})
	assert.NoError(t, err)
	assert.NotEmpty(t, vcOutput)

	credentialJWT, err := getJSONElement(vcOutput, "$.credentialJwt")
	assert.NoError(t, err)
	assert.NotEmpty(t, credentialJWT)
	SetValue(credentialManifestContext, "credentialJWT", credentialJWT)
}

func TestBatchCreateCredentialsIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	issuerDID, err := GetValue(credentialManifestContext, "issuerDID")
	assert.NoError(t, err)
	assert.NotEmpty(t, issuerDID)

	verificationMethodID, err := GetValue(credentialManifestContext, "verificationMethodID")
	assert.NoError(t, err)
	assert.NotEmpty(t, verificationMethodID)

	schemaID, err := GetValue(credentialManifestContext, "schemaID")
	assert.NoError(t, err)
	assert.NotEmpty(t, schemaID)

	vcsOutput, err := BatchCreateVerifiableCredentials(batchCredInputParams{
		IssuerID:             issuerDID.(string),
		VerificationMethodID: verificationMethodID.(string),
		SchemaID:             schemaID.(string),
		SubjectID0:           issuerDID.(string),
		SubjectID1:           issuerDID.(string),
		Suspendable0:         true,
		Revocable1:           true,
	})
	assert.NoError(t, err)
	assert.NotEmpty(t, vcsOutput)

	credentialJWT, err := getJSONElement(vcsOutput, "$.credentials[0].credentialJwt")
	assert.NoError(t, err)
	assert.NotEmpty(t, credentialJWT)

	credentialJWT1, err := getJSONElement(vcsOutput, "$.credentials[1].credentialJwt")
	assert.NoError(t, err)
	assert.NotEmpty(t, credentialJWT1)

	credentialID0, err := getJSONElement(vcsOutput, "$.credentials[0].credential.id")
	assert.NoError(t, err)

	credentialID1, err := getJSONElement(vcsOutput, "$.credentials[1].credential.id")
	assert.NoError(t, err)

	SetValue(credentialManifestContext, "credentialID0", credentialID0)
	SetValue(credentialManifestContext, "credentialID1", credentialID1)
}

func idFromURL(id string) string {
	lastIdx := strings.LastIndex(id, "/")
	return id[lastIdx+1:]
}

func TestBatchUpdateCredentialStatusIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	fullCredentialID0, err := GetValue(credentialManifestContext, "credentialID0")
	assert.NoError(t, err)
	assert.NotEmpty(t, fullCredentialID0)

	fullCredentialID1, err := GetValue(credentialManifestContext, "credentialID1")
	assert.NoError(t, err)
	assert.NotEmpty(t, fullCredentialID1)

	credentialID0 := idFromURL(fullCredentialID0.(string))
	credentialID1 := idFromURL(fullCredentialID1.(string))
	updatesOutput, err := BatchUpdateVerifiableCredentialStatuses(batchUpdateStatusInputParams{
		CredentialID0: credentialID0,
		Suspended0:    true,
		CredentialID1: credentialID1,
		Revoked1:      true,
	})
	assert.NoError(t, err)
	assert.NotEmpty(t, updatesOutput)

	id0, err := getJSONElement(updatesOutput, "$.credentialStatuses[0].id")
	assert.NoError(t, err)
	assert.Equal(t, credentialID0, id0)

	sus, err := getJSONElement(updatesOutput, "$.credentialStatuses[0].suspended")
	assert.NoError(t, err)
	assert.Equal(t, "true", sus)

	id1, err := getJSONElement(updatesOutput, "$.credentialStatuses[1].id")
	assert.NoError(t, err)
	assert.Equal(t, credentialID1, id1)

	rev, err := getJSONElement(updatesOutput, "$.credentialStatuses[1].revoked")
	assert.NoError(t, err)
	assert.Equal(t, "true", rev)
}

func TestBatchCreate100CredentialsIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	issuerDID, err := GetValue(credentialManifestContext, "issuerDID")
	assert.NoError(t, err)
	assert.NotEmpty(t, issuerDID)

	verificationMethodID, err := GetValue(credentialManifestContext, "verificationMethodID")
	assert.NoError(t, err)
	assert.NotEmpty(t, verificationMethodID)

	schemaID, err := GetValue(credentialManifestContext, "schemaID")
	assert.NoError(t, err)
	assert.NotEmpty(t, schemaID)

	// This test is simply about making sure we can create the maximum configured by default.
	vcsOutput, err := BatchCreate100VerifiableCredentials(credInputParams{
		IssuerID:             issuerDID.(string),
		VerificationMethodID: verificationMethodID.(string),
		SchemaID:             schemaID.(string),
		SubjectID:            issuerDID.(string),
	})
	assert.NoError(t, err)
	assert.NotEmpty(t, vcsOutput)
}

func TestCreateCredentialManifestIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	issuerDID, err := GetValue(credentialManifestContext, "issuerDID")
	assert.NoError(t, err)
	assert.NotEmpty(t, issuerDID)

	verificationMethodID, err := GetValue(credentialManifestContext, "verificationMethodID")
	assert.NoError(t, err)
	assert.NotEmpty(t, verificationMethodID)

	schemaID, err := GetValue(credentialManifestContext, "schemaID")
	assert.NoError(t, err)
	assert.NotEmpty(t, schemaID)

	cmOutput, err := CreateCredentialManifest(credManifestParams{
		IssuerID:             issuerDID.(string),
		VerificationMethodID: verificationMethodID.(string),
		SchemaID:             schemaID.(string),
	})
	assert.NoError(t, err)

	presentationDefinitionID, err := getJSONElement(cmOutput, "$.credential_manifest.presentation_definition.id")
	assert.NoError(t, err)
	assert.NotEmpty(t, presentationDefinitionID)
	SetValue(credentialManifestContext, "presentationDefinitionID", presentationDefinitionID)

	manifestID, err := getJSONElement(cmOutput, "$.credential_manifest.id")
	assert.NoError(t, err)
	assert.NotEmpty(t, manifestID)
	SetValue(credentialManifestContext, "manifestID", manifestID)
}

func TestCreateIssuanceTemplateIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	issuerDID, err := GetValue(credentialManifestContext, "issuerDID")
	assert.NoError(t, err)
	assert.NotEmpty(t, issuerDID)

	verificationMethodID, err := GetValue(credentialManifestContext, "verificationMethodID")
	assert.NoError(t, err)
	assert.NotEmpty(t, verificationMethodID)

	schemaID, err := GetValue(credentialManifestContext, "schemaID")
	assert.NoError(t, err)
	assert.NotEmpty(t, schemaID)

	cmOutput, err := CreateCredentialManifest(credManifestParams{
		IssuerID:             issuerDID.(string),
		VerificationMethodID: verificationMethodID.(string),
		SchemaID:             schemaID.(string),
	})
	assert.NoError(t, err)

	manifestID, err := getJSONElement(cmOutput, "$.credential_manifest.id")
	assert.NoError(t, err)
	assert.NotEmpty(t, manifestID)
	SetValue(credentialManifestContext, "manifestWithIssuanceTemplateID", manifestID)

	presentationDefinitionID, err := getJSONElement(cmOutput, "$.credential_manifest.presentation_definition.id")
	assert.NoError(t, err)
	assert.NotEmpty(t, presentationDefinitionID)
	SetValue(credentialManifestContext, "presentationDefinitionWithIssuanceTemplateID", presentationDefinitionID)

	itOutput, err := CreateIssuanceTemplate(issuanceTemplateParams{
		SchemaID:             schemaID.(string),
		ManifestID:           manifestID,
		IssuerID:             issuerDID.(string),
		VerificationMethodID: verificationMethodID.(string),
	})
	assert.NoError(t, err)

	issuanceTemplateID, err := getJSONElement(itOutput, "$.id")
	assert.NoError(t, err)
	assert.NotEmpty(t, issuanceTemplateID)
	SetValue(credentialManifestContext, "issuanceTemplateID", issuanceTemplateID)
}

func TestSubmitApplicationWithIssuanceTemplateIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	credentialJWT, err := GetValue(credentialManifestContext, "credentialJWT")
	assert.NoError(t, err)
	assert.NotEmpty(t, credentialJWT)

	presentationDefinitionID, err := GetValue(credentialManifestContext, "presentationDefinitionWithIssuanceTemplateID")
	assert.NoError(t, err)
	assert.NotEmpty(t, presentationDefinitionID)

	manifestID, err := GetValue(credentialManifestContext, "manifestWithIssuanceTemplateID")
	assert.NoError(t, err)
	assert.NotEmpty(t, manifestID)

	aliceDID, err := GetValue(credentialManifestContext, "aliceDID")
	assert.NoError(t, err)
	assert.NotEmpty(t, aliceDID)

	aliceKID, err := GetValue(credentialManifestContext, "aliceKID")
	assert.NoError(t, err)
	assert.NotEmpty(t, aliceKID)

	aliceDIDPrivateKey, err := GetValue(credentialManifestContext, "aliceDIDPrivateKey")
	assert.NoError(t, err)
	assert.NotEmpty(t, aliceDIDPrivateKey)

	credAppJWT, err := CreateCredentialApplicationJWT(credApplicationParams{
		DefinitionID: presentationDefinitionID.(string),
		ManifestID:   manifestID.(string),
	}, credentialJWT.(string), aliceDID.(string), aliceKID.(string), aliceDIDPrivateKey)
	assert.NoError(t, err)
	assert.NotEmpty(t, credAppJWT)

	submitApplicationOutput, err := SubmitApplication(applicationParams{
		ApplicationJWT: credAppJWT,
	})
	assert.NoError(t, err)
	assert.NotEmpty(t, submitApplicationOutput)

	isDone, err := getJSONElement(submitApplicationOutput, "$.done")
	assert.NoError(t, err)
	assert.Equal(t, "true", isDone)

	credentialResponseID, err := getJSONElement(submitApplicationOutput, "$.result.response.credential_response.id")
	assert.NoError(t, err)

	opCredentialResponse, err := getJSONElement(submitApplicationOutput, "$.result.response")
	assert.NoError(t, err)

	responsesOutput, err := get(endpoint + version + "manifests/responses/" + credentialResponseID)
	assert.NoError(t, err)

	assert.JSONEq(t, responsesOutput, opCredentialResponse)
}
func TestSubmitAndReviewApplicationIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	credentialJWT, err := GetValue(credentialManifestContext, "credentialJWT")
	assert.NoError(t, err)
	assert.NotEmpty(t, credentialJWT)

	presentationDefinitionID, err := GetValue(credentialManifestContext, "presentationDefinitionID")
	assert.NoError(t, err)
	assert.NotEmpty(t, presentationDefinitionID)

	manifestID, err := GetValue(credentialManifestContext, "manifestID")
	assert.NoError(t, err)
	assert.NotEmpty(t, manifestID)

	aliceDID, err := GetValue(credentialManifestContext, "aliceDID")
	assert.NoError(t, err)
	assert.NotEmpty(t, aliceDID)

	aliceKID, err := GetValue(credentialManifestContext, "aliceKID")
	assert.NoError(t, err)
	assert.NotEmpty(t, aliceKID)

	aliceDIDPrivateKey, err := GetValue(credentialManifestContext, "aliceDIDPrivateKey")
	assert.NoError(t, err)
	assert.NotEmpty(t, aliceDIDPrivateKey)

	credAppJWT, err := CreateCredentialApplicationJWT(credApplicationParams{
		DefinitionID: presentationDefinitionID.(string),
		ManifestID:   manifestID.(string),
	}, credentialJWT.(string), aliceDID.(string), aliceKID.(string), aliceDIDPrivateKey)
	assert.NoError(t, err)
	assert.NotEmpty(t, credAppJWT)

	submitApplicationOutput, err := SubmitApplication(applicationParams{
		ApplicationJWT: credAppJWT,
	})
	assert.NoError(t, err)
	assert.NotEmpty(t, submitApplicationOutput)

	isDone, err := getJSONElement(submitApplicationOutput, "$.done")
	assert.NoError(t, err)
	assert.Equal(t, "false", isDone)
	opID, err := getJSONElement(submitApplicationOutput, "$.id")
	assert.NoError(t, err)

	reviewApplicationOutput, err := ReviewApplication(reviewApplicationParams{
		ID:       storage.StatusObjectID(opID),
		Approved: true,
		Reason:   "oh yeah im testing",
	})
	assert.NoError(t, err)

	crManifestID, err := getJSONElement(reviewApplicationOutput, "$.credential_response.manifest_id")
	assert.NoError(t, err)
	assert.Equal(t, manifestID, crManifestID)

	vc, err := getJSONElement(reviewApplicationOutput, "$.verifiableCredentials[0]")
	assert.NoError(t, err)
	assert.NotEmpty(t, vc)
	_, _, typedVC, err := parsing.ToCredential(vc)
	assert.NoError(t, err)
	assert.Equal(t, "Mister", typedVC.CredentialSubject["givenName"])
	assert.Equal(t, "Tee", typedVC.CredentialSubject["familyName"])

	operationOutput, err := get(endpoint + version + "operations/" + opID)
	assert.NoError(t, err)
	isDone, err = getJSONElement(operationOutput, "$.done")
	assert.NoError(t, err)
	assert.Equal(t, "true", isDone)

	opCredentialResponse, err := getJSONElement(operationOutput, "$.result.response")
	assert.NoError(t, err)
	assert.JSONEq(t, reviewApplicationOutput, opCredentialResponse)
}
