package main

import (
	"github.com/asdine/storm"
	"log"
)

// DBHandler struct
type DBHandler struct {
	rawdb *storm.DB
	conf  *Config
}

// Insert function
func (h *DBHandler) Insert(object interface{}) error {
	err := h.rawdb.Save(object)
	if err != nil {
		log.Println("Could not insert object: ", err.Error())
		return err
	}
	return nil
}

// Find function
func (h *DBHandler) Find(first string, second string, object interface{}) error {
	err := h.rawdb.One(first, second, object)
	if err != nil {
		return err
	}
	return nil
}

// Update function
func (h *DBHandler) Update(object interface{}) error {
	err := h.rawdb.Update(object)
	if err != nil {
		return err
	}
	return nil
}

