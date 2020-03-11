package main

import (
	"errors"
	"github.com/bwmarrin/discordgo"
	"log"
	"math/rand"
	"strings"
	"time"
)

type StatTracker struct {

	scraper *Scraper
	dg *discordgo.Session

	Cases string
	Deaths string
	Recovered string
	Unresolved string

}

func NewStatTracker(dg *discordgo.Session) (statTracker *StatTracker) {

	statTracker = &StatTracker{dg: dg}
	statTracker.scraper = NewScraper()

	return statTracker
}

func (h *StatTracker) Run() {
	min := 10
	max := 20

	for{
		log.Println("Running Update")
		err := h.Update()
		if err != nil {
			log.Println("Could not retrieve updated stats: "+err.Error())
		} else {
			log.Println("Running Post")
			err = h.PostToDiscord()
			if err != nil {
				log.Println("Could not post stats to Discord: "+err.Error())
			}
		}
		duration := rand.Intn(max - min) + min
		time.Sleep(time.Duration(duration)*time.Minute)
	}
}

func (h *StatTracker) Update() (err error) {

	root, err := h.scraper.GetSiteRoot("https://docs.google.com/spreadsheets/u/0/d/e/2PACX-1vR30F8lYP3jG7YOq8es0PBpJIE5yvRVZffOyaqC0GgMBN6yt0Q-NI8pxS7hd1F9dYXnowSC6zpZmW9D/pubhtml/sheet?headers=false&gid=0")
	if err != nil {
		return err
	}

	statusTableBox := root.FindAll("table")
	if len(statusTableBox) > 0 {
		for _, table := range statusTableBox {
			trList := table.FindAll("tr")
			for _, tr := range trList {
				thID0R3 := tr.FindAll("th", "id", "0R3")
				if len(thID0R3) == 1 {
					tdList := tr.FindAll("td")
					if len(tdList) == 5 {
						h.Cases = tdList[0].Text()
						h.Deaths = tdList[1].Text()
						h.Recovered = tdList[2].Text()
						h.Unresolved = tdList[3].Text()
					}
				}
			}

		}
	} else {
		return errors.New("could not BNO find table data")
	}

	return nil
}


func (h *StatTracker) PostToDiscord() (err error) {

	for _, guild := range h.dg.State.Guilds {

		channels, err := h.dg.GuildChannels(guild.ID)
		if err != nil {
			return err
		}

		updated := false
		for sliceID, channel := range channels {
			if strings.Contains(channel.Name, "CONFIRMED:"){
				if channel.Name != "🌍 CONFIRMED: "+h.Cases {
					time.Sleep(1*time.Second)
					updated = true
					_, _ = h.dg.ChannelEdit(channel.ID, "🌍 CONFIRMED: "+h.Cases)
					channels = MoveChannel(channels, sliceID, 0)
				}
			}
			if strings.Contains(channel.Name, "FATALITIES:"){
				if channel.Name != "💀 FATALITIES: "+h.Deaths {
					time.Sleep(1*time.Second)
					updated = true
					_, _ = h.dg.ChannelEdit(channel.ID, "💀 FATALITIES: "+h.Deaths)
					channels = MoveChannel(channels, sliceID, 1)
				}
			}
			if strings.Contains(channel.Name, "UNRESOLVED:"){
				if channel.Name != "🔬 UNRESOLVED: "+h.Unresolved {
					time.Sleep(1*time.Second)
					updated = true
					_, _ = h.dg.ChannelEdit(channel.ID, "🔬 UNRESOLVED: "+h.Unresolved)
					channels = MoveChannel(channels, sliceID, 2)
				}
			}
			if strings.Contains(channel.Name, "RECOVERIES:"){
				if channel.Name != "🌞 RECOVERIES: "+h.Recovered {
					time.Sleep(1*time.Second)
					updated = true
					_, _ = h.dg.ChannelEdit(channel.ID, "🌞 RECOVERIES: "+h.Recovered)
					channels = MoveChannel(channels, sliceID, 3)
				}
			}

			if strings.Contains(channel.Name, "UPDATED:"){
				t, err := TimeIn(time.Now(), "America/New_York")
				if err != nil {
					return err
				}

				time.Sleep(1*time.Second)
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
			time.Sleep(5*time.Second)
		}
	}
	return nil
}

