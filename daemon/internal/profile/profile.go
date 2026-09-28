// Package profile defines the closed cohort schema. A profile is rejected when
// its signals contradict each other, so a composite lie cannot be activated.
package profile

import (
	"bytes"
	"encoding/json"
	"fmt"
	"regexp"
	"slices"
	"strings"
)

const DefaultID = "cohort-win11-chrome-en-us"

var (
	idPattern     = regexp.MustCompile(`^[a-z0-9][a-z0-9-]{0,63}$`)
	localePattern = regexp.MustCompile(`^[a-z]{2}(-[A-Z]{2})?$`)
	chromeMajor   = regexp.MustCompile(`Chrome/(\d+)\.`)
	brandVersion  = regexp.MustCompile(`^\d+$`)
	allowedCores  = []int{4, 8, 12, 16}
	allowedMemory = []int{4, 8, 16}
	allowedPolicy = "disable_non_proxied_udp"
	zoneOffsets   = map[string][]int{
		"UTC":                 {0},
		"Europe/London":       {-60, 0},
		"Europe/Berlin":       {-120, -60},
		"America/New_York":    {240, 300},
		"America/Chicago":     {300, 360},
		"America/Denver":      {360, 420},
		"America/Los_Angeles": {420, 480},
	}
)

// Profile is one coherent browser cohort. Fields are closed: unknown JSON
// keys are rejected.
type Profile struct {
	ID                    string        `json:"id"`
	UserAgent             string        `json:"userAgent"`
	UserAgentData         UserAgentData `json:"userAgentData"`
	Platform              string        `json:"platform"`
	Locale                string        `json:"locale"`
	Languages             []string      `json:"languages"`
	TimeZone              string        `json:"timeZone"`
	TimezoneOffsetMinutes int           `json:"timezoneOffsetMinutes"`
	Screen                Screen        `json:"screen"`
	HardwareConcurrency   int           `json:"hardwareConcurrency"`
	DeviceMemory          int           `json:"deviceMemory"`
	WebGL                 WebGL         `json:"webgl"`
	WebRTCPolicy          string        `json:"webRtcPolicy"`
	Fonts                 []string      `json:"fonts"`
}

// UserAgentData is the User-Agent Client Hints subset the cohort fixes.
type UserAgentData struct {
	Brands          []Brand `json:"brands"`
	Mobile          *bool   `json:"mobile"`
	Platform        string  `json:"platform"`
	PlatformVersion string  `json:"platformVersion"`
	Architecture    string  `json:"architecture"`
	Bitness         string  `json:"bitness"`
	Model           string  `json:"model"`
}

// Brand is one User-Agent Client Hints brand entry.
type Brand struct {
	Brand   string `json:"brand"`
	Version string `json:"version"`
}

// Screen is the cohort display. Values are CSS pixels and a device pixel ratio.
type Screen struct {
	Width            int     `json:"width"`
	Height           int     `json:"height"`
	DevicePixelRatio float64 `json:"devicePixelRatio"`
}

// WebGL is the renderer string paired to the claimed OS.
type WebGL struct {
	Vendor   string `json:"vendor"`
	Renderer string `json:"renderer"`
}

// Parse decodes one profile object and rejects unknown fields, trailing
// values, and internal contradictions.
func Parse(data []byte) (Profile, error) {
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	var p Profile
	if err := dec.Decode(&p); err != nil {
		return Profile{}, fmt.Errorf("invalid profile: %w", err)
	}
	if dec.More() {
		return Profile{}, fmt.Errorf("invalid profile: trailing data")
	}
	if err := p.Validate(); err != nil {
		return Profile{}, err
	}
	return p, nil
}

// Validate checks that the cohort agrees with itself.
func (p Profile) Validate() error {
	if !idPattern.MatchString(p.ID) {
		return fmt.Errorf("invalid profile: id is required")
	}
	if strings.TrimSpace(p.UserAgent) == "" {
		return fmt.Errorf("invalid profile: userAgent is required")
	}
	if strings.Contains(p.UserAgent, "Headless") {
		return fmt.Errorf("invalid profile: user agent contradicts platform")
	}
	if err := p.validatePlatform(); err != nil {
		return err
	}
	if !localePattern.MatchString(p.Locale) {
		return fmt.Errorf("invalid profile: locale is invalid")
	}
	if len(p.Languages) == 0 || p.Languages[0] != p.Locale {
		return fmt.Errorf("invalid profile: languages must start with locale")
	}
	for _, lang := range p.Languages {
		if !localePattern.MatchString(lang) {
			return fmt.Errorf("invalid profile: languages must start with locale")
		}
	}
	offsets, ok := zoneOffsets[p.TimeZone]
	if !ok || !slices.Contains(offsets, p.TimezoneOffsetMinutes) {
		return fmt.Errorf("invalid profile: time zone offset is not valid for timeZone")
	}
	if p.Screen.Width < 800 || p.Screen.Width > 7680 || p.Screen.Height < 600 || p.Screen.Height > 4320 {
		return fmt.Errorf("invalid profile: screen is invalid")
	}
	if p.Screen.DevicePixelRatio < 1 || p.Screen.DevicePixelRatio > 3 {
		return fmt.Errorf("invalid profile: screen is invalid")
	}
	if !slices.Contains(allowedCores, p.HardwareConcurrency) {
		return fmt.Errorf("invalid profile: hardwareConcurrency is not a common value")
	}
	if !slices.Contains(allowedMemory, p.DeviceMemory) {
		return fmt.Errorf("invalid profile: deviceMemory is not a common value")
	}
	if p.WebRTCPolicy != allowedPolicy {
		return fmt.Errorf("invalid profile: webRtcPolicy is invalid")
	}
	if len(p.Fonts) == 0 {
		return fmt.Errorf("invalid profile: fonts are required")
	}
	return nil
}

