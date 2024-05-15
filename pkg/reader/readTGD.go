package reader

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/sf1/go-card/smartcard"
)

type fileDefinition struct {
	Identifier       uint16
	Length           int
	RequireSignature bool
}

var /*const*/ allowedCardNumberCharacters = regexp.MustCompile(`^[A-Z0-9a-z]+$`)

//3. PROTOCOLO DE TRANSFERENCIA DE LOS DATOS ALMACENADOS EN TARJETAS DE TACÓGRAFO
// 3.5. Visión general de los comandos y los códigos de error

var headerFiles = []string{"ICC", "IC"}
var bodyFiles = []string{
	"Card_Certificate", "CA_Certificate", "Identification", "Card_Download",
	"Driving_License_Info", "Events_Data", "Faults_Data", "Driver_Activity_Data",
	"Vehicles_Used", "Places", "Current_Usage", "Control_Activity_Data",
	"Specific_Conditions"}

var definitions = map[string]*fileDefinition{
	"ICC":                        {Identifier: 0x0002, Length: 25, RequireSignature: false},
	"IC":                         {Identifier: 0x0005, Length: 8, RequireSignature: false},
	"Application_identification": {Identifier: 0x0501, Length: 10, RequireSignature: true},
	"Card_Certificate":           {Identifier: 0xC100, Length: 194, RequireSignature: false},
	"CA_Certificate":             {Identifier: 0xC108, Length: 194, RequireSignature: false},
	"Identification":             {Identifier: 0x0520, Length: 143, RequireSignature: true},
	"Card_Download":              {Identifier: 0x050e, Length: 4, RequireSignature: true},
	"Driving_License_Info":       {Identifier: 0x0521, Length: 53, RequireSignature: true},
	"Events_Data":                {Identifier: 0x0502, Length: 0, RequireSignature: true},
	"Faults_Data":                {Identifier: 0x0503, Length: 0, RequireSignature: true},
	"Driver_Activity_Data":       {Identifier: 0x0504, Length: 0, RequireSignature: true},
	"Vehicles_Used":              {Identifier: 0x0505, Length: 0, RequireSignature: true},
	"Places":                     {Identifier: 0x0506, Length: 0, RequireSignature: true},
	"Current_Usage":              {Identifier: 0x0507, Length: 19, RequireSignature: true},
	"Control_Activity_Data":      {Identifier: 0x0508, Length: 46, RequireSignature: true},
	"Specific_Conditions":        {Identifier: 0x0522, Length: 280, RequireSignature: true},
}

func ReadTGD(inputFilename string, card *smartcard.Card) (filename string, err error) {

	filename = inputFilename

	var tgdData []byte
	for _, name := range headerFiles {
		log.Println(name)
		fileBytes, err := processFile(card, definitions[name])
		if err != nil {
			return "", fmt.Errorf("error processing file %s: %v", name, err)
		}
		tgdData = append(tgdData, fileBytes...)
	}
	err = selectTacograph(card, 1)
	if err != nil {
		return "", fmt.Errorf("error selecting tacograph file: %v", err)
	}
	log.Println("Application_identification")
	appIdBytes, err := processFile(card, definitions["Application_identification"])
	if err != nil {
		return "", fmt.Errorf("error processing file Application_identification: %v", err)
	}
	parseEFApplicationIdentification(appIdBytes[5:15])
	tgdData = append(tgdData, appIdBytes...)

	for _, name := range bodyFiles {
		log.Println(name)
		fileBytes, err := processFile(card, definitions[name])
		if err != nil {
			return "", fmt.Errorf("error processing file %s: %v", name, err)
		}
		if name == "Identification" {
			generatedFilename, err := buildTGDFilename(fileBytes[5:16])
			if err != nil {
				return "", fmt.Errorf("error generating filename: %v", err)
			}
			if filename != "" {
				log.Println("Ignoring generated filename to used the specified one:", generatedFilename)
			} else {
				filename = generatedFilename
			}
		}
		tgdData = append(tgdData, fileBytes...)
	}
	err = writeTGD(tgdData, filename)
	return
}

func buildTGDFilename(input []byte) (string, error) {
	cardNumber := string(input[1:11])
	cardNumber = strings.TrimRight(cardNumber, " ")
	if !allowedCardNumberCharacters.MatchString(cardNumber) {
		return "", fmt.Errorf("card number contains not-allowed characters: %s", cardNumber)
	}
	return fmt.Sprintf("C_%s_%02X_%s.TGD",
		cardNumber,
		input[0],
		time.Now().Format("060102_1504"),
	), nil
}

func writeTGD(input []byte, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("cannot create file: %v", err)
	}
	defer file.Close()
	byteCount, err := file.Write(input)
	if err != nil {
		return fmt.Errorf("failed to write TGD data: %v", err)
	}
	if byteCount != len(input) {
		return fmt.Errorf("written distict number of required bytes (%d): %d", len(input), byteCount)
	}
	return nil
}

