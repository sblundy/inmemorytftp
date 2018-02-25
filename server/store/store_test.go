package store

import (
	"bytes"
	"testing"
)

func TestStore_GetEmpty(t *testing.T) {
	sut := New()
	_, prs := sut.Get("test.txt")
	if prs {
		t.Error("Expected to be empty")
	}
}

func TestStore_PutGet(t *testing.T) {
	sut := New()
	sut.Put("test.txt", []byte("test value"))
	contents, prs := sut.Get("test.txt")
	if !prs {
		t.Error("Expected to be empty")
	}

	if !bytes.Equal(contents, []byte("test value")) {
		t.Error("contents don't match", contents)
	}
}

func TestStore_PutGetDifferentFile(t *testing.T) {
	sut := New()
	sut.Put("test.txt", []byte("test value"))
	_, prs := sut.Get("not-test.txt")
	if prs {
		t.Error("File not expected")
	}
}
