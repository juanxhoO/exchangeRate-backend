package currency

import (
	"testing"
	"time"
)

func TestUser_Fields(t *testing.T) {
	user := Currency{
		ID:        1,
		Name:      "testuser",
		Status:    true,
		Code:      "",
		Rate:      0,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if user.ID != 1 {
		t.Errorf("Expected ID to be 1, got %d", user.ID)
	}

	if user.Name != "testuser" {
		t.Errorf("Expected UserName to be 'testuser', got %s", user.Name)
	}

	if user.Rate != 1 {
		t.Errorf("Expected Email to be '1', got %s", user.Rate)
	}

	if !user.Status {
		t.Errorf("Expected Status to be true, got %t", user.Status)
	}
}

func TestUser_TimeFields(t *testing.T) {
	now := time.Now()
	user := Currency{
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
	user := Currency{}

	if user.ID != 0 {
		t.Errorf("Expected ID to be 0, got %d", user.ID)
	}

	if user.Name != "" {
		t.Errorf("Expected UserName to be empty, got %s", user.Name)
	}

	if user.Status {
		t.Errorf("Expected Status to be false, got %t", user.Status)
	}

	if !user.CreatedAt.IsZero() {
		t.Errorf("Expected CreatedAt to be zero, got %v", user.CreatedAt)
	}

	if !user.UpdatedAt.IsZero() {
		t.Errorf("Expected UpdatedAt to be zero, got %v", user.UpdatedAt)
	}
}
