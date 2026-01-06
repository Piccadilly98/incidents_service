package error_worker

import (
	"database/sql"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
)

type userErrorInfo struct {
	text string
	code int
}

type dbErrorinfo struct {
	isUser        bool
	pattern       string
	code          int
	messageResult string
}

type ErrorWorker struct {
	userPatterns       []userErrorInfo
	dbPatterns         []dbErrorinfo
	isLoggingUserError bool
	userErrorLogger    *log.Logger
	dbErrorLogger      *log.Logger
}

func NewErrorWorker(isLoggingUserError bool) *ErrorWorker {
	ew := &ErrorWorker{
		userErrorLogger:    log.New(os.Stdout, "[USER ERROR]  ", log.Ldate|log.Ltime),
		dbErrorLogger:      log.New(os.Stderr, "[SERVER DB ERROR]  ", log.Ldate|log.Ltime),
		isLoggingUserError: isLoggingUserError,
	}
	ew.initErrors()
	return ew
}

func (ew *ErrorWorker) initErrors() {
	//DTO VALIDATION
	ew.AddNewUserError("cannot be", http.StatusBadRequest)
	ew.AddNewUserError("latitude incorrect", http.StatusBadRequest)
	ew.AddNewUserError("longitude incorrect", http.StatusBadRequest)
	ew.AddNewUserError("latitide: incorrect format", http.StatusBadRequest)
	ew.AddNewUserError("longitude: incorrect format", http.StatusBadRequest)
	ew.AddNewUserError("invalid type incident_id", http.StatusBadRequest)
	ew.AddNewUserError("no data for update", http.StatusBadRequest)
	ew.AddNewUserError("is not uuid", http.StatusBadRequest)
	ew.AddNewUserError("is not integer", http.StatusBadRequest)

	//service
	ew.AddNewUserError("very long", http.StatusBadRequest)
	ew.AddNewUserError("invalid status", http.StatusBadRequest)
	ew.AddNewUserError("unexpected status", http.StatusBadRequest)
	ew.AddNewUserError("invalid incident_id", http.StatusNotFound)
	ew.AddNewUserError("unable to update archived incident", http.StatusConflict)
	ew.AddNewUserError("incident already archived", http.StatusConflict)
	ew.AddNewUserError("invalid page_num", http.StatusBadRequest)
	ew.AddNewUserError("EOF", http.StatusBadRequest)
	ew.AddNewUserError("must be", http.StatusBadRequest)
	ew.AddNewUserError("invalid page", http.StatusBadRequest)

	//db - user error
	ew.AddNewDbError(true, "violates foreign key", "invalid request", http.StatusBadRequest)
	ew.AddNewDbError(true, "invalid input", "invalid request", http.StatusBadRequest)
	ew.AddNewDbError(true, "invalid format", "invalid request", http.StatusBadRequest)
	ew.AddNewDbError(true, "duplicate key value", "id is not unique", http.StatusConflict)
	ew.AddNewDbError(true, "value too long", "value too long", http.StatusBadRequest)
	ew.AddNewDbError(true, "duplicate key value", "already exists", http.StatusConflict)
	ew.AddNewDbError(true, "EOF", "body empty", http.StatusBadRequest)

	//db - server err
	ew.AddNewDbError(false, "connection refused", "service unavailable", http.StatusServiceUnavailable)
	ew.AddNewDbError(false, "no such host", "service unavailable", http.StatusServiceUnavailable)
	ew.AddNewDbError(false, "host", "service unavailable", http.StatusServiceUnavailable)
	ew.AddNewDbError(false, "does not exist", "service unavailable", http.StatusServiceUnavailable)
	ew.AddNewDbError(false, "connections", "service unavailable", http.StatusServiceUnavailable)
	ew.AddNewDbError(false, " many clients", "service unavailable", http.StatusServiceUnavailable)
	ew.AddNewDbError(false, "shutting down", "service unavailable", http.StatusServiceUnavailable)
	ew.AddNewDbError(false, "network is unreachable", "service unavailable", http.StatusServiceUnavailable)
	ew.AddNewDbError(false, "network", "service unavailable", http.StatusServiceUnavailable)
	ew.AddNewDbError(false, "syntax", "service unavailable", http.StatusServiceUnavailable)
}

func (ew *ErrorWorker) AddNewUserError(pattern string, statusCode int) {
	ew.userPatterns = append(ew.userPatterns, userErrorInfo{
		text: pattern,
		code: statusCode,
	})
}

func (ew *ErrorWorker) AddNewDbError(isUser bool, pattern, resultMsg string, code int) {
	ew.dbPatterns = append(ew.dbPatterns, dbErrorinfo{
		isUser:        isUser,
		pattern:       pattern,
		messageResult: resultMsg,
		code:          code,
	})
}

func (ew *ErrorWorker) ProcessError(err error) (int, error) {
	if err == nil {
		return http.StatusOK, nil
	}

	errStr := strings.ToLower(err.Error())

	if errors.Is(err, io.EOF) {
		if ew.isLoggingUserError {
			ew.dbErrorLogger.Printf("DB/User error [code %d]: %v", http.StatusBadRequest, err)
		}
		return http.StatusBadRequest, fmt.Errorf("body empty")
	}

	switch {
	case err == sql.ErrNoRows:
		ew.userErrorLogger.Printf("Not found id: %s\n", errStr)
		return http.StatusNotFound, fmt.Errorf("not found id")
	case strings.Contains(errStr, "context canceled"):
		return -1, nil
	case strings.Contains(errStr, "deadline exceeded"):
		ew.dbErrorLogger.Printf("Deadline exceeded: %v", err)
		return http.StatusGatewayTimeout, fmt.Errorf("request timeout")
	}

	for _, pattern := range ew.dbPatterns {
		if strings.Contains(errStr, pattern.pattern) {
			if pattern.isUser {
				if ew.isLoggingUserError {
					ew.userErrorLogger.Printf("User error [response code %d]: %v", pattern.code, err)
				}
			} else {
				ew.dbErrorLogger.Printf("DB error [code %d]: %v", pattern.code, err)
			}
			return pattern.code, errors.New(pattern.messageResult)
		}
	}

	for _, pattern := range ew.userPatterns {
		if strings.Contains(errStr, pattern.text) {
			if ew.isLoggingUserError {
				ew.userErrorLogger.Printf("User error [response code %d]: %v", pattern.code, err)
			}
			return pattern.code, err
		}
	}

	ew.dbErrorLogger.Printf("CRITICAL - Unknown error: %v", err)
	return http.StatusInternalServerError, fmt.Errorf("internal server error")
}
