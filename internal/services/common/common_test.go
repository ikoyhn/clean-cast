package common

import (
	"testing"
	"time"
)

func TestParseLastBuildDate_RoundTripsAcrossZones(t *testing.T) {
	zones := []*time.Location{
		time.UTC,
		time.FixedZone("+0545", 5*3600+45*60),
		time.FixedZone("+0630", 6*3600+30*60),
		time.FixedZone("-0930", -(9*3600 + 30*60)),
		time.FixedZone("EDT", -4*3600),
	}

	for _, loc := range zones {
		now := time.Now().In(loc).Truncate(time.Second)

		parsed, err := ParseLastBuildDate(FormatLastBuildDate(now))
		if err != nil {
			t.Errorf("zone %s: round trip failed: %v", loc, err)
			continue
		}
		if !parsed.Equal(now) {
			t.Errorf("zone %s: got %v, want %v", loc, parsed, now)
		}
	}
}

func TestParseLastBuildDate_AcceptsLegacyRFC1123(t *testing.T) {
	legacy := "Mon, 18 Aug 2026 22:06:52 UTC"

	parsed, err := ParseLastBuildDate(legacy)
	if err != nil {
		t.Fatalf("expected legacy RFC1123 value to parse: %v", err)
	}

	want := time.Date(2026, time.August, 18, 22, 6, 52, 0, time.UTC)
	if !parsed.Equal(want) {
		t.Fatalf("got %v, want %v", parsed, want)
	}
}

func TestParseLastBuildDate_RejectsGarbage(t *testing.T) {
	if _, err := ParseLastBuildDate("not a date"); err == nil {
		t.Fatal("expected an error for an unparseable value")
	}
}
