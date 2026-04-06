package cql

import (
	"fmt"
	"strings"
)

// TimeRanges maps shorthand time ranges to CQL date expressions.
var TimeRanges = map[string]string{
	"1d":  `startOfDay("-1d")`,
	"7d":  `startOfDay("-7d")`,
	"30d": `startOfDay("-30d")`,
	"90d": `startOfDay("-90d")`,
	"6M":  `startOfMonth("-6M")`,
	"1y":  `startOfYear("-1y")`,
	"2y":  `startOfYear("-2y")`,
	"5y":  `startOfYear("-5y")`,
}

// TranslateOptions holds optional CQL filter parameters.
type TranslateOptions struct {
	Spaces        []string
	Labels        []string
	TitlesOnly    bool
	ModifiedAfter string // shorthand or ISO date
	CreatedAfter  string
}

// Translate converts a natural language query into a CQL string.
func Translate(query string, opts TranslateOptions) (string, error) {
	q := strings.TrimSpace(query)
	if q == "" {
		return "", fmt.Errorf("query cannot be empty")
	}

	var clauses []string

	// Text search clause
	escaped := escape(q)
	if opts.TitlesOnly {
		clauses = append(clauses, fmt.Sprintf(`title ~ "%s"`, escaped))
	} else {
		clauses = append(clauses, fmt.Sprintf(`text ~ "%s"`, escaped))
	}

	// Space filter
	if len(opts.Spaces) > 0 {
		clauses = append(clauses, orClause("space", opts.Spaces))
	}

	// Label filter
	if len(opts.Labels) > 0 {
		clauses = append(clauses, orClause("label", opts.Labels))
	}

	// Type filter (always page + blogpost)
	clauses = append(clauses, `(type = "page" OR type = "blogpost")`)

	// Date filters
	if opts.ModifiedAfter != "" {
		clauses = append(clauses, dateClause("lastmodified", opts.ModifiedAfter))
	}
	if opts.CreatedAfter != "" {
		clauses = append(clauses, dateClause("created", opts.CreatedAfter))
	}

	return strings.Join(clauses, " AND ") + " ORDER BY lastmodified DESC", nil
}

func orClause(field string, values []string) string {
	var cleaned []string
	for _, v := range values {
		v = strings.TrimSpace(v)
		if v != "" {
			cleaned = append(cleaned, v)
		}
	}
	if len(cleaned) == 1 {
		return fmt.Sprintf(`%s = "%s"`, field, escape(cleaned[0]))
	}
	parts := make([]string, len(cleaned))
	for i, v := range cleaned {
		parts[i] = fmt.Sprintf(`%s = "%s"`, field, escape(v))
	}
	return "(" + strings.Join(parts, " OR ") + ")"
}

func dateClause(field, value string) string {
	if expr, ok := TimeRanges[value]; ok {
		return fmt.Sprintf("%s >= %s", field, expr)
	}
	return fmt.Sprintf(`%s >= "%s"`, field, value)
}

func escape(s string) string {
	s = strings.ReplaceAll(s, `"`, `\"`)
	s = strings.ReplaceAll(s, "&", " and ")
	return s
}
