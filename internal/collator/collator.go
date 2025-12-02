package collator

import (
	"cmp"
	"os"
	"strings"

	"golang.org/x/text/collate"
	"golang.org/x/text/language"
)

// New returns a locale-aware collator derived from environment variables.
func New() *collate.Collator {
	tag := localeTagFromEnv()
	if tag == "" {
		return collate.New(language.Und)
	}
	return collate.New(language.Make(tag))
}

func localeTagFromEnv() string {
	locale := cmp.Or(
		os.Getenv("LC_ALL"),
		os.Getenv("LC_COLLATE"),
		os.Getenv("LANG"),
	)

	// Strip encoding (e.g. ".UTF-8") and modifiers (e.g. "@euro")
	if idx := strings.Index(locale, "."); idx != -1 {
		locale = locale[:idx]
	}
	if idx := strings.Index(locale, "@"); idx != -1 {
		locale = locale[:idx]
	}

	locale = strings.TrimSpace(locale)
	locale = strings.ReplaceAll(locale, "_", "-")
	return locale
}
