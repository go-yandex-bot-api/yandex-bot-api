// Package pagination provides utilities for pagination.
package pagination

import (
	"fmt"

	yabotapi "github.com/go-yandex-bot-api/yandex-bot-api"
)

// Payload represents the JSON payload sent when a pagination button is clicked.
type Payload struct {
	Page int `json:"page"`
}

// PaginateSlice is a math helper that calculates bounds for slicing an array.
// Returns the start and end indexes for slicing items[start:end], and the total number of pages.
// Safe to use: if start > len, it will return start=0, end=0.
func PaginateSlice(totalItems, page, limit int) (start, end, totalPages int) {
	if totalItems <= 0 || limit <= 0 {
		return 0, 0, 0
	}

	totalPages = (totalItems + limit - 1) / limit

	if page < 1 {
		page = 1
	}
	if page > totalPages {
		page = totalPages
	}

	start = (page - 1) * limit
	end = start + limit

	if end > totalItems {
		end = totalItems
	}

	return start, end, totalPages
}

// NewPaginationRow generates a row of buttons (InlineSuggestButton) for navigating between pages.
// actionName is the Name of the ServerAction that will be triggered.
// It creates "⬅️ Назад", "1 / 5", "Вперед ➡️" buttons.
func NewPaginationRow(currentPage, totalPages int, actionName string) []yabotapi.InlineSuggestButton {
	var row []yabotapi.InlineSuggestButton

	// Если страниц 0 или 1, навигация не нужна
	if totalPages <= 1 {
		return row
	}

	// Clamp currentPage to a valid range
	if currentPage < 1 {
		currentPage = 1
	}
	if currentPage > totalPages {
		currentPage = totalPages
	}

	// Кнопка "Назад"
	if currentPage > 1 {
		row = append(row, yabotapi.InlineSuggestButton{
			Title: "⬅️ Назад",
			Directives: []yabotapi.Directive{
				yabotapi.NewServerActionDirective(actionName, Payload{Page: currentPage - 1}),
			},
		})
	}

	// Индикатор текущей страницы (При нажатии просто обновит текущую страницу)
	row = append(row, yabotapi.InlineSuggestButton{
		Title: fmt.Sprintf("%d / %d", currentPage, totalPages),
		Directives: []yabotapi.Directive{
			yabotapi.NewServerActionDirective(actionName, Payload{Page: currentPage}),
		},
	})

	// Кнопка "Вперед"
	if currentPage < totalPages {
		row = append(row, yabotapi.InlineSuggestButton{
			Title: "Вперед ➡️",
			Directives: []yabotapi.Directive{
				yabotapi.NewServerActionDirective(actionName, Payload{Page: currentPage + 1}),
			},
		})
	}

	return row
}
