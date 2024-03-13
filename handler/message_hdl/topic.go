package message_hdl

import "strings"

const (
	singleLvlWildcard byte = '+'
	multiLvlWildcard  byte = '#'
	slash             byte = '/'
)

func matchTopic(topic, str string, tParts []string) bool {
	shift := 0
	strLen := len(str)
	var i int
	for i = 0; i < len(topic); i++ {
		if i+shift >= strLen {
			return false
		}
		switch topic[i] {
		case singleLvlWildcard:
			pos := strings.IndexByte(str[i+shift:], slash)
			if pos < 0 {
				pos = len(str[i+shift:])
			}
			tPart := str[i+shift : i+shift+pos]
			shift += len(tPart) - 1
			tParts = append(tParts, tPart)
		case multiLvlWildcard:
			tParts = append(tParts, str[i+shift:])
			return true
		case str[i+shift]:
		default:
			return false
		}
	}
	if i+shift < strLen {
		return false
	}
	return true
}
