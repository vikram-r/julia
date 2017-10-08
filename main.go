package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Config struct {
	w, h, cX, cY, moveX, moveY float64
	max                        int
}

func main() {
	w, h, err := terminalDimensions()
	if err != nil {
		panic(err)
	}
	cfg := Config{
		w:     float64(w),
		h:     float64(h),
		cX:    -0.7,
		cY:    0.27015,
		max:   255,
		moveX: 0,
		moveY: 0,
	}

	res := make(chan result, 20)

	// start calculating julia sets
	go worker(cfg, res)

	// holds the output for each zoom value
	cache := map[float64]string{}

	// mutex protects the cache
	var mutex = &sync.Mutex{}
	// readWG ensures that reads from cache don't get starved
	var readWG sync.WaitGroup

	// start listening for results from the worker
	go cacheWriter(cache, res, readWG, mutex)

	ticker := time.NewTicker(time.Duration(150) * time.Millisecond)
	defer ticker.Stop()

	currZoom := 1.0
	for range ticker.C {
		// reads have priority
		readWG.Add(1)
		mutex.Lock()

		var output string
		output, ok := cache[currZoom]
		delete(cache, currZoom)
		mutex.Unlock()
		readWG.Done()

		if !ok {
			// cache miss
			output = julia(cfg, currZoom)
		}
		fmt.Print(output)

		currZoom += .01
	}
}

// result is the message sent over the channel
type result struct {
	output string
	zoom   float64
}

// worker performs julia set calculations, and publishes results onto a buffered channel (to prevent oom)
func worker(cfg Config, out chan<- result) {
	zoom := 1.0
	for {
		out <- result{
			output: julia(cfg, zoom),
			zoom:   zoom,
		}
		zoom += .01
	}
}

// cacheWriter listens on the channel, and writes all results to the map
func cacheWriter(cache map[float64]string, in <-chan result, readWG sync.WaitGroup, mutex *sync.Mutex) {
	for r := range in {
		readWG.Wait()
		mutex.Lock()

		cache[r.zoom] = r.output

		mutex.Unlock()
	}
}

func julia(cfg Config, zoom float64) string {
	canvas := makeCanvas(int(cfg.w), int(cfg.h))
	for y := 0.0; y < cfg.h; y++ {
		for x := 0.0; x < cfg.w; x++ {
			zx := 1.5*(x-cfg.w/2)/(0.5*zoom*cfg.w) + cfg.moveX
			zy := 1.0*(y-cfg.h/2)/(0.5*zoom*cfg.h) + cfg.moveY
			done := func() bool {
				return (zx*zx + zy*zy) >= 4
			}
			for i := 0; i < cfg.max && !done(); i++ {
				tmp := zx*zx - zy*zy + cfg.cX
				zy, zx = 2.0*zx*zy+cfg.cY, tmp
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
