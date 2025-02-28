package manifest

import (
	"context"

	"github.com/TBD54566975/ssi-sdk/credential/manifest"
	sdkutil "github.com/TBD54566975/ssi-sdk/util"
	"github.com/goccy/go-json"
	"github.com/lestrrat-go/jwx/jws"
	"github.com/lestrrat-go/jwx/jwt"

	didint "github.com/tbd54566975/ssi-service/internal/did"
	"github.com/tbd54566975/ssi-service/internal/keyaccess"
)

func (s Service) verifyManifestJWT(ctx context.Context, token keyaccess.JWT) (*manifest.CredentialManifest, error) {
	// parse headers
	headers, err := keyaccess.GetJWTHeaders([]byte(token))
	if err != nil {
		return nil, sdkutil.LoggingErrorMsg(err, "could not parse JWT headers")
	}
	jwtKID, ok := headers.Get(jws.KeyIDKey)
	if !ok {
		return nil, sdkutil.LoggingNewError("JWT does not contain a kid")
	}
	kid, ok := jwtKID.(string)
	if !ok {
		return nil, sdkutil.LoggingNewError("JWT kid is not a string")
	}

	// parse token
	parsed, err := jwt.Parse([]byte(token))
	if err != nil {
		return nil, sdkutil.LoggingErrorMsg(err, "could not parse JWT")
	}

	claims := parsed.PrivateClaims()
	claimsJSONBytes, err := json.Marshal(claims)
	if err != nil {
		return nil, sdkutil.LoggingErrorMsg(err, "could not marshal claims")
	}

	var parsedManifest CredentialManifestContainer
	if err = json.Unmarshal(claimsJSONBytes, &parsedManifest); err != nil {
		return nil, sdkutil.LoggingErrorMsg(err, "unmarshalling claims into manifest")
	}

	if err = didint.VerifyTokenFromDID(ctx, s.didResolver, parsedManifest.Manifest.Issuer.ID, kid, token); err != nil {
		return nil, sdkutil.LoggingErrorMsg(err, "verifying manifest JWT")
	}
	return &parsedManifest.Manifest, nil
}
