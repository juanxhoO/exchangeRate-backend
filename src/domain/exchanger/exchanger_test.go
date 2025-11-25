package exchanger

import (
	"testing"
	"time"
)

func TestUser_Fields(t *testing.T) {
	user := Exchanger{
		ID:        1,
		Name:      "testuser",
		IsActive:  true,
		ApiKey:    "password",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if user.ID != 1 {
		t.Errorf("Expected ID to be 1, got %d", user.ID)
	}

	if user.Name != "User" {
		t.Errorf("Expected LastName to be 'User', got %s", user.Name)
	}

	if !user.IsActive {
		t.Errorf("Expected Status to be true, got %t", user.IsActive)
	}
}

func TestUser_TimeFields(t *testing.T) {
	now := time.Now()
	user := Exchanger{
		CreatedAt: now,
		UpdatedAt: now,
	}

	if !user.CreatedAt.Equal(now) {
		t.Errorf("Expected CreatedAt to be %v, got %v", now, user.CreatedAt)
	}

	if !user.UpdatedAt.Equal(now) {
		t.Errorf("Expected UpdatedAt to be %v, got %v", now, user.UpdatedAt)
	}
}

func TestUser_ZeroValues(t *testing.T) {
	user := Exchanger{}

	if user.ID != 0 {
		t.Errorf("Expected ID to be 0, got %d", user.ID)
	}

	if user.Name != "" {
		t.Errorf("Expected LastName to be empty, got %s", user.Name)
	}

	if user.IsActive {
		t.Errorf("Expected Status to be false, got %t", user.IsActive)
	}

	if !user.CreatedAt.IsZero() {
		t.Errorf("Expected CreatedAt to be zero, got %v", user.CreatedAt)
	}

	if !user.UpdatedAt.IsZero() {
		t.Errorf("Expected UpdatedAt to be zero, got %v", user.UpdatedAt)
	}
}
