package msclient

import (
	"context"
	"fmt"
	"github.com/microsoftgraph/msgraph-sdk-go/models"
	graphusers "github.com/microsoftgraph/msgraph-sdk-go/users"
)

func (c *microsoftGraph) MyMailBox(token Token, mailProperties ...string) MailBox {
	return &myMailBox{token: token, mailProperties: mailProperties}
}

type MailBox interface {
	GetMails(ctx context.Context, size int32) (Mails, error)
}

type myMailBox struct {
	mailProperties []string
	token          Token
}

func (m *myMailBox) GetMails(ctx context.Context, size int32) (Mails, error) {

	requestParameters := &graphusers.ItemMessagesRequestBuilderGetQueryParameters{
		Top:    &size,
		Select: m.mailProperties,
	}
	configuration := &graphusers.ItemMessagesRequestBuilderGetRequestConfiguration{
		QueryParameters: requestParameters,
	}

	client, err := microsoftGraphClient(ctx, m.token)
	if err != nil {
		return nil, fmt.Errorf("token to ms graph client failed:%v", err)
	}
	mailBox := client.Me().Messages()
	messages, err := mailBox.Get(ctx, configuration)
	if err != nil {
		return nil, fmt.Errorf("connect outlook mailbox failed: %v", err)
	}

	var (
		mails = make(Mails, size)
		i     int
	)
	for _, message := range messages.GetValue() {
		if id := message.GetId(); id != nil {
			message, err = mailBox.ByMessageId(*id).Get(ctx, nil)
			if err != nil {
				return nil, fmt.Errorf("get mail detail failed:%v", err)
			}
			mails[i] = message
			i++
		}
	}
	return mails, nil
}

type Mails []models.Messageable
