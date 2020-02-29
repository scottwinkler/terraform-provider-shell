#!/bin/bash
echo "reading.."
echo "writing some error" >&2

IN=$(cat)
echo "stdin: ${IN}" #the old state

#business logic
cat ex.json >&3 #must write state to >&3