package report

import (
	"fmt"

	"golang.org/x/tools/cover"
)

type Profile = cover.Profile
type ProfileBlock = cover.ProfileBlock

func ParseProfiles(path string) ([]Profile, error) {
	profiles, err := cover.ParseProfiles(path)
	if err != nil {
		return nil, fmt.Errorf("parse profile: %w", err)
	}

	result := make([]Profile, 0, len(profiles))
	for _, profile := range profiles {
		result = append(result, *profile)
	}

	return result, nil
}
