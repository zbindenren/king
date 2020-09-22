package king

import (
	"os"
	"strings"

	"github.com/alecthomas/kong"
)

var (
	ignoredFlagsNames map[string]bool = map[string]bool{
		"help":     true,
		"env-help": true,
	}
)

// EnvResolver returns a Resolver that retrieves values from environment variables.
//
// Hyphens in flag names are replaced with underscores.
// Flag names are prefixed with app name and converted to uppercase.
//
//  Usage:
//  ctx := kong.Parse(&cli,
//      kong.Resolvers(pfkong.EnvResolver()),
//      )
//  }
func EnvResolver() kong.Resolver {
	var f kong.ResolverFunc = func(context *kong.Context, parent *kong.Path, flag *kong.Flag) (interface{}, error) {
		if ok := ignoredFlagsNames[flag.Name]; ok {
			return nil, nil
		}
		raw, ok := os.LookupEnv(toEnvVarName(context.Model.Name, flag.Value))
		if !ok {
			return nil, nil
		}
		return raw, nil
	}

	return f
}

func toEnvVarName(prefix string, value *kong.Value) string {
	if prefix != "" {
		prefix += "_"
	}

	return strings.ToUpper(prefix + strings.ReplaceAll(value.Name, "-", "_"))
}
