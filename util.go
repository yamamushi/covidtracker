package main

import (
	"errors"
	"github.com/bwmarrin/discordgo"
	"strings"
	"time"
)

func TimeIn(t time.Time, name string) (time.Time, error) {
	loc, err := time.LoadLocation(name)
	if err == nil {
		t = t.In(loc)
	}
	return t, err
}

func InsertChannel(array []*discordgo.Channel, value *discordgo.Channel, index int) []*discordgo.Channel {
	return append(array[:index], append([]*discordgo.Channel{value}, array[index:]...)...)
}

func RemoveChannel(array []*discordgo.Channel, index int) []*discordgo.Channel {
	return append(array[:index], array[index+1:]...)
}

func MoveChannel(array []*discordgo.Channel, srcIndex int, dstIndex int) []*discordgo.Channel {
	value := array[srcIndex]
	return InsertChannel(RemoveChannel(array, srcIndex), value, dstIndex)
}

// GetRoleIDByName function
func GetRoleIDByName(s *discordgo.Session, guildID string, name string) (roleid string, err error) {
	name = strings.Title(name)
	roles, err := s.GuildRoles(guildID)
	if err != nil {
		return "", err
	}
	for _, role := range roles {
		if role.Name == name {
			return role.ID, nil
		}
	}
	return "", errors.New("Role ID Not Found: " + name)
}

// GetChannelIDByName function
func GetChannelIDByName(s *discordgo.Session, guildID string, name string) (channelID string, err error) {
	//name = strings.Title(name)
	channels, err := s.GuildChannels(guildID)
	if err != nil {
		return "", err
	}
	for _, channel := range channels {
		if channel.Name == name {
			return channel.ID, nil
		}
	}
	return "", errors.New("Channel ID Not Found: " + name)
}

// GetRoleNameByID function
func GetRoleNameByID(roleID string, guildID string, s *discordgo.Session) (rolename string, err error) {

	roles, err := s.GuildRoles(guildID)
	if err != nil {
		return "", err
	}

	for _, role := range roles {

		if role.ID == roleID {
			return role.Name, nil
		}
	}

	return "", errors.New("Role " + roleID + " not found in guild " + guildID)
}

// SplitCommandFromArg function
func SplitCommandFromArgs(input []string) (command string, message string) {

	// Remove the prefix from our command
	command = input[0]
	payload := RemoveStringFromSlice(input, command)

	for _, value := range payload {
		message = message + value + " "
	}
	return command, message
}

// RemoveStringFromSlice function
func RemoveStringFromSlice(s []string, r string) []string {
	for i, v := range s {
		if v == r {
			return append(s[:i], s[i+1:]...)
		}
	}
	return s
}

type timeSlice []CaseEntry

func (p timeSlice) Len() int {
	return len(p)
}

func (p timeSlice) Less(i, j int) bool {
	return p[i].Time < p[j].Time
}

func (p timeSlice) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}
