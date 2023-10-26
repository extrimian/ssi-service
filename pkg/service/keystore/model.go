package keystore

import (
	gocrypto "crypto"

	"github.com/extrimian/ssi-sdk/crypto"
	"github.com/extrimian/ssi-sdk/crypto/jwx"
)

type StoreKeyRequest struct {
	ID               string
	Type             crypto.KeyType
	Controller       string
	PrivateKeyBase58 string
}

type GetKeyRequest struct {
	ID string
}

type GetKeyResponse struct {
	ID         string
	Type       crypto.KeyType
	Controller string
	CreatedAt  string
	Revoked    bool
	RevokedAt  string
	Key        gocrypto.PrivateKey
}

type GetKeyDetailsRequest struct {
	ID string
}

type GetKeyDetailsResponse struct {
	ID           string
	Type         crypto.KeyType
	Controller   string
	CreatedAt    string
	Revoked      bool
	RevokedAt    string
	PublicKeyJWK jwx.PublicKeyJWK
}

type RevokeKeyRequest struct {
	ID string
}
