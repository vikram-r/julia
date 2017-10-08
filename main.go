package main

import (
	"fmt"
	"os"
	"os/exec"
	"runtime/debug"
	"strconv"
	"strings"
	"time"
	"bytes"
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
	cX, cY := -0.7, 0.27015

	waitTime := time.Duration(100) * time.Millisecond

	z := 1.0
	for {
		result := julia(float64(w), float64(h), cX, cY, 255, z, 0, 0)
		fmt.Print(result)
		z += .01
		time.Sleep(waitTime)
	}
}

func julia(w float64, h float64, cX float64, cY float64, max int, zoom float64, moveX float64, moveY float64) string {
	// TODO canvas is to improve readability, and can be removed if it affects performance
	canvas := makeCanvas(int(w), int(h))
	for y := 0.0; y < h; y++ {
		for x := 0.0; x < w; x++ {
			zx := 1.5 * (x - w/2) / (0.5 * zoom * w) + moveX
			zy := 1.0 * (y - h/2) / (0.5 * zoom * h) + moveY
			done := func() bool {
				return (zx*zx + zy*zy) >= 4
			}
			for i := 0; i < max && !done(); i++ {
				tmp := zx*zx - zy*zy + cX
				zy, zx = 2.0*zx*zy+cY, tmp
			}
			if !done() {
				canvas[int(y)][int(x)] = "*"
			} else {
				canvas[int(y)][int(x)] = " "
			}
		}
	}
	return canvasToString(canvas)
}

func canvasToString(canvas [][]string) string {
	var buf bytes.Buffer
	for y := range canvas {
		for x := range canvas[y] {
			buf.WriteString(canvas[y][x])
		}
		buf.WriteByte('\n')
	}
	return buf.String()
}

func makeCanvas(w int, h int) [][]string {
	canvas := make([][]string, h)
	for i := range canvas {
		canvas[i] = make([]string, w)
	}
	return canvas
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
