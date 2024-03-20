package message_hdl

import "strings"

const (
	singleLvlWildcard byte = '+'
	multiLvlWildcard  byte = '#'
	slash             byte = '/'
)

func parseTopic(topic, str string, tParts ...*string) bool {
	shift := 0
	strLen := len(str)
	pCount := 0
	var i int
	for i = 0; i < len(topic); i++ {
		if i+shift >= strLen {
			return false
		}
		switch topic[i] {
		case str[i+shift]:
		case singleLvlWildcard:
			pos := strings.IndexByte(str[i+shift:], slash)
			if pos < 0 {
				pos = len(str[i+shift:])
			}
			tp := str[i+shift : i+shift+pos]
			*tParts[pCount] = tp
			pCount++
			shift += len(tp) - 1
		case multiLvlWildcard:
			*tParts[pCount] = str[i+shift:]
			return true
		default:
			return false
		}
	}
	if i+shift < strLen {
		return false
	}
	return true
}
