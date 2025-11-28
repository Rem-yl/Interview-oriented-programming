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

// ==================== DEL Command Tests ====================

// TestDelHandler_SingleKey tests deleting a single existing key
func TestDelHandler_SingleKey(t *testing.T) {
	s := store.NewStore()
	r := NewRouter(s)

	// SET a key first
	setCmdResp := "*3\r\n$3\r\nSET\r\n$4\r\nkey1\r\n$6\r\nvalue1\r\n"
	reader := strings.NewReader(setCmdResp)
	p := protocol.NewParser(reader)
	cmd, err := p.Parse()
	if err != nil {
		t.Fatal(err)
	}
	r.Route(cmd)

	// DEL the key
	delCmdResp := "*2\r\n$3\r\nDEL\r\n$4\r\nkey1\r\n"
	reader = strings.NewReader(delCmdResp)
	p = protocol.NewParser(reader)
	cmd, err = p.Parse()
	if err != nil {
		t.Fatal(err)
	}

	resp := r.Route(cmd)
	if resp.Type != protocol.IntType {
		t.Errorf("expected IntType, got %v", resp.Type)
	}
	if resp.Int != 1 {
		t.Errorf("expected 1 deleted, got %d", resp.Int)
	}

	// Verify key is deleted
	_, exists := s.Get("key1")
	if exists {
		t.Error("key should be deleted")
	}

	logger.Debugf("DEL single key: deleted %d key(s)", resp.Int)
}

// TestDelHandler_MultipleKeys tests deleting multiple keys
func TestDelHandler_MultipleKeys(t *testing.T) {
	s := store.NewStore()
	r := NewRouter(s)

	// SET multiple keys
	keys := []string{"key1", "key2", "key3"}
	for _, key := range keys {
		setCmdResp := fmt.Sprintf("*3\r\n$3\r\nSET\r\n$%d\r\n%s\r\n$5\r\nvalue\r\n", len(key), key)
		reader := strings.NewReader(setCmdResp)
		p := protocol.NewParser(reader)
		cmd, err := p.Parse()
		if err != nil {
			t.Fatal(err)
		}
		r.Route(cmd)
	}

	// DEL all keys
	delCmdResp := "*4\r\n$3\r\nDEL\r\n$4\r\nkey1\r\n$4\r\nkey2\r\n$4\r\nkey3\r\n"
	reader := strings.NewReader(delCmdResp)
	p := protocol.NewParser(reader)
	cmd, err := p.Parse()
	if err != nil {
		t.Fatal(err)
	}

	resp := r.Route(cmd)
	if resp.Type != protocol.IntType {
		t.Errorf("expected IntType, got %v", resp.Type)
	}
	if resp.Int != 3 {
		t.Errorf("expected 3 deleted, got %d", resp.Int)
	}

	logger.Debugf("DEL multiple keys: deleted %d key(s)", resp.Int)
}

// TestDelHandler_NonExistent tests deleting a non-existent key
func TestDelHandler_NonExistent(t *testing.T) {
	s := store.NewStore()
	r := NewRouter(s)

	// DEL a non-existent key
	delCmdResp := "*2\r\n$3\r\nDEL\r\n$10\r\nnonexistent\r\n"
	reader := strings.NewReader(delCmdResp)
	p := protocol.NewParser(reader)
	cmd, err := p.Parse()
	if err != nil {
		t.Fatal(err)
	}

	resp := r.Route(cmd)
	if resp.Type != protocol.IntType {
		t.Errorf("expected IntType, got %v", resp.Type)
	}
	if resp.Int != 0 {
		t.Errorf("expected 0 deleted, got %d", resp.Int)
	}

	logger.Debugf("DEL non-existent key: deleted %d key(s)", resp.Int)
}

// TestDelHandler_MixedKeys tests deleting mix of existing and non-existing keys
func TestDelHandler_MixedKeys(t *testing.T) {
	s := store.NewStore()
	r := NewRouter(s)

	// SET only one key
	setCmdResp := "*3\r\n$3\r\nSET\r\n$4\r\nkey1\r\n$5\r\nvalue\r\n"
	reader := strings.NewReader(setCmdResp)
	p := protocol.NewParser(reader)
	cmd, err := p.Parse()
	if err != nil {
		t.Fatal(err)
	}
	r.Route(cmd)

	// DEL existing and non-existing keys
	delCmdResp := "*3\r\n$3\r\nDEL\r\n$4\r\nkey1\r\n$4\r\nkey2\r\n"
	reader = strings.NewReader(delCmdResp)
	p = protocol.NewParser(reader)
	cmd, err = p.Parse()
	if err != nil {
		t.Fatal(err)
	}

	resp := r.Route(cmd)
	if resp.Type != protocol.IntType {
		t.Errorf("expected IntType, got %v", resp.Type)
	}
	if resp.Int != 1 {
		t.Errorf("expected 1 deleted, got %d", resp.Int)
	}

	logger.Debugf("DEL mixed keys: deleted %d key(s)", resp.Int)
}

