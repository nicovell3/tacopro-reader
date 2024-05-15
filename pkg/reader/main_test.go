package reader

import (
	"fmt"
	"log"
	"testing"

	"github.com/sf1/go-card/smartcard"
)

func selectSmartCard() (card *smartcard.Card, ctx *smartcard.Context, err error) {
	var currentReader *smartcard.Reader
	ctx, err = smartcard.EstablishContext()
	if err != nil {
		return nil, nil, err
	}
	readers, err := ctx.ListReadersWithCard()
	if err != nil {
		ctx.Release()
		return nil, nil, err
	}
	if len(readers) == 0 {
		ctx.Release()
		return nil, nil, fmt.Errorf("please insert smart card")
	}
	log.Println("Multiple readers (using the first one with present card)")
	for _, reader := range readers {
		log.Println(reader.Name(), reader.IsCardPresent())
		if currentReader == nil && reader.IsCardPresent() {
			currentReader = reader
		}
	}
	if currentReader == nil {
		ctx.Release()
		return nil, nil, fmt.Errorf("please insert smart card")
	}

	card, err = currentReader.Connect()
	return card, ctx, err
	//fmt.Printf("Card ATR: %s\n", card.ATR())
}

func TestMain(t *testing.T) {
	card, ctx, err := selectSmartCard()
	if err != nil {
		t.Fatalf("Error selecting card: %v\n", err)
	}
	defer ctx.Release()
	defer card.Disconnect()

	/*
		// This has never been tested
		err = selectTacograph(card, 2)
		if err != nil {
			log.Println("Card is not generation 2:", err)
		} else {
			t.Fatalf("Generation 2 card? No error reported selecting SMRTD")
		}
	*/

	err = selectTacograph(card, 1)
	if err != nil {
		t.Fatalf("Error selecting tacograph file: %v\n", err)
	}

	appIdBytes, err := processFile(card, definitions["Application_identification"])
	if err != nil {
		t.Fatalf("Error processing file Application_identification: %v\n", err)
	}
	parseEFApplicationIdentification(appIdBytes[5:15])
}
