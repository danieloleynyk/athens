package actions

import (
	"fmt"
	"github.com/gomods/athens/pkg/config"
	"github.com/gomods/athens/pkg/log"
	"github.com/gomods/athens/pkg/storage"
	"github.com/sirupsen/logrus"
)

type App struct {
	Store storage.Backend
	Lggr  *log.Logger
}

func NewApp(conf *config.Config) (App, error) {
	logLvl, err := logrus.ParseLevel(conf.LogLevel)
	if err != nil {
		return App{}, err
	}
	lggr := log.New(conf.CloudRuntime, logLvl)

	store, err := GetStorage(conf.StorageType, conf.Storage, conf.TimeoutDuration())
	if err != nil {
		err = fmt.Errorf("error getting storage configuration (%s)", err)
		return App{}, err
	}

	return App{Store: store, Lggr: lggr}, nil
}
