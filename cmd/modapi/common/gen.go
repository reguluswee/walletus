package common

import (
	"fmt"
	"net/url"
	"reflect"
	"strconv"
	"time"
)

const TOKEN_DURATION = 10 * 24 * time.Hour

type HeaderParam struct {
	AppId     string
	AuthToken string
	Ts        string
	Ver       string
	RequestId string
	XAuth     string
}

func (h HeaderParam) Join() string {
	return fmt.Sprintf("%s%s%s%s", h.AppId, h.RequestId, h.Ts, h.Ver)
}

type QueryParams[T any] struct {
	Data T
}

func (q *QueryParams[T]) BuildQueryString() string {
	params := url.Values{}
	val := reflect.ValueOf(q.Data)
	typ := reflect.TypeOf(q.Data)

	for i := range val.NumField() {
		field := val.Field(i)
		fieldType := typ.Field(i)

		jsonTag := fieldType.Tag.Get("json")
		if jsonTag == "" || jsonTag == "-" {
			continue
		}
		jsonKey := jsonTag
		if commaIdx := len(jsonTag) - len(",omitempty"); commaIdx > 0 && jsonTag[commaIdx:] == ",omitempty" {
			jsonKey = jsonTag[:commaIdx]
		}

		switch field.Kind() {
		case reflect.String:
			if field.String() != "" {
				params.Set(jsonKey, field.String())
			}
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			if field.Int() != 0 {
				params.Set(jsonKey, strconv.FormatInt(field.Int(), 10))
			}
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			if field.Uint() != 0 {
				params.Set(jsonKey, strconv.FormatUint(field.Uint(), 10))
			}
		case reflect.Float32, reflect.Float64:
			if field.Float() != 0 {
				params.Set(jsonKey, strconv.FormatFloat(field.Float(), 'f', -1, 64))
			}
		case reflect.Ptr:
			if !field.IsNil() {
				elem := field.Elem()
				switch elem.Kind() {
				case reflect.String:
					params.Set(jsonKey, elem.String())
				case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
					params.Set(jsonKey, strconv.FormatInt(elem.Int(), 10))
				case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
					params.Set(jsonKey, strconv.FormatUint(elem.Uint(), 10))
				case reflect.Float32, reflect.Float64:
					params.Set(jsonKey, strconv.FormatFloat(elem.Float(), 'f', -1, 64))
				}
			}
		}
	}
	queryString := params.Encode()
	if queryString != "" {
		return "?" + queryString
	}
	return ""
}
