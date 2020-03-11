package main

import (
	"errors"
	"sync"
)

type StatsDB struct {
	db          *DBHandler
	CaseEntryquerylocker sync.RWMutex
	cityquerylocker sync.RWMutex
	countryquerylocker sync.RWMutex
	casequerylocker sync.RWMutex
}

type CaseEntryStats struct {
	Name string
	CaseEntryOfEmergency bool
	Confirmed string
	New string
	Deaths string
	Cities []CityStats
}

type CityStats struct {
	Name string
	Confirmed string
	New string
	Deaths string
}

type CountryStats struct {
	Name string
	Cased string
	Deaths string
	Serious string
	Critical string
	Recovered string
}

type CaseEntry struct {
	ID string
	CaseID string
	Date string
	CaseEntry string
	County string
	Text string
}

func NewStatsDB(db *DBHandler) (statsDB *StatsDB){
	statsDB = &StatsDB{db: db}
	return statsDB
}

func (h *StatsDB) GetEmptyCaseEntry() (entry CaseEntry, err error) {
	uuid, err := GetUUID()
	if err != nil {
		return entry, err
	}
	caseEntry := CaseEntry{ID: uuid}
	return caseEntry, nil
}

func (h *StatsDB) SetCaseEntry(CaseEntry CaseEntry) (err error) {
	CaseEntrydb, err := h.GetAllCaseEntryDB()
	if len(CaseEntrydb) < 1 {
		err = h.AddCaseEntryToDB(CaseEntry)
		if err != nil {
			return err
		}
		return nil
	}

	err = h.RemoveCaseEntryFromDB(CaseEntrydb[0])
	if err != nil {
		return err
	}

	err = h.AddCaseEntryToDB(CaseEntry)
	if err != nil {
		return err
	}

	return nil

}

// AddCaseEntryToDB function
func (h *StatsDB) AddCaseEntryToDB(entry CaseEntry) (err error) {
	h.casequerylocker.Lock()
	defer h.casequerylocker.Unlock()

	db := h.db.rawdb.From("CaseEntryDB")
	err = db.Save(&entry)
	return err
}

// RemoveCaseEntryFromDB function
func (h *StatsDB) RemoveCaseEntryFromDB(entry CaseEntry) (err error) {
	h.casequerylocker.Lock()
	defer h.casequerylocker.Unlock()

	db := h.db.rawdb.From("CaseEntryDB")
	err = db.DeleteStruct(&entry)
	return err
}

// RemoveCaseEntryFromDBByID function
func (h *StatsDB) RemoveCaseEntryFromDBByID(caseID string) (err error) {
	CaseEntry, err := h.GetCaseEntryFromDB(caseID)
	if err != nil {
		return err
	}

	err = h.RemoveCaseEntryFromDB(CaseEntry)
	if err != nil {
		return err
	}

	return nil
}

// GetCaseEntryFromDB function
func (h *StatsDB) GetCaseEntryFromDB(caseID string) (entry CaseEntry, err error) {
	CaseEntryDB, err := h.GetAllCaseEntryDB()
	if err != nil {
		return entry, err
	}

	for _, i := range CaseEntryDB {
		if i.CaseID == caseID {
			return i, nil
		}
	}
	return entry, errors.New("No record found")
}

// GetAllCaseEntryDB function
func (h *StatsDB) GetAllCaseEntryDB() (entryList []CaseEntry, err error) {
	h.casequerylocker.Lock()
	defer h.casequerylocker.Unlock()

	db := h.db.rawdb.From("CaseEntryDB")
	err = db.All(&entryList)
	if err != nil {
		return entryList, err
	}

	return entryList, nil
}

func (h *StatsDB) UpdateCaseEntry(entry CaseEntry) (err error) {
	h.casequerylocker.Lock()
	defer h.casequerylocker.Unlock()

	db := h.db.rawdb.From("CaseEntryDB")

	err = db.DeleteStruct(&entry)
	if err != nil {
		return err
	}
	err = db.Save(&entry)
	return err
}


