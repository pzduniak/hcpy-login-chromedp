package main

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/chromedp/cdproto/target"
	"github.com/chromedp/chromedp"
	"github.com/dchest/uniuri"
)

// performInteractiveAuth completes an OAuth dance using an interactive Chrome session
func performInteractiveAuth(ctx context.Context) (string, error) {
	allocCtx, cancel := chromedp.NewExecAllocator(ctx, chromedp.Flag("headless", false))
	defer cancel()

	chromeCtx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	// prepare pkce challenge
	verifier := uniuri.NewLen(32)
	verifierChallengeRaw := sha256.Sum256([]byte(verifier))
	verifierChallenge := base64.RawURLEncoding.EncodeToString(verifierChallengeRaw[:])

	// nonce and state
	nonce := uniuri.NewLen(16)
	state := uniuri.NewLen(16)

	// assemble the request
	loginQuery := url.Values{
		"response_type":         {"code"},
		"prompt":                {"login"},
		"code_challenge":        {verifierChallenge},
		"code_challenge_method": {"S256"},
		"client_id":             {hcAppID},
		"scope":                 {strings.Join(hcScopes, " ")},
		"nonce":                 {nonce},
		"state":                 {state},
		"redirect_uri":          {hcRedirectURI},
	}
	loginPageURL := hcAPIBaseURLs[selectedRegion] + "authorize?" + loginQuery.Encode()

	authCodeCh := make(chan string)
	chromedp.ListenBrowser(chromeCtx, func(ev interface{}) {
		switch e := ev.(type) {
		case *target.EventTargetInfoChanged:
			// only match HC redirect URLs
			if !hcRedirectURLRE.Match([]byte(e.TargetInfo.URL)) {
				return
			}
			parsedURL, err := url.Parse(e.TargetInfo.URL)
			if err != nil {
				panic(err)
			}

			code := parsedURL.Query().Get("code")
			if code == "" {
				panic("Code is missing from URL: " + e.TargetInfo.URL)
			}
			authCodeCh <- code
		}
	})

	if err := chromedp.Run(
		chromeCtx,
		chromedp.Navigate(loginPageURL),
	); err != nil {
		panic(err)
	}

	// block until we get a parsed oauth code
	responseCode := <-authCodeCh

	// shut down chrome, since we're done with the interactive session
	tctx, tcancel := context.WithTimeout(chromeCtx, 10*time.Second)
	defer tcancel()
	if err := chromedp.Cancel(tctx); err != nil {
		log.Printf("Failed to shut down Chrome: %v", err)
	}

	// complete the oauth flow
	tokenValues := url.Values{
		"grant_type":    {"authorization_code"},
		"client_id":     {hcAppID},
		"code_verifier": {verifier},
		"code":          {responseCode},
		"redirect_uri":  {hcRedirectURI},
	}
	resp, err := http.PostForm(hcAPIBaseURLs[selectedRegion]+"token", tokenValues)
	if err != nil {
		panic(err)
	}
	var authResponse struct {
		AccessToken  string `json:"access_token"`
		ExpiresIn    int    `json:"expires_in"`
		IDToken      string `json:"id_token"`
		RefreshToken string `json:"refresh_token"`
		Scope        string `json:"scope"`
		TokenType    string `json:"token_type"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&authResponse); err != nil {
		panic(err)
	}
	if authResponse.AccessToken == "" {
		panic("Failed to retrieve an access token")
	}

	return authResponse.AccessToken, nil
}
