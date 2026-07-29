// Package types provides the data structures for Yandex Messenger API.
package types

import "encoding/json"

// SuggestButtons represents a keyboard with buttons under a message.
type SuggestButtons struct {
	Layout  string                  `json:"layout"`  // "false" for 1D, "true" for 2D
	Persist bool                    `json:"persist"` // If true, buttons persist after the message
	Buttons [][]InlineSuggestButton `json:"-"`       // Can be []InlineSuggestButton or [][]InlineSuggestButton
}

// MarshalJSON customizes the JSON representation of SuggestButtons.
func (s *SuggestButtons) MarshalJSON() ([]byte, error) {
	if s == nil {
		return []byte("null"), nil
	}
	type Alias SuggestButtons
	if s.Layout == "true" {
		gridButtons := s.Buttons
		if gridButtons == nil {
			gridButtons = make([][]InlineSuggestButton, 0)
		}
		return json.Marshal(&struct {
			*Alias
			Buttons [][]InlineSuggestButton `json:"buttons"`
		}{
			Alias:   (*Alias)(s),
			Buttons: gridButtons,
		})
	}

	var flatButtons []InlineSuggestButton
	for _, row := range s.Buttons {
		flatButtons = append(flatButtons, row...)
	}

	if flatButtons == nil {
		flatButtons = make([]InlineSuggestButton, 0)
	}

	return json.Marshal(&struct {
		*Alias
		Buttons []InlineSuggestButton `json:"buttons"`
	}{
		Alias:   (*Alias)(s),
		Buttons: flatButtons,
	})
}

// InlineSuggestButton represents a button displayed under a message.
type InlineSuggestButton struct {
	ID         string      `json:"id,omitempty"`
	Title      string      `json:"title,omitempty"`
	Directives []Directive `json:"directives,omitempty"`
}

// ActionButtons represents action buttons (e.g. Like/Dislike) under a message.
type ActionButtons struct {
	Buttons []ActionButton `json:"buttons"`
}

// ActionButton represents a single action button.
type ActionButton struct {
	ID         string            `json:"id,omitempty"`
	Title      string            `json:"title"`
	Icon       *ActionButtonIcon `json:"icon,omitempty"`
	Directives []Directive       `json:"directives,omitempty"`
}

// ActionButtonIcon represents an icon for an action button.
type ActionButtonIcon struct {
	Type  string `json:"type"`  // Should be "messenger_icons"
	Value string `json:"value"` // "like", "pressed_like", "dislike", "pressed_dislike"
}

// Directive represents an action executed when a button is pressed.
type Directive struct {
	// Type can be "open_uri", "send_message", "server_action", "set_elements_state"
	Type           string   `json:"type"`
	URI            string   `json:"uri,omitempty"`             // for open_uri
	Text           string   `json:"text,omitempty"`            // for send_message
	Name           string   `json:"name,omitempty"`            // for server_action
	Payload        any      `json:"payload,omitempty"`         // for send_message, server_action (JSON)
	IDs            []string `json:"ids,omitempty"`             // for set_elements_state
	State          string   `json:"state,omitempty"`           // for set_elements_state ("disabled", "loading")
	TimeoutSeconds int      `json:"timeout_seconds,omitempty"` // for set_elements_state (1-60)
}

// NewSuggestButtons creates a simple row of inline buttons.
// The persist parameter controls whether the buttons disappear after being clicked:
// - If persist=false, the keyboard is "one-time" and disappears after the user clicks a button or sends any message.
// - If persist=true, the keyboard remains persistently pinned to the input field after interaction.
func NewSuggestButtons(persist bool, buttons ...InlineSuggestButton) *SuggestButtons {
	return &SuggestButtons{
		Layout:  "false",
		Persist: persist,
		Buttons: [][]InlineSuggestButton{buttons},
	}
}

// NewSuggestButtonsGrid creates a multi-row grid of inline buttons.
// The persist parameter controls whether the buttons disappear after being clicked:
// - If persist=false, the keyboard is "one-time" and disappears after the user clicks a button or sends any message.
// - If persist=true, the keyboard remains persistently pinned to the input field after interaction.
func NewSuggestButtonsGrid(persist bool, rows ...[]InlineSuggestButton) *SuggestButtons {
	return &SuggestButtons{
		Layout:  "true",
		Persist: persist,
		Buttons: rows,
	}
}

// NewOpenURIDirective creates a directive to open a URL.
func NewOpenURIDirective(uri string) Directive {
	return Directive{
		Type: "open_uri",
		URI:  uri,
	}
}

// normalizePayload converts primitive payload values (numbers, strings, booleans) into a valid JSON object.
func normalizePayload(payload any) any {
	if payload == nil {
		return nil
	}
	switch payload.(type) {
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64, bool, string:
		return map[string]any{"value": payload}
	default:
		return payload
	}
}

