// Package king is a library to configure the command line parser
// https://github.com/alecthomas/kong
package king

import (
	"context"
	"regexp"
	"sort"
	"strings"

	"github.com/alecthomas/kong"
)

const redactChar = `*`

// Config is used to create DefaultOptions.
type Config struct {
	Context     context.Context
	Name        string
	Description string
	BuildInfo   *BuildInfo
	ConfigPaths []string
}

func (c Config) pathString() string {
	return strings.Join(c.ConfigPaths, ",")
}

// DefaultOptions creates a set of opinionated options.
func DefaultOptions(c Config) []kong.Option {
	if c.ConfigPaths == nil {
		c.ConfigPaths = configsForApp(c.Name)
	}

	vars := kong.Vars{
		configPathsKey: c.pathString(),
	}

	if c.BuildInfo != nil {
		vars[versionKey] = c.BuildInfo.Version(c.Name)

		for k, v := range c.BuildInfo.asMap("king_") {
			vars[k] = v
		}
	}

	opts := []kong.Option{
		kong.Name(c.Name),
		kong.Description(c.Description),
		kong.HelpFormatter(newHelpFormatter(c.Name)),
		kong.ConfigureHelp(kong.HelpOptions{
			Compact: true,
		}),
		kong.UsageOnError(),
		kong.Resolvers(EnvResolver()),
		kong.Configuration(YAML, c.ConfigPaths...),
		vars,
	}

	if c.Context != nil {
		opts = append(opts, bindContext(c.Context))
	}

	return opts
}

// Map is a map with string as key and interface{} as value.
type Map map[string]interface{}

// FlagMap returns the flags and corresponding values from *kong.Context.
//
// To prevent logging sensitive flag values it is possible to provide
// a list of regular expressions. Flag values of flag names that match are
// redacted by '*'.
func FlagMap(ctx *kong.Context, redactFlags ...*regexp.Regexp) Map {
	m := Map{}
	for _, f := range ctx.Flags() {
		m[f.Name] = ctx.FlagValue(f)
	}

	m = m.redact(redactFlags...)

	b := newBuildInfo("king_", ctx.Model.Vars())
	if b != nil {
		m[buildInfoKey] = b
		for k, v := range b.asMap("") {
			m[k] = v
		}
	}

	return m
}

func (m Map) redact(keyRegexp ...*regexp.Regexp) Map {

	r := redactor(keyRegexp)
	nm := Map{}

	for k, v := range m {
		nm[k] = r(k, v)
	}

	return nm
}

// Add adds key and values.
func (m Map) Add(keyVals ...string) Map {
	nm := Map{}

	for k, v := range m {
		nm[k] = v
	}

	max := len(keyVals)
	if max%2 != 0 {
		max--
	}

	for i := 0; i < max; i += 2 {
		nm[keyVals[i]] = keyVals[i+1]
	}

	return nm
}

// Rm removes keys from Map.
func (m Map) Rm(keys ...string) Map {
	nm := Map{}

	for k, v := range m {
		if contains(keys, k) {
			continue
		}

		nm[k] = v
	}

	return nm
}

// List returns the flag and values as list (sorted by keys).
func (m Map) List() []interface{} {
	l := make([]interface{}, 0, 2*len(m))

	for _, k := range m.keys() {
		l = append(l, k, m[k])
	}

	return l
}

func (m Map) keys() []string {
	keys := make([]string, 0, len(m))

	for k := range m {
		if k == buildInfoKey {
			continue
		}

		keys = append(keys, k)
	}

	sort.Strings(keys)

	return keys
}

func bindContext(ctx context.Context) kong.Option {
	return kong.BindTo(ctx, (*context.Context)(nil))
}

func redactor(targets []*regexp.Regexp) func(string, interface{}) interface{} {
	return func(key string, value interface{}) interface{} {
		s, ok := value.(string)
		if !ok {
			return value
		}

		for _, t := range targets {
			if t.MatchString(strings.ToLower(key)) {
				return strings.Repeat(string(redactChar), len(s))
			}
		}

		return value
	}
}

func contains(list []string, item string) bool {
	for _, itm := range list {
		if itm == item {
			return true
		}
	}

	return false
}

func newHelpFormatter(appName string) func(*kong.Value) string {
	return func(value *kong.Value) string {
		suffix := "($" + value.Tag.Env + ")"

		if value.Tag.Env == "" {
			envName := toEnvVarName(appName, value)
			suffix = "($" + envName + ")"
		}

		switch {
		case strings.HasSuffix(value.Help, "."):
			return value.Help[:len(value.Help)-1] + " " + suffix + "."
		case value.Help == "":
			return suffix
		default:
			return value.Help + " " + suffix
		}
	}
}
