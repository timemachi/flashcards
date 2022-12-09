package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"
)

func main() {
	//check command-line argument

	var importFile string
	flag.StringVar(&importFile, "import_from", "", "Enter import file name to load cards")
	var exportFile string
	flag.StringVar(&exportFile, "export_to", "", "Enter export file name to save the flashCards")
	flag.Parse()
	//initiate all
	reader := bufio.NewReader(os.Stdin)
	cardMap := make(map[string]string)
	wrongAnswerStatistic := make(map[string]int)
	var stringBuilder strings.Builder

	if importFile != "" {
		importCardsByFileName(importFile, &stringBuilder, cardMap)
	}
out:
	for {
		answer := handleInput(reader, &stringBuilder, "input the action (add, remove,"+
			" import, export, ask, exit, log, hardest card, reset stats)")
		switch answer {
		case "add":
			addCard(reader, &stringBuilder, cardMap)
		case "remove":
			removeCard(reader, &stringBuilder, cardMap)
		case "ask":
			ask(reader, &stringBuilder, cardMap, wrongAnswerStatistic)
		case "export":
			export(reader, &stringBuilder, cardMap)
		case "import":
			importCards(reader, &stringBuilder, cardMap)
		case "log":
			logDialogue(reader, &stringBuilder)
		case "hardest card":
			stats(&stringBuilder, wrongAnswerStatistic)
		case "reset stats":
			resetStats(&stringBuilder, wrongAnswerStatistic)
		case "exit":
			if exportFile != "" {
				exportByFileName(exportFile, &stringBuilder, cardMap)
			}
			fmt.Println("Bye bye!")
			break out
		}
	}
}

func stats(builder *strings.Builder, statistic map[string]int) {
	if len(statistic) == 0 {
		fmt.Println("There are no cards with errors.")
		builder.WriteString("There are no cards with errors.\n")
		return
	}
	//sort keys of statistic:https://www.geeksforgeeks.org/how-to-sort-golang-map-by-keys-or-values/#:~:text=To%20sort%20a%20map%20by,keys%20appearing%20as%20it%20is.
	keys := make([]string, 0, len(statistic))
	for k := range statistic {
		keys = append(keys, k)
	}
	sort.SliceStable(keys, func(i, j int) bool {
		return statistic[keys[i]] > statistic[keys[j]]
	})

	errorNum := statistic[keys[0]]
	hardestError := make([]string, 0, len(keys))
	for _, errorCard := range keys {
		if statistic[errorCard] == errorNum {
			hardestError = append(hardestError, errorCard)
		} else {
			break
		}
	}
	if len(hardestError) == 1 {
		fmt.Printf("The hardest card is \"%s\". You have %d errors answering it.\n\n", hardestError[0], errorNum)
		builder.WriteString(fmt.Sprintf("The hardest card is \"%s\". You have %d errors answering it.\n", hardestError[0], errorNum))
	} else {
		var helper strings.Builder
		helper.WriteString("The hardest cards are ")
		for index, errorCardName := range hardestError {
			helper.WriteString(fmt.Sprintf("\"%s\"", errorCardName))
			if index < len(hardestError)-1 {
				helper.WriteString(", ")
			} else {
				helper.WriteString(". ")
			}
		}
		helper.WriteString(fmt.Sprintf("You have %d errors answering them.", errorNum))
		fmt.Println(helper.String())
		builder.WriteString(helper.String() + "\n")
	}
}

func resetStats(builder *strings.Builder, statistic map[string]int) {
	for k, _ := range statistic {
		delete(statistic, k)
	}
	fmt.Println("Card statistics have been reset.")
	builder.WriteString("Card statistics have been reset.\n")
	fmt.Println()
}

func logDialogue(reader *bufio.Reader, builder *strings.Builder) {
	filename := handleInput(reader, builder, "File name:")
	//check if log file exist: https://stackoverflow.com/questions/12518876/how-to-check-if-a-file-exists-in-go
	if _, err := os.Stat(filename); errors.Is(err, os.ErrNotExist) {
		// log file does not exist: create it
		_, err := os.Create(filename)
		if err != nil {
			log.Fatal(err)
		}
	}

	file, err := os.OpenFile(filename, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println("filename is " + filename)
		log.Fatal(err)
	}
	defer file.Close()
	fmt.Println("The log has been saved.")
	builder.WriteString("The log has been saved.\n")
	_, err1 := fmt.Fprintln(file, builder.String())
	if err1 != nil {
		log.Fatal(err1)
	}
	builder.Reset()
}

func export(reader *bufio.Reader, builder *strings.Builder, cardMap map[string]string) {
	fileName := handleInput(reader, builder, "File name:")
	exportByFileName(fileName, builder, cardMap)
}
func exportByFileName(fileName string, builder *strings.Builder, cardMap map[string]string) {
	file, err := os.Create(fileName)
	if err != nil {
		fmt.Println("File not found.")
		builder.WriteString("File not found.\n")
		return
	}
	defer file.Close()
	for k, v := range cardMap {
		currCard := flashCard{Card: k, Definition: v}
		err := exportCards(file, currCard)
		if err != nil {
			log.Fatal(err)
		}
	}
	fmt.Printf("%d cards have been saved.\n\n", len(cardMap))
	builder.WriteString(fmt.Sprintf("%d cards have been saved.\n\n", len(cardMap)))
}
func exportCards(file *os.File, card flashCard) error {
	cardJson, err := json.Marshal(card)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintln(file, string(cardJson))
	if err != nil {
		return err
	}
	return nil
}
func importCards(reader *bufio.Reader, builder *strings.Builder, cardMap map[string]string) {
	fileName := handleInput(reader, builder, "File name:")
	importCardsByFileName(fileName, builder, cardMap)
}

