//
//	yaml/scanner/chars.go
//	goray
//
//	Created by Ross Light on 2010-06-28.
//

package scanner

func isBreak(c byte) bool {
	return c == '\n' || c == '\r'
}

func isWhitespace(c byte) bool {
	return c == ' ' || c == '\t'
}

func isDigit(c byte) bool {
	return c >= '0' && c <= '9'
}

func isLetter(c byte) bool {
	return (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z')
}

func isHexDigit(c byte) bool {
	return isDigit(c) || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')
}

func isWordChar(c byte) bool {
	return isLetter(c) || isDigit(c) || c == '_' || c == '-'
}

func asHex(c byte) (val int, ok bool) {
	ok = true
	switch {
	case c >= '0' && c <= '9':
		val = int(c - '0')
	case c >= 'A' && c <= 'F':
		val = int(c - 'A' + 0x09)
	case c >= 'a' && c <= 'f':
		val = int(c - 'a' + 0x09)
	default:
		ok = false
	}
	return
}

func asDecimal(c byte) (val int, ok bool) {
	if !isDigit(c) {
		return 0, false
	}
	return int(c - '0'), true
}
