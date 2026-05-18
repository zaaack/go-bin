package service

import (
	"testing"
	"time"
)

func TestParseExpire(t *testing.T) {
	now := time.Date(2026, 5, 19, 10, 0, 0, 0, time.UTC)
	cases := []struct {
		name string
		in   string
		want *time.Time
	}{
		{name: "never", in: "never", want: nil},
		{name: "three months", in: "3mo", want: ptrTime(now.AddDate(0, 3, 0))},
		{name: "one year", in: "1y", want: ptrTime(now.AddDate(1, 0, 0))},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ParseExpire(tc.in, now)
			if err != nil {
				t.Fatalf("ParseExpire() error = %v", err)
			}
			if tc.want == nil && got != nil {
				t.Fatalf("ParseExpire() = %v, want nil", got)
			}
			if tc.want != nil {
				if got == nil {
					t.Fatalf("ParseExpire() = nil, want %v", tc.want)
				}
				if !got.Equal(*tc.want) {
					t.Fatalf("ParseExpire() = %v, want %v", got, tc.want)
				}
			}
		})
	}
}

func TestSlugifyKeepsChinese(t *testing.T) {
	got := Slugify("测试报告 2026版.pdf")
	if got != "测试报告-2026版-pdf" {
		t.Fatalf("Slugify() = %q", got)
	}
}

func TestSummaryTwoLines(t *testing.T) {
	got := SummaryTwoLines("line1\nline2\nline3")
	if got != "line1\nline2..." {
		t.Fatalf("SummaryTwoLines() = %q", got)
	}
}

func ptrTime(ti time.Time) *time.Time { return &ti }
