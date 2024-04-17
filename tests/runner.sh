#!/bin/bash

if [[ $1 -eq "0" ]] || [[ -z $1 ]]
then
  for i in ./basic/*.in; do
    last=$(../grader ../build/engine < $i 2>&1 | tail -n -1) 
    echo $i test: $last
  done
fi

if [[ $1 -eq "1" ]] || [[ -z $1 ]]
then
  for i in ./custom/*.in; do
    last=$(../grader ../build/engine < $i 2>&1 | tail -n -1) 
    echo $i test: $last
  done
fi

