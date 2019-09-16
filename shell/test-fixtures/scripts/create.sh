#!/bin/bash
echo "creating..."
echo "writing some error" >&2

IN=$(cat)
echo "stdin: ${IN}" #the old state, not useful for create step since the old state was empty

#business logic
/bin/cat <<END >${filename}
  {"out1": "${out1}"}
END

cat ${filename} >&3