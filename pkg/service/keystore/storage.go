package keystore

import (
	"context"
	"time"

	"github.com/TBD54566975/ssi-sdk/crypto"
	"github.com/TBD54566975/ssi-sdk/crypto/jwx"
	sdkutil "github.com/TBD54566975/ssi-sdk/util"
	"github.com/benbjohnson/clock"
	"github.com/goccy/go-json"
	"github.com/mr-tron/base58"
	"github.com/pkg/errors"

	"github.com/tbd54566975/ssi-service/internal/encryption"
	"github.com/tbd54566975/ssi-service/pkg/storage"
)

// StoredKey represents a common data model to store data on all key types
type StoredKey struct {
	ID         string         `json:"id"`
	Controller string         `json:"controller"`
	KeyType    crypto.KeyType `json:"keyType"`
	Base58Key  string         `json:"key"`
	Revoked    bool           `json:"revoked"`
	RevokedAt  string         `json:"revokedAt"`
	CreatedAt  string         `json:"createdAt"`
}

// KeyDetails represents a common data model to get information about a key, without revealing the key itself
type KeyDetails struct {
	ID           string           `json:"id"`
	Controller   string           `json:"controller"`
	KeyType      crypto.KeyType   `json:"keyType"`
	Revoked      bool             `json:"revoked"`
	RevokedAt    string           `json:"revokedAt"`
	CreatedAt    string           `json:"createdAt"`
	PublicKeyJWK jwx.PublicKeyJWK `json:"publicKeyJwk"`
}

type ServiceKey struct {
	Base58Key  string
	Base58Salt string
}

const (
	namespace             = "keystore"
	serviceInternalSuffix = "service-internal"
	publicNamespaceSuffix = "public-keys"
	keyNotFoundErrMsg     = "key not found"

	ServiceKeyEncryptionKey  = "ssi-service-key-encryption-key"
	ServiceDataEncryptionKey = "ssi-service-data-key"
)

var (
	serviceInternalNamespace = storage.Join(namespace, serviceInternalSuffix)
	publicKeyNamespace       = storage.Join(namespace, publicNamespaceSuffix)
)

type Storage struct {
	db        storage.ServiceStorage
	tx        storage.Tx
	encrypter encryption.Encrypter
	decrypter encryption.Decrypter
	Clock     clock.Clock
}

func NewKeyStoreStorage(db storage.ServiceStorage, e encryption.Encrypter, d encryption.Decrypter, writer storage.Tx) (*Storage, error) {
	s := &Storage{
		db:        db,
		encrypter: e,
		decrypter: d,
		Clock:     clock.New(),
		tx:        db,
	}
	if writer != nil {
		s.tx = writer
	}
	if s.encrypter == nil {
		s.encrypter = encryption.NoopEncrypter
	}
	if s.decrypter == nil {
		s.decrypter = encryption.NoopDecrypter
	}

	return s, nil
}

// ensureEncryptionKeyExists makes sure that the service key that will be used for encryption exists. This function is
// idempotent, so that multiple instances of ssi-service can call it on boot.
func ensureEncryptionKeyExists(config encryption.ExternalEncryptionConfig, provider storage.ServiceStorage, namespace, encryptionMaterialKey string) error {
	if config.GetMasterKeyURI() != "" {
		return nil
	}

	watchKeys := []storage.WatchKey{{
		Namespace: namespace,
		Key:       encryptionMaterialKey,
	}}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := provider.Execute(ctx, func(ctx context.Context, tx storage.Tx) (any, error) {
		// Create the key only if it doesn't already exist.
		gotKey, err := getServiceKey(ctx, provider, namespace, encryptionMaterialKey)
		if gotKey == nil && err.Error() == keyNotFoundErrMsg {
			serviceKey, err := GenerateServiceKey()
			if err != nil {
				return nil, errors.Wrap(err, "generating service key")
			}

			key := ServiceKey{
				Base58Key: serviceKey,
			}
			if err := storeServiceKey(ctx, tx, key, namespace, encryptionMaterialKey); err != nil {
				return nil, err
			}
			return nil, nil
		}
		return nil, err
	}, watchKeys)
	if err != nil {
		return err
	}
	return nil
}

// NewServiceEncryption creates a pair of Encrypter and Decrypter with the given configuration.
func NewServiceEncryption(db storage.ServiceStorage, cfg encryption.ExternalEncryptionConfig, key string) (encryption.Encrypter, encryption.Decrypter, error) {
	if !cfg.EncryptionEnabled() {
		return nil, nil, nil
	}

	if len(cfg.GetMasterKeyURI()) != 0 {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return encryption.NewExternalEncrypter(ctx, cfg)
	}

	if err := ensureEncryptionKeyExists(cfg, db, serviceInternalNamespace, key); err != nil {
		return nil, nil, errors.Wrap(err, "ensuring that the encryption key exists")
	}
	encSuite := encryption.NewXChaCha20Poly1305EncrypterWithKeyResolver(func(ctx context.Context) ([]byte, error) {
		return getServiceKey(ctx, db, serviceInternalNamespace, key)
	})
	return encSuite, encSuite, nil
}

