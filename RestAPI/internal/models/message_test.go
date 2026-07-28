package models

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestToQueueMessage_CanonicalSchema 는 발행 페이로드가 C++ 소비측 계약을 만족하는지 고정한다.
// MainServerConsumer 는 id/sub_id 를 문자열, message 를 객체로 요구하고(MainServerConsumer.cpp),
// MainServer 와 DBWorker 는 message.content 를 읽는다. 이 구조가 깨지면 브로드캐스트가 전부 폐기된다.
func TestToQueueMessage_CanonicalSchema(t *testing.T) {
	req := &MessageRequest{
		UserID:  "test_user",
		Command: "chat_message",
		SubID:   "session_1",
		Content: "Hello!",
	}
	require.NoError(t, req.Validate())

	raw, err := req.ToQueueMessage("11111111-2222-3333-4444-555555555555").ToJSON()
	require.NoError(t, err)

	var decoded map[string]interface{}
	require.NoError(t, json.Unmarshal(raw, &decoded))

	id, ok := decoded["id"].(string)
	require.True(t, ok, "id must be a string")
	assert.Equal(t, "test_user", id, "id 는 사용자 식별자여야 한다 (메시지 UUID 가 아니다)")

	subID, ok := decoded["sub_id"].(string)
	require.True(t, ok, "sub_id must be present as a string (omitempty 금지)")
	assert.Equal(t, "session_1", subID)

	payload, ok := decoded["message"].(map[string]interface{})
	require.True(t, ok, "message must be a JSON object")
	assert.Equal(t, "chat_message", payload["command"])
	assert.Equal(t, "Hello!", payload["content"])

	publisher, ok := decoded["publisher_information"].(map[string]interface{})
	require.True(t, ok, "publisher_information must be a JSON object")
	assert.Equal(t, "11111111-2222-3333-4444-555555555555", publisher["message_id"],
		"발행측 UUID 는 publisher_information.message_id 로 추적된다")
}

// sub_id 를 생략해도 키 자체는 남아야 한다 — Consumer 가 존재와 문자열 타입을 함께 검사한다.
func TestToQueueMessage_SubIDAlwaysPresent(t *testing.T) {
	req := &MessageRequest{UserID: "u", Command: "c", Content: "x"}
	require.NoError(t, req.Validate())

	raw, err := req.ToQueueMessage("mid").ToJSON()
	require.NoError(t, err)

	var decoded map[string]interface{}
	require.NoError(t, json.Unmarshal(raw, &decoded))

	value, exists := decoded["sub_id"]
	require.True(t, exists, "sub_id key must exist even when empty")
	assert.IsType(t, "", value)
}

func TestMessageRequest_Validate(t *testing.T) {
	tests := []struct {
		name    string
		request MessageRequest
		wantErr bool
	}{
		{"valid", MessageRequest{UserID: "u", Command: "c", Content: "x"}, false},
		{"missing user_id", MessageRequest{Command: "c", Content: "x"}, true},
		{"missing command", MessageRequest{UserID: "u", Content: "x"}, true},
		{"missing content", MessageRequest{UserID: "u", Command: "c"}, true},
		{"priority too high", MessageRequest{UserID: "u", Command: "c", Content: "x", Priority: 4}, true},
		{"priority too low", MessageRequest{UserID: "u", Command: "c", Content: "x", Priority: -1}, true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.request.Validate()
			if tc.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, 2, tc.request.Priority, "priority 미지정 시 normal(2) 로 채워진다")
		})
	}
}
