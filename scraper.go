package main

import (
	"github.com/anaskhan96/soup"
	"sync"
)

type Scraper struct {
	querylocker sync.RWMutex
}

func NewScraper() (scraper *Scraper) {
	scraper = &Scraper{}
	return scraper
}

func (h *Scraper) GetSiteRoot(url string) (root soup.Root, err error) {
	h.querylocker.Lock()
	defer h.querylocker.Unlock()
	soup.SetDebug(false)
	soup.Header("Accept-Language", "en-us")
	resp, err := soup.Get(url) // Append page=1000 so we get the last page
	if err != nil {
		return root, err
	}

	root = soup.HTMLParse(resp)
	return root, nil
}
