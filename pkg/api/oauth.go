package api

import (
	"fmt"
	"golang.org/x/oauth2"
	"log"
	"net/http"
)

type OAuthHandler struct {
	ClientID     string `split_words:"true"`
	ClientSecret string `split_words:"true"`
	AuthURL      string `split_words:"true"`
	TokenURL     string `split_words:"true"`
}

func (x OAuthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	conf := &oauth2.Config{
		ClientID:     x.ClientID,
		ClientSecret: x.ClientSecret,
		Scopes:       []string{"identity"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  x.AuthURL,
			TokenURL: x.TokenURL,
		},
		RedirectURL: "http://localhost:8080",
	}

	// Redirect user to consent page to ask for permission
	// for the scopes specified above.
	url := conf.AuthCodeURL("state", oauth2.ApprovalForce)
	fmt.Printf("Visit the URL for the auth dialog: %v", url)

	// Use the authorization code that is pushed to the redirect
	// URL. Exchange will do the handshake to retrieve the
	// initial access token. The HTTP Client returned by
	// conf.Client will refresh the token as necessary.
	var code string
	if _, err := fmt.Scan(&code); err != nil {
		log.Fatal(err)
	}
	tok, err := conf.Exchange(ctx, code)
	if err != nil {
		log.Fatal(err)
	}

	client := conf.Client(ctx, tok)
	client.Get("...")

}