// TODO(gabe): support more robust service key operations, including rotation, and caching
func storeServiceKey(ctx context.Context, tx storage.Tx, key ServiceKey, namespace string, skKey string) error {
	keyBytes, err := json.Marshal(key)
	if err != nil {
		return sdkutil.LoggingErrorMsg(err, "could not marshal service key")
	}
	if err = tx.Write(ctx, namespace, skKey, keyBytes); err != nil {
		return sdkutil.LoggingErrorMsg(err, "could store marshal service key")
	}
	return nil
}

func getServiceKey(ctx context.Context, db storage.ServiceStorage, namespace, skKey string) ([]byte, error) {
	storedKeyBytes, err := db.Read(ctx, namespace, skKey)
	if err != nil {
		return nil, sdkutil.LoggingErrorMsg(err, "could not get service key")
	}
	if len(storedKeyBytes) == 0 {
		return nil, sdkutil.LoggingNewError(keyNotFoundErrMsg)
	}

	var stored ServiceKey
	if err = json.Unmarshal(storedKeyBytes, &stored); err != nil {
		return nil, sdkutil.LoggingErrorMsg(err, "could not unmarshal service key")
	}

	keyBytes, err := base58.Decode(stored.Base58Key)
	if err != nil {
		return nil, errors.Wrap(err, "could not decode service key")
	}

	return keyBytes, nil
}

func (kss *Storage) StoreKey(ctx context.Context, key StoredKey) error {
	// TODO(gabe): conflict checking on key id
	id := key.ID
	if id == "" {
		return sdkutil.LoggingNewError("could not store key without an ID")
	}

	keyBytes, err := json.Marshal(key)
	if err != nil {
		return sdkutil.LoggingErrorMsg(err, "deserializing key from base58")
	}

	skBytes, err := base58.Decode(key.Base58Key)
	if err != nil {
		return sdkutil.LoggingErrorMsg(err, "deserializing key from base58")
	}

	secretKey, err := crypto.BytesToPrivKey(skBytes, key.KeyType)
	if err != nil {
		return sdkutil.LoggingErrorMsg(err, "reconstructing private key from input")
	}

	publicJWK, _, err := jwx.PrivateKeyToPrivateKeyJWK(key.ID, secretKey)
	if err != nil {
		return sdkutil.LoggingErrorMsg(err, "reconstructing JWK")
	}

	publicBytes, err := json.Marshal(publicJWK)
	if err != nil {
		return sdkutil.LoggingErrorMsg(err, "marshalling JWK")
	}

	if err := kss.tx.Write(ctx, publicKeyNamespace, id, publicBytes); err != nil {
		return sdkutil.LoggingErrorMsgf(err, "writing public key")
	}

	// encrypt key before storing
	encryptedKey, err := kss.encrypter.Encrypt(ctx, keyBytes, nil)
	if err != nil {
		return sdkutil.LoggingErrorMsgf(err, "could not encrypt key: %s", key.ID)
	}

	return kss.tx.Write(ctx, namespace, id, encryptedKey)
}

// RevokeKey revokes a key by setting the revoked flag to true.
func (kss *Storage) RevokeKey(ctx context.Context, id string) error {
	key, err := kss.GetKey(ctx, id)
	if err != nil {
		return err
	}
	if key == nil {
		return sdkutil.LoggingNewErrorf("key not found: %s", id)
	}

	key.Revoked = true
	key.RevokedAt = kss.Clock.Now().Format(time.RFC3339)
	return kss.StoreKey(ctx, *key)
}

func (kss *Storage) GetKey(ctx context.Context, id string) (*StoredKey, error) {
	storedKeyBytes, err := kss.db.Read(ctx, namespace, id)
	if err != nil {
		return nil, sdkutil.LoggingErrorMsgf(err, "getting key details for key: %s", id)
	}
	if len(storedKeyBytes) == 0 {
		return nil, sdkutil.LoggingNewErrorf("could not find key details for key: %s", id)
	}

	// decrypt key before unmarshalling
	decryptedKey, err := kss.decrypter.Decrypt(ctx, storedKeyBytes, nil)
	if err != nil {
		return nil, sdkutil.LoggingErrorMsgf(err, "could not decrypt key: %s", id)
	}

	var stored StoredKey
	if err = json.Unmarshal(decryptedKey, &stored); err != nil {
		return nil, sdkutil.LoggingErrorMsgf(err, "unmarshalling stored key: %s", id)
	}
	return &stored, nil
}

func (kss *Storage) GetKeyDetails(ctx context.Context, id string) (*KeyDetails, error) {
	stored, err := kss.GetKey(ctx, id)
	if err != nil {
		return nil, sdkutil.LoggingErrorMsgf(err, "reading details for private key %q", id)
	}

	storedPublicKeyBytes, err := kss.db.Read(ctx, publicKeyNamespace, id)
	if err != nil {
		return nil, sdkutil.LoggingErrorMsgf(err, "reading details for public key %q", id)
	}
	var storedPublicKey jwx.PublicKeyJWK
	if err = json.Unmarshal(storedPublicKeyBytes, &storedPublicKey); err != nil {
		return nil, sdkutil.LoggingErrorMsgf(err, "unmarshalling public key")
	}

	return &KeyDetails{
		ID:           stored.ID,
		Controller:   stored.Controller,
		KeyType:      stored.KeyType,
		CreatedAt:    stored.CreatedAt,
		Revoked:      stored.Revoked,
		PublicKeyJWK: storedPublicKey,
	}, nil
}
