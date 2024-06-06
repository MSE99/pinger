package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/template/html/v2"
)

func main() {
	startHTTPServerAndCheckers(context.Background())
}

var (
	socketsGuard                      = &sync.Mutex{}
	sockets      map[chan string]bool = map[chan string]bool{}
)

func startHTTPServerAndCheckers(mainCtx context.Context) {
	getStatusOnly := flag.Bool("status", false, "If set to true, will fetch the status of the services, report errors and immediately exit.")
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

	conf, err := loadOrCreateConfigAt("config.json")
	if err != nil {
		log.Panic(err)
	}

	if *getStatusOnly {
		checkOnAll(conf.Apps, context.Background())
		return
	}

	ctx, cancelFunc := signal.NotifyContext(mainCtx, os.Interrupt, os.Kill)
	defer cancelFunc()

	for _, def := range conf.Apps {
		startChecker(ctx, def)
	}

	engine := html.New("./views", ".html")

	app := fiber.New(fiber.Config{
		Views: engine,
	})

	app.Get("/", func(c *fiber.Ctx) error {
		results := checkOnAll(conf.Apps, c.Context())
		return c.Render("index", results)
	})

	app.Use("/ws", func(c *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			c.Locals("allowed", true)
			return c.Next()
		}

		return fiber.ErrUpgradeRequired
	})

	app.Get("/ws", websocket.New(func(c *websocket.Conn) {
		log.Println("Websocket watcher connected, starting receiver goroutine.")

		messages := make(chan string, 10)

		defer func() {
			socketsGuard.Lock()
			defer socketsGuard.Unlock()

			delete(sockets, messages)
			close(messages)
		}()

		func() {
			socketsGuard.Lock()
			defer socketsGuard.Unlock()

			sockets[messages] = true
		}()

		for {
			select {
			case <-ctx.Done():
				return
			case s := <-messages:
				err := c.WriteJSON(s)
				if err != nil {
					log.Println(err)
				}
			}
		}
	}))

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
