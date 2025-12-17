package errors

import (
	"errors"
	"net/http"
)

type ErrorType string
type ErrorMessage string

const (
	NotFound        ErrorType    = "NotFound"
	notFoundMessage ErrorMessage = "record not found"

	ValidationError        ErrorType    = "ValidationError"
	validationErrorMessage ErrorMessage = "validation error"

	ResourceAlreadyExists     ErrorType    = "ResourceAlreadyExists"
	alreadyExistsErrorMessage ErrorMessage = "resource already exists"

	RepositoryError        ErrorType    = "RepositoryError"
	repositoryErrorMessage ErrorMessage = "error in repository operation"

	NotAuthenticated             ErrorType    = "NotAuthenticated"
	notAuthenticatedErrorMessage ErrorMessage = "not Authenticated"

	TokenGeneratorError        ErrorType    = "TokenGeneratorError"
	tokenGeneratorErrorMessage ErrorMessage = "error in token generation"

	NotAuthorized             ErrorType    = "NotAuthorized"
	notAuthorizedErrorMessage ErrorMessage = "not authorized"

	UnknownError        ErrorType    = "UnknownError"
	unknownErrorMessage ErrorMessage = "something went wrong"
)

type AppError struct {
	Err  error
	Type ErrorType
	Message string
	Meta map[string]string
}
func NewResourceAlreadyExists(field string) *AppError {
	msg := field + " already exists"

	return &AppError{
		Err:     errors.New(msg),
		Type:    ResourceAlreadyExists,
		Message: msg,
		Meta: map[string]string{
			"field": field,
		},
	}
}


func NewAppError(err error, errType ErrorType) *AppError {
	return &AppError{
		Err:  err,
		Type: errType,
	}
}

func NewAppErrorWithType(errType ErrorType) *AppError {
	var err error
	var message string

	switch errType {
	case NotFound:
		err = errors.New(string(notFoundMessage))
		message = string(notFoundMessage)
	case ValidationError:
		err = errors.New(string(validationErrorMessage))
		message = string(validationErrorMessage)
	case ResourceAlreadyExists:
		err = errors.New(string(alreadyExistsErrorMessage))
		message = string(alreadyExistsErrorMessage)
	case RepositoryError:
		err = errors.New(string(repositoryErrorMessage))
		message = string(repositoryErrorMessage)
	case NotAuthenticated:
		err = errors.New(string(notAuthenticatedErrorMessage))
		message = string(notAuthenticatedErrorMessage)
	case NotAuthorized:
		err = errors.New(string(notAuthorizedErrorMessage))
		message = string(notAuthorizedErrorMessage)
	case TokenGeneratorError:
		err = errors.New(string(tokenGeneratorErrorMessage))
	default:
		err = errors.New(string(unknownErrorMessage))
		message = string(unknownErrorMessage)
	}

	return &AppError{
		Err:  err,
		Type: errType,
		Message: message,
	}
}

func (appErr *AppError) Error() string {
	return appErr.Err.Error()
}

// AppErrorToHTTP maps an AppError to an HTTP status code and message
func AppErrorToHTTP(appErr *AppError) (int, string) {
	switch appErr.Type {
	case NotFound:
		return http.StatusNotFound, appErr.Error()
	case ValidationError:
		return http.StatusBadRequest, appErr.Error()
	case RepositoryError:
		return http.StatusInternalServerError, appErr.Error()
	case ResourceAlreadyExists:
		return http.StatusConflict, appErr.Error()

	case NotAuthenticated:
		return http.StatusUnauthorized, appErr.Error()
	case NotAuthorized:
		return http.StatusForbidden, appErr.Error()
	default:
		return http.StatusInternalServerError, "Internal Server Error"
	}
}
