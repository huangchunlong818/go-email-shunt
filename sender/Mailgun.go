package sender

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"go.uber.org/zap"
	"io"
	"mime/multipart"
	"net/http"
)

type Client struct{}

type EmailDomain struct {
	Domain string
	APIKey string
}

type MailgunSender struct {
	domain *EmailDomain
}

func NewMailgunSender(domain *EmailDomain) *MailgunSender {
	return &MailgunSender{domain: domain}
}

// mailgun发送邮件
func (m *MailgunSender) Send(content *MailContent) (*EmailSendResult, error) {
	from := fmt.Sprintf("%s <%s>", content.From.Name, content.From.Email)
	to := content.To.Email
	subject := content.Subject
	html := content.Html
	unsubscribeUrl := content.UnsubscribeUrl
	replyTo := content.ReplyTo.Email
	tags := content.Tags

	multipartBody := &bytes.Buffer{}
	writer := multipart.NewWriter(multipartBody)

	err := m.addFormField(writer, "from", from)
	if err != nil {
		zap.S().Errorf("Error adding form field:%#v", err)
	}
	err = m.addFormField(writer, "to", to)
	if err != nil {
		return nil, err
	}
	err = m.addFormField(writer, "subject", subject)
	if err != nil {
		return nil, err
	}
	err = m.addFormField(writer, "html", html)
	if err != nil {
		return nil, err
	}

	if unsubscribeUrl != "" {
		err = m.addFormField(writer, "h:List-Unsubscribe", fmt.Sprintf("<%s>", unsubscribeUrl))
		if err != nil {
			return nil, err
		}
	}

	if replyTo != "" {
		err = m.addFormField(writer, "h:Reply-To", replyTo)
		if err != nil {
			return nil, err
		}
	}

	for _, tag := range tags {
		err = m.addFormField(writer, "o:tag", tag)
		if err != nil {
			return nil, err
		}
	}

	_ = writer.Close()

	mailgunApiUrl := fmt.Sprintf("https://api.mailgun.net/v3/%s/messages", m.domain.Domain)
	req, err := http.NewRequestWithContext(context.Background(), "POST", mailgunApiUrl, multipartBody)
	if err != nil {
		zap.S().Errorf("Error creating request: %#v", err)
	}

	req.SetBasicAuth("api", m.domain.APIKey)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		zap.S().Errorf("Error sending request:%#v", err)
	}

	body, _ := io.ReadAll(resp.Body)
	var bodyData map[string]interface{}
	err = json.Unmarshal(body, &bodyData)
	if err != nil {
		return nil, err
	}
	messageId := bodyData["id"].(string)

	result := &EmailSendResult{
		Success:   resp.StatusCode >= 200 && resp.StatusCode < 300,
		Body:      string(body),
		Status:    resp.StatusCode,
		MessageId: messageId,
	}

	return result, nil
}

// 添加form-data字段
func (m *MailgunSender) addFormField(writer *multipart.Writer, name, value string) error {
	part, err := writer.CreateFormFile(name, "")
	if err != nil {
		return err
	}
	_, err = part.Write([]byte(value))
	return err
}
