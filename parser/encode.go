package parser

import "fmt"

func Encode(decodedItem interface{}) (string, error) {
	res := ""

	switch decodedItem := decodedItem.(type) {
	case string:
		res += encodeString(decodedItem)

	case int:
		res += encodeInt(decodedItem)

	case []interface{}:
		encodedList, err := encodeList(decodedItem)
		if err != nil {
			return "", err
		}

		res += encodedList
	case map[string]interface{}:
		encodedDictionary, err := encodeDictionary(decodedItem)
		if err != nil {
			return "", err
		}

		res += encodedDictionary
	}

	return res, nil
}

func encodeString(str string) string {
	return fmt.Sprintf("%d:%s", len(str), str)
}

func encodeInt(num int) string {
	return fmt.Sprintf("i%de", num)
}

func encodeDictionary(dict map[string]interface{}) (string, error) {
	res := "d"

	for key, item := range dict {
		encodedItem, err := Encode(item)
		if err != nil {
			return "", err
		}

		encodedKey := fmt.Sprintf("%d:%s", len(key), key)

		encodedDictionaryKeyValPair := encodedKey + encodedItem

		res += encodedDictionaryKeyValPair
	}

	res += "e"

	return res, nil
}

func encodeList(li []interface{}) (string, error) {
	res := "l"

	for _, item := range li {
		encodedItem, err := Encode(item)
		if err != nil {
			return "", err
		}

		res += encodedItem
	}

	res += "e"

	return res, nil
}
