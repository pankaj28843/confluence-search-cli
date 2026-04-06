package cql

import (
	"strings"
	"testing"
)

func TestTranslateBasic(t *testing.T) {
	result, err := Translate("deployment process", TranslateOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(result, `text ~ "deployment process"`) {
		t.Errorf("expected text clause, got: %s", result)
	}
	if !strings.Contains(result, "ORDER BY lastmodified DESC") {
		t.Errorf("expected ORDER BY, got: %s", result)
	}
}

func TestTranslateWithSpaces(t *testing.T) {
	result, err := Translate("docs", TranslateOptions{Spaces: []string{"ENG", "OPS"}})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(result, `space = "ENG"`) {
		t.Errorf("expected ENG space, got: %s", result)
	}
	if !strings.Contains(result, `space = "OPS"`) {
		t.Errorf("expected OPS space, got: %s", result)
	}
}

func TestTranslateTitlesOnly(t *testing.T) {
	result, err := Translate("api", TranslateOptions{TitlesOnly: true})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(result, `title ~ "api"`) {
		t.Errorf("expected title clause, got: %s", result)
	}
}

func TestTranslateModifiedAfter(t *testing.T) {
	result, err := Translate("docs", TranslateOptions{ModifiedAfter: "30d"})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(result, `startOfDay("-30d")`) {
		t.Errorf("expected date shorthand, got: %s", result)
	}
}

func TestTranslateModifiedAfterISO(t *testing.T) {
	result, err := Translate("docs", TranslateOptions{ModifiedAfter: "2025-01-01"})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(result, `"2025-01-01"`) {
		t.Errorf("expected ISO date, got: %s", result)
	}
}

func TestTranslateEmptyQuery(t *testing.T) {
	_, err := Translate("", TranslateOptions{})
	if err == nil {
		t.Error("expected error for empty query")
	}
}

func TestTranslateEscapesAmpersand(t *testing.T) {
	result, err := Translate("Q&A docs", TranslateOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(result, "&") {
		t.Errorf("expected ampersand escaped, got: %s", result)
	}
	if !strings.Contains(result, "Qand") {
		t.Errorf("expected 'and' replacement, got: %s", result)
	}
}

func TestTranslateTypeFilter(t *testing.T) {
	result, err := Translate("test", TranslateOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(result, `type = "page"`) {
		t.Errorf("expected page type filter, got: %s", result)
	}
	if !strings.Contains(result, `type = "blogpost"`) {
		t.Errorf("expected blogpost type filter, got: %s", result)
	}
}
