package email

import (
	"bytes"
	"html/template"
)

const resetPath = "auth/password-reset"

const resetTemplate = `
<!DOCTYPE html>
<html lang="en">
<head>
	<meta charset="UTF-8">
	<meta http-equiv="X-UA-Compatible" content="IE=edge">
	<meta name="viewport" content="width=device-width, initial-scale=1.0">
	<title>Password Reset</title>
</head>
<body>
	<h1>Password Reset</h1>
	<br/>
	<p>Hello {{.firstName}} {{.lastName}}</p>
	<br/>
	<p>We received a request to reset your password. Please click the link below to reset your password.</p>
	<a href="{{.ResetURL}}">Reset Password</a>
	<p><small>Or copy this link: {{.ResetURL}}</small></p>
	<br/>
	<p>If you did not request a password reset, please ignore this email.</p>
	<br/>
	<p>Thank you,</p>
	<p>Kiwi Script Team</p>
</body>
`

type resetEmailData struct {
	FirstName string
	LastName  string
	ResetURL  string
}

type ResetEmailOptions struct {
	Email      string
	FirstName  string
	LastName   string
	ResetToken string
}

func (m *Mail) SendResetEmail(options ResetEmailOptions) error {
	t, err := template.New("reset").Parse(resetTemplate)

	if err != nil {
		return err
	}

	var emailContent bytes.Buffer
	if err = t.Execute(&emailContent, resetEmailData{
		FirstName: options.FirstName,
		LastName:  options.LastName,
		ResetURL:  m.buildUrl(resetPath, options.ResetToken),
	}); err != nil {
		return err
	}

	return m.sendMail(options.Email, "Password Reset", emailContent.String())
}
