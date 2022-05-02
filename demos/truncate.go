package main

import (
	"log"
	"os"
	"strconv"
)

func main() {
	args := os.Args[1:]

	length, err := strconv.Atoi(args[0])

	file, err := os.OpenFile("file.txt", os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		log.Println("failed to create file")
		return
	}
	defer file.Close()

	err = file.Truncate(int64(length))
	if err != nil {
		log.Println("failed to truncate")
	}

	file.WriteString("Hello world!")
	file.Sync()
}
