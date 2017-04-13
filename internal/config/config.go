package config

import (
	"io/ioutil"
	"os"

	"github.com/naoina/toml"
	"github.com/naoina/toml/ast"
)

func Decode(path string) *ast.Table {
	var (
		err   error
		raw   []byte
		table *ast.Table
	)

	if raw, err = ioutil.ReadFile(path); err != nil {
		panic(err)
	}

	raw = []byte(os.ExpandEnv(string(raw)))

	if table, err = toml.Parse(raw); err != nil {
		panic(err)
	}

	return table
}
