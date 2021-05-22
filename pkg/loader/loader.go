package loader

import (
	"context"
	"fmt"
	"github.com/gomods/athens/pkg/errors"
	"github.com/gomods/athens/pkg/log"
	"github.com/gomods/athens/pkg/storage"
	"github.com/spf13/afero"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

type moduleInfo struct {
	version string
	src     string
}

func LoadModules(ctx context.Context, srcDir string, store storage.Backend, lgger *log.Logger) error {
	const op errors.Op = "loader.LoadModules"

	Fs := afero.NewOsFs()
	tempDir, err := afero.TempDir(Fs, "", "athens_load")
	if err != nil {
		return errors.E(op, err)
	}
	defer Fs.RemoveAll(tempDir)

	dumps, err := unpackDumps(srcDir, tempDir)
	if err != nil {
		return errors.E(op, err)
	}

	modules := make(map[string]moduleInfo)

	for _, dump := range dumps {
		dumpDir := filepath.Join(tempDir, dump)
		_, err := listDirs(dumpDir)
		if err != nil {
			return errors.E(op, "vcs:listDirs failed", err)
		}

		filepath.Walk(dumpDir, func(path string, info fs.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if ok, _ := isAllFiles(path); ok {
				module := strings.TrimSuffix(path, info.Name())
				module = strings.TrimPrefix(module, dumpDir)
				module = strings.Trim(module, "/")

				modules[module] = moduleInfo{
					version: info.Name(),
					src:     path,
				}
			}

			return nil
		})
	}

	for moduleName, moduleInfo := range modules {
		lgger.Printf("saving %s...", moduleName)
		if err := loadModule(ctx, store, moduleName, moduleInfo); err != nil {
			return err
		}
	}

	return nil
}

func loadModule(ctx context.Context, store storage.Backend, moduleName string, moduleInfo moduleInfo) error {
	const op errors.Op = "loader.LoadModules"
	moduleBaseDir := moduleInfo.src

	goModFileDir := filepath.Join(moduleBaseDir, "go.mod")
	goModFile, err := os.ReadFile(goModFileDir)
	if err != nil {
		return errors.E(op, "failed loading go.mod file", err)
	}

	zipDir := filepath.Join(moduleBaseDir, "source.zip")
	zip, err := os.Open(zipDir)
	if err != nil {
		return errors.E(op, "failed loading source.zip file", err)
	}

	infoDir := filepath.Join(moduleBaseDir, fmt.Sprintf("%s.info", moduleInfo.version))
	info, err := os.ReadFile(infoDir)
	if err != nil {
		return errors.E(op, "failed loading moduleInfo file", err)
	}

	err = store.Save(ctx, moduleName, moduleInfo.version, goModFile, zip, info)
	if err != nil && !errors.Is(err, 409) {
		return err
	}

	return nil
}