// TestDelHandler_NoArgs tests DEL with no arguments
func TestDelHandler_NoArgs(t *testing.T) {
	s := store.NewStore()
	r := NewRouter(s)

	// DEL with no arguments
	delCmdResp := "*1\r\n$3\r\nDEL\r\n"
	reader := strings.NewReader(delCmdResp)
	p := protocol.NewParser(reader)
	cmd, err := p.Parse()
	if err != nil {
		t.Fatal(err)
	}

	resp := r.Route(cmd)
	if resp.Type != protocol.ErrorType {
		t.Errorf("expected ErrorType, got %v", resp.Type)
	}

	logger.Debugf("DEL with no args error: %s", resp.Str)
}

// ==================== EXISTS Command Tests ====================

// TestExistsHandler_ExistingKey tests EXISTS for an existing key
func TestExistsHandler_ExistingKey(t *testing.T) {
	s := store.NewStore()
	r := NewRouter(s)

	// SET a key first
	setCmdResp := "*3\r\n$3\r\nSET\r\n$4\r\nname\r\n$5\r\nAlice\r\n"
	reader := strings.NewReader(setCmdResp)
	p := protocol.NewParser(reader)
	cmd, err := p.Parse()
	if err != nil {
		t.Fatal(err)
	}
	r.Route(cmd)

	// EXISTS check
	existsCmdResp := "*2\r\n$6\r\nEXISTS\r\n$4\r\nname\r\n"
	reader = strings.NewReader(existsCmdResp)
	p = protocol.NewParser(reader)
	cmd, err = p.Parse()
	if err != nil {
		t.Fatal(err)
	}

	resp := r.Route(cmd)
	if resp.Type != protocol.IntType {
		t.Errorf("expected IntType, got %v", resp.Type)
	}
	if resp.Int != 1 {
		t.Errorf("expected 1 (exists), got %d", resp.Int)
	}

	logger.Debugf("EXISTS existing key: %d", resp.Int)
}

// TestExistsHandler_NonExistent tests EXISTS for a non-existent key
func TestExistsHandler_NonExistent(t *testing.T) {
	s := store.NewStore()
	r := NewRouter(s)

	// EXISTS check for non-existent key
	existsCmdResp := "*2\r\n$6\r\nEXISTS\r\n$10\r\nnonexistent\r\n"
	reader := strings.NewReader(existsCmdResp)
	p := protocol.NewParser(reader)
	cmd, err := p.Parse()
	if err != nil {
		t.Fatal(err)
	}

	resp := r.Route(cmd)
	if resp.Type != protocol.IntType {
		t.Errorf("expected IntType, got %v", resp.Type)
	}
	if resp.Int != 0 {
		t.Errorf("expected 0 (not exists), got %d", resp.Int)
	}

	logger.Debugf("EXISTS non-existent key: %d", resp.Int)
}

// TestExistsHandler_AfterDelete tests EXISTS after deleting a key
func TestExistsHandler_AfterDelete(t *testing.T) {
	s := store.NewStore()
	r := NewRouter(s)

	// SET a key
	setCmdResp := "*3\r\n$3\r\nSET\r\n$4\r\ntemp\r\n$5\r\nvalue\r\n"
	reader := strings.NewReader(setCmdResp)
	p := protocol.NewParser(reader)
	cmd, err := p.Parse()
	if err != nil {
		t.Fatal(err)
	}
	r.Route(cmd)

	// DELETE the key
	delCmdResp := "*2\r\n$3\r\nDEL\r\n$4\r\ntemp\r\n"
	reader = strings.NewReader(delCmdResp)
	p = protocol.NewParser(reader)
	cmd, err = p.Parse()
	if err != nil {
		t.Fatal(err)
	}
	r.Route(cmd)

	// EXISTS check should return 0
	existsCmdResp := "*2\r\n$6\r\nEXISTS\r\n$4\r\ntemp\r\n"
	reader = strings.NewReader(existsCmdResp)
	p = protocol.NewParser(reader)
	cmd, err = p.Parse()
	if err != nil {
		t.Fatal(err)
	}

	resp := r.Route(cmd)
	if resp.Type != protocol.IntType {
		t.Errorf("expected IntType, got %v", resp.Type)
	}
	if resp.Int != 0 {
		t.Errorf("expected 0 after delete, got %d", resp.Int)
	}

	logger.Debugf("EXISTS after delete: %d", resp.Int)
}

// ==================== KEYS Command Tests ====================

