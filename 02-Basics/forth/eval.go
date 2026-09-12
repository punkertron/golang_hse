//go:build !solution

package main

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type Evaluator struct {
	stack              []int
	customInstructions map[string][]string
}

// NewEvaluator creates evaluator.
func NewEvaluator() *Evaluator {
	return &Evaluator{
		stack:              make([]int, 0),
		customInstructions: make(map[string][]string),
	}
}

func (e *Evaluator) parseCommand(command string) ([]string, error) {
	if commands, ok := e.customInstructions[command]; ok {
		return commands, nil
	}

	switch command {
	case "dup", "over", "drop", "swap", "+", "-", "*", "/":
		return []string{command}, nil
	}

	if _, err := strconv.Atoi(command); err == nil {
		return []string{command}, nil
	}

	return nil, fmt.Errorf("error '%s': non-existing command", command)
}

func (e *Evaluator) createCommand(commands []string) (int, error) {
	n := len(commands)
	if n < 4 {
		return 0, errors.New("not enough args to create command")
	}

	if _, err := strconv.Atoi(commands[1]); err == nil {
		return 0, fmt.Errorf("trying to redefine number '%s'", commands[1])
	}

	command := make([]string, 0, 3)

	for i := 2; i < n; i++ {
		if commands[i] == ";" {
			e.customInstructions[commands[1]] = command
			return i, nil
		}

		parsed, err := e.parseCommand(commands[i])
		if err != nil {
			return 0, err
		}

		command = append(command, parsed...)
	}

	return 0, errors.New("no symbol \";\" at the end of the new command")
}

func (e *Evaluator) execCommand(command string) error {
	if commands, ok := e.customInstructions[command]; ok {
		for _, comm := range commands {
			if err := e.execPrimitive(comm); err != nil {
				return err
			}
		}

		return nil
	}

	return e.execPrimitive(command)
}

func (e *Evaluator) execPrimitive(command string) error {
	if num, err := strconv.Atoi(command); err == nil {
		e.stack = append(e.stack, num)
		return nil
	}

	n := len(e.stack)

	switch command {
	case "over", "swap", "+", "-", "*", "/":
		if n < 2 {
			return fmt.Errorf("error '%s': not enough elements", command)
		}
	case "dup", "drop":
		if n < 1 {
			return fmt.Errorf("error '%s': not enough elements", command)
		}
	}

	switch command {
	case "dup":
		e.stack = append(e.stack, e.stack[n-1])
	case "over":
		e.stack = append(e.stack, e.stack[n-2])
	case "drop":
		e.stack = e.stack[:n-1]
	case "swap":
		e.stack[n-2], e.stack[n-1] = e.stack[n-1], e.stack[n-2]
	case "+":
		res := e.stack[n-2] + e.stack[n-1]
		e.stack = e.stack[:n-2]
		e.stack = append(e.stack, res)
	case "-":
		res := e.stack[n-2] - e.stack[n-1]
		e.stack = e.stack[:n-2]
		e.stack = append(e.stack, res)
	case "*":
		res := e.stack[n-2] * e.stack[n-1]
		e.stack = e.stack[:n-2]
		e.stack = append(e.stack, res)
	case "/":
		if e.stack[n-1] == 0 {
			return fmt.Errorf("error '/': integer divide by zero")
		}

		res := e.stack[n-2] / e.stack[n-1]
		e.stack = e.stack[:n-2]
		e.stack = append(e.stack, res)
	default:
		return fmt.Errorf("error '%s': non-existing command", command)
	}

	return nil
}

// Process evaluates sequence of words or definition.
//
// Returns resulting stack state and an error.
func (e *Evaluator) Process(row string) ([]int, error) {
	commands := strings.Fields(strings.ToLower(row))

	for i := 0; i < len(commands); i++ {
		command := commands[i]
		if command == ":" {
			c, err := e.createCommand(commands[i:])
			if err != nil {
				return []int{}, err
			}

			i += c
		} else {
			if err := e.execCommand(commands[i]); err != nil {
				return []int{}, err
			}
		}
	}

	return e.stack, nil
}
