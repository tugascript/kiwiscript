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

package services

import (
	"context"

	db "github.com/kiwiscript/kiwiscript_go/providers/database"
)

func (s *Services) FindTagsBySeriesID(ctx context.Context, seriesID int32) ([]db.Tag, *ServiceError) {
	log := s.log.WithGroup("service.series.FindTagsBySeriesID").With("seriesID", seriesID)
	log.InfoContext(ctx, "Getting tags by series ID")

	tags, err := s.database.FindTagsBySeriesID(ctx, seriesID)
	if err != nil {
		log.ErrorContext(ctx, "Error getting tags", "error", err)
		return nil, FromDBError(err)
	}

	log.InfoContext(ctx, "Tags found")
	return tags, nil
}
