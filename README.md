## cptest

Copy all example test cases from a problem statement into a file. Then test your code in one command!

Let's assume that you have a shortcut for compiling your code. Also your directory with code looks like this
```
$ ls
app Makefile main.cpp
```

Create a file, say, `inputs.txt` and copy the test cases in the following format:
```
input: test 1
---
output for test 1
===
input: test 2
---
output for test 2
===
input: test 3
---
output for test 3
```

That is, just separate input and output with `---` and test cases themselves with `===`.

Now, simply run in the directory
```
$ cptest
=== RUN  Test 1
=== RUN  Test 2
=== RUN  Test 3
[TODO]
```

You can also explicitly provide the working directory and/or the path to the inputs/executable. In general
```
Usage:
    cptest -i INPUTS -e EXECUTABLE [WORKING_DIR]
```


### Build

To build and run the application, do (assuming you've cloned the repository)
```
$ cd cptest
$ go build ./cmd/cptest
$ ./cptest
```
