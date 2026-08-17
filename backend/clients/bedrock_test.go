package clients

import (
	"testing"

	smithy "github.com/aws/smithy-go"
)

func TestClassifyBedrockError(t *testing.T) {
	err := &smithy.GenericAPIError{Code: "ThrottlingException"}
	msg := ClassifyBedrockError(err)
	if msg != "The AI service is busy right now. Please try again in a moment." {
		t.Errorf("unexpected message: %s", msg)
	}
}
