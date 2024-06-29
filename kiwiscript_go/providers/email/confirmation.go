package email

import (
	"bytes"
	"html/template"
)

const confirmationPath = "auth/confirm"

const confirmationTemplate = `
<!DOCTYPE html>
<html lang="en">
<head>
	<meta charset="UTF-8">
	<meta http-equiv="X-UA-Compatible" content="IE=edge">
	<meta name="viewport" content="width=device-width, initial-scale=1.0">
	<title>Email Confirmation</title>
</head>
<body>
	<h1>Email Confirmation</h1>
	<br/>
	<p>Welcome {{.firstName}} {{.lastName}}</p>
	<br/>
	<p>Thank you for signing up to Kiwi Script. Please click the link below to confirm your email address.</p>
	<a href="{{.ConfirmationURL}}">Confirm Email</a>
	<p><small>Or copy this link: {{.confirmationURL}}</small></p>
	<br/>
	<p>Thank you,</p>
	<p>Kiwi Script Team</p>
</body>
`

type confirmationEmailData struct {
	FirstName       string
	LastName        string
	ConfirmationURL string
}

type ConfirmationEmailOptions struct {
	Email             string
	FirstName         string
	LastName          string
	ConfirmationToken string
}

func (m *Mail) SendConfirmationEmail(options ConfirmationEmailOptions) error {
	t, err := template.New("confirmation").Parse(confirmationTemplate)

	if err != nil {
		return err
	}

	var emailContent bytes.Buffer
	if err = t.Execute(&emailContent, confirmationEmailData{
		FirstName:       options.FirstName,
		LastName:        options.LastName,
		ConfirmationURL: m.buildUrl(confirmationPath, options.ConfirmationToken),
	}); err != nil {
		return err
	}

	return m.sendMail(options.Email, "Email Confirmation", emailContent.String())
}
