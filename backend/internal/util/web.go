package util

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

const (
	ContextRequestID = "request_id"
	ContextUserID    = "user_id"
	ContextUsername  = "username"
	ContextRole      = "role"
)

type AppError struct {
	Status  int
	Code    string
	Message string
	Cause   error
}

func (e *AppError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Cause)
	}
	return e.Message
}

func (e *AppError) Unwrap() error { return e.Cause }

func BadRequest(code, message string, cause error) *AppError {
	return &AppError{Status: http.StatusBadRequest, Code: code, Message: message, Cause: cause}
}

func Unauthorized(message string) *AppError {
	return &AppError{Status: http.StatusUnauthorized, Code: "AUTH_REQUIRED", Message: message}
}

func Forbidden(message string) *AppError {
	return &AppError{Status: http.StatusForbidden, Code: "FORBIDDEN", Message: message}
}

func NotFound(entity string) *AppError {
	return &AppError{Status: http.StatusNotFound, Code: "NOT_FOUND", Message: entity + " not found"}
}

func Conflict(code, message string, cause error) *AppError {
	return &AppError{Status: http.StatusConflict, Code: code, Message: message, Cause: cause}
}

func Unprocessable(code, message string, cause error) *AppError {
	return &AppError{Status: http.StatusUnprocessableEntity, Code: code, Message: message, Cause: cause}
}

func Internal(cause error) *AppError {
	return &AppError{Status: http.StatusInternalServerError, Code: "INTERNAL_ERROR", Message: "internal service error", Cause: cause}
}

type Envelope struct {
	Data      any    `json:"data,omitempty"`
	Error     *Error `json:"error,omitempty"`
	RequestID string `json:"request_id"`
}

type Error struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type Page struct {
	Items any   `json:"items"`
	Total int64 `json:"total"`
	Page  int   `json:"page"`
	Size  int   `json:"size"`
}

func OK(c *gin.Context, data any) {
	c.JSON(http.StatusOK, Envelope{Data: data, RequestID: RequestID(c)})
}

func Created(c *gin.Context, data any) {
	c.JSON(http.StatusCreated, Envelope{Data: data, RequestID: RequestID(c)})
}

func NoContent(c *gin.Context) { c.Status(http.StatusNoContent) }

func Fail(c *gin.Context, err error) {
	var appErr *AppError
	if errors.As(err, &appErr) {
		c.JSON(appErr.Status, Envelope{Error: &Error{Code: appErr.Code, Message: appErr.Message}, RequestID: RequestID(c)})
		return
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusNotFound, Envelope{Error: &Error{Code: "NOT_FOUND", Message: "resource not found"}, RequestID: RequestID(c)})
		return
	}
	c.JSON(http.StatusInternalServerError, Envelope{Error: &Error{Code: "INTERNAL_ERROR", Message: "internal service error"}, RequestID: RequestID(c)})
}

func BindJSON(c *gin.Context, target any) bool {
	if err := c.ShouldBindJSON(target); err != nil {
		Fail(c, BadRequest("INVALID_INPUT", "request body failed validation", err))
		return false
	}
	return true
}

func ParamID(c *gin.Context) (uint, bool) {
	raw := c.Param("id")
	value, err := strconv.ParseUint(raw, 10, 64)
	if err != nil || value == 0 {
		Fail(c, BadRequest("INVALID_ID", "path id must be a positive integer", err))
		return 0, false
	}
	return uint(value), true
}

func QueryPage(c *gin.Context) (int, int) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "50"))
	if page < 1 {
		page = 1
	}
	if size < 1 {
		size = 20
	}
	if size > 200 {
		size = 200
	}
	return page, size
}

func RequestID(c *gin.Context) string { return c.GetString(ContextRequestID) }
func Username(c *gin.Context) string  { return c.GetString(ContextUsername) }
func Role(c *gin.Context) string      { return c.GetString(ContextRole) }

func UserID(c *gin.Context) uint {
	value, exists := c.Get(ContextUserID)
	if !exists {
		return 0
	}
	id, _ := value.(uint)
	return id
}

func SearchTerm(c *gin.Context) string {
	return strings.TrimSpace(c.Query("search"))
}
