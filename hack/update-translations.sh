#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

CMD_FILES=("pkg/cmd/cmd.go" "pkg/cmd/apply/apply.go" "pkg/cmd/compile/compile.go" "pkg/cmd/deps/deps.go" "pkg/cmd/destroy/destroy.go" \
"pkg/cmd/init/init.go" "pkg/cmd/preview/preview.go" "pkg/cmd/version/version.go" "pkg/engine/backend/backend.go")
TRANSLATIONS_DIR=${TRANSLATIONS_BASE_DIR:-"pkg/util/i18n/translations/kusion"}
TRANSLATIONS_FOLDER=${TRANSLATIONS_FOLDER:-"LC_MESSAGES"}
TRANSLATIONS_FUNC=${TRANSLATIONS_FUNC:-"i18n.T"}
TRANSLATIONS_LANG=("en_US" "zh_CN")
DEFAULT_TRANSLATIONS_LANG=${DEFAULT_TRANSLATIONS_LANG:-"en_US"}
POT_FILE=${POT_FILE:-"template.pot"}
PO_FILE=${PO_FILE:-"kusion.po"}
MO_FILE=${MO_FILE:-"kusion.mo"}

if ! which go-xgettext > /dev/null; then
  echo "Can not find go-xgettext, install with: go install github.com/gosexy/gettext/go-xgettext."
  exit 1
fi

if ! which msginit > /dev/null; then
  echo "Can not find msginit, install with: brew install gettext."
  exit 1
fi

if ! which msgmerge > /dev/null; then
  echo "Can not find msgmerge, install with: brew install gettext."
  exit 1
fi

if ! which msgfmt > /dev/null; then
  echo "Can not find msgfmt, install with: brew install gettext."
  exit 1
fi

cmdFileList=""
for file in "${CMD_FILES[@]}"
do
  cmdFileList="${cmdFileList} ${file}"
done
cmdFileList=${cmdFileList:1}
tmpPotFilePath="${TRANSLATIONS_DIR}/tmp.pot"
go-xgettext -k "${TRANSLATIONS_FUNC}" -o "${tmpPotFilePath}" ${cmdFileList}
echo "derive .pot file succeeded."

for lang in "${TRANSLATIONS_LANG[@]}"
do
  poFilePath="${TRANSLATIONS_DIR}/${lang}/${TRANSLATIONS_FOLDER}/${PO_FILE}"
  tmpPoFilePath="${TRANSLATIONS_DIR}/${lang}/${TRANSLATIONS_FOLDER}/tmp.po"
  if [ -e "${poFilePath}" ]; then
    echo "start updating .po file of lang ${lang}."
    msgmerge "${poFilePath}" "${tmpPotFilePath}" -o "${tmpPoFilePath}" > /dev/null # actually, there is no use of "> /dev/null", cause msgmerge 0.21.1 will output err even if succeeded.
    echo "update .po file of lang ${lang} succeeded."
  else
    echo "start initializing .po file of lang ${lang}."
    msginit -l "${lang}" -o "${tmpPoFilePath}" -i "${tmpPotFilePath}" --no-translator > /dev/null
    echo "initialize .po file of lang ${lang} succeeded."
  fi
  if [ "${lang}" = "${DEFAULT_TRANSLATIONS_LANG}" ]; then
    echo "start updating .mo file of default lang ${lang}"
    tmpMoFilePath="${TRANSLATIONS_DIR}/${lang}/${TRANSLATIONS_FOLDER}/tmp.mo"
    msgfmt -c -v -o "${tmpMoFilePath}" "${tmpPoFilePath}" > /dev/null
    echo "update .mo file of default lang ${lang} succeeded"
  fi
done

potFilePath="${TRANSLATIONS_DIR}/${POT_FILE}"
mv "${tmpPotFilePath}" "${potFilePath}"
for lang in "${TRANSLATIONS_LANG[@]}"
do
  poFilePath="${TRANSLATIONS_DIR}/${lang}/${TRANSLATIONS_FOLDER}/${PO_FILE}"
  tmpPoFilePath="${TRANSLATIONS_DIR}/${lang}/${TRANSLATIONS_FOLDER}/tmp.po"
  mv "${tmpPoFilePath}" "${poFilePath}"
  if [ "${lang}" = "${DEFAULT_TRANSLATIONS_LANG}" ]; then
    moFilePath="${TRANSLATIONS_DIR}/${lang}/${TRANSLATIONS_FOLDER}/${MO_FILE}"
    tmpMoFilePath="${TRANSLATIONS_DIR}/${lang}/${TRANSLATIONS_FOLDER}/tmp.mo"
    mv "${tmpMoFilePath}" "${moFilePath}"
  fi
done
echo "update translations files succeeded."
