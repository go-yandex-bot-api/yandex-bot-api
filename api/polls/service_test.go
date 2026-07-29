package polls

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-yandex-bot-api/yandex-bot-api/core"
	"github.com/go-yandex-bot-api/yandex-bot-api/types"
)

func TestService_CreatePoll(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/bot/v1/messages/createPoll/" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		var req CreateRequest
		_ = json.NewDecoder(r.Body).Decode(&req)
		if req.Title != "Sample Poll" || len(req.Answers) != 2 {
			t.Errorf("unexpected poll payload: %+v", req)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(types.SendResponse{
			Ok:        true,
			MessageID: 301,
		})
	}))
	defer server.Close()

	client := core.NewClient("test-token", core.WithAPIURL(server.URL+"/bot/v1/"))
	svc := NewService(client)

	resp, err := svc.CreatePoll(context.Background(), CreateRequest{
		ChatID:  "c1",
		Title:   "Sample Poll",
		Answers: []string{"Yes", "No"},
	})
	if err != nil || resp.MessageID != 301 {
		t.Fatalf("unexpected CreatePoll result: %+v, err: %v", resp, err)
	}
}

func TestService_GetResults(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/bot/v1/polls/getResults/" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.URL.Query().Get("message_id") != "301" {
			t.Errorf("expected message_id=301, got %s", r.URL.Query().Get("message_id"))
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(GetResultsResponse{
			Ok:         true,
			VotedCount: 5,
			Answers:    map[string]int{"Yes": 3, "No": 2},
		})
	}))
	defer server.Close()

	client := core.NewClient("test-token", core.WithAPIURL(server.URL+"/bot/v1/"))
	svc := NewService(client)

	res, err := svc.GetResults(context.Background(), GetResultsRequest{
		ChatID:    "c1",
		MessageID: 301,
	})
	if err != nil || res.VotedCount != 5 {
		t.Fatalf("unexpected GetResults result: %+v, err: %v", res, err)
	}
}

func TestService_GetVoters(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/bot/v1/polls/getVoters/" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.URL.Query().Get("answer_id") != "0" {
			t.Errorf("expected answer_id=0, got %s", r.URL.Query().Get("answer_id"))
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(GetVotersResponse{
			Ok:         true,
			VotedCount: 1,
			Cursor:     json.RawMessage(`"next_page_cursor"`),
			Votes: []PollVoter{
				{User: types.Sender{Login: "user1"}},
			},
		})
	}))
	defer server.Close()

	client := core.NewClient("test-token", core.WithAPIURL(server.URL+"/bot/v1/"))
	svc := NewService(client)

	ansID := 0
	voters, err := svc.GetVoters(context.Background(), GetVotersRequest{
		ChatID:    "c1",
		MessageID: 301,
		AnswerID:  &ansID,
	}.WithLimit(10))
	if err != nil || voters.VotedCount != 1 {
		t.Fatalf("unexpected GetVoters result: %+v, err: %v", voters, err)
	}

	if voters.NextCursor() != "next_page_cursor" {
		t.Errorf("expected NextCursor 'next_page_cursor', got '%s'", voters.NextCursor())
	}
}
