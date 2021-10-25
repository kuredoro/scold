## scold
[![Coverage Status](https://coveralls.io/repos/github/kuredoro/scold/badge.svg?branch=main)](https://coveralls.io/github/kuredoro/scold?branch=main)
[![GoReport](https://goreportcard.com/badge/github.com/kuredoro/scold)](https://goreportcard.com/report/github.com/kuredoro/scold)
[![PkgGoDev](https://pkg.go.dev/badge/github.com/kuredoro/scold)](https://pkg.go.dev/github.com/kuredoro/scold)
[![Actions Status](https://github.com/kuredoro/scold/workflows/build/badge.svg)](https://github.com/kuredoro/scold/actions)
[![Release](https://img.shields.io/github/release/kuredoro/scold.svg?style=flat-square)](https://github.com/kuredoro/scold/releases/latest)

A tool to speed up the testing of competitive programming code on multiple inputs. I.e., the purpose is to minimize the number of keypresses between finishing writing code and submitting it, knowing that the code passes on the provided test inputs.

* [Install](#install)
  * [Windows](#windows)
  * [Linux](#linux)
* [User guide](#user-guide)
  * [Command line interface](#command-line-interface)
  * [inputs\.txt format](#inputstxt-format)
  * [How are outputs compared?](#how-are-outputs-compared)
  * [Verdicts](#verdicts)
    * [OK: Test pass](#ok-test-pass)
    * [WA: Wrong answer](#wa-wrong-answer)
    * [RE: Runtime error](#re-runtime-error)
    * [TL: Time limit exceeded](#tl-time-limit-exceeded)
    * [IE: Internal error](#ie-internal-error)
  * [Test suite configuration](#test-suite-configuration)
    * [Specifying time limit](#specifying-time-limit)
    * [Specifying floating point precision](#specifying-floating-point-precision)
* [Building](#building)

## Install

### Windows

On Windows go to the releases page on the right and download the scold executable for your architecture. Rename it to `scold.exe`. Additionally, you can create a folder, add it to the PATH and put `scold.exe` in there to access scold from the console.

You can do all of this automatically by using `scoop`. scoop is a minimalistic install helper for Windows that allows you to install, update, and manage various command-line utilities and applications (python, go, node.js) from within the console in just one line. Get it from [scoop.sh](https://scoop.sh). If you're a fan of Linux, you will be pleased that you can install all of the Linux tools via scoop and enjoy your Linux habits on Windows. Also, if you need some app, just `scoop search` to see if you can have it without a hassle.

When you installed scoop, run
```
> scoop update
```

And then add the bucket with the install script and install scold.
```
> scoop bucket add kuredoro https://github.com/kuredoro/scoop-bucket
> scoop install scold
```

### Linux

You can use AUR on ArchLinux to install scold. Via `yay` AUR helper it would be:
```
$ yay -S scold
```

Don't forget to star packages on AUR if you liked them (─‿‿─).

For other distros, you'll have to build it from the source, but don't worry, it's just one line. See [Building](#building).

## User guide

### Command-line interface

```
scold [options] EXECUTABLE [ARGS...]
```

scold requires an executable to run. Any arguments written after the executable are forwarded to it. This way, one can call `scold node index` to test a Node.js code. The options related to the scold are, therefore, specified before the executable.

Possible arguments:

* `-i`, `--inputs` -- specifies the path to the test suite. Default: `inputs.txt`.
* `-j`, `--jobs` -- specifies the number of executables to run concurrently. Default: CPU count.
* `--no-colors` -- disables colored output. Useful for environments that cannot render color, like Sublime Text console.
* `--no-progress` -- disables progress bar. Since the progress bar requires some specific features of a console emulator, it might not work everywhere.

### `inputs.txt` format

The format is simple:
```
===     (1)
line 1  ┐
line 2  │
---     ├─ Test 1
line 1  │
line 2  ┘
===     (1)
===
line 1  ┐
line 2  │
---     ├─ Test 2
line 1  │
line 2  ┘
===     (1)
```

In other words, your inputs are separated from outputs with `---`, and test cases are separated with `===`. The empty test cases (1) are ignored and allowed. All of the lines in the input and output sections will be newline terminated when parsed by scold.

The first test can specify a set of **test suite options** and follows a different format
```
tl = 10s
prec = 8
===
line 1
line 2
---
line 1
line 2
```

See [Test suite configuration](#test-suite-configuration).

### How are outputs compared?

**TL;DR:** The comparison routine is more or less equivalent to a program that reads the program's output from the `stdin` and compares it against correct values.

The key concepts in scold are the **lexeme** and the **lexeme type**. Lexeme is a string consisting of printable characters or a single newline. The program's output and the test's answer are parsed and turned into a sequence of lexemes while discarding all whitespaces, tabs, etc. between them.

Given this sequence, the **lexeme type** is deduced for each lexeme. The lexeme type specifies what type of data the lexeme holds. Currently, there are string, integer and floating-point number types. There is a specialization, or an "is-a", relation between the types. For example, every integer is a string, but not every string is an integer, hence integer specializes string type. In fact, specialization relation is the weak order relation: `>=`, and it follows that: string `>=` floating-point number `>=` integer.

For each lexeme at the same position in both sequences (program's output and the answer) their **common type** is deduced by taking the least specialized type among the two lexemes. Then a comparison routine is invoked that performs highlighting of the mismatched parts depending on the deduced common type. For example, if a discrepancy is found in an integer, the whole integer should be highlighted, instead of individual characters, otherwise, it would be annoying and not logically correct.

Additional measures are taken to treat excessive newlines rationally. If a misplaced newline is encountered (meaning that the other lexeme is not a newline), the lexemes after this newline are skipped until a non-newline lexeme is encountered. This ensures that in case of an excessive newline, the comparison highlighting stays consistent and valid.

### Verdicts

#### `OK`: Test pass

Example:
```
--- OK: Test 1 (0.123s)
```

`OK` verdict reflects the answer and the program's output being conceptually equal and that the program finished within the specified time frame.

Even if the program has some `stderr` output, it won't be displayed.

#### `WA`: Wrong answer

Example:
```
--- WA:	Test 1 (0.523s)
Input:
5
5 4 3 2 1

Answer:
5
1 2 3 4 5\n

Output:
5
1 2 4 3 5

Stderr:
Some debug information...
```

`WA` verdict signalizes that the program's output somehow differs from the correct answer so much that the output cannot be any longer considered correct.

The discrepancies are highlighted by default and the behavior of the highlight is different depending on the nature of data the inconsistency is found in. If the lexemes that are different are

* floating-point numbers -- the sign, the whole part or/and the fractional part up until the specified precision is highlighted.
* integer numbers -- the sign or/and all of the digits are highlighted.
* strings -- the individual characters are highlighted.

Moreover, if a number is too long, it's treated as a string.

The newlines are not visible, but if one is missing or misplaced, it will be highlighted and rendered as text, in this example, as `\n`. Since a misplaced newline can skew all of the lexemes that follow it, the extra newlines are ignored as if they have never been in the input while comparing further lexemes.

Next, to aid debugging, the `stderr` is also being printed (if it has anything). This way, a printf-style debugging can not interfere with the output that needs to be compared. Further adjustments to the user's code can create a special logger that will output to `stderr` and that can be easily turned off before sending the code to the online judge.

Some OJ add `ONLINE_JUDGE` preprocessor macro when compiling C/C++ code, like Codeforces or UVa. You can use `#ifdef ONLINE_JUDGE` to conditionally turn off debugging when sending to the OJ. Or do this in reverse: define a macro that will be present when compiling locally, like `g++ main.cpp -DLOCAL_BUILD` and use it instead. More ideas and for more languages can be found in this [Codeforces thread](https://codeforces.cc/blog/entry/14118).


#### `RE`: Runtime error

Example:
```
--- RE:	Test 1 (0.005s)
Input:
1

Answer:
1\n

Exit code: 255

Output:
Something from stdout

Stderr:
Something from stderr
```

Runtime errors are detected by looking at the exit code. If it has a non-zero value -- it's a `RE`. When this happens, additional information is always printed: the exit code and the `stderr`.

#### `TL`: Time limit exceeded

Example:
```
--- TL:	Test 1 (6.000s)
Input:
abcde

Answer:
edcba\n
```

When the program exceeds the default time limit or the one specified by the user (see [Specifying time limit](#specifying-time-limit)), the program is terminated and the test is failed. Nor the output, nor the `stderr` are shown.

#### `IE`: Internal error

Example (on Linux):
```
--- IE:	Test 160 (0.169s)
Input:
55

Answer:
55\n

Error:
executable: fork/exec /home/kuredoro/contest_code/a.out: too many open files
```

The internal error is a failed test because scold could not perform what it was designed to do. The situations when IE pops out are extremely rare but sometimes can occur. In the example above, the problem is that the executable `a.out` was opened too many times simultaneously exceeding the limit Linux allows an executable to be opened at the same time (on the machine in question). The IE can also appear when scold panics itself, in which case it might be a potential bug. As always, read what the error says and, if anything, ask for help or file a bug on the [issue tracker](https://github.com/kuredoro/scold/issues).

### Test suite configuration

A set of key-value pairs can be specified at the very top of `inputs.txt`. For example:
```
tl = 10s
prec = 8
===
input 1
---
output 1
===
input 2
---
output 2
```

A key-value pair is a line with an equality sign. The key and the value are located to the left and to the right of the sign, respectively. They both are space-trimmed. So, `"  two words =   are  parsed"` is parsed as: `key="two words"` and `value="are  parsed"`.

#### Specifying time limit

Syntax:
```
tl = <digits> [ '.' <digits> ] <unit>
unit ::=  "ns" | "us" | "µs" | "ms" | "s" | "m" | "h"
```

Examples:
```
tl = 1ms
tl = 6.66s
tl = 0.01m
tl = 0s
```

The `tl` option specifies the time limit for the test suite, overriding the default value. The value for the time limit should contain a unit suffix and may contain a fractional part. "us" and "µs" both correspond to microseconds. If `tl` is specified to be `0`, then the time limit is considered to be infinite.

#### Specifying floating point precision

Syntax:
```
prec = <digits>
```

Examples:
```
prec = 0
prec = 12
```

The `prec` option specifies how many digits after the decimal point should be considered when comparing floating-point lexemes. The value of 0 tells scold to ignore the fractional part.

## Building

To build `scold` you'll need an installation of `go`. Installing it should be as simple as installing base-devel package (─‿‿─).

E.g. on Arch:
```
$ sudo pacman -S go
```

Then clone the repository, and invoke `go build`.
```
$ git clone https://github.com/kuredoro/scold.git
$ cd scold
$ go build ./cmd/scold
```

Finally, move the executable to a folder accessible from PATH, like
```
$ sudo mv scold /usr/bin
```

You're ready to go!
