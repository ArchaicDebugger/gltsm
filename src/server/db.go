package main

import (
	"fmt"
	dbmodels "gltsm/models/db"
	"strings"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func ensureDbCreated() {
	db_user := envVariable("DB_USER")
	db_pass := envVariable("DB_PASS")
	db_host := envVariable("DB_HOST")
	db_port := envVariable("DB_PORT")
	db_name := envVariable("DB_NAME")
	conn_str := fmt.Sprintf("user=%s password=%s host=%s dbname=%s port=%s sslmode=disable", db_user, db_pass, db_host, db_name, db_port)
	master_conn_str := fmt.Sprintf("user=%s password=%s host=%s dbname=postgres port=%s sslmode=disable", db_user, db_pass, db_host, db_port)

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

	db, err := gorm.Open(postgres.Open(conn_str), &gorm.Config{})

	if err != nil {
		panic(err)
	}

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
