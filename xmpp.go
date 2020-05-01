package fcm

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	xmpp "github.com/mattn/go-xmpp"
)

const (
	productionEndpoint = "fcm-xmpp.googleapis.com:5235"
	testingEndpoint    = "fcm-xmpp.googleapis.com:5236"
	cmdTypeSend        = 0
	cmdTypeRenew       = 1
	cmdTypeClose       = 2
	defaultTimeout     = time.Second * 15
)

// xmpp 客户端
type XmppClient struct {
	user, passwd string
	debug        bool
	client       *xmpp.Client
	cmds         chan (*cmd)
	retInfo      map[string]*sendInfo
}

type cmd struct {
	t int
	d interface{}
}

type sendInfo struct {
	ret chan *Response
	msg *Message
}

// 新建一个连接
func (c *XmppClient) renewConn() error {
	host := productionEndpoint
	if c.debug {
		host = testingEndpoint
	}
	client, err := xmpp.NewClient(host, c.user, c.passwd, c.debug)
	if err != nil {
		log.Println(err)
		return err
	}
	c.client = client
	return nil
}

// fcm必须要用tls
// user是<sendId>@gcm.googleapis.com
// passwd 就是 appKey
func NewXmppClient(user, passwd string, debug bool, cmdsize int) (*XmppClient, error) {
	c := &XmppClient{
		user:    user,
		passwd:  passwd,
		debug:   debug,
		cmds:    make(chan *cmd, cmdsize),
		retInfo: make(map[string]*sendInfo),
	}
	err := c.renewConn()
	if err != nil {
		return nil, err
	}
	c.run()
	return c, nil
}

// 发送消息
func (c *XmppClient) Send(msg *Message, timeout ...time.Duration) (*Response, error) {
	sendInfo := &sendInfo{
		msg: msg,
		ret: make(chan *Response),
	}
	si := &cmd{
		t: cmdTypeSend,
		d: sendInfo,
	}
	c.cmds <- si

	to := defaultTimeout
	if len(timeout) > 0 {
		to = timeout[0]
	}
	resp := new(Response)
	select {
	case <-time.After(to):
		resp = &Response{
			StatusCode: 503,
			Results: []Result{
				{
					MessageID:      msg.MessageId,
					RegistrationID: msg.To,
					Error:          ErrorServiceUnavaliable,
				},
			},
		}
	case resp = <-sendInfo.ret:
	}
	delete(c.retInfo, msg.MessageId)
	return resp, nil
}

// 发送不直接返回结果
func (c *XmppClient) AsyncSend(msg *Message) error {
	sendInfo := &sendInfo{
		msg: msg,
		ret: make(chan *Response),
	}
	si := &cmd{
		t: cmdTypeSend,
		d: sendInfo,
	}
	c.cmds <- si
	return nil
}

func (c *XmppClient) send(si *sendInfo) error {
	data, _ := json.Marshal(si.msg)
	xmppDataStr := fmt.Sprintf(
		`<message id="">
			<gcm xmlns="google:mobile:data">
				%s
			</gcm>
		</message>`, string(data))
	_, err := c.client.SendOrg(xmppDataStr)
	if err != nil {
		c.retInfo[si.msg.MessageId] = si
	}
	return err
}

// 关闭
func (c *XmppClient) Close() error {
	c.cmds <- &cmd{t: cmdTypeClose}
	if c.client != nil {
		return c.client.Close()
	}
	return nil
}

func (c *XmppClient) run() {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Println(r)
			}
		}()
		for {
			select {
			case cmd := <-c.cmds:
				switch cmd.t {
				case cmdTypeSend:
					sendInfo := cmd.d.(*sendInfo)
					c.send(sendInfo)
				case cmdTypeRenew:
					c.renewConn()
				case cmdTypeClose:
					return
				}
			default:
				chat, err := c.client.Recv()
				if err != nil {
					log.Println(err)
				}
				response := new(Response)
				switch v := chat.(type) {
				case xmpp.Chat:
					if len(v.Other) > 0 {
						xmppResp := new(xmppResponse)
						json.Unmarshal([]byte(v.Other[0]), xmppResp)
						if strings.EqualFold(xmppResp.MessageType, "ack") {
							response.StatusCode = 200
						} else {
							log.Printf("%+v", v)
						}
						response.Results = append(response.Results, Result{
							MessageID:      xmppResp.MessageId,
							RegistrationID: xmppResp.Form,
							Error:          errMap[xmppResp.Error],
						})
						if si, ok := c.retInfo[xmppResp.MessageId]; ok {
							si.ret <- response
						}
						if errMap[xmppResp.Error] == ErrorConnectionDraining {
							go func() {
								c.cmds <- &cmd{t: cmdTypeRenew}
							}()
						}
					}
				case xmpp.Presence:
					log.Printf(" presence: %+v", v)
				case xmpp.IQ:
					log.Printf("iq: %+v", v)
				}
			}
		}
	}()
}

type xmppResponse struct {
	MessageType      string `json:"message_type"`
	MessageId        string `json:"message_id"`
	Form             string `json:"from"`
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
}
