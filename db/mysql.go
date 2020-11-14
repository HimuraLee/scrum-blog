package db

import (
	"blog/config"
	"github.com/sirupsen/logrus"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
	"time"
)

const (
	PostTable = "post"
	CateTable = "cate"
	TagTable = "tag"
	UserTable = "user"
	PostTagTable = "post_tag"
)

type SqlClient struct {
	*gorm.DB
}

func MustNewSqlClient(cfg *config.Config) *SqlClient {
	db, err := gorm.Open(mysql.Open(cfg.MySQL.Source), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},
	})
	if err != nil {
		logrus.Fatalf("fail to connect mysql, %s", err)
	}
	DB, err := db.DB()
	if err != nil {
		logrus.Fatalf("fail to return db.DB(), ", err)
	}
	if cfg.MySQL.LogMode {
		db.Logger = logger.Default
	}
	DB.SetMaxIdleConns(cfg.MySQL.MaxIdleConns)
	DB.SetMaxOpenConns(0)
	DB.SetConnMaxLifetime(time.Hour)
	return &SqlClient{db}
}