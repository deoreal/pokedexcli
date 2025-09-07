package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"math/rand"

	"github.com/deoreal/pokedexcli/internal/pokecache"
)

type config struct {
	nextURL     *string
	previousURL *string
	cache       *pokecache.Cache
	pokedex     map[string]Pokemon // map of caught pokemon
}

type cliCommand struct {
	name        string
	description string
	callback    func(*config, ...[]string) error
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

type LocationAreaResponse struct {
	ID                   int    `json:"id"`
	Name                 string `json:"name"`
	GameIndex            int    `json:"game_index"`
	EncounterMethodRates []struct {
		EncounterMethod struct {
			Name string `json:"name"`
			URL  string `json:"url"`
		} `json:"encounter_method"`
		VersionDetails []struct {
			Rate    int `json:"rate"`
			Version struct {
				Name string `json:"name"`
				URL  string `json:"url"`
			} `json:"version"`
		} `json:"version_details"`
	} `json:"encounter_method_rates"`
	Location struct {
		Name string `json:"name"`
		URL  string `json:"url"`
	} `json:"location"`
	Names []struct {
		Language struct {
			Name string `json:"name"`
			URL  string `json:"url"`
		} `json:"language"`
		Name string `json:"name"`
	} `json:"names"`
	PokemonEncounters []struct {
		Pokemon struct {
			Name string `json:"name"`
			URL  string `json:"url"`
		} `json:"pokemon"`
		VersionDetails []struct {
			EncounterDetails []struct {
				Chance          int   `json:"chance"`
				ConditionValues []any `json:"condition_values"`
				MaxLevel        int   `json:"max_level"`
				Method          struct {
					Name string `json:"name"`
					URL  string `json:"url"`
				} `json:"method"`
				MinLevel int `json:"min_level"`
			} `json:"encounter_details"`
			MaxChance int `json:"max_chance"`
			Version   struct {
				Name string `json:"name"`
				URL  string `json:"url"`
			} `json:"version"`
		} `json:"version_details"`
	} `json:"pokemon_encounters"`
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
	"explore": {
		name:        "explore",
		description: "Displays the Pokémon in a location area",
		callback:    commandExplore,
	},
	"catch": {
		name:        "catch",
		description: "Try to catch a Pokémon by name",
		callback:    commandCatch,
	},
	"inspect": {
		name:        "inspect",
		description: "Prints the stats of a Pokémon",
		callback:    commandInspect,
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

	commandName := in[0]
	if cmd, ok := Commands[commandName]; !ok {
		fmt.Println("Unknown command")
	} else {
		var err error
		// Pass arguments for commands that expect them (all except help, exit, map, mapb)
		switch commandName {
		case "explore", "catch":
			err = cmd.callback(cfg, in[1:])
		default:
			err = cmd.callback(cfg)
		}
		if err != nil {
			fmt.Println("Error occurred:", err)
		}
	}
}

// makeRequest handles HTTP requests with caching
func makeRequest(url string, cache *pokecache.Cache) ([]byte, error) {
	// Check cache first
	if data, found := cache.Get(url); found {
		return data, nil
	}

	// Make HTTP request
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bad status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}

	// Add to cache
	cache.Add(url, body)

	return body, nil
}

func main() {
	// Initialize cache with 5 second interval
	cache := pokecache.NewCache(5 * time.Second)

	cfg := &config{
		cache:   cache,
		pokedex: make(map[string]Pokemon),
	}

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

func commandHelp(cfg *config, args ...[]string) error {
	fmt.Println()
	fmt.Println("Welcome to the Pokedex!")
	fmt.Println("Usage:")
	fmt.Println()
	fmt.Println("help: Displays a help message")
	fmt.Println("map: Displays the names of 20 location areas")
	fmt.Println("mapb: Displays the previous 20 location areas")
	fmt.Println("explore <location-area-name>: Displays the Pokémon in a location area")
	fmt.Println("catch <pokemon-name>: Try to catch a Pokémon by name")
	fmt.Println("exit: Exit the Pokedex")
	fmt.Println()
	return nil
}

func commandExplore(cfg *config, args ...[]string) error {
	if len(args) == 0 || len(args[0]) == 0 {
		fmt.Println("You must provide a location area name")
		return nil
	}

	locationAreaName := args[0][0]
	url := fmt.Sprintf("https://pokeapi.co/api/v2/location-area/%s", locationAreaName)

	// Use cached request
	body, err := makeRequest(url, cfg.cache)
	if err != nil {
		return fmt.Errorf("failed to fetch location area data: %w", err)
	}

	var locationAreaResp LocationAreaResponse
	err = json.Unmarshal(body, &locationAreaResp)
	if err != nil {
		return fmt.Errorf("error unmarshaling JSON: %w", err)
	}

	fmt.Printf("\nExploring %s...\n", locationAreaName)
	fmt.Println("Found Pokémon:")

	if len(locationAreaResp.PokemonEncounters) == 0 {
		fmt.Println(" - No Pokémon found in this area")
	} else {
		for _, encounter := range locationAreaResp.PokemonEncounters {
			fmt.Printf(" - %s\n", encounter.Pokemon.Name)
		}
	}
	fmt.Println()

	return nil
}

func commandExit(cfg *config, args ...[]string) error {
	cfg.cache.Stop()
	fmt.Println("Closing the Pokedex... Goodbye!")
	os.Exit(0)
	return nil // This line won't be reached due to os.Exit(0)
}

func commandMap(cfg *config, args ...[]string) error {
	url := "https://pokeapi.co/api/v2/location-area"

	// If we have a next URL from previous pagination, use it
	if cfg.nextURL != nil {
		url = *cfg.nextURL
	}

	// Use cached request
	body, err := makeRequest(url, cfg.cache)
	if err != nil {
		return err
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

// Pokemon struct for storing caught Pokemon
type Pokemon struct {
	Name           string `json:"name"`
	BaseExperience int    `json:"base_experience"`
}

func commandCatch(cfg *config, args ...[]string) error {
	if len(args) == 0 || len(args[0]) == 0 {
		fmt.Println("You must provide a Pokémon name")
		return nil
	}
	pokemonName := args[0][0]
	fmt.Printf("Throwing a Pokeball at %s...\n", pokemonName)

	url := fmt.Sprintf("https://pokeapi.co/api/v2/pokemon/%s", pokemonName)
	body, err := makeRequest(url, cfg.cache)
	if err != nil {
		fmt.Printf("Could not find Pokémon: %s\n", pokemonName)
		return nil
	}

	var pokeResp struct {
		Name           string `json:"name"`
		BaseExperience int    `json:"base_experience"`
	}
	err = json.Unmarshal(body, &pokeResp)
	if err != nil {
		fmt.Println("Error parsing Pokémon data")
		return nil
	}

	// Already caught?
	if _, ok := cfg.pokedex[pokeResp.Name]; ok {
		fmt.Printf("%s is already in your Pokedex!\n", pokeResp.Name)
		return nil
	}

	// Determine catch chance: base 50%, minus (base_experience / 2)%, min 1%, max 90%
	catchChance := 50 - pokeResp.BaseExperience/2
	if catchChance < 1 {
		catchChance = 1
	}
	if catchChance > 90 {
		catchChance = 90
	}

	rand.Seed(time.Now().UnixNano())
	roll := rand.Intn(100) + 1 // 1-100

	if roll <= catchChance {
		fmt.Printf("Congratulations! You caught %s!\n", pokeResp.Name)
		cfg.pokedex[pokeResp.Name] = Pokemon{
			Name:           pokeResp.Name,
			BaseExperience: pokeResp.BaseExperience,
		}
	} else {
		fmt.Printf("%s escaped!\n", pokeResp.Name)
	}

	return nil
}

func commandMapB(cfg *config, args ...[]string) error {
	if cfg.previousURL == nil {
		fmt.Println("You're on the first page")
		return nil
	}

	url := *cfg.previousURL

	// Use cached request
	body, err := makeRequest(url, cfg.cache)
	if err != nil {
		return err
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
