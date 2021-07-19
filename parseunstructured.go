package fragment

import (
	"errors"
	"fmt"
	"strconv"
	"unicode"

	"github.com/ludvigalden/go-typemeta"
)

// ParseUnstructured parses a value and turns it into a fragment. It can be a `fragment.Unstructured`, which wil simply be returned,
// a list of strings, a string-interface map, or a fragment string such as "fullName, profile { createdAt }".
func ParseUnstructured(v interface{}) (Unstructured, error) {
	// else if vf, ok := v.(ValueFragment); ok {
	// 	return vf.EnsureFragment().Fragment().Unstructured(), nil
	// }
	if v == nil {
		return Unstructured{}, nil
	} else if fv, ok := v.(Unstructured); ok {
		return fv, nil
	} else if fv, ok := v.(Struct); ok {
		return fv.ToUnstructured(), nil
	} else if expr, ok := v.(string); ok {
		return parseString(trimFragmentChars([]rune(expr)))
	} else if fields, ok := v.([]string); ok {
		return NewUnstructured().Add(fields...), nil
	} else if fieldsMap, ok := v.(map[string]interface{}); ok {
		fragment := NewUnstructured()
		for fieldName, fieldv := range fieldsMap {
			fieldFragment, err := ParseUnstructured(fieldv)
			if err != nil {
				return fragment, NewError(err).Register(fieldName)
			}
			fragment = fragment.Set(fieldName, fieldFragment)
		}
		return fragment, nil
	}
	return Unstructured{}, errors.New("cannot parse fragment value with type " + typemeta.Get(v).String())
}

func parseString(fragmentChars []rune) (Unstructured, error) {
	fragmentCharsLen := len(fragmentChars)
	emptyFragment := false
	if fragmentCharsLen > 1 && fragmentChars[0] == leftBrace && fragmentChars[fragmentCharsLen-1] == rightBrace {
		if fragmentCharsLen > 2 {
			fragmentChars = fragmentChars[1 : fragmentCharsLen-1]
		} else {
			fragmentChars = []rune{}
			emptyFragment = true
		}
		fragmentCharsLen = len(fragmentChars)
	}
	if emptyFragment {
		return NewEmptyUnstructured(), nil
	}
	fragment := NewUnstructured()
	if fragmentCharsLen == 0 {
		return fragment, nil
	}
	fieldParts, err := splitFragmentByComma(fragmentChars, fragmentCharsLen)
	if err != nil {
		return fragment, err
	}
	for _, fieldPartChars := range fieldParts {
		fieldPartChars = trimFragmentChars(fieldPartChars)
		fieldPartSpaceParts, err := splitFragment(fieldPartChars, len(fieldPartChars))
		fieldPartSpacePartsLen := len(fieldPartSpaceParts)
		if err != nil {
			return fragment, errors.New("failed parsing \"" + string(fieldPartChars) + "\": " + err.Error())
		} else if fieldPartSpacePartsLen > 2 {
			return fragment, errors.New("failed parsing \"" + string(fieldPartChars) + "\": contains more than one space")
		} else {
			fieldName := trimFragmentChars(fieldPartSpaceParts[0])
			if fieldPartSpacePartsLen == 1 {
				fragment = fragment.Add(string(fieldName))
			} else if fieldPartSpacePartsLen == 2 {
				fieldFragment, err := parseString(fieldPartSpaceParts[1])
				if err != nil {
					return fragment, errors.New("\"" + string(fieldName) + "\": " + err.Error())
				} else if fieldFragment.fields != nil {
					fragment = fragment.Set(string(fieldName), fieldFragment)
				} else {
					fragment = fragment.Add(string(fieldName))
				}
			} else {
				return fragment, errors.New("unexpected amount of parts " + fmt.Sprint(fieldPartSpaceParts))
			}
		}
	}
	return fragment, nil
}

func splitFragmentByComma(chars []rune, charsLen int) ([][]rune, error) {
	parts := [][]rune{}
	var partStart int
	var braceDepth int
	maxIndex := charsLen - 1
	for i := 0; i < charsLen; i++ {
		char := chars[i]
		if char == leftBrace {
			braceDepth++
		} else if char == rightBrace {
			if braceDepth == 0 {
				return parts, errors.New("unexpected closing paranthesis at index " + strconv.Itoa(i) + " in fragment \"" + string(chars) + "\"")
			}
			braceDepth--
		}
		if i == maxIndex {
			partChars := chars[partStart:]
			if len(partChars) != 0 {
				parts = append(parts, partChars)
			}
		} else if braceDepth == 0 && char == comma {
			partChars := chars[partStart:i]
			partStart = i + 1
			if len(partChars) != 0 {
				parts = append(parts, partChars)
			}
		}
	}
	if braceDepth != 0 {
		return parts, errors.New("missing closing brace in \"" + string(chars) + "\"")
	}
	return parts, nil
}

func splitFragment(chars []rune, charsLen int) ([][]rune, error) {
	parts := [][]rune{}
	var partStart int
	var braceDepth int
	for i := 0; i < charsLen; i++ {
		char := chars[i]
		if char == leftBrace {
			if braceDepth == 0 {
				partChars := chars[partStart:i]
				partStart = i
				if len(partChars) != 0 {
					parts = append(parts, partChars)
				}
			}
			braceDepth++
		} else if char == rightBrace {
			if braceDepth == 0 {
				return parts, errors.New("unexpected closing brace at index " + strconv.Itoa(i) + " in \"" + string(chars) + "\"")
			}
			braceDepth--
			if braceDepth == 0 {
				partChars := chars[partStart : i+1]
				partStart = i + 1
				if len(partChars) != 0 {
					parts = append(parts, partChars)
				}
			}
		} else if i == charsLen-1 {
			partChars := chars[partStart:]
			if len(partChars) != 0 {
				parts = append(parts, partChars)
			}
		}
	}
	if braceDepth != 0 {
		return parts, errors.New("missing closing brace in \"" + string(chars) + "\"")
	}
	return parts, nil
}

const (
	leftBrace  = rune('{')
	rightBrace = rune('}')
	comma      = rune(',')
)

func trimFragmentChars(chars []rune) []rune {
	charsLen := len(chars)
	if charsLen == 0 {
		return chars
	}
	left := -1
	for i := 0; i < charsLen; i++ {
		if unicode.IsSpace(chars[i]) {
			continue
		}
		left = i
		break
	}
	right := -1
	for i := charsLen - 1; i >= left; i-- {
		if unicode.IsSpace(chars[i]) {
			continue
		}
		right = i
		break
	}
	return chars[left : right+1]
}
