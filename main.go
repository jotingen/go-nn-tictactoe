package main

import (
	"encoding/json"
	"fmt"
	network "github.com/jotingen/go-neuralnetwork"
	"io/ioutil"
	"os"
)

var (
	tt    []*network.Network
	board = []float64{.5, .5, .5, .5, .5, .5, .5, .5, .5}
)

func main() {

	var illegal []int

	total := 10
	for i := 0; i < total; i++ {
		tt = append(tt, network.New([]int{10, 729, 81, 9}))
		illegal = append(illegal, 0)
	}

	jsonString, err := json.MarshalIndent(tt, "", "  ")
	if err != nil {
		fmt.Println("Error converting to JSON:", err)
		os.Exit(1)
	}

	//fmt.Println(string(jsonString))
	ioutil.WriteFile("test.init", jsonString, 0644)
	//os.Exit(0)

	totalAttempts := 0
	for p := 0; p < 2; p++ {
		for i := 0; i < total; i++ {
			for j := 0; j < total; j++ {

				if i != j {
					attempts := 0
					noillegals := false
					for !noillegals {
						thisillegal := []int{0, 0}

						reset()

						thisillegal[0] += moveAI("X", tt[i])
						thisillegal[1] += moveAI("O", tt[j])
						thisillegal[0] += moveAI("X", tt[i])
						thisillegal[1] += moveAI("O", tt[j])
						thisillegal[0] += moveAI("X", tt[i])
						thisillegal[1] += moveAI("O", tt[j])
						thisillegal[0] += moveAI("X", tt[i])
						thisillegal[1] += moveAI("O", tt[j])
						thisillegal[0] += moveAI("X", tt[i])

						//print()

						reset()

						thisillegal[1] += moveAI("X", tt[j])
						thisillegal[0] += moveAI("O", tt[i])
						thisillegal[1] += moveAI("X", tt[j])
						thisillegal[0] += moveAI("O", tt[i])
						thisillegal[1] += moveAI("X", tt[j])
						thisillegal[0] += moveAI("O", tt[i])
						thisillegal[1] += moveAI("X", tt[j])
						thisillegal[0] += moveAI("O", tt[i])
						thisillegal[1] += moveAI("X", tt[j])

						//print()

						if thisillegal[0] == 0 && thisillegal[1] == 0 {
							noillegals = true
						}
						illegal[i] += thisillegal[0]
						illegal[j] += thisillegal[1]

						attempts++
						totalAttempts++

						if totalAttempts%100 == 0 || totalAttempts == 1 {
							jsonString, err := json.MarshalIndent(tt, "", "  ")
							if err != nil {
								fmt.Println("Error converting to JSON:", err)
								os.Exit(1)
							}

							ioutil.WriteFile(fmt.Sprintf("test.%05d", totalAttempts), jsonString, 0644)
						}
						fmt.Printf("%dv%d: %5d Attempts, Illegal moves: %d : %d      \r", i, j, attempts, illegal, thisillegal)
					}
					fmt.Printf("%dv%d: %5d Attempts, Illegal moves: %d             \n", i, j, attempts, illegal)
				}
			}
		}
	}

}

func moveAI(player string, tt *network.Network) (illegal int) {
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
