package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/fox-one/fox-data/cmd"
	"github.com/fox-one/fox-data/server"
	"github.com/fox-one/gin-contrib/session"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

var (
	// NAME name
	NAME = "pubsrc"
	// VERSION version
	VERSION = "null"
	// BUILD build
	BUILD = "null"
)

func main() {
	app := &cli.App{
		Name:        NAME,
		Version:     VERSION + "." + BUILD,
		Description: "Fox Portfolio pubsrc",
		Flags: []cli.Flag{
			&cli.BoolFlag{Name: "debug, d"},
		},
	}

	s, err := newSession()
	if err != nil {
		panic(err)
	}

	ctx := s.Context()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	s = s.WithContext(ctx)
	defer s.Close()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-quit
		log.Debug("quit app...")
		cancel()

		select {
		case <-time.After(time.Second * 3):
			log.Fatal("quit app timeout")
		}
	}()

	app.Before = func(c *cli.Context) error {
		if c.Bool("debug") {
			log.SetLevel(log.DebugLevel)
		}

		return nil
	}

	app.Commands = append(app.Commands, cli.Command{
		Name: "hc",
		Action: func(c *cli.Context) error {
			if err := s.MysqlRead().DB().Ping(); err != nil {
				return err
			}

			if err := s.MysqlWrite().DB().Ping(); err != nil {
				return err
			}

			return nil
		},
	})

	app.Commands = append(app.Commands, cli.Command{
		Name: "setdb",
		Action: func(c *cli.Context) error {
			return session.Setdb(s.Session)
		},
	})

	app.Commands = append(app.Commands, cli.Command{
		Name:  "server",
		Flags: []cli.Flag{cli.IntFlag{Name: "port", Value: 7000}},
		Action: func(c *cli.Context) error {
			return server.Run(s.Session, c.Int("port"))
		},
	})

	app.Commands = append(app.Commands, cli.Command{
		Name: "slug",
		Action: func(c *cli.Context) error {
			return cmd.MatchSlug(s.Session)
		},
	})

	app.Commands = append(app.Commands, cli.Command{
		Name: "price",
		Action: func(c *cli.Context) error {
			return cmd.CrawlHistoryPrices(s.Session)
		},
	})

	if err := app.Run(os.Args); err != nil {
		log.Errorf("app exit with error: %s", err)
		os.Exit(1)
	}
}
