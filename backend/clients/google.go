package clients

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const googleTokenInfoURL = "https://oauth2.googleapis.com/tokeninfo"

var (
	// ErrGoogleTokenInvalid means Google rejected the ID token, or it did not
	// belong to this application / a verified email address.
	ErrGoogleTokenInvalid = errors.New("clients: google id token is invalid")
	// ErrGoogleVerificationUnavailable means the token could not be checked
	// because Google's verification endpoint was unreachable or failed.
	ErrGoogleVerificationUnavailable = errors.New("clients: google token verification unavailable")
)

// GoogleIdentity is the subset of a verified Google ID token that the
// application needs to create or link an account.
type GoogleIdentity struct {
	Subject       string
	Email         string
	EmailVerified bool
	Name          string
	Picture       string
}

// GoogleTokenVerifier verifies Google ID tokens (the "credential" returned by
// Google Identity Services in the browser) and extracts the user's identity.
//
// Verification is delegated to Google's tokeninfo endpoint, which validates
// the token signature, issuer, and expiry. The verifier additionally makes
// sure the token was issued for this application and that the email address
// has been verified by Google.
type GoogleTokenVerifier struct {
	clientID     string
	tokenInfoURL string
	httpClient   *http.Client
	now          func() time.Time
}

// NewGoogleTokenVerifier creates a verifier for ID tokens issued to clientID.
// When httpClient is nil a client with a 10 second timeout is used.
func NewGoogleTokenVerifier(clientID string, httpClient *http.Client) (*GoogleTokenVerifier, error) {
	clientID = strings.TrimSpace(clientID)
	if clientID == "" {
		return nil, fmt.Errorf("google client id is required but was empty")
	}

	if httpClient == nil {
		httpClient = &http.Client{Timeout: 10 * time.Second}
	}

	return &GoogleTokenVerifier{
		clientID:     clientID,
		tokenInfoURL: googleTokenInfoURL,
		httpClient:   httpClient,
		now:          time.Now,
	}, nil
}

// Verify checks a Google ID token and returns the identity it describes.
func (v *GoogleTokenVerifier) Verify(ctx context.Context, idToken string) (*GoogleIdentity, error) {
	idToken = strings.TrimSpace(idToken)
	if idToken == "" {
		return nil, fmt.Errorf("%w: empty token", ErrGoogleTokenInvalid)
	}

	endpoint := v.tokenInfoURL + "?id_token=" + url.QueryEscape(idToken)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("%w: build request", ErrGoogleVerificationUnavailable)
	}

	resp, err := v.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrGoogleVerificationUnavailable, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode >= 400 && resp.StatusCode < 500 {
			return nil, fmt.Errorf("%w: tokeninfo status %d", ErrGoogleTokenInvalid, resp.StatusCode)
		}
		return nil, fmt.Errorf("%w: tokeninfo status %d", ErrGoogleVerificationUnavailable, resp.StatusCode)
	}

	var payload googleTokenInfo
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, fmt.Errorf("%w: decode tokeninfo response", ErrGoogleVerificationUnavailable)
	}

	if err := v.validate(&payload); err != nil {
		return nil, err
	}

	return &GoogleIdentity{
		Subject:       payload.Subject,
		Email:         strings.ToLower(strings.TrimSpace(payload.Email)),
		EmailVerified: payload.EmailVerified.Truthy(),
		Name:          strings.TrimSpace(payload.Name),
		Picture:       payload.Picture,
	}, nil
}

func (v *GoogleTokenVerifier) validate(payload *googleTokenInfo) error {
	if payload.Subject == "" {
		return fmt.Errorf("%w: missing subject", ErrGoogleTokenInvalid)
	}

	if payload.Issuer != "accounts.google.com" && payload.Issuer != "https://accounts.google.com" {
		return fmt.Errorf("%w: unexpected issuer", ErrGoogleTokenInvalid)
	}

	if payload.Audience != v.clientID {
		return fmt.Errorf("%w: audience mismatch", ErrGoogleTokenInvalid)
	}

	if strings.TrimSpace(payload.Email) == "" || !payload.EmailVerified.Truthy() {
		return fmt.Errorf("%w: email is missing or unverified", ErrGoogleTokenInvalid)
	}

	if payload.ExpiresAt.set && v.now().Unix() >= payload.ExpiresAt.value {
		return fmt.Errorf("%w: token expired", ErrGoogleTokenInvalid)
	}

	return nil
}

// googleTokenInfo is the JSON payload returned by Google's tokeninfo endpoint.
// `email_verified` and `exp` are encoded as strings by some Google services,
// so both accept string and native JSON forms.
type googleTokenInfo struct {
	Issuer        string    `json:"iss"`
	Audience      string    `json:"aud"`
	Subject       string    `json:"sub"`
	Email         string    `json:"email"`
	EmailVerified flexBool  `json:"email_verified"`
	Name          string    `json:"name"`
	Picture       string    `json:"picture"`
	ExpiresAt     flexInt64 `json:"exp"`
}

// flexBool decodes a JSON value that may be a boolean or a string ("true").
type flexBool struct {
	value bool
	set   bool
}

func (b *flexBool) UnmarshalJSON(data []byte) error {
	b.set = true

	var boolean bool
	if err := json.Unmarshal(data, &boolean); err == nil {
		b.value = boolean
		return nil
	}

	var text string
	if err := json.Unmarshal(data, &text); err == nil {
		b.value = strings.EqualFold(strings.TrimSpace(text), "true")
		return nil
	}

	return fmt.Errorf("clients: cannot decode google tokeninfo boolean: %s", string(data))
}

// Truthy reports whether the value was true. A missing value is false.
func (b flexBool) Truthy() bool { return b.set && b.value }

// flexInt64 decodes a JSON number or a numeric string.
type flexInt64 struct {
	value int64
	set   bool
}

func (n *flexInt64) UnmarshalJSON(data []byte) error {
	n.set = true

	var number int64
	if err := json.Unmarshal(data, &number); err == nil {
		n.value = number
		return nil
	}

	var text string
	if err := json.Unmarshal(data, &text); err == nil {
		parsed, err := strconv.ParseInt(strings.TrimSpace(text), 10, 64)
		if err != nil {
			return fmt.Errorf("clients: cannot decode google tokeninfo integer: %s", text)
		}
		n.value = parsed
		return nil
	}

	return fmt.Errorf("clients: cannot decode google tokeninfo integer: %s", string(data))
}
