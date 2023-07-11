#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

CMD_DIR=${CMD_DIR:-"pkg/cmd"}
CMD_FILES=("cmd.go" "apply/apply.go" "check/check.go" "compile/compile.go" "deps/deps.go" "destroy/destroy.go" \
"diff/diff.go" "env/env.go" "init/init.go" "ls/ls.go" "preview/preview.go" "version/version.go")
TRANSLATIONS_DIR=${TRANSLATIONS_BASE_DIR:-"pkg/util/i18n/translations/kusion"}
TRANSLATIONS_FOLDER=${TRANSLATIONS_FOLDER:-"LC_MESSAGES"}
TRANSLATIONS_FUNC=${TRANSLATIONS_FUNC:-"i18n.T"}
TRANSLATIONS_LANG=("en_US" "zh_CN")
POT_FILE=${POT_FILE:-"template.pot"}
PO_FILE=${PO_FILE:-"kusion.po"}

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

cmdFileList=""
for file in "${CMD_FILES[@]}"
do
  cmdFile="${CMD_DIR}/${file}"
  cmdFileList="${cmdFileList} ${cmdFile}"
done
cmdFileList=${cmdFileList:1}
tmpPotFilePath="${TRANSLATIONS_DIR}/tmp.pot"
go-xgettext -k "${TRANSLATIONS_FUNC}" -o "${tmpPotFilePath}" ${cmdFileList}
echo "derive .pot file from ${CMD_DIR} .go files succeeded."

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
done

potFilePath="${TRANSLATIONS_DIR}/${POT_FILE}"
mv "${tmpPotFilePath}" "${potFilePath}"
for lang in "${TRANSLATIONS_LANG[@]}"
do
  poFilePath="${TRANSLATIONS_DIR}/${lang}/${TRANSLATIONS_FOLDER}/${PO_FILE}"
  tmpPoFilePath="${TRANSLATIONS_DIR}/${lang}/${TRANSLATIONS_FOLDER}/tmp.po"
  mv "${tmpPoFilePath}" "${poFilePath}"
done
echo "update translations files succeeded."
