package main

import (
	"os"
	"fmt"

	"github.com/whitecypher/vgo/app"
)

func main() {
	vgo := app.New(MustGetwd(), os.Getenv("GOPATH"))
	err := vgo.Run(os.Args)
	if err != nil {
		fmt.Println(err)
	}
}
