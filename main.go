package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	network "github.com/jotingen/go-neuralnetwork"
	"github.com/kokardy/listing"
	"io/ioutil"
	"math/rand"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync/atomic"
	"time"
)

type net struct {
	Net  network.Network
	wins uint64
}

const (
	MAXFAIL uint64 = 1000000
)

var (
	tt      []net
	total   int
	err     error
	gen     uint64
	games   uint64
	passes  uint64
	illegal []uint64
	pairs   [][]int
	running []bool
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
		for i := 0; i < total; i++ {
			illegal = append(illegal, 0)
			running = append(running, false)
		}
		fmt.Println("Total", total)
	} else {

		total = 1000
		for i := 0; i < total; i++ {
			tt = append(tt, net{network.New([]int{19, 729, 81, 9}), 0})
			illegal = append(illegal, 0)
			running = append(running, false)
		}

	}

	if len(os.Args) > 2 {
		//Two arguments means a human is playing
		board := newBoard()
	Game:
		for move := 0; move < 9; move++ {
			if move%2 == 0 {
				print(board)
				won := false
				fmt.Print("Type cell #: ")
				reader := bufio.NewReader(os.Stdin)
				input, _ := reader.ReadString('\n')
				mymove, err := strconv.Atoi(strings.TrimSpace(input))
				if err != nil {
					//If it is an invalid move, insult the player
					fmt.Println("Not a number you dummy")
					os.Exit(1)
				}
				mymove--
				if board[mymove] == 0 || board[mymove] == 1 || mymove > 8 {
					//If it is an invalid move, insult the player
					fmt.Println("You dummy")
					os.Exit(1)
				}
				board[mymove] = 1
				won = checkWin(board, 1)
				if won {
					fmt.Println("You win!")
					break Game
				}
			}
			if move%2 == 1 {
				won, _, _ := moveAI("O", tt[0].Net, board)
				if won {
					print(board)
					fmt.Println("You lose!")
					break Game
				}
			}
		}
	} else {
		fmt.Println("Generating unique pair list")

		//Generate list to permutate
		var test []int
		for i := 0; i < total; i++ {
			test = append(test, i)
			pairs = append(pairs, []int{i, i})
		}
		ss := listing.IntReplacer(test)
		for perm := range listing.Permutations(ss, 2, false, 10000) {

			//Process into pairs
			//TODO do this less hacky
			p := fmt.Sprint(perm)
			p = strings.Trim(p, "[")
			p = strings.Trim(p, "]")
			words := strings.Fields(p)
			var pair []int
			first, _ := strconv.Atoi(words[0])
			second, _ := strconv.Atoi(words[1])
			pair = append(pair, first)
			pair = append(pair, second)
			pairs = append(pairs, pair)

		}

		for gen = 1; gen <= 1000; gen++ {
			fight(gen)
			time.Sleep(time.Second)
			fmt.Printf("\nGen %d: Wins: [ ", gen)
			for i := 0; i < len(tt); i++ {
				fmt.Printf("%d ", tt[i].wins)
			}
			fmt.Println("]")

			sort.Slice(tt, func(i, j int) bool {
				return tt[i].wins > tt[j].wins
			})

			jsonString, err := json.MarshalIndent(tt, "", "  ")
			if err != nil {
				fmt.Println("Error converting to JSON:", err)
				os.Exit(1)
			}

			ioutil.WriteFile(fmt.Sprintf("gen.%05d", gen), jsonString, 0644)

			//Remove lowest 20
			replaceChild := 0
			replaceNew := 0
			parents := 0
			if total == 1 {
				replaceChild = 0
				replaceNew = 0
				parents = 1
			} else if total == 2 {
				replaceChild = 0
				replaceNew = 1
				parents = 1
			} else if total == 3 {
				replaceChild = 0
				replaceNew = 1
				parents = 1
			} else if total == 4 {
				replaceChild = 1
				replaceNew = 1
				parents = 2
			} else {
				replaceChild = total / 3
				replaceNew = 2
				parents = total / 2
			}
			tt = tt[:len(tt)-replaceChild-replaceNew]

			//Replace middle 15 with children
			for r := 0; r < replaceChild; r++ {
				ttnew := net{network.New([]int{19, 729, 81, 9}), 0}
				parentA := rand.Intn(parents)
				parentB := rand.Intn(parents)
				for i := 0; i < len(ttnew.Net.Neurons); i++ {
					for j := 0; j < len(ttnew.Net.Neurons[i]); j++ {
						parent := 0
						if rand.Intn(1) == 1 {
							parent = parentA
						} else {
							parent = parentB
						}
						for k := 0; k < len(ttnew.Net.Neurons[i][j].Weight); k++ {
							if rand.Intn(99) > 3 {
								//Replace with parent if not mutating
								ttnew.Net.Neurons[i][j].Weight[k] = tt[parent].Net.Neurons[i][j].Weight[k]
							}
						}
					}
				}
				tt = append(tt, ttnew)
			}

			//Replace lowest 5 with random new ones
			for r := 0; r < replaceNew; r++ {
				tt = append(tt, net{network.New([]int{19, 729, 81, 9}), 0})
			}
		}
	}

	jsonString, err := json.MarshalIndent(tt, "", "  ")
	if err != nil {
		fmt.Println("Error converting to JSON:", err)
		os.Exit(1)
	}

	ioutil.WriteFile("gen.final", jsonString, 0644)

}

