package util

// IsPossiblyNegative checks whether the integer could be a negative int32,
// unless its value is bigger than int32.max.
func IsPossiblyNegative(n int) bool {
    if n >> 32 == 0 {
        return n & (1 << 30) != 0
    }

    return n < 0
}
