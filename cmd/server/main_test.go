package main

import "testing"

func ptr[T any](v T) *T {
	return &v
}

func TestOverride_String(t *testing.T) {
	flag := "default"
	override(&flag, ptr("env_value"))
	if flag != "env_value" {
		t.Errorf("expected 'env_value', got %q", flag)
	}
}

func TestOverride_String_Nil_KeepsFlag(t *testing.T) {
	flag := "default"
	override(&flag, (*string)(nil))
	if flag != "default" {
		t.Errorf("expected 'default', got %q", flag)
	}
}

func TestOverride_String_Empty_OverridesFlag(t *testing.T) {
	flag := "default"
	override(&flag, ptr(""))
	if flag != "" {
		t.Errorf("expected '', got %q", flag)
	}
}

func TestOverride_Int(t *testing.T) {
	flag := 300
	override(&flag, ptr(600))
	if flag != 600 {
		t.Errorf("expected 600, got %d", flag)
	}
}

func TestOverride_Int_Zero_OverridesFlag(t *testing.T) {
	flag := 300
	override(&flag, ptr(0))
	if flag != 0 {
		t.Errorf("expected 0, got %d", flag)
	}
}

func TestOverride_Int_Nil_KeepsFlag(t *testing.T) {
	flag := 300
	override(&flag, (*int)(nil))
	if flag != 300 {
		t.Errorf("expected 300, got %d", flag)
	}
}

func TestOverride_Bool_True(t *testing.T) {
	flag := false
	override(&flag, ptr(true))
	if !flag {
		t.Error("expected true")
	}
}

func TestOverride_Bool_False_OverridesFlag(t *testing.T) {
	flag := true
	override(&flag, ptr(false))
	if flag {
		t.Error("expected false")
	}
}

func TestOverride_Bool_Nil_KeepsFlag(t *testing.T) {
	flag := true
	override(&flag, (*bool)(nil))
	if !flag {
		t.Error("expected true")
	}
}
