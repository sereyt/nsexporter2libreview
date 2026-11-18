package nightscout

import (
	"encoding/json"
	"testing"
	"time"
)

func TestGlucoseEntryTimestampFallbackForNightscout15(t *testing.T) {
	// Nightscout 15 does not return dateString/sysTime anymore.
	raw := `{
		"_id":"6af87112fe001431c08c6e69",
		"type":"sgv",
		"sgv":164,
		"date":1794666767667,
		"created_at":"2026-11-14T14:32:47.667Z",
		"direction":"Flat",
		"device":"Libre2",
		"app":"AAPS"
	}`

	entry := new(GlucoseEntry)
	if err := json.Unmarshal([]byte(raw), entry); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	want := time.UnixMilli(1794666767667)
	if got := entry.Timestamp(); !got.Equal(want) {
		t.Errorf("Timestamp() = %v, want %v", got, want)
	}
	if got := entry.FactoryTimestamp(); !got.Equal(want) {
		t.Errorf("FactoryTimestamp() = %v, want %v", got, want)
	}
	if entry.FactoryTimestamp().IsZero() {
		t.Error("FactoryTimestamp() must not be zero")
	}
}

func TestGlucoseEntryUsesSysTimeWhenPresent(t *testing.T) {
	raw := `{
		"sgv":120,
		"date":1794666767667,
		"sysTime":"2026-11-14T14:32:45Z",
		"dateString":"2026-11-14T17:32:47+03:00"
	}`

	entry := new(GlucoseEntry)
	if err := json.Unmarshal([]byte(raw), entry); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	want := time.Date(2026, 11, 14, 14, 32, 45, 0, time.UTC)
	if got := entry.FactoryTimestamp(); !got.Equal(want) {
		t.Errorf("FactoryTimestamp() = %v, want %v", got, want)
	}
}

func TestNSTimeAcceptsQuotedMilliseconds(t *testing.T) {
	var ts NSTime
	if err := json.Unmarshal([]byte(`"1794666767667"`), &ts); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	want := time.UnixMilli(1794666767667)
	if !ts.Time().Equal(want) {
		t.Errorf("Time() = %v, want %v", ts.Time(), want)
	}
}

func TestGlucoseEntriesDownsampleWithDateTime(t *testing.T) {
	first := time.UnixMilli(1794666767667)
	second := first.Add(5 * time.Minute)
	third := first.Add(10 * time.Minute)

	makeEntry := func(ts time.Time, sgv float64) *GlucoseEntry {
		t := NSTime(ts)
		return &GlucoseEntry{Date: &t, Sgv: SVG(sgv)}
	}

	entries := GlucoseEntries{
		makeEntry(third, 130),
		makeEntry(second, 120),
		makeEntry(first, 110),
	}

	got := entries.Downsample(DownsampleDuration(6 * time.Minute))
	if got.Len() != 2 {
		t.Fatalf("downsampled length = %d, want 2", got.Len())
	}
	if !(*got)[0].Timestamp().Equal(third) || !(*got)[1].Timestamp().Equal(first) {
		t.Errorf("unexpected downsampled timestamps: %v, %v", (*got)[0].Timestamp(), (*got)[1].Timestamp())
	}
}
