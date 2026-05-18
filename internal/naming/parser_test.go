package naming

import "testing"

func TestExtractDocumentID(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{
			name:  "valid document path",
			input: `{"pfad":"http://paperless.local/documents/123/"}`,
			want:  "123",
		},
		{
			name:    "invalid json",
			input:   `{"pfad":`,
			wantErr: true,
		},
		{
			name:    "missing document id",
			input:   `{"pfad":"http://paperless.local/other/123/"}`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ExtractDocumentID(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("expected %q, got %q", tt.want, got)
			}
		})
	}
}

func TestBuildFinalTitle(t *testing.T) {
	got := BuildFinalTitle("A", "B", "C", "D", "2025_03", "2025")
	want := "A_B_C_D_2025_03"
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}

	got = BuildFinalTitle("A", "AB", "B", "", "", "2025")
	want = "A_AB_2025"
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestCheckTitles(t *testing.T) {
	content := "Sender GmbH\nKalenderjahr 01.02.24\n12. März 2025\nTotal zu Ihren Lasten CHF 1'234.50"

	res := CheckTitles(
		content,
		map[string]string{"sender gmbh": "Sender"},
		map[string]string{"kalenderjahr": "Jahresbezug"},
		map[string]string{"irrelevant": "X"},
		map[string]string{},
		[]string{"Kalenderjahr"},
		[]string{"Total zu Ihren Lasten CHF"},
	)

	if res.Title1 != "Sender" {
		t.Fatalf("expected Title1 Sender, got %q", res.Title1)
	}
	if res.Title2 != "Jahresbezug" {
		t.Fatalf("expected Title2 Jahresbezug, got %q", res.Title2)
	}
	if res.Year != "2024" {
		t.Fatalf("expected Year 2024, got %q", res.Year)
	}
	if res.DateString != "2025_03" {
		t.Fatalf("expected DateString 2025_03, got %q", res.DateString)
	}
	if res.Amount != 1234.50 {
		t.Fatalf("expected Amount 1234.50, got %f", res.Amount)
	}
}

func TestHelpers(t *testing.T) {
	year, err := extractYear("Datum: 31.12.24")
	if err != nil {
		t.Fatalf("unexpected error extracting year: %v", err)
	}
	if year != "2024" {
		t.Fatalf("expected year 2024, got %q", year)
	}

	amount, err := extractCHFAmount("Bitte zahlen CHF 999.90")
	if err != nil {
		t.Fatalf("unexpected error extracting amount: %v", err)
	}
	if amount != 999.90 {
		t.Fatalf("expected amount 999.90, got %f", amount)
	}

	date := findDate("1. März 2026")
	if date != "2026_03" {
		t.Fatalf("expected date 2026_03, got %q", date)
	}
}

