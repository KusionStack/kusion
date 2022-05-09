# Translations README

This is a basic sketch of the workflow needed to add translations:

# Adding/Updating Translations

## New languages
Create `pkg/util/i18n/translations/kusion/<language>/LC_MESSAGES/kusion.po`. There's
no need to update `translations/test/...` which is only used for unit tests.

Once you've added a new language, you'll need to register it in
`pkg/util/i18n/i18n.go` by adding it to the `knownTranslations` map.

## Extracting strings
Once the strings are wrapped, you can extract strings from go files using
the `go-xgettext` command which can be installed with:

```console
go get github.com/gosexy/gettext/go-xgettext
```

## Adding new translations
Edit the appropriate `kusion.po` file, `poedit` is a popular open source tool
for translations. You can load the `pkg/util/i18n/translations/kusion/template.pot` file
to find messages that might be missing.

Once you are done with your `kusion.po` file, generate the corresponding `kusion.mo`
file. `poedit` does this automatically on save.

We use the English translation as the `msgid`.

# Using translations

To use translations, you simply need to add:
```go
import pkg/i18n
...
// Get a translated string
translated := i18n.T("Your message in english here")

// Get a translated plural string
translated := i18n.T("You had % items", items)

// Translated error
return i18n.Error("Something bad happened")

// Translated plural error
return i18n.Error("%d bad things happened")
```
