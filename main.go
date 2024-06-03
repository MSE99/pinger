package main

import (
	"flag"
	"fmt"
	"log"
)

func main() {
	genConfigFlag := flag.Bool("config", false, "If this flag is passed to pinger, it will generate a config file.")

	flag.Parse()

	if *genConfigFlag {
		fmt.Println("✨ Generating default config...")

		err := storeDefaultConfigIn("config.json")
		if err != nil {
			log.Panic(err)
		}
		return
	}
}
