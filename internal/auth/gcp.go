package auth

import (
	"ClassiFaaS/internal/globals"
	"context"
	"fmt"
	"net/url"
	"os"

	"google.golang.org/api/idtoken"
)

// GetGoogleIdentityToken returns an OAuth2 ID token for authenticating requests
// to the specified targetURL. It uses service account credentials defined in
// globals.GCPServiceAccount.
//
// The returned token is valid for approximately one hour.

func GetGoogleIdentityToken(targetURL string) (string, error) {
	data, err := os.ReadFile(globals.GCPServiceAccount)
	if err != nil {
		return "", fmt.Errorf("failed to read service account JSON: %w", err)
	}

	ctx := context.Background()
	u, err := url.Parse(targetURL)
	if err != nil {
		return "", fmt.Errorf("invalid target URL: %w", err)
	}
	u.RawQuery = ""
	cleanURL := u.String()

	tokenSource, err := idtoken.NewTokenSource(ctx, cleanURL, idtoken.WithCredentialsJSON(data))
	if err != nil {
		return "", fmt.Errorf("failed to create ID token source: %w", err)
	}

	token, err := tokenSource.Token()
	if err != nil {
		return "", fmt.Errorf("failed to retrieve ID token: %w", err)
	}

	return token.AccessToken, nil
}
