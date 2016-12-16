package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"qvl.io/ghbackup/ghbackup"
)

const (
	// Printed for -help, -h or with wrong number of arguments
	usage = `Usage: %s account directory

  account    github user or organization name to get the repositories from
  directory  path to save the repositories to

`
	secretUsage = `Authentication secret for Github API.
	Can use the users password or a personal access token (https://github.com/settings/tokens).
	Authentication increases rate limiting (https://developer.github.com/v3/#rate-limiting) and enables backup of private repositories.`
)

// Get command line arguments and start updating repositories
func main() {
	// Flags
	secret := flag.String("secret", "", secretUsage)
	verboseFlag := flag.Bool("verbose", false, "print progress information")

	// Parse args
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, usage, os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()
	args := flag.Args()
	if len(args) != 2 {
		flag.Usage()
		os.Exit(1)
	}

	// Log errors with line numbers
	errLogger := log.New(os.Stderr, "", log.LstdFlags|log.Lshortfile|log.LUTC)

	// Logger for verbose mode
	var verboseLogger *log.Logger
	if *verboseFlag {
		verboseLogger = log.New(os.Stderr, "", log.LstdFlags|log.LUTC)
	} else {
		verboseLogger = log.New(ioutil.Discard, "", 0)
	}

	updates := make(chan ghbackup.Update)

	go func() {
		for u := range updates {
			switch u.Type {
			case ghbackup.UErr:
				errLogger.Println(u.Message)
			case ghbackup.UInfo:
				verboseLogger.Println(u.Message)
			}
		}
	}()

	err := ghbackup.Run(ghbackup.Config{
		Account: args[0],
		Dir:     args[1],
		Secret:  *secret,
		Updates: updates,
	})
	if err != nil {
		errLogger.Println(err)
		os.Exit(1)
	}
}
