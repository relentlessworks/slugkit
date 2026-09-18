package slug

import "testing"

func TestGenerateBasic(t *testing.T) {
	tests := []struct {
		text string
		want string
	}{
		{"Hello World", "hello-world"},
		{"Hello, World!", "hello-world"},
		{"  Hello   World  ", "hello-world"},
		{"hello-world", "hello-world"},
		{"hello_world", "hello-world"},
		{"hello.world", "hello-world"},
		{"Hello World 123", "hello-world-123"},
		{"", ""},
		{"!!!", ""},
		{"a", "a"},
		{"a b c", "a-b-c"},
	}
	for _, tt := range tests {
		got := Generate(tt.text, DefaultOptions())
		if got != tt.want {
			t.Errorf("Generate(%q) = %q, want %q", tt.text, got, tt.want)
		}
	}
}

func TestGenerateSeparators(t *testing.T) {
	opts := DefaultOptions()
	opts.Separator = SepUnderscore
	if got := Generate("Hello World", opts); got != "hello_world" {
		t.Errorf("underscore separator: got %q", got)
	}

	opts.Separator = SepDot
	if got := Generate("Hello World", opts); got != "hello.world" {
		t.Errorf("dot separator: got %q", got)
	}
}

func TestGenerateUppercase(t *testing.T) {
	opts := DefaultOptions()
	opts.Lowercase = false
	if got := Generate("Hello World", opts); got != "Hello-World" {
		t.Errorf("uppercase: got %q", got)
	}
}

func TestGenerateMaxLength(t *testing.T) {
	opts := DefaultOptions()
	opts.MaxLength = 10
	got := Generate("Hello Wonderful World", opts)
	if len(got) > 10 {
		t.Errorf("max length exceeded: got %q (len %d)", got, len(got))
	}
	if got != "hello" {
		t.Errorf("max length: got %q, want %q", got, "hello")
	}
}

func TestGenerateUnicode(t *testing.T) {
	tests := []struct {
		text string
		want string
	}{
		{"Héllo Wörld", "hello-world"},
		{"Café Münster", "cafe-munster"},
		{"Ñoño", "nono"},
		{"Über", "uber"},
		{"Résumé", "resume"},
		{"naïve", "naive"},
		{"Zürich", "zurich"},
		{"Ångström", "angstrom"},
		{"Æsir", "aesir"},
		{"Œuvre", "oeuvre"},
		{"Æblegrøft", "aeblegroft"},
	}
	for _, tt := range tests {
		got := Generate(tt.text, DefaultOptions())
		if got != tt.want {
			t.Errorf("Generate(%q) = %q, want %q", tt.text, got, tt.want)
		}
	}
}

func TestGenerateCyrillic(t *testing.T) {
	tests := []struct {
		text string
		want string
	}{
		{"Привет мир", "privet-mir"},
		{"Москва", "moskva"},
		{"Санкт-Петербург", "sankt-peterburg"},
	}
	for _, tt := range tests {
		got := Generate(tt.text, DefaultOptions())
		if got != tt.want {
			t.Errorf("Generate(%q) = %q, want %q", tt.text, got, tt.want)
		}
	}
}

func TestGenerateGreek(t *testing.T) {
	tests := []struct {
		text string
		want string
	}{
		{"Αθήνα", "athena"},
		{"Ελληνικά", "ellenika"},
	}
	for _, tt := range tests {
		got := Generate(tt.text, DefaultOptions())
		if got != tt.want {
			t.Errorf("Generate(%q) = %q, want %q", tt.text, got, tt.want)
		}
	}
}

func TestGenerateSymbols(t *testing.T) {
	tests := []struct {
		text string
		want string
	}{
		{"C++ Programming", "c-programming"},
		{"C# vs F#", "c-vs-f"},
		{"100% Done", "100-done"},
		{"user@example.com", "user-example-com"},
		{"Price: $100", "price-100"},
		{"A & B", "a-b"},
		{"A/B/C", "a-b-c"},
	}
	for _, tt := range tests {
		got := Generate(tt.text, DefaultOptions())
		if got != tt.want {
			t.Errorf("Generate(%q) = %q, want %q", tt.text, got, tt.want)
		}
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		slug string
		sep  string
		want bool
	}{
		{"hello-world", "-", true},
		{"hello_world", "_", true},
		{"hello.world", ".", true},
		{"hello", "-", true},
		{"hello123", "-", true},
		{"", "-", false},
		{"-hello", "-", false},
		{"hello-", "-", false},
		{"hello--world", "-", false},
		{"hello world", "-", false},
		{"hello@world", "-", false},
		{"Hello-World", "-", true}, // uppercase is valid, just not default-generated
	}
	for _, tt := range tests {
		got := Validate(tt.slug, tt.sep)
		if got != tt.want {
			t.Errorf("Validate(%q, %q) = %v, want %v", tt.slug, tt.sep, got, tt.want)
		}
	}
}

func TestGenerateNumbers(t *testing.T) {
	tests := []struct {
		text string
		want string
	}{
		{"Room 101", "room-101"},
		{"Version 2.0", "version-2-0"},
		{"3.14 Pi", "3-14-pi"},
		{"1000000", "1000000"},
	}
	for _, tt := range tests {
		got := Generate(tt.text, DefaultOptions())
		if got != tt.want {
			t.Errorf("Generate(%q) = %q, want %q", tt.text, got, tt.want)
		}
	}
}

func TestGenerateMixedScripts(t *testing.T) {
	// Mixed Latin + Cyrillic
	got := Generate("Hello Привет", DefaultOptions())
	if got != "hello-privet" {
		t.Errorf("mixed scripts: got %q", got)
	}
}

func TestGenerateNewlines(t *testing.T) {
	got := Generate("Hello\nWorld", DefaultOptions())
	if got != "hello-world" {
		t.Errorf("newline: got %q", got)
	}
	got = Generate("Hello\r\nWorld", DefaultOptions())
	if got != "hello-world" {
		t.Errorf("crlf: got %q", got)
	}
}

func TestGenerateTabs(t *testing.T) {
	got := Generate("Hello\tWorld", DefaultOptions())
	if got != "hello-world" {
		t.Errorf("tab: got %q", got)
	}
}
