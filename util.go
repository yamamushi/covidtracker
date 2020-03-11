package main

import (
	"github.com/bwmarrin/discordgo"
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
