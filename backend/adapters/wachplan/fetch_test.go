package wachplan

import (
	_ "embed"
	"strings"
	"testing"
)

//go:embed testdata/luss.json
var lussFixture []byte

//go:embed testdata/error.json
var errorFixture []byte

func TestParseResponse_Luss(t *testing.T) {
	t.Parallel()

	r, err := parseResponse(lussFixture)
	if err != nil {
		t.Fatalf("parseResponse: %v", err)
	}

	want := rawResponse{
		ID:           "172315",
		AppID:        "wassertemperatur-luss-see",
		DevID:        "eui-a840414281840166",
		TTNTimestamp: "2026-05-01T12:51:10.765868442Z",
		GtwID:        "wasserwacht-muenchen-west",
		GtwRSSI:      "-121",
		GtwSNR:       "-7",
		DevValue1:    "16.79",
		DevValue2:    "14.56",
		DevValue3:    "38.4",
		DevValue4:    "2.947",
	}
	if r != want {
		t.Errorf("parsed = %+v, want %+v", r, want)
	}
}

func TestParseResponse_UpstreamError(t *testing.T) {
	t.Parallel()

	_, err := parseResponse(errorFixture)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "bad sensor id") {
		t.Errorf("error = %v, want it to contain %q", err, "bad sensor id")
	}
}

func TestParseFloatPtr(t *testing.T) {
	t.Parallel()

	cases := []struct {
		in     string
		wantOK bool
		want   float64
	}{
		{"14.56", true, 14.56},
		{"-7", true, -7},
		{"0", true, 0},
		{"", false, 0},
		{"not a number", false, 0},
	}
	for _, c := range cases {
		got := parseFloatPtr(c.in)
		if c.wantOK {
			if got == nil {
				t.Errorf("parseFloatPtr(%q) = nil, want %v", c.in, c.want)
				continue
			}
			if *got != c.want {
				t.Errorf("parseFloatPtr(%q) = %v, want %v", c.in, *got, c.want)
			}
		} else if got != nil {
			t.Errorf("parseFloatPtr(%q) = %v, want nil", c.in, *got)
		}
	}
}

func TestParseIntPtr32(t *testing.T) {
	t.Parallel()

	cases := []struct {
		in     string
		wantOK bool
		want   int32
	}{
		{"-121", true, -121},
		{"0", true, 0},
		{"", false, 0},
		{"3.14", false, 0},
	}
	for _, c := range cases {
		got := parseIntPtr32(c.in)
		if c.wantOK {
			if got == nil {
				t.Errorf("parseIntPtr32(%q) = nil, want %v", c.in, c.want)
				continue
			}
			if *got != c.want {
				t.Errorf("parseIntPtr32(%q) = %v, want %v", c.in, *got, c.want)
			}
		} else if got != nil {
			t.Errorf("parseIntPtr32(%q) = %v, want nil", c.in, *got)
		}
	}
}
