// SPDX-FileCopyrightText: 2020 Ethel Morgan
//
// SPDX-License-Identifier: MIT

package logger

import (
	"context"
	"fmt"
	"sync"

	log "github.com/sirupsen/logrus"
)

type (
	// log, ctx := logger.FromContext(ctx)
	// log.WithField("error", err).Warninging(...)
	// log.AddField("transport", udn)
	// log.Info(...)
	Logger interface {
		// AddField adds a field to the current logger.
		AddField(name string, value interface{})

		// WithField forks a logger, adding context.
		WithField(name string, value interface{}) Logger

		// WithError is a convenience method for one-off forks to log error messages under the key "error".
		WithError(err error) Logger

		// Fork returns a copy of the Logger and a fork of the context.Context to pass through.
		Fork(context.Context) (Logger, context.Context)

		Debug(string)
		Info(string)
		Warning(string)
		Error(string)
		Fatal(string)
	}

	logger struct {
		mu     sync.Mutex
		values map[string]interface{}
	}

	// contextKey is a separate type to prevent collisions with other packages.
	contextKey int
)

const (
	loggerKey contextKey = iota
)

func FromContext(ctx context.Context) (Logger, context.Context) {
	maybeLogger := ctx.Value(loggerKey)

	if maybeLogger == nil {
		l := Background()
		return l, context.WithValue(ctx, loggerKey, l)
	}

	if l, ok := maybeLogger.(*logger); ok {
		return l, ctx
	}

	panic(fmt.Sprintf("expected logger in context, found %+v", maybeLogger))
}
func Background() Logger {
	return &logger{
		values: map[string]interface{}{},
	}
}

func (l *logger) AddField(name string, value interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()

	// TODO: check it doesn't exist already?
	l.values[name] = value
}
func (l *logger) WithField(name string, value interface{}) Logger {
	clone, _ := l.Fork(context.Background())
	clone.AddField(name, value)
	return clone
}
func (l *logger) WithError(err error) Logger {
	return l.WithField("error", err)
}

func (l *logger) Fork(ctx context.Context) (Logger, context.Context) {
	l.mu.Lock()
	defer l.mu.Unlock()

	clone := &logger{
		values: map[string]interface{}{},
	}
	for k, v := range l.values {
		clone.values[k] = v
	}
	return clone, context.WithValue(ctx, loggerKey, clone)
}

func (l *logger) Debug(message string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	log.WithFields(log.Fields(l.values)).Debug(message)
}
func (l *logger) Info(message string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	log.WithFields(log.Fields(l.values)).Info(message)
}
func (l *logger) Warning(message string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	log.WithFields(log.Fields(l.values)).Warning(message)
}
func (l *logger) Error(message string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	log.WithFields(log.Fields(l.values)).Error(message)
}
func (l *logger) Fatal(message string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	log.WithFields(log.Fields(l.values)).Fatal(message)
}
