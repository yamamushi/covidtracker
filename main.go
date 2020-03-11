package main

import (
	"flag"
	"github.com/asdine/storm"
	"github.com/bwmarrin/discordgo"
	"log"
	"os"
	"os/signal"
	"syscall"
)

// Variables used for command line parameters
var (
	ConfPath string
)

func init() {
	// Read our command line options
	flag.StringVar(&ConfPath, "c", "bot.conf", "Path to Config File")
	flag.Parse()

	_, err := os.Stat(ConfPath)
	if err != nil {
		log.Fatal("Config file is missing: ", ConfPath)
		//flag.Usage()
		//os.Exit(1)
	}
}

func main() {

	log.Println("|| Starting CovidTracker Bot ||")
	//log.SetOutput(ioutil.Discard)

	// Setup our tmp directory
	_, err := os.Stat("tmp")
	if err != nil {
		if os.IsNotExist(err) {
			err = os.Mkdir("tmp", os.FileMode(0777))
			if err != nil {
				log.Fatal("Could not make tmp directory! " + err.Error())
			}
		}
	}

	// Verify we can actually read our config file
	conf, err := ReadConfig(ConfPath)
	if err != nil {
		log.Println("error reading config file at: ", ConfPath)
		return
	}

	// Create a new Discord session using the provided bot token.
	dg, err := discordgo.New("Bot " + conf.DiscordConfig.Token)
	if err != nil {
		log.Println("error creating Discord session,", err)
		return
	}
	log.Println("Opening Connection to Discord")
	err = dg.Open()
	dg.State.TrackMembers = true
	dg.State.TrackChannels = true
	dg.State.TrackEmojis = true
	dg.State.TrackPresences = true
	dg.State.TrackRoles = true
	dg.State.TrackVoice = true
	if err != nil {
		log.Fatal("Error Opening Connection: ", err)
	}
	log.Println("Connection Established")
	defer dg.Close()

	// Create / open our embedded database
	db, err := storm.Open(conf.DBConfig.DBFile)
	if err != nil {
		log.Fatal(err)
		return
	}
	defer db.Close()
	// Run a quick first time db configuration to verify that it is working properly
	log.Println("Checking Database")
	dbhandler := DBHandler{conf: &conf, rawdb: db}

	log.Println("|| Initializing Stat Tracker ||")
	statTracker := NewStatTracker(dg, &dbhandler)
	//go statTracker.RunSidebarUpdater()
	//go statTracker.RunCountryDataUpdater()
	go statTracker.RunUSADataUpdater()

	log.Println("|| Initializing Command Parser ||")
	commandParser := NewCommandParser(dg, &conf, statTracker.db)
	dg.AddHandler(commandParser.Read)

	log.Println("|| Main Handler Initialized ||")

	// Wait here until CTRL-C or other term signal is received.
	log.Println("Bot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

}
