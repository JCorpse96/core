package copybook

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

type Format interface {
	getDecimals() int
	getType() string
	getElements() []Element
}

type Float struct {
	TypeOf   string `json:"typeOf"`
	Decimals int    `json:"decimals"`
	Format   `json:"-"`
}

func (f Float) getDecimals() int {
	return f.Decimals
}

func (f Float) getType() string {
	return f.TypeOf
}

type Integer struct {
	TypeOf string `json:"typeOf"`
	Format `json:"-"`
}

func (i Integer) getType() string {
	return i.TypeOf
}

type String struct {
	TypeOf string `json:"typeOf"`
	Format `json:"-"`
}

func (s String) getType() string {
	return s.TypeOf
}

type Array struct {
	TypeOf   string    `json:"typeOf"`
	Elements []Element `json:"elements"`
	Format   `json:"-"`
}

func (a Array) getType() string {
	return a.TypeOf
}

func (a Array) getElements() []Element {
	return a.Elements
}

type ParentElement struct {
	TypeOf     string  `json:"typeOf"`
	Subelement Element `json:"subelement"`
	Format     `json:"-"`
}

func (p ParentElement) getType() string {
	return p.TypeOf
}

type Element struct {
	ElementIndex     int    `json:"index"`
	ElementName      string `json:"name"`
	ElementMaxLength int    `json:"maxLength"`
	ElementFormat    Format `json:"format"`
}

type HXXX struct {
	Elements []Element `json:"elements"`
}

func MapJson(js string) (map[string]interface{}, error) {
	var data map[string]interface{}
	err := json.Unmarshal([]byte(js), &data)
	if err != nil {
		return nil, fmt.Errorf("Unable to coerce to object: %s", err.Error())
	}
	return data, nil
}

func (h HXXX) RenderCopybook(mapping map[string]string, request map[string]interface{}) string {
	return GetPaddedElements(h, mapping, request)
}

func GetPaddedElements(h HXXX, mapping map[string]string, request map[string]interface{}) string {
	final_elements := make(map[int]string)
	for _, element := range h.Elements {
		element_name := element.ElementName
		mapped := mapping[element_name]
		if element.ElementFormat.getType() == "array" {
			getRepeatedElements(element, mapped, mapping, request, final_elements)

		} else {
			index, padded := getPaddedElement(element, request, mapped)
			final_elements[index] = padded
		}
	}
	rendered_copy := join_elements(final_elements)
	return rendered_copy
}

func padElement(field string, element Element) string {
	var padded string
	if element.ElementFormat.getType() == "string" {
		padded = paddingRight(field, element.ElementMaxLength, "-")
	} else if element.ElementFormat.getType() == "number" {
		switch element.ElementFormat.(type) {
		case Float:
			field = strings.Replace(field, ".", "", 1)
		}
		padded = paddingLeft(field, element.ElementMaxLength, "0")
	}
	return padded
}

func getPaddedElement(element Element, request map[string]interface{}, mapped string) (int, string) {
	field := get_value(request, strings.Split(mapped, ".")).(string)

	return element.ElementIndex, padElement(field, element)
}

func getRepeatedElements(element Element, mapped string, mapping map[string]string, request map[string]interface{}, final_elements map[int]string) {
	repeated := get_value(request, strings.Split(mapped, ".")).([]interface{})
	elements := make(map[int]string)
	for i := range repeated {
		for k, sub_ele := range repeated[i].(map[string]interface{}) {
			mapped_key := mapped + "." + k
			var copy_key string
			for key, val := range mapping {
				if val == mapped_key {
					copy_key = key
					break
				}
			}
			var copy_element_match Element
			for _, cpy_element := range element.ElementFormat.getElements() {
				if cpy_element.ElementName == copy_key {
					copy_element_match = cpy_element
					break
				}
			}
			padded := padElement(sub_ele.(string), copy_element_match)
			index := i*len(repeated[i].(map[string]interface{})) + copy_element_match.ElementIndex
			elements[index] = padded
		}
	}
	if len(repeated) < element.ElementMaxLength {
		j := len(repeated)
		for j < element.ElementMaxLength {
			for _, cpy_element := range element.ElementFormat.getElements() {
				index := j*(len(element.ElementFormat.getElements())) + cpy_element.ElementIndex
				padded := padElement("", cpy_element)
				elements[index] = padded
			}
			j++
		}
	}
	joined := join_elements(elements)
	final_elements[element.ElementIndex] = joined
}

func join_elements(data map[int]string) string {
	concated_elements := ""
	var indexes []int
	for k := range data {
		indexes = append(indexes, k)
	}
	sort.Ints(indexes)
	for _, index := range indexes {
		concated_elements += data[index]
	}
	return concated_elements
}

func get_value(data map[string]interface{}, key []string) interface{} {
	var value interface{}
	if len(key) == 1 {
		value = data[key[0]]
	} else {
		newData, ok := data[key[0]].(map[string]interface{})
		if ok {
			value = get_value(newData, key[1:])
		}

	}
	return value
}

func paddingRight(value string, length int, padCharacter string) string {
	var padCountInt = 1 + ((length - len(padCharacter)) / len(padCharacter))
	var retStr = value + strings.Repeat(padCharacter, padCountInt)
	return retStr[:length]
}

func paddingLeft(value string, length int, padCharacter string) string {
	var padCountInt = 1 + ((length - len(padCharacter)) / len(padCharacter))
	var retStr = strings.Repeat(padCharacter, padCountInt) + value

	return retStr[(len(retStr) - length):]
}
