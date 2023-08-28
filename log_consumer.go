package main

import (
	"context"
	"errors"
	"github.com/testcontainers/testcontainers-go"
	"log"
	"strings"
	"time"
)

type LogConsoleConsumer struct {
	name                          string
	logger                        *log.Logger
	isInitializationComplete      bool
	showOutput                    bool
	containerInitializationString string
}

func NewLogConsoleConsumer(name string, logger *log.Logger, containerInitializationString string, showOutput bool) *LogConsoleConsumer {
	result := &LogConsoleConsumer{
		name:                          name,
		logger:                        logger,
		containerInitializationString: containerInitializationString,
		showOutput:                    showOutput,
	}

	if containerInitializationString == "" {
		result.isInitializationComplete = true
	}

	return result
}

func (g *LogConsoleConsumer) Accept(l testcontainers.Log) {
	if g.containerInitializationString != "" && strings.Contains(string(l.Content), g.containerInitializationString) {
		g.isInitializationComplete = true
	}

	if g.showOutput {
		g.logger.Printf("%s - %v", g.name, string(l.Content))
	}
}

func (g *LogConsoleConsumer) WaitForContainerToStart() error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)

	defer cancel()

	var resp = make(chan bool)

	go func() {
		for {
			if !g.isInitializationComplete {
				time.Sleep(5 * time.Second)
			} else {
				resp <- true
				return
			}
		}
	}()

	select {
	case <-ctx.Done():
		return errors.New("context timeout, stopping the operation")
	case <-resp:
		g.logger.Println("container got initialized")
		return nil
	}
}
