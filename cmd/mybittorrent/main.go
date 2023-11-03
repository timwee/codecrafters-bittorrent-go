package main

import (
	// Uncomment this line to pass the first stage
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"unicode"

	bencode "github.com/jackpal/bencode-go" // Available if you need it!
)

// Example:
// - 5:hello -> hello
// - 10:hello12345 -> hello12345
func decodeBencode(bencodedString string) (interface{}, error) {
	if unicode.IsDigit(rune(bencodedString[0])) {
		var firstColonIndex int

		for i := 0; i < len(bencodedString); i++ {
			if bencodedString[i] == ':' {
				firstColonIndex = i
				break
			}
		}

		lengthStr := bencodedString[:firstColonIndex]

		length, err := strconv.Atoi(lengthStr)
		if err != nil {
			return "", err
		}

		return bencodedString[firstColonIndex+1 : firstColonIndex+1+length], nil
	} else if number, intErr := decodeBenEncdoedNumber(bencodedString); intErr == nil {
		return number, intErr
	} else if data, err := bencode.Decode(strings.NewReader(bencodedString)); err == nil {
		return data, err
		// tr
	} else {
		return "", fmt.Errorf("only strings are supported at the moment")
	}
}

func decodeBenEncdoedNumber(bencodedString string) (int, error) {
	strSize := len(bencodedString)
	if strSize < 3 {
		return -1, fmt.Errorf("not long enough for ben encoded int")
	}
	if bencodedString[0] != 'i' || bencodedString[strSize-1] != 'e' {
		return -1, fmt.Errorf("doesn't have i and e surrounding the string")
	}
	return strconv.Atoi(bencodedString[1 : strSize-1])

}

func readFileToString(fileName string) (string, error) {
	b, err := os.ReadFile(fileName)
	if err != nil {
		return "", err
	}

	return string(b), nil
}

func printInfo(m map[string]interface{}) {
	if url, ok := m["announce"]; ok {
		fmt.Printf("Tracker URL: %s\n", url)
		if info, ok := m["info"]; ok {
			if mInfo, ok := info.(map[string]interface{}); ok {
				// fmt.Print(mInfo)
				fmt.Printf("Length: %d\n", mInfo["length"])

				var buf bytes.Buffer
				if marshalErr := bencode.Marshal(&buf, info); marshalErr == nil {
					hasher := sha1.New()
					hasher.Write(buf.Bytes())
					// shaInfo := base64.URLEncoding.EncodeToString(hasher.Sum(nil))
					shaInfo := hex.EncodeToString(hasher.Sum(nil))
					fmt.Printf("Info Hash: %s", shaInfo)
				}
			}
		}
	}
}

func main() {
	// You can use print statements as follows for debugging, they'll be visible when running tests.
	// fmt.Println("Logs from your program will appear here!")

	command := os.Args[1]

	if command == "decode" {
		bencodedValue := os.Args[2]

		decoded, err := decodeBencode(bencodedValue)
		if err != nil {
			fmt.Println(err)
			return
		}

		jsonOutput, _ := json.Marshal(decoded)
		fmt.Println(string(jsonOutput))
	} else if command == "info" {
		fileName := os.Args[2]
		contents, err := readFileToString(fileName)
		if err != nil {
			fmt.Println(err)
			return
		}
		decoded, err := decodeBencode(contents)
		if err != nil {
			fmt.Println(err)
			return
		}
		if m, ok := decoded.(map[string]interface{}); ok {
			printInfo(m)
		} else {
			fmt.Println("Error encountered in parsing torrent file for info command")
		}
	} else {
		fmt.Println("Unknown command: " + command)
		os.Exit(1)
	}
}
