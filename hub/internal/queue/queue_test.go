package queue

import "testing"

func TestNormalizeLocalQueueURLUsesInternalEndpoint(t *testing.T) {
	client := &Client{endpoint: "http://localstack:4566", local: true}
	got, err := client.normalizeQueueURL("http://sqs.us-east-1.localhost.localstack.cloud:4566/000000000000/r2-cell-a-cometa-commands")
	if err != nil {
		t.Fatal(err)
	}
	want := "http://localstack:4566/000000000000/r2-cell-a-cometa-commands"
	if got != want {
		t.Fatalf("normalized queue URL = %q, want %q", got, want)
	}
}

func TestNormalizeQueueURLLeavesProductionURLUntouched(t *testing.T) {
	client := &Client{endpoint: "https://sqs.us-east-1.amazonaws.com", local: false}
	input := "https://sqs.us-east-1.amazonaws.com/123/queue"
	got, err := client.normalizeQueueURL(input)
	if err != nil {
		t.Fatal(err)
	}
	if got != input {
		t.Fatalf("production queue URL changed from %q to %q", input, got)
	}
}
