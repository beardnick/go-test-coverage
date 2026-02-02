package report

import (
	"fmt"

	"golang.org/x/tools/cover"
)

func ParseProfiles(path string) ([]*cover.Profile, error) {
	profiles, err := cover.ParseProfiles(path)
	if err != nil {
		return nil, fmt.Errorf("parse profile: %w", err)
	}

	return profiles, nil
}
