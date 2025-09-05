package main

import (
	"context"
	"flag"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Range struct {
	Start  int
	Finish int
}

type Data struct {
	Timeout int
	File    string
	Range   []Range
}

type arrayFlags []string

func (i *arrayFlags) String() string {
	return strings.Join(*i, ", ")
}

func (i *arrayFlags) Set(value string) error {
	*i = append(*i, value)
	return nil
}

func parseRange(s string) (Range, error) {
	parts := strings.Split(s, ":")
	if len(parts) != 2 {
		return Range{}, fmt.Errorf("incorrect range %s", s)
	}
	start, err1 := strconv.Atoi(parts[0])
	finish, err2 := strconv.Atoi(parts[1])
	if err1 != nil || err2 != nil {
		return Range{}, fmt.Errorf("error: %s", s)
	}
	return Range{Start: start, Finish: finish}, nil
}

func isPrime(number int) bool {
	if number < 2 {
		return false
	}
	for i := 2; i <= int(math.Sqrt(float64(number))); i++ {
		if number%i == 0 {
			return false
		}
	}
	return true
}

func PrimeNumber(ctx context.Context, wg *sync.WaitGroup, primerange Range, out chan<- int) {
	defer wg.Done()

	for i := primerange.Start; i <= primerange.Finish; i++ {
		select {
		case <-ctx.Done():
			return
		default:
			if isPrime(i) {
				out <- i
			}
		}
	}
}

func main() {
	timeOut := flag.Int("timeout", 0, "")
	fileName := flag.String("file", "", "")

	var ranges arrayFlags
	flag.Var(&ranges, "range", "")

	flag.Parse()

	if *timeOut == 0 || *fileName == "" || len(ranges) == 0 {
		flag.Usage()
	}

	file, err := os.Create(*fileName)
	if err != nil {
		return
	}
	defer file.Close()
	var data Data
	data.Timeout = *timeOut
	for _, r := range ranges {
		pr, err := parseRange(r)
		if err != nil {
			fmt.Println("Error:", err)
		}
		data.Range = append(data.Range, pr)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(data.Timeout)*time.Second)
	defer cancel()

	intCh := make(chan int)
	var wg sync.WaitGroup

	for _, r := range data.Range {
		wg.Add(1)
		go PrimeNumber(ctx, &wg, r, intCh)
	}

	go func() {
		for prime := range intCh {
			if _, err := fmt.Fprintln(file, prime); err != nil {
				fmt.Println("Error: ", err)
			}
		}
	}()

	go func() {
		wg.Wait()
		close(intCh)
	}()

	<-ctx.Done()
	fmt.Println("Finished:", ctx.Err())
}
