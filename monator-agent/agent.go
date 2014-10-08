package main

import (
	"flag"
	"fmt"
	"os"
	"time"

    "github.com/dullgiulio/monator/monator"
)

type Config struct {
	PrintAverages  time.Duration
	Verbose        bool
	DefaultHeaders map[string]string
}

var config Config

func (c *Config) LoadDefaults() {
	c.DefaultHeaders = make(map[string]string)
	c.DefaultHeaders["User-Agent"] = "Mozilla/5.0 (comptabile; Linux) Monator/1.0 HTML/5.0"
	c.DefaultHeaders["Accept"] = "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8"
	c.DefaultHeaders["Cache-Control"] = "no-cache"
	c.DefaultHeaders["Pragma"] = "no-cache"
}

func (c *Config) ParseFlags() []string {
	// Parse options from the arguments
	flag.DurationVar(&c.PrintAverages, "print-averages", time.Duration(0), "Print load averages at a fixed frequency")
	flag.BoolVar(&c.Verbose, "verbose", false, "Print all checks, even if status doesn't change")
	flag.Parse()

	return flag.Args()
}

func loadChecksFromJson(checks *monator.CheckContainer, files []string) error {
	for _, file := range files {
		if err := checks.LoadFromJson(file); err != nil {
			return fmt.Errorf("Error loading %s: %s", file, err)
		}
	}

	return nil
}

func main() {
	config.LoadDefaults()
	args := config.ParseFlags()

	if len(args) > 0 {
		checks := monator.NewCheckContainer()

		if err := loadChecksFromJson(checks, args); err != nil {
			// TODO: Just print the error to stderr
			panic(err)
			os.Exit(3)
		} else {
			checks.SetHeaders(config.DefaultHeaders)

            ch := make(chan *monator.CheckResult)

			checks.StartChecks(ch)
			checks.EmitAverages(ch, config.PrintAverages)

			checks.ReadChannel(ch, config.Verbose)

			// Never reached, really.
			close(ch)
		}
	} else {
		fmt.Printf("Usage: %s [OPTIONS] <json-configuration...>\n", os.Args[0])
		os.Exit(2)
	}

	os.Exit(0)
}
