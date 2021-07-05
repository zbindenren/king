package king

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
	"text/tabwriter"

	"github.com/alecthomas/kong"
)

const (
	configPathsKey = "king_config_paths"
)

// ShowConfig can be used to show information about the parsed configuration files.
type ShowConfig bool

// BeforeApply is the actual show-config command.
func (s ShowConfig) BeforeApply(app *kong.Kong, vars kong.Vars) error {
	fmt.Fprintln(app.Stderr, "Configuration files:")
	w := tabwriter.NewWriter(app.Stderr, 0, 0, 1, ' ', 0)

	for _, f := range Configs(vars) {
		file := kong.ExpandPath(f)

		f, err := os.Open(filepath.Clean(file))
		if err != nil {
			if os.IsNotExist(err) {
				fmt.Fprintf(w, "  %s\tnot found\n", file)
				continue
			}

			if os.IsPermission(err) {
				fmt.Fprintf(w, "  %s\tpermission denied\n", file)
				continue
			}

			return err
		}

		f.Close()
		fmt.Fprintf(w, "  %s\tparsed\n", file)
	}

	w.Flush()

	app.Exit(0)

	return nil
}

// Configs returns all configured absolute paths form kong.Vars.
func Configs(vars kong.Vars) []string {
	paths := []string{}
	for _, f := range strings.Split(vars[configPathsKey], ",") {
		paths = append(paths, kong.ExpandPath(f))
	}

	return paths
}

func configsForApp(name string) []string {
	return []string{
		"./" + name + ".yaml",
		path.Join("~/.config/", name, "config.yaml"),
		"/etc/" + name + "/config.yaml",
	}
}
