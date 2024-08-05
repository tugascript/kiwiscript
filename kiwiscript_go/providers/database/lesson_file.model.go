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

package db

import "github.com/google/uuid"

type LessonFileModel struct {
	ID       uuid.UUID
	Ext      string
	Name     string
	URL      string
	LessonID int32
}

type ToLessonFileModel interface {
	ToLessonFileModel(url string) *LessonFileModel
}

func (lf *LessonFile) ToLessonFileModel(url string) *LessonFileModel {
	return &LessonFileModel{
		ID:       lf.ID,
		Ext:      lf.Ext,
		Name:     lf.Name,
		URL:      url,
		LessonID: lf.LessonID,
	}
}
