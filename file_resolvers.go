package king

import (
	"io"
	"io/ioutil"

	"github.com/BurntSushi/toml"
	"github.com/alecthomas/kong"
	"gopkg.in/yaml.v3"
)

// FileResolver represents a kong fileresolver.
type FileResolver string

// All supported file resolvers.
const (
	YAML FileResolver = "yaml"
	TOML FileResolver = "toml"
)

// NewFileResolver creates a new fileresolver.
func NewFileResolver(f FileResolver) kong.ConfigurationLoader {
	switch f {
	case TOML:
		return tomlResolver
	default:
		return yamlResolver
	}
}

func yamlResolver(r io.Reader) (kong.Resolver, error) {
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

func tomlResolver(r io.Reader) (kong.Resolver, error) {
	values := map[string]interface{}{}

	data, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}

	if err := toml.Unmarshal(data, &values); err != nil {
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
