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
	<p>Hello {{.FirstName}} {{.LastName}}</p>
	<br/>
	<p>Your access code is: <strong>{{.Code}}</strong></p>
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
	RequestID string
	Email     string
	FirstName string
	LastName  string
	Code      string
}

func (m *Mail) SendCodeEmail(ctx context.Context, opts CodeEmailOptions) error {
	log := m.buildLogger(opts.RequestID, "SendCodeEmail").With(
		"firstName", opts.FirstName,
		"lastName", opts.LastName,
	)
	log.DebugContext(ctx, "Sending code email...")
	t, err := template.New("code").Parse(codeTemplate)

	if err != nil {
		log.ErrorContext(ctx, "Failed to parse email template", "error", err)
		return err
	}

	data := codeEmailData{
		FirstName: opts.FirstName,
		LastName:  opts.LastName,
		Code:      opts.Code,
	}
	var emailContent bytes.Buffer
	if err := t.Execute(&emailContent, data); err != nil {
		log.ErrorContext(ctx, "Failed to execute email template", "error", err)
		return err
	}

	return m.sendMail(opts.Email, "Access Code", emailContent.String())
}
