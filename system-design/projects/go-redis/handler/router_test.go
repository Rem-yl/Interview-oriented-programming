package handler

import (
	"fmt"
	"go-redis/logger"
	"go-redis/protocol"
	"go-redis/store"
	"strings"
	"testing"
)

func TestRegister(t *testing.T) {
	s := store.NewStore()
	r := NewRouter(s)

	logger.Debugf("handler: %v", r.handlers)
}

func TestPingHandler(t *testing.T) {
	s := store.NewStore()
	r := NewRouter(s)

	cmdResp := "*1\r\n$4\r\nPING\r\n"
	reader := strings.NewReader(cmdResp)
	p := protocol.NewParser(reader)
	cmd, err := p.Parse()
	if err != nil {
		t.Error(err)
	}

	resp := r.Route(cmd)
	if resp.Type != protocol.StringType {
		t.Error("resp is not string types")
	}

	logger.Debug(resp.Str)
}

func TestSetHandler(t *testing.T) {
	s := store.NewStore()
	r := NewRouter(s)

	cmdResp := "*3\r\n$3\r\nSET\r\n$4\r\nname\r\n$5\r\nAlice\r\n"
	reader := strings.NewReader(cmdResp)
	p := protocol.NewParser(reader)
	cmd, err := p.Parse()
	if err != nil {
		t.Error(err)
	}

	resp := r.Route(cmd)
	if resp.Str != "OK" {
		t.Error("resp is not correct: ", resp.Str)
	}

	value, exists := r.db.Get("name")
	if !exists {
		t.Error("key not store")
	}

	logger.Debug(value)
}

// TestGetHandler_String tests GET command for string values
func TestGetHandler_String(t *testing.T) {
	s := store.NewStore()
	r := NewRouter(s)

	// First SET a string value
	setCmdResp := "*3\r\n$3\r\nSET\r\n$4\r\nname\r\n$5\r\nAlice\r\n"
	reader := strings.NewReader(setCmdResp)
	p := protocol.NewParser(reader)
	cmd, err := p.Parse()
	if err != nil {
		t.Fatal(err)
	}
	r.Route(cmd)

	// Then GET the value
	getCmdResp := "*2\r\n$3\r\nGET\r\n$4\r\nname\r\n"
	reader = strings.NewReader(getCmdResp)
	p = protocol.NewParser(reader)
	cmd, err = p.Parse()
	if err != nil {
		t.Fatal(err)
	}

	resp := r.Route(cmd)
	if resp.Type != protocol.BulkStringType {
		t.Errorf("expected BulkStringType, got %v", resp.Type)
	}
	if resp.Str != "Alice" {
		t.Errorf("expected 'Alice', got '%s'", resp.Str)
	}
	if resp.IsNull {
		t.Error("response should not be null")
	}

	logger.Debugf("GET name: %s", resp.Str)
}

// TestGetHandler_Integer tests GET command for integer values
func TestGetHandler_Integer(t *testing.T) {
	s := store.NewStore()
	r := NewRouter(s)

	// Manually set an integer value in store
	r.db.Set("count", int64(42))

	// GET the integer value
	getCmdResp := "*2\r\n$3\r\nGET\r\n$5\r\ncount\r\n"
	reader := strings.NewReader(getCmdResp)
	p := protocol.NewParser(reader)
	cmd, err := p.Parse()
	if err != nil {
		t.Fatal(err)
	}

	resp := r.Route(cmd)
	if resp.Type != protocol.IntType {
		t.Errorf("expected IntType, got %v", resp.Type)
	}
	if resp.Int != 42 {
		t.Errorf("expected 42, got %d", resp.Int)
	}
	if resp.IsNull {
		t.Error("response should not be null")
	}

	logger.Debugf("GET count: %d", resp.Int)
}

