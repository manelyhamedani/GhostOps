package logprocessor

import (
	"bufio"
	"os"
)

// Read file line by line and call lineHandler on each line
func ReadLine(file *os.File, lineHandler func(string) error) error {
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()
		if err := lineHandler(line); err != nil {
			return err
		}
	}

	return scanner.Err()
}
