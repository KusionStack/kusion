package i18n

import (
	"archive/zip"
	"bytes"
	"embed"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/chai2010/gettext-go"

	"kusionstack.io/kusion/pkg/log"
)

const (
	DomainKusion = "kusion"
	domainTest   = "test"

	LangEnUS      = "en_US"
	LangZhCN      = "zh_CN"
	LangValueEnUS = "en_US.UTF-8"
	LangValueZhCN = "zh_CN.UTF-8"

	EnvKeyLanguage   = "LANGUAGE"
	EnvKeyLcAll      = "LC_ALL"
	EnvKeyLcMessages = "LC_MESSAGES"
	EnvKeyLang       = "LANG"

	transEmbedDir = "translations"
	transFileDir  = "LC_MESSAGES"
	poSuffix      = ".po"
	moSuffix      = ".mo"
	zipSuffix     = ".zip"
	pluralSuffix  = ".plural"
)

//go:embed translations
var translations embed.FS

var knownTranslations = map[string][]string{
	DomainKusion: {LangEnUS, LangZhCN},
	// only used for unit tests.
	domainTest: {LangEnUS},
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
		fmt.Sprintf("%s/%s/%s/%s%s", root, langStr, transFileDir, root, poSuffix),
		fmt.Sprintf("%s/%s/%s/%s%s", root, langStr, transFileDir, root, moSuffix),
	}

	log.Infof("Setting language to %s", langStr)
	log.Infof("load translation files %s, %s", translationFiles[0], translationFiles[1])
	buf := new(bytes.Buffer)
	w := zip.NewWriter(buf)

	// Make sure to check the error on Close.
	for _, file := range translationFiles {
		filename := transEmbedDir + "/" + file
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
	gettext.BindLocale(gettext.New(root, root+zipSuffix, buf.Bytes()))
	gettext.SetDomain(root)
	gettext.SetLanguage(langStr)
	return nil
}

func loadSystemLanguage() string {
	// implements the following locale priority order: LANGUAGE, LC_ALL, LC_MESSAGES, LANG
	// similarly to: https://www.gnu.org/software/gettext/manual/html_node/Locale-Environment-Variables.html
	langStr := os.Getenv(EnvKeyLanguage)
	if langStr == "" {
		langStr = os.Getenv(EnvKeyLcAll)
	}
	if langStr == "" {
		langStr = os.Getenv(EnvKeyLcMessages)
	}
	if langStr == "" {
		langStr = os.Getenv(EnvKeyLang)
	}

	if langStr == "" {
		log.Infof("Couldn't find the %s, %s, %s or %s environment variables, defaulting to %s",
			EnvKeyLanguage, EnvKeyLcAll, EnvKeyLcMessages, EnvKeyLang, LangEnUS)
		return LangEnUS
	}
	// the value is kind like en_US.UTF-8
	pieces := strings.Split(langStr, ".")
	if len(pieces) != 2 {
		log.Infof("Unexpected system language (%s), defaulting to %s", langStr, LangEnUS)
		return LangEnUS
	}
	return pieces[0]
}

func findLanguage(root string, getLanguageFn func() string) string {
	langStr := getLanguageFn()

	trans := knownTranslations[root]
	for ix := range trans {
		if trans[ix] == langStr {
			return langStr
		}
	}
	log.Infof("Couldn't find translations for %s, using default", langStr)
	return LangEnUS
}

// T translates a string, possibly substituting arguments into it along
// the way. If len(args) is > 0, args1 is assumed to be the plural value
// and plural translation is used.
func T(defaultValue string, args ...int) string {
	if len(args) == 0 {
		return gettext.PGettext("", defaultValue)
	}
	return fmt.Sprintf(gettext.PNGettext("", defaultValue, defaultValue+pluralSuffix, args[0]),
		args[0])
}

// Errorf produces an error with a translated error string.
// Substitution is performed via the `T` function above, following
// the same rules.
func Errorf(defaultValue string, args ...int) error {
	return errors.New(T(defaultValue, args...))
}
