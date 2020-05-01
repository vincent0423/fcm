package fcm

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"time"
)

// 旧版的HTTP协议
// 需要注意的是: 应用服务器以 HTTP POST 请求的形式发送消息，并等待响应。
// 此机制是同步的，且发送者无法在收到响应前发送其他消息。

// fcm 服务常量
const (
	// 传递消息的http请求Url
	defaultEndpoint = "https://fcm.googleapis.com/fcm/send"
	// 最大的TTL，4周
	MAX_TTL = 2419200
	// 一次发送最大接收对象数量
	MAX_REGISTRATIONID_NUM = 1000
	// 通知优先级-高
	PRIORITY_HIGH = "high"
	// 通知优先级-普通
	PRIORITY_NORMAL = "normal"
	// 超时返回，要重试同一请求
	RETRY_AFTER_HEADER = "Retry-After"
	// 描述性错误键值
	ERROR_KEY = "error"
)

// 自定义错误
var (
	ErrInvalidAPIKey = errors.New("client API Key is invalid")
)

// http 方式客户端
type HttpClient struct {
	apiKey   string
	client   *http.Client
	endpoint string
	// v1
	projectId string
	v1        bool
}

// 新建一个Http方式的客户端
// 默认会初始化一个 http Client
func NewHttpClient(apiKey string, opts ...Option) (*HttpClient, error) {
	if apiKey == "" {
		return nil, ErrInvalidAPIKey
	}
	c := &HttpClient{
		apiKey:   apiKey,
		endpoint: defaultEndpoint,
		client:   &http.Client{},
	}
	for _, opt := range opts {
		if err := opt(c); err != nil {
			return nil, err
		}
	}
	return c, nil
}

// 可选配置项
type Option func(*HttpClient) error

// 使用自定义的httpclient
// 默认会有一个httpclinet
// 将client设置为nil，会在每次发送请求的时候，重新创建一个client
func WithClient(client *http.Client) Option {
	return func(c *HttpClient) error {
		c.client = client
		return nil
	}
}

// 发送消息
func (c *HttpClient) Send(msg *Message) (*Response, error) {
	// validate
	if err := msg.Validate(); err != nil {
		return nil, err
	}
	// marshal message
	data, err := json.Marshal(msg)
	if err != nil {
		return nil, err
	}
	return c.send(data)
}

// 发送消息，在遇到错误时候重试
func (c *HttpClient) SendWithRetry(msg *Message, retryAttempts int) (*Response, error) {
	// 验证
	if err := msg.Validate(); err != nil {
		return nil, err
	}
	data, err := json.Marshal(msg)
	if err != nil {
		return nil, err
	}

	resp := new(Response)
	err = retry(func() error {
		var err error
		resp, err = c.send(data)
		return err
	}, retryAttempts)
	return resp, err
}

// 发送请求
func (c *HttpClient) send(data []byte) (*Response, error) {
	// 创建请求
	req, err := http.NewRequest("POST", c.endpoint, bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}

	// 添加请求头
	req.Header.Add("Authorization", fmt.Sprintf("key=%s", c.apiKey))
	req.Header.Add("Content-Type", "application/json")

	httpClient := c.client
	if httpClient == nil {
		httpClient = &http.Client{}
	}
	// 执行嗯请求
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, connectionError(err.Error())
	}
	defer resp.Body.Close()

	// 构造返回
	response := new(Response)
	response.StatusCode = resp.StatusCode
	response.RetryAfter = resp.Header.Get(RETRY_AFTER_HEADER)
	// 检查返回状态
	if resp.StatusCode != http.StatusOK {
		return response, serverError(fmt.Sprintf("%d error: %s", resp.StatusCode, resp.Status))
	}
	if err := json.NewDecoder(resp.Body).Decode(response); err != nil {
		return nil, err
	}
	return response, nil
}

const (
	minBackoff = 100 * time.Millisecond
	maxBackoff = 1 * time.Minute
)

// 指数退避
func retry(fn func() error, attempts int) error {
	var attempt int
	for {
		err := fn()
		if err == nil {
			return nil
		}

		if tErr, ok := err.(net.Error); !ok || !tErr.Temporary() {
			return err
		}

		attempt++
		backoff := minBackoff * time.Duration(attempt*attempt)
		if attempt > attempts || backoff > maxBackoff {
			return err
		}
		time.Sleep(backoff)
	}
}
