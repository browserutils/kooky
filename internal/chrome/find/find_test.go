package find

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCookieDBCandidatesPrefersExistingPaths(t *testing.T) {
	root := t.TempDir()
	prof := `Default`

	// Neither path exists → empty.
	if got := cookieDBCandidates(root, prof); len(got) != 0 {
		t.Fatalf("expected no candidates, got %v", got)
	}

	// Legacy Cookies only (common on some Chrome/Linux layouts).
	legacy := filepath.Join(root, prof, `Cookies`)
	if err := os.MkdirAll(filepath.Dir(legacy), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(legacy, []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	got := cookieDBCandidates(root, prof)
	if len(got) != 1 || got[0] != legacy {
		t.Fatalf("legacy only: got %v, want [%s]", got, legacy)
	}

	// Network/Cookies present → preferred first; both returned if both exist.
	network := filepath.Join(root, prof, `Network`, `Cookies`)
	if err := os.MkdirAll(filepath.Dir(network), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(network, []byte("y"), 0o600); err != nil {
		t.Fatal(err)
	}
	got = cookieDBCandidates(root, prof)
	if len(got) != 2 {
		t.Fatalf("both exist: got %v", got)
	}
	if got[0] != network || got[1] != legacy {
		t.Fatalf("order: got %v, want network then legacy", got)
	}

	// Network only.
	if err := os.Remove(legacy); err != nil {
		t.Fatal(err)
	}
	got = cookieDBCandidates(root, prof)
	if len(got) != 1 || got[0] != network {
		t.Fatalf("network only: got %v, want [%s]", got, network)
	}
}
