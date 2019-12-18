// Package duration provides a partial implementation of ISO8601 durations.
// Constant values are assumed for non-constant timespans for convenience
// 1 Day = 24 Hours
// 1 Month = 30 Days
// 1 Year = 365 Days
package duration

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"regexp"
	"strconv"
	"text/template"
	"time"
)

const (
	// Zero is zero value for duration. ISO-8601 mandates at least one element for a zero duration (4.4.3.2.c - page 21)
	Zero = "P0D"
	// Day is an average day duration
	Day = time.Hour * 24
	// Week is an average week duration
	Week = Day * 7
	// Month is an average month duration
	Month = Day * 30
	// Year is an average year duration
	Year = Day * 365
)

var (
	// ErrBadFormat is returned when parsing fails
	ErrBadFormat = errors.New("invalid ISO8601 string")

	tmpl = template.Must(template.New("duration").Parse(`P{{if .IsZero}}0D{{else}}{{if and .Weeks .IsWeeksOnly}}{{.Weeks}}W{{else}}{{if .Years}}{{.Years}}Y{{end}}{{if .Days}}{{.Days}}D{{end}}{{if .HasTimePart}}T{{if .Hours}}{{.Hours}}H{{end}}{{if .Minutes}}{{.Minutes}}M{{end}}{{if .Seconds}}{{.Seconds}}S{{end}}{{end}}{{end}}{{end}}`))

	full = regexp.MustCompile(`P((?P<year>\d+)Y)?((?P<month>\d+)M)?((?P<day>\d+)D)?(T((?P<hour>\d+)H)?((?P<minute>\d+)M)?((?P<second>[\d\.]+)S)?)?`)
	week = regexp.MustCompile(`P((?P<week>\d+)W)`)
)

// Duration is ISO8601 type for your structs that will ensure correct validation and formatting per ISO8601 standard
type Duration struct {
	time.Duration
}

// Years returns duration in whole years
func (d *Duration) Years() float64 {
	x := float64(d.Duration / Year)
	return math.Floor(x)
}

// Weeks returns duration in whole weeks
func (d *Duration) Weeks() float64 {
	x := d.Days()
	y := float64(Week / Day)
	if math.Mod(x, y) > 0 {
		return 0
	}
	return x / y
}

// Days returns duration in whole days
func (d *Duration) Days() float64 {
	x := float64(d.Duration / Day)
	y := float64(Year / Day)
	return math.Mod(x, y)
}

// Hours returns duration in whole hours
func (d *Duration) Hours() float64 {
	x := float64(d.Duration / time.Hour)
	y := float64(Day / time.Hour)
	return math.Mod(x, y)
}

// Minutes returns duration in whole minutes
func (d *Duration) Minutes() float64 {
	x := float64(d.Duration / time.Minute)
	y := float64(time.Hour / time.Minute)
	return math.Mod(x, y)
}

// Seconds returns duration in whole seconds
func (d *Duration) Seconds() float64 {
	return (float64(d.Duration.Nanoseconds()) - math.Trunc(d.Duration.Minutes())*float64(time.Minute)) / float64(time.Second)
}

// ParseString parses string into ISO8601 duration
func ParseString(dur string) (*Duration, error) {
	var (
		match []string
		re    *regexp.Regexp
	)

	if week.MatchString(dur) {
		match = week.FindStringSubmatch(dur)
		re = week
	} else if full.MatchString(dur) {
		match = full.FindStringSubmatch(dur)
		re = full
	} else {
		return nil, ErrBadFormat
	}

	d := time.Duration(0)

	for i, name := range re.SubexpNames() {
		part := match[i]
		if i == 0 || name == "" || part == "" {
			continue
		}

		val, err := strconv.ParseFloat(part, 64)
		if err != nil {
			return nil, err
		}

		switch name {
		case "year":
			d += time.Duration(val) * Year
		case "month":
			d += time.Duration(val) * Month
		case "week":
			d += time.Duration(val) * Week
		case "day":
			d += time.Duration(val) * Day
		case "hour":
			d += time.Duration(val) * time.Hour
		case "minute":
			d += time.Duration(val) * time.Minute
		case "second":
			d += time.Duration(int(val)) * time.Second
			// handle fractional seconds
			val = (val - math.Trunc(val)) * 1000.0
			d += time.Duration(int(val)) * time.Millisecond
			val = (val - math.Trunc(val)) * 1000.0
			d += time.Duration(int(val)) * time.Microsecond
			val = (val - math.Trunc(val)) * 1000.0
			d += time.Duration(int(val)) * time.Nanosecond
		default:
			return nil, fmt.Errorf("unknown field %s", name)
		}
	}

	return &Duration{d}, nil
}

// String prints out the value passed in. It's not strictly according to the
// ISO spec, but it's pretty close. In particular, months are not returned.
// Instead, it returns a value in days (1D ~ 364D) or weeks (1W ~ 52W)
// whenever possible.
func (d *Duration) String() string {
	var s bytes.Buffer

	err := tmpl.Execute(&s, d)
	if err != nil {
		panic(err)
	}

	return s.String()
}

// IsZero checks whether duration is zero. Zero value is set to
func (d *Duration) IsZero() bool {
	return d.Duration == 0
}

// IsWeeksOnly checks whether duration in whole weeks
func (d *Duration) IsWeeksOnly() bool {
	return time.Duration(time.Duration(d.Weeks())*Week) == d.Duration
}

// HasTimePart checks if duration has time section
func (d *Duration) HasTimePart() bool {
	return d.Hours() != 0.0 || d.Minutes() != 0.0 || d.Seconds() != 0.0
}

// ToDuration converst ISO8001 duration into time.Duration
func (d *Duration) ToDuration() time.Duration {
	return d.Duration
}

// UnmarshalJSON validates json against Duration type
func (d *Duration) UnmarshalJSON(data []byte) error {
	b := bytes.NewBuffer(data)
	dec := json.NewDecoder(b)
	var s string
	if err := dec.Decode(&s); err != nil {
		return err
	}
	t, err := ParseString(s)
	if err != nil {
		return err
	}
	*d = *t
	return nil
}

// MarshalJSON converts Duration into json value
func (d Duration) MarshalJSON() ([]byte, error) {
	var b bytes.Buffer
	enc := json.NewEncoder(&b)
	s := d.String()
	enc.Encode(s)
	return b.Bytes(), nil
}
