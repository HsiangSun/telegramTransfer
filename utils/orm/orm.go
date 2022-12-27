package orm

import (
	"fmt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"os"
	"telgramTransfer/model"
	"telgramTransfer/utils/config"
	"telgramTransfer/utils/log"
)

var Gdb *gorm.DB

func InitDb() {
	sqlPath := fmt.Sprintf("%s%s%s%s%s", config.AppPath, string(os.PathSeparator), "db", string(os.PathSeparator), "transfer.db")
	log.Sugar.Infof("SqlPath:%s", sqlPath)
	db, err := gorm.Open(sqlite.Open(sqlPath), &gorm.Config{
		//Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Sugar.Errorf("open database err:%s", err)
	}
	err = db.AutoMigrate(model.Group{})
	if err != nil {
		log.Sugar.Errorf("orm auto migrate have error:%s", err)
	}
	database, _ := db.DB()
	database.SetMaxOpenConns(2)
	err = database.Ping()
	if err != nil {
		log.Sugar.Errorf("db pring:%s", err)
	}
	Gdb = db
}
