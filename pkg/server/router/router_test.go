package router

import (
	didsdk "github.com/extrimian/ssi-sdk/did"

	"github.com/extrimian/ssi-service/config"
	"github.com/extrimian/ssi-service/pkg/service/framework"
)

// generic test config to be used by all tests in this package

type testService struct{}

func (s *testService) Type() framework.Type {
	return "test"
}

func (s *testService) Status() framework.Status {
	return framework.Status{Status: "ready"}
}

func (s *testService) Config() config.ServicesConfig {
	return config.ServicesConfig{
		StorageProvider:  "bolt",
		KeyStoreConfig:   config.KeyStoreServiceConfig{},
		DIDConfig:        config.DIDServiceConfig{Methods: []string{string(didsdk.KeyMethod)}},
		CredentialConfig: config.CredentialServiceConfig{},
	}
}
