package sender

type MailContent struct {
	EmailType      int
	From           EmailAddress
	To             EmailAddress
	Subject        string
	Html           string
	UnsubscribeUrl string
	ReplyTo        EmailAddress
	Tags           map[string]string
}

type EmailAddress struct {
	Email string
	Name  string
}

type EmailSendResult struct {
	Success   bool   `json:"success"`
	Body      string `json:"body"`
	Status    int    `json:"status"`
	MessageId string `json:"message_id"`
}

type EmailSendInfo struct {
	EmailTypes []int     `json:"emailTypes"`
	DomainIds  []int     `json:"domainIds"`
	EmailInfo  EmailInfo `json:"emailInfo"`
}

type EmailInfo struct {
	FromEmail string `json:"fromEmail"`
}

// Sendgrid邮件发送
type Sendgrid struct {
	Personalizations []Personalizations `json:"personalizations"`
	From             From               `json:"from"`
	ReplyTo          ReplyTo            `json:"reply_to"`
	Subject          string             `json:"subject"`
	Content          []Content          `json:"content"`
	Headers          Headers            `json:"headers"`
	MailSettings     MailSettings       `json:"mail_settings"`
}

type ToEmail struct {
	Name  string `json:"name,omitempty"`
	Email string `json:"email"`
}

type Personalizations struct {
	To []ToEmail `json:"to"`
}

type From struct {
	Name  string `json:"name,omitempty"`
	Email string `json:"email,omitempty"`
}

type ReplyTo struct {
	Email string `json:"email,omitempty"`
}

type Content struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}

type Headers struct {
	ListUnsubscribe string `json:"List-Unsubscribe,omitempty"`
}

type SandboxMode struct {
	Enable bool `json:"enable"`
}

type MailSettings struct {
	SandboxMode SandboxMode `json:"sandbox_mode"`
}
