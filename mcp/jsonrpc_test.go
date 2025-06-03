package mcp

import (
	"encoding/json"
	"testing"
)

func TestRequestMarshalUnmarshal(t *testing.T) {
	// Create a request
	req := NewRequest("test_method", map[string]interface{}{
		"param1": "value1",
		"param2": 42,
	}, 1)
	
	// Marshal to JSON
	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("Failed to marshal request: %v", err)
	}
	
	// Unmarshal back to request
	var parsedReq Request
	err = json.Unmarshal(data, &parsedReq)
	if err != nil {
		t.Fatalf("Failed to unmarshal request: %v", err)
	}
	
	// Check values
	if parsedReq.JSONRPC != Version {
		t.Errorf("Expected jsonrpc version %s, got %s", Version, parsedReq.JSONRPC)
	}
	
	if parsedReq.Method != "test_method" {
		t.Errorf("Expected method %s, got %s", "test_method", parsedReq.Method)
	}
	
	if parsedReq.ID != float64(1) {
		t.Errorf("Expected ID %v, got %v", 1, parsedReq.ID)
	}
	
	// Check params
	params, ok := parsedReq.Params.(map[string]interface{})
	if !ok {
		t.Fatalf("Expected params to be map[string]interface{}, got %T", parsedReq.Params)
	}
	
	if params["param1"] != "value1" {
		t.Errorf("Expected param1 to be %s, got %v", "value1", params["param1"])
	}
	
	if params["param2"] != float64(42) {
		t.Errorf("Expected param2 to be %d, got %v", 42, params["param2"])
	}
}

func TestResponseMarshalUnmarshal(t *testing.T) {
	// Create a successful response
	successResp := NewResponse(map[string]interface{}{
		"result": "success",
		"value":  123,
	}, 1)
	
	// Marshal to JSON
	successData, err := json.Marshal(successResp)
	if err != nil {
		t.Fatalf("Failed to marshal success response: %v", err)
	}
	
	// Unmarshal back to response
	var parsedSuccessResp Response
	err = json.Unmarshal(successData, &parsedSuccessResp)
	if err != nil {
		t.Fatalf("Failed to unmarshal success response: %v", err)
	}
	
	// Check values
	if parsedSuccessResp.JSONRPC != Version {
		t.Errorf("Expected jsonrpc version %s, got %s", Version, parsedSuccessResp.JSONRPC)
	}
	
	if parsedSuccessResp.ID != float64(1) {
		t.Errorf("Expected ID %v, got %v", 1, parsedSuccessResp.ID)
	}
	
	if parsedSuccessResp.Error != nil {
		t.Errorf("Expected error to be nil, got %v", parsedSuccessResp.Error)
	}
	
	// Create an error response
	errorResp := NewErrorResponse(
		InvalidRequestCode,
		InvalidRequestMsg,
		map[string]interface{}{"detail": "Invalid method"}, 
		2,
	)
	
	// Marshal to JSON
	errorData, err := json.Marshal(errorResp)
	if err != nil {
		t.Fatalf("Failed to marshal error response: %v", err)
	}
	
	// Unmarshal back to response
	var parsedErrorResp Response
	err = json.Unmarshal(errorData, &parsedErrorResp)
	if err != nil {
		t.Fatalf("Failed to unmarshal error response: %v", err)
	}
	
	// Check values
	if parsedErrorResp.JSONRPC != Version {
		t.Errorf("Expected jsonrpc version %s, got %s", Version, parsedErrorResp.JSONRPC)
	}
	
	if parsedErrorResp.ID != float64(2) {
		t.Errorf("Expected ID %v, got %v", 2, parsedErrorResp.ID)
	}
	
	if parsedErrorResp.Error == nil {
		t.Fatal("Expected error to be non-nil")
	}
	
	if parsedErrorResp.Error.Code != InvalidRequestCode {
		t.Errorf("Expected error code %d, got %d", InvalidRequestCode, parsedErrorResp.Error.Code)
	}
	
	if parsedErrorResp.Error.Message != InvalidRequestMsg {
		t.Errorf("Expected error message %s, got %s", InvalidRequestMsg, parsedErrorResp.Error.Message)
	}
}

