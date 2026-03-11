package prompt

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

type Prompter interface {
	Prompt(label string) (string, error)
	PromptSecret(label string) (string, error)
}

type Terminal struct {
	in  io.Reader
	out io.Writer
}

func NewTerminal(in io.Reader, out io.Writer) *Terminal {
	return &Terminal{
		in:  in,
		out: out,
	}
}

func (t *Terminal) Prompt(label string) (string, error) {
	fmt.Fprint(t.out, label)
	reader := bufio.NewReader(t.in)
	value, err := reader.ReadString('\n')
	if err != nil && err != io.EOF {
		return "", err
	}
	return strings.TrimSpace(value), nil
}

func (t *Terminal) PromptSecret(label string) (string, error) {
	value, err := t.promptTTYSecret(label)
	if err == nil {
		return value, nil
	}

	fmt.Fprintln(t.out, "Token input will be visible.")
	return t.Prompt(label)
}

func (t *Terminal) promptTTYSecret(label string) (string, error) {
	tty, err := os.OpenFile("/dev/tty", os.O_RDWR, 0)
	if err != nil {
		return "", err
	}
	defer tty.Close()

	fmt.Fprint(tty, label)

	disable := exec.Command("stty", "-echo")
	disable.Stdin = tty
	disable.Stdout = io.Discard
	disable.Stderr = io.Discard
	if err := disable.Run(); err != nil {
		return "", err
	}

	defer func() {
		enable := exec.Command("stty", "echo")
		enable.Stdin = tty
		enable.Stdout = io.Discard
		enable.Stderr = io.Discard
		_ = enable.Run()
		fmt.Fprintln(tty)
	}()

	reader := bufio.NewReader(tty)
	value, err := reader.ReadString('\n')
	if err != nil && err != io.EOF {
		return "", err
	}
	return strings.TrimSpace(value), nil
}
