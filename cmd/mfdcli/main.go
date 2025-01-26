package main

import (
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/kmpm/go-x52pro/public/x52pro"
)

var (
	x *x52pro.X52Pro
)

func logging() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	slog.SetDefault(logger)
}

func main() {
	logging()
	x, err := x52pro.New()
	if err != nil {
		fmt.Println("Error: ", err)
		return
	}
	defer x.Close()
	// check := func(err error) {
	// 	if err != nil {
	// 		fmt.Println(err)
	// 		x.Close()
	// 		os.Exit(1)
	// 	}
	// }

	fmt.Println("Hello, World!")
	x.AddPage("page1", true)
	time.Sleep(2 * time.Second)
	// check(p.SetLine(0, "p1 l0"))
	// check(p.SetLine(1, "p1 l1"))
	// check(p.SetLine(2, "p1 l2"))

	x.AddPage("page 2", false)
	// check(p2.SetLine(0, "page 2, line 0"))
	// check(p2.SetLine(1, "page 2, line 1"))
	// check(p2.SetLine(2, "page 2, line 2"))

	// wait for ctrl-c
	done := make(chan os.Signal, 1)
	signal.Notify(done, syscall.SIGINT, syscall.SIGTERM)
	fmt.Println("Blocking, press ctrl+c to continue...")
	<-done // Will block here until user hits ctrl+c

}
