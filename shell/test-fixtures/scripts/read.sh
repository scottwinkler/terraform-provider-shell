#!/bin/bash
echo "reading.."
echo "writing some error" >&2

IN=$(cat)
echo "stdin: ${IN}" #the old state

#business logic
cat ${filename}