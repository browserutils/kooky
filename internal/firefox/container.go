package firefox

import (
	"encoding/json"
	"strings"
)

// for the official "Firefox Multi-Account Containers" addon

func (s *CookieStore) initContainersMap() error {
	if s.Containers != nil || s.contFile == nil {
		return nil
	}

	conts := &containers{}
	err := json.NewDecoder(s.contFile).Decode(conts)
	if err != nil {
		return err
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
		contMap[cont.UserContextID] = name
	}
	s.Containers = contMap

	return nil
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

// TODO names of default container
