package firefox

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// for the official "Firefox Multi-Account Containers" addon

// defaultContainerLabels maps Firefox l10nId values to English labels
// for the four built-in container types.
// Default containers use l10nId (localization) instead of a plain name field;
// Firefox localizes these per locale at runtime, but we use the English labels
// as a reasonable fallback since we have no access to the l10n system.
var defaultContainerLabels = map[string]string{
	`user-context-personal`: `Personal`,
	`user-context-work`:     `Work`,
	`user-context-banking`:  `Banking`,
	`user-context-shopping`: `Shopping`,
}

func (s *CookieStore) initContainersMap() {
	if s.Containers != nil || s.contFile == nil {
		return
	}
	s.Containers, s.containersErr = parseContainersJSON(s.contFile)
}

func (s *SessionCookieStore) initSessionContainersMap() {
	if s.Containers != nil {
		return
	}
	contFileName := filepath.Join(s.profileDir, `containers.json`)
	f, err := os.Open(contFileName)
	if err != nil {
		return
	}
	defer f.Close()
	s.Containers, s.containersErr = parseContainersJSON(f)
}

// parseContainersJSON reads the containers.json file written by Firefox
// (and the "Firefox Multi-Account Containers" addon).
//
// The file lists container identities. Default containers (Personal, Work,
// Banking, Shopping) use an l10nId field for localized display; custom
// containers (e.g. Facebook from Mozilla's Facebook Container extension)
// use a plain name field. Internal containers (thumbnail, webextStorageLocal)
// are identified by a "userContextIdInternal." name prefix and are skipped.
func parseContainersJSON(r io.Reader) (map[int]string, error) {
	conts := &containers{}
	err := json.NewDecoder(r).Decode(conts)
	if err != nil {
		return nil, err
	}

	contMap := make(map[int]string)
	for _, cont := range conts.Identities {
		var name string
		if cont.Name != nil {
			name = *cont.Name
			if strings.HasPrefix(name, `userContextIdInternal.`) {
				name = ``
			}
		}
		// fall back to l10nId for default containers (Personal, Work, Banking, Shopping)
		if name == `` && cont.L10nID != nil {
			name = defaultContainerLabels[*cont.L10nID]
		}
		contMap[cont.UserContextID] = name
	}
	return contMap, nil
}

type containers struct {
	Identities []struct {
		AccessKey     *string `json:"accessKey,omitempty"`
		Color         string  `json:"color"`
		Icon          string  `json:"icon"`
		L10nID        *string `json:"l10nID,omitempty"`
		Name          *string `json:"name,omitempty"`
		Public        bool    `json:"public"`
		TelemetryID   *int    `json:"telemetryId,omitempty"`
		UserContextID int     `json:"userContextId"`
	} `json:"identities"`
	LastUserContextID int `json:"lastUserContextId"`
	Version           int `json:"version"`
}

// parseOriginAttributes parses the originAttributes column from moz_cookies.
//
// Format: "^key1=value1&key2=value2" (leading ^ is stripped).
//
// Known attributes:
//   - userContextId: container tab identity (1=Personal, 2=Work, 3=Banking, 4=Shopping by default)
//   - partitionKey: CHIPS (Cookies Having Independent Partitioned State) top-level site,
//     e.g. "%28https%2Cexample.com%29" (URL-encoded "(https,example.com)")
//   - firstPartyDomain: First-Party Isolation (FPI, used by Tor Browser / privacy.firstparty.isolate)
//   - privateBrowsingId: private browsing session (not persisted to disk)
//   - geckoViewSessionContextId: Android GeckoView embedding context
func parseOriginAttributes(s string) map[string]string {
	s = strings.TrimPrefix(s, `^`)
	attrs := make(map[string]string)
	for _, part := range strings.Split(s, `&`) {
		if k, v, ok := strings.Cut(part, `=`); ok && len(k) > 0 {
			attrs[k] = v
		}
	}
	return attrs
}
