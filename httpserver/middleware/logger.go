// Package middleware provides HTTP middleware for Fiber applications.
// It includes logging, recovery, and other common middleware functionality.
package middleware

import (
	"strconv"
	"strings"

	"github.com/rdashevsky/go-pkgs/logger"

	"github.com/gofiber/fiber/v2"
)

func buildRequestMessage(ctx *fiber.Ctx) string {
	var result strings.Builder

	result.WriteString(ctx.IP())
	result.WriteString(" - ")
	result.WriteString(ctx.Method())
	result.WriteString(" ")
	result.WriteString(ctx.OriginalURL())
	result.WriteString(" - ")
	result.WriteString(strconv.Itoa(ctx.Response().StatusCode()))
	result.WriteString(" ")
	result.WriteString(strconv.Itoa(len(ctx.Response().Body())))

	return result.String()
}

// Logger returns a Fiber middleware that logs HTTP requests.
// It logs the client IP, method, URL, status code, and response body size for each request.
func Logger(l logger.LoggerI) func(c *fiber.Ctx) error {
	return func(ctx *fiber.Ctx) error {
		err := ctx.Next()

		l.Info(buildRequestMessage(ctx))

		return err
	}
}
