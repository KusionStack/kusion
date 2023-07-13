package i18n

import (
	"os"
	"testing"
)

var knownTestLocale = "en_US.UTF-8"

func TestTranslation(t *testing.T) {
	err := LoadTranslations(domainTest, func() string { return LangEnUS })
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	result := T("test_string")
	if result != "foo" {
		t.Errorf("expected: %s, saw: %s", "foo", result)
	}
}

func TestTranslationPlural(t *testing.T) {
	err := LoadTranslations(domainTest, func() string { return LangEnUS })
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
	// We must back up and restore env vars before setting test values in tests
	// otherwise we are risking to break other tests/test cases
	// which rely on the same env vars
	envVarsToBackup := []string{EnvKeyLcMessages, EnvKeyLang, EnvKeyLcAll}
	expectedStrEnUSLocale := "foo"

	testCases := []struct {
		name        string
		setenvFn    func()
		expectedStr string
	}{
		{
			name:        "Only LC_ALL is set",
			setenvFn:    func() { _ = os.Setenv(EnvKeyLcAll, knownTestLocale) },
			expectedStr: expectedStrEnUSLocale,
		},
		{
			name:        "Only LC_MESSAGES is set",
			setenvFn:    func() { _ = os.Setenv(EnvKeyLcMessages, knownTestLocale) },
			expectedStr: expectedStrEnUSLocale,
		},
		{
			name:        "Only LANG",
			setenvFn:    func() { _ = os.Setenv(EnvKeyLang, knownTestLocale) },
			expectedStr: expectedStrEnUSLocale,
		},
		{
			name: "LC_MESSAGES overrides LANG",
			setenvFn: func() {
				_ = os.Setenv(EnvKeyLang, "be_BY.UTF-8") // Unknown locale
				_ = os.Setenv(EnvKeyLcMessages, knownTestLocale)
			},
			expectedStr: expectedStrEnUSLocale,
		},
		{
			name: "LC_ALL overrides LANG",
			setenvFn: func() {
				_ = os.Setenv(EnvKeyLang, "be_BY.UTF-8") // Unknown locale
				_ = os.Setenv(EnvKeyLcAll, knownTestLocale)
			},
			expectedStr: expectedStrEnUSLocale,
		},
		{
			name: "LC_ALL overrides LC_MESSAGES",
			setenvFn: func() {
				_ = os.Setenv(EnvKeyLcMessages, "be_BY.UTF-8") // Unknown locale
				_ = os.Setenv(EnvKeyLcAll, knownTestLocale)
			},
			expectedStr: expectedStrEnUSLocale,
		},
		{
			name:        "Unknown locale in LANG",
			setenvFn:    func() { _ = os.Setenv(EnvKeyLang, "be_BY.UTF-8") },
			expectedStr: expectedStrEnUSLocale,
		},
		{
			name:        "Unknown locale in LC_MESSAGES",
			setenvFn:    func() { _ = os.Setenv(EnvKeyLcMessages, "be_BY.UTF-8") },
			expectedStr: expectedStrEnUSLocale,
		},
		{
			name:        "Unknown locale in LC_ALL",
			setenvFn:    func() { _ = os.Setenv(EnvKeyLcAll, "be_BY.UTF-8") },
			expectedStr: expectedStrEnUSLocale,
		},
		{
			name:        "Invalid env var",
			setenvFn:    func() { _ = os.Setenv(EnvKeyLcMessages, "fake.locale.UTF-8") },
			expectedStr: expectedStrEnUSLocale,
		},
		{
			name:        "No env vars",
			setenvFn:    func() {},
			expectedStr: expectedStrEnUSLocale,
		},
	}

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			envBk := make(map[string]string)

			for _, envKey := range envVarsToBackup {
				if envValue := os.Getenv(envKey); envValue != "" {
					_ = os.Unsetenv(envKey)
					envBk[envKey] = envValue
				}
			}
			defer func(bk map[string]string) {
				for k, v := range bk {
					_ = os.Setenv(k, v)
				}
			}(envBk)

			test.setenvFn()

			err := LoadTranslations(domainTest, nil)
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
