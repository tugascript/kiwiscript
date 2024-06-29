package email

import (
	"bytes"
	"html/template"
)

const codeTemplate = `
<!DOCTYPE html>
<html lang="en">
<head>
	<meta charset="UTF-8">
	<meta http-equiv="X-UA-Compatible" content="IE=edge">
	<meta name="viewport" content="width=device-width, initial-scale=1.0">
	<title>Access Code</title>
</head>
<body>
	<h1>Access Code</h1>
	<br/>
	<p>Hello {{.firstName}} {{.lastName}}</p>
	<br/>
	<p>Your access code is: <strong>{{.code}}</strong></p>
	<br/>
	<p>Happy coding,</p>
	<p>Kiwi Script Team</p>
</body>
`

type codeEmailData struct {
	FirstName string
	LastName  string
	Code      string
}

type CodeEmailOptions struct {
	Email     string
	FirstName string
	LastName  string
	Code      string
}

func (m *Mail) SendCodeEmail(options CodeEmailOptions) error {
	t, err := template.New("code").Parse(codeTemplate)

	if err != nil {
		return err
	}

	var emailContent bytes.Buffer
	if err = t.Execute(&emailContent, codeEmailData{
		FirstName: options.FirstName,
		LastName:  options.LastName,
		Code:      options.Code,
	}); err != nil {
		return err
	}

	return m.sendMail(options.Email, "Access Code", emailContent.String())
}
