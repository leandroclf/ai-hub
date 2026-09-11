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

func TestResourceNameIsolatesEnvironmentAndCell(t *testing.T) {
	logicalName := "cometa-commands"
	devCellA := &Client{namespace: "dev-cell-a"}
	homCellA := &Client{namespace: "hom-cell-a"}
	devCellB := &Client{namespace: "dev-cell-b"}

	got := map[string]bool{
		devCellA.resourceName(logicalName): true,
		homCellA.resourceName(logicalName): true,
		devCellB.resourceName(logicalName): true,
	}
	if len(got) != 3 {
		t.Fatalf("broker resources collided across environment/cell namespaces: %#v", got)
	}
	for _, want := range []string{"dev-cell-a-cometa-commands", "hom-cell-a-cometa-commands", "dev-cell-b-cometa-commands"} {
		if !got[want] {
			t.Fatalf("missing namespaced broker resource %q in %#v", want, got)
		}
	}
}
