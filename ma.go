package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"time"
)

// Константы для настройки буфера
const (
	bufferSize    = 5               // Размер кольцевого буфера
	flushInterval = 2 * time.Second // Интервал опустошения буфера
)

func main() {
	// Канал для сигнала завершения работы
	done := make(chan struct{})
	defer close(done)

	// Создание и соединение стадий пайплайна
	numbers := readNumbers(done)
	filtered := filterNegative(done, numbers)
	divisibleBy3 := filterNotDivisibleBy3(done, filtered)
	buffered := bufferStage(done, divisibleBy3)

	// Потребление и вывод результатов
	for num := range buffered {
		fmt.Printf("Получены данные: %d\n", num)
	}
}

// readNumbers читает числа из консоли и отправляет их в канал
func readNumbers(done <-chan struct{}) <-chan int {
	output := make(chan int)
	go func() {
		defer close(output)
		scanner := bufio.NewScanner(os.Stdin)
		fmt.Println("Введите целые числа (для завершения введите 'q'):")
		for scanner.Scan() {
			if scanner.Text() == "q" {
				return
			}
			num, err := strconv.Atoi(scanner.Text())
			if err != nil {
				fmt.Println("Некорректный ввод. Пожалуйста, введите целое число.")
				continue
			}
			select {
			case output <- num:
			case <-done:
				return
			}
		}
	}()
	return output
}

// filterNegative фильтрует отрицательные числа
func filterNegative(done <-chan struct{}, input <-chan int) <-chan int {
	output := make(chan int)
	go func() {
		defer close(output)
		for num := range input {
			if num >= 0 {
				select {
				case output <- num:
				case <-done:
					return
				}
			}
		}
	}()
	return output
}

// filterNotDivisibleBy3 фильтрует числа, не кратные 3, исключая 0
func filterNotDivisibleBy3(done <-chan struct{}, input <-chan int) <-chan int {
	output := make(chan int)
	go func() {
		defer close(output)
		for num := range input {
			if num != 0 && num%3 == 0 {
				select {
				case output <- num:
				case <-done:
					return
				}
			}
		}
	}()
	return output
}

// bufferStage реализует буферизацию данных с периодическим опустошением
func bufferStage(done <-chan struct{}, input <-chan int) <-chan int {
	output := make(chan int)
	go func() {
		defer close(output)
		buffer := make([]int, 0, bufferSize)
		ticker := time.NewTicker(flushInterval)
		defer ticker.Stop()

		for {
			select {
			case <-done:
				return
			case num, ok := <-input:
				if !ok {
					flushBuffer(done, output, buffer)
					return
				}
				buffer = append(buffer, num)
				if len(buffer) == bufferSize {
					flushBuffer(done, output, buffer)
					buffer = buffer[:0]
				}
			case <-ticker.C:
				if len(buffer) > 0 {
					flushBuffer(done, output, buffer)
					buffer = buffer[:0]
				}
			}
		}
	}()
	return output
}

// flushBuffer отправляет все числа из буфера в выходной канал
func flushBuffer(done <-chan struct{}, output chan<- int, buffer []int) {
	for _, num := range buffer {
		select {
		case output <- num:
		case <-done:
			return
		}
	}
}
