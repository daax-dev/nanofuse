package main

import "testing"

func TestParseMountSpecs(t *testing.T) {
	t.Run("shorthand bind", func(t *testing.T) {
		m, err := parseMountSpecs([]string{"/host/data:/data"})
		if err != nil {
			t.Fatal(err)
		}
		if len(m) != 1 || m[0].Source != "/host/data" || m[0].Target != "/data" || m[0].Type != "bind" || m[0].ReadOnly {
			t.Fatalf("unexpected: %+v", m)
		}
	})
	t.Run("shorthand ro", func(t *testing.T) {
		m, err := parseMountSpecs([]string{"/host:/data:ro"})
		if err != nil {
			t.Fatal(err)
		}
		if !m[0].ReadOnly {
			t.Fatalf("want ro, got %+v", m[0])
		}
	})
	t.Run("shorthand windows drive source", func(t *testing.T) {
		m, err := parseMountSpecs([]string{`C:\data:/data:ro`})
		if err != nil {
			t.Fatal(err)
		}
		if m[0].Source != `C:\data` || m[0].Target != "/data" || !m[0].ReadOnly {
			t.Fatalf("unexpected: %+v", m[0])
		}
	})
	t.Run("shorthand rw suffix", func(t *testing.T) {
		m, err := parseMountSpecs([]string{"/host:/data:rw"})
		if err != nil {
			t.Fatal(err)
		}
		if m[0].Source != "/host" || m[0].Target != "/data" || m[0].ReadOnly {
			t.Fatalf("unexpected: %+v", m[0])
		}
	})
	t.Run("kv full", func(t *testing.T) {
		m, err := parseMountSpecs([]string{"src=/h,dst=/g,type=bind,ro"})
		if err != nil {
			t.Fatal(err)
		}
		if m[0].Source != "/h" || m[0].Target != "/g" || m[0].Type != "bind" || !m[0].ReadOnly {
			t.Fatalf("unexpected: %+v", m[0])
		}
	})
	t.Run("kv tmpfs", func(t *testing.T) {
		m, err := parseMountSpecs([]string{"type=tmpfs,dst=/scratch"})
		if err != nil {
			t.Fatal(err)
		}
		if m[0].Type != "tmpfs" || m[0].Target != "/scratch" {
			t.Fatalf("unexpected: %+v", m[0])
		}
	})
	t.Run("unknown key", func(t *testing.T) {
		if _, err := parseMountSpecs([]string{"foo=bar"}); err == nil {
			t.Fatal("want error")
		}
	})
	t.Run("bad shorthand", func(t *testing.T) {
		if _, err := parseMountSpecs([]string{"justone"}); err == nil {
			t.Fatal("want error")
		}
	})
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
}
