package fcm

import (
	"context"

	firebase "firebase.google.com/go"
	"firebase.google.com/go/messaging"
	"google.golang.org/api/option"
)

const (
	// v1版本接口
	httpV1Endpoint = "https://fcm.googleapis.com/v1/projects/%s/messages:send"
	// v1接口权限范围
	fcmScope = "https://www.googleapis.com/auth/firebase.messaging"
)

type HttpV1Client struct {
	*messaging.Client
}

func NewHttpV1Client(credentialFile string) (*HttpV1Client, error) {
	app, err := firebase.NewApp(context.Background(), &firebase.Config{},
		option.WithServiceAccountFile(credentialFile),
		option.WithScopes(fcmScope))
	if err != nil {
		return nil, err
	}
	client, err := app.Messaging(context.Background())
	if err != nil {
		return nil, err
	}
	return &HttpV1Client{client}, nil
}