func processFile(card *smartcard.Card, file *fileDefinition) (output []byte, err error) {
	output = convertInt16ToBytes(file.Identifier)
	output = append(output, 0x00)
	err = selectFile(card, file.Identifier)
	if err != nil {
		return nil, fmt.Errorf("error selecting file: %v", err)
	}
	if file.RequireSignature {
		err = performHash(card)
		if err != nil {
			return nil, fmt.Errorf("error creating hash: %v", err)
		}
	}
	fileContents, err := readFile(card, file.Length)
	if err != nil {
		return nil, fmt.Errorf("error reading file: %v", err)
	}
	output = append(output, convertInt16ToBytes(uint16(len(fileContents)))...)
	output = append(output, fileContents...)
	if file.RequireSignature {
		output = append(output, convertInt16ToBytes(file.Identifier)...)
		output = append(output, 0x01, 0x00, 0x80)
		signature, err := getHash(card)
		if err != nil {
			return nil, fmt.Errorf("error downloading hash: %v", err)
		}
		if len(signature) != 0x80 {
			return nil, fmt.Errorf("signature with incorrect length: %d", len(signature))
		}
		output = append(output, signature...)
	}
	return
}

func selectFile(card *smartcard.Card, fileID uint16) error {
	_, err := sendCommand(card, append([]byte{0x00, 0xA4, 0x02, 0x0C, 0x02}, convertInt16ToBytes(fileID)...))
	return err
}

func readFile(card *smartcard.Card, fileLength int) (output []byte, err error) {
	if fileLength < 1 {
		return nil, fmt.Errorf("size should not be zero")
	}
	var part []byte
	var i uint16
	for i = 0; fileLength > 0; fileLength -= 0xFF {
		partId := convertInt16ToBytes(i * 0xFF)
		part, err = sendCommand(card, []byte{0x00, 0xB0, partId[0], partId[1], getIntAsSingleByte(fileLength)})
		if err != nil {
			return
		}
		output = append(output, part...)
		i++
	}
	return
}

func performHash(card *smartcard.Card) error {
	_, err := sendCommand(card, []byte{0x80, 0x2A, 0x90, 0x00})
	return err
}

func getHash(card *smartcard.Card) ([]byte, error) {
	output, err := sendCommand(card, []byte{0x00, 0x2A, 0x9E, 0x9A, 0x80})
	return output, err
}

func selectTacograph(card *smartcard.Card, generation int) error {
	//TCS_37
	aidData := []byte{0x00, 0xA4, 0x04, 0x0C, 0x06, 0xFF}
	versionID := "TACHO"
	if generation == 2 {
		versionID = "SMRDT"
	}
	aidData = append(aidData, []byte(versionID)...)
	output, err := sendCommand(card, aidData)
	if err != nil {
		return err
	}
	if len(output) != 0 {
		return fmt.Errorf("unexpected output in select tacograph command: %X", output)
	}
	return nil
}

func sendCommand(card *smartcard.Card, input []byte) ([]byte, error) {
	command := smartcard.CommandAPDU(input)
	response, err := card.TransmitAPDU(command)
	if err != nil {
		return nil, err
	}
	var byteArray []byte = response
	if len(byteArray) < 2 {
		return byteArray, fmt.Errorf("less than 2 bytes returned")
	} else if bytes.Equal(byteArray[len(byteArray)-2:], []byte{0x90, 0x00}) {
		return byteArray[:len(byteArray)-2], nil
	} else {
		return byteArray, fmt.Errorf("%X", byteArray[len(byteArray)-2:])
	}
}

func convertInt16ToBytes(input uint16) []byte {
	return []byte{byte(input >> 8), byte(input & 0x00FF)}
}

func getIntAsSingleByte(input int) byte {
	if input > 255 {
		return 255
	}
	return byte(input)
}

func parseEFApplicationIdentification(input []byte) {
	eventsPerType := int(input[1+2])
	definitions["Events_Data"].Length = eventsPerType * 24 * 6
	faultsPerType := int(input[1+2+1])
	definitions["Faults_Data"].Length = faultsPerType * 24 * 2
	activityStructureLength := binary.BigEndian.Uint16(input[1+2+1+1 : 1+2+1+1+2])
	definitions["Driver_Activity_Data"].Length = int(activityStructureLength) + 4
	vehicleRecords := binary.BigEndian.Uint16(input[1+2+1+1+2 : 1+2+1+1+2+2])
	definitions["Vehicles_Used"].Length = int(vehicleRecords)*31 + 2
	placeRecords := int(input[1+2+1+1+2+2])
	definitions["Places"].Length = int(placeRecords)*10 + 1
}
