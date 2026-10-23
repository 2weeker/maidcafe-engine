package tripcode

import "strings"

const Mark = "◆"

const saltFrom = ":;<=>?@[\\]^_`"
const saltTo = "ABCDEFGabcdef"

func Classic(key string) string {
	if key == "" {
		return ""
	}
	k := []byte(key)
	padded := append(append([]byte{}, k...), 'H', '.')
	var salt [2]byte
	for i := 0; i < 2; i++ {
		c := padded[i+1]
		if c < '.' || c > 'z' {
			c = '.'
		} else if j := strings.IndexByte(saltFrom, c); j >= 0 {
			c = saltTo[j]
		}
		salt[i] = c
	}
	full := crypt(k, salt)
	return full[len(full)-10:]
}
