package main

import (
	"encoding/json"
	"fmt"
	network "github.com/jotingen/go-neuralnetwork"
	"io/ioutil"
	"os"
)

var (
	tt    []network.Network
	board = []float64{.5, .5, .5, .5, .5, .5, .5, .5, .5}
	total int
	err error
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
			tt = append(tt, network.New([]int{10, 729, 81, 9}))
		}

	}


	for gen := 1; gen <= 2; gen++ {
	fight(gen)
} 

	jsonString, err := json.MarshalIndent(tt, "", "  ")
	if err != nil {
		fmt.Println("Error converting to JSON:", err)
		os.Exit(1)
	}

	ioutil.WriteFile("test.final", jsonString, 0644)

}

func fight(gen int) {
	var illegal []int
	var wins []int
	for i := 0; i < total; i++ {
		illegal = append(illegal, 0)
		wins = append(wins, 0)
	}
	totalAttempts := 0
	passillegal := true
	for passillegal {
		passillegal = false
		for i := 0; i < total; i++ {
			wins[i] = 0
		}
		for i := 0; i < total; i++ {
			for j := i + 1; j < total; j++ {

				attempts := 0
				noillegals := false
				for !noillegals {
					thisillegal := []int{0, 0}

					reset()
				Game0:
					for move := 0; move < 9; move++ {
						if move%2 == 0 {
							won, illegalmoves := moveAI("X", tt[i])
							thisillegal[0] += illegalmoves
							if won {
								wins[i]++
								break Game0
							}
						}
						if move%2 == 1 {
							won, illegalmoves := moveAI("O", tt[j])
							thisillegal[1] += illegalmoves
							if won {
								wins[j]++
								break Game0
							}
						}
					}

					//print()

					reset()
				Game1:
					for move := 1; move < 9; move++ {
						if move%2 == 0 {
							won, illegalmoves := moveAI("X", tt[j])
							thisillegal[1] += illegalmoves
							if won {
								wins[j]++
								break Game1
							}
						}
						if move%2 == 1 {
							won, illegalmoves := moveAI("O", tt[i])
							thisillegal[0] += illegalmoves
							if won {
								wins[i]++
								break Game1
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

					attempts++
					totalAttempts++

					fmt.Printf("Gen %d: %5d Attempts, Illegal moves: %d      \r", gen, attempts, illegal)
				}
				fmt.Printf("Gen %d: %5d Attempts, Illegal moves: %d      \r", gen, attempts, illegal)
				//fmt.Printf("%dv%d: %5d Attempts, Illegal moves: %d             \n", i, j, attempts, illegal)
			}
		}
	}

	fmt.Println()
	fmt.Printf("Gen %d: Wins: %d\n", gen, wins)
}

func moveAI(player string, tt network.Network) (win bool, illegal int) {
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
		spot := tt.Calc(append(board, p))
		for xy := range spot {
			if spot[xy] > spot[largest_xy] {
				largest_xy = xy
			}
		}

		if spot[largest_xy] < .5 {
			//Nothing crossed the threshold, try and bring up outputs
			target := []float64{1, 1, 1, 1, 1, 1, 1, 1, 1, p}
			tt.Train(board, target)
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
			tt.Train(board, target)
			illegal++
		}

		//fmt.Printf("Attempts: %d  %5.3f\r", illegal, spot)
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
