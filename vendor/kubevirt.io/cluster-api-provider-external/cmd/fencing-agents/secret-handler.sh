#!/bin/bash

extra_args=""
env | grep SECRETPATH_ | sed s/SECRETPATH_// > /tmp/secrets

if [ "x$ARG_FORMAT" = xcli ]; then
    while IFS= read -r line; do
	field=$(echo $line | awk -F= '{print $1}')
	secretpath=$(echo $line | awk -F= '{print $2}')
	extra_args="$extra_args --$field $(cat $secretpath)"
    done < /tmp/secrets
else
    while IFS= read -r line; do
	field=$(echo $line | awk -F= '{print $1}')
	secretpath=$(echo $line | awk -F= '{print $2}')
	secretvalue=$(cat $secretpath)
	export field="$secretpath"
    done < /tmp/secrets
fi

echo "Fixed command: $* $extra_args"
$* $extra_args
