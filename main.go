package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/template/html/v2"
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

	log.Println("Starting pinger")

	conf, err := loadConfigFromFile("config.json")
	if err != nil {
		log.Panic(err)
	}

	ctx, cancelFunc := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer cancelFunc()

	for _, def := range conf.Apps {
		startChecker(ctx, def)
	}

	engine := html.New("./views", ".html")

	app := fiber.New(fiber.Config{
		Views: engine,
	})

	app.Get("/", func(c *fiber.Ctx) error {
		results := checkOnAll(conf.Apps)
		return c.Render("index", results)
	})

	go func() {
		app.Listen(":9111")
	}()

	<-ctx.Done()

	shutdownErr := app.ShutdownWithTimeout(time.Second * 60)
	if shutdownErr != nil {
		log.Println("HTTP Server shutdown error: ", shutdownErr)
	}

	log.Println("Shutting down pinger")
}
