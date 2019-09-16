#!/bin/bash
echo "updating..."
echo "writing some error" >&2

IN=$(cat)
echo "stdin: ${IN}" #the old state

#business logic
/bin/cat <<END >${filename}
  {"out1": "${out1}"}
END