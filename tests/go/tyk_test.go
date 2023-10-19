package tyk

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
)

func TestTyk(t *testing.T) {
	ctx := context.Background()

	tc, err := NewTykContainers(ctx)
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		if err := tc.CleanUp(ctx); err != nil {
			t.Fatal(err)
		}
	})

	resp, err := http.Get(fmt.Sprintf("%s/keyless-test/get", tc.URI))
	if err != nil {
		t.Fatal(err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status code %d. Got %d.", http.StatusOK, resp.StatusCode)
	}

	var data map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		t.Fatal(err)
	}

	headers, ok := data["headers"].(map[string]interface{})
	if !ok {
		t.Fatalf("Expected headers to be a map. Got '%T'", data["headers"])
	}

	if headers["Custom-Header"] != "hello world" {
		t.Fatalf("Expected header value to be 'hello world'. Got '%s'", headers["Custom-Header"])
	}
}
