package i18n

import (
	"archive/zip"
	"bytes"
	"embed"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/chai2010/gettext-go/gettext"

	"kusionstack.io/kusion/pkg/log"
)

//go:embed translations
var translations embed.FS

var knownTranslations = map[string][]string{
	"kusion": {
		"default",
		"en_US",
		"zh_CN",
	},
	// only used for unit tests.
	"test": {
		"default",
		"en_US",
	},
}

func loadSystemLanguage() string {
	// Implements the following locale priority order: LC_ALL, LC_MESSAGES, LANG
	// Similarly to: https://www.gnu.org/software/gettext/manual/html_node/Locale-Environment-Variables.html
	langStr := os.Getenv("LC_ALL")
	if langStr == "" {
		langStr = os.Getenv("LC_MESSAGES")
	}
	if langStr == "" {
		langStr = os.Getenv("LANG")
	}

	if langStr == "" {
		log.Infof("Couldn't find the LC_ALL, LC_MESSAGES or LANG environment variables, defaulting to en_US")
		return "default"
	}
	pieces := strings.Split(langStr, ".")
	if len(pieces) != 2 {
		log.Infof("Unexpected system language (%s), defaulting to en_US", langStr)
		return "default"
	}
	return pieces[0]
}

func findLanguage(root string, getLanguageFn func() string) string {
	langStr := getLanguageFn()

	translations := knownTranslations[root]
	for ix := range translations {
		if translations[ix] == langStr {
			return langStr
		}
	}
	log.Infof("Couldn't find translations for %s, using default", langStr)
	return "default"
}

// LoadTranslations loads translation files. getLanguageFn should return a language
// string (e.g. 'en-US'). If getLanguageFn is nil, then the loadSystemLanguage function
// is used, which uses the 'LANG' environment variable.
func LoadTranslations(root string, getLanguageFn func() string) error {
	if getLanguageFn == nil {
		getLanguageFn = loadSystemLanguage
	}

	langStr := findLanguage(root, getLanguageFn)
	translationFiles := []string{
		fmt.Sprintf("%s/%s/LC_MESSAGES/kusion.po", root, langStr),
		fmt.Sprintf("%s/%s/LC_MESSAGES/kusion.mo", root, langStr),
	}

	log.Infof("Setting language to %s", langStr)
	// TODO: list the directory and load all files.
	buf := new(bytes.Buffer)
	w := zip.NewWriter(buf)

	// Make sure to check the error on Close.
	for _, file := range translationFiles {
		filename := "translations/" + file
		f, err := w.Create(file)
		if err != nil {
			return err
		}
		data, err := translations.ReadFile(filename)
		if err != nil {
			return err
		}
		if _, err := f.Write(data); err != nil {
			return nil
		}
	}
	if err := w.Close(); err != nil {
		return err
	}
	gettext.BindTextdomain("kusion", root+".zip", buf.Bytes())
	gettext.Textdomain("kusion")
	gettext.SetLocale(langStr)
	return nil
}

// T translates a string, possibly substituting arguments into it along
// the way. If len(args) is > 0, args1 is assumed to be the plural value
// and plural translation is used.
func T(defaultValue string, args ...int) string {
	if len(args) == 0 {
		return gettext.PGettext("", defaultValue)
	}
	return fmt.Sprintf(gettext.PNGettext("", defaultValue, defaultValue+".plural", args[0]),
		args[0])
}

// Errorf produces an error with a translated error string.
// Substitution is performed via the `T` function above, following
// the same rules.
func Errorf(defaultValue string, args ...int) error {
	return errors.New(T(defaultValue, args...))
}
