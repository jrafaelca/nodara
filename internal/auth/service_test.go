package auth

import "testing"

func TestHashPassword(t *testing.T) {
	hash, err := HashPassword("a long enough password")
	if err != nil {
		t.Fatalf("HashPassword() error = %v", err)
	}
	if !VerifyPassword(hash, "a long enough password") {
		t.Fatal("VerifyPassword() did not accept the original password")
	}
	if VerifyPassword(hash, "another password") {
		t.Fatal("VerifyPassword() accepted an invalid password")
	}
}

func TestValidatePassword(t *testing.T) {
	if err := ValidatePassword("short"); err == nil {
		t.Fatal("ValidatePassword() accepted a short password")
	}
	if err := ValidatePassword("twelve chars"); err != nil {
		t.Fatalf("ValidatePassword() error = %v", err)
	}
}
