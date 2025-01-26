// SPDX-FileCopyrightText: 2025 Peter Magnusson <me@kmpm.se>
//
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/kmpm/go-x52pro/internal/helper"
	"github.com/kmpm/go-x52pro/public/x52pro"
)

var (
	x *x52pro.X52Pro
)

func check(err error) {
	if err != nil {
		fmt.Println(err)
		x.Close()
		os.Exit(1)
	}
}

func logging() {
	opts := &helper.FilteredHandlerOptions{
		SlogOpts: slog.HandlerOptions{
			Level: slog.LevelInfo,
		},
	}
	handler := helper.NewFilteredHandler(*opts)
	// handler := slog.NewTextHandler(os.Stdout, opts)
	logger := slog.New(handler)
	slog.SetDefault(logger)

	// classic
	// logger := slog.NewLogLogger(handler, slog.LevelError)
}

func main() {
	logging()
	var err error
	x, err = x52pro.New()
	if err != nil {
		fmt.Println("Error: ", err)
		return
	}
	defer x.Close()

	fmt.Println("Page 1")

	// NOTE: This will NOT call the OnPageChanged callback
	p, err := x.AddPage("page1", true)
	check(err)
	check(p.SetLine(0, "p1 l0"))
	check(p.SetLine(1, "p1 l1"))
	// check(p.SetLine(2, "p1 l2"))

	fmt.Println("Page 2")
	p2, err := x.AddPage("page2", false)
	check(err)
	check(p2.SetLine(0, "page 2, line 0"))
	check(p2.SetLine(1, "page 2, line 1"))
	check(p2.SetLine(2, "page 2, line 2"))

	fmt.Println("Type", x.GetType())

	go func() {
		loopNo := 0
		var err error
		defer func() {
			check(err)
		}()

		for {
			loopNo++
			if err = x.SetString("page1", 2, fmt.Sprintf("Loop: %d", loopNo)); err != nil {
				slog.Error("error setting string in loop", "error", err)
			}
			time.Sleep(250 * time.Millisecond)

		}
	}()

	// wait for ctrl-c
	done := make(chan os.Signal, 1)
	signal.Notify(done, syscall.SIGINT, syscall.SIGTERM)
	fmt.Println("Blocking, press ctrl+c to continue...")
	<-done // Will block here until user hits ctrl+c

}
