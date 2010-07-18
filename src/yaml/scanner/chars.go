//
//	yaml/scanner/chars.go
//	goray
//
//	Created by Ross Light on 2010-06-28.
//

package scanner

// isBreak returns whether a byte is a linebreak character.
func isBreak(c byte) bool {
	return c == '\n' || c == '\r'
}

// isWhitespace returns whether a byte is a space or a tab (but not a linebreak).
func isWhitespace(c byte) bool {
	return c == ' ' || c == '\t'
}

// isDigit returns whether a byte is an ASCII digit [0-9].
func isDigit(c byte) bool {
	return c >= '0' && c <= '9'
}

// isLetter returns whether a byte is an ASCII letter [A-Za-z]
func isLetter(c byte) bool {
	return (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z')
}

// isHexDigit returns whether a byte is a hexadecimal digit [0-9A-Fa-f].
func isHexDigit(c byte) bool {
	return isDigit(c) || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')
}

// isWordChar returns whether a byte is alphanumeric [-_0-9A-Za-z].
func isWordChar(c byte) bool {
	return isLetter(c) || isDigit(c) || c == '_' || c == '-'
}

// asHex returns the integer value of a single hexadecimal digit.
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

// asDecimal returns the integer value of a single decimal digit.
func asDecimal(c byte) (val int, ok bool) {
	if !isDigit(c) {
		return 0, false
	}
	return int(c - '0'), true
}
