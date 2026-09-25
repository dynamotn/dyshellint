package lint

import (
	"encoding/json"
	"fmt"
	"io"
)

// Report prints findings in the requested format and answers whether the run
// has to fail.
type Report struct {
	Findings []Finding
	// Format is "text" or "json".
	Format string
	// WarningsAsErrors turns the ⚠️ CONSIDER tier into a failure too.
	WarningsAsErrors bool
}

// Write prints the report and returns true when the run should fail.
func (r Report) Write(w io.Writer) (failed bool, err error) {
	Sort(r.Findings)
	var errors, warnings int
	for _, finding := range r.Findings {
		if finding.Severity == SeverityError {
			errors++
		} else {
			warnings++
		}
	}
	if r.Format == "json" {
		encoder := json.NewEncoder(w)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(struct {
			Findings []Finding `json:"findings"`
			Errors   int       `json:"errors"`
			Warnings int       `json:"warnings"`
		}{r.Findings, errors, warnings}); err != nil {
			return true, err
		}
		return errors > 0 || (r.WarningsAsErrors && warnings > 0), nil
	}
	for _, finding := range r.Findings {
		if _, err := fmt.Fprintln(w, finding.String()); err != nil {
			return true, err
		}
	}
	if len(r.Findings) == 0 {
		if _, err := fmt.Fprintln(w, "No style guide violations found."); err != nil {
			return true, err
		}
		return false, nil
	}
	if _, err := fmt.Fprintf(w, "\n%d error(s), %d warning(s)\n", errors, warnings); err != nil {
		return true, err
	}
	return errors > 0 || (r.WarningsAsErrors && warnings > 0), nil
}
