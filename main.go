package main

import (
	"fmt"
	"os"
	"runtime/debug"
	"os/exec"
	"strings"
	"strconv"
)

func main() {
	// TODO temporary for debugging purposes
	defer func() {
		if r := recover(); r != nil {
			fmt.Println(os.Stderr, "error: ", r)
			fmt.Printf("trace: %s", debug.Stack())
			os.Exit(1)
		}
	}()

	w, h, err := terminalDimensions()
	if err != nil {
		panic(err)
	}
	fmt.Println(w, ",", h)
}

func terminalDimensions() (width, height int, e error) {
	stty := exec.Command("stty", "size")
	stty.Stdin = os.Stdin
	if r, err := stty.Output(); err != nil {
		return 0, 0, err
	} else {
		//h w\n
		p := strings.Split(strings.TrimSpace(string(r)), " ")

		if height, err = strconv.Atoi(p[0]); err != nil {
			return 0, 0, err
		}
		if width, err = strconv.Atoi(p[1]); err != nil {
			return 0, 0, err
		}
		return width, height, nil
	}
}