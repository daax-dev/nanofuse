package main

import (
	"testing"

	"github.com/daax-dev/nanofuse/internal/client"
)

func TestParseMountSpecs(t *testing.T) {
	tests := []struct {
		name    string
		spec    string
		wantErr bool
		want    client.Mount
	}{
		{name: "shorthand bind", spec: "/host/data:/data", want: client.Mount{Source: "/host/data", Target: "/data", Type: "bind"}},
		{name: "shorthand ro", spec: "/host:/data:ro", want: client.Mount{Source: "/host", Target: "/data", Type: "bind", ReadOnly: true}},
		{name: "windows drive source", spec: `C:\data:/data:ro`, want: client.Mount{Source: `C:\data`, Target: "/data", Type: "bind", ReadOnly: true}},
		{name: "rw suffix", spec: "/host:/data:rw", want: client.Mount{Source: "/host", Target: "/data", Type: "bind"}},
		{name: "kv full", spec: "src=/h,dst=/g,type=bind,ro", want: client.Mount{Source: "/h", Target: "/g", Type: "bind", ReadOnly: true}},
		{name: "kv ro mixed case", spec: "src=/h,dst=/g,type=bind,ro=True", want: client.Mount{Source: "/h", Target: "/g", Type: "bind", ReadOnly: true}},
		{name: "kv ro false", spec: "src=/h,dst=/g,type=bind,ro=False", want: client.Mount{Source: "/h", Target: "/g", Type: "bind", ReadOnly: false}},
		{name: "kv ro invalid", spec: "src=/h,dst=/g,ro=maybe", wantErr: true},
		{name: "kv ro empty value", spec: "src=/h,dst=/g,ro=", wantErr: true},
		{name: "kv empty src value", spec: "src=,dst=/g", wantErr: true},
		{name: "kv bare key no value", spec: "src,dst=/g", wantErr: true},
		{name: "empty spec rejected", spec: "", wantErr: true},
		{name: "whitespace spec rejected", spec: "   ", wantErr: true},
		{name: "kv tmpfs", spec: "type=tmpfs,dst=/scratch", want: client.Mount{Target: "/scratch", Type: "tmpfs"}},
		{name: "unknown key", spec: "foo=bar", wantErr: true},
		{name: "bad shorthand", spec: "justone", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m, err := parseMountSpecs([]string{tt.spec})
			if tt.wantErr {
				if err == nil {
					t.Fatal("want error")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if len(m) != 1 || m[0] != tt.want {
				t.Fatalf("got %+v want %+v", m, tt.want)
			}
		})
	}
}

func TestParseSecretSpecs(t *testing.T) {
	t.Run("env", func(t *testing.T) {
		s, err := parseSecretSpecs([]string{"name=API_TOKEN,source=vault://kv/token"})
		if err != nil {
			t.Fatal(err)
		}
		if s[0].Name != "API_TOKEN" || s[0].Source != "vault://kv/token" {
			t.Fatalf("unexpected: %+v", s[0])
		}
	})
	t.Run("file", func(t *testing.T) {
		s, err := parseSecretSpecs([]string{"name=tls,type=file,target=/etc/tls/key.pem,source=spire://"})
		if err != nil {
			t.Fatal(err)
		}
		if s[0].Type != "file" || s[0].Target != "/etc/tls/key.pem" {
			t.Fatalf("unexpected: %+v", s[0])
		}
	})
	t.Run("value rejected", func(t *testing.T) {
		if _, err := parseSecretSpecs([]string{"name=x,value=hunter2"}); err == nil {
			t.Fatal("want error: values must not be accepted")
		}
	})
	t.Run("missing name", func(t *testing.T) {
		if _, err := parseSecretSpecs([]string{"source=vault://x"}); err == nil {
			t.Fatal("want error")
		}
	})
	t.Run("unknown key", func(t *testing.T) {
		if _, err := parseSecretSpecs([]string{"name=x,bogus=1"}); err == nil {
			t.Fatal("want error")
		}
	})
	t.Run("empty rejected", func(t *testing.T) {
		if _, err := parseSecretSpecs([]string{""}); err == nil {
			t.Fatal("want error for empty --secret")
		}
	})
}
