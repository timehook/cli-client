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
	delay := flag.Int("delay", 5, "delay in seconds")
	flag.Parse()

	if os.Getenv("TIMEHOOK_KEY") == "" {
		fmt.Println("TIMEHOOK_KEY enviroment variable not defined")
		os.Exit(1)
	}

	client := timehook.New(os.Getenv("TIMEHOOK_KEY"), http.DefaultClient)
	out, errc, succ := client.RegisterAndPoll(*URL, *body, *delay, 1*time.Second)
	for {
		select {
		case msg := <-out:
			fmt.Fprint(os.Stdout, msg)
		case msg := <-errc:
			fmt.Fprint(os.Stderr, msg)
		case isSucc := <-succ:
			if isSucc {
				os.Exit(0)
			}
			os.Exit(1)
		}
	}
}
