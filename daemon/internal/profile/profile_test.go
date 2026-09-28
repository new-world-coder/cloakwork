package profile_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/cloakwork/cloakwork/daemon/internal/profile"
)

func TestEmbeddedCohortIsCoherent(t *testing.T) {
	store, err := profile.LoadEmbedded()
	if err != nil {
		t.Fatal(err)
	}
	got := store.Active()
	if got.ID != profile.DefaultID {
		t.Fatalf("active id = %s", got.ID)
	}
	if got.Platform != "Win32" || got.TimeZone != "America/New_York" || got.TimezoneOffsetMinutes != 300 {
		t.Fatalf("reference cohort drifted: %+v", got)
	}
	ids := store.IDs()
	if len(ids) != 1 || ids[0] != profile.DefaultID {
		t.Fatalf("ids = %v", ids)
	}
}

func TestSchemaRejection(t *testing.T) {
	base := active(t)
	tests := []struct {
		name    string
		mutate  func(*profile.Profile)
		raw     func([]byte) []byte
		wantErr string
	}{
		{
			name: "unknown field",
			raw: func(b []byte) []byte {
				var m map[string]json.RawMessage
				if err := json.Unmarshal(b, &m); err != nil {
					t.Fatal(err)
				}
				m["canvasNoise"] = []byte("true")
				out, err := json.Marshal(m)
				if err != nil {
					t.Fatal(err)
				}
				return out
			},
			wantErr: "unknown field",
		},
		{
			name: "windows ua with apple gpu",
			mutate: func(p *profile.Profile) {
				p.WebGL.Renderer = "ANGLE (Apple, Apple M2, OpenGL 4.1)"
			},
			wantErr: "webgl renderer contradicts platform",
		},
		{
			name: "time zone offset disagrees",
			mutate: func(p *profile.Profile) {
				p.TimezoneOffsetMinutes = 1
			},
			wantErr: "time zone offset is not valid",
		},
		{
			name: "language list disagrees with locale",
			mutate: func(p *profile.Profile) {
				p.Languages = []string{"fr-FR"}
			},
			wantErr: "languages must start with locale",
		},
		{
			name: "headless user agent",
			mutate: func(p *profile.Profile) {
				p.UserAgent = strings.Replace(p.UserAgent, "Chrome/", "HeadlessChrome/", 1)
			},
			wantErr: "user agent contradicts platform",
		},
		{
			name: "missing user agent",
			mutate: func(p *profile.Profile) {
				p.UserAgent = ""
			},
			wantErr: "userAgent is required",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, err := json.Marshal(base)
			if err != nil {
				t.Fatal(err)
			}
			if tt.raw != nil {
				body = tt.raw(body)
			} else {
				p := base
				tt.mutate(&p)
				body, err = json.Marshal(p)
				if err != nil {
					t.Fatal(err)
				}
			}
			_, err = profile.Parse(body)
			if err == nil {
				t.Fatal("expected rejection")
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("error %q does not contain %q", err.Error(), tt.wantErr)
			}
		})
	}
}

func TestUnknownProfile(t *testing.T) {
	store, err := profile.LoadEmbedded()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.Get("missing"); err == nil {
		t.Fatal("expected unknown profile")
	}
}

func active(t *testing.T) profile.Profile {
	t.Helper()
	store, err := profile.LoadEmbedded()
	if err != nil {
		t.Fatal(err)
	}
	return store.Active()
}
