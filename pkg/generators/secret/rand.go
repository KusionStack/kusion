package secret

import (
	"math/rand"
	"sync"
	"time"
	"unsafe"
)

// Since rand.NewSource() doesn't provide safety under concurrent use,
// we need to use sync.Mutex here.
var rng = struct {
	sync.Mutex
	rand *rand.Rand
}{
	rand: rand.New(rand.NewSource(time.Now().UnixNano())),
}

const (
	// We omit vowels from the set of available characters to reduce the chances
	// of "bad words" being formed.
	alphanums = "bcdfghjklmnpqrstvwxzBCDFGHJKLMNPQRSTVWXZ2456789"
	// No. of bits required to index into alphanums string.
	alphanumsIdxBits = 5
	// Mask used to extract last alphanumsIdxBits of an int.
	alphanumsIdxMask = 1<<alphanumsIdxBits - 1
	// No. of random letters we can extract from a single int63.
	maxAlphanumsPerInt = 63 / alphanumsIdxBits
)

// GenerateRandomString generates a random alphanumeric string, without vowels and which is
// n characters long.
func GenerateRandomString(n int) string {
	b := make([]byte, n)
	rng.Lock()
	defer rng.Unlock()

	// A src.Int63() generates 63 random bits, enough for letterIdxMax characters!
	for i, cache, remain := n-1, rng.rand.Int63(), maxAlphanumsPerInt; i >= 0; {
		if remain == 0 {
			cache, remain = rng.rand.Int63(), maxAlphanumsPerInt
		}
		if idx := int(cache & alphanumsIdxMask); idx < len(alphanums) {
			b[i] = alphanums[idx]
			i--
		}
		cache >>= alphanumsIdxBits
		remain--
	}

	return *(*string)(unsafe.Pointer(&b))
}
