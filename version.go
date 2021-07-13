package king

import (
	"bytes"
	"errors"
	"fmt"
	"runtime"
	"strings"
	"text/template"
	"time"

	"github.com/alecthomas/kong"
)

const (
	versionKey      = "king_version"
	buildInfoKey    = "king_build_info"
	versionInfoTmpl = `
{{.Program}}, version {{.Version}} (revision: {{.Revision}})
  build date:       {{.Date}}
  go version:       {{.GoVersion}}
`
)

// VersionFlag displays the version information stored in "version" key form kong.Vars.
//
// Use this flag to show version information.
type VersionFlag bool

// BeforeApply is the actual version command.
func (v VersionFlag) BeforeApply(app *kong.Kong, vars kong.Vars) error {
	fmt.Fprintln(app.Stdout, vars[versionKey])
	app.Exit(0)

	return nil
}

// BuildInfo represents build information.
type BuildInfo struct {
	version   string
	revision  string
	goVersion string
	location  *time.Location
	date      time.Time
}

// NewBuildInfo creates BuildInformation from version, revision and date. These
// values are typically set with ldflags (via goreleaser for example).
//
// The date has to be in time.RFC3339 format and the revision must be
// at least 8 chars long.
func NewBuildInfo(version string, opts ...Option) (*BuildInfo, error) {
	b := BuildInfo{
		version:   version,
		location:  time.UTC,
		goVersion: runtime.Version(),
		date:      time.Date(1977, time.May, 25, 0, 0, 0, 0, time.UTC),
	}

	for _, opt := range opts {
		if err := opt(&b); err != nil {
			return nil, err
		}
	}

	return &b, nil
}

// Option is a BuildInfo functional option.
type Option func(*BuildInfo) error

// WithRevision sets the git commit revision.
func WithRevision(r string) Option {
	return func(b *BuildInfo) error {
		if len(r) < 8 {
			return errors.New("build revision must be at least 8 chars long")
		}

		b.revision = r[:8]

		return nil
	}
}

// WithDateString sets the build date (RFC3339).
func WithDateString(date string) Option {
	return func(b *BuildInfo) error {
		d, err := time.Parse(time.RFC3339, date)
		if err != nil {
			return err
		}

		b.date = d

		return nil
	}
}

// WithDate sets the build date.
func WithDate(date time.Time) Option {
	return func(b *BuildInfo) error {
		b.date = date

		return nil
	}
}

// WithLocation sets the timezone for the build date.
func WithLocation(loc string) Option {
	return func(b *BuildInfo) error {
		l, err := time.LoadLocation(loc)
		if err != nil {
			return err
		}

		b.location = l

		return nil
	}
}

// Version returns the version information.
func (b *BuildInfo) Version(program string) Version {
	return Version{
		Version:   b.version,
		Revision:  b.revision,
		Date:      b.date.In(b.location).Format(time.RFC3339),
		GoVersion: b.goVersion,
		Program:   program,
	}
}

// Version represents the version of a go program.
type Version struct {
	Version   string `json:"version" yaml:"version"`
	Revision  string `json:"revision" yaml:"revision"`
	Date      string `json:"date" yaml:"date"`
	GoVersion string `json:"go_version" yaml:"go_version"`
	Program   string `json:"program" yaml:"program"`
}

func (v Version) String() string {
	t := template.Must(template.New("version").Parse(versionInfoTmpl))

	var buf bytes.Buffer
	if err := t.ExecuteTemplate(&buf, "version", v); err != nil {
		panic(err)
	}

	return strings.TrimSpace(buf.String())
}

func (b *BuildInfo) asMap(prefix string) map[string]string {
	m := map[string]string{
		prefix + "buildinfo-version":  b.version,
		prefix + "buildinfo-date":     b.date.In(b.location).Format(time.RFC3339),
		prefix + "buildinfo-go":       b.goVersion,
		prefix + "buildinfo-revision": b.revision,
	}

	if prefix != "" {
		m[prefix+"buildinfo-location"] = b.location.String()
	}

	return m
}

func newBuildInfo(prefix string, m map[string]string) *BuildInfo {
	if _, ok := m[prefix+"buildinfo-version"]; !ok {
		return nil
	}

	d, _ := time.Parse(time.RFC3339, m[prefix+"buildinfo-date"])
	l, _ := time.LoadLocation(m[prefix+"buildinfo-location"])

	return &BuildInfo{
		version:   m[prefix+"buildinfo-version"],
		revision:  m[prefix+"buildinfo-revision"],
		goVersion: m[prefix+"buildinfo-go"],
		date:      d,
		location:  l,
	}
}
