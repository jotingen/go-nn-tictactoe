package main

import (
	"encoding/json"
	"fmt"
	network "github.com/jotingen/go-neuralnetwork"
	"io/ioutil"
	"os"
	"sort"
)

type net struct {
	Net  network.Network
	wins int
}

var (
	tt    []net
	board = []float64{.5, .5, .5, .5, .5, .5, .5, .5, .5}
	total int
	err   error
)

func main() {

	if len(os.Args) > 1 {
		fmt.Println("Loading", os.Args[1])
		file, e := ioutil.ReadFile(os.Args[1])
		if e != nil {
			fmt.Printf("File error: %v\n", e)
			os.Exit(1)
		}
		json.Unmarshal(file, &tt)
		total = len(tt)
		fmt.Println("Total", total)
	} else {

		total = 25
		for i := 0; i < total; i++ {
			tt = append(tt, net{network.New([]int{10, 729, 81, 9}), 0})
		}

	}

	for gen := 1; gen <= 10; gen++ {
		fight(gen)
		fmt.Printf("Gen %d: Wins: [ ", gen)
		for i := 0; i < len(tt); i++ {
			fmt.Printf("%d ", tt[i].wins)
		}
		fmt.Println("]")

		jsonString, err := json.MarshalIndent(tt, "", "  ")
		if err != nil {
			fmt.Println("Error converting to JSON:", err)
			os.Exit(1)
		}

		ioutil.WriteFile(fmt.Sprintf("gen.%05d", gen), jsonString, 0644)

		sort.Slice(tt, func(i, j int) bool {
			return tt[i].wins > tt[j].wins
		})

		//Replace lowest 5 with random new ones
		replace := 5
		tt = tt[:len(tt)-replace]
		for r := 0; r < replace; r++ {
			tt = append(tt, net{network.New([]int{10, 729, 81, 9}), 0})
		}
	}

	jsonString, err := json.MarshalIndent(tt, "", "  ")
	if err != nil {
		fmt.Println("Error converting to JSON:", err)
		os.Exit(1)
	}

	ioutil.WriteFile("gen.final", jsonString, 0644)

}

func fight(gen int) {
	var illegal []int
	for i := 0; i < total; i++ {
		illegal = append(illegal, 0)
	}
	passillegal := true
	games := 0
	for passillegal {
		passillegal = false
		for i := 0; i < total; i++ {
			tt[i].wins = 0
		}
		for i := 0; i < total; i++ {
			for j := 0; j < total; j++ {
				if i != j {
					noillegals := false
					for !noillegals {
						thisillegal := []int{0, 0}

						reset()
					Game:
						for move := 0; move < 9; move++ {
							if move%2 == 0 {
								won, illegalmoves := moveAI("X", tt[i].Net)
								thisillegal[0] += illegalmoves
								if won {
									tt[i].wins++
									games++
									break Game
								}
							}
							if move%2 == 1 {
								won, illegalmoves := moveAI("O", tt[j].Net)
								thisillegal[1] += illegalmoves
								if won {
									tt[j].wins++
									games++
									break Game
								}
							}
						}

						//print()

						if thisillegal[0] == 0 && thisillegal[1] == 0 {
							noillegals = true
						} else {
							passillegal = true
						}

						illegal[i] += thisillegal[0]
						illegal[j] += thisillegal[1]

						if games%100 == 0 {
							fmt.Printf("Gen %d: %5d Games, Illegal moves: %d\r", gen, games, illegal)
						}
					}
					if games%100 == 0 {
						fmt.Printf("Gen %d: %5d Games, Illegal moves: %d\r", gen, games, illegal)
					}
				}
			}
		}
	}
	fmt.Printf("Gen %d: %5d Games, Illegal moves: %d\n", gen, games, illegal)

}

func moveAI(player string, net network.Network) (win bool, illegal int) {
	var p float64
	if player == "X" {
		p = 1
	} else if player == "O" {
		p = 0
	} else {
		panic("Invalid Player")
	}

	illegal = 0
	largest_xy := 0
	valid := false
	for !valid {
		spot := net.Calc(append(board, p))
		for xy := range spot {
			if spot[xy] > spot[largest_xy] {
				largest_xy = xy
			}
		}

		if spot[largest_xy] < .5 {
			//Nothing crossed the threshold, try and bring up outputs
			target := []float64{1, 1, 1, 1, 1, 1, 1, 1, 1, p}
			net.Train(board, target)
			illegal++
		} else if board[largest_xy] == .5 {
			//If it is a valid move, allow it
			valid = true
		} else {
			//If it is an invalid move, train it to not prefer this move and try again until we get a valid move
			target := []float64{0, 0, 0, 0, 0, 0, 0, 0, 0, p}
			for xy := range board {
				if board[xy] == .5 {
					target[xy] = spot[xy]
				}
			}
			net.Train(board, target)
			illegal++
		}

	}
	board[largest_xy] = p
	if board[2] == p && board[1] == p && board[0] == p {
		win = true
	}
	if board[5] == p && board[4] == p && board[3] == p {
		win = true
	}
	if board[8] == p && board[7] == p && board[6] == p {
		win = true
	}
	if board[8] == p && board[5] == p && board[2] == p {
		win = true
	}
	if board[7] == p && board[4] == p && board[1] == p {
		win = true
	}
	if board[6] == p && board[3] == p && board[0] == p {
		win = true
	}
	if board[8] == p && board[4] == p && board[0] == p {
		win = true
	}
	if board[6] == p && board[4] == p && board[2] == p {
		win = true
	}

	return
}

func print() {
	fmt.Printf(" %s | %s | %s\n", printXO(board[8]), printXO(board[7]), printXO(board[6]))
	fmt.Println("---+---+---")
	fmt.Printf(" %s | %s | %s\n", printXO(board[5]), printXO(board[4]), printXO(board[3]))
	fmt.Println("---+---+---")
	fmt.Printf(" %s | %s | %s\n", printXO(board[2]), printXO(board[1]), printXO(board[0]))
	fmt.Println()
}

func printXO(value float64) (xo string) {
	if value == 1 {
		return "X"
	} else if value == 0 {
		return "0"
	} else {
		return " "
	}
}

func reset() {
	board = []float64{.5, .5, .5, .5, .5, .5, .5, .5, .5}
}
