package king

import (
	"fmt"
	"os"
	"path"
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

	for _, file := range Configs(vars) {
		filePath := kong.ExpandPath(file)
		found := "parsed"

		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			found = "not found"
		}

		fmt.Fprintf(w, "  %s\t%v\n", filePath, found)
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
