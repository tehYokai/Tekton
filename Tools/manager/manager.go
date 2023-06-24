package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"time"
)

func main() {
	// Define flags
	duration := flag.Duration("time", 1*time.Second, "the duration of the timer")
	taskName := flag.String("task", "Unnamed task", "the name of the task")

	// Parse the flags
	flag.Parse()

	// Get the total seconds
	totalSeconds := int((*duration).Seconds())

	// Start the countdown
	for i := totalSeconds; i >= 0; i-- {
		fmt.Printf("\rTimer: %d seconds remaining", i)
		time.Sleep(1 * time.Second)
	}

	// Execute the 'say' command
	cmd := exec.Command("say", fmt.Sprintf("your time of %v for task %s is up", *duration, *taskName))
	err := cmd.Run()
	if err != nil {
		panic(err)
	}

	// Open CSV file
	file, err := os.OpenFile("tasks.csv", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	// Write to CSV file
	writer := csv.NewWriter(file)
	defer writer.Flush()
	writer.Write([]string{time.Now().Format(time.RFC3339), *taskName, duration.String()})
}
