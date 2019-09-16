#!/bin/bash
echo "deleting..."
echo "writing some error" >&2

IN=$(cat)
echo "stdin: ${IN}" #the old state

#business logic
rm -rf ${filename}