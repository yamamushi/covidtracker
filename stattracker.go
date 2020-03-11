package main

import (
	"errors"
	"github.com/bwmarrin/discordgo"
	"log"
	"math/rand"
	"sort"
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
				_, _ = h.dg.ChannelEdit(channel.ID, "UPDATED: "+t.Format("02/01 @ 3:04 PM MST"))
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
		log.Println("Updating USA Stats")
		err := h.UpdateUSAStats()
		if err != nil {
			log.Println("Error retrieving USA Stats: " + err.Error())
		} else {
			log.Println("Posting Latest USA Events")
			err = h.PostLatestEvent()
			if err != nil {
				log.Println("Error Posting Latest Event: " + err.Error())
			}
		}
		time.Sleep(2 * time.Minute)
	}
}

func (h *StatTracker) UpdateUSAStats() (err error) {
	root, err := h.scraper.GetSiteRoot("http://coronavirus.1point3acres.com/en/")
	if err != nil {
		return err
	}

	eventsDiv := root.Find("div", "class", "ant-table-wrapper responsive-table")
	trList := eventsDiv.FindAll("tr")
	for _, tr := range trList {
		tdlist := tr.FindAll("td")
		if len(tdlist) == 6 {
			eventRecord, err := h.db.GetEmptyCaseEntry()
			if err != nil {
				log.Println("Error retrieving new event record: " + err.Error())
			}

			span := tdlist[0].FindAll("span")
			if len(span) > 0 {
				//log.Println("Cases: " + span[0].Text())
				eventRecord.CasesRange = span[0].Text()
			} else {
				break
			}
			//log.Println("Date: " + tdlist[1].Text())
			eventRecord.Date = tdlist[1].Text()
			//log.Println("State: " + tdlist[2].Text())
			eventRecord.State = tdlist[2].Text()
			//log.Println("County: " + tdlist[3].Text())
			eventRecord.County = tdlist[3].Text()

			// Descriptions don't work right now, everything is in chinese because it's handled in the browser
			// And cloudflare is blocking headless browser requests
			/*
				descSpan := tdlist[4].FindAll("span")
				if len(descSpan) > 0 {
					log.Println("Description: " + descSpan[0].Text())
				}
			*/
			aSpan := tdlist[5].FindAll("a")
			if len(aSpan) > 0 {
				//log.Println("Link: " + aSpan[0].Attrs()["href"])
				eventRecord.Link = aSpan[0].Attrs()["href"]
			}

			_, err = h.db.GetCaseEntryFromDB(eventRecord.CasesRange)
			if err != nil {
				if err.Error() != "No record found" {
					return errors.New("error checking event entry: " + eventRecord.CasesRange)
				} else {
					tmp, err := h.db.GetEmptyCaseEntry()
					if err != nil {
						return err
					}
					log.Println("Adding Event Record: " + eventRecord.CasesRange)
					eventRecord.ID = tmp.ID
					eventRecord.Posted = false
					eventRecord.Time = int(time.Now().UnixNano())
					err = h.db.AddCaseEntryToDB(eventRecord)
					if err != nil {
						return errors.New("error adding event: " + eventRecord.CasesRange + " " + err.Error())
					}
				}
			}
		}
	}

	return nil
}

func (h *StatTracker) PostLatestEvent() (err error) {
	events, err := h.db.GetAllCaseEntryDB()
	if err != nil {
		return err
	}

	if len(events) < 1 {
		return nil
	}

	var events_map = make(map[string]CaseEntry)
	for i, event := range events {
		events_map[strconv.Itoa(i)] = event
	}

	date_sorted_events := make(timeSlice, 0, len(events_map))
	for _, d := range events {
		date_sorted_events = append(date_sorted_events, d)
	}
	sort.Sort(date_sorted_events)
	//currentEvent := date_sorted_events[len(date_sorted_events)-1]
	//log.Println("Attempting: " + currentEvent.CasesRange)

	for i := len(date_sorted_events) - 1; i >= 0; i-- {
		currentEvent := date_sorted_events[i]
		//log.Println(currentEvent.Time)
		if !currentEvent.Posted && currentEvent.Time > 1583945039705795000 {
			currentEvent.Posted = true
			err = h.db.SetCaseEntry(currentEvent)
			if err != nil {
				return err
			}

			log.Println("Posting: " + currentEvent.CasesRange)
			for _, guild := range h.dg.State.Guilds {

				channels, err := h.dg.GuildChannels(guild.ID)
				if err != nil {
					break
				}
				for _, channel := range channels {
					if strings.Contains(channel.Name, "us-cases") {
						output := ":newspaper: "
						output = output + "\n"
						output = output + "Case #s: " + currentEvent.CasesRange + "\n"
						output = output + "Date: " + currentEvent.Date + "\n"

						stateRole, err := GetRoleIDByName(h.dg, guild.ID, currentEvent.State)
						if err == nil {
							output = output + "State: <@&" + stateRole + ">" + "\n"
						} else {
							output = output + "State: " + currentEvent.State + "\n"
						}

						output = output + "County: " + currentEvent.County + "\n"
						output = output + "Link: " + currentEvent.Link + "\n"

						_, _ = h.dg.ChannelMessageSend(channel.ID, output)
					}
				}
			}
		}
	}

	return nil
}
