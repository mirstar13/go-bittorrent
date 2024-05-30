package parser

import (
	"fmt"
	"sort"
	"unicode"
)

func Decode(bencodedString string) (interface{}, int, error) {
	switch {
	default:
		return "", 0, fmt.Errorf("format not supported")

	case unicode.IsDigit(rune(bencodedString[0])):
		return decodeString(bencodedString)

	case bencodedString[0] == 'i':
		return decodeInt(bencodedString)

	case bencodedString[0] == 'l':
		return decodeList(bencodedString)

	case bencodedString[0] == 'd':
		return decodeDictionary(bencodedString)

	}
}

func decodeString(str string) (interface{}, int, error) {
	size := 0
	colPos := 0

	for i := 0; i < len(str); i++ {
		if str[i] == ':' {
			colPos = i
			break
		}
	}

	_, err := fmt.Sscanf(str, "%d:", &size)
	if err != nil {
		return "", 0, err
	}

	return str[colPos+1 : colPos+1+size], size + len(fmt.Sprintf("%d:", size)), nil
}

func decodeInt(str string) (interface{}, int, error) {
	num := 0

	_, err := fmt.Sscanf(str, "i%de", &num)
	if err != nil {
		return 0, 0, err
	}

	return num, len(fmt.Sprintf("%d", num)) + 2, nil
}

func decodeDictionary(dict string) (map[string]interface{}, int, error) {
	res := map[string]interface{}{}
	i := 1

	for i < len(dict) && dict[i] != 'e' {
		key, bencodeLength, err := Decode(dict[i:])
		if err != nil {
			return nil, 0, err
		}

		i += bencodeLength

		value, bencodeLength, err := Decode(dict[i:])
		if err != nil {
			return nil, 0, err
		}

		res[key.(string)] = value

		i += bencodeLength
	}

	keys := make([]string, 0, len(res))

	for k := range res {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	sortedRes := map[string]interface{}{}

	for _, k := range keys {
		sortedRes[k] = res[k]
	}

	return sortedRes, i + 1, nil
}

func decodeList(li string) ([]interface{}, int, error) {
	res := []interface{}{}
	i := 1

	for i < len(li) && li[i] != 'e' {
		if li[i] == 'i' || li[i] == 'l' || unicode.IsDigit(rune(li[i])) {
			s, offset, _ := Decode(li[i:])

			i += offset
			res = append(res, s)
		} else {
			i += 1
		}
	}

	return res, i + 1, nil
}
