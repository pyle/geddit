// Copyright 2012 Jimmy Zelinskie. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package geddit

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path"
	"testing"
)

type RewriteTransport struct {
	Transport http.RoundTripper
	URL       *url.URL
}

func (t RewriteTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.URL.Scheme = t.URL.Scheme
	req.URL.Host = t.URL.Host
	req.URL.Path = path.Join(t.URL.Path, req.URL.Path)
	rt := t.Transport
	if rt == nil {
		rt = http.DefaultTransport
	}
	return rt.RoundTrip(req)
}

func testTools(code int, body string) (*httptest.Server, *OAuthSession) {
	// Dummy server to write JSON body provided
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(code)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, body)
	}))

	u, err := url.Parse(server.URL)
	if err != nil {
		log.Fatalf("Failed to parse local server URL: %v", err)
	}
	o := &OAuthSession{Client: http.DefaultClient, UserAgent: "Geddit Test"}
	o.Client.Transport = RewriteTransport{URL: u}

	return server, o
}

// Test defaults o fresh OAuthSession type.
func TestNewOAuthSession(t *testing.T) {
	// server := httptest.NewServer()
	o, err := NewOAuthSession("user", "pw", "agent", "http://", &http.Client{})
	if err != nil {
		t.Fatal(err)
	}

	if o.OAuthConfig.ClientID != "user" {
		t.Errorf("o.OAuthConfig.ClientID is incorrect, expected 'user', got %v", o.OAuthConfig.ClientID)
	}
	if o.OAuthConfig.ClientSecret != "pw" {
		t.Errorf("o.OAuthConfig.ClientSecret is incorrect, expected 'pw', got %v", o.OAuthConfig.ClientSecret)
	}
	if o.OAuthConfig.RedirectURL != "http://" {
		t.Errorf("o.OAuthConfig.RedirectURL is incorrect, expected 'http://', got %v", o.OAuthConfig.RedirectURL)
	}
	if o.UserAgent != "agent" {
		t.Errorf("o.UserAgent is incorrect, expected 'agent', got %v", o.UserAgent)
	}
	if o.Client == nil {
		t.Error("http client not created")
	}
}

func TestMe(t *testing.T) {
	server, oauth := testTools(200, `{"has_mail": false, "name": "aggrolite", "is_friend": false, "created": 1278447313.0, "suspension_expiration_utc": null, "hide_from_robots": true, "is_suspended": false, "modhash": "XXX", "created_utc": 1278418513.0, "link_karma": 2327, "comment_karma": 1233, "over_18": true, "is_gold": false, "is_mod": true, "id": "45xiz", "gold_expiration": null, "inbox_count": 0, "has_verified_email": true, "gold_creddits": 0, "has_mod_mail": false}`)
	defer server.Close()

	me, err := oauth.Me()
	if err != nil {
		t.Errorf("Me() Test failed: %v", err)
	}
	// Sanity check just a few fields?
	if me.Name != "aggrolite" {
		t.Fatalf("Me() returned unexpected name: %s", me.Name)
	}
	if me.ID != "45xiz" {
		t.Fatalf("Me() returned unexpected ID: %s", me.ID)
	}
	if me.String() != "aggrolite (2327-1233)" {
		t.Fatalf("Me.String() returns unexpected result: %s", me.String())
	}
	fmt.Println(me)

}
