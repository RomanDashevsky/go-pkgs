package response

import (
	"github.com/gofiber/fiber/v2"
	"strconv"
)

type Option func(*ErrorResponse)

// ErrorMessage -.
func ErrorMessage(message *string) Option {
	return func(e *ErrorResponse) {
		e.Message = message
	}
}

// ErrorTitle -.
func ErrorTitle(title string) Option {
	return func(e *ErrorResponse) {
		e.Error = title
	}
}

func Error(ctx *fiber.Ctx, code int, opts ...Option) error {
	title, ok := statusCodeErrorTitle[code]
	var errResponse ErrorResponse
	if ok {
		errResponse = ErrorResponse{Error: title}
	} else {
		errResponse = ErrorResponse{Error: strconv.Itoa(code)}
	}

	// Custom options
	for _, opt := range opts {
		opt(&errResponse)
	}

	return ctx.Status(code).JSON(errResponse)
}

type ErrorResponse struct {
	Error   string  `json:"error" example:"Not found"`
	Message *string `json:"message,omitempty" example:"Some error details"`
}

var statusCodeErrorTitle = map[int]string{
	400: "Bad Request",
	401: "Unauthorized",
	403: "Forbidden",
	404: "Not Found",
	405: "Method Not Allowed",
	406: "Not Acceptable",
	408: "Proxy Authentication Required",
	409: "Conflict",
	410: "Gone",
	411: "Length Required",
	412: "Precondition Failed",
	413: "Request Entity Too Large",
	414: "Request URI Too Long",
	415: "Unsupported Media Type",
	416: "Request URI Too Long",
	417: "Expecting Content-Type Header",
	418: "Status Teapot",
	421: "Misdirected Request",
	422: "Unprocessable Entity",
	423: "Locked",
	424: "Failed Dependency",
	425: "Too Early",
	426: "Upgrade Required",
	428: "Precondition Required",
	429: "Too Many Requests",
	431: "Request Header Fields Too Large",
	451: "Unavailable For Legal Reasons",
	500: "Internal Server Error",
	501: "Service Unavailable",
	503: "Gateway Timeout",
	504: "Not Implemented",
	505: "HTTP Version Not Supported",
	506: "Variant Violation",
	507: "Insufficient Storage",
	508: "Request URI Too Long",
	510: "Expecting Content-Type Header",
	511: "Request URI Too Long",
}