// TestKeysHandler_AllKeys tests KEYS * to get all keys
func TestKeysHandler_AllKeys(t *testing.T) {
	s := store.NewStore()
	r := NewRouter(s)

	// SET multiple keys
	testKeys := []string{"key1", "key2", "key3"}
	for _, key := range testKeys {
		setCmdResp := fmt.Sprintf("*3\r\n$3\r\nSET\r\n$%d\r\n%s\r\n$5\r\nvalue\r\n", len(key), key)
		reader := strings.NewReader(setCmdResp)
		p := protocol.NewParser(reader)
		cmd, err := p.Parse()
		if err != nil {
			t.Fatal(err)
		}
		r.Route(cmd)
	}

	// KEYS *
	keysCmdResp := "*2\r\n$4\r\nKEYS\r\n$1\r\n*\r\n"
	reader := strings.NewReader(keysCmdResp)
	p := protocol.NewParser(reader)
	cmd, err := p.Parse()
	if err != nil {
		t.Fatal(err)
	}

	resp := r.Route(cmd)
	if resp.Type != protocol.ArrayType {
		t.Errorf("expected ArrayType, got %v", resp.Type)
	}
	if len(resp.Array) != 3 {
		t.Errorf("expected 3 keys, got %d", len(resp.Array))
	}

	logger.Debugf("KEYS *: found %d key(s)", len(resp.Array))
}

// TestKeysHandler_PrefixPattern tests KEYS with prefix pattern
func TestKeysHandler_PrefixPattern(t *testing.T) {
	s := store.NewStore()
	r := NewRouter(s)

	// SET keys with different prefixes
	testData := map[string]string{
		"user:1":  "Alice",
		"user:2":  "Bob",
		"config:1": "value1",
		"config:2": "value2",
	}

	for key, value := range testData {
		setCmdResp := fmt.Sprintf("*3\r\n$3\r\nSET\r\n$%d\r\n%s\r\n$%d\r\n%s\r\n", len(key), key, len(value), value)
		reader := strings.NewReader(setCmdResp)
		p := protocol.NewParser(reader)
		cmd, err := p.Parse()
		if err != nil {
			t.Fatal(err)
		}
		r.Route(cmd)
	}

	// KEYS user:*
	keysCmdResp := "*2\r\n$4\r\nKEYS\r\n$6\r\nuser:*\r\n"
	reader := strings.NewReader(keysCmdResp)
	p := protocol.NewParser(reader)
	cmd, err := p.Parse()
	if err != nil {
		t.Fatal(err)
	}

	resp := r.Route(cmd)
	if resp.Type != protocol.ArrayType {
		t.Errorf("expected ArrayType, got %v", resp.Type)
	}
	if len(resp.Array) != 2 {
		t.Errorf("expected 2 keys matching 'user:*', got %d", len(resp.Array))
	}

	// Verify all returned keys start with "user:"
	for _, keyVal := range resp.Array {
		if !strings.HasPrefix(keyVal.Str, "user:") {
			t.Errorf("expected key to start with 'user:', got '%s'", keyVal.Str)
		}
	}

	logger.Debugf("KEYS user:*: found %d key(s)", len(resp.Array))
}

// TestKeysHandler_SuffixPattern tests KEYS with suffix pattern
func TestKeysHandler_SuffixPattern(t *testing.T) {
	s := store.NewStore()
	r := NewRouter(s)

	// SET keys with different suffixes
	testData := map[string]string{
		"firstname": "Alice",
		"lastname":  "Smith",
		"age":       "30",
		"username":  "alice123",
	}

	for key, value := range testData {
		setCmdResp := fmt.Sprintf("*3\r\n$3\r\nSET\r\n$%d\r\n%s\r\n$%d\r\n%s\r\n", len(key), key, len(value), value)
		reader := strings.NewReader(setCmdResp)
		p := protocol.NewParser(reader)
		cmd, err := p.Parse()
		if err != nil {
			t.Fatal(err)
		}
		r.Route(cmd)
	}

	// KEYS *name
	keysCmdResp := "*2\r\n$4\r\nKEYS\r\n$5\r\n*name\r\n"
	reader := strings.NewReader(keysCmdResp)
	p := protocol.NewParser(reader)
	cmd, err := p.Parse()
	if err != nil {
		t.Fatal(err)
	}

	resp := r.Route(cmd)
	if resp.Type != protocol.ArrayType {
		t.Errorf("expected ArrayType, got %v", resp.Type)
	}
	if len(resp.Array) != 3 {
		t.Errorf("expected 3 keys matching '*name', got %d", len(resp.Array))
	}

	// Verify all returned keys end with "name"
	for _, keyVal := range resp.Array {
		if !strings.HasSuffix(keyVal.Str, "name") {
			t.Errorf("expected key to end with 'name', got '%s'", keyVal.Str)
		}
	}

	logger.Debugf("KEYS *name: found %d key(s)", len(resp.Array))
}

