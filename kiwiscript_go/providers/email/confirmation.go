// Copyright (C) 2024 Afonso Barracha
//
// This file is part of KiwiScript.
//
// KiwiScript is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// KiwiScript is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with KiwiScript.  If not, see <https://www.gnu.org/licenses/>.

package email

import (
	"bytes"
	"context"
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
	<p>Welcome {{.FirstName}} {{.LastName}}</p>
	<br/>
	<p>Thank you for signing up to Kiwi Script. Please click the link below to confirm your email address.</p>
	<a href="{{.ConfirmationURL}}">Confirm Email</a>
	<p><small>Or copy this link: {{.ConfirmationURL}}</small></p>
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
	RequestID         string
	Email             string
	FirstName         string
	LastName          string
	ConfirmationToken string
}

func (m *Mail) SendConfirmationEmail(ctx context.Context, opts ConfirmationEmailOptions) error {
	log := m.buildLogger(opts.RequestID, "SendConfirmationEmail").With(
		"firstName", opts.FirstName,
		"lastName", opts.LastName,
	)
	log.DebugContext(ctx, "Sending confirmation email...")

	t, err := template.New("confirmation").Parse(confirmationTemplate)
	if err != nil {
		log.ErrorContext(ctx, "Failed to parse email template", "error", err)
		return err
	}

	data := confirmationEmailData{
		FirstName:       opts.FirstName,
		LastName:        opts.LastName,
		ConfirmationURL: m.buildUrl(confirmationPath, opts.ConfirmationToken),
	}
	var emailContent bytes.Buffer
	if err := t.Execute(&emailContent, data); err != nil {
		log.ErrorContext(ctx, "Failed to execute email template", "error", err)
		return err
	}

	return m.sendMail(opts.Email, "Email Confirmation", emailContent.String())
}
