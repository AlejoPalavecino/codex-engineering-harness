package integrations

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestProbeNegotiatesAndListsToolsWithoutCallingThem(t *testing.T) {
	for _, sse := range []bool{false, true} {
		t.Run(fmt.Sprint(sse), func(t *testing.T) {
			var methods []string
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Header.Get("Authorization") != "Bearer test-secret" {
					t.Error("missing environment credential")
				}
				if r.Method == "DELETE" {
					w.WriteHeader(204)
					return
				}
				var request struct {
					Method string `json:"method"`
					ID     int    `json:"id"`
				}
				json.NewDecoder(r.Body).Decode(&request)
				methods = append(methods, request.Method)
				if request.Method == "notifications/initialized" {
					w.WriteHeader(202)
					return
				}
				result := `{"protocolVersion":"2025-03-26","capabilities":{"tools":{}},"serverInfo":{"name":"fixture","version":"1"}}`
				if request.Method == "tools/list" {
					if r.Header.Get("Mcp-Session-Id") != "fixture-session" {
						t.Error("session not propagated")
					}
					result = `{"tools":[{"name":"example"}]}`
				} else {
					w.Header().Set("Mcp-Session-Id", "fixture-session")
				}
				message := fmt.Sprintf(`{"jsonrpc":"2.0","id":%d,"result":%s}`, request.ID, result)
				if sse {
					w.Header().Set("Content-Type", "text/event-stream")
					fmt.Fprint(w, "event: message\ndata: "+message+"\n\n")
				} else {
					w.Header().Set("Content-Type", "application/json")
					fmt.Fprint(w, message)
				}
			}))
			defer server.Close()
			count, err := probe(context.Background(), server.Client(), server.URL, "test-secret")
			if err != nil || count != 1 {
				t.Fatal(count, err)
			}
			if strings.Join(methods, ",") != "initialize,notifications/initialized,tools/list" {
				t.Fatal(methods)
			}
		})
	}
}

func TestProbeReportsAuthProtocolAndTimeoutSafely(t *testing.T) {
	for _, status := range []int{401, 403, 429, 500} {
		s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(status)
			fmt.Fprint(w, "secret-from-provider")
		}))
		_, err := probe(context.Background(), s.Client(), s.URL, "")
		s.Close()
		if err == nil || strings.Contains(err.Error(), "secret-from-provider") {
			t.Fatal(err)
		}
	}
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { fmt.Fprint(w, `{"jsonrpc":"2.0","id":1,"result":{}}`) }))
	defer s.Close()
	if _, err := probe(context.Background(), s.Client(), s.URL, ""); err == nil {
		t.Fatal("accepted non-MCP result")
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Nanosecond)
	defer cancel()
	if _, err := probe(ctx, s.Client(), s.URL, ""); err == nil {
		t.Fatal("missing timeout")
	}
}

func TestSSESkipsNotificationsAndUnrelatedResponses(t *testing.T) {
	stream := "data: {\"jsonrpc\":\"2.0\",\"method\":\"notifications/message\",\"params\":{}}\n\n" +
		"data: {\"jsonrpc\":\"2.0\",\"id\":\"server-ping\",\"method\":\"ping\"}\n\n" +
		"data: {\"jsonrpc\":\"2.0\",\"id\":99,\"result\":{}}\n\n" +
		"data: {\"jsonrpc\":\"2.0\",\"id\":2,\"result\":{\"tools\":[]}}\n\n"
	pingReplied := false
	result, err := rpcResult(strings.NewReader(stream), true, 2, func(method string, id json.RawMessage) error {
		if method != "ping" || string(id) != `"server-ping"` {
			t.Fatal(method, string(id))
		}
		pingReplied = true
		return nil
	})
	if err != nil || string(result) != `{"tools":[]}` {
		t.Fatal(string(result), err)
	}
	if !pingReplied {
		t.Fatal("server ping was not handled")
	}
}

func TestProbeDoesNotFollowRedirectWithCredentials(t *testing.T) {
	called := false
	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { called = true }))
	defer target.Close()
	redirect := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, target.URL, http.StatusTemporaryRedirect)
	}))
	defer redirect.Close()
	if _, err := probe(context.Background(), redirect.Client(), redirect.URL, "test-secret"); err == nil {
		t.Fatal("followed redirect")
	}
	if called {
		t.Fatal("forwarded request to redirected host")
	}
}

func TestProbeAnswersServerPingBeforeToolsResult(t *testing.T) {
	pong := make(chan struct{})
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "DELETE" {
			w.WriteHeader(204)
			return
		}
		var message struct {
			Method string          `json:"method"`
			ID     json.RawMessage `json:"id"`
			Result json.RawMessage `json:"result"`
		}
		if err := json.NewDecoder(r.Body).Decode(&message); err != nil {
			t.Error(err)
			w.WriteHeader(400)
			return
		}
		switch message.Method {
		case "initialize":
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, `{"jsonrpc":"2.0","id":1,"result":{"protocolVersion":"2025-03-26","serverInfo":{"name":"fixture"},"capabilities":{"tools":{}}}}`)
		case "notifications/initialized":
			w.WriteHeader(202)
		case "tools/list":
			w.Header().Set("Content-Type", "text/event-stream")
			fmt.Fprint(w, "data: {\"jsonrpc\":\"2.0\",\"id\":\"server-ping\",\"method\":\"ping\"}\n\n")
			w.(http.Flusher).Flush()
			select {
			case <-pong:
			case <-r.Context().Done():
				return
			}
			fmt.Fprint(w, "data: {\"jsonrpc\":\"2.0\",\"id\":2,\"result\":{\"tools\":[]}}\n\n")
		case "":
			if string(message.ID) != `"server-ping"` || string(message.Result) != "{}" {
				t.Error("invalid pong")
			}
			close(pong)
			w.WriteHeader(202)
		default:
			t.Error("unexpected method", message.Method)
			w.WriteHeader(400)
		}
	}))
	defer s.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if _, err := probe(ctx, s.Client(), s.URL, ""); err != nil {
		t.Fatal(err)
	}
}
