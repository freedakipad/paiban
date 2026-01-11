// Package errors 提供统一的错误处理框架
package errors

import (
	"errors"
	"fmt"
	"net/http"
)

// Code 错误码
type Code string

const (
	// 通用错误码
	CodeUnknown       Code = "UNKNOWN"
	CodeInternal      Code = "INTERNAL_ERROR"
	CodeInvalidInput  Code = "INVALID_INPUT"
	CodeNotFound      Code = "NOT_FOUND"
	CodeAlreadyExists Code = "ALREADY_EXISTS"
	CodeUnauthorized  Code = "UNAUTHORIZED"
	CodeForbidden     Code = "FORBIDDEN"
	CodeTimeout       Code = "TIMEOUT"
	CodeRateLimited   Code = "RATE_LIMITED"

	// 排班引擎相关
	CodeConstraintViolation   Code = "CONSTRAINT_VIOLATION"
	CodeNoFeasibleSolution    Code = "NO_FEASIBLE_SOLUTION"
	CodeScheduleConflict      Code = "SCHEDULE_CONFLICT"
	CodeInsufficientResources Code = "INSUFFICIENT_RESOURCES"
	CodeInvalidTimeRange      Code = "INVALID_TIME_RANGE"

	// 派单相关
	CodeNoAvailableEmployee Code = "NO_AVAILABLE_EMPLOYEE"
	CodeDispatchFailed      Code = "DISPATCH_FAILED"
	CodeOrderNotAssignable  Code = "ORDER_NOT_ASSIGNABLE"
	CodeAreaNotCovered      Code = "AREA_NOT_COVERED"

	// 数据相关
	CodeDatabaseError  Code = "DATABASE_ERROR"
	CodeValidationFail Code = "VALIDATION_FAILED"
)

// AppError 应用错误
type AppError struct {
	Code       Code                   `json:"code"`
	Message    string                 `json:"message"`
	Details    string                 `json:"details,omitempty"`
	HTTPStatus int                    `json:"-"`
	Cause      error                  `json:"-"`
	Fields     map[string]interface{} `json:"fields,omitempty"`
}

// Error 实现 error 接口
func (e *AppError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// Unwrap 返回底层错误
func (e *AppError) Unwrap() error {
	return e.Cause
}

// WithDetails 添加详细信息
func (e *AppError) WithDetails(details string) *AppError {
	e.Details = details
	return e
}

// WithCause 添加原因
func (e *AppError) WithCause(cause error) *AppError {
	e.Cause = cause
	return e
}

// WithField 添加字段
func (e *AppError) WithField(key string, value interface{}) *AppError {
	if e.Fields == nil {
		e.Fields = make(map[string]interface{})
	}
	e.Fields[key] = value
	return e
}

// New 创建新错误
func New(code Code, message string) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		HTTPStatus: codeToHTTPStatus(code),
	}
}

// Wrap 包装错误
func Wrap(err error, code Code, message string) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		HTTPStatus: codeToHTTPStatus(code),
		Cause:      err,
	}
}

// codeToHTTPStatus 错误码转HTTP状态码
func codeToHTTPStatus(code Code) int {
	switch code {
	case CodeInvalidInput, CodeValidationFail, CodeInvalidTimeRange:
		return http.StatusBadRequest
	case CodeUnauthorized:
		return http.StatusUnauthorized
	case CodeForbidden:
		return http.StatusForbidden
	case CodeNotFound:
		return http.StatusNotFound
	case CodeAlreadyExists, CodeScheduleConflict:
		return http.StatusConflict
	case CodeRateLimited:
		return http.StatusTooManyRequests
	case CodeTimeout:
		return http.StatusGatewayTimeout
	case CodeNoFeasibleSolution, CodeNoAvailableEmployee:
		return http.StatusUnprocessableEntity
	default:
		return http.StatusInternalServerError
	}
}

// Is 检查错误是否为特定类型
func Is(err error, code Code) bool {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.Code == code
	}
	return false
}

// GetCode 获取错误码
func GetCode(err error) Code {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.Code
	}
	return CodeUnknown
}

// GetHTTPStatus 获取HTTP状态码
func GetHTTPStatus(err error) int {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.HTTPStatus
	}
	return http.StatusInternalServerError
}

// 预定义错误
var (
	ErrNotFound            = New(CodeNotFound, "资源不存在")
	ErrInvalidInput        = New(CodeInvalidInput, "输入参数无效")
	ErrUnauthorized        = New(CodeUnauthorized, "未授权访问")
	ErrForbidden           = New(CodeForbidden, "禁止访问")
	ErrInternal            = New(CodeInternal, "内部错误")
	ErrTimeout             = New(CodeTimeout, "操作超时")
	ErrNoFeasibleSolution  = New(CodeNoFeasibleSolution, "无可行解")
	ErrConstraintViolation = New(CodeConstraintViolation, "违反约束条件")
)

// InvalidInput 创建输入无效错误
func InvalidInput(field, reason string) *AppError {
	return New(CodeInvalidInput, fmt.Sprintf("字段 '%s' 无效: %s", field, reason))
}

// NotFound 创建资源不存在错误
func NotFound(resource, id string) *AppError {
	return New(CodeNotFound, fmt.Sprintf("%s '%s' 不存在", resource, id))
}

// ConstraintViolation 创建约束违反错误
func ConstraintViolation(constraint, details string) *AppError {
	return New(CodeConstraintViolation, fmt.Sprintf("违反约束 '%s': %s", constraint, details))
}

// NoFeasibleSolution 创建无可行解错误
func NoFeasibleSolution(reason string) *AppError {
	return New(CodeNoFeasibleSolution, reason)
}

// ScheduleConflict 创建排班冲突错误
func ScheduleConflict(empID, date, details string) *AppError {
	return New(CodeScheduleConflict, fmt.Sprintf("员工 %s 在 %s 存在排班冲突: %s", empID, date, details))
}

// NoAvailableEmployee 创建无可用员工错误
func NoAvailableEmployee(orderID, reason string) *AppError {
	return New(CodeNoAvailableEmployee, fmt.Sprintf("订单 %s 无可用员工: %s", orderID, reason))
}

// ValidationErrors 验证错误集合
type ValidationErrors struct {
	Errors []ValidationError `json:"errors"`
}

// ValidationError 单个验证错误
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// Error 实现 error 接口
func (ve *ValidationErrors) Error() string {
	if len(ve.Errors) == 0 {
		return "验证失败"
	}
	return fmt.Sprintf("验证失败: %s - %s", ve.Errors[0].Field, ve.Errors[0].Message)
}

// Add 添加验证错误
func (ve *ValidationErrors) Add(field, message string) {
	ve.Errors = append(ve.Errors, ValidationError{Field: field, Message: message})
}

// HasErrors 检查是否有错误
func (ve *ValidationErrors) HasErrors() bool {
	return len(ve.Errors) > 0
}

// ToAppError 转换为 AppError
func (ve *ValidationErrors) ToAppError() *AppError {
	err := New(CodeValidationFail, "验证失败")
	err.Fields = make(map[string]interface{})
	for _, e := range ve.Errors {
		err.Fields[e.Field] = e.Message
	}
	return err
}
