package pagination

import (
	"testing"
)

func TestPaginateSlice(t *testing.T) {
	tests := []struct {
		name       string
		totalItems int
		page       int
		limit      int
		wantStart  int
		wantEnd    int
		wantTotal  int
	}{
		{"Normal pagination page 1", 10, 1, 3, 0, 3, 4},
		{"Normal pagination page 2", 10, 2, 3, 3, 6, 4},
		{"Normal pagination last page", 10, 4, 3, 9, 10, 4},
		{"Page out of bounds (too large)", 10, 100, 3, 9, 10, 4},
		{"Page out of bounds (negative)", 10, -5, 3, 0, 3, 4},
		{"Zero items", 0, 1, 3, 0, 0, 0},
		{"Negative limit", 10, 1, -1, 0, 0, 0},
		{"Exact division", 9, 3, 3, 6, 9, 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			start, end, total := PaginateSlice(tt.totalItems, tt.page, tt.limit)
			if start != tt.wantStart || end != tt.wantEnd || total != tt.wantTotal {
				t.Errorf("PaginateSlice(%d, %d, %d) = (%d, %d, %d); want (%d, %d, %d)",
					tt.totalItems, tt.page, tt.limit, start, end, total, tt.wantStart, tt.wantEnd, tt.wantTotal)
			}
		})
	}
}

func TestNewPaginationRow(t *testing.T) {
	// Should return empty row if only 1 page
	row := NewPaginationRow(1, 1, "test_action")
	if len(row) != 0 {
		t.Errorf("Expected 0 buttons for 1 page, got %d", len(row))
	}

	// Should return 2 buttons for page 1/5 (Page indicator + Forward)
	row = NewPaginationRow(1, 5, "test_action")
	if len(row) != 2 {
		t.Errorf("Expected 2 buttons for first page, got %d", len(row))
	}
	if row[1].Title != "Вперед ➡️" {
		t.Errorf("Expected 'Forward' button, got %s", row[1].Title)
	}

	// Should return 3 buttons for page 2/5 (Back + Page indicator + Forward)
	row = NewPaginationRow(2, 5, "test_action")
	if len(row) != 3 {
		t.Errorf("Expected 3 buttons for middle page, got %d", len(row))
	}

	// Should return 2 buttons for page 5/5 (Back + Page indicator)
	row = NewPaginationRow(5, 5, "test_action")
	if len(row) != 2 {
		t.Errorf("Expected 2 buttons for last page, got %d", len(row))
	}
	if row[0].Title != "⬅️ Назад" {
		t.Errorf("Expected 'Back' button, got %s", row[0].Title)
	}
}
