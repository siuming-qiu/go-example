package main

import (
	"bufio"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"
)

func main() {
	fmt.Printf("please enter a number: \n")
	rand.Seed(time.Now().Unix())
	secretNum := rand.Intn(100)
	fmt.Println(secretNum)
	reader := bufio.NewReader(os.Stdin)
	for {
		input, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("readstring err", err)
			continue
		}
		input = strings.TrimSuffix(input, "\n")
		guessNum, err := strconv.Atoi(input)
		if err != nil {
			fmt.Println("please enter a number")
			continue
		}
		if guessNum < secretNum {
			fmt.Println("your guess is smaller")
		} else if guessNum > secretNum {
			fmt.Println("your guess is bigger")
		} else {
			fmt.Println("your guess is right!")
			break
		}

	}
}
