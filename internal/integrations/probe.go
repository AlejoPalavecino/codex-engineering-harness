package integrations

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

func probe(ctx context.Context, client *http.Client, endpoint, token string) (int, error) {
	// Do not forward credentials or RPC requests to redirected hosts.
	copyClient := *client
	copyClient.CheckRedirect = func(*http.Request, []*http.Request) error { return errors.New("redirección no admitida") }
	session, protocol := "", "2025-03-26"
	request := func(method string, id int, params any) (json.RawMessage, error) {
		body := map[string]any{"jsonrpc": "2.0", "method": method}
		if id != 0 {
			body["id"] = id
		}
		if params != nil {
			body["params"] = params
		}
		data, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		r, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(data))
		if err != nil {
			return nil, errors.New("URL de diagnóstico inválida")
		}
		r.Header.Set("Content-Type", "application/json")
		r.Header.Set("Accept", "application/json, text/event-stream")
		if token != "" {
			r.Header.Set("Authorization", "Bearer "+token)
		}
		if session != "" {
			r.Header.Set("Mcp-Session-Id", session)
		}
		if method != "initialize" {
			r.Header.Set("MCP-Protocol-Version", protocol)
		}
		response, err := copyClient.Do(r)
		if err != nil {
			return nil, errors.New("no se pudo conectar; revisá red, URL o tiempo de espera")
		}
		defer response.Body.Close()
		if response.StatusCode == 401 || response.StatusCode == 403 {
			return nil, errors.New("el proveedor requiere autenticación válida o permisos adicionales; revisá la guía")
		}
		if response.StatusCode < 200 || response.StatusCode >= 300 {
			return nil, fmt.Errorf("el proveedor respondió HTTP %d; reintentá o revisá su estado", response.StatusCode)
		}
		if id == 0 {
			return nil, nil
		}
		if method == "initialize" {
			session = response.Header.Get("Mcp-Session-Id")
		}
		return rpcResult(response.Body, strings.Contains(response.Header.Get("Content-Type"), "text/event-stream"), id, func(method string, requestID json.RawMessage) error {
			if method != "ping" {
				return errors.New("el servidor solicitó una operación no soportada por este diagnóstico")
			}
			data, _ := json.Marshal(map[string]any{"jsonrpc": "2.0", "id": requestID, "result": map[string]any{}})
			ping, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(data))
			if err != nil {
				return errors.New("no se pudo responder al ping MCP")
			}
			ping.Header = r.Header.Clone()
			if session != "" {
				ping.Header.Set("Mcp-Session-Id", session)
			}
			pong, err := copyClient.Do(ping)
			if err != nil {
				return errors.New("no se pudo responder al ping MCP")
			}
			defer pong.Body.Close()
			if pong.StatusCode < 200 || pong.StatusCode >= 300 {
				return errors.New("el servidor rechazó la respuesta de ping MCP")
			}
			return nil
		})
	}
	initialized, err := request("initialize", 1, map[string]any{"protocolVersion": protocol, "capabilities": map[string]any{}, "clientInfo": map[string]string{"name": "chuli-integrations", "version": "1"}})
	if err != nil {
		return 0, err
	}
	// Session cleanup has its own small deadline even when the probe timed out.
	defer func() {
		if session == "" {
			return
		}
		cleanup, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		r, err := http.NewRequestWithContext(cleanup, http.MethodDelete, endpoint, nil)
		if err != nil {
			return
		}
		r.Header.Set("Mcp-Session-Id", session)
		r.Header.Set("MCP-Protocol-Version", protocol)
		if token != "" {
			r.Header.Set("Authorization", "Bearer "+token)
		}
		if response, err := copyClient.Do(r); err == nil {
			response.Body.Close()
		}
	}()
	var info struct {
		Protocol string `json:"protocolVersion"`
		Server   struct {
			Name string `json:"name"`
		} `json:"serverInfo"`
	}
	if json.Unmarshal(initialized, &info) != nil || info.Server.Name == "" {
		return 0, errors.New("el servidor no devolvió una inicialización MCP válida")
	}
	switch info.Protocol {
	case "2024-11-05", "2025-03-26", "2025-06-18":
		protocol = info.Protocol
	default:
		return 0, errors.New("versión MCP no soportada por este diagnóstico; comprobá desde Codex")
	}
	if _, err = request("notifications/initialized", 0, nil); err != nil {
		return 0, err
	}
	listed, err := request("tools/list", 2, map[string]any{})
	if err != nil {
		return 0, err
	}
	var tools struct {
		Tools []json.RawMessage `json:"tools"`
	}
	if json.Unmarshal(listed, &tools) != nil || tools.Tools == nil {
		return 0, errors.New("el servidor no devolvió un listado MCP válido")
	}
	return len(tools.Tools), nil
}

func rpcResult(reader io.Reader, sse bool, id int, requests ...func(string, json.RawMessage) error) (json.RawMessage, error) {
	reader = io.LimitReader(reader, 1024*1024)
	decode := func(data []byte) (json.RawMessage, error) {
		var response struct {
			Version string          `json:"jsonrpc"`
			ID      int             `json:"id"`
			Result  json.RawMessage `json:"result"`
			Error   json.RawMessage `json:"error"`
		}
		if json.Unmarshal(data, &response) != nil || response.Version != "2.0" || response.ID != id || len(response.Result) == 0 || (len(response.Error) > 0 && string(response.Error) != "null") {
			return nil, errors.New("respuesta MCP inválida o error del proveedor; no se mostraron datos remotos")
		}
		return response.Result, nil
	}
	if !sse {
		data, err := io.ReadAll(reader)
		if err != nil {
			return nil, errors.New("no se pudo leer la respuesta MCP")
		}
		return decode(data)
	}
	scanner := bufio.NewScanner(reader)
	scanner.Buffer(make([]byte, 4096), 1024*1024)
	var event []string
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "data:") {
			event = append(event, strings.TrimPrefix(strings.TrimPrefix(line, "data:"), " "))
		}
		if line == "" && len(event) > 0 {
			data := []byte(strings.Join(event, "\n"))
			event = nil
			var envelope struct {
				Version string          `json:"jsonrpc"`
				ID      json.RawMessage `json:"id"`
				Method  string          `json:"method"`
			}
			if json.Unmarshal(data, &envelope) != nil || envelope.Version != "2.0" {
				return nil, errors.New("evento MCP inválido")
			}
			if envelope.Method != "" {
				if len(envelope.ID) > 0 && string(envelope.ID) != "null" && len(requests) > 0 {
					if err := requests[0](envelope.Method, envelope.ID); err != nil {
						return nil, err
					}
				}
				continue
			}
			var responseID int
			if json.Unmarshal(envelope.ID, &responseID) != nil || responseID != id {
				continue
			}
			return decode(data)
		}
	}
	return nil, errors.New("el flujo MCP terminó sin una respuesta completa")
}
