package nightscout

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/blutz1982/go-nsexporter-libreview/pkg/rest"
)

type GlucoseGetter interface {
	Glucose() GlucoseInterface
}

type GlucoseInterface interface {
	List(ctx context.Context, opts ListOptions) (*GlucoseEntries, error)
}

type glucose struct {
	client rest.Interface
}

// newGlucose returns a Glucos
func newGlucose(c Client) *glucose {
	return &glucose{
		client: c.RESTClient(),
	}
}

func (g glucose) List(ctx context.Context, opts ListOptions) (result *GlucoseEntries, err error) {
	result = &GlucoseEntries{}
	r := g.client.Get().
		Resource("entries").
		Name(opts.Kind).
		Param("find[date][$gte]", strconv.FormatInt(opts.DateFrom.UTC().UnixMilli(), 10)).
		Param("find[date][$lte]", strconv.FormatInt(opts.DateTo.UTC().UnixMilli(), 10)).
		Param("count", strconv.Itoa(opts.Count))

	err = r.Do(ctx).Into(result)
	if err != nil {
		return result, NewNightscoutError(err, "cant retreive list glucose entries")
	}

	return
}

type SVG float64

func (svg SVG) HighOutOfRange(max int) string {
	if int(svg) >= max {
		return "true"
	} else {
		return "false"
	}
}

func (svg SVG) Float64() float64 {
	return float64(svg)
}

func (svg SVG) MMol() float64 {
	return svg.Float64() / 18
}

func (svg SVG) LowOutOfRange(min int) string {
	if int(svg) <= min {
		return "true"
	} else {
		return "false"
	}
}

type NSTime time.Time

func (t *NSTime) Time() time.Time {
	if t == nil {
		return time.Time{}
	}
	return time.Time(*t)
}

func (t *NSTime) UnmarshalJSON(data []byte) error {
	value := strings.TrimSpace(string(data))
	if value == "" || value == "null" || value == `""` {
		*t = NSTime(time.Time{})
		return nil
	}

	// Nightscout normally sends `date` as milliseconds, but some
	// uploaders/older versions may send it as a quoted string or ISO date.
	value = strings.Trim(value, `"`)

	millis, err := strconv.ParseFloat(value, 64)
	if err == nil {
		*t = NSTime(time.UnixMilli(int64(millis)))
		return nil
	}

	for _, layout := range []string{time.RFC3339Nano, "2006-01-02T15:04:05.9999999Z07:00"} {
		ts, parseErr := time.Parse(layout, value)
		if parseErr == nil {
			*t = NSTime(ts)
			return nil
		}
	}

	return err
}

func (t *NSTime) MarshalJSON() ([]byte, error) {
	return []byte(strconv.FormatInt(t.Time().UnixMilli(), 10)), nil
}

func (t *NSTime) MarshalYAML() (interface{}, error) {
	return t.Time().UnixMilli(), nil
}

type GlucoseEntry struct {
	ID           string    `json:"_id"`
	App          string    `json:"app"`
	Device       string    `json:"device"`
	Date         *NSTime   `json:"date"`
	CreatedAt    string    `json:"created_at"`
	DateString   time.Time `json:"dateString"`
	Sgv          SVG       `json:"sgv"`
	Delta        float64   `json:"delta"`
	Direction    string    `json:"direction"`
	Type         string    `json:"type"`
	Filtered     float64   `json:"filtered"`
	Unfiltered   float64   `json:"unfiltered"`
	Units        string    `json:"units"`
	Rssi         int       `json:"rssi"`
	Noise        int       `json:"noise"`
	SysTime      time.Time `json:"sysTime"`
	DataType     int       `json:"dataType"`
	RecordNumber int       `json:"recordNumber"`
	UtcOffset    int       `json:"utcOffset"`
	Subject      string    `json:"subject"`
	Mills        int64     `json:"mills"`
}

func (g *GlucoseEntry) Kind() string {
	return "GlucoseEntry"
}

// Timestamp returns the time when the glucose value was measured.
// Nightscout 15 no longer returns dateString/sysTime, so `date` is the
// primary source. Older responses may only have dateString/created_at.
func (e *GlucoseEntry) Timestamp() time.Time {
	if e == nil {
		return time.Time{}
	}
	if e.Date != nil && !e.Date.Time().IsZero() {
		return e.Date.Time()
	}
	if !e.DateString.IsZero() {
		return e.DateString
	}
	if e.CreatedAt != "" {
		if ts, err := time.Parse(time.RFC3339Nano, e.CreatedAt); err == nil {
			return ts
		}
	}
	return time.Time{}
}

// FactoryTimestamp returns the sensor/factory time for LibreView.
// It falls back to the measurement date because Nightscout 15 entries
// do not contain `sysTime` anymore. Without this fallback LibreView gets
// a zero factory timestamp (0001-01-01) and silently ignores the readings.
func (e *GlucoseEntry) FactoryTimestamp() time.Time {
	if e == nil {
		return time.Time{}
	}
	if !e.SysTime.IsZero() {
		return e.SysTime
	}
	return e.Timestamp()
}

type GlucoseEntries []*GlucoseEntry

func (r *GlucoseEntries) Append(e *GlucoseEntry) {
	*r = append(*r, e)
}

func (r GlucoseEntries) Len() int {
	return len(r)
}

type GlucoseFilterFunc func(*GlucoseEntry) bool

func OnlyAfter(date time.Time) GlucoseFilterFunc {
	return func(e *GlucoseEntry) bool {
		return e.Timestamp().After(date)
	}
}

func (es *GlucoseEntries) Filter(fn GlucoseFilterFunc) (result *GlucoseEntries) {
	result = &GlucoseEntries{}
	es.Visit(func(e *GlucoseEntry, _ error) error {
		if fn(e) {
			result.Append(e)
		}
		return nil
	})
	return result
}

type DownsampleFunc func() time.Duration

func DownsampleDuration(d time.Duration) DownsampleFunc {
	return func() time.Duration {
		return d
	}
}

func (es *GlucoseEntries) Downsample(f DownsampleFunc) *GlucoseEntries {

	var lastTS time.Time

	return es.Filter(func(e *GlucoseEntry) bool {
		ts := e.Timestamp()
		if lastTS.IsZero() {
			lastTS = ts
			return true
		}

		if lastTS.Sub(ts) > f() {
			lastTS = ts
			return true
		}

		return false
	})

}

type VisitorFunc func(*GlucoseEntry, error) error

func (es GlucoseEntries) Visit(fn VisitorFunc) error {
	var err error
	for _, entry := range es {
		if err = fn(entry, err); err != nil {
			return err
		}
	}
	return nil
}
