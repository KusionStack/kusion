package i18n

import (
	"os"
	"testing"
)

var knownTestLocale = "en_US.UTF-8"

func TestTranslation(t *testing.T) {
	err := LoadTranslations("test", func() string { return "default" })
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	result := T("test_string")
	if result != "foo" {
		t.Errorf("expected: %s, saw: %s", "foo", result)
	}
}

func TestTranslationPlural(t *testing.T) {
	err := LoadTranslations("test", func() string { return "default" })
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	result := T("test_plural", 3)
	if result != "there were 3 items" {
		t.Errorf("expected: %s, saw: %s", "there were 3 items", result)
	}

	result = T("test_plural", 1)
	if result != "there was 1 item" {
		t.Errorf("expected: %s, saw: %s", "there was 1 item", result)
	}
}

func TestTranslationUsingEnvVar(t *testing.T) {
	// We must backup and restore env vars before setting test values in tests
	// othervise we are risking to break other tests/test cases
	// which rely on the same env vars
	envVarsToBackup := []string{"LC_MESSAGES", "LANG", "LC_ALL"}
	expectedStrEnUSLocale := "baz"
	expectedStrFallback := "foo"

	testCases := []struct {
		name        string
		setenvFn    func()
		expectedStr string
	}{
		{
			name:        "Only LC_ALL is set",
			setenvFn:    func() { os.Setenv("LC_ALL", knownTestLocale) },
			expectedStr: expectedStrEnUSLocale,
		},
		{
			name:        "Only LC_MESSAGES is set",
			setenvFn:    func() { os.Setenv("LC_MESSAGES", knownTestLocale) },
			expectedStr: expectedStrEnUSLocale,
		},
		{
			name:        "Only LANG",
			setenvFn:    func() { os.Setenv("LANG", knownTestLocale) },
			expectedStr: expectedStrEnUSLocale,
		},
		{
			name: "LC_MESSAGES overrides LANG",
			setenvFn: func() {
				os.Setenv("LANG", "be_BY.UTF-8") // Unknown locale
				os.Setenv("LC_MESSAGES", knownTestLocale)
			},
			expectedStr: expectedStrEnUSLocale,
		},
		{
			name: "LC_ALL overrides LANG",
			setenvFn: func() {
				os.Setenv("LANG", "be_BY.UTF-8") // Unknown locale
				os.Setenv("LC_ALL", knownTestLocale)
			},
			expectedStr: expectedStrEnUSLocale,
		},
		{
			name: "LC_ALL overrides LC_MESSAGES",
			setenvFn: func() {
				os.Setenv("LC_MESSAGES", "be_BY.UTF-8") // Unknown locale
				os.Setenv("LC_ALL", knownTestLocale)
			},
			expectedStr: expectedStrEnUSLocale,
		},
		{
			name:        "Unknown locale in LANG",
			setenvFn:    func() { os.Setenv("LANG", "be_BY.UTF-8") },
			expectedStr: expectedStrFallback,
		},
		{
			name:        "Unknown locale in LC_MESSAGES",
			setenvFn:    func() { os.Setenv("LC_MESSAGES", "be_BY.UTF-8") },
			expectedStr: expectedStrFallback,
		},
		{
			name:        "Unknown locale in LC_ALL",
			setenvFn:    func() { os.Setenv("LC_ALL", "be_BY.UTF-8") },
			expectedStr: expectedStrFallback,
		},
		{
			name:        "Invalid env var",
			setenvFn:    func() { os.Setenv("LC_MESSAGES", "fake.locale.UTF-8") },
			expectedStr: expectedStrFallback,
		},
		{
			name:        "No env vars",
			setenvFn:    func() {},
			expectedStr: expectedStrFallback,
		},
	}

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			envBk := make(map[string]string)

			for _, envKey := range envVarsToBackup {
				if envValue := os.Getenv(envKey); envValue != "" {
					os.Unsetenv(envKey)
					envBk[envKey] = envValue
				}
			}
			defer func(bk map[string]string) {
				for k, v := range bk {
					os.Setenv(k, v)
				}
			}(envBk)

			test.setenvFn()

			err := LoadTranslations("test", nil)
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			result := T("test_string")
			if result != test.expectedStr {
				t.Errorf("expected: %s, saw: %s", test.expectedStr, result)
			}
		})
	}
}
