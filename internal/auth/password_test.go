package auth

import (
	"strings"
	"testing"

	"golang.org/x/crypto/bcrypt"
)

func TestHashPassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
		wantErr  bool
	}{
		{
			name:     "valid password",
			password: "mySecurePassword123",
			wantErr:  false,
		},
		{
			name:     "empty password",
			password: "",
			wantErr:  false, // bcrypt allows empty passwords
		},
		{
			name:     "long password",
			password: strings.Repeat("a", 72), // bcrypt limit is 72 bytes
			wantErr:  false,
		},
		{
			name:     "very long password (exceeds bcrypt limit)",
			password: strings.Repeat("a", 73),
			wantErr:  true, // bcrypt errors on passwords exceeding 72 bytes
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hashed, err := HashPassword(tt.password)
			if (err != nil) != tt.wantErr {
				t.Errorf("HashPassword() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				// Verify the hash is not empty
				if hashed == "" {
					t.Error("HashPassword() returned empty hash")
				}
				// Verify the hash is different from the original password
				if hashed == tt.password {
					t.Error("HashPassword() returned unhashed password")
				}
				// Verify the hash starts with bcrypt identifier
				if !strings.HasPrefix(hashed, "$2a$") && !strings.HasPrefix(hashed, "$2b$") && !strings.HasPrefix(hashed, "$2y$") {
					t.Errorf("HashPassword() hash format invalid: %s", hashed)
				}
			}
		})
	}
}

func TestHashPasswordUniqueness(t *testing.T) {
	password := "testPassword123"

	// Generate two hashes of the same password
	hash1, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword() failed: %v", err)
	}

	hash2, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword() failed: %v", err)
	}

	// Verify hashes are different due to unique salts
	if hash1 == hash2 {
		t.Error("HashPassword() should generate unique hashes for the same password")
	}

	// But both should verify correctly
	if err := VerifyPassword(hash1, password); err != nil {
		t.Errorf("VerifyPassword() failed for hash1: %v", err)
	}
	if err := VerifyPassword(hash2, password); err != nil {
		t.Errorf("VerifyPassword() failed for hash2: %v", err)
	}
}

func TestVerifyPassword(t *testing.T) {
	password := "correctPassword123"
	hashedPassword, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword() failed: %v", err)
	}

	tests := []struct {
		name           string
		hashedPassword string
		password       string
		wantErr        bool
	}{
		{
			name:           "correct password",
			hashedPassword: hashedPassword,
			password:       password,
			wantErr:        false,
		},
		{
			name:           "incorrect password",
			hashedPassword: hashedPassword,
			password:       "wrongPassword",
			wantErr:        true,
		},
		{
			name:           "empty password against hash",
			hashedPassword: hashedPassword,
			password:       "",
			wantErr:        true,
		},
		{
			name:           "case sensitive check",
			hashedPassword: hashedPassword,
			password:       "CORRECTPASSWORD123",
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := VerifyPassword(tt.hashedPassword, tt.password)
			if (err != nil) != tt.wantErr {
				t.Errorf("VerifyPassword() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr && err != bcrypt.ErrMismatchedHashAndPassword {
				t.Errorf("VerifyPassword() expected bcrypt.ErrMismatchedHashAndPassword, got %v", err)
			}
		})
	}
}

func TestVerifyPasswordWithInvalidHash(t *testing.T) {
	tests := []struct {
		name           string
		hashedPassword string
		password       string
	}{
		{
			name:           "invalid hash format",
			hashedPassword: "not-a-bcrypt-hash",
			password:       "password",
		},
		{
			name:           "empty hash",
			hashedPassword: "",
			password:       "password",
		},
		{
			name:           "truncated hash",
			hashedPassword: "$2a$10$",
			password:       "password",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := VerifyPassword(tt.hashedPassword, tt.password)
			if err == nil {
				t.Error("VerifyPassword() should return error for invalid hash")
			}
		})
	}
}

func TestHashAndVerifyWorkflow(t *testing.T) {
	// Simulate new user registration workflow
	passwords := []string{
		"user123!@#",
		"AnotherPassword456",
		"special-chars!@#$%^&*()",
	}

	for _, password := range passwords {
		t.Run("workflow_"+password, func(t *testing.T) {
			// Step 1: Hash password when creating user
			hashed, err := HashPassword(password)
			if err != nil {
				t.Fatalf("Failed to hash password: %v", err)
			}

			// Step 2: Verify correct password returns true
			if err := VerifyPassword(hashed, password); err != nil {
				t.Errorf("Failed to verify correct password: %v", err)
			}

			// Step 3: Verify incorrect password returns false
			wrongPassword := password + "wrong"
			if err := VerifyPassword(hashed, wrongPassword); err == nil {
				t.Error("Should not verify incorrect password")
			}
		})
	}
}
