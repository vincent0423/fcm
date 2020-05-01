package fcm

// http和xmpp请求的消息体结构

import (
	"errors"
	"strings"
)

var (
	// 错误的消息结构
	ErrInvalidMessage = errors.New("message is invalid")
	// 错误的目标结构
	ErrInvalidTarget = errors.New("topic is invalid or registration ids are not set")
	// 消息发送对象超过上限 1000
	ErrToManyRegIDs = errors.New("too many registrations ids")
	// 消息维持时间太长
	ErrInvalidTimeToLive = errors.New("messages time-to-live is invalid")
)

// 所有平台的通知合集
type Notification struct {
	Title            string `json:"title,omitempty"`
	Body             string `json:"body,omitempty"`
	Icon             string `json:"icon,omitempty"`
	ClickAction      string `json:"click_action,omitempty"`
	Color            string `json:"color,omitempty"`
	Sound            string `json:"sound,omitempty"`
	Tag              string `json:"tag,omitempty"`
	BodyLocKey       string `json:"body_loc_key,omitempty"`
	BodyLocArgs      string `json:"body_loc_args,omitempty"`
	TitleLocKey      string `json:"title_loc_key,omitempty"`
	TitleLocArgs     string `json:"title_loc_args,omitempty"`
	AndroidChannelID string `json:"android_channel_id,omitempty"`
	Badge            string `json:"badge,omitempty"`
	Subtitle         string `json:"subtitle,omitempty"`
}

// 包含了推送对象，消息内容等
type Message struct {
	// 可选,可以是单个设备注册令牌，也可以是单个主题
	To string `json:"to,omitempty"`
	// 可选,此参数用于指定多播消息（发送到多个注册令牌的消息）的接收者。
	RegistrationIDs []string `json:"registration_ids,omitempty"`
	// 可选,此参数指定用于确定消息目标的逻辑条件表达式。
	Condition string `json:"condition,omitempty"`
	// 可选,此参数用于指定一组可折叠的消息
	CollapseKey string `json:"collapse_key,omitempty"`
	// 可选,设置消息的优先级
	Priority string `json:"priority,omitempty"`
	// 可选,在 iOS 中，使用此字段表示 APNs 有效负载中的 content-available。
	// 如果发送通知或消息时此字段设为 true，将会唤醒处于非活动状态的客户端应用，
	// 且消息将作为静默通知通过 APNs 发送，而不是通过 FCM 连接服务器发送。
	// 请注意，APNs 中的静默通知不保证能传送，并且可能受多种因素影响（比如用户开启低电量模式、强制退出应用等）。
	// 在 Android 系统上，数据消息默认唤醒应用。Chrome 中目前不支持此功能。
	ContentAvailable bool `json:"content_available,omitempty"`
	// 可选,目前仅适用于运行 iOS 10 及更高版本的设备
	MutableContent bool `json:"mutable_content"`
	// 可选,此参数指定当设备离线时消息在 FCM 存储中保留的时长（以秒为单位）。受支持的最长生存时间为 4 周，默认值为 4 周。
	TimeToLive int `json:"time_to_live,omitempty"`
	// 可选,此参数让应用服务器能够请求消息传递的确认。
	DeliveryReceiptRequested bool `json:"delivery_receipt_requested,omitempty"`
	// 可选,此值为true时候,消息会在app激活才发送
	DelayWhileIdle bool `json:"delay_while_idle,omitempty"`
	// 可选,仅适用于 Android 应用,此参数用于指定应用的软件包名称，其注册令牌必须匹配才能接收消息。
	RestrictedPackageName string `json:"restricted_package_name,omitempty"`
	// 可选,此参数设置为 true 时，开发者可在不实际发送消息的情况下对请求进行测试。
	DryRun bool `json:"dry_run,omitempty"`
	// 可选,此参数用于指定消息有效负载的自定义键值对。
	Data map[string]interface{} `json:"data,omitempty"`
	// 可选,此参数用于指定通知有效负载的用户可见的预定义键值对。
	// 如果为发送到 iOS 设备的消息提供了通知有效负载，或者将 content_available 选项设为 true，
	// 消息将通过 APNs 发送，否则会通过 FCM 连接服务器发送。
	Notification *Notification `json:"notification,omitempty"`
	// xmpp
	MessageId string `json:"message_id,omitempty"`
}

// 验证消息的有效性
func (msg *Message) Validate() error {
	if msg == nil {
		return ErrInvalidMessage
	}

	opCnt := strings.Count(msg.Condition, "&&") + strings.Count(msg.Condition, "||")
	if msg.To == "" && (msg.Condition == "" || opCnt > 2) && len(msg.RegistrationIDs) == 0 {
		return ErrInvalidTarget
	}
	if len(msg.RegistrationIDs) > MAX_REGISTRATIONID_NUM {
		return ErrToManyRegIDs
	}
	if msg.TimeToLive > MAX_TTL {
		return ErrInvalidTimeToLive
	}
	return nil
}
