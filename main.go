package main

import (
	"flag"
	"github.com/scribble-rs/scribble.rs/database"
	"log"
	"math/rand"
	"os"
	"strconv"
	"time"

	"github.com/scribble-rs/scribble.rs/communication"
)

func main() {
	portHTTPFlag := flag.Int("portHTTP", -1, "defines the port to be used for http mode")
	flag.Parse()

	var portHTTP int
	if *portHTTPFlag != -1 {
		portHTTP = *portHTTPFlag
		log.Printf("Listening on port %d sourced from portHTTP flag.\n", portHTTP)
	} else {
		//Support for heroku, as heroku expects applications to use a specific port.
		envPort, _ := os.LookupEnv("PORT")
		parsed, parseError := strconv.ParseInt(envPort, 10, 16)
		if parseError == nil {
			portHTTP = int(parsed)
			log.Printf("Listening on port %d sourced from PORT environment variable\n", portHTTP)
		} else {
			portHTTP = 8080
			log.Printf("Listening on default port %d\n", portHTTP)
		}
	}

	//Setting the seed in order for the petnames to be random.
	rand.Seed(time.Now().UnixNano())

	database.Init()
	playerRecord1 := &database.PlayerRecord{
		ID:        "Player1",
		Name:      "First Player",
		HighScore: 10,
	}
	playerRecord2 := &database.PlayerRecord{
		ID:        "Player2",
		Name:      "Second Player",
		HighScore: 30,
	}
	playerRecord3 := &database.PlayerRecord{
		ID:        "Player3",
		Name:      "Third Player",
		HighScore: 20,
	}
	var err error
	if err = database.PutPlayerRecord(playerRecord1); err != nil {
		log.Fatal(err)
	}
	if err = database.PutPlayerRecord(playerRecord2); err != nil {
		log.Fatal(err)
	}
	if err = database.PutPlayerRecord(playerRecord3); err != nil {
		log.Fatal(err)
	}
	var playerRecords []database.PlayerRecord
	if playerRecords, err = database.GetPlayerRecords(); err != nil {
		log.Fatal(err)
	}
	for _, playerRecord := range playerRecords {
		log.Print(playerRecord)
	}

	log.Println("Started.")

	//If this ever fails, it will return and print a fatal logger message
	log.Fatal(communication.Serve(portHTTP))
}