func fight(gen uint64) {
	for i := 0; i < total; i++ {
		illegal[i] = 0
	}
	passillegal := true
	passes = 0
	for passillegal {
		passes++
		for i := 0; i < total; i++ {
			tt[i].wins = 1
		}
		games = 0
		passillegal = false
		var mypairs [][]int
		for i := range pairs {
			var mypair []int
			mypair = append(mypair, pairs[i][0])
			mypair = append(mypair, pairs[i][1])
			mypairs = append(mypairs, mypair)
		}

		for len(mypairs) > 0 {
			i := 0
			if len(mypairs) > 1 {
				i = rand.Intn(len(mypairs))
			}
			if !running[mypairs[i][0]] && !running[mypairs[i][1]] {
				games++
				p0 := mypairs[i][0]
				p1 := mypairs[i][1]
				mypairs = append(mypairs[:i], mypairs[i+1:]...)
				running[p0] = true
				running[p1] = true
				go func() {
					passillegal = play(p0, p1, true) || passillegal
					running[p0] = false
					running[p1] = false
				}()
			}
			//runningCount := 0
			//for r := 0; r < len(running); r++ {
			//	if running[r] {
			//		runningCount++
			//	}
			//}
			//for runningCount >= 8 {
			//	time.Sleep(time.Nanosecond*1000)
			//	runningCount = 0
			//	for r := 0; r < len(running); r++ {
			//		if running[r] {
			//			runningCount++
			//		}
			//	}
			//}
			time.Sleep(time.Nanosecond * 1000)
		}
		runningCount := 1
		for r := 0; r < len(running); r++ {
			if running[r] {
				runningCount++
			}
		}
		for runningCount == 0 {
			time.Sleep(time.Nanosecond * 1000)
			runningCount = 0
			for r := 0; r < len(running); r++ {
				if running[r] {
					runningCount++
				}
			}
		}

	}

}

func boardAI(board []float64) (boardForAI []float64) {

	for b := range board {
		if board[b] == 1 {
			boardForAI = append(boardForAI, 1)
		} else {
			boardForAI = append(boardForAI, 0)
		}
	}
	for b := range board {
		if board[b] == 0 {
			boardForAI = append(boardForAI, 1)
		} else {
			boardForAI = append(boardForAI, 0)
		}
	}
	return
}

