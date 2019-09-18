#!/bin/bash
echo "updating..."
echo "writing some error" >&2

IN=$(cat)
echo "stdin: ${IN}" #the old state

#business logic, create a large string and write a JSON stucture to file
TESTDATA=$(head -c $testdatasize < /dev/zero | tr '\0' '\141')

/bin/cat <<END >${filename}
  {"data":"${TESTDATA}", "out1": "${out1}"}
END
