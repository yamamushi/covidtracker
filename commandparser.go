package main

import (
	"errors"
	"github.com/bwmarrin/discordgo"
	"github.com/dustin/go-humanize"
	emoji "github.com/tmdvs/Go-Emoji-Utils"
	"log"
	"strconv"
	"strings"
)

type CommandParser struct {
	dg    *discordgo.Session
	stats *StatsDB
	conf  *Config
}

func NewCommandParser(dg *discordgo.Session, conf *Config, stats *StatsDB) (parser *CommandParser) {
	parser = &CommandParser{dg: dg, conf: conf, stats: stats}
	return parser
}

func (h *CommandParser) Read(s *discordgo.Session, m *discordgo.MessageCreate) {
	// very important to set this first!
	cp := h.conf.BotConfig.CP

	if m.Author.ID == s.State.User.ID {
		return
	}

	// Ignore bots
	if m.Author.Bot {
		return
	}

	message := strings.Fields(strings.ToLower(m.Content))
	if len(message) < 1 {
		log.Println(m.Content)
		return
	}

	command, payload := SplitCommandFromArgs(message)

	command = strings.ToLower(command)
	payload = strings.ToLower(payload)

	// If the message is "ping" reply with "Pong!"
	if command == cp+"ping" {
		_, _ = s.ChannelMessageSend(m.ChannelID, "Pong!")
		return
	}

	if command == cp+"stats" {
		h.CheckStats(payload, s, m)
		return
	}

	if len(m.Mentions) > 0 {
		for _, mention := range m.Mentions {
			if mention.ID == s.State.User.ID {
				h.CheckStats(payload, s, m)
				return
			}
		}
	}

}

func (h *CommandParser) CheckStats(payload string, s *discordgo.Session, m *discordgo.MessageCreate) {

	var err error
	if payload == "" {
		countryStats, err := h.stats.GetAllCountryStatsDB()
		if err != nil {
			_, _ = s.ChannelMessageSend(m.ChannelID, "Error: Could not access Global Stats database: "+err.Error())
			return
		}
		var cases, deaths, critical, serious, recovered int
		for _, stat := range countryStats {
			if stat.Cases != "" && stat.Cases != "-" {
				val, err := strconv.Atoi(strings.ReplaceAll(stat.Cases, ",", ""))
				if err != nil {
					_, _ = s.ChannelMessageSend(m.ChannelID, "Error: Could not parse stat entry for "+stat.Name+": "+err.Error())
					return
				}
				cases = cases + val
			}
			if stat.Deaths != "" && stat.Deaths != "-" {
				val, err := strconv.Atoi(strings.ReplaceAll(stat.Deaths, ",", ""))
				if err != nil {
					_, _ = s.ChannelMessageSend(m.ChannelID, "Error: Could not parse stat entry for "+stat.Name+": "+err.Error())
					return
				}
				deaths = deaths + val
			}
			if stat.Critical != "" && stat.Critical != "-" {
				val, err := strconv.Atoi(strings.ReplaceAll(stat.Critical, ",", ""))
				if err != nil {
					_, _ = s.ChannelMessageSend(m.ChannelID, "Error: Could not parse stat entry for "+stat.Name+": "+err.Error())
					return
				}
				critical = critical + val
			}
			if stat.Serious != "" && stat.Serious != "-" {
				val, err := strconv.Atoi(strings.ReplaceAll(stat.Serious, ",", ""))
				if err != nil {
					_, _ = s.ChannelMessageSend(m.ChannelID, "Error: Could not parse stat entry for "+stat.Name+": "+err.Error())
					return
				}
				serious = serious + val
			}
			if stat.Recovered != "" && stat.Recovered != "-" {
				val, err := strconv.Atoi(strings.ReplaceAll(stat.Recovered, ",", ""))
				if err != nil {
					_, _ = s.ChannelMessageSend(m.ChannelID, "Error: Could not parse stat entry for "+stat.Name+": "+err.Error())
					return
				}
				recovered = recovered + val
			}
		}

		unresolved := cases - recovered

		output := ":bulb: Global Case Stats ```\n"
		output = output + "Cases: " + humanize.Comma(int64(cases)) + "\n"
		output = output + "Deaths: " + humanize.Comma(int64(deaths)) + "\n"
		output = output + "Critical: " + humanize.Comma(int64(critical)) + "\n"
		output = output + "Serious: " + humanize.Comma(int64(serious)) + "\n"
		output = output + "Unresolved: " + humanize.Comma(int64(unresolved)) + "\n"
		output = output + "Recovered: " + humanize.Comma(int64(recovered)) + "\n"
		output = output + "```"

		_, _ = s.ChannelMessageSend(m.ChannelID, output)
		return
	}

	emojis := emoji.FindAll(payload)
	if len(emojis) > 0 {
		payload, err = h.FlagToCountry(payload)
		if err != nil {
			_, _ = s.ChannelMessageSend(m.ChannelID, "Error reading input: "+payload)
		}
	}

	payload = strings.TrimSpace(payload)
	if payload == "china" {
		payload = "mainland china"
	}
	if payload == "usa" || payload == "us" {
		payload = "united states"
	}
	country, err := h.stats.GetCountryStatFromDB(payload)
	if err == nil {
		output := ":bulb: ```\n"
		if country.Name != "" {
			output = output + "Country: " + strings.Title(country.Name) + "\n"
		}
		if country.Cases != "" {
			output = output + "Cases: " + country.Cases + "\n"
		} else {
			output = output + "Cases: -\n"
		}
		if country.Deaths != "" {
			output = output + "Deaths: " + country.Deaths + "\n"
		} else {
			output = output + "Deaths: -\n"
		}
		if country.Critical != "" {
			output = output + "Critical: " + country.Critical + "\n"
		} else {
			output = output + "Critical: -\n"
		}
		if country.Serious != "" {
			output = output + "Serious: " + country.Serious + "\n"
		} else {
			output = output + "Serious: -\n"
		}
		if country.Recovered != "" {
			output = output + "Recovered: " + country.Recovered + "\n"
		} else {
			output = output + "Recovered: -\n"
		}
		output = output + "```"
		_, _ = s.ChannelMessageSend(m.ChannelID, output)
		return
	}

	_, _ = s.ChannelMessageSend(m.ChannelID, "No data for "+strings.Title(payload)+" found.")
}

func (h *CommandParser) FlagToCountry(payload string) (country string, err error) {

	emojis := emoji.FindAll(payload)
	if len(emojis) > 0 {
		found := emojis[0].Match.(emoji.Emoji)
		description := strings.TrimPrefix(found.Descriptor, "Flag:")
		description = strings.TrimSpace(description)
		return strings.ToLower(description), nil
	} else {
		return "", errors.New("no emoji match found")
	}
}
