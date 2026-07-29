package types

import (
	"testing"
)

func TestKeyboards_SuggestButtons(t *testing.T) {
	b1 := InlineSuggestButton{Title: "B1"}
	b2 := InlineSuggestButton{Title: "B2"}

	// 1D Array
	kb1 := NewSuggestButtons(true, b1, b2)
	if kb1.Layout != "false" || !kb1.Persist {
		t.Error("NewSuggestButtons failed")
	}

	// 2D Array
	kb2 := NewSuggestButtonsGrid(false, []InlineSuggestButton{b1}, []InlineSuggestButton{b2})
	if kb2.Layout != "true" || kb2.Persist {
		t.Error("NewSuggestButtonsGrid failed")
	}
}

func TestNewDirectives(t *testing.T) {
	d1 := NewOpenURIDirective("http://test.com")
	if d1.Type != "open_uri" || d1.URI != "http://test.com" {
		t.Error("NewOpenURIDirective failed")
	}

	d2 := NewSendMessageDirective("Hello", "payload")
	if d2.Type != "send_message" || d2.Text != "Hello" || d2.Payload == nil {
		t.Error("NewSendMessageDirective failed")
	}

	d3 := NewServerActionDirective("btn_click", map[string]string{"k": "v"})
	if d3.Type != "server_action" || d3.Name != "btn_click" {
		t.Error("NewServerActionDirective failed")
	}

	d5 := NewServerActionDirective("btn_num", 0)
	if d5.Type != "server_action" || d5.Name != "btn_num" || d5.Payload == nil {
		t.Error("NewServerActionDirective normalization failed")
	}
}

func TestKeyboardBuilder(t *testing.T) {
	kb := NewKeyboardBuilder(true).
		Columns(2).
		AddSimpleButton("B1", "action1").
		AddButton("B2", "action2", "payload2").
		AddURLButton("B3", "https://yandex.ru").
		AddTextButton("B4", "Text message")

	res := kb.Build()
	if !res.Persist {
		t.Error("expected Persist=true")
	}
	if len(res.Buttons) != 2 {
		t.Fatalf("expected 2 rows for 4 buttons with Columns(2), got %d rows", len(res.Buttons))
	}
	if len(res.Buttons[0]) != 2 || len(res.Buttons[1]) != 2 {
		t.Errorf("expected 2 buttons per row, got row0: %d, row1: %d", len(res.Buttons[0]), len(res.Buttons[1]))
	}
}
