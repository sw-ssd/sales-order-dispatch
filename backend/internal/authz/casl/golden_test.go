package casl

import (
	"encoding/json"
	"os"
	"testing"
)

type goldenFile struct {
	Cases []struct {
		Name     string            `json:"name"`
		Rules    []json.RawMessage `json:"rules"`
		Identity struct {
			UserID       string `json:"user_id"`
			CompanyID    string `json:"company_id"`
			DepartmentID string `json:"department_id"`
			CustomerID   string `json:"customer_id"`
		} `json:"identity"`
		Checks []struct {
			Action   string         `json:"action"`
			Subject  string         `json:"subject"`
			Instance map[string]any `json:"instance"`
			Expect   bool           `json:"expect"`
			Query    struct {
				Denied bool `json:"denied"`
				Or     int  `json:"or"`
				And    int  `json:"and"`
			} `json:"query"`
		} `json:"checks"`
	} `json:"cases"`
}

func loadGolden(t *testing.T) goldenFile {
	t.Helper()
	raw, err := os.ReadFile("testdata/cases.json")
	if err != nil {
		t.Fatal(err)
	}
	var gf goldenFile
	if err := json.Unmarshal(raw, &gf); err != nil {
		t.Fatal(err)
	}
	if len(gf.Cases) == 0 {
		t.Fatal("golden fixture has no cases")
	}
	return gf
}

func TestGoldenCan(t *testing.T) {
	for _, c := range loadGolden(t).Cases {
		t.Run(c.Name, func(t *testing.T) {
			var rules []Rule
			for _, rr := range c.Rules {
				var jr struct {
					Action     string         `json:"action"`
					Subject    string         `json:"subject"`
					Conditions map[string]any `json:"conditions"`
					Inverted   bool           `json:"inverted"`
				}
				if err := json.Unmarshal(rr, &jr); err != nil {
					t.Fatal(err)
				}
				conds, err := ParseConditions(jr.Conditions)
				if err != nil {
					t.Fatal(err)
				}
				rules = append(rules, Rule{Action: jr.Action, Subject: jr.Subject, Conditions: conds, Inverted: jr.Inverted})
			}
			id := Identity{c.Identity.UserID, c.Identity.CompanyID, c.Identity.DepartmentID, c.Identity.CustomerID}
			e := NewEvaluator(rules, id)
			for _, chk := range c.Checks {
				got := e.Can(chk.Action, chk.Subject, chk.Instance)
				if got != chk.Expect {
					t.Errorf("can(%s, %s, %v) = %v, want %v", chk.Action, chk.Subject, chk.Instance, got, chk.Expect)
				}
			}
		})
	}
}
