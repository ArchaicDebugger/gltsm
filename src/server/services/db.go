package services

import (
	"fmt"
	dbmodels "gltsm/models/db"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func EnvVariable(key string) string {
	err := godotenv.Load(".env")

	if err != nil {
		panic(err)
	}

	return os.Getenv(key)
}

func GetConnectionString() string {
	db_user := EnvVariable("DB_USER")
	db_pass := EnvVariable("DB_PASS")
	db_host := EnvVariable("DB_HOST")
	db_port := EnvVariable("DB_PORT")
	db_name := EnvVariable("DB_NAME")
	return fmt.Sprintf("user=%s password=%s host=%s dbname=%s port=%s sslmode=disable", db_user, db_pass, db_host, db_name, db_port)
}

func GetGormConnection() *gorm.DB {
	conn_str := GetConnectionString()
	db, err := gorm.Open(postgres.Open(conn_str), &gorm.Config{
		//Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		panic(err)
	}

	return db
}

func EnsureDbCreated() {
	db_name := EnvVariable("DB_NAME")
	conn_str := GetConnectionString()
	master_conn_str := strings.Replace(conn_str, db_name, "postgres", -1)

	master_db, err := gorm.Open(postgres.Open(master_conn_str), &gorm.Config{})

	if err != nil {
		panic(err)
	}

	err = master_db.Exec(fmt.Sprintf("CREATE DATABASE %s", db_name)).Error
	if err != nil && !strings.Contains(err.Error(), "already exists") {
		panic(err)
	} else if err != nil {
		fmt.Println("Database already there")
	}

	master_open_db, err := master_db.DB()
	if err != nil {
		panic(err)
	}

	err = master_open_db.Close()
	if err != nil {
		panic(err)
	}

	db := GetGormConnection()

	err = db.AutoMigrate(
		&dbmodels.User{},
		&dbmodels.Artist{},
		&dbmodels.Album{},
		&dbmodels.Track{},
		&dbmodels.Image{},
		&dbmodels.ListeningHistory{},
	)

	if err != nil {
		panic(err)
	}

	open_db, err := db.DB()
	if err != nil {
		panic(err)
	}

	err = open_db.Close()
	if err != nil {
		panic(err)
	}
}