func importCardsByFileName(fileName string, builder *strings.Builder, cardMap map[string]string) {
	file, err := os.Open(fileName)
	if err != nil {
		fmt.Println("File not found.")
		builder.WriteString("File not found.\n")
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	numOfImport := 0

	for scanner.Scan() {
		//import card
		var card flashCard
		err := json.Unmarshal([]byte(scanner.Text()), &card)
		if err != nil {
			log.Fatal(err)
		}
		cardMap[card.Card] = card.Definition
		numOfImport++
	}
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%d cards have been loaded.\n", numOfImport)
	builder.WriteString(fmt.Sprintf("%d cards have been loaded.\n", numOfImport))
}

// handle answer of user
func handleInput(reader *bufio.Reader, builder *strings.Builder, question string) string {
	fmt.Println(question)
	builder.WriteString(question + "\n")
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)
	builder.WriteString(input + "\n")
	return input
}
func ask(reader *bufio.Reader, builder *strings.Builder, cardMap map[string]string, statistic map[string]int) {
	time, err := strconv.Atoi(handleInput(reader, builder, "How many times to ask?"))
	if err != nil {
		fmt.Println("Please input number !")
		builder.WriteString("Please input number !\n")
		return
	}
bigLoop:
	for true {
		keys := make([]string, len(cardMap))
		i := 0
		for k := range cardMap {
			keys[i] = k
			i++
		}
		for _, k := range keys {
			v := cardMap[k]
			ans := handleInput(reader, builder, fmt.Sprintf("Print the definition of \"%s\":", k))
			if ans == v {
				fmt.Println("Correct!")
				builder.WriteString("Correct!\n")
			} else {
				statistic[k] += 1
				wrongBut := false
				for kk, vv := range cardMap {
					if vv == ans {
						fmt.Printf("Wrong. The right answer is \"%s\", but your definition is correct for \"%s\".\n", v, kk)
						builder.WriteString(fmt.Sprintf("Wrong. The right answer is \"%s\", but your definition is correct for \"%s\".\n", v, kk))
						wrongBut = true
						break
					}
				}
				if !wrongBut {
					fmt.Printf("Wrong. The right answer is \"%s\".\n", v)
					builder.WriteString(fmt.Sprintf("Wrong. The right answer is \"%s\".\n", v))
				}
			}
			time--
			if time == 0 {
				break bigLoop
			}
		}
	}
	fmt.Println()
}
func removeCard(reader *bufio.Reader, builder *strings.Builder, cardMap map[string]string) {
	removedCard := handleInput(reader, builder, "Which card?")
	if cardMap[removedCard] == "" {
		fmt.Printf("Can't remove \"%s\": there is no such card.\n", removedCard)
		builder.WriteString(fmt.Sprintf("Can't remove \"%s\": there is no such card.\n", removedCard))
	} else {
		delete(cardMap, removedCard)
		fmt.Println("The card has been removed.")
		builder.WriteString("The card has been removed.\n")
	}
	fmt.Println()
}
func addCard(reader *bufio.Reader, builder *strings.Builder, cardMap map[string]string) {
	card := createCard(reader, builder, cardMap)
	cardMap[card.Card] = card.Definition
	fmt.Printf("The pair (\"%s\":\"%s\") has been added.\n", card.Card, card.Definition)
	builder.WriteString(fmt.Sprintf("The pair (\"%s\":\"%s\") has been added.\n", card.Card, card.Definition))
	fmt.Println()
}

// create a new card
func createCard(reader *bufio.Reader, builder *strings.Builder, cardMap map[string]string) flashCard {
	term := strings.TrimSpace(handleInput(reader, builder, fmt.Sprintf("The card :")))
	//check if input term already exist
	isValidTerm := false
	for !isValidTerm {
		if cardMap[term] != "" {
			term = strings.TrimSpace(handleInput(reader, builder, fmt.Sprintf("The term \"%s\" already exists. Try again:", term)))
			continue
		}
		isValidTerm = true
	}

	cardDefinition := strings.TrimSpace(handleInput(reader, builder, fmt.Sprintf("The definition for the card:")))
	//check if input definition already exist
	isValidDefinition := false
	for !isValidDefinition {
		isValid := true
	mapLoop:
		for _, v := range cardMap {
			if cardDefinition == v {
				cardDefinition = strings.TrimSpace(handleInput(reader, builder, fmt.Sprintf("The definition \"%s\" already exists. Try again:", cardDefinition)))
				isValid = false
				break mapLoop
			}
		}
		if isValid {
			isValidDefinition = true
		}
	}
	var card flashCard
	card.set(term, cardDefinition)
	return card
}

type flashCard struct {
	Card       string `json:"card,omitempty"`
	Definition string `json:"definition,omitempty"`
}

func (receiver *flashCard) set(card, definition string) {
	receiver.Card = card
	receiver.Definition = definition
}
func (receiver *flashCard) print() {
	fmt.Printf("Card:\n%s\nDefinition:\n%s", receiver.Card, receiver.Definition)
}
