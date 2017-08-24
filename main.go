package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	network "github.com/jotingen/go-neuralnetwork"
	"io/ioutil"
	"math/rand"
	"os"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
)

type net struct {
	Net  network.Network
	wins uint64
}

var (
	tt      []net
	total   int
	err     error
	gen     uint64
	games   uint64
	illegal []uint64
)

func main() {
	runtime.GOMAXPROCS(4)

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
		}
		fmt.Println("Total", total)
	} else {

		total = 1000
		for i := 0; i < total; i++ {
			tt = append(tt, net{network.New([]int{10, 729, 81, 9}), 0})
			illegal = append(illegal, 0)
		}

	}

	if len(os.Args) > 2 {

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
				if board[2] == 1 && board[1] == 1 && board[0] == 1 {
					won = true
				}
				if board[5] == 1 && board[4] == 1 && board[3] == 1 {
					won = true
				}
				if board[8] == 1 && board[7] == 1 && board[6] == 1 {
					won = true
				}
				if board[8] == 1 && board[5] == 1 && board[2] == 1 {
					won = true
				}
				if board[7] == 1 && board[4] == 1 && board[1] == 1 {
					won = true
				}
				if board[6] == 1 && board[3] == 1 && board[0] == 1 {
					won = true
				}
				if board[8] == 1 && board[4] == 1 && board[0] == 1 {
					won = true
				}
				if board[6] == 1 && board[4] == 1 && board[2] == 1 {
					won = true
				}
				if won {
					fmt.Println("You win!")
					break Game
				}
			}
			if move%2 == 1 {
				won, _ := moveAI("O", tt[0].Net, board)
				if won {
					print(board)
					fmt.Println("You lose!")
					break Game
				}
			}
		}
	} else {
		for gen = 1; gen <= 10; gen++ {
			fight(gen)
			sort.Slice(tt, func(i, j int) bool {
				return tt[i].wins > tt[j].wins
			})
			fight(gen)
			fmt.Printf("\nGen %d: Wins: [ ", gen)
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

			//Remove lowest 20
			replaceChild := total / 3
			replaceNew := total / 5
			parents := total / 5
			tt = tt[:len(tt)-replaceChild-replaceNew]

			//Replace middle 15 with children
			for r := 0; r < replaceChild; r++ {
				ttnew := net{network.New([]int{10, 729, 81, 9}), 0}
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
				tt = append(tt, net{network.New([]int{10, 729, 81, 9}), 0})
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
	games = 0
	for passillegal {
		passillegal = false
		for i := 0; i < total; i++ {
			tt[i].wins = 0
		}

		var wg sync.WaitGroup
		for i := 0; i < total; i++ {
			ndx := i;
			wg.Add(1)
			go func() {
				passillegal = play(ndx, ndx, false) || passillegal
				wg.Done()
			}()
			if (i+1)%4 == 0 {
		wg.Wait()
			}
		}
		wg.Wait()
		for i := 0; i < total; i++ {
			for j := 0; j < total; j++ {
				k := (i + 1*(total/4)) % total
				l := (j + 1*(total/4)) % total
				m := (i + 2*(total/4)) % total
				n := (j + 2*(total/4)) % total
				o := (i + 3*(total/4)) % total
				p := (j + 3*(total/4)) % total
				wg.Add(1)
				go func() {
					passillegal = play(i, j, true) || passillegal
					wg.Done()
				}()
				if k != i && k != j && l != i && l != j {
					wg.Add(1)
					go func() {
						passillegal = play(k, l, false) || passillegal
						wg.Done()
					}()
				}
				if m != k && m != l && n != k && n != l {
					if m != i && m != j && n != i && l != j {
						wg.Add(1)
						go func() {
							passillegal = play(m, n, false) || passillegal
							wg.Done()
						}()
					}
				}
				if o != m && o != n && p != m && p != n {
					if o != k && o != l && p != k && p != l {
						if o != i && o != j && p != i && p != j {
							wg.Add(1)
							go func() {
								passillegal = play(o, p, false) || passillegal
								wg.Done()
							}()
						}
					}
				}
				wg.Wait()
			}
		}
	}

}

func play(i int, j int, allowWin bool) (illegals bool) {
	noillegals := false
	for !noillegals {
		thisillegal := []uint64{0, 0}

		board := newBoard()
	Game:
		for move := 0; move < 9; move++ {
			if move%2 == 0 {
				won, illegalmoves := moveAI("X", tt[i].Net, board)
				thisillegal[0] += illegalmoves
				if won {
					if allowWin {
						atomic.AddUint64(&tt[i].wins, 1)
					}
					atomic.AddUint64(&games, 1)
					break Game
				}
			}
			if move%2 == 1 {
				won, illegalmoves := moveAI("O", tt[j].Net, board)
				thisillegal[1] += illegalmoves
				if won {
					if allowWin {
						atomic.AddUint64(&tt[j].wins, 1)
					}
					atomic.AddUint64(&games, 1)
					break Game
				}
			}
			if move == 8 {
				atomic.AddUint64(&games, 1)
			}
		}

		//print()

		if thisillegal[0] == 0 && thisillegal[1] == 0 {
			noillegals = true
		} else {
			illegals = true
		}

		atomic.AddUint64(&illegal[i], thisillegal[0])
		atomic.AddUint64(&illegal[j], thisillegal[1])

		fmt.Printf("Gen %d: %5d Games, Illegal moves: %d\r", gen, games, illegal)
	}
	fmt.Printf("Gen %d: %5d Games, Illegal moves: %d\r", gen, games, illegal)
	return
}

func moveAI(player string, net network.Network, board []float64) (win bool, illegal uint64) {
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
