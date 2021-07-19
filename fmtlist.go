package fragment

import (
	"fmt"
	"reflect"
)

// fmtListAnd formats a list like "item1, item2, and item3"
func fmtListAnd(languageTag string, items ...interface{}) string {
	return fmtList(items, languageTag, false)
}

// fmtListOr formats a list like "item1, item2, or item3"
func fmtListOr(languageTag string, items ...interface{}) string {
	return fmtList(items, languageTag, true)
}

func fmtList(v []interface{}, languageTag string, or bool) string {
	items := stringSlice(v...)
	itemsLen := len(items)
	if itemsLen == 0 {
		return ""
	} else if itemsLen == 1 {
		return items[0]
	}
	fmtSpec, ok := listFmtSpecs[languageTag]
	if !ok {
		fmtSpec = listFmtSpecs["en"]
	}
	sep := fmtSpec.andSep
	lastComma := fmtSpec.andLastComma
	if or {
		sep = fmtSpec.orSep
		lastComma = fmtSpec.orLastComma
	}
	if itemsLen == 2 {
		return items[0] + " " + sep + " " + items[1]
	}
	str := items[0]
	commaItems := items[1 : itemsLen-1]
	for _, commaItem := range commaItems {
		str += ", " + commaItem
	}
	if lastComma {
		str += ","
	}
	str += " " + sep + " " + items[itemsLen-1]
	return str
}

// stringSlice returns a string slice from the specified items
func stringSlice(items ...interface{}) []string {
	strItems := []string{}
	for _, item := range items {
		if itemstr, ok := item.(string); ok {
			strItems = append(strItems, itemstr)
		} else if itemstrs, ok := item.([]string); ok {
			strItems = append(strItems, itemstrs...)
		} else {
			rv := reflect.ValueOf(item)
			for rv.Kind() == reflect.Ptr {
				if rv.IsNil() {
					break
				}
				rv = rv.Elem()
			}
			if rv.IsZero() {
				continue
			}
			if kind := rv.Kind(); kind == reflect.Slice || kind == reflect.Array {
				for i := 0; i < rv.Len(); i++ {
					iv := rv.Index(i)
					for iv.Kind() == reflect.Ptr {
						if iv.IsNil() {
							break
						}
						iv = iv.Elem()
					}
					if iv.IsZero() {
						continue
					}
					if iv.Kind() == reflect.String {
						strItems = append(strItems, iv.String())
					} else {
						strItems = append(strItems, fmt.Sprint(iv.Interface()))
					}
				}
			} else {
				strItems = append(strItems, fmt.Sprint(rv.Interface()))
			}
		}
	}
	return strItems
}

// quotedItems returns an array of quoted string from a list of items
func quotedItems(items ...interface{}) []string {
	strItems := stringSlice(items...)
	quoted := []string{}
	for _, str := range strItems {
		if str[0:1] != "\"" {
			str = "\"" + str + "\""
		}
		quoted = append(quoted, str)
	}
	return quoted
}

// quotedStringInteraces returns an array of interfaces from an array of strings, which also will be quoted
func quotedStringInteraces(strs ...string) []interface{} {
	interfaces := []interface{}{}
	for _, str := range strs {
		if str[0:1] != "\"" {
			str = "\"" + str + "\""
		}
		interfaces = append(interfaces, str)
	}
	return interfaces
}

var listFmtSpecs = map[string]listFmtSpec{
	"en":    {"and", true, "or", true},
	"sv-SE": {"och", false, "eller", false},
	"da-DK": {"og", false, "eller", false},
	"no-NO": {"og", false, "eller", false},
	"fi-FI": {"ja", false, "tai", false},
	"fr-FR": {"et", false, "ou", false},
	"de-DE": {"und", false, "oder", false},
}

type listFmtSpec struct {
	andSep       string
	andLastComma bool
	orSep        string
	orLastComma  bool
}
