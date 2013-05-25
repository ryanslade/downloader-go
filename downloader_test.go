package main

import (
	"testing"
)

func TestTitleInShowList(t *testing.T) {
	title := "Green Paradise - The Bahamas 1x4"
	shows := []string{"Breaking Bad", "Walking Dead", "Green Paradise", "Something Else"}

	if !titleInShowList(title, shows) {
		t.Error("Show should match")
	}

	shows = []string{"Breaking Bad", "Walking Dead", "green paradise", "Something Else"}
	if !titleInShowList(title, shows) {
		t.Error("Matches should ignore case")
	}
}

func TestGetLines(t *testing.T) {
	shows, err := getLines("shows.txt")

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if len(shows) != 3 {
		t.Errorf("Expected 3 shows, got %v", len(shows))
	}
}
