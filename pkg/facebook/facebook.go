package facebook

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/virtual-vgo/vvgo/pkg/http_wrappers"
	"github.com/virtual-vgo/vvgo/pkg/log"
	"net/http"
	"net/url"
)

var logger = log.Logger()
var ErrNon200Response = errors.New("non-200 response from facebook")
var ErrInvalidOAuthCode = errors.New("invalid oauth code")

type Client struct {
	config Config
}

func NewClient(config Config) *Client {
	return &Client{config: config}
}

type Config struct {
	Endpoint          string `default:"https://graph.facebook.com/v8.0"`
	OAuthClientID     string `envconfig:"oauth_client_id"`
	OAuthClientSecret string `envconfig:"oauth_client_secret"`
	OAuthRedirectURI  string `envconfig:"oauth_redirect_uri"`
}

var client *Client

func Initialize(config Config) {
	client = NewClient(config)
}

func LoginURL(state string) string { return client.LoginURL(state) }

func QueryOAuth(ctx context.Context, code string) (*OAuthToken, error) {
	return client.QueryOAuth(ctx, code)
}

func UserHasGroup(ctx context.Context, accessToken string, userID string, wantGroupID string) (bool, error) {
	return client.UserHasGroup(ctx, accessToken, userID, wantGroupID)
}

func (x Client) LoginURL(state string) string {
	query := make(url.Values)
	query.Set("client_id", x.config.OAuthClientID)
	query.Set("redirect_uri", x.config.OAuthRedirectURI)
	query.Set("state", state)
	query.Set("scope", "groups_access_member_info")
	return "https://www.facebook.com/v8.0/dialog/oauth?" + query.Encode()
}

type OAuthToken struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}

// Query oauth token from discord.
// We use the authorization code grant.
func (x Client) QueryOAuth(ctx context.Context, code string) (*OAuthToken, error) {
	if code == "" {
		return nil, ErrInvalidOAuthCode
	}

	query := make(url.Values)
	query.Set("client_id", x.config.OAuthClientID)
	query.Set("redirect_uri", x.config.OAuthRedirectURI)
	query.Set("client_secret", x.config.OAuthClientSecret)
	query.Set("code", code)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, x.config.Endpoint+"/oauth/access_token?"+query.Encode(), nil)
	if err != nil {
		return nil, fmt.Errorf("http.NewRequestWithContext() failed: %v", err)
	}

	var oauthToken OAuthToken
	if _, err := doFacebookRequest(req, &oauthToken); err != nil {
		return nil, err
	}
	return &oauthToken, nil
}

// Check if the user is in the given group.
// https://developers.facebook.com/docs/graph-api/reference/user/groups
func (x Client) UserHasGroup(ctx context.Context, accessToken string, userID string, wantGroupID string) (bool, error) {
	type Group struct {
		ID   string `json:"id"`
		Name string
	}

	type Page struct {
		Data json.RawMessage `json:"data"`
		Next string          `json:"next"`
	}

	for {
		query := make(url.Values)
		query.Set("fields", "id,name")
		query.Set("access_token", accessToken)
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, x.config.Endpoint+"/"+userID+"/groups?"+query.Encode(), nil)
		if err != nil {
			return false, fmt.Errorf("http.NewRequestWithContext() failed: %v", err)
		}

		var page Page
		if _, err := doFacebookRequest(req, &page); err != nil {
			return false, err
		}
		var groups []Group
		if err := json.Unmarshal(page.Data, &groups); err != nil {
			return false, err
		}

		for _, group := range groups {
			if group.ID == wantGroupID {
				return true, nil
			}
		}

		if page.Next == "" {
			return false, nil
		}
	}
}

// performs the http request and logs results
func doFacebookRequest(req *http.Request, dest interface{}) (*http.Response, error) {
	resp, err := http_wrappers.DoRequest(req)
	switch {
	case err != nil:
		logger.WithError(err).Error("http_wrappers.DoRequest() failed")

	case resp.StatusCode != http.StatusOK:
		err = ErrNon200Response
		var buf bytes.Buffer
		_, _ = buf.ReadFrom(resp.Body)

		logger.WithFields(logrus.Fields{
			"method": req.Method,
			"status": resp.StatusCode,
			"url":    req.URL.String(),
			"body":   buf.String(),
		}).Error("non-200 response from facebook")

	default:
		logger.WithFields(logrus.Fields{
			"method": req.Method,
			"status": resp.StatusCode,
			"url":    req.URL.String(),
		}).Info("facebook api request complete")
		err = json.NewDecoder(resp.Body).Decode(dest)
		if err != nil {
			logger.WithError(err).Error("json.Decode() failed")
		}
	}
	return resp, err
}
