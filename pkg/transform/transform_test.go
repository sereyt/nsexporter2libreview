package transform

import (
	"testing"
	"time"

	"github.com/blutz1982/go-nsexporter-libreview/pkg/nightscout"
)

func TestNSToLibreScheduledGlucoseEntryUsesDateAsFactoryTimestamp(t *testing.T) {
	measuredAt := time.UnixMilli(1794666767667)
	date := nightscout.NSTime(measuredAt)

	entry := &nightscout.GlucoseEntry{
		Date:      &date,
		Sgv:       nightscout.SVG(164),
		Direction: "Flat",
	}

	got := NSToLibreScheduledGlucoseEntry(entry)

	if got.ExtendedProperties.FactoryTimestamp.IsZero() {
		t.Fatal("factoryTimestamp must not be zero")
	}
	if !got.ExtendedProperties.FactoryTimestamp.Equal(measuredAt.UTC()) {
		t.Errorf("factoryTimestamp = %v, want %v", got.ExtendedProperties.FactoryTimestamp, measuredAt.UTC())
	}
	if !got.Timestamp.Equal(measuredAt) {
		t.Errorf("timestamp = %v, want %v", got.Timestamp, measuredAt)
	}
	if got.ValueInMgPerDl != 164 {
		t.Errorf("valueInMgPerDl = %v, want 164", got.ValueInMgPerDl)
	}
	if got.RecordNumber == 0 {
		t.Error("recordNumber must not be zero")
	}
}

func TestNSToLibreUnscheduledGlucoseEntryUsesDateAsFactoryTimestamp(t *testing.T) {
	measuredAt := time.UnixMilli(1794666767667)
	date := nightscout.NSTime(measuredAt)

	entry := &nightscout.GlucoseEntry{
		Date:      &date,
		Sgv:       nightscout.SVG(164),
		Direction: "FortyFiveUp",
	}

	got := NSToLibreUnscheduledGlucoseEntry(entry)

	if got.ExtendedProperties.FactoryTimestamp.IsZero() {
		t.Fatal("factoryTimestamp must not be zero")
	}
	if !got.ExtendedProperties.FactoryTimestamp.Equal(measuredAt.UTC()) {
		t.Errorf("factoryTimestamp = %v, want %v", got.ExtendedProperties.FactoryTimestamp, measuredAt.UTC())
	}
	if got.ExtendedProperties.TrendArrow != "Rising" {
		t.Errorf("trendArrow = %q, want Rising", got.ExtendedProperties.TrendArrow)
	}
	if got.RecordNumber == 0 {
		t.Error("recordNumber must not be zero")
	}
}

func TestNSToLibreGlucoseEntryPrefersSysTime(t *testing.T) {
	measuredAt := time.UnixMilli(1794666767667)
	sysTime := measuredAt.Add(-2 * time.Second)
	date := nightscout.NSTime(measuredAt)

	entry := &nightscout.GlucoseEntry{
		Date:    &date,
		SysTime: sysTime,
		Sgv:     nightscout.SVG(164),
	}

	got := NSToLibreScheduledGlucoseEntry(entry)

	if !got.ExtendedProperties.FactoryTimestamp.Equal(sysTime.UTC()) {
		t.Errorf("factoryTimestamp = %v, want %v", got.ExtendedProperties.FactoryTimestamp, sysTime.UTC())
	}
}
