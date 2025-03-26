package king_test

import (
	"fmt"
	"os"
	"regexp"
	"runtime"
	"sort"
	"strings"
	"testing"

	"github.com/alecthomas/kong"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zbindenren/king"
)

type envMap map[string]string

func tempEnv(env envMap) func() {
	for k, v := range env {
		os.Setenv(k, v)
	}

	return func() {
		for k := range env {
			os.Unsetenv(k)
		}
	}
}

func writeFile(t *testing.T, data []byte) (filePath string, cleanup func()) {
	tmpfile, err := os.CreateTemp("", "test")
	if err != nil {
		t.Error(t, err)
	}

	if _, err := tmpfile.Write(data); err != nil {
		t.Error(t, err)
	}

	return tmpfile.Name(), func() {
		tmpfile.Close()

		if err := os.Remove(tmpfile.Name()); err != nil {
			t.Error(err)
		}
	}
}

type cli struct {
	FromFlag        string `help:"Value from flag."`
	FromAutoEnv     string `help:"From auto env."`
	FromConfig      string `help:"From config."`
	OverrideEnv     string `help:"Override env."`
	OverrideConfig  string `help:"Override config."`
	OverrideAutoEnv string `help:"Override config." env:"ENV"`
}

func TestYAMLAndAutoEnvResolvers(t *testing.T) {
	autoEnvValue := "fromAutoEnv"
	envValue := "fromEnv"
	flagValue := "fromFlag"
	cfgValue := "fromConfig"
	cleanup := tempEnv(envMap{
		"TEST_FROM_AUTO_ENV": autoEnvValue,
		"TEST_OVERRIDE_ENV":  autoEnvValue,
		"ENV":                envValue,
	})

	defer cleanup()

	path, cleanUpFile := writeFile(t, fmt.Appendf(nil, `---
from-config: "%s"
override-config: "%s"
`, cfgValue, cfgValue),
	)
	defer cleanUpFile()

	expected := cli{
		FromAutoEnv:     autoEnvValue,
		FromConfig:      cfgValue,
		OverrideEnv:     flagValue,
		OverrideConfig:  flagValue,
		OverrideAutoEnv: envValue,
	}

	c := cli{}
	buf := &strings.Builder{}
	opts := king.DefaultOptions(
		king.Config{
			Name:        "test",
			Description: "A application to test.",
			ConfigPaths: []string{path},
		},
	)
	opts = append(opts, kong.Writers(buf, buf))
	parser, err := kong.New(&c, opts...)
	require.NoError(t, err)
	_, err = parser.Parse([]string{
		"--override-env=" + flagValue,
		"--override-config=" + flagValue,
	})
	require.NoError(t, err)
	assert.Equal(t, expected, c)
}

func TestTOMLResolver(t *testing.T) {
	cfgValue := "fromConfig"
	path, cleanUpFile := writeFile(t, fmt.Appendf(nil, `from-config="%s"
override-config="%s"
`, cfgValue, cfgValue),
	)
	defer cleanUpFile()

	expected := cli{
		FromConfig:     cfgValue,
		OverrideConfig: cfgValue,
	}

	cli := cli{}
	buf := &strings.Builder{}
	opts := king.DefaultOptions(
		king.Config{
			Name:         "test",
			Description:  "A application to test.",
			ConfigPaths:  []string{path},
			FileResolver: king.TOML,
		},
	)
	opts = append(opts, kong.Writers(buf, buf))
	parser, err := kong.New(&cli, opts...)
	require.NoError(t, err)
	_, err = parser.Parse([]string{})
	require.NoError(t, err)
	assert.Equal(t, expected, cli)
}

func TestVersion(t *testing.T) {
	buf := &strings.Builder{}
	b, err := king.NewBuildInfo("1.0.0",
		king.WithDateString("2020-09-22T11:11:10+02:00"),
		king.WithRevision("12345678"),
		king.WithLocation("Europe/Zurich"),
	)

	require.NoError(t, err)

	opts := king.DefaultOptions(
		king.Config{
			Name:        "test",
			Description: "A application to test.",
			BuildInfo:   b,
		},
	)
	opts = append(opts, kong.Writers(buf, buf))
	cli := struct {
		Version king.VersionFlag `help:"Show version."`
	}{}
	parser, err := kong.New(&cli, opts...)
	require.NoError(t, err)

	parser.Exit = func(int) {}
	_, err = parser.Parse([]string{"--version"})
	require.NoError(t, err)

	expected := fmt.Sprintf(`test, version 1.0.0 (revision: 12345678)
  build date:       2020-09-22T11:11:10+02:00
  go version:       %s
`, runtime.Version())

	assert.Equal(t, expected, buf.String())
}

func TestHelp(t *testing.T) {
	buf := &strings.Builder{}
	opts := king.DefaultOptions(
		king.Config{
			Name:        "test",
			Description: "A application to test.",
		},
	)
	opts = append(opts, kong.Writers(buf, buf))
	c := cli{}

	parser, err := kong.New(&c, opts...)
	require.NoError(t, err)

	parser.Exit = func(int) {}
	_, err = parser.Parse([]string{"--help"})
	require.NoError(t, err)

	expected := `Usage: test [flags]

A application to test.

Flags:
  -h, --help                      Show context-sensitive help ($TEST_HELP).
      --from-flag=STRING          Value from flag ($TEST_FROM_FLAG).
      --from-auto-env=STRING      From auto env ($TEST_FROM_AUTO_ENV).
      --from-config=STRING        From config ($TEST_FROM_CONFIG).
      --override-env=STRING       Override env ($TEST_OVERRIDE_ENV).
      --override-config=STRING    Override config ($TEST_OVERRIDE_CONFIG).
      --override-auto-env=STRING
                                  Override config ($ENV).
`
	assert.Equal(t, expected, buf.String())
}

func TestFlagMap(t *testing.T) {
	opts := king.DefaultOptions(
		king.Config{
			Name:        "test",
			Description: "A application to test.",
		},
	)
	buf := &strings.Builder{}
	opts = append(opts, kong.Writers(buf, buf))
	c := cli{}

	parser, err := kong.New(&c, opts...)
	require.NoError(t, err)

	parser.Exit = func(int) {}

	ctx, err := parser.Parse([]string{"--help", "--override-config=will_be_redacted"})
	require.NoError(t, err)

	l := king.FlagMap(ctx, regexp.MustCompile("override-config")).Rm("override-auto-env", "help").Add("version", "1.0", "commit", "123456789", "not_added").List()
	expected := []interface{}{
		"commit",
		"123456789",
		"from-auto-env",
		"",
		"from-config",
		"",
		"from-flag",
		"",
		"override-config",
		"****************",
		"override-env",
		"",
		"version",
		"1.0",
	}
	assert.Equal(t, expected, l)

	reg := prometheus.NewRegistry()
	king.FlagMap(ctx, regexp.MustCompile("override-config")).Add("bool-flag", "true").Rm("help").Register("program", reg)

	a, err := reg.Gather()
	require.NoError(t, err)

	labels := []string{}

	for _, m := range a {
		for _, b := range m.GetMetric() {
			for _, label := range b.GetLabel() {
				labels = append(labels, strings.TrimSpace(label.String()))
			}
		}
	}

	expectedLabels := []string{
		`name:"name"  value:"bool-flag"`,
		`name:"program"  value:"program"`,
		`name:"value"  value:"true"`,
	}
	sort.Strings(expectedLabels)
	sort.Strings(labels)

	assert.Equal(t, expectedLabels, labels)
}
