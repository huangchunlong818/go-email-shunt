package sender

import (
	"bytes"
	"emailWarming/utils"
	"encoding/json"
	"errors"
	"time"
)

type SendgridSender struct {
	domain *EmailDomain
}

func NewSendgridSender(domain *EmailDomain) *SendgridSender {
	return &SendgridSender{domain: domain}
}

// sendgrid发送邮件
func (s *SendgridSender) Send(content *MailContent) (*EmailSendResult, error) {
	var data = Sendgrid{
		Personalizations: []Personalizations{
			{
				To: []ToEmail{
					{
						Name:  content.To.Name,
						Email: content.To.Email,
					},
				},
			},
		},
		From: From{
			Name:  content.From.Name,
			Email: content.From.Email,
		},
		ReplyTo: ReplyTo{
			Email: content.ReplyTo.Email,
		},
		Subject: content.Subject,
		Content: []Content{
			{
				Type:  "text/html",
				Value: content.Html,
			},
		},
		Headers: Headers{
			ListUnsubscribe: "",
		},
		MailSettings: MailSettings{
			SandboxMode{
				Enable: false,
			},
		},
	}
	if content.UnsubscribeUrl != "" {
		data.Headers.ListUnsubscribe = "<" + content.UnsubscribeUrl + ">"
	}

	mailgunApiUrl := "https://api.sendgrid.com/v3/mail/send"
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	reader := bytes.NewReader(jsonData)
	headerMap := map[string]string{
		"Authorization": "Bearer " + s.domain.APIKey,
		"Content-Type":  "application/json",
	}
	timeout := 60 * time.Second
	resp, err := utils.ApiCall("POST", mailgunApiUrl, reader, headerMap, timeout)
	if err != nil {
		return nil, err
	}
	if resp == nil {
		return nil, errors.New("resp is nil")
	}

	messageId := resp.RspHeader.Get("X-Message-Id")
	status := resp.HttpCode
	isSendingSuccess := status >= 200 && status < 300
	body := resp.Data

	return &EmailSendResult{
		Success:   isSendingSuccess,
		Body:      string(body),
		Status:    status,
		MessageId: messageId,
	}, nil
}
