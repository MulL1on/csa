package main

import (
	"fmt"
	"math/rand"
	"time"
)

func main() {
	maxNum := 100
	rand.Seed(time.Now().UnixNano())
	secretNumber := rand.Intn(maxNum)
	fmt.Println("Please input your guess")

	// 判断我们猜的数字和随机数的大小
	for {
		var guess int
		// 输入我们猜的数字
		_, err := fmt.Scanf("%d", &guess)
		// Go语言中处理错误的方法
		if err != nil {
			fmt.Println("Invalid input. Please enter an integer value")
			return
		}
		fmt.Println("You guess is", guess)

		if guess > secretNumber {
			fmt.Println("Your guess is bigger than the secret number")
			continue
		} else if guess < secretNumber {
			fmt.Println("Your guess is smaller than the secret number")
			continue
		} else {
			fmt.Println("Congratulations! You got the secret number!")
			break
		}
	}

}
