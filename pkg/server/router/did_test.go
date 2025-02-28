package router

import (
	"context"
	"strings"
	"testing"

	"github.com/TBD54566975/ssi-sdk/crypto"
	didsdk "github.com/TBD54566975/ssi-sdk/did"
	"github.com/stretchr/testify/assert"
	"github.com/tbd54566975/ssi-service/pkg/service/common"
	"github.com/tbd54566975/ssi-service/pkg/testutil"
	"gopkg.in/h2non/gock.v1"

	"github.com/tbd54566975/ssi-service/config"
	"github.com/tbd54566975/ssi-service/pkg/service/did"
	"github.com/tbd54566975/ssi-service/pkg/service/framework"
)

func TestDIDRouter(t *testing.T) {
	t.Run("Nil Service", func(tt *testing.T) {
		didRouter, err := NewDIDRouter(nil)
		assert.Error(tt, err)
		assert.Empty(tt, didRouter)
		assert.Contains(tt, err.Error(), "service cannot be nil")
	})

	t.Run("Bad Service", func(tt *testing.T) {
		didRouter, err := NewDIDRouter(&testService{})
		assert.Error(tt, err)
		assert.Empty(tt, didRouter)
		assert.Contains(tt, err.Error(), "could not create DID router with service type: test")
	})

	for _, test := range testutil.TestDatabases {
		t.Run(test.Name, func(t *testing.T) {
			// TODO: Fix pagesize issue on redis - https://github.com/TBD54566975/ssi-service/issues/538
			if !strings.Contains(test.Name, "Redis") {
				t.Run("List DIDs supports paging", func(tt *testing.T) {
					db := test.ServiceStorage(tt)
					assert.NotEmpty(tt, db)
					keyStoreService := testKeyStoreService(tt, db)
					methods := []string{didsdk.KeyMethod.String()}
					serviceConfig := config.DIDServiceConfig{Methods: methods, LocalResolutionMethods: methods}
					didService, err := did.NewDIDService(serviceConfig, db, keyStoreService, nil)
					assert.NoError(tt, err)
					assert.NotEmpty(tt, didService)
					createDID(tt, didService)
					createDID(tt, didService)

					one := 1
					listDIDsResponse1, err := didService.ListDIDsByMethod(context.Background(),
						did.ListDIDsRequest{
							Method: didsdk.KeyMethod,
							PageRequest: &common.Page{
								Size: one,
							},
						})

					assert.NoError(tt, err)
					assert.Len(tt, listDIDsResponse1.DIDs, 1)
					assert.NotEmpty(tt, listDIDsResponse1.NextPageToken)

					listDIDsResponse2, err := didService.ListDIDsByMethod(context.Background(),
						did.ListDIDsRequest{
							Method: didsdk.KeyMethod,
							PageRequest: &common.Page{
								Size:  one,
								Token: listDIDsResponse1.NextPageToken,
							},
						})

					assert.NoError(tt, err)
					assert.Len(tt, listDIDsResponse2.DIDs, 1)
					assert.Empty(tt, listDIDsResponse2.NextPageToken)

				})
			}

			t.Run("DID Service Test", func(tt *testing.T) {
				db := test.ServiceStorage(tt)
				assert.NotEmpty(tt, db)

				keyStoreService := testKeyStoreService(tt, db)
				methods := []string{didsdk.KeyMethod.String()}
				serviceConfig := config.DIDServiceConfig{Methods: methods, LocalResolutionMethods: methods}
				didService, err := did.NewDIDService(serviceConfig, db, keyStoreService, nil)
				assert.NoError(tt, err)
				assert.NotEmpty(tt, didService)

				// check type and status
				assert.Equal(tt, framework.DID, didService.Type())
				assert.Equal(tt, framework.StatusReady, didService.Status().Status)

				// get unknown handler
				_, err = didService.GetDIDByMethod(context.Background(), did.GetDIDRequest{Method: "bad"})
				assert.Error(tt, err)
				assert.Contains(tt, err.Error(), "could not get handler for method<bad>")

				supported := didService.GetSupportedMethods()
				assert.NotEmpty(tt, supported)
				assert.Len(tt, supported.Methods, 1)
				assert.Equal(tt, didsdk.KeyMethod, supported.Methods[0])

				// bad key type
				_, err = didService.CreateDIDByMethod(context.Background(), did.CreateDIDRequest{Method: didsdk.KeyMethod, KeyType: "bad"})
				assert.Error(tt, err)
				assert.Contains(tt, err.Error(), "could not create did:key")

				// good key type
				createDIDResponse, err := didService.CreateDIDByMethod(context.Background(), did.CreateDIDRequest{Method: didsdk.KeyMethod, KeyType: crypto.Ed25519})
				assert.NoError(tt, err)
				assert.NotEmpty(tt, createDIDResponse)

				// check the DID is a did:key
				assert.Contains(tt, createDIDResponse.DID.ID, "did:key")

				// get it back
				getDIDResponse, err := didService.GetDIDByMethod(context.Background(), did.GetDIDRequest{Method: didsdk.KeyMethod, ID: createDIDResponse.DID.ID})
				assert.NoError(tt, err)
				assert.NotEmpty(tt, getDIDResponse)

				// make sure it's the same value
				assert.Equal(tt, createDIDResponse.DID.ID, getDIDResponse.DID.ID)

				// create a second DID
				createDIDResponse2, err := didService.CreateDIDByMethod(context.Background(), did.CreateDIDRequest{Method: didsdk.KeyMethod, KeyType: crypto.Ed25519})
				assert.NoError(tt, err)
				assert.NotEmpty(tt, createDIDResponse2)

				// get all DIDs back
				getDIDsResponse, err := didService.ListDIDsByMethod(context.Background(), did.ListDIDsRequest{Method: didsdk.KeyMethod})
				assert.NoError(tt, err)
				assert.NotEmpty(tt, getDIDsResponse)
				assert.Len(tt, getDIDsResponse.DIDs, 2)

				knownDIDs := map[string]bool{createDIDResponse.DID.ID: true, createDIDResponse2.DID.ID: true}
				for _, gotDID := range getDIDsResponse.DIDs {
					if _, ok := knownDIDs[gotDID.ID]; !ok {
						tt.Error("got unknown DID")
					} else {
						delete(knownDIDs, gotDID.ID)
					}
				}
				assert.Len(tt, knownDIDs, 0)

				// delete dids
				err = didService.SoftDeleteDIDByMethod(context.Background(), did.DeleteDIDRequest{Method: didsdk.KeyMethod, ID: createDIDResponse.DID.ID})
				assert.NoError(tt, err)

				err = didService.SoftDeleteDIDByMethod(context.Background(), did.DeleteDIDRequest{Method: didsdk.KeyMethod, ID: createDIDResponse2.DID.ID})
				assert.NoError(tt, err)

				// get all DIDs back
				getDIDsResponse, err = didService.ListDIDsByMethod(context.Background(), did.ListDIDsRequest{Method: didsdk.KeyMethod})
				assert.NoError(tt, err)
				assert.NotEmpty(tt, getDIDsResponse)
				assert.Len(tt, getDIDsResponse.DIDs, 0)

				// get deleted DIDs back
				getDIDsResponse, err = didService.ListDIDsByMethod(context.Background(), did.ListDIDsRequest{Method: didsdk.KeyMethod, Deleted: true})
				assert.NoError(tt, err)
				assert.NotEmpty(tt, getDIDsResponse)
				assert.Len(tt, getDIDsResponse.DIDs, 2)
			})

			t.Run("DID Web Service Test", func(tt *testing.T) {
				db := test.ServiceStorage(tt)
				assert.NotEmpty(tt, db)

				keyStoreService := testKeyStoreService(tt, db)
				methods := []string{didsdk.KeyMethod.String(), didsdk.WebMethod.String()}
				serviceConfig := config.DIDServiceConfig{Methods: methods, LocalResolutionMethods: methods}
				didService, err := did.NewDIDService(serviceConfig, db, keyStoreService, nil)
				assert.NoError(tt, err)
				assert.NotEmpty(tt, didService)

				// check type and status
				assert.Equal(tt, framework.DID, didService.Type())
				assert.Equal(tt, framework.StatusReady, didService.Status().Status)

				// get unknown handler
				_, err = didService.GetDIDByMethod(context.Background(), did.GetDIDRequest{Method: "bad"})
				assert.Error(tt, err)
				assert.Contains(tt, err.Error(), "could not get handler for method<bad>")

				supported := didService.GetSupportedMethods()
				assert.NotEmpty(tt, supported)
				assert.Len(tt, supported.Methods, 2)

				assert.ElementsMatch(tt, supported.Methods, []didsdk.Method{didsdk.KeyMethod, didsdk.WebMethod})

				gock.Off()
				gock.New("https://example.com").
					Get("/.well-known/did.json").
					Reply(200).
					BodyString("")
				// bad key type
				createOpts := did.CreateWebDIDOptions{DIDWebID: "did:web:example.com"}
				_, err = didService.CreateDIDByMethod(context.Background(), did.CreateDIDRequest{Method: didsdk.WebMethod, KeyType: "bad", Options: createOpts})
				assert.Error(tt, err)
				assert.Contains(tt, err.Error(), "key type <bad> not supported")

				gock.Off()
				gock.New("https://example.com").
					Get("/.well-known/did.json").
					Reply(404)
				// good key type
				createDIDResponse, err := didService.CreateDIDByMethod(context.Background(), did.CreateDIDRequest{Method: didsdk.WebMethod, KeyType: crypto.Ed25519, Options: createOpts})
				assert.NoError(tt, err)
				assert.NotEmpty(tt, createDIDResponse)

				// check the DID is a did:key
				assert.Contains(tt, createDIDResponse.DID.ID, "did:web")

				// get it back
				getDIDResponse, err := didService.GetDIDByMethod(context.Background(), did.GetDIDRequest{Method: didsdk.WebMethod, ID: createDIDResponse.DID.ID})
				assert.NoError(tt, err)
				assert.NotEmpty(tt, getDIDResponse)

				// make sure it's the same value
				assert.Equal(tt, createDIDResponse.DID.ID, getDIDResponse.DID.ID)

				gock.Off()
				gock.New("https://tbd.website").
					Get("/.well-known/did.json").
					Reply(404)
				// create a second DID
				createOpts = did.CreateWebDIDOptions{DIDWebID: "did:web:tbd.website"}
				createDIDResponse2, err := didService.CreateDIDByMethod(context.Background(), did.CreateDIDRequest{Method: didsdk.WebMethod, KeyType: crypto.Ed25519, Options: createOpts})
				assert.NoError(tt, err)
				assert.NotEmpty(tt, createDIDResponse2)

				// get all DIDs back
				getDIDsResponse, err := didService.ListDIDsByMethod(context.Background(), did.ListDIDsRequest{Method: didsdk.WebMethod})
				assert.NoError(tt, err)
				assert.NotEmpty(tt, getDIDsResponse)
				assert.Len(tt, getDIDsResponse.DIDs, 2)

				knownDIDs := map[string]bool{createDIDResponse.DID.ID: true, createDIDResponse2.DID.ID: true}
				for _, gotDID := range getDIDsResponse.DIDs {
					if _, ok := knownDIDs[gotDID.ID]; !ok {
						tt.Error("got unknown DID")
					} else {
						delete(knownDIDs, gotDID.ID)
					}
				}
				assert.Len(tt, knownDIDs, 0)

				// delete dids
				err = didService.SoftDeleteDIDByMethod(context.Background(), did.DeleteDIDRequest{Method: didsdk.WebMethod, ID: createDIDResponse.DID.ID})
				assert.NoError(tt, err)

				err = didService.SoftDeleteDIDByMethod(context.Background(), did.DeleteDIDRequest{Method: didsdk.WebMethod, ID: createDIDResponse2.DID.ID})
				assert.NoError(tt, err)

				// get all DIDs back
				getDIDsResponse, err = didService.ListDIDsByMethod(context.Background(), did.ListDIDsRequest{Method: didsdk.WebMethod})
				assert.NoError(tt, err)
				assert.NotEmpty(tt, getDIDsResponse)
				assert.Len(tt, getDIDsResponse.DIDs, 0)
			})
		})
	}
}

func createDID(tt *testing.T, didService *did.Service) {
	createDIDResponse, err := didService.CreateDIDByMethod(context.Background(), did.CreateDIDRequest{Method: didsdk.KeyMethod, KeyType: crypto.Ed25519})
	assert.NoError(tt, err)
	assert.NotEmpty(tt, createDIDResponse)
}
