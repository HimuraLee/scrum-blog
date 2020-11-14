package service

import (
	"blog/config"
	"blog/db"
	"context"
	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
	"net/http"
)

type BackendService struct {
	cfg *config.Config
	echo *echo.Echo
	sql *db.SqlClient
}

func NewBackendService(cfg *config.Config) *BackendService {
	svc := &BackendService{
		cfg:        cfg,
		echo:       echo.New(),
		sql: db.MustNewSqlClient(cfg),
	}
	return svc
}

func (svc *BackendService) Start() {
	svc.ApiAdminRouter()
	svc.echo.Validator = NewCustomValidator()
	go func() {
		err := svc.echo.Start(svc.cfg.HTTP.Addr)
		if err != nil && err != http.ErrServerClosed {
			logrus.Fatal("fail to start echo", "error", err)
		}
	}()
}

func (svc *BackendService) Shutdown() {
	svc.echo.Shutdown(context.Background())
	logrus.Info("shutdown echo")
}