package main

import (
	"context"
	"log"
	"math/rand"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

const (
	port = ":8080"
)

var proverbs = []string{
	"Don't communicate by sharing memory, share memory by communicating.",
	"Concurrency is not parallelism.",
	"Channels orchestrate; mutexes serialize.",
	"The bigger the interface, the weaker the abstraction.",
	"Make the zero value useful.",
	"interface{} says nothing.",
	"Gofmt's style is no one's favorite, yet gofmt is everyone's favorite.",
	"A little copying is better than a little dependency.",
	"Syscall must always be guarded with build tags.",
	"Cgo must always be guarded with build tags.",
	"Cgo is not Go.",
	"With the unsafe package there are no guarantees.",
	"Clear is better than clever.",
	"Reflection is never clear.",
	"Errors are values.",
	"Don't just check errors, handle them gracefully.",
	"Design the architecture, name the components, document the details.",
	"Documentation is for users.",
	"Don't panic.",
}

func handleConnection(conn net.Conn, ctx context.Context, wg *sync.WaitGroup) {
	defer func() {
		wg.Done()
		conn.Close()
	}()

	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			proverb := getRandomProverb()
			_, err := conn.Write([]byte(proverb + "\n"))
			if err != nil {
				log.Printf("Ошибка отправки данных: %v\n", err)
				return
			}
		case <-ctx.Done():
			log.Println("Контекст истек или был отменен")
			return
		}
	}
}

func getRandomProverb() string {
	return proverbs[rand.Intn(len(proverbs))]
}

func main() {
	rand.Seed(time.Now().UnixNano())

	ln, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatal(err)
	}
	defer ln.Close()

	log.Printf("Сервер запущен на порту %s\n", port)

	var wg sync.WaitGroup

	// Обработка сигналов
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan
		cancel()
	}()

	for {
		select {
		case <-ctx.Done():
			log.Println("Завершение сервера...")
			return
		default:
			conn, err := ln.Accept()
			if err != nil {
				log.Println(err)
				continue
			}

			wg.Add(1)
			go handleConnection(conn, ctx, &wg)
		}
	}
}
