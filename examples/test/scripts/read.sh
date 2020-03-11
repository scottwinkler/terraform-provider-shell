#!/bin/bash
echo "reading.."
echo "writing some error" >&2

IN=$(cat)
echo "stdin: ${IN}" #the old state

#business logic
cat ex.json # Last JSON object written to stdout is taken to be state