func play(i int, j int, allowWin bool) (sawIllegals bool) {
	illegals := true
	sawIllegals = false

	//Cut off a hopeless network
	if illegal[i] >= MAXFAIL {
		atomic.AddUint64(&tt[i].wins, -tt[i].wins)
		return false
	}
	if illegal[j] >= MAXFAIL {
		atomic.AddUint64(&tt[j].wins, -tt[j].wins)
		return false
	}

	//Run games until we have a game with no illegal moves
	for illegals {
		thisillegal := []uint64{0, 0}

		board := newBoard()
		var boards0 [][]float64
		var boards1 [][]float64
		var moves0 [][]float64
		var moves1 [][]float64

	Game:
		for m := 0; m < 9; m++ {
			thisBoard := make([]float64, len(board))
			copy(thisBoard, board)
			if m%2 == 0 {
				boards0 = append(boards0, thisBoard)
				won, illegalmoves, move := moveAI("X", tt[i].Net, board)
				moves0 = append(moves0, move)
				thisillegal[0] += illegalmoves
				atomic.AddUint64(&illegal[i], illegalmoves)
				if won {
					if allowWin {
						atomic.AddUint64(&tt[i].wins, 1)
					}
					for w := range moves0 {
						tt[i].Net.Train(append(boardAI(boards0[w]), 1), moves0[w])
					}
					break Game
				} else {
					if illegal[i] >= MAXFAIL {
						atomic.AddUint64(&tt[i].wins, -tt[i].wins)

						return false
					}
				}
			}
			if m%2 == 1 {
				boards1 = append(boards1, thisBoard)
				won, illegalmoves, move := moveAI("O", tt[j].Net, board)
				moves1 = append(moves1, move)
				thisillegal[1] += illegalmoves
				atomic.AddUint64(&illegal[j], illegalmoves)
				if won {
					if allowWin {
						atomic.AddUint64(&tt[j].wins, 1)
					}
					for w := range moves1 {
						//fmt.Printf("WINNER TRAIN %4.2f %4.2f\n", boards1[w], moves1[w])
						tt[j].Net.Train(append(boardAI(boards1[w]), 0), moves1[w])
					}
					break Game
				} else {
					if illegal[j] >= MAXFAIL {
						atomic.AddUint64(&tt[j].wins, -tt[j].wins)

						return false
					}
				}
			}
		}

		if thisillegal[0] == 0 && thisillegal[1] == 0 {
			illegals = false
		} else {
			sawIllegals = true
		}

		//atomic.AddUint64(&illegal[i], thisillegal[0])
		//atomic.AddUint64(&illegal[j], thisillegal[1])

		fmt.Printf("Gen %d: Pass %d: %5d:%d Games, Illegal moves: %d\r", gen, passes, games, len(pairs), illegal)
	}
	fmt.Printf("Gen %d: Pass %d: %5d:%d Games, Illegal moves: %d\r", gen, passes, games, len(pairs), illegal)
	return
}

func moveAI(player string, net network.Network, board []float64) (win bool, illegal uint64, move []float64) {
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
	for !valid && illegal <= MAXFAIL {
		spot := net.Calc(append(boardAI(board), p))
		for xy := range spot {
			if spot[xy] > spot[largest_xy] {
				largest_xy = xy
			}
		}

		if spot[largest_xy] < .5 {
			//Nothing crossed the threshold, try and bring up outputs
			target := []float64{1, 1, 1, 1, 1, 1, 1, 1, 1}
			for xy := range board {
				if board[xy] != .5 {
					target[xy] = 0
				}
			}
			//fmt.Printf("%d SPEAKUP TRAIN %4.2f %4.2f %4.2f\n", illegal, board, spot, target)
			net.Train(append(boardAI(board), p), target)
			illegal++
		} else if board[largest_xy] == .5 {
			//If it is a valid move, allow it
			valid = true
		} else {
			//If it is an invalid move, train it to not prefer this move and try again until we get a valid move
			target := []float64{0, 0, 0, 0, 0, 0, 0, 0, 0}
			for xy := range board {
				if board[xy] != .5 {
					target[xy] = -1
				}
			}
			//fmt.Printf("%d ILLEGAL TRAIN %4.2f %4.2f %4.2f\n", illegal, board, spot, target)
			net.Train(append(boardAI(board), p), target)
			//net.Print()
			//os.Exit(0)
			illegal++
		}

	}
	move = []float64{0, 0, 0, 0, 0, 0, 0, 0, 0}
	move[largest_xy] = 1
	board[largest_xy] = p
	win = checkWin(board, p)

	return
}

func print(board []float64) {
	fmt.Printf(" %s | %s | %s\n", printXO(board[8], 9), printXO(board[7], 8), printXO(board[6], 7))
	fmt.Println("---+---+---")
	fmt.Printf(" %s | %s | %s\n", printXO(board[5], 6), printXO(board[4], 5), printXO(board[3], 4))
	fmt.Println("---+---+---")
	fmt.Printf(" %s | %s | %s\n", printXO(board[2], 3), printXO(board[1], 2), printXO(board[0], 1))
	fmt.Println()
}

func printXO(value float64, place int) (xo string) {
	if value == 1 {
		return "X"
	} else if value == 0 {
		return "0"
	} else {
		return fmt.Sprint(place)
	}
}

func newBoard() (board []float64) {
	return []float64{.5, .5, .5, .5, .5, .5, .5, .5, .5}
}

func checkWin(board []float64, p float64) (win bool) {
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
