package main

import (
	"fmt"
	"net"
	"time"
)

func main4() {
	l, err := net.Listen("tcp", "127.0.0.1:9000")
	if err != nil {
		return
	}

	go client()

	buf := make([]byte, 32*1024)

	defer l.Close()

	for {
		conn, err := l.Accept()
		if err != nil {
			return
		}

		go handle(conn, buf)
	}
}

func handle(conn net.Conn, buf []byte) {
	for i := 1; i < 256; i++ {
		n, err := conn.Write(buf)
		if err != nil {
			fmt.Println("failed to write")
		}
		fmt.Printf("%d: wrote %d\n", i, n)
	}

	fmt.Println("done!")
}

func client() {
	conn, err := net.Dial("tcp", "127.0.0.1:9000")
	if err != nil {
		fmt.Println("Failed to connect")
	}

	time.Sleep(5 * time.Second)

	buf := make([]byte, 128*1024)
	n, err := conn.Read(buf)
	if err != nil {
		fmt.Println("failed to read")
	}

	fmt.Printf("read %d\n", n)

	select {}
}
