#/bin/bash

HEADER="// Copyright 2024 Francis Michael Stephens. All rights reserved.  Use of this
// source code is governed by an MIT license that can be found in the LICENSE
// file."

echo "$HEADER" > copyright-file

for file in $(find . -type f -name \*.go -not -path "./vendor/*"); do
	echo $file
	line=$(head -n 1 $file)
	if [[ $line != "// Copyright"* ]]; then
		echo "$HEADER" > copyright-file
		cat $file >> copyright-file
		mv copyright-file $file
	fi
done

rm copyright-file
