package main

import (
	"errors"
	"sync"
)

type StatsDB struct {
	db                   *DBHandler
	CaseEntryquerylocker sync.RWMutex
	cityquerylocker      sync.RWMutex
	countryquerylocker   sync.RWMutex
	casequerylocker      sync.RWMutex
}

type StateStats struct {
	ID                   string
	Name                 string
	CaseEntryOfEmergency bool
	Confirmed            string
	New                  string
	Deaths               string
	Cities               []CityStats
}

type CityStats struct {
	ID        string
	Name      string
	Confirmed string
	New       string
	Deaths    string
}

type CountryStat struct {
	ID        string
	Name      string
	Cases     string
	Deaths    string
	Serious   string
	Critical  string
	Recovered string
}

type CaseEntry struct {
	ID         string
	CasesRange string
	Date       string
	Link       string
	County     string
	Text       string
	State      string
	Posted     bool
	Time       int
}

func NewStatsDB(db *DBHandler) (statsDB *StatsDB) {
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

func (h *StatsDB) SetCaseEntry(entry CaseEntry) (err error) {
	entries, err := h.GetAllCaseEntryDB()
	if len(entries) < 1 {
		err = h.AddCaseEntryToDB(entry)
		if err != nil {
			return err
		}
		return nil
	}

	err = h.RemoveCaseEntryFromDB(entry)
	if err != nil {
		return err
	}

	err = h.AddCaseEntryToDB(entry)
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
/*
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
*/

// GetCaseEntryFromDB function
func (h *StatsDB) GetCaseEntryFromDB(cases string) (entry CaseEntry, err error) {
	CaseEntryDB, err := h.GetAllCaseEntryDB()
	if err != nil {
		return entry, err
	}

	for _, i := range CaseEntryDB {
		if i.CasesRange == cases {
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

func (h *StatsDB) GetEmptyCountryStat() (stat CountryStat, err error) {
	uuid, err := GetUUID()
	if err != nil {
		return stat, err
	}
	stat = CountryStat{ID: uuid}
	return stat, nil
}

func (h *StatsDB) SetCountryStat(stat CountryStat) (err error) {
	countryStatsDB, err := h.GetAllCountryStatsDB()
	if len(countryStatsDB) < 1 {
		err = h.AddCountryStatToDB(stat)
		if err != nil {
			return err
		}
		return nil
	}

	err = h.RemoveCountryStatFromDB(stat)
	if err != nil {
		return err
	}

	err = h.AddCountryStatToDB(stat)
	if err != nil {
		return err
	}

	return nil

}

// AddCaseEntryToDB function
func (h *StatsDB) AddCountryStatToDB(stat CountryStat) (err error) {
	h.countryquerylocker.Lock()
	defer h.countryquerylocker.Unlock()

	db := h.db.rawdb.From("CountryStatsDB")
	err = db.Save(&stat)
	return err
}

// RemoveCaseEntryFromDB function
func (h *StatsDB) RemoveCountryStatFromDB(stat CountryStat) (err error) {
	h.countryquerylocker.Lock()
	defer h.countryquerylocker.Unlock()

	db := h.db.rawdb.From("CountryStatsDB")
	err = db.DeleteStruct(&stat)
	return err
}

// RemoveCaseEntryFromDBByID function
func (h *StatsDB) RemoveCountryStatFromDBByName(name string) (err error) {
	CaseEntry, err := h.GetCaseEntryFromDB(name)
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
func (h *StatsDB) GetCountryStatFromDB(name string) (entry CountryStat, err error) {
	stats, err := h.GetAllCountryStatsDB()
	if err != nil {
		return entry, err
	}

	for _, i := range stats {
		if i.Name == name {
			return i, nil
		}
	}
	return entry, errors.New("No record found")
}

// GetAllCaseEntryDB function
func (h *StatsDB) GetAllCountryStatsDB() (stats []CountryStat, err error) {
	h.countryquerylocker.Lock()
	defer h.countryquerylocker.Unlock()

	db := h.db.rawdb.From("CountryStatsDB")
	err = db.All(&stats)
	if err != nil {
		return stats, err
	}

	return stats, nil
}

func (h *StatsDB) UpdateCountryStat(stat CountryStat) (err error) {
	h.countryquerylocker.Lock()
	defer h.countryquerylocker.Unlock()

	db := h.db.rawdb.From("CountryStatsDB")

	err = db.DeleteStruct(&stat)
	if err != nil {
		return err
	}
	err = db.Save(&stat)
	return err
}
