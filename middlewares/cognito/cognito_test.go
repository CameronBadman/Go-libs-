package cognito

import (
	"testing"
)

func TestConfigValidateMissingRegion(t *testing.T) {
	config := Config{
		UserPoolID: "ap-southeast-2_example",
		ClientID:   "client-id",
		TokenUse:   TokenUseAccess,
	}

	if err := config.validate(); err == nil {
		t.Fatal("expected error")
	}
}

func TestConfigValidateMissingUserPoolID(t *testing.T) {
	config := Config{
		Region:   "ap-southeast-2",
		ClientID: "client-id",
		TokenUse: TokenUseAccess,
	}

	if err := config.validate(); err == nil {
		t.Fatal("expected error")
	}
}

func TestConfigValidateMissingClientID(t *testing.T) {
	config := Config{
		Region:     "ap-southeast-2",
		UserPoolID: "ap-southeast-2_example",
		TokenUse:   TokenUseAccess,
	}

	if err := config.validate(); err == nil {
		t.Fatal("expected error")
	}
}

func TestConfigValidateMissingTokenUse(t *testing.T) {
	config := Config{
		Region:     "ap-southeast-2",
		UserPoolID: "ap-southeast-2_example",
		ClientID:   "client-id",
	}

	if err := config.validate(); err == nil {
		t.Fatal("expected error")
	}
}

func TestConfigValidateInvalidTokenUse(t *testing.T) {
	config := Config{
		Region:     "ap-southeast-2",
		UserPoolID: "ap-southeast-2_example",
		ClientID:   "client-id",
		TokenUse:   "refresh",
	}

	if err := config.validate(); err == nil {
		t.Fatal("expected error")
	}
}

func TestConfigValidateAccessTokenUse(t *testing.T) {
	config := Config{
		Region:     "ap-southeast-2",
		UserPoolID: "ap-southeast-2_example",
		ClientID:   "client-id",
		TokenUse:   TokenUseAccess,
	}

	if err := config.validate(); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestConfigValidateIDTokenUse(t *testing.T) {
	config := Config{
		Region:     "ap-southeast-2",
		UserPoolID: "ap-southeast-2_example",
		ClientID:   "client-id",
		TokenUse:   TokenUseID,
	}

	if err := config.validate(); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestConfigIssuer(t *testing.T) {
	config := Config{
		Region:     "ap-southeast-2",
		UserPoolID: "ap-southeast-2_example",
		ClientID:   "client-id",
		TokenUse:   TokenUseAccess,
	}

	expected := "https://cognito-idp.ap-southeast-2.amazonaws.com/ap-southeast-2_example"

	if got := config.issuer(); got != expected {
		t.Fatalf("expected issuer %q, got %q", expected, got)
	}
}

func TestConfigJWKSURL(t *testing.T) {
	config := Config{
		Region:     "ap-southeast-2",
		UserPoolID: "ap-southeast-2_example",
		ClientID:   "client-id",
		TokenUse:   TokenUseAccess,
	}

	expected := "https://cognito-idp.ap-southeast-2.amazonaws.com/ap-southeast-2_example/.well-known/jwks.json"

	if got := config.jwksURL(); got != expected {
		t.Fatalf("expected jwks url %q, got %q", expected, got)
	}
}
