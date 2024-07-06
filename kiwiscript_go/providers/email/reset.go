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
	<p>Hello {{.FirstName}} {{.LastName}}</p>
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
