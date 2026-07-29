// Package router provides routing capabilities.
package router

import (
	"fmt"
	"log"
	"time"
)

// LoggerMiddleware is a built-in middleware that logs every incoming update
// and the time it took to process it.
func LoggerMiddleware() MiddlewareFunc {
	return func(next HandlerFunc) HandlerFunc {
		return func(c *Context) error {
			start := time.Now()

			sender := "Unknown"
			if c.Update.From != nil {
				if c.Update.From.Login != "" {
					sender = string(c.Update.From.Login)
				} else if c.Update.From.DisplayName != "" {
					sender = c.Update.From.DisplayName
				}
			}

			log.Printf("[Router] Received update ID: %d from: %s", c.Update.UpdateID, sender)

			// Call the next handler in the chain
			err := next(c)

			if err != nil {
				log.Printf("[Router] ❌ Error processing update ID: %d: %v", c.Update.UpdateID, err)
			} else {
				log.Printf("[Router] ✅ Processed update ID: %d in %v", c.Update.UpdateID, time.Since(start))
			}
			return err
		}
	}
}

// RecoverMiddleware is a built-in middleware that catches panics inside handlers
// to prevent the entire bot application from crashing.
func RecoverMiddleware() MiddlewareFunc {
	return func(next HandlerFunc) HandlerFunc {
		return func(c *Context) (err error) {
			defer func() {
				if r := recover(); r != nil {
					log.Printf("[Router] 🚨 PANIC RECOVERED in handler for update ID %d: %v", c.Update.UpdateID, r)
					err = fmt.Errorf("panic recovered: %v", r)
				}
			}()

			// Call the next handler in the chain
			return next(c)
		}
	}
}
