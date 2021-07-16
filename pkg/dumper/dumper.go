package dumper

import (
	"context"
	"github.com/gomods/athens/pkg/config"
	"github.com/gomods/athens/pkg/errors"
	"github.com/gomods/athens/pkg/index/mem"
	"github.com/gomods/athens/pkg/log"
	"github.com/gomods/athens/pkg/module"
	"github.com/gomods/athens/pkg/stash"
	"github.com/gomods/athens/pkg/storage/fs"
	"github.com/spf13/afero"
	"golang.org/x/sync/errgroup"
	"os"
	"path/filepath"
)

// DumpModules gets the modules that are referenced in go.mod, and packages them to a tar.gz file
func DumpModules(ctx context.Context, goModDir, outDir string, cfg *config.Config, lggr *log.Logger) error {
	const op errors.Op = "dumper.DumpModules"

	Fs := afero.NewOsFs()
	tempDir, err := afero.TempDir(Fs, "", "athens_dump")
	if err != nil {
		return errors.E(op, err)
	}
	defer Fs.RemoveAll(tempDir)

	goModDir, err = filepath.Abs(goModDir)
	if err != nil {
		return errors.E(op, "failed getting go mod dir output directory", err)
	}

	modfile, err := New(goModDir)
	if err != nil {
		return errors.E(op, "failed parsing go.mod file", err)
	}

	st, err := initializeStasher(tempDir, cfg)
	if err != nil {
		return errors.E(op, "failed initializing stasher", err)
	}

	errs, ctx := errgroup.WithContext(ctx)

	fetchedTotal := 0
	failedTotal := 0
	modules := modfile.GetModules()
	for _, m := range modules {
		func(path, version string) {
			lggr.Printf("fetching %s@%s", path, version)
			errs.Go(func() error {
				try := 0
				numOfTries := 5
				success := false

				for !success && try < numOfTries {
					if _, err := st.Stash(ctx, path, version); err != nil {
						lggr.Warnf("retrying fetching %s@%s", path, version)
						try++
					} else {
						success = true
					}
				}

				if !success {
					failedTotal++
					lggr.Errorf("failed fetching %s@%s {%v} (%d/%d)", path, version, err, failedTotal, len(modules))
					return err
				}

				fetchedTotal++
				lggr.Printf("successfully fetched %s@%s (%d/%d)", path, version, fetchedTotal, len(modules))
				return nil
			})
		}(m.Path, m.Version)
	}

	if err := errs.Wait(); err != nil {
		lggr.Errorf("an error occurred during fetch, continuing...")
	}

	lggr.Println("successfully fetched all modules")
	if err := compress(tempDir, outDir); err != nil {
		return err
	}
	return nil
}

func initializeStasher(storageDir string, cfg *config.Config) (stash.Stasher, error) {
	const op errors.Op = "pack.initializeStasher"

	storageDir, err := filepath.Abs(storageDir)
	if err != nil {
		return nil, errors.E(op, "failed getting absolute output directory", err)
	}

	afs := afero.NewOsFs()
	if err := afs.MkdirAll(storageDir, os.ModeDir|os.ModePerm); err != nil {
		return nil, errors.E(op, "failed initializing output directory", err)
	}

	s, err := fs.NewStorage(storageDir, afs)
	if err != nil {
		return nil, errors.E(op, "failed initializing backend storage: ", err)
	}

	f, err := module.NewGoGetFetcher(cfg.GoBinary, "", nil, afs)
	if err != nil {
		return nil, errors.E(op, "failed initializing fetcher: ", err)
	}

	indexer := mem.New()
	return stash.New(f, s, indexer, stash.WithPool(cfg.GoGetWorkers)), nil
}
