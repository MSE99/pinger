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
)

type flags struct {
	getStatusOnly bool
	genConfigFlag bool
}

func main() {
	opts := flags{}

	flag.BoolVar(&opts.getStatusOnly, "status", false, "If set to true, will fetch the status of the services, report errors and immediately exit.")
	flag.BoolVar(&opts.genConfigFlag, "config", false, "If this flag is passed to pinger, it will generate a config file.")
	flag.Parse()

	ctx, cancelFunc := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer cancelFunc()
	startHTTPServerAndCheckers(ctx, opts)
}

var (
	socketsGuard                                 = &sync.Mutex{}
	sockets      map[chan statusCheckResult]bool = map[chan statusCheckResult]bool{}
)

func startHTTPServerAndCheckers(ctx context.Context, opts flags) {
	if opts.genConfigFlag {
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

	if opts.getStatusOnly {
		checkOnAll(conf.Apps, context.Background())
		return
	}

	for _, def := range conf.Apps {
		startChecker(ctx, def)
	}

	app := fiber.New()

	app.Static("/", "./www")

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
		log.Println("Websocket watcher connected, starting sender goroutine.")

		messages := make(chan statusCheckResult, 10)

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

		disconnectedChan := make(chan bool)

		go func() {
			for {
				_, _, err := c.ReadMessage()

				if err != nil {
					disconnectedChan <- true
					return
				}

				select {
				case <-ctx.Done():
					return
				default:
					continue
				}
			}
		}()

		defer log.Println("Shutting down sender & reader goroutines.")

		checkResults := checkOnAll(conf.Apps, ctx)
		_ = c.WriteJSON(checkResults)

		for {
			select {
			case <-ctx.Done():
				return
			case <-disconnectedChan:
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
		port := conf.Port
		if port == 0 {
			port = 9111
		}

		app.Listen(fmt.Sprintf(":%d", port))
	}()

	<-ctx.Done()

	shutdownErr := app.ShutdownWithTimeout(time.Second * 60)
	if shutdownErr != nil {
		log.Println("HTTP Server shutdown error: ", shutdownErr)
	}

	log.Println("Shutting down pinger")
}
