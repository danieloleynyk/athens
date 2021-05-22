package dumper

import (
	"context"
	"github.com/gomods/athens/pkg/errors"
	"github.com/gomods/athens/pkg/index/mem"
	"github.com/gomods/athens/pkg/module"
	"github.com/gomods/athens/pkg/stash"
	"github.com/gomods/athens/pkg/storage/fs"
	"github.com/spf13/afero"
	"golang.org/x/sync/errgroup"
	"log"
	"os"
	"path/filepath"
)

// DumpModules gets the modules that are referenced in go.mod, and packages them to a tar.gz file
func DumpModules(ctx context.Context, goModDir, outDir string) error {
	const op errors.Op = "pack.DumpModules"
	numOfWorkers := 4

	fs := afero.NewOsFs()
	tempDir, err := afero.TempDir(fs, "", "athens_dump")
	if err != nil {
		return errors.E(op, err)
	}
	defer fs.RemoveAll(tempDir)

	goModDir, err = filepath.Abs(goModDir)
	if err != nil {
		return errors.E(op, "failed getting go mod dir output directory", err)
	}

	modfile, err := New(goModDir)
	if err != nil {
		return errors.E(op, "failed parsing go.mod file", err)
	}

	st, err := initializeStasher(tempDir, numOfWorkers)
	if err != nil {
		return errors.E(op, "failed initializing stasher", err)
	}

	errs, ctx := errgroup.WithContext(ctx)

	for _, m := range modfile.GetModules() {
		func (path, version string) {
			log.Printf("fetching %s@%s\n", path, version)
			errs.Go(func() error {
				if _, err := st.Stash(ctx, m.Path, m.Version); err != nil {
					return err
				}

				log.Printf("successfully fetched %s@%s\n", path, version)
				return nil
			})
		} (m.Path, m.Version)
	}

	if err := errs.Wait(); err != nil {
		log.Println(err)
		return err
	}

	log.Println("successfully fetched all modules")
	if err := compress(tempDir, outDir); err != nil {
		return err
	}
	return nil
}

func initializeStasher(storageDir string, numOfWorkers int) (stash.Stasher, error) {
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

	f, err := module.NewGoGetFetcher("/usr/bin/go", "", nil, afs)
	if err != nil {
		return nil, errors.E(op, "failed initializing fetcher: ", err)
	}

	indexer := mem.New()
	return stash.New(f, s, indexer, stash.WithPool(numOfWorkers)), nil
}
