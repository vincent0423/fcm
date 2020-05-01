package fcm

import (
	"encoding/json"
	"errors"
	"time"
)

var (
	// 目标token没有设置
	ErrMissingRegistration = errors.New("missing registration token")
	// 目标token错误
	ErrInvalidRegistration = errors.New("invalid registration token")
	// 目标token没有注册或者取消注册
	ErrNotRegistered = errors.New("unregistered device")
	// 包名错误
	ErrInvalidPackageName = errors.New("invalid package name")
	// 注册令牌和发送者id不匹配
	ErrMismatchSenderID = errors.New("mismatched sender id")
	// 消息太大，一般消息上限4096字节，主题消息2048字节
	ErrMessageTooBig = errors.New("message is too big")
	// 负载数据中含有非法key值
	ErrInvalidDataKey = errors.New("invalid data key")
	// time_to_alive 值错误
	ErrInvalidTTL = errors.New("invalid time to live")
	// 服务器无法处理请求，可以重试
	ErrUnavailable = connectionError("timeout")
	// 服务器在尝试处理请求时遇到错误,可以重试
	ErrInternalServerError = serverError("internal server error")
	// 消息投递率过高
	ErrDeviceMessageRateExceeded = errors.New("device message rate exceeded")
	// 主题消息投递率过高
	ErrTopicsMessageRateExceeded = errors.New("topics message rate exceeded")

	// xmpp
	// 有很多种可能
	ErrorInvalidJson = errors.New("invalid message")
	// 注册令牌错误
	ErrorBadRegistration = errors.New("registration is invalid")
	// 设备未注册
	ErrorDeviceUnRegistered = errors.New("device is unregistered")
	// senderID 不匹配
	ErrorSenderIdMismatch = errors.New("sender id mismatch")
	// 你给的ack消息不对
	ErrorBadAck = errors.New("bad ack")
	// 服务器无法及时处理请求，可重试
	ErrorServiceUnavaliable = errors.New("service unavaliable")
	// 服务器在尝试处理请求时遇到了错误，可重试
	ErrorInternalServerError = errors.New("internal server error")
	// 消息投递率过高
	ErrorDeviceMessageRateExceeded = errors.New("device message exceeded")
	// 消息投递率过高
	ErrorTopicMessageRateExceeded = errors.New("topic message exceeded")
	// 因为正在排空连接，所以无法处理消息。
	ErrorConnectionDraining = errors.New("connection drainning")
)

var (
	errMap = map[string]error{
		"MissingRegistration":          ErrMissingRegistration,
		"InvalidRegistration":          ErrInvalidRegistration,
		"NotRegistered":                ErrNotRegistered,
		"InvalidPackageName":           ErrInvalidPackageName,
		"MismatchSenderId":             ErrMismatchSenderID,
		"MessageTooBig":                ErrMessageTooBig,
		"InvalidDataKey":               ErrInvalidDataKey,
		"InvalidTtl":                   ErrInvalidTTL,
		"Unavailable":                  ErrUnavailable,
		"InternalServerError":          ErrInternalServerError,
		"DeviceMessageRateExceeded":    ErrDeviceMessageRateExceeded,
		"TopicsMessageRateExceeded":    ErrTopicsMessageRateExceeded, // following xmpp error
		"INVALID_JSON":                 ErrorInvalidJson,
		"BAD_REGISTRATION":             ErrorBadRegistration,
		"DEVICE_UNREGISTERED":          ErrorDeviceUnRegistered,
		"SENDER_ID_MISMATCH":           ErrorSenderIdMismatch,
		"BAD_ACK":                      ErrorBadAck,
		"SERVICE_UNAVAILABLE":          ErrorServiceUnavaliable,
		"INTERNAL_SERVER_ERROR":        ErrorInternalServerError,
		"DEVICE_MESSAGE_RATE_EXCEEDED": ErrorDeviceMessageRateExceeded,
		"TOPICS_MESSAGE_RATE_EXCEEDED": ErrorTopicMessageRateExceeded,
		"CONNECTION_DRAINING":          ErrorConnectionDraining, // 连接排空
	}
)

// 连接错误类型
type connectionError string

func (err connectionError) Error() string {
	return string(err)
}

func (err connectionError) Temporary() bool {
	return true
}

func (err connectionError) Timeout() bool {
	return true
}

// 服务错误
type serverError string

func (err serverError) Error() string {
	return string(err)
}

func (serverError) Temporary() bool {
	return true
}

func (serverError) Timeout() bool {
	return false
}

// fcm返回结构
type Response struct {
	MulticastID  int64    `json:"multicast_id"`
	Success      int      `json:"success"`
	Failure      int      `json:"failure"`
	CanonicalIDs int      `json:"canonical_ids"`
	StatusCode   int      `json:"error_code"`
	Results      []Result `json:"results"`
	RetryAfter   string
}

// 获取重试时间
func (r *Response) GetRetryAfterTime() (time.Duration, error) {
	return time.ParseDuration(r.RetryAfter)
}

// 消息传递的结果
type Result struct {
	MessageID      string `json:"message_id"`
	RegistrationID string `json:"registration_id"`
	Error          error  `json:"error"`
}

// 实现
func (r *Result) UnmarshalJSON(data []byte) error {
	var result struct {
		MessageID      string `json:"message_id"`
		RegistrationID string `json:"registration_id"`
		Error          string `json:"error"`
	}

	if err := json.Unmarshal(data, &result); err != nil {
		return err
	}

	r.MessageID = result.MessageID
	r.RegistrationID = result.RegistrationID
	if e, ok := errMap[result.Error]; ok {
		r.Error = e
	} else {
	}
	return nil
}

// 检查设备是否为注册
func (r Result) Unregistered() bool {
	switch r.Error {
	case ErrNotRegistered, ErrMismatchSenderID, ErrMissingRegistration, ErrInvalidRegistration,
		ErrorBadRegistration, ErrorDeviceUnRegistered, ErrorSenderIdMismatch:
		return true

	default:
		return false
	}
}
