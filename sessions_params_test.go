package requests

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/wangluozhe/requests/url"
)

func TestSessionPreservesLegacyQueryEncoding(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(r.RequestURI))
	}))
	defer server.Close()

	const value = "a%20b%26c%3d%E4%B8%AD%E6%96%87%2b%25"
	params := url.NewParams()
	params.Set("q", value)
	for _, test := range []struct {
		name          string
		rawQuery      string
		params        any
		sessionParams *url.Params
		wantQuery     string
	}{
		{
			name:      "url_only",
			rawQuery:  "q=" + value + "&plus=a+b&empty=",
			wantQuery: "q=" + value + "&plus=a+b&empty",
		},
		{
			name:      "params",
			params:    params,
			wantQuery: "q=" + value,
		},
		{
			name:      "map",
			params:    map[string]string{"q": value},
			wantQuery: "q=" + value,
		},
		{
			name:      "string",
			params:    "a%20b=" + value + "&q=one&q=two&empty=",
			wantQuery: "a%20b=" + value + "&q=one&q=two&empty",
		},
		{
			name:      "merge_url",
			rawQuery:  "base=x%20y%2f&plus=a+b",
			params:    params,
			wantQuery: "base=x%20y%2f&plus=a+b&q=" + value,
		},
		{
			name:          "merge_session",
			rawQuery:      "base=x%20y%2f",
			params:        params,
			sessionParams: url.ParseParams("plus=a+b&q=old%20value"),
			wantQuery:     "base=x%20y%2f&plus=a+b&q=" + value,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			session := NewSession()
			defer session.Close()
			session.Timeout = 5 * time.Second
			session.Params = test.sessionParams
			rawURL := server.URL + "/search"
			if test.rawQuery != "" {
				rawURL += "?" + test.rawQuery
			}
			response, err := session.Get(rawURL, &url.Request{Params: test.params})
			if err != nil {
				t.Fatal(err)
			}
			defer response.Close()
			if got, want := response.Text(), "/search?"+test.wantQuery; got != want {
				t.Fatalf("server received RequestURI %q, want %q", got, want)
			}
		})
	}
}
