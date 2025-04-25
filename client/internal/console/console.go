package console

import (
	"bufio"
	"client/internal/domain"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

type ConsoleClient struct {
	baseURL string
	reader  *bufio.Reader
}

func NewConsoleClient() *ConsoleClient {
	return &ConsoleClient{
		baseURL: "http://localhost:8080",
		reader:  bufio.NewReader(os.Stdin),
	}
}

func (c *ConsoleClient) Run() error {
	for {
		fmt.Println("\n1. Create game")
		fmt.Println("2. Join game")
		fmt.Println("3. Exit")
		fmt.Print("Choose option: ")

		option, err := c.reader.ReadString('\n')
		if err != nil {
			return err
		}
		option = strings.TrimSpace(option)

		switch option {
		case "1":
			if err := c.createGame(); err != nil {
				fmt.Printf("Error: %v\n", err)
			}
		case "2":
			if err := c.joinGame(); err != nil {
				fmt.Printf("Error: %v\n", err)
			}
		case "3":
			return nil
		default:
			fmt.Println("Invalid option")
		}
	}
}

func (c *ConsoleClient) createGame() error {
	fmt.Print("Enter number of players (2-4): ")
	numPlayers, err := c.readInt(2, 4)
	if err != nil {
		return err
	}

	fmt.Print("Enter your name: ")
	name, err := c.reader.ReadString('\n')
	if err != nil {
		return err
	}
	name = strings.TrimSpace(name)

	gameID, playerID, err := c.createNewGame(numPlayers, name)
	if err != nil {
		return err
	}

	fmt.Printf("\nGame created! ID: %s\n", gameID)
	fmt.Println("Waiting for other players to join...")

	for {
		status, err := c.checkGameStatus(gameID)
		if err != nil {
			return err
		}

		if status.Started {
			fmt.Println("\nGame started!")
			break
		}

		fmt.Printf("\rPlayers: %d/%d", status.Players, status.MaxPlayers)
		time.Sleep(2 * time.Second)
	}

	return c.playerMenu(gameID, playerID)
}

func (c *ConsoleClient) joinGame() error {
	fmt.Print("Enter game ID: ")
	gameID, err := c.reader.ReadString('\n')
	if err != nil {
		return err
	}
	gameID = strings.TrimSpace(gameID)

	fmt.Print("Enter your name: ")
	name, err := c.reader.ReadString('\n')
	if err != nil {
		return err
	}
	name = strings.TrimSpace(name)

	data := fmt.Sprintf("game_id=%s&name=%s", gameID, name)
	resp, err := http.Post(c.baseURL+"/join", "application/x-www-form-urlencoded", strings.NewReader(data))
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    // Обрабатываем ошибки сервера
    if resp.StatusCode != http.StatusOK {
        var errorResp struct {
            Error string `json:"error"`
        }
        if err := json.NewDecoder(resp.Body).Decode(&errorResp); err != nil {
            return fmt.Errorf("failed to join game")
        }
        return fmt.Errorf(errorResp.Error)
    }

	var result struct {
		PlayerID string `json:"player_id"`
		Started  bool   `json:"started"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return err
	}

	if !result.Started {
		fmt.Println("\nWaiting for other players...")
		for {
			status, err := c.checkGameStatus(gameID)
			if err != nil {
				return err
			}

			if status.Started {
				fmt.Println("\nGame started!")
				break
			}

			fmt.Printf("\rPlayers: %d/%d", status.Players, status.MaxPlayers)
			time.Sleep(2 * time.Second)
		}
	}

	return c.playerMenu(gameID, result.PlayerID)
}

func (c *ConsoleClient) playerMenu(gameID, playerID string) error {
	for {
		fmt.Println("\n1. Make guess")
		fmt.Println("2. Exit game")
		fmt.Print("Choose option: ")

		option, err := c.reader.ReadString('\n')
		if err != nil {
			return err
		}
		option = strings.TrimSpace(option)

		switch option {
		case "1": 
		fmt.Print("Enter your guess (4 digits): ")
		guess, err := c.reader.ReadString('\n')
		if err != nil {
			return err
		}
		guess = strings.TrimSpace(guess)

		if len(guess) != 4 {
			fmt.Println("Guess must be 4 digits")
			continue
		}

		result, err := c.makeGuess(gameID, playerID, guess)
		if err != nil {
			return err
		}

		fmt.Printf("Black markers: %d (correct digit and position)\n", result.Black)
		fmt.Printf("White markers: %d (correct digit, wrong position)\n", result.White)

		if result.GameOver {
			if result.IsWinner {
				fmt.Println("Congratulations! You won!")
			} else {
				fmt.Println("Game over! You didn't guess the code.")
			}
			fmt.Print("Press Enter to continue...")
			c.reader.ReadString('\n')
			return nil
		}

		case "2":
			return nil
		default:
			fmt.Println("Invalid option")
		}
	}
}

func (c *ConsoleClient) makeGuess(gameID, playerID, guess string) (*domain.GameState, error) {
    data := fmt.Sprintf("game_id=%s&player_id=%s&guess=%s", gameID, playerID, guess)
    resp, err := http.Post(c.baseURL+"/guess", "application/x-www-form-urlencoded", strings.NewReader(data))
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    var result struct {
        Black    int  `json:"black"`
        White    int  `json:"white"`
        IsWinner bool `json:"isWinner"`
        GameOver bool `json:"gameOver"`
    }
    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return nil, err
    }

    return &domain.GameState{
        Black:    result.Black,
        White:    result.White,
        IsWinner: result.IsWinner,
        GameOver: result.GameOver,
    }, nil
}

func (c *ConsoleClient) createNewGame(maxPlayers int, name string) (string, string, error) {
	data := fmt.Sprintf("max_players=%d&creator_name=%s", maxPlayers, name)
	resp, err := http.Post(c.baseURL+"/create", "application/x-www-form-urlencoded", strings.NewReader(data))
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	var result struct {
		GameID   string `json:"game_id"`
		PlayerID string `json:"player_id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", "", err
	}

	return result.GameID, result.PlayerID, nil
}

type GameStatus struct {
    Started     bool `json:"started"`
    Players     int  `json:"players"`
	Finished    bool `json:"finished"` 
    MaxPlayers  int  `json:"max_players"`
}

func (c *ConsoleClient) checkGameStatus(gameID string) (GameStatus, error) {
    resp, err := http.Get(fmt.Sprintf("%s/game/%s/status", c.baseURL, gameID))
    if err != nil {
        return GameStatus{}, err
    }
    defer resp.Body.Close()

    var status GameStatus
    if err := json.NewDecoder(resp.Body).Decode(&status); err != nil {
        return GameStatus{}, err
    }

    return status, nil
}

func (c *ConsoleClient) readInt(min, max int) (int, error) {
	input, err := c.reader.ReadString('\n')
	if err != nil {
		return 0, err
	}
	input = strings.TrimSpace(input)
	
	num, err := strconv.Atoi(input)
	if err != nil || num < min || num > max {
		return min, fmt.Errorf("please enter a number between %d and %d", min, max)
	}
	
	return num, nil
}