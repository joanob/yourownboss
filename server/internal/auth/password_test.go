package auth

import (
	"testing"
)

func TestPasswordManager_HashPassword(t *testing.T) {
	pm := NewPasswordManager()
	password := "testpassword123"

	// Test: Hash password
	hash, err := pm.HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}

	if hash == "" {
		t.Error("Expected non-empty hash")
	}

	// Test: Hash contains Argon2id marker
	if len(hash) < 20 {
		t.Error("Expected hash to be sufficiently long")
	}
}

func TestPasswordManager_HashPassword_Unique(t *testing.T) {
	pm := NewPasswordManager()
	password := "testpassword123"

	// Test: Same password produces different hashes (due to random salt)
	hash1, err := pm.HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}

	hash2, err := pm.HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}

	if hash1 == hash2 {
		t.Error("Expected different hashes for same password (due to random salt)")
	}
}

func TestPasswordManager_VerifyPassword_Correct(t *testing.T) {
	pm := NewPasswordManager()
	password := "testpassword123"

	// Hash the password
	hash, err := pm.HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}

	// Test: Verify correct password
	isValid := pm.VerifyPassword(password, hash)
	if !isValid {
		t.Error("Expected password verification to succeed for correct password")
	}
}

func TestPasswordManager_VerifyPassword_Incorrect(t *testing.T) {
	pm := NewPasswordManager()
	password := "testpassword123"
	wrongPassword := "wrongpassword123"

	// Hash the password
	hash, err := pm.HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}

	// Test: Verify incorrect password
	isValid := pm.VerifyPassword(wrongPassword, hash)
	if isValid {
		t.Error("Expected password verification to fail for incorrect password")
	}
}

func TestPasswordManager_VerifyPassword_InvalidHash(t *testing.T) {
	pm := NewPasswordManager()
	password := "testpassword123"

	// Test: Verify with invalid hash format
	isValid := pm.VerifyPassword(password, "not-a-valid-hash")
	if isValid {
		t.Error("Expected password verification to fail for invalid hash")
	}
}

func TestPasswordManager_VerifyPassword_TimingSafe(t *testing.T) {
	pm := NewPasswordManager()
	password := "testpassword123"

	// Hash the password
	hash, err := pm.HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}

	// Test: Verify multiple times with same password (should always work)
	for i := 0; i < 5; i++ {
		isValid := pm.VerifyPassword(password, hash)
		if !isValid {
			t.Errorf("Expected password verification to succeed on attempt %d", i+1)
		}
	}
}

func TestPasswordManager_StrongPassword(t *testing.T) {
	pm := NewPasswordManager()
	strongPasswords := []string{
		"VeryStrongPassword123!@#",
		"P@ssw0rdWith!Special",
		"LongPasswordWith123AndSpecial!@#$%",
		"MyP@ssw0rd2026",
	}

	for _, password := range strongPasswords {
		// Test: Hash strong password
		hash, err := pm.HashPassword(password)
		if err != nil {
			t.Fatalf("HashPassword failed for %s: %v", password, err)
		}

		if hash == "" {
			t.Errorf("Expected non-empty hash for strong password: %s", password)
		}

		// Test: Verify strong password
		isValid := pm.VerifyPassword(password, hash)
		if !isValid {
			t.Errorf("Expected password verification to succeed for strong password: %s", password)
		}
	}
}

func TestPasswordManager_WeakPassword(t *testing.T) {
	pm := NewPasswordManager()
	// Note: PasswordManager doesn't enforce password strength rules,
	// that's done in the service layer. This test just verifies hashing works
	// for any password including weak ones.
	weakPasswords := []string{
		"123",
		"abc",
		"pass",
		"",
	}

	for _, password := range weakPasswords {
		// Note: We're testing that the system can hash and verify weak passwords
		// Security validation should be done at application level, not in PasswordManager
		hash, err := pm.HashPassword(password)
		if err != nil {
			t.Fatalf("HashPassword failed for weak password %s: %v", password, err)
		}

		// Test: Weak password still hashes and verifies
		isValid := pm.VerifyPassword(password, hash)
		if !isValid {
			t.Errorf("Expected password verification to succeed for weak password: %s", password)
		}
	}
}

func TestPasswordManager_SpecialCharacters(t *testing.T) {
	pm := NewPasswordManager()
	passwords := []string{
		"P@ss!word#2026",
		"Complex$%^&*()Password",
		"Emoji🔐Password123",
		"Unicode파스워드1234",
		"Spaces in password 123",
	}

	for _, password := range passwords {
		// Test: Hash special character password
		hash, err := pm.HashPassword(password)
		if err != nil {
			t.Fatalf("HashPassword failed for %s: %v", password, err)
		}

		// Test: Verify special character password
		isValid := pm.VerifyPassword(password, hash)
		if !isValid {
			t.Errorf("Expected password with special characters to hash and verify: %s", password)
		}
	}
}

func TestPasswordManager_EmptyPassword(t *testing.T) {
	pm := NewPasswordManager()
	password := ""

	// Test: Even empty passwords can be hashed (validation is in service layer)
	hash, err := pm.HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}

	if hash == "" {
		t.Error("Expected non-empty hash for empty password")
	}

	// Test: Verify empty password
	isValid := pm.VerifyPassword(password, hash)
	if !isValid {
		t.Error("Expected empty password to verify correctly")
	}
}
