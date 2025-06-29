package schema

import (
	"net/http"
)

type BasicResponse struct {
	StatusCode int    `json:"statusCode"`
	Status     string `json:"status"`
	Message    string `json:"message,omitempty"`
}

type ErrorResponse struct {
	StatusCode int    `json:"statusCode"`
	Status     string `json:"status"`
	Message    string `json:"message,omitempty"`
}

type BasicResponseData[T any] struct {
	BasicResponse
	Data T `json:"data"`
}

var BadParamsErrorResponse = ErrorResponse{
	StatusCode: http.StatusBadRequest,
	Status:     "error",
	Message:    "bad parameters",
}

var InternalServerErrorResponse = ErrorResponse{
	StatusCode: http.StatusInternalServerError,
	Status:     "error",
	Message:    "internal server error",
}

var OkBasicResponse = BasicResponse{
	StatusCode: http.StatusOK,
	Status:     "success",
}
