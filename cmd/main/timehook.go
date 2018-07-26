package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/timehook/cli-client/timehook"
)

func main() {
	URL := flag.String("url", "https://httpstat.us/200", "webhook URL")
	body := flag.String("body", `{"msg" : "from timehook client"}`, "webhook body in JSON")
	sec := flag.Int("sec", 5, "delay in seconds")
	flag.Parse()

	if os.Getenv("TIMEHOOK_KEY") == "" {
		fmt.Println("TIMEHOOK_KEY enviroment variable not defined")
		os.Exit(1)
	}

	client := timehook.New(os.Getenv("TIMEHOOK_KEY"), http.DefaultClient)
	proc := client.RegisterAndPoll(*URL, *body, *sec, 1*time.Second)
	for msg := range proc.C {
		fmt.Fprint(os.Stdout, msg)
	}

	if proc.IsSucceeded() {
		os.Exit(0)
	}
	os.Exit(1)
}
