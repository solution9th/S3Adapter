package to

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/private/protocol"
	"github.com/aws/aws-sdk-go/private/protocol/xml/xmlutil"
)

var errValueNotSet = fmt.Errorf("value not set")

// UnmarshalRequest unmarshals the REST request to struct
func UnmarshalRequest(ctx context.Context, r *http.Request, v interface{}) (string, error) {

	v1 := reflect.Indirect(reflect.ValueOf(v))
	return unmarshal(ctx, r, v1)
}

func unmarshal(ctx context.Context, r *http.Request, v reflect.Value) (string, error) {

	for i := 0; i < v.NumField(); i++ {
		var err error
		m, field := v.Field(i), v.Type().Field(i)

		if n := field.Name; n[0:1] == strings.ToLower(n[0:1]) {
			continue
		}
		if !m.IsValid() {
			continue
		}
		name := field.Tag.Get("locationName")
		if name == "" {
			name = field.Name
		}

		switch field.Tag.Get("location") {
		case "header":
			err = unmarshalHeader(ctx, m, r.Header.Get(name), field.Tag)
			if err != nil {
				return field.Name, err
			}
		case "headers":
			prefix := field.Tag.Get("locationName")
			err = unmarshalHeaderMap(ctx, m, r.Header, prefix)
			if err != nil {
				return field.Name, err
			}
		case "querystring":
			err = unmarshalQuery(ctx, m, r.URL.Query().Get(name), field.Tag)
			if err != nil {
				return field.Name, err
			}
		}
	}
	return "", unmarshalBody(ctx, r, v)
}

// MarshalResponse marshal struct to REST response
func MarshalResponse(ctx context.Context, w http.ResponseWriter, v interface{}) error {

	v1 := reflect.Indirect(reflect.ValueOf(v))
	return marshal(ctx, w, v1)
}

func marshal(ctx context.Context, w http.ResponseWriter, v reflect.Value) error {

	if !v.CanAddr() {
		return nil
	}

	for i := 0; i < v.NumField(); i++ {
		m, field := v.Field(i), v.Type().Field(i)
		if n := field.Name; n[0:1] == strings.ToLower(n[0:1]) {
			continue
		}
		if (m.Kind() != reflect.Struct && m.IsNil()) || !m.IsValid() {
			continue
		}

		name := field.Tag.Get("locationName")
		if name == "" {
			name = field.Name
		}

		switch field.Tag.Get("location") {
		case "header":
			err := marshalHeader(ctx, w.Header(), m, name, field.Tag)
			if err != nil {
				return err
			}
		case "headers":
			err := marshalHeaderMap(ctx, m, w.Header(), field.Tag)
			if err != nil {
				return err
			}
		}
	}

	return marshalBody(ctx, w, v)
}

func marshalBody(ctx context.Context, w http.ResponseWriter, v reflect.Value) error {
	field, ok := v.Type().FieldByName("_")
	if !ok {
		return nil
	}

	payloadName := field.Tag.Get("payload")
	if payloadName == "" {
		return nil
	}
	pfield, _ := v.Type().FieldByName(payloadName)
	ptag := pfield.Tag.Get("type")

	payload := reflect.Indirect(v.FieldByName(payloadName))
	if !payload.IsValid() || payload.Interface() == nil {
		return nil
	}

	if ptag == "" || ptag == "structure" {
		var buf bytes.Buffer
		buf.WriteString(xml.Header)
		err := xmlutil.BuildXML(payload.Interface(), xml.NewEncoder(&buf))
		if err != nil {
			return err
		}
		w.Write(buf.Bytes())
		return nil
	}

	switch reader := payload.Interface().(type) {
	case io.ReadSeeker:
		data, err := ioutil.ReadAll(reader)
		if err != nil {
			return err
		}
		w.Write(data)
	case io.ReadCloser:
		data, err := ioutil.ReadAll(reader)
		if err != nil {
			return err
		}
		w.Write(data)
	case []byte:
		w.Write(reader)
	case string:
		w.Write([]byte(reader))
	default:
		return fmt.Errorf("unknown payload type %s", payload.Type())
	}
	return nil
}

func marshalHeader(ctx context.Context, h http.Header, v reflect.Value, headerName string, tag reflect.StructTag) error {

	str, err := convertType(v, tag)
	if err == errValueNotSet {
		return nil
	} else if err != nil {
		return err
	}

	headerName = strings.TrimSpace(headerName)
	str = strings.TrimSpace(str)
	h.Add(headerName, str)
	return nil
}

func marshalHeaderMap(ctx context.Context, v reflect.Value, header http.Header, tag reflect.StructTag) error {

	prefix := tag.Get("locationName")
	for _, key := range v.MapKeys() {
		str, err := convertType(v.MapIndex(key), tag)
		if err == errValueNotSet {
			continue
		} else if err != nil {
			return err
		}
		keyStr := strings.TrimSpace(key.String())
		str = strings.TrimSpace(str)

		header.Add(prefix+keyStr, str)
	}

	return nil
}

func unmarshalHeaderMap(ctx context.Context, v reflect.Value, headers http.Header, prefix string) error {
	switch v.Interface().(type) {
	case map[string]*string:
		out := map[string]*string{}
		for k, v := range headers {
			k = http.CanonicalHeaderKey(k)
			if strings.HasPrefix(strings.ToLower(k), strings.ToLower(prefix)) {
				out[k[len(prefix):]] = &v[0]
			}
		}
		v.Set(reflect.ValueOf(out))
	}
	return nil
}

