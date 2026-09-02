package chatbot

import "testing"

func TestValidateSelect(t *testing.T) {
	ok := []string{
		"SELECT * FROM samples",
		"  select count(*) from test_results where flag = 'hi'  ",
		"WITH x AS (SELECT id FROM samples) SELECT * FROM x",
		"SELECT * FROM samples;",
	}
	for _, q := range ok {
		if _, err := validateSelect(q); err != nil {
			t.Errorf("validateSelect(%q) = %v, want nil", q, err)
		}
	}

	bad := []string{
		"",
		"DELETE FROM samples",
		"UPDATE samples SET status = 'done'",
		"SELECT * FROM samples; DROP TABLE samples",
		"SELECT * FROM samples -- comment",
		"SELECT * FROM samples /* x */",
		"BEGIN NULL; END;",
	}
	for _, q := range bad {
		if _, err := validateSelect(q); err == nil {
			t.Errorf("validateSelect(%q) = nil, want error", q)
		}
	}
}
