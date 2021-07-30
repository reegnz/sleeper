package pkg

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"
)

// ReadStdin is a generator function that returns lines
// of stdin from a channel
func ReadLines(r io.Reader) <-chan string {
	lines := make(chan string)
	go func() {
		defer close(lines)
		scanner := bufio.NewScanner(r)
		for scanner.Scan() {
			lines <- scanner.Text()
		}
	}()
	return lines
}

func WriteLines(w io.Writer) chan<- string {
	out := make(chan string)
	go func() {
		writer := bufio.NewWriter(w)
		for {
			data, ok := <-out
			if !ok {
				break
			}
			writer.WriteString(data)
		}
	}()
	return out
}

// ScheduledJobs spec
type ScheduledJobs struct {
	At      time.Time
	Command []string
}

// ParseLines parses dates in string format from a channel and
// puts parsed strings into another channel
// assumes it's the only one writing the dates channel, so it
func ParseLines(lines <-chan string) <-chan *ScheduledJobs {
	items := make(chan *ScheduledJobs)
	go func() {
		for line := range lines {
			columns := strings.Split(line, " ")
			dateString := columns[0]
			parsedTime, err := time.Parse(time.RFC3339, dateString)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed parsing time '%s': %v", dateString, err)
				break
			}

			items <- &ScheduledJobs{
				At: parsedTime,
			}
		}
		close(items)
	}()
	return items
}

// Sleep sleeps concurrently for all dates arriving in the channel
func Sleep(items <-chan *ScheduledJobs) sync.WaitGroup {
	results := make(chan string)
	var wg sync.WaitGroup
	for item := range items {
		go func(job *ScheduledJobs) {
			wg.Add(1)
			doAfter(at, func() {
				cmd := exec.Command(job.Command[0], job.Command[1:]...)
				var out bytes.Buffer
				cmd.Stdout = &out
				err := cmd.Run()
				if err != nil {
					log.Fatal(err)
				}
				results <- string(out.Bytes())
				wg.Done()
			})
		}(&item)
		wg.Wait()
		close(results)
	}
	return wg
}

// Printer whatever
func Printer(strings <-chan string) {
	for string := range strings {
		fmt.Print(string)
	}
}

func doAfter(date time.Time, fn func()) {
	go func() {
		<-time.After(time.Until(date))
		fn()
	}()
}

func makeAsync(fn func() interface{}) <-chan interface{} {
	c := make(chan interface{})
	go func() {
		defer close(c)
		c <- fn()
	}()
	return c
}

func asyncSquare(x int) <-chan int {
	make(chan int)
	squarex := func() interface{} {
		return square(x)
	}
	return makeAsync(squarex)
}

func square(x int) int {
	return x * x
}