func TestParse(t *testing.T) {
	// Valid request
	validJSON := []byte(`{"jsonrpc":"2.0","method":"test","params":{"foo":"bar"},"id":1}`)
	req, err := Parse(validJSON)
	if err != nil {
		t.Errorf("Failed to parse valid request: %v", err)
	}
	
	if req.Method != "test" {
		t.Errorf("Expected method %s, got %s", "test", req.Method)
	}
	
	// Invalid JSON
	invalidJSON := []byte(`{"jsonrpc":"2.0","method`)
	_, err = Parse(invalidJSON)
	if err == nil {
		t.Error("Expected error for invalid JSON, got nil")
	}
	
	// Invalid version
	invalidVersion := []byte(`{"jsonrpc":"1.0","method":"test","id":1}`)
	_, err = Parse(invalidVersion)
	if err == nil {
		t.Error("Expected error for invalid version, got nil")
	}
	
	// Missing method
	missingMethod := []byte(`{"jsonrpc":"2.0","id":1}`)
	_, err = Parse(missingMethod)
	if err == nil {
		t.Error("Expected error for missing method, got nil")
	}
}

func TestParseResponse(t *testing.T) {
	// Valid success response
	validSuccess := []byte(`{"jsonrpc":"2.0","result":{"foo":"bar"},"id":1}`)
	resp, err := ParseResponse(validSuccess)
	if err != nil {
		t.Errorf("Failed to parse valid success response: %v", err)
	}
	
	if resp.Error != nil {
		t.Errorf("Expected error to be nil, got %v", resp.Error)
	}
	
	// Valid error response
	validError := []byte(`{"jsonrpc":"2.0","error":{"code":-32600,"message":"Invalid request"},"id":1}`)
	resp, err = ParseResponse(validError)
	if err != nil {
		t.Errorf("Failed to parse valid error response: %v", err)
	}
	
	if resp.Error == nil {
		t.Error("Expected error to be non-nil")
	}
	
	// Invalid - both result and error
	invalidBoth := []byte(`{"jsonrpc":"2.0","result":{},"error":{"code":-32600,"message":"Invalid request"},"id":1}`)
	_, err = ParseResponse(invalidBoth)
	if err == nil {
		t.Error("Expected error for both result and error, got nil")
	}
	
	// Invalid - neither result nor error
	invalidNeither := []byte(`{"jsonrpc":"2.0","id":1}`)
	_, err = ParseResponse(invalidNeither)
	if err == nil {
		t.Error("Expected error for neither result nor error, got nil")
	}
}

func TestIsBatch(t *testing.T) {
	// Batch request
	batchJSON := []byte(`[{"jsonrpc":"2.0","method":"m1","id":1},{"jsonrpc":"2.0","method":"m2","id":2}]`)
	if !IsBatch(batchJSON) {
		t.Error("Expected batch detection to return true")
	}
	
	// Single request
	singleJSON := []byte(`{"jsonrpc":"2.0","method":"test","id":1}`)
	if IsBatch(singleJSON) {
		t.Error("Expected batch detection to return false for single request")
	}
	
	// Empty input
	emptyJSON := []byte(``)
	if IsBatch(emptyJSON) {
		t.Error("Expected batch detection to return false for empty input")
	}
}

func TestParseBatch(t *testing.T) {
	// Valid batch
	validBatch := []byte(`[
		{"jsonrpc":"2.0","method":"m1","id":1},
		{"jsonrpc":"2.0","method":"m2","id":2}
	]`)
	
	requests, err := ParseBatch(validBatch)
	if err != nil {
		t.Errorf("Failed to parse valid batch: %v", err)
	}
	
	if len(requests) != 2 {
		t.Errorf("Expected 2 requests, got %d", len(requests))
	}
	
	if requests[0].Method != "m1" || requests[1].Method != "m2" {
		t.Errorf("Methods don't match expected values")
	}
	
	// Invalid JSON
	invalidJSON := []byte(`[{"jsonrpc":"2.0",`)
	_, err = ParseBatch(invalidJSON)
	if err == nil {
		t.Error("Expected error for invalid JSON, got nil")
	}
	
	// Invalid batch - invalid version
	invalidVersion := []byte(`[
		{"jsonrpc":"2.0","method":"m1","id":1},
		{"jsonrpc":"1.0","method":"m2","id":2}
	]`)
	
	_, err = ParseBatch(invalidVersion)
	if err == nil {
		t.Error("Expected error for invalid version in batch, got nil")
	}
	
	// Invalid batch - missing method
	invalidMethod := []byte(`[
		{"jsonrpc":"2.0","method":"m1","id":1},
		{"jsonrpc":"2.0","id":2}
	]`)
	
	_, err = ParseBatch(invalidMethod)
	if err == nil {
		t.Error("Expected error for missing method in batch, got nil")
	}
}