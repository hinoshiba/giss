#!/bin/bash
if [ $# -ne 1 ]; then
  echo "$#"
  exit 1
fi

Version="${1}"
echo "creating ${Version}"

DIR="Release/${Version}"
README="${DIR}/release.md"
mkdir -p ${DIR}

echo "# ${Version}" > ${README}
echo "## changelog" >> ${README}
git log $(git describe --tags --abbrev=0)..HEAD --oneline >> ${README}
git tag "${Version}"
#git push --tag

export GOPATH="`pwd`"
ls -1 src/giss/exec | while read row ; do
  echo "compile ${row}"
  GOOS=linux GOARCH=amd64 go install -ldflags "-s -w -X giss/values.Version=${Version}" giss/exec/$row
  GOOS=windows GOARCH=amd64 go install -ldflags "-s -w -X giss/values.Version=${Version}" giss/exec/$row
  GOOS=darwin GOARCH=amd64 go install -ldflags "-s -w -X giss/values.Version=${Version}" giss/exec/$row
done

ls -1 bin/ | while read row ; do
  echo "archiving ${row}"
  zip -r "${DIR}/${row}.zip" "bin/${row}"
  tar cvfz "${DIR}/${row}.tar.gz" "bin/${row}"
done

echo "done"
exit 0
