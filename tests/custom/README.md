## Test Generator script

## Build

1. Run the following command to build

```
g++ -std=c++20 test_generator.cpp -o test_generator
```

2. The executable accepts 3 command line arguments.

- The 1st arg represents the number of test cases.
- The 2nd arg represents the number of instruments to generate.
- The 3rd arg represents the number of commands in each test case.

An example run to produce 10 test cases with 30 commands over 2 instruments is as follows.

```
./test_generator 10 2 30
```
