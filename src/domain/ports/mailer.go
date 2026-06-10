package mailer

type EmailMessage struct {
    To      []string
    Subject string
    Body    string
    IsHTML  bool
}

type IMailer interface {
    Send(msg EmailMessage) error
}
