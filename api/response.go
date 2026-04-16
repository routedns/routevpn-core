package api

import "github.com/gofiber/fiber/v3"

// Envelope is the standard JSON response wrapper.
type Envelope struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   *APIError   `json:"error,omitempty"`
	Meta    *Meta       `json:"meta,omitempty"`
}

type APIError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type Meta struct {
	Page    int   `json:"page"`
	PerPage int   `json:"per_page"`
	Total   int64 `json:"total"`
}

func OK(c fiber.Ctx, data interface{}) error {
	return c.JSON(Envelope{Success: true, Data: data})
}

func OKMeta(c fiber.Ctx, data interface{}, meta *Meta) error {
	return c.JSON(Envelope{Success: true, Data: data, Meta: meta})
}

func Created(c fiber.Ctx, data interface{}) error {
	return c.Status(201).JSON(Envelope{Success: true, Data: data})
}

func Fail(c fiber.Ctx, status int, msg string) error {
	return c.Status(status).JSON(Envelope{
		Success: false,
		Error:   &APIError{Code: status, Message: msg},
	})
}

func BadRequest(c fiber.Ctx, msg string) error   { return Fail(c, 400, msg) }
func Unauthorized(c fiber.Ctx, msg string) error { return Fail(c, 401, msg) }
func Forbidden(c fiber.Ctx, msg string) error    { return Fail(c, 403, msg) }
func NotFound(c fiber.Ctx, msg string) error     { return Fail(c, 404, msg) }
func Internal(c fiber.Ctx, msg string) error     { return Fail(c, 500, msg) }
