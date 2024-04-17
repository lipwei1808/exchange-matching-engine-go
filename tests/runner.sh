#!/bin/bash
echo $1
if [ $1 == 0 ]
then
  echo hello
elif [ $1 == 1 ]
then
elif [ $1 == ""]
  echo hello2
fi

for i in ./*.in; do
  last=$(../grader ../build/engine < $i 2>&1 | tail -n -1) 
  echo $i test: $last
done
