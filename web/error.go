package web

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-chi/chi/middleware"
	en "github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	ja_translations "github.com/go-playground/validator/v10/translations/ja"
)

// ErrType is a kind of error understood by the client.
type ErrType string

// Error types for the API.
const (
	ErrTypeForbidden    ErrType = "ForbiddenError"
	ErrTypeUnauthorized ErrType = "UnauthorizedError"
	ErrTypeNotFound     ErrType = "NotFoundError"
	ErrTypeValidation   ErrType = "ValidationError"
	ErrRequiredQuery    ErrType = "RequiredQuery"
	ErrTypeInternal     ErrType = "InternalError"
	ErrTypeDuplicate    ErrType = "DuplicateError"
)

// Error Message
const (
	ErrMessageNotFoundUser = "ユーザが見つかりません。"
)

// StatusCode returns the HTTP status code that corresponds to this error type.
func (e ErrType) StatusCode() int {
	switch e {
	case ErrTypeForbidden:
		return http.StatusForbidden
	case ErrTypeUnauthorized:
		return http.StatusUnauthorized
	case ErrTypeNotFound:
		return http.StatusNotFound
	case ErrTypeValidation,
		ErrTypeDuplicate:
		return http.StatusBadRequest
	}
	return http.StatusInternalServerError
}
func (e ErrType) String() string {
	return string(e)
}

// ErrorResponse is an error message sent to clients.
type ErrorResponse struct {
	Type       string `json:"type"`
	Message    string `json:"message"`
	SysMessage string `json:"sys_message"`
}

// Error is a standardized wrapper for errors that contains API metadata.
type Error struct {
	Type    ErrType
	Message string
	Cause   error
}

func GetErrorMessages(err error) string {
	en := en.New()
	uni := ut.New(en, en)
	trans, _ := uni.GetTranslator("en")
	ja_translations.RegisterDefaultTranslations(validate, trans)
	errs := err.(validator.ValidationErrors)
	var errorsString []string
	for _, e := range errs.Translate(trans) {
		errorsString = append(errorsString, e)
	}
	return strings.Join(errorsString, ",")
}

// MarshalJSON marshals this Error as an ErrorResponse.
func (err Error) MarshalJSON() ([]byte, error) {
	resp := ErrorResponse{
		Type:    err.Type.String(),
		Message: err.Message,
	}
	switch err.Cause.(type) {
	case validator.ValidationErrors:
		if err.Cause != nil {
			resp.Message = GetErrorMessages(err.Cause)
		}
	default:
		resp.Message = err.Message
	}
	if err.Cause != nil {
		resp.SysMessage = err.Cause.Error()
	}
	return json.Marshal(resp)
}
func (err Error) Error() string {
	if err.Cause != nil {
		return fmt.Sprint(err.Type, err.Cause)
	}
	return fmt.Sprint(err.Type, err.Message)
}

// Unwrap implements the errors.Unwrapper interface and returns the original cause.
func (err Error) Unwrap() error {
	return err.Cause
}

// ErrorFunc represents a convention to wrap errors with standard API metadata.
type ErrorFunc func(error) Error

func ErrForbidden(cause error) Error {
	return Error{
		Type:    ErrTypeForbidden,
		Message: "リソースに対する権限がありません",
		Cause:   cause,
	}
}
func ErrUnauthorized(cause error) Error {
	return Error{
		Type:    ErrTypeUnauthorized,
		Message: "認情情報に不正があります",
		Cause:   cause,
	}
}
func ErrNotFound(cause error) Error {
	return Error{
		Type:    ErrTypeNotFound,
		Message: "リソースが見つかりません",
		Cause:   cause,
	}
}
func ErrRequiredQueryNotFound(cause error) Error {
	return Error{
		Type:    ErrRequiredQuery,
		Message: "クエリが見つかりません",
		Cause:   cause,
	}
}
func ErrValidation(cause error) Error {
	return Error{
		Type:    ErrTypeValidation,
		Message: "入力データに不正があります",
		Cause:   cause,
	}
}
func ErrValidationWithMessage(cause error, message string) Error {
	return Error{
		Type:    ErrTypeValidation,
		Message: message,
		Cause:   cause,
	}
}
func ErrValidationWithMessageAndErrType(cause error, message string, errType ErrType) Error {
	return Error{
		Type:    errType,
		Message: message,
		Cause:   cause,
	}
}
func ErrDuplicated() Error {
	return Error{
		Type:    ErrTypeDuplicate,
		Message: "データが重複しています",
	}
}

var ErrInvalidPasswordToken = Error{
	Type:    ErrTypeInternal,
	Message: "パスワードリセットトークンが不正です",
}

// PanicHandler recovers from panics and sends the appropriate JSON-formatted error message.
// http.ErrAbortHandler is a special exception which silently aborts the request, as per the standard library.
//
// Errors are handled as follows:
//
//   panic(Error) is sent as-is
//   panic(ErrorFunc(err)) uses err for SysMessage
//   panic(ErrorFunc) is equivalent to panic(ErrorFunc(nil))
//   panic(error) or anything else becomes ErrTypeInternal
//
// Internal errors dump the stack.
// You should only panic inside of handlers.
func PanicHandler(next http.Handler) http.Handler {
	const errMessageInternal = "サーバー内部でエラーが発生しました"
	fn := func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if ex := recover(); ex != nil && ex != http.ErrAbortHandler {
				var resp Error
				switch x := ex.(type) {
				case Error:
					resp = x
				case ErrorFunc:
					resp = x(nil)
				case func(error) Error:
					resp = x(nil)
				case error:
					resp = Error{
						Type:    ErrTypeInternal,
						Message: errMessageInternal,
						Cause:   x,
					}
				default:
					resp = Error{
						Type:    ErrTypeInternal,
						Message: errMessageInternal,
						Cause:   fmt.Errorf("panic: %v", x),
					}
				}
				status := resp.Type.StatusCode()
				middleware.PrintPrettyStack(ex)
				RenderJSON(w, status, resp)
			}
		}()
		next.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}
