// Package cognito provides AWS Cognito JWT verification for generaljwt middleware.
package cognito

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"go-libs/middlewares/generaljwt"

	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/lestrrat-go/jwx/v2/jwt"
)

const (
	TokenUseAccess = "access"
	TokenUseID     = "id"
)

type Config struct {
	Region     string
	UserPoolID string
	ClientID   string

	// TokenUse should usually be "access" for API authorization.
	// Use "id" only if you intentionally want to accept ID tokens.
	TokenUse string
}

type Verifier struct {
	config Config
	jwks   jwk.Set
}

func NewVerifier(ctx context.Context, config Config) (*Verifier, error) {
	if err := config.validate(); err != nil {
		return nil, err
	}

	jwksURL := config.jwksURL()

	set, err := jwk.Fetch(ctx, jwksURL)
	if err != nil {
		return nil, fmt.Errorf("fetch cognito jwks: %w", err)
	}

	return &Verifier{
		config: config,
		jwks:   set,
	}, nil
}

func (v *Verifier) Verify(ctx context.Context, rawToken string) (*generaljwt.Claims, error) {
	if v == nil {
		return nil, errors.New("nil cognito verifier")
	}

	if v.jwks == nil {
		return nil, errors.New("cognito jwks not configured")
	}

	token, err := jwt.Parse(
		[]byte(rawToken),
		jwt.WithKeySet(v.jwks),
		jwt.WithValidate(true),
		jwt.WithIssuer(v.config.issuer()),
	)
	if err != nil {
		return nil, fmt.Errorf("parse and validate jwt: %w", err)
	}

	if err := v.validateTokenUse(token); err != nil {
		return nil, err
	}

	if err := v.validateClient(token); err != nil {
		return nil, err
	}

	return mapClaims(token), nil
}

func (v *Verifier) validateTokenUse(token jwt.Token) error {
	tokenUse, ok := stringClaim(token, "token_use")
	if !ok {
		return errors.New("missing token_use claim")
	}

	if tokenUse != v.config.TokenUse {
		return fmt.Errorf("invalid token_use: got %q, want %q", tokenUse, v.config.TokenUse)
	}

	return nil
}

func (v *Verifier) validateClient(token jwt.Token) error {
	switch v.config.TokenUse {
	case TokenUseAccess:
		clientID, ok := stringClaim(token, "client_id")
		if !ok {
			return errors.New("missing client_id claim")
		}

		if clientID != v.config.ClientID {
			return errors.New("invalid client_id claim")
		}

		return nil

	case TokenUseID:
		audience := token.Audience()
		for _, aud := range audience {
			if aud == v.config.ClientID {
				return nil
			}
		}

		return errors.New("invalid audience claim")

	default:
		return fmt.Errorf("unsupported token_use config: %q", v.config.TokenUse)
	}
}

func mapClaims(token jwt.Token) *generaljwt.Claims {
	raw := make(map[string]any)

	iter := token.Iterate(context.Background())
	for iter.Next(context.Background()) {
		pair := iter.Pair()
		key, ok := pair.Key.(string)
		if !ok {
			continue
		}

		raw[key] = pair.Value
	}

	claims := &generaljwt.Claims{
		Subject: token.Subject(),
		Raw:     raw,
	}

	if username, ok := stringClaim(token, "username"); ok {
		claims.Username = username
	} else if username, ok := stringClaim(token, "cognito:username"); ok {
		claims.Username = username
	}

	if groups, ok := stringSliceClaim(token, "cognito:groups"); ok {
		claims.Groups = groups
	}

	if scope, ok := stringClaim(token, "scope"); ok {
		claims.Scopes = strings.Fields(scope)
	}

	return claims
}

func stringClaim(token jwt.Token, key string) (string, bool) {
	value, ok := token.Get(key)
	if !ok {
		return "", false
	}

	s, ok := value.(string)
	return s, ok
}

func stringSliceClaim(token jwt.Token, key string) ([]string, bool) {
	value, ok := token.Get(key)
	if !ok {
		return nil, false
	}

	switch v := value.(type) {
	case []string:
		return v, true

	case []any:
		out := make([]string, 0, len(v))
		for _, item := range v {
			s, ok := item.(string)
			if !ok {
				return nil, false
			}
			out = append(out, s)
		}
		return out, true

	default:
		return nil, false
	}
}

func (c Config) validate() error {
	if c.Region == "" {
		return errors.New("missing cognito region")
	}

	if c.UserPoolID == "" {
		return errors.New("missing cognito user pool id")
	}

	if c.ClientID == "" {
		return errors.New("missing cognito client id")
	}

	if c.TokenUse == "" {
		return errors.New("missing cognito token use")
	}

	if c.TokenUse != TokenUseAccess && c.TokenUse != TokenUseID {
		return fmt.Errorf("invalid cognito token use: %q", c.TokenUse)
	}

	return nil
}

func (c Config) issuer() string {
	return fmt.Sprintf(
		"https://cognito-idp.%s.amazonaws.com/%s",
		c.Region,
		c.UserPoolID,
	)
}

func (c Config) jwksURL() string {
	return c.issuer() + "/.well-known/jwks.json"
}
