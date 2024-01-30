// Package logging is a shared, minimal logging module that wraps Zerolog
package logging

import (
	"errors"
	"fmt"
	"io"

	"github.com/rs/zerolog"
)

//go:generate mockgen -source=logging.go -destination mock_logger.go -package mocks

var invalidRate = errors.New("invalid rate")

// LogLevel represents available logging levels
type LogLevel int

const (
	DebugLevel LogLevel = iota
	InfoLevel
	WarningLevel
	ErrorLevel
)

// Logger represents the builtin logger for Maelstrom and some basic access and tuning methods
type Logger interface {
	// Level returns the log level to log at.
	//level(ll LogLevel) *zerolog.Event

	// Log logs a message with the provided msg and a series of key value pairs
	// keys are string, but values are cast to strings
	Log(level LogLevel, msg string, kv ...KV)

	// LogWithError allows for an additional error to be logged alongside the msg and key value pairs
	LogWithError(level LogLevel, msg string, e error, kv ...KV)
}

type logger struct {
	logLevel LogLevel
	writer   zerolog.Logger
}

func (l logger) level(ll LogLevel) *zerolog.Event {
	switch ll {
	case DebugLevel:
		return l.writer.Debug()
	case InfoLevel:
		return l.writer.Info()
	case WarningLevel:
		return l.writer.Warn()
	case ErrorLevel:
		return l.writer.Error()
	default:
		return l.writer.Info()
	}
}

type KV struct {
	Key   string
	Value any
}

func (l logger) Log(level LogLevel, msg string, kv ...KV) {
	chain := l.level(level)
	if chain == nil {
		return
	}
	for _, v := range kv {
		chain.Str(v.Key, fmt.Sprintf("%v", v.Value))
	}
	chain.Msg(msg)
}

func (l logger) LogWithError(level LogLevel, msg string, e error, kv ...KV) {
	chain := l.level(level).Err(e)
	if chain == nil {
		return
	}
	for _, v := range kv {
		chain.Str(v.Key, fmt.Sprintf("%v", v.Value))
	}
	chain.Msg(msg)
}

func NewLogger(level LogLevel, rate uint32, writer io.Writer, serviceKey, service string) (Logger, error) {
	if rate < 1 {
		return nil, invalidRate
	}
	return logger{
		logLevel: level,
		writer: zerolog.New(writer).
			Sample(&zerolog.BasicSampler{N: rate}).
			Level(zerolog.Level(level)).
			With().
			Str(serviceKey, service).
			Timestamp().
			Logger(),
	}, nil
}
