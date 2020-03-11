package main

import (
	"errors"
	"github.com/bwmarrin/discordgo"
	"log"
	"math/rand"
	"strconv"
	"strings"
	"time"
)

type StatTracker struct {
	scraper *Scraper
	dg      *discordgo.Session
	db      *StatsDB

	GlobalCases      string
	GlobalDeaths     string
	GlobalRecovered  string
	GlobalUnresolved string
}

func NewStatTracker(dg *discordgo.Session, db *DBHandler) (statTracker *StatTracker) {

	statTracker = &StatTracker{dg: dg}
	statTracker.scraper = NewScraper()
	statTracker.db = NewStatsDB(db)

	return statTracker
}

func (h *StatTracker) RunCountryDataUpdater() {
	min := 10
	max := 20

	for {
		log.Println("Running Update")
		err := h.UpdateCountryStats()
		if err != nil {
			log.Println("Could not retrieve updated stats: " + err.Error())
		} else {
			log.Println("Running Post")
			err = h.PostSidebarStatsToDiscord()
			if err != nil {
				log.Println("Could not post stats to Discord: " + err.Error())
			}
		}
		duration := rand.Intn(max-min) + min
		time.Sleep(time.Duration(duration) * time.Minute)
	}
}

func (h *StatTracker) UpdateCountryStats() (err error) {
	root, err := h.scraper.GetSiteRoot("https://docs.google.com/spreadsheets/u/0/d/e/2PACX-1vR30F8lYP3jG7YOq8es0PBpJIE5yvRVZffOyaqC0GgMBN6yt0Q-NI8pxS7hd1F9dYXnowSC6zpZmW9D/pubhtml/sheet?headers=false&gid=0")
	if err != nil {
		return err
	}

	statusTableBox := root.FindAll("table")
	if len(statusTableBox) > 0 {
		for _, table := range statusTableBox {

			trList := table.FindAll("tr")
			for _, tr := range trList {

				thList := tr.FindAll("th")
				for _, th := range thList {

					thID := th.Attrs()["id"]
					if thID == "" {
						break
					}

					if thID == "0R3" {
						tdList := tr.FindAll("td")
						if len(tdList) == 5 {
							h.GlobalCases = tdList[0].Text()
							h.GlobalDeaths = tdList[1].Text()
							h.GlobalRecovered = tdList[2].Text()
							h.GlobalUnresolved = tdList[3].Text()
						}
					}

					thID = strings.TrimPrefix(thID, "0R")
					rowID, err := strconv.Atoi(thID)
					if err != nil {
						return errors.New("could not read ID of table properties")
					}
					if rowID > 5 {
						countryStat := CountryStat{}
						tdList := tr.FindAll("td")
						if len(tdList) > 5 {
							countryStat.Name = strings.ToLower(tdList[0].Text())
							if countryStat.Name == "total" {
								break
							}
							if countryStat.Name == "" {
								divs := tr.FindAll("div")
								for _, div := range divs {
									if div.Text() == "Diamond Princess" {
										countryStat.Name = strings.ToLower(div.Text())
									}
								}
							}
							if countryStat.Name == "" {
								break
							}

							log.Println("Processing: " + countryStat.Name)
							countryStat.Cases = tdList[1].Text()
							countryStat.Deaths = tdList[2].Text()
							countryStat.Serious = tdList[3].Text()
							countryStat.Critical = tdList[4].Text()
							countryStat.Recovered = tdList[5].Text()
						} else {
							break
						}

						existing, err := h.db.GetCountryStatFromDB(countryStat.Name)
						if err != nil {
							if err.Error() != "No record found" {
								return errors.New("error checking country stat: " + countryStat.Name)
							} else {
								tmp, err := h.db.GetEmptyCountryStat()
								if err != nil {
									return err
								}
								countryStat.ID = tmp.ID
							}
						} else {
							countryStat.ID = existing.ID
						}

						err = h.db.AddCountryStatToDB(countryStat)
						if err != nil {
							return errors.New("error adding country stat: " + countryStat.Name)
						}
					}

				}
			}
		}
	} else {
		return errors.New("could not BNO find table data")
	}

	return nil
}

func (h *StatTracker) PostSidebarStatsToDiscord() (err error) {
	for _, guild := range h.dg.State.Guilds {
		channels, err := h.dg.GuildChannels(guild.ID)
		if err != nil {
			return err
		}

		updated := false
		for sliceID, channel := range channels {
			if strings.Contains(channel.Name, "CONFIRMED:") {
				if channel.Name != "🌍 CONFIRMED: "+h.GlobalCases {
					time.Sleep(1 * time.Second)
					updated = true
					_, _ = h.dg.ChannelEdit(channel.ID, "🌍 CONFIRMED: "+h.GlobalCases)
					channels = MoveChannel(channels, sliceID, 0)
				}
			}
			if strings.Contains(channel.Name, "FATALITIES:") {
				if channel.Name != "💀 FATALITIES: "+h.GlobalDeaths {
					time.Sleep(1 * time.Second)
					updated = true
					_, _ = h.dg.ChannelEdit(channel.ID, "💀 FATALITIES: "+h.GlobalDeaths)
					channels = MoveChannel(channels, sliceID, 1)
				}
			}
			if strings.Contains(channel.Name, "UNRESOLVED:") {
				if channel.Name != "🔬 UNRESOLVED: "+h.GlobalUnresolved {
					time.Sleep(1 * time.Second)
					updated = true
					_, _ = h.dg.ChannelEdit(channel.ID, "🔬 UNRESOLVED: "+h.GlobalUnresolved)
					channels = MoveChannel(channels, sliceID, 2)
				}
			}
			if strings.Contains(channel.Name, "RECOVERIES:") {
				if channel.Name != "🌞 RECOVERIES: "+h.GlobalRecovered {
					time.Sleep(1 * time.Second)
					updated = true
					_, _ = h.dg.ChannelEdit(channel.ID, "🌞 RECOVERIES: "+h.GlobalRecovered)
					channels = MoveChannel(channels, sliceID, 3)
				}
			}

			if strings.Contains(channel.Name, "UPDATED:") {
				t, err := TimeIn(time.Now(), "America/New_York")
				if err != nil {
					return err
				}

				time.Sleep(1 * time.Second)
				updated = true
				_, _ = h.dg.ChannelEdit(channel.ID, "UPDATED: "+t.Format("02/01 @ 3:04 p.m. MST"))
				channels = MoveChannel(channels, sliceID, 4)
			}
		}
		err = h.dg.GuildChannelsReorder(guild.ID, channels)
		if err != nil {
			return err
		}
		if updated {
			time.Sleep(5 * time.Second)
		}
	}
	return nil
}

func (h *StatTracker) RunUSADataUpdater() {
	for {
		h.UpdateUSAStats()
		time.Sleep(10 * time.Minute)
	}
}

func (h *StatTracker) UpdateUSAStats() (err error) {
	_, err = h.scraper.GetSiteRoot("https://coronavirus.1point3acres.com/en")
	if err != nil {
		return err
	}
	return nil
}
