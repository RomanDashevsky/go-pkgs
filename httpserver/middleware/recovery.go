package middleware

import (
	"fmt"
	"runtime/debug"
	"strings"

	"github.com/gofiber/fiber/v2"
	fiberRecover "github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/rdashevsky/go-pkgs/logger"
)

func buildPanicMessage(ctx *fiber.Ctx, err interface{}) string {
	var result strings.Builder

	result.WriteString(ctx.IP())
	result.WriteString(" - ")
	result.WriteString(ctx.Method())
	result.WriteString(" ")
	result.WriteString(ctx.OriginalURL())
	result.WriteString(" PANIC DETECTED: ")
	result.WriteString(fmt.Sprintf("%v\n%s\n", err, debug.Stack()))

	return result.String()
}

func logPanic(l logger.LoggerI) func(c *fiber.Ctx, err interface{}) {
	return func(ctx *fiber.Ctx, err interface{}) {
		l.Error(buildPanicMessage(ctx, err))
	}
}

// Recovery returns a Fiber middleware that recovers from panics.
// It logs panic details including stack trace and returns a 500 Internal Server Error.
func Recovery(l logger.LoggerI) func(c *fiber.Ctx) error {
	return fiberRecover.New(fiberRecover.Config{
		EnableStackTrace:  true,
		StackTraceHandler: logPanic(l),
	})
}