// when has => _ struct{} `type:"structure" payload:"LegalHold"`
func unmarshalBody(ctx context.Context, r *http.Request, v reflect.Value) error {
	field, ok := v.Type().FieldByName("_")
	if !ok {
		return nil
	}
	payloadName := field.Tag.Get("payload")
	if payloadName == "" {
		return nil
	}
	pfield, _ := v.Type().FieldByName(payloadName)
	ptag := pfield.Tag.Get("type")
	// 嵌入 struct 暂不解决
	if ptag == "" || ptag == "structure" {
		return nil
	}
	payload := v.FieldByName(payloadName)
	if !payload.IsValid() {
		return nil
	}
	switch payload.Interface().(type) {
	case []byte:
		defer r.Body.Close()
		b, err := ioutil.ReadAll(r.Body)
		if err != nil {
			return err
		}
		payload.Set(reflect.ValueOf(b))
	case *string:
		defer r.Body.Close()
		b, err := ioutil.ReadAll(r.Body)
		if err != nil {
			return err
		}
		str := string(b)
		payload.Set(reflect.ValueOf(&str))
	default:
		switch payload.Type().String() {
		case "io.ReadCloser":
			payload.Set(reflect.ValueOf(r.Body))
		case "io.ReadSeeker":
			b, err := ioutil.ReadAll(r.Body)
			if err != nil {
				return err
			}
			payload.Set(reflect.ValueOf(bytes.NewReader(b)))
			// payload.Set(reflect.ValueOf(ioutil.NopCloser(bytes.NewReader(b))))
		default:
			io.Copy(ioutil.Discard, r.Body)
			defer r.Body.Close()
			err := fmt.Errorf("unknown payload type %s", payload.Type())
			return err
		}
	}

	return nil
}

func unmarshalHeader(ctx context.Context, v reflect.Value, header string, tag reflect.StructTag) error {
	if !v.IsValid() || (header == "" && v.Elem().Kind() != reflect.String) {
		return nil
	}
	return setValue(ctx, v, header, tag)
}

func unmarshalQuery(ctx context.Context, v reflect.Value, params string, tag reflect.StructTag) error {
	if params == "" {
		return nil
	}

	return setValue(ctx, v, params, tag)
}

func setValue(ctx context.Context, v reflect.Value, param string, tag reflect.StructTag) error {

	switch v.Interface().(type) {
	case *string:
		v.Set(reflect.ValueOf(&param))
	case []byte:
		b, err := base64.StdEncoding.DecodeString(param)
		if err != nil {
			return err
		}
		v.Set(reflect.ValueOf(&b))
	case *bool:
		b, err := strconv.ParseBool(param)
		if err != nil {
			return err
		}
		v.Set(reflect.ValueOf(&b))
	case *int64:
		i, err := strconv.ParseInt(param, 10, 64)
		if err != nil {
			return err
		}
		v.Set(reflect.ValueOf(&i))
	case *float64:
		f, err := strconv.ParseFloat(param, 64)
		if err != nil {
			return err
		}
		v.Set(reflect.ValueOf(&f))
	case *time.Time:
		format := tag.Get("timestampFormat")
		if len(format) == 0 {
			format = protocol.RFC822TimeFormatName
		}
		t, err := protocol.ParseTime(format, param)
		if err != nil {
			return err
		}
		v.Set(reflect.ValueOf(&t))
	case aws.JSONValue:
		escaping := protocol.NoEscape
		if tag.Get("location") == "header" {
			escaping = protocol.Base64Escape
		}
		m, err := protocol.DecodeJSONValue(param, escaping)
		if err != nil {
			return err
		}
		v.Set(reflect.ValueOf(m))
	default:
		err := fmt.Errorf("Unsupported value for param %v (%s)", v.Interface(), v.Type())
		return err
	}
	return nil
}

func convertType(v reflect.Value, tag reflect.StructTag) (str string, err error) {
	v = reflect.Indirect(v)
	if !v.IsValid() {
		return "", errValueNotSet
	}

	switch value := v.Interface().(type) {
	case string:
		str = value
	case []byte:
		str = base64.StdEncoding.EncodeToString(value)
	case bool:
		str = strconv.FormatBool(value)
	case int64:
		str = strconv.FormatInt(value, 10)
	case float64:
		str = strconv.FormatFloat(value, 'f', -1, 64)
	case time.Time:
		format := tag.Get("timestampFormat")
		if len(format) == 0 {
			format = protocol.RFC822TimeFormatName
			if tag.Get("location") == "querystring" {
				format = protocol.ISO8601TimeFormatName
			}
		}
		str = protocol.FormatTime(format, value)
	case aws.JSONValue:
		if len(value) == 0 {
			return "", errValueNotSet
		}
		escaping := protocol.NoEscape
		if tag.Get("location") == "header" {
			escaping = protocol.Base64Escape
		}
		str, err = protocol.EncodeJSONValue(value, escaping)
		if err != nil {
			return "", fmt.Errorf("unable to encode JSONValue, %v", err)
		}
	default:
		err := fmt.Errorf("unsupported value for param %v (%s)", v.Interface(), v.Type())
		return "", err
	}
	return str, nil
}
