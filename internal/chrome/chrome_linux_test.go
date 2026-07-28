//go:build linux && !android

package chrome

import (
	"strings"
	"testing"
)

// TestSecretServiceCollectionPriority documents the search order used when
// falling back from Service.SearchItems: default alias before login.
// Regression: Chrome Safe Storage on gnome-keyring is often in the default
// keyring while a separate "login" collection exists and is empty for Chrome.
func TestSecretServiceCollectionPriority(t *testing.T) {
	preferred := []string{
		defaultCollectionAlias,
		loginCollectionPath,
	}
	if preferred[0] != `/org/freedesktop/secrets/aliases/default` {
		t.Fatalf("default alias should be first, got %q", preferred[0])
	}
	if preferred[1] != `/org/freedesktop/secrets/collection/login` {
		t.Fatalf("login collection should be second, got %q", preferred[1])
	}
	// Ensure go-keyring's old hard-coded login path is not the only target.
	if strings.Contains(defaultCollectionAlias, `login`) {
		t.Fatal("default alias must not be the login collection")
	}
}
