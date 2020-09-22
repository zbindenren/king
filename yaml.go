package king

import (
	"io"

	"github.com/alecthomas/kong"
	"gopkg.in/yaml.v3"
)

// YAML returns a Resolver that retrieves values from a YAML source.
func YAML(r io.Reader) (kong.Resolver, error) {
	values := map[string]interface{}{}

	err := yaml.NewDecoder(r).Decode(&values)
	if err != nil {
		return nil, err
	}

	var f kong.ResolverFunc = func(ctx *kong.Context, parent *kong.Path, flag *kong.Flag) (interface{}, error) {
		raw, ok := values[flag.Name]
		if !ok {
			return nil, nil
		}
		return raw, nil
	}

	return f, nil
}
