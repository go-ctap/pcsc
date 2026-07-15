//go:build darwin || linux

package pcsc

import (
	"errors"
	"sync"
	"testing"
)

func TestUnavailableNativeLibraryReturnsError(t *testing.T) {
	originalOpen := openNativeLibrary
	originalEnsure := ensureNativeLibrary
	t.Cleanup(func() {
		openNativeLibrary = originalOpen
		ensureNativeLibrary = originalEnsure
	})

	loadErr := errors.New("library not found")
	loadCalls := 0
	openNativeLibrary = func(string, int) (uintptr, error) {
		loadCalls++
		return 0, loadErr
	}
	ensureNativeLibrary = sync.OnceValue(loadNativeLibrary)

	var enumerateErr error
	for reader, err := range Enumerate() {
		if reader != nil {
			t.Fatalf("Enumerate reader = %#v, want nil", reader)
		}
		enumerateErr = err
	}
	if !errors.Is(enumerateErr, ErrUnavailable) {
		t.Fatalf("Enumerate error = %v, want ErrUnavailable", enumerateErr)
	}
	if !errors.Is(enumerateErr, loadErr) {
		t.Fatalf("Enumerate error = %v, want wrapped load error", enumerateErr)
	}

	card, err := Open("reader")
	if card != nil {
		t.Fatalf("Open card = %#v, want nil", card)
	}
	if !errors.Is(err, ErrUnavailable) {
		t.Fatalf("Open error = %v, want ErrUnavailable", err)
	}
	if loadCalls != 1 {
		t.Fatalf("native library load calls = %d, want 1", loadCalls)
	}
}
