package main

import "testing"

func TestOverride_String(t *testing.T) {
	flag := "default"
	override(&flag, "env_value")
	if flag != "env_value" {
		t.Errorf("expected 'env_value', got %q", flag)
	}
}

func TestOverride_String_EmptyEnv(t *testing.T) {
	flag := "default"
	override(&flag, "")
	if flag != "default" {
		t.Errorf("expected 'default', got %q", flag)
	}
}

func TestOverride_Int(t *testing.T) {
	flag := 300
	override(&flag, 600)
	if flag != 600 {
		t.Errorf("expected 600, got %d", flag)
	}
}

func TestOverride_Int_ZeroEnv(t *testing.T) {
	flag := 300
	override(&flag, 0)
	if flag != 300 {
		t.Errorf("expected 300, got %d", flag)
	}
}

func TestOverride_Bool_True(t *testing.T) {
	flag := false
	override(&flag, true)
	if !flag {
		t.Error("expected true")
	}
}

func TestOverride_Bool_FalseEnv(t *testing.T) {
	flag := true
	override(&flag, false)
	if !flag {
		t.Error("expected true (false env should not override)")
	}
}
