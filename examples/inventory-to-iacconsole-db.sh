#!/bin/bash

inventory_path=$1

post_to_iacconsoledb () {
  # $1 - filename

  dimpath=$1
    dimpath=${dimpath#*$inventory_path}
    dimpath=${dimpath%%.json}
    printf "\n-------------- $dimpath -------------\n"

  curl --header "Content-Type: application/json" \
  --request POST \
  -d @$1\
  "$IACCONSOLE_API_URL/v1/dimension/$dimpath?workspace=master&source=inventory&readonly=true" -q
  printf "\n------------------------\n\n"
}

for file in $(find $inventory_path -name '*.json')
do
   post_to_iacconsoledb $file
done
