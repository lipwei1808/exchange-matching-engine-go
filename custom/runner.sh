#!/bin/bash

for i in ./*.in; do
  last=$(../grader_arm64 ../build/engine < $i 2>&1 | tail -n -1) 
  echo $i test: $last
done