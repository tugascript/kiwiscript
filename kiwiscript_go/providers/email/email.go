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
	"net/smtp"
)

type Mail struct {
	auth           smtp.Auth
	port           string
	host           string
	address        string
	frontendDomain string
}

func NewMail(username, password, port, host, name, frontendDomain string) *Mail {
	return &Mail{
		auth:           smtp.PlainAuth("", username, password, host),
		port:           port,
		host:           host,
		address:        name + " <" + username + ">",
		frontendDomain: frontendDomain,
	}
}

func (m *Mail) sendMail(to string, subject, body string) error {
	addr := m.host + ":" + m.port
	msg := []byte("To: " + to + "\r\n" + "Subject: " + subject + "\r\n" + body)
	return smtp.SendMail(addr, m.auth, m.address, []string{to}, msg)
}

func (m *Mail) buildUrl(path, token string) string {
	return "https://" + m.frontendDomain + "/" + path + "/" + token
}