// TestGetHandler_KeyNotExists tests GET command for non-existent key
func TestGetHandler_KeyNotExists(t *testing.T) {
	s := store.NewStore()
	r := NewRouter(s)

	// GET a non-existent key
	getCmdResp := "*2\r\n$3\r\nGET\r\n$7\r\nnotfound\r\n"
	reader := strings.NewReader(getCmdResp)
	p := protocol.NewParser(reader)
	cmd, err := p.Parse()
	if err != nil {
		t.Fatal(err)
	}

	resp := r.Route(cmd)
	if resp.Type != protocol.BulkStringType {
		t.Errorf("expected BulkStringType, got %v", resp.Type)
	}
	if !resp.IsNull {
		t.Error("response should be null for non-existent key")
	}

	logger.Debug("GET non-existent key returned null")
}

// TestGetHandler_TooFewArgs tests GET command with no arguments
func TestGetHandler_TooFewArgs(t *testing.T) {
	s := store.NewStore()
	r := NewRouter(s)

	// GET with no key
	getCmdResp := "*1\r\n$3\r\nGET\r\n"
	reader := strings.NewReader(getCmdResp)
	p := protocol.NewParser(reader)
	cmd, err := p.Parse()
	if err != nil {
		t.Fatal(err)
	}

	resp := r.Route(cmd)
	if resp.Type != protocol.ErrorType {
		t.Errorf("expected ErrorType, got %v", resp.Type)
	}
	if !strings.Contains(resp.Str, "too less args") {
		t.Errorf("expected 'too less args' error, got '%s'", resp.Str)
	}

	logger.Debugf("GET with no args error: %s", resp.Str)
}

// TestGetHandler_TooManyArgs tests GET command with too many arguments
func TestGetHandler_TooManyArgs(t *testing.T) {
	s := store.NewStore()
	r := NewRouter(s)

	// GET with multiple keys (invalid)
	getCmdResp := "*3\r\n$3\r\nGET\r\n$4\r\nkey1\r\n$4\r\nkey2\r\n"
	reader := strings.NewReader(getCmdResp)
	p := protocol.NewParser(reader)
	cmd, err := p.Parse()
	if err != nil {
		t.Fatal(err)
	}

	resp := r.Route(cmd)
	if resp.Type != protocol.ErrorType {
		t.Errorf("expected ErrorType, got %v", resp.Type)
	}
	if !strings.Contains(resp.Str, "too many args") {
		t.Errorf("expected 'too many args' error, got '%s'", resp.Str)
	}

	logger.Debugf("GET with too many args error: %s", resp.Str)
}

// TestSetHandler_TooFewArgs tests SET command with insufficient arguments
func TestSetHandler_TooFewArgs(t *testing.T) {
	s := store.NewStore()
	r := NewRouter(s)

	// SET with only key, no value
	setCmdResp := "*2\r\n$3\r\nSET\r\n$4\r\nkey1\r\n"
	reader := strings.NewReader(setCmdResp)
	p := protocol.NewParser(reader)
	cmd, err := p.Parse()
	if err != nil {
		t.Fatal(err)
	}

	resp := r.Route(cmd)
	if resp.Type != protocol.ErrorType {
		t.Errorf("expected ErrorType, got %v", resp.Type)
	}

	logger.Debugf("SET with too few args error: %s", resp.Str)
}

// TestSetHandler_TooManyArgs tests SET command with too many arguments
func TestSetHandler_TooManyArgs(t *testing.T) {
	s := store.NewStore()
	r := NewRouter(s)

	// SET with extra arguments
	setCmdResp := "*4\r\n$3\r\nSET\r\n$4\r\nkey1\r\n$5\r\nvalue\r\n$5\r\nextra\r\n"
	reader := strings.NewReader(setCmdResp)
	p := protocol.NewParser(reader)
	cmd, err := p.Parse()
	if err != nil {
		t.Fatal(err)
	}

	resp := r.Route(cmd)
	if resp.Type != protocol.ErrorType {
		t.Errorf("expected ErrorType, got %v", resp.Type)
	}

	logger.Debugf("SET with too many args error: %s", resp.Str)
}

