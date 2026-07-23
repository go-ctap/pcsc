//go:build darwin || linux || windows

package pcsc

import (
	"context"
	"errors"
	"testing"
)

func TestControlForwardsRequestAndReturnsResponse(t *testing.T) {
	original := scardControl
	t.Cleanup(func() { scardControl = original })

	input := []byte{1, 2, 3}
	scardControl = func(
		_ scardHandle,
		controlCode scardDWORD,
		inputPointer uintptr,
		inputLength scardDWORD,
		outputPointer uintptr,
		outputLength scardDWORD,
		bytesReturned *scardDWORD,
	) scardResult {
		if controlCode != 42 {
			t.Errorf("control code = %d, want 42", controlCode)
		}
		if inputPointer == 0 || inputLength != scardDWORD(len(input)) {
			t.Errorf("input pointer = 0x%x, length = %d", inputPointer, inputLength)
		}
		if outputPointer == 0 || outputLength != maxResponseBufSize {
			t.Errorf("output pointer = 0x%x, length = %d", outputPointer, outputLength)
		}

		*bytesReturned = 0

		return 0
	}

	card := &Card{handle: scardHandle(1)}
	response, err := card.Control(context.Background(), 42, input)
	if err != nil {
		t.Fatalf("Control: %v", err)
	}
	if len(response) != 0 {
		t.Fatalf("response = %x, want empty", response)
	}
}

func TestControlDoesNotRetryAfterInsufficientBuffer(t *testing.T) {
	original := scardControl
	t.Cleanup(func() { scardControl = original })

	calls := 0
	scardControl = func(
		_ scardHandle,
		_ scardDWORD,
		_ uintptr,
		_ scardDWORD,
		_ uintptr,
		_ scardDWORD,
		bytesReturned *scardDWORD,
	) scardResult {
		calls++
		*bytesReturned = maxResponseBufSize + 1

		return scardResultFromCodeForTest(scardEInsufficientBuf)
	}

	card := &Card{handle: scardHandle(1)}
	_, err := card.Control(context.Background(), 42, nil)
	var pcscErr *Error
	if !errors.As(err, &pcscErr) || pcscErr.Code != scardEInsufficientBuf {
		t.Fatalf("Control error = %v, want insufficient buffer", err)
	}
	if calls != 1 {
		t.Fatalf("SCardControl calls = %d, want 1", calls)
	}
}

func TestGetAttributeQueriesSize(t *testing.T) {
	original := scardGetAttrib
	t.Cleanup(func() { scardGetAttrib = original })

	calls := 0
	scardGetAttrib = func(_ scardHandle, attribute scardDWORD, value uintptr, length *scardDWORD) scardResult {
		calls++
		if attribute != 7 {
			t.Errorf("attribute = %d, want 7", attribute)
		}

		if value != 0 {
			t.Errorf("value pointer = 0x%x, want zero for size query", value)
		}

		*length = 0

		return 0
	}

	card := &Card{handle: scardHandle(1)}
	value, err := card.GetAttribute(Attribute(7))
	if err != nil {
		t.Fatalf("GetAttribute: %v", err)
	}
	if value != nil {
		t.Fatalf("value = %x, want nil", value)
	}
	if calls != 1 {
		t.Fatalf("SCardGetAttrib calls = %d, want 1", calls)
	}
}

func TestSetAttributeForwardsValue(t *testing.T) {
	original := scardSetAttrib
	t.Cleanup(func() { scardSetAttrib = original })

	want := []byte("value")
	scardSetAttrib = func(_ scardHandle, attribute scardDWORD, value uintptr, length scardDWORD) scardResult {
		if attribute != 7 {
			t.Errorf("attribute = %d, want 7", attribute)
		}
		if value == 0 || length != scardDWORD(len(want)) {
			t.Errorf("value pointer = 0x%x, length = %d", value, length)
		}

		return 0
	}

	card := &Card{handle: scardHandle(1)}
	if err := card.SetAttribute(Attribute(7), want); err != nil {
		t.Fatalf("SetAttribute: %v", err)
	}
}
