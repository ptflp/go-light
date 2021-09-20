package db

import (
	light "github.com/ptflp/go-light"
	"github.com/ptflp/go-light/components"
	"github.com/ptflp/go-light/migration"
	"go.uber.org/zap"
)

func NewRepositories(cmps components.Componenter) light.Repositories {

	mainDB, err := NewDB(cmps.Logger(), cmps.Config().DB)
	if err != nil {
		cmps.Logger().Fatal("db initialization error", zap.Error(err))
	}
	migrator := migration.NewMigrator(mainDB)
	err = migrator.Migrate()
	if err != nil {
		cmps.Logger().Fatal("error on migration apply", zap.Error(err))
	}

	r := light.Repositories{
		Users: NewUserRepository(mainDB),
	}

	return r
}
