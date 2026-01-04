package error_worker

import (
	"database/sql"
	"errors"
	"fmt"
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

	//service
	ew.AddNewUserError("very long", http.StatusBadRequest)
	ew.AddNewUserError("invalid status", http.StatusBadRequest)
	ew.AddNewUserError("unexpected status", http.StatusBadRequest)
	ew.AddNewUserError("invalid incident_id", http.StatusNotFound)

	//db - user error
	ew.AddNewDbError("violates foreign key", "invalid request", http.StatusBadRequest)
	ew.AddNewDbError("invalid input", "invalid request", http.StatusBadRequest)
	ew.AddNewDbError("invalid format", "invalid request", http.StatusBadRequest)
	ew.AddNewDbError("duplicate key value", "already exists", http.StatusConflict)
	ew.AddNewDbError("value too long", "value too long", http.StatusBadRequest)
	ew.AddNewDbError("duplicate key value", "already exists", http.StatusConflict)

	//db - server err
	ew.AddNewDbError("connection refused", "service unavailable", http.StatusServiceUnavailable)
	ew.AddNewDbError("no such host", "service unavailable", http.StatusServiceUnavailable)
	ew.AddNewDbError("host", "service unavailable", http.StatusServiceUnavailable)
	ew.AddNewDbError("does not exist", "service unavailable", http.StatusServiceUnavailable)
	ew.AddNewDbError("connections", "service unavailable", http.StatusServiceUnavailable)
	ew.AddNewDbError(" many clients", "service unavailable", http.StatusServiceUnavailable)
	ew.AddNewDbError("shutting down", "service unavailable", http.StatusServiceUnavailable)
	ew.AddNewDbError("network is unreachable", "service unavailable", http.StatusServiceUnavailable)
	ew.AddNewDbError("network", "service unavailable", http.StatusServiceUnavailable)
	ew.AddNewDbError("syntax", "service unavailable", http.StatusServiceUnavailable)
}

func (ew *ErrorWorker) AddNewUserError(pattern string, statusCode int) {
	ew.userPatterns = append(ew.userPatterns, userErrorInfo{
		text: pattern,
		code: statusCode,
	})
}

func (ew *ErrorWorker) AddNewDbError(pattern, resultMsg string, code int) {
	ew.dbPatterns = append(ew.dbPatterns, dbErrorinfo{
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

	switch {
	case err == sql.ErrNoRows:
		ew.userErrorLogger.Printf("Not found: %s\n", errStr)
		return http.StatusNotFound, fmt.Errorf("not found")
	case strings.Contains(errStr, "context canceled"):
		return -1, nil
	case strings.Contains(errStr, "deadline exceeded"):
		ew.dbErrorLogger.Printf("Deadline exceeded: %v", err)
		return http.StatusGatewayTimeout, fmt.Errorf("request timeout")
	}

	for _, pattern := range ew.dbPatterns {
		if strings.Contains(errStr, pattern.pattern) {
			ew.dbErrorLogger.Printf("DB error [code %d]: %v", pattern.code, err)
			return pattern.code, errors.New(pattern.messageResult)
		}
	}

	for _, pattern := range ew.userPatterns {
		if strings.Contains(errStr, pattern.text) {
			if ew.isLoggingUserError {
				ew.userErrorLogger.Printf("User error [code %d]: %v", pattern.code, err)
			}
			return pattern.code, err
		}
	}

	ew.dbErrorLogger.Printf("CRITICAL - Unknown error: %v", err)
	return http.StatusInternalServerError, fmt.Errorf("internal server error")
}