// TestSetAndGet_MultipleKeys tests setting and getting multiple keys
func TestSetAndGet_MultipleKeys(t *testing.T) {
	s := store.NewStore()
	r := NewRouter(s)

	testCases := []struct {
		key   string
		value string
	}{
		{"user:1:name", "Alice"},
		{"user:2:name", "Bob"},
		{"user:3:name", "Charlie"},
	}

	// SET multiple keys
	for _, tc := range testCases {
		setCmdResp := fmt.Sprintf("*3\r\n$3\r\nSET\r\n$%d\r\n%s\r\n$%d\r\n%s\r\n", len(tc.key), tc.key, len(tc.value), tc.value)
		reader := strings.NewReader(setCmdResp)
		p := protocol.NewParser(reader)
		cmd, err := p.Parse()
		if err != nil {
			t.Fatal(err)
		}

		resp := r.Route(cmd)
		if resp.Str != "OK" {
			t.Errorf("SET failed for key %s: %s", tc.key, resp.Str)
		}
	}

	// GET and verify all keys
	for _, tc := range testCases {
		getCmdResp := fmt.Sprintf("*2\r\n$3\r\nGET\r\n$%d\r\n%s\r\n", len(tc.key), tc.key)
		reader := strings.NewReader(getCmdResp)
		p := protocol.NewParser(reader)
		cmd, err := p.Parse()
		if err != nil {
			t.Fatal(err)
		}

		resp := r.Route(cmd)
		if resp.Str != tc.value {
			t.Errorf("expected '%s', got '%s' for key %s", tc.value, resp.Str, tc.key)
		}
	}

	logger.Debug("Multiple SET/GET operations completed successfully")
}

// TestSetOverwrite tests overwriting an existing key
func TestSetOverwrite(t *testing.T) {
	s := store.NewStore()
	r := NewRouter(s)

	key := "config"

	// SET initial value
	setCmdResp1 := "*3\r\n$3\r\nSET\r\n$6\r\nconfig\r\n$3\r\nold\r\n"
	reader := strings.NewReader(setCmdResp1)
	p := protocol.NewParser(reader)
	cmd, err := p.Parse()
	if err != nil {
		t.Fatal(err)
	}
	r.Route(cmd)

	// SET new value (overwrite)
	setCmdResp2 := "*3\r\n$3\r\nSET\r\n$6\r\nconfig\r\n$3\r\nnew\r\n"
	reader = strings.NewReader(setCmdResp2)
	p = protocol.NewParser(reader)
	cmd, err = p.Parse()
	if err != nil {
		t.Fatal(err)
	}
	r.Route(cmd)

	// GET and verify new value
	getCmdResp := "*2\r\n$3\r\nGET\r\n$6\r\nconfig\r\n"
	reader = strings.NewReader(getCmdResp)
	p = protocol.NewParser(reader)
	cmd, err = p.Parse()
	if err != nil {
		t.Fatal(err)
	}

	resp := r.Route(cmd)
	if resp.Str != "new" {
		t.Errorf("expected 'new', got '%s'", resp.Str)
	}

	logger.Debugf("Key '%s' successfully overwritten", key)
}

// TestUnknownCommand tests routing to an unknown command
func TestUnknownCommand(t *testing.T) {
	s := store.NewStore()
	r := NewRouter(s)

	// Try an unregistered command
	cmdResp := "*2\r\n$6\r\nDELETE\r\n$4\r\nkey1\r\n"
	reader := strings.NewReader(cmdResp)
	p := protocol.NewParser(reader)
	cmd, err := p.Parse()
	if err != nil {
		t.Fatal(err)
	}

	resp := r.Route(cmd)
	if resp.Type != protocol.ErrorType {
		t.Errorf("expected ErrorType for unknown command, got %v", resp.Type)
	}

	logger.Debugf("Unknown command error: %s", resp.Str)
}