// TestKeysHandler_NoMatch tests KEYS with pattern that matches nothing
func TestKeysHandler_NoMatch(t *testing.T) {
	s := store.NewStore()
	r := NewRouter(s)

	// SET some keys
	setCmdResp := "*3\r\n$3\r\nSET\r\n$4\r\nkey1\r\n$5\r\nvalue\r\n"
	reader := strings.NewReader(setCmdResp)
	p := protocol.NewParser(reader)
	cmd, err := p.Parse()
	if err != nil {
		t.Fatal(err)
	}
	r.Route(cmd)

	// KEYS nonexistent*
	keysCmdResp := "*2\r\n$4\r\nKEYS\r\n$12\r\nnonexistent*\r\n"
	reader = strings.NewReader(keysCmdResp)
	p = protocol.NewParser(reader)
	cmd, err = p.Parse()
	if err != nil {
		t.Fatal(err)
	}

	resp := r.Route(cmd)
	if resp.Type != protocol.ArrayType {
		t.Errorf("expected ArrayType, got %v", resp.Type)
	}
	if len(resp.Array) != 0 {
		t.Errorf("expected 0 keys (empty array), got %d", len(resp.Array))
	}

	logger.Debugf("KEYS nonexistent*: found %d key(s)", len(resp.Array))
}

// TestKeysHandler_EmptyStore tests KEYS on an empty store
func TestKeysHandler_EmptyStore(t *testing.T) {
	s := store.NewStore()
	r := NewRouter(s)

	// KEYS * on empty store
	keysCmdResp := "*2\r\n$4\r\nKEYS\r\n$1\r\n*\r\n"
	reader := strings.NewReader(keysCmdResp)
	p := protocol.NewParser(reader)
	cmd, err := p.Parse()
	if err != nil {
		t.Fatal(err)
	}

	resp := r.Route(cmd)
	if resp.Type != protocol.ArrayType {
		t.Errorf("expected ArrayType, got %v", resp.Type)
	}
	if len(resp.Array) != 0 {
		t.Errorf("expected 0 keys (empty array), got %d", len(resp.Array))
	}

	logger.Debugf("KEYS * on empty store: found %d key(s)", len(resp.Array))
}

// TestKeysHandler_ExactMatch tests KEYS with exact key name (no wildcards)
func TestKeysHandler_ExactMatch(t *testing.T) {
	s := store.NewStore()
	r := NewRouter(s)

	// SET multiple keys
	testKeys := []string{"exactkey", "otherkey", "anotherkey"}
	for _, key := range testKeys {
		setCmdResp := fmt.Sprintf("*3\r\n$3\r\nSET\r\n$%d\r\n%s\r\n$5\r\nvalue\r\n", len(key), key)
		reader := strings.NewReader(setCmdResp)
		p := protocol.NewParser(reader)
		cmd, err := p.Parse()
		if err != nil {
			t.Fatal(err)
		}
		r.Route(cmd)
	}

	// KEYS exactkey (exact match, no wildcard)
	keysCmdResp := "*2\r\n$4\r\nKEYS\r\n$8\r\nexactkey\r\n"
	reader := strings.NewReader(keysCmdResp)
	p := protocol.NewParser(reader)
	cmd, err := p.Parse()
	if err != nil {
		t.Fatal(err)
	}

	resp := r.Route(cmd)
	if resp.Type != protocol.ArrayType {
		t.Errorf("expected ArrayType, got %v", resp.Type)
	}
	if len(resp.Array) != 1 {
		t.Errorf("expected 1 key for exact match, got %d", len(resp.Array))
	}
	if len(resp.Array) > 0 && resp.Array[0].Str != "exactkey" {
		t.Errorf("expected 'exactkey', got '%s'", resp.Array[0].Str)
	}

	logger.Debugf("KEYS exactkey: found %d key(s)", len(resp.Array))
}

// TestKeysHandler_NoArgs tests KEYS with no arguments
func TestKeysHandler_NoArgs(t *testing.T) {
	s := store.NewStore()
	r := NewRouter(s)

	// KEYS with no pattern
	keysCmdResp := "*1\r\n$4\r\nKEYS\r\n"
	reader := strings.NewReader(keysCmdResp)
	p := protocol.NewParser(reader)
	cmd, err := p.Parse()
	if err != nil {
		t.Fatal(err)
	}

	resp := r.Route(cmd)
	if resp.Type != protocol.ErrorType {
		t.Errorf("expected ErrorType, got %v", resp.Type)
	}

	logger.Debugf("KEYS with no args error: %s", resp.Str)
}
