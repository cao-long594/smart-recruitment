package db

import (
	"fmt"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Open 仅建立数据库连接，不执行 AutoMigrate。
// 请先在 MySQL 中执行仓库根目录 [scripts/schema.sql](scripts/schema.sql) 创建表结构。
func Open(dsn string) (*gorm.DB, error) {
	gdb, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		return nil, fmt.Errorf("gorm open: %w", err)
	}
	return gdb, nil
}