func (p Profile) validatePlatform() error {
	ua := p.UserAgent
	gl := p.WebGL.Renderer
	switch p.Platform {
	case "Win32":
		if !strings.Contains(ua, "Windows NT") || strings.Contains(ua, "Macintosh") || strings.Contains(ua, "Linux") {
			return fmt.Errorf("invalid profile: user agent contradicts platform")
		}
		if p.UserAgentData.Platform != "Windows" {
			return fmt.Errorf("invalid profile: user agent contradicts platform")
		}
		if strings.Contains(gl, "Apple") || strings.Contains(gl, "Mesa") || !strings.Contains(gl, "Direct3D") {
			return fmt.Errorf("invalid profile: webgl renderer contradicts platform")
		}
		if !slices.Contains(p.Fonts, "Segoe UI") {
			return fmt.Errorf("invalid profile: fonts contradict platform")
		}
	case "MacIntel":
		if !strings.Contains(ua, "Macintosh") || strings.Contains(ua, "Windows NT") {
			return fmt.Errorf("invalid profile: user agent contradicts platform")
		}
		if p.UserAgentData.Platform != "macOS" {
			return fmt.Errorf("invalid profile: user agent contradicts platform")
		}
		if strings.Contains(gl, "Direct3D") || !strings.Contains(gl, "Apple") {
			return fmt.Errorf("invalid profile: webgl renderer contradicts platform")
		}
		if !slices.Contains(p.Fonts, "Helvetica") {
			return fmt.Errorf("invalid profile: fonts contradict platform")
		}
	case "Linux x86_64":
		if !strings.Contains(ua, "Linux") || strings.Contains(ua, "Android") || strings.Contains(ua, "Windows NT") {
			return fmt.Errorf("invalid profile: user agent contradicts platform")
		}
		if p.UserAgentData.Platform != "Linux" {
			return fmt.Errorf("invalid profile: user agent contradicts platform")
		}
		if strings.Contains(gl, "Direct3D") || strings.Contains(gl, "Apple") {
			return fmt.Errorf("invalid profile: webgl renderer contradicts platform")
		}
		if !slices.Contains(p.Fonts, "DejaVu Sans") {
			return fmt.Errorf("invalid profile: fonts contradict platform")
		}
	default:
		return fmt.Errorf("invalid profile: platform is invalid")
	}
	if err := validateClientHints(p); err != nil {
		return err
	}
	if strings.TrimSpace(p.WebGL.Vendor) == "" {
		return fmt.Errorf("invalid profile: webgl renderer contradicts platform")
	}
	return nil
}

func validateClientHints(p Profile) error {
	uad := p.UserAgentData
	if uad.Mobile == nil || *uad.Mobile {
		return fmt.Errorf("invalid profile: user agent contradicts platform")
	}
	if uad.PlatformVersion == "" || (uad.Architecture != "x86" && uad.Architecture != "arm") {
		return fmt.Errorf("invalid profile: user agent contradicts platform")
	}
	if uad.Bitness != "64" && uad.Bitness != "32" {
		return fmt.Errorf("invalid profile: user agent contradicts platform")
	}
	if len(uad.Model) > 64 {
		return fmt.Errorf("invalid profile: user agent contradicts platform")
	}
	if len(uad.Brands) == 0 {
		return fmt.Errorf("invalid profile: user agent contradicts platform")
	}
	var chromeBrand string
	for _, b := range uad.Brands {
		if b.Brand == "" || !brandVersion.MatchString(b.Version) {
			return fmt.Errorf("invalid profile: user agent contradicts platform")
		}
		if b.Brand == "Google Chrome" {
			chromeBrand = b.Version
		}
	}
	if chromeBrand == "" {
		return fmt.Errorf("invalid profile: user agent contradicts platform")
	}
	m := chromeMajor.FindStringSubmatch(p.UserAgent)
	if m == nil || m[1] != chromeBrand {
		return fmt.Errorf("invalid profile: user agent contradicts platform")
	}
	return nil
}