// NewSendMessageDirective creates a directive to send a message on behalf of the user.
func NewSendMessageDirective(text string, payload any) Directive {
	return Directive{
		Type:    "send_message",
		Text:    text,
		Payload: normalizePayload(payload),
	}
}

// NewServerActionDirective creates a directive to send a server action silently.
func NewServerActionDirective(name string, payload any) Directive {
	return Directive{
		Type:    "server_action",
		Name:    name,
		Payload: normalizePayload(payload),
	}
}

// NewSetElementsStateDirective creates a directive to change the state of elements (e.g., disable buttons).
func NewSetElementsStateDirective(ids []string, state string, timeoutSeconds int) Directive {
	return Directive{
		Type:           "set_elements_state",
		IDs:            ids,
		State:          state,
		TimeoutSeconds: timeoutSeconds,
	}
}

// NewSimpleActionButton creates a simple InlineSuggestButton for a server action callback WITHOUT any payload.
func NewSimpleActionButton(title, actionName string) InlineSuggestButton {
	return InlineSuggestButton{
		Title: title,
		Directives: []Directive{
			NewServerActionDirective(actionName, nil),
		},
	}
}

// NewActionButton creates an InlineSuggestButton that triggers a server action callback with an optional payload.
func NewActionButton(title, actionName string, payload any) InlineSuggestButton {
	return InlineSuggestButton{
		Title: title,
		Directives: []Directive{
			NewServerActionDirective(actionName, payload),
		},
	}
}

// NewURLButton creates an InlineSuggestButton that opens a public URI.
func NewURLButton(title, uri string) InlineSuggestButton {
	return InlineSuggestButton{
		Title: title,
		Directives: []Directive{
			NewOpenURIDirective(uri),
		},
	}
}

// NewTextButton creates an InlineSuggestButton that sends a text message on click.
func NewTextButton(title, text string) InlineSuggestButton {
	return InlineSuggestButton{
		Title: title,
		Directives: []Directive{
			NewSendMessageDirective(text, nil),
		},
	}
}

// KeyboardBuilder assists in dynamically constructing multi-row button grids.
type KeyboardBuilder struct {
	persist bool
	cols    int
	buttons []InlineSuggestButton
}

// NewKeyboardBuilder creates a builder for constructing button grids.
func NewKeyboardBuilder(persist bool) *KeyboardBuilder {
	const defaultCols = 2
	return &KeyboardBuilder{persist: persist, cols: defaultCols}
}

// Columns sets the maximum number of buttons per row.
func (b *KeyboardBuilder) Columns(cols int) *KeyboardBuilder {
	if cols > 0 {
		b.cols = cols
	}
	return b
}

// AddSimpleButton appends a simple action button without payload.
func (b *KeyboardBuilder) AddSimpleButton(title, actionName string) *KeyboardBuilder {
	b.buttons = append(b.buttons, NewSimpleActionButton(title, actionName))
	return b
}

// AddButton appends an action button with payload.
func (b *KeyboardBuilder) AddButton(title, actionName string, payload any) *KeyboardBuilder {
	b.buttons = append(b.buttons, NewActionButton(title, actionName, payload))
	return b
}

// AddURLButton appends a URL button.
func (b *KeyboardBuilder) AddURLButton(title, uri string) *KeyboardBuilder {
	b.buttons = append(b.buttons, NewURLButton(title, uri))
	return b
}

// AddTextButton appends a text sending button.
func (b *KeyboardBuilder) AddTextButton(title, text string) *KeyboardBuilder {
	b.buttons = append(b.buttons, NewTextButton(title, text))
	return b
}

// AddRawButton appends an existing InlineSuggestButton to the builder.
func (b *KeyboardBuilder) AddRawButton(btn InlineSuggestButton) *KeyboardBuilder {
	b.buttons = append(b.buttons, btn)
	return b
}

// AddRow appends a pre-built slice of inline buttons to the builder.
func (b *KeyboardBuilder) AddRow(buttons ...InlineSuggestButton) *KeyboardBuilder {
	b.buttons = append(b.buttons, buttons...)
	return b
}

// Build compiles the accumulated buttons into a SuggestButtons grid.
func (b *KeyboardBuilder) Build() *SuggestButtons {
	var rows [][]InlineSuggestButton
	currentRow := make([]InlineSuggestButton, 0, b.cols)

	for i, btn := range b.buttons {
		currentRow = append(currentRow, btn)
		if len(currentRow) == b.cols || i == len(b.buttons)-1 {
			rows = append(rows, currentRow)
			currentRow = make([]InlineSuggestButton, 0, b.cols)
		}
	}
	return NewSuggestButtonsGrid(b.persist, rows...)
}
