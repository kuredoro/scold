package cptest

// StringError is an error type whose values can be constant and compared against deterministically with == operator. An error type that solves the problems of sentinel errors.
type StringError string

func (e StringError) Error() string {
	return string(e)
}

