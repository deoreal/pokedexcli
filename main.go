package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

type config struct {
	nextURL     *string
	previousURL *string
}

type cliCommand struct {
	name        string
	description string
	callback    func(*config) error
}

type LocationAreasResponse struct {
	Count    int     `json:"count"`
	Next     *string `json:"next"`
	Previous *string `json:"previous"`
	Results  []struct {
		Name string `json:"name"`
		URL  string `json:"url"`
	} `json:"results"`
}

var Commands = map[string]cliCommand{
	"exit": {
		name:        "exit",
		description: "Exit the Pokedex",
		callback:    commandExit,
	},
	"help": {
		name:        "help",
		description: "Displays a help message",
		callback:    commandHelp,
	},
	"map": {
		name:        "map",
		description: "Displays the names of 20 location areas",
		callback:    commandMap,
	},
	"mapb": {
		name:        "mapb",
		description: "Displays the previous 20 location areas",
		callback:    commandMapB,
	},
}

// trimMultipleSpaces removes all leading and trailing spaces and reduces all spaces to single spaces
func trimMultipleSpaces(text string) string {
	stop := 0
	for strings.Contains(text, "  ") {
		text = strings.ReplaceAll(text, "  ", " ")
		if stop > 100 {
			break
		}
		stop++
	}

	return text
}

func cleanInput(text string) []string {
	var res []string
	text = strings.ToLower(text)
	text = strings.TrimSpace(text)
	text = trimMultipleSpaces(text)

	if text == "" || text == " " {
		return res
	}

	res = strings.Split(text, " ")
	return res
}

func processInput(input string, cfg *config) {
	in := cleanInput(input)

	if len(in) == 0 {
		return
	}

	if cmd, ok := Commands[in[0]]; !ok {
		fmt.Println("Unknown command")
	} else {
		err := cmd.callback(cfg)
		if err != nil {
			fmt.Println("Error occurred:", err)
		}
	}
}

func main() {
	cfg := &config{}

	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("Pokedex > ")

		if !scanner.Scan() {
			break
		}
		input := strings.TrimSpace(scanner.Text())

		if input == "" {
			continue
		}

		processInput(input, cfg)

		if err := scanner.Err(); err != nil {
			fmt.Fprintf(os.Stderr, "Error reading input: %v\n", err)
		}
	}

	fmt.Println("Ciao")
}

func commandHelp(cfg *config) error {
	fmt.Println()
	fmt.Println("Welcome to the Pokedex!")
	fmt.Println("Usage:")
	fmt.Println()
	fmt.Println("help: Displays a help message")
	fmt.Println("map: Displays the names of 20 location areas")
	fmt.Println("mapb: Displays the previous 20 location areas")
	fmt.Println("exit: Exit the Pokedex")
	fmt.Println()
	return nil
}

func commandExit(cfg *config) error {
	fmt.Println("Closing the Pokedex... Goodbye!")
	os.Exit(0)
	return nil // This line won't be reached due to os.Exit(0)
}

func commandMap(cfg *config) error {
	url := "https://pokeapi.co/api/v2/location-area"

	// If we have a next URL from previous pagination, use it
	if cfg.nextURL != nil {
		url = *cfg.nextURL
	}

	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("error reading response body: %w", err)
	}

	var locationAreasResp LocationAreasResponse
	err = json.Unmarshal(body, &locationAreasResp)
	if err != nil {
		return fmt.Errorf("error unmarshaling JSON: %w", err)
	}

	// Update config with new pagination URLs
	cfg.nextURL = locationAreasResp.Next
	cfg.previousURL = locationAreasResp.Previous

	// Display the location areas
	fmt.Println()
	for _, result := range locationAreasResp.Results {
		fmt.Println(result.Name)
	}
	fmt.Println()

	return nil
}

func commandMapB(cfg *config) error {
	if cfg.previousURL == nil {
		fmt.Println("You're on the first page")
		return nil
	}

	url := *cfg.previousURL

	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("error reading response body: %w", err)
	}

	var locationAreasResp LocationAreasResponse
	err = json.Unmarshal(body, &locationAreasResp)
	if err != nil {
		return fmt.Errorf("error unmarshaling JSON: %w", err)
	}

	// Update config with new pagination URLs
	cfg.nextURL = locationAreasResp.Next
	cfg.previousURL = locationAreasResp.Previous

	// Display the location areas
	fmt.Println()
	for _, result := range locationAreasResp.Results {
		fmt.Println(result.Name)
	}
	fmt.Println()

	return nil
}
