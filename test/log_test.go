//go:build e2e
// +build e2e

package tests

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"cloud.google.com/go/pubsub"
	"gotest.tools/assert"
)

func TestLog(t *testing.T) {
	postBody := pubsub.Message{
		Attributes: map[string]string{
			"att1": "att1Val",
		},
		Data: []byte(`{"test":"123"}`),
		ID:   "1",
	}
	postBodyBytes, err := json.Marshal(postBody)
	resp, err := http.Post("http://localhost:8080/api/v1/pubsub", "application/json", strings.NewReader(string(postBodyBytes)))
	assert.NilError(t, err)
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	fmt.Printf("Body: %v\n", string(body))
	assert.NilError(t, fmt.Errorf("ttt"))
}
