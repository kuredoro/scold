## cptest
[![Coverage Status](https://coveralls.io/repos/github/kuredoro/cptest/badge.svg?branch=master)](https://coveralls.io/github/kuredoro/cptest?branch=master) [![Actions Status](https://github.com/kuredoro/cptest/workflows/build/badge.svg)](https://github.com/kuredoro/cptest/actions)

Copy all example test cases from a problem statement into a file. Then test your code in one command!

### Hmm?

Let's assume that you have a shortcut for compiling your code. Also the directory with your code looks like this
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
```

That is, just separate input and output with `---` and test cases themselves with `===`.

Now, simply run in the directory
```
$ cptest
found executable: /home/username/path/to/app
using default time limit: 6s
=== RUN Test 1
=== RUN Test 2
=== RUN Test 3
--- WA: Test 1 (0.005s)
Input:
input: test 1

Answer:
output for test 1

Output:
no

--- WA: Test 2 (0.006s)
Input:
input: test 2

Answer:
output for test 2

Output:
no
```

(Your code was determining if the given string was a palindrome, by the way.)

You can also explicitly provide the working directory and/or the path to the inputs/executable. In general
```
Usage:
    cptest -i INPUTS -e EXECUTABLE [WORKING_DIR]
```

You can simply amend the command you're executing with the shortcut and that's it! No more risking getting a WA on the first test or wasting time rechecking the given examples!

#### Time limit

You can also specify a time limit for the tests. Instead of the first test, you can provide a set of key-value pairs. Only one key is supported for now, though -- `tl`. (If you have any ideas for more, an issue is welcome!)

Assuming, the task is to compute fibonacci numbers, somebody wrote a O(2^n) implementation. This `inputs.txt` will fail:
```
tl = 10.0
===
1
---
1
===
2
---
1
===
47
---
2971215073

```

Running `cptest`, we get
```
$ cptest
found executable: /home/username/path/to/app
=== RUN	Test 1
=== RUN	Test 2
=== RUN	Test 3
--- OK:	Test 1 (0.006s)
--- OK:	Test 2 (0.006s)
--- TL:	Test 3 (10.000s)
Input:
47

Answer:
2971215073

FAIL
2/3 passed
```

### Build

To build and run the application, do (assuming you've cloned the repository)
```
$ cd cptest
$ go build ./cmd/cptest
$ ./cptest
```

You can add it to your PATH to call it from anywhere in the console. On Linux that could be `/usr/bin/` directory.
