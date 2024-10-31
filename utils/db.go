package utils

import (
	"go.mills.io/bitcask/v2"
)

var db *bitcask.Bitcask

func InitDB() {
	db, _ = bitcask.Open("./db")
}

func GetDB() *bitcask.Bitcask {
	if db == nil {
		InitDB()
	}

	return db
}
