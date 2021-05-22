package dumper

import (
	"fmt"
	xmodfile "golang.org/x/mod/modfile"
	"golang.org/x/mod/module"
	"io/ioutil"
	"os"
	"path/filepath"
)

type Modfile interface {
	GetModules() (dependencies []module.Version)
}

type modfile struct {
	*xmodfile.File
}

type Module struct {
	Path string
	Version string
}

func New(path string) (*modfile, error) {
	path = filepath.Join(path, "go.mod")
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open module file: %w", err)
	}
	defer file.Close()

	moduleBytes, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("read module file: %w", err)
	}

	moduleFile, err := xmodfile.Parse(path, moduleBytes, nil)
	if err != nil {
		return nil, fmt.Errorf("parse module file: %w", err)
	}

	if moduleFile.Module == nil {
		return nil, fmt.Errorf("parsing module returned nil module")
	}

	return &modfile{moduleFile}, nil
}

// GetModules gets the dependencies in the go.mod file
func (mf *modfile) GetModules() (dependencies []Module) {
	for _, dependency := range mf.File.Require {
		dependencies = append(dependencies, Module{dependency.Mod.Path, dependency.Mod.Version})
	}

	return
}
