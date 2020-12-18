package database

import (
	"fmt"
	"errors"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"log"
	"sort"
)

type PlayerRecord struct {
	Name       string
	TotalScore int
}

const (
	tableName = "ScribbleServicePlayers"
)

var (
	svc *dynamodb.DynamoDB
)

func Init() {
	mySession := session.Must(session.NewSession())
	svc = dynamodb.New(mySession)
}

func GetPlayerRecords() ([]PlayerRecord, error) {
	var playerRecords []PlayerRecord

	input := &dynamodb.ScanInput{
		TableName: aws.String(tableName),
	}

	if err := svc.ScanPages(input, func(page *dynamodb.ScanOutput, last bool) bool {
		var recs []PlayerRecord

		err := dynamodbattribute.UnmarshalListOfMaps(page.Items, &recs)
		if err != nil {
			panic(fmt.Sprintf("failed to unmarshal Dynamodb Scan Items, %v", err))
		}

		playerRecords = append(playerRecords, recs...)

		return true
	}); err != nil {
		return nil, err
	}

	log.Printf("Successfully retrieved %v player records:\n", len(playerRecords))
	for _, playerRecord := range playerRecords {
		log.Print(playerRecord)
	}

	return playerRecords, nil
}

func GetPlayerRecord(name string) (PlayerRecord, error) {
	var playerRecord PlayerRecord

	result, err := svc.GetItem(&dynamodb.GetItemInput{
		TableName: aws.String(tableName),
		Key: map[string]*dynamodb.AttributeValue{
			"Name": {
				S: aws.String(name),
			},
		},
	})
	if err != nil {
		fmt.Println(err.Error())
		return playerRecord, err
	}
	if result.Item == nil {
		msg := "Could not find '" + name + "'"
		return playerRecord, errors.New(msg)
	}
	
	err = dynamodbattribute.UnmarshalMap(result.Item, &playerRecord)
	if err != nil {
		panic(fmt.Sprintf("Failed to unmarshal Record, %v", err))
	}

	return playerRecord, nil
}

func PutPlayerRecord(record *PlayerRecord) error {
	av, err := dynamodbattribute.MarshalMap(record)
	if err != nil {
		log.Printf("failed to DynamoDB marshal Record, %v", err)
		return err
	}

	input := &dynamodb.PutItemInput{
		TableName: aws.String(tableName),
		Item:      av,
	}

	if _, err = svc.PutItem(input); err != nil {
		log.Printf("failed to put Record to DynamoDB, %v", err)
		return err
	}

	log.Printf("Successfully put player record: %v", record)

	return nil
}

func GetTotalScores() []PlayerRecord {
	playerRecords, err := GetPlayerRecords()
	if err != nil {
		return playerRecords
	}

	sort.Slice(playerRecords, func(i, j int) bool {
		return playerRecords[i].TotalScore > playerRecords[j].TotalScore
	})

	return playerRecords
}
