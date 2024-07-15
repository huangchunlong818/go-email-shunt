package sender

type EmailSenderInterface interface {
	Send(*MailContent) (*EmailSendResult, error)
}
