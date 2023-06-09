package httpserver

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"
	"reflect"
	"strconv"
	"strings"

	"github.com/cockroachdb/errors"
	"github.com/goccy/go-json"
	"github.com/golang/protobuf/proto"
	"github.com/spf13/cast"
	"github.com/tidwall/gjson"

	"github.com/milvus-io/milvus-proto/go-api/v2/commonpb"
	"github.com/milvus-io/milvus-proto/go-api/v2/milvuspb"
	"github.com/milvus-io/milvus-proto/go-api/v2/schemapb"

	"github.com/milvus-io/milvus/internal/log"
	"github.com/milvus-io/milvus/internal/util/funcutil"
)

func getPrimaryField(schema *schemapb.CollectionSchema) *schemapb.FieldSchema {
	primaryField := schema.Fields[0]
	for _, field := range schema.Fields {
		if field.IsPrimaryKey {
			primaryField = field
			break
		}
	}
	return primaryField
}

func getDim(field *schemapb.FieldSchema) (int64, error) {
	dimensionInSchema, err := funcutil.GetAttrByKeyFromRepeatedKV("dim", field.TypeParams)
	if err != nil {
		return 0, err
	}
	dim, err := strconv.Atoi(dimensionInSchema)
	if err != nil {
		return 0, err
	}
	return int64(dim), nil
}

func anyToColumns(rows []map[string]interface{}, sch *schemapb.CollectionSchema) ([]*schemapb.FieldData, error) {
	rowsLen := len(rows)
	if rowsLen == 0 {
		return []*schemapb.FieldData{}, errors.New("0 length column")
	}

	isDynamic := sch.EnableDynamicField
	var dim int64 = 0

	nameColumns := make(map[string]interface{})
	fieldData := make(map[string]*schemapb.FieldData)
	for _, field := range sch.Fields {
		// skip auto id pk field
		if field.IsPrimaryKey && field.AutoID {
			continue
		}
		var data interface{}
		switch field.DataType {
		case schemapb.DataType_Bool:
			data = make([]bool, 0, rowsLen)
		case schemapb.DataType_Int8:
			data = make([]int8, 0, rowsLen)
		case schemapb.DataType_Int16:
			data = make([]int16, 0, rowsLen)
		case schemapb.DataType_Int32:
			data = make([]int32, 0, rowsLen)
		case schemapb.DataType_Int64:
			data = make([]int64, 0, rowsLen)
		case schemapb.DataType_Float:
			data = make([]float32, 0, rowsLen)
		case schemapb.DataType_Double:
			data = make([]float64, 0, rowsLen)
		case schemapb.DataType_String:
			data = make([]string, 0, rowsLen)
		case schemapb.DataType_JSON:
			data = make([][]byte, 0, rowsLen)
		case schemapb.DataType_FloatVector:
			data = make([][]float32, 0, rowsLen)
			dim, _ = getDim(field)
		case schemapb.DataType_BinaryVector:
			data = make([][]byte, 0, rowsLen)
			dim, _ = getDim(field)
		}
		nameColumns[field.Name] = data
		fieldData[field.Name] = &schemapb.FieldData{
			Type:      field.DataType,
			FieldName: field.Name,
			FieldId:   field.FieldID,
			IsDynamic: field.IsDynamic,
		}
	}
	if dim == 0 {
		return nil, errors.New("cannot find dimension")
	}

	dynamicCol := make([][]byte, 0, rowsLen)

	for _, row := range rows {
		// collection schema name need not to be same, since receiver could has other names
		v := reflect.ValueOf(row)
		set, err := reflectValueCandi(v)
		if err != nil {
			return nil, err
		}

		for idx, field := range sch.Fields {
			// skip dynamic field if visible
			if isDynamic && field.IsDynamic {
				continue
			}
			// skip auto id pk field
			if field.IsPrimaryKey && field.AutoID {
				// remove pk field from candidates set, avoid adding it into dynamic column
				delete(set, field.Name)
				continue
			}

			candi, ok := set[field.Name]
			if !ok {
				return nil, fmt.Errorf("row %d does not has field %s", idx, field.Name)
			}
			switch field.DataType {
			case schemapb.DataType_Bool:
				nameColumns[field.Name] = append(nameColumns[field.Name].([]bool), candi.v.Interface().(bool))
			case schemapb.DataType_Int8:
				nameColumns[field.Name] = append(nameColumns[field.Name].([]int8), candi.v.Interface().(int8))
			case schemapb.DataType_Int16:
				nameColumns[field.Name] = append(nameColumns[field.Name].([]int16), candi.v.Interface().(int16))
			case schemapb.DataType_Int32:
				nameColumns[field.Name] = append(nameColumns[field.Name].([]int32), candi.v.Interface().(int32))
			case schemapb.DataType_Int64:
				nameColumns[field.Name] = append(nameColumns[field.Name].([]int64), candi.v.Interface().(int64))
			case schemapb.DataType_Float:
				nameColumns[field.Name] = append(nameColumns[field.Name].([]float32), candi.v.Interface().(float32))
			case schemapb.DataType_Double:
				nameColumns[field.Name] = append(nameColumns[field.Name].([]float64), candi.v.Interface().(float64))
			case schemapb.DataType_String:
				nameColumns[field.Name] = append(nameColumns[field.Name].([]string), candi.v.Interface().(string))
			case schemapb.DataType_JSON:
				nameColumns[field.Name] = append(nameColumns[field.Name].([][]byte), candi.v.Interface().([]byte))
			case schemapb.DataType_FloatVector:
				nameColumns[field.Name] = append(nameColumns[field.Name].([][]float32), candi.v.Interface().([]float32))
			case schemapb.DataType_BinaryVector:
				nameColumns[field.Name] = append(nameColumns[field.Name].([][]byte), candi.v.Interface().([]byte))
			}

			delete(set, field.Name)
		}

		if isDynamic {
			m := make(map[string]interface{})
			for name, candi := range set {
				m[name] = candi.v.Interface()
			}
			bs, err := json.Marshal(m)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal dynamic field %w", err)
			}
			dynamicCol = append(dynamicCol, bs)
			if err != nil {
				return nil, fmt.Errorf("failed to append value to dynamic field %w", err)
			}
		}
	}
	columns := make([]*schemapb.FieldData, 0, len(nameColumns))
	for name, column := range nameColumns {
		colData, _ := fieldData[name]
		switch colData.Type {
		case schemapb.DataType_Bool:
			colData.Field = &schemapb.FieldData_Scalars{
				Scalars: &schemapb.ScalarField{
					Data: &schemapb.ScalarField_BoolData{
						BoolData: &schemapb.BoolArray{
							Data: column.([]bool),
						},
					},
				},
			}
			break
		case schemapb.DataType_Int8:
			colData.Field = &schemapb.FieldData_Scalars{
				Scalars: &schemapb.ScalarField{
					Data: &schemapb.ScalarField_IntData{
						IntData: &schemapb.IntArray{
							Data: column.([]int32),
						},
					},
				},
			}
		case schemapb.DataType_Int16:
			colData.Field = &schemapb.FieldData_Scalars{
				Scalars: &schemapb.ScalarField{
					Data: &schemapb.ScalarField_IntData{
						IntData: &schemapb.IntArray{
							Data: column.([]int32),
						},
					},
				},
			}
		case schemapb.DataType_Int32:
			colData.Field = &schemapb.FieldData_Scalars{
				Scalars: &schemapb.ScalarField{
					Data: &schemapb.ScalarField_IntData{
						IntData: &schemapb.IntArray{
							Data: column.([]int32),
						},
					},
				},
			}
		case schemapb.DataType_Int64:
			colData.Field = &schemapb.FieldData_Scalars{
				Scalars: &schemapb.ScalarField{
					Data: &schemapb.ScalarField_LongData{
						LongData: &schemapb.LongArray{
							Data: column.([]int64),
						},
					},
				},
			}
		case schemapb.DataType_Float:
			colData.Field = &schemapb.FieldData_Scalars{
				Scalars: &schemapb.ScalarField{
					Data: &schemapb.ScalarField_FloatData{
						FloatData: &schemapb.FloatArray{
							Data: column.([]float32),
						},
					},
				},
			}
		case schemapb.DataType_Double:
			colData.Field = &schemapb.FieldData_Scalars{
				Scalars: &schemapb.ScalarField{
					Data: &schemapb.ScalarField_DoubleData{
						DoubleData: &schemapb.DoubleArray{
							Data: column.([]float64),
						},
					},
				},
			}
		case schemapb.DataType_String:
			colData.Field = &schemapb.FieldData_Scalars{
				Scalars: &schemapb.ScalarField{
					Data: &schemapb.ScalarField_StringData{
						StringData: &schemapb.StringArray{
							Data: column.([]string),
						},
					},
				},
			}
		case schemapb.DataType_JSON:
			colData.Field = &schemapb.FieldData_Scalars{
				Scalars: &schemapb.ScalarField{
					Data: &schemapb.ScalarField_BytesData{
						BytesData: &schemapb.BytesArray{
							Data: column.([][]byte),
						},
					},
				},
			}
		case schemapb.DataType_FloatVector:
			arr, err := convertFloatVectorToArray(column.([][]float32), dim)
			if err != nil {
				return nil, err
			}
			colData.Field = &schemapb.FieldData_Vectors{
				Vectors: &schemapb.VectorField{
					Dim: dim,
					Data: &schemapb.VectorField_FloatVector{
						FloatVector: &schemapb.FloatArray{
							Data: arr,
						},
					},
				},
			}
		case schemapb.DataType_BinaryVector:
			arr, err := convertBinaryVectorToArray(column.([][]byte), dim)
			if err != nil {
				return nil, err
			}
			colData.Field = &schemapb.FieldData_Vectors{
				Vectors: &schemapb.VectorField{
					Dim: dim,
					Data: &schemapb.VectorField_BinaryVector{
						BinaryVector: arr,
					},
				},
			}
		}
		columns = append(columns, colData)
	}
	if isDynamic {
		columns = append(columns, &schemapb.FieldData{
			Type:      0,
			FieldName: "",
			Field: &schemapb.FieldData_Scalars{
				Scalars: &schemapb.ScalarField{
					Data: &schemapb.ScalarField_JsonData{
						JsonData: &schemapb.JSONArray{
							Data: dynamicCol,
						},
					},
				},
			},
			FieldId:   0,
			IsDynamic: true,
		})
	}
	return columns, nil
}

func convertFloatVectorToArray(vector [][]float32, dim int64) ([]float32, error) {
	floatArray := make([]float32, 0)
	for _, arr := range vector {
		if int64(len(arr)) != dim {
			return nil, errors.New("wrong")
		}
		for i := int64(0); i < dim; i++ {
			floatArray = append(floatArray, arr[i])
		}
	}
	return floatArray, nil
}

func convertBinaryVectorToArray(vector [][]byte, dim int64) ([]byte, error) {
	binaryArray := make([]byte, 0)
	for _, arr := range vector {
		if int64(len(arr)) != dim {
			return nil, errors.New("wrong")
		}
		for i := int64(0); i < dim; i++ {
			binaryArray = append(binaryArray, arr[i])
		}
	}
	return binaryArray, nil
}

type fieldCandi struct {
	name    string
	v       reflect.Value
	options map[string]string
}

func reflectValueCandi(v reflect.Value) (map[string]fieldCandi, error) {
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	result := make(map[string]fieldCandi)
	switch v.Kind() {
	case reflect.Map: // map[string]interface{}
		iter := v.MapRange()
		for iter.Next() {
			key := iter.Key().String()
			result[key] = fieldCandi{
				name: key,
				v:    iter.Value(),
			}
		}
		return result, nil
	case reflect.Struct:
		for i := 0; i < v.NumField(); i++ {
			ft := v.Type().Field(i)
			name := ft.Name
			tag, ok := ft.Tag.Lookup(MilvusTag)

			settings := make(map[string]string)
			if ok {
				if tag == MilvusSkipTagValue {
					continue
				}
				settings = parseTagSetting(tag, MilvusTagSep)
				fn, has := settings[MilvusTagName]
				if has {
					// overwrite column to tag name
					name = fn
				}
			}
			_, ok = result[name]
			// duplicated
			if ok {
				return nil, fmt.Errorf("column has duplicated name: %s when parsing field: %s", name, ft.Name)
			}

			v := v.Field(i)
			if v.Kind() == reflect.Array {
				v = v.Slice(0, v.Len()-1)
			}

			result[name] = fieldCandi{
				name:    name,
				v:       v,
				options: settings,
			}
		}

		return result, nil
	default:
		return nil, fmt.Errorf("unsupport row type: %s", v.Kind().String())
	}
}

func parseTagSetting(str string, sep string) map[string]string {
	settings := map[string]string{}
	names := strings.Split(str, sep)

	for i := 0; i < len(names); i++ {
		j := i
		if len(names[j]) > 0 {
			for {
				if names[j][len(names[j])-1] == '\\' {
					i++
					names[j] = names[j][0:len(names[j])-1] + sep + names[i]
					names[i] = ""
				} else {
					break
				}
			}
		}

		values := strings.Split(names[j], ":")
		k := strings.TrimSpace(strings.ToUpper(values[0]))

		if len(values) >= 2 {
			settings[k] = strings.Join(values[1:], ":")
		} else if k != "" {
			settings[k] = k
		}
	}

	return settings
}

func vector2Placeholder(vectors []float32) *commonpb.PlaceholderValue {
	var placeHolderType commonpb.PlaceholderType
	ph := &commonpb.PlaceholderValue{
		Tag:    "$0",
		Values: make([][]byte, 0, len(vectors)),
	}
	if len(vectors) == 0 {
		return ph
	}
	placeHolderType = commonpb.PlaceholderType_FloatVector

	ph.Type = placeHolderType
	ph.Values = append(ph.Values, serialize(vectors))
	return ph
}

func serialize(fv []float32) []byte {
	data := make([]byte, 0, 4*len(fv)) // float32 occupies 4 bytes
	buf := make([]byte, 4)
	for _, f := range fv {
		binary.LittleEndian.PutUint32(buf, math.Float32bits(f))
		data = append(data, buf...)
	}
	return data
}

func vector2PlaceholderGroupBytes(floatArray []float32) []byte {
	phg := &commonpb.PlaceholderGroup{
		Placeholders: []*commonpb.PlaceholderValue{
			vector2Placeholder(floatArray),
		},
	}

	bs, _ := proto.Marshal(phg)
	return bs
}

func getDynamicOutputFields(schema *schemapb.CollectionSchema, outputFields []string) ([]string, []string) {
	fieldNames := []string{}
	for _, field := range schema.Fields {
		if field.IsPrimaryKey && !containsString(outputFields, field.Name) {
			outputFields = append(outputFields, field.Name)
		}
		fieldNames = append(fieldNames, field.Name)
	}
	if schema.EnableDynamicField {
		return outputFields, getNotExistData(outputFields, fieldNames)
	} else {
		return outputFields, []string{}
	}
}

func getNotExistData(arrayA []string, allArray []string) []string {
	// 使用一个 map 来记录数组 B 中的所有元素
	arrayBMap := make(map[string]bool)
	for _, value := range allArray {
		arrayBMap[value] = true
	}

	// 遍历数组 A，检查每个元素是否在该 map 中出现过
	resultArray := []string{}
	for _, value := range arrayA {
		if _, ok := arrayBMap[value]; !ok {
			resultArray = append(resultArray, value)
		}
	}

	return resultArray
}

func convertRange(field *schemapb.FieldSchema, result gjson.Result) (string, error) {
	var resultStr string
	fieldType := field.DataType

	if fieldType == schemapb.DataType_Int64 {
		dataArray := []int64{}
		for _, data := range result.Array() {
			if data.Type == gjson.String {
				value, err := cast.ToInt64E(data.Str)
				if err != nil {
					return "", err
				}
				dataArray = append(dataArray, value)
			} else {
				value, err := cast.ToInt64E(data.Raw)
				if err != nil {
					return "", err
				}
				dataArray = append(dataArray, value)
			}
		}
		resultStr = JoinArray(dataArray)
	} else if fieldType == schemapb.DataType_VarChar {
		dataArray := []string{}
		for _, data := range result.Array() {
			value, err := cast.ToStringE(data.Str)
			if err != nil {
				return "", err
			}
			dataArray = append(dataArray, value)
		}
		resultStr = JoinArray(dataArray)
	}
	return resultStr, nil
}

func JoinArray(data interface{}) string {
	var buffer bytes.Buffer
	arr := reflect.ValueOf(data)

	for i := 0; i < arr.Len(); i++ {
		if i > 0 {
			buffer.WriteString(",")
		}

		buffer.WriteString(fmt.Sprintf("%v", arr.Index(i)))
	}

	return buffer.String()
}

func checkGetPrimaryKey(coll *milvuspb.DescribeCollectionResponse, idResult gjson.Result) string {
	primaryField := getPrimaryField(coll.Schema)
	resultStr, _ := convertRange(primaryField, idResult)
	filter := primaryField.Name + " in [" + resultStr + "]"
	return filter
}

func containsString(arr []string, s string) bool {
	for _, str := range arr {
		if str == s {
			return true
		}
	}
	return false
}

func checkAndSetData(body string, collDescResp *milvuspb.DescribeCollectionResponse, req *InsertReq) error {
	reallyDataArray := []map[string]interface{}{}
	dataResult := gjson.Get(body, "data")
	dataResultArray := dataResult.Array()

	fieldNames := []string{}
	for _, field := range collDescResp.Schema.Fields {
		fieldNames = append(fieldNames, field.Name)
	}

	for _, data := range dataResultArray {
		reallyData := map[string]interface{}{}
		vectorArray := []float32{}
		if data.Type == gjson.JSON {
			for _, field := range collDescResp.Schema.Fields {
				fieldType := field.DataType
				fieldName := field.Name

				dataString := gjson.Get(data.Raw, fieldName).String()

				// 如果AutoId 则报错
				if field.IsPrimaryKey && collDescResp.Schema.AutoID {
					if dataString != "" {
						return errors.New(fmt.Sprintf("[checkAndSetData] fieldName %s AutoId already open, not support insert data %s", fieldName, dataString))
					} else {
						continue
					}
				}

				// 转换对应数据,同时检查数据合法性
				if fieldType == schemapb.DataType_FloatVector || fieldType == schemapb.DataType_BinaryVector {
					for _, vector := range gjson.Get(data.Raw, fieldName).Array() {
						vectorArray = append(vectorArray, cast.ToFloat32(vector.Num))
					}
					reallyData[fieldName] = vectorArray
				} else if fieldType == schemapb.DataType_Int8 {
					result, err := cast.ToInt8E(dataString)
					if err != nil {
						return errors.New(fmt.Sprintf("[checkAndSetData] dataString %s cast to int8 error: %s", dataString, err.Error()))
					}
					reallyData[fieldName] = result
				} else if fieldType == schemapb.DataType_Int16 {
					result, err := cast.ToInt16E(dataString)
					if err != nil {
						return errors.New(fmt.Sprintf("[checkAndSetData] dataString %s cast to int16 error: %s", dataString, err.Error()))
					}
					reallyData[fieldName] = result
				} else if fieldType == schemapb.DataType_Int32 {
					result, err := cast.ToInt32E(dataString)
					if err != nil {
						return errors.New(fmt.Sprintf("[checkAndSetData] dataString %s cast to int32 error: %s", dataString, err.Error()))
					}
					reallyData[fieldName] = result
				} else if fieldType == schemapb.DataType_Int64 {
					result, err := cast.ToInt64E(dataString)
					if err != nil {
						return errors.New(fmt.Sprintf("[checkAndSetData] dataString %s cast to int64 error: %s", dataString, err.Error()))
					}
					reallyData[fieldName] = result
				} else if fieldType == schemapb.DataType_VarChar || fieldType == schemapb.DataType_String {
					reallyData[fieldName] = dataString
				} else if fieldType == schemapb.DataType_Float {
					result, err := cast.ToFloat64E(dataString)
					if err != nil {
						return errors.New(fmt.Sprintf("[checkAndSetData] dataString %s cast to float64 error: %s", dataString, err.Error()))
					}
					reallyData[fieldName] = result
				} else if fieldType == schemapb.DataType_Bool {
					result, err := cast.ToBoolE(dataString)
					if err != nil {
						return errors.New(fmt.Sprintf("[checkAndSetData] dataString %s cast to bool error: %s", dataString, err.Error()))
					}
					reallyData[fieldName] = result
				} else {
					errMsg := fmt.Sprintf("[checkAndSetData] not support fieldName %s dataType %s", fieldName, fieldType)
					return errors.New(errMsg)
				}
			}

			// fill dynamic schema
			if collDescResp.Schema.EnableDynamicField {
				for mapKey, mapValue := range data.Map() {
					if !containsString(fieldNames, mapKey) {
						mapValueStr := mapValue.String()
						if mapValue.Type == gjson.True || mapValue.Type == gjson.False {
							reallyData[mapKey] = cast.ToBool(mapValueStr)
						} else if mapValue.Type == gjson.String {
							reallyData[mapKey] = mapValueStr
						} else if mapValue.Type == gjson.Number {
							if strings.Contains(mapValue.Raw, ".") {
								reallyData[mapKey] = cast.ToFloat64(mapValue.Raw)
							} else {
								reallyData[mapKey] = cast.ToInt64(mapValueStr)
							}
						} else if mapValue.Type == gjson.JSON {
							reallyData[mapKey] = mapValue.Value()
						} else {

						}
					}
				}
			}

			reallyDataArray = append(reallyDataArray, reallyData)
		} else {
			errMsg := fmt.Sprintf("[checkAndSetData] dataType %s not Json", data.Type)
			return errors.New(errMsg)
		}
	}
	req.Data = reallyDataArray
	return nil
}

func buildQueryResp(rowsNum int64, dynamicOutputFields []string, fieldDataList []*schemapb.FieldData, scores []float32) ([]map[string]interface{}, error) {
	if len(fieldDataList) == 0 {
		return nil, errors.New("no columns")
	}

	querysResp := []map[string]interface{}{}

	columnNum := len(fieldDataList)
	if rowsNum == int64(0) {
		dataType := fieldDataList[0].Type
		if dataType == schemapb.DataType_Bool {
			rowsNum = int64(len(fieldDataList[0].GetScalars().GetBoolData().Data))
		} else if dataType == schemapb.DataType_Int8 {
			rowsNum = int64(len(fieldDataList[0].GetScalars().GetIntData().Data))
		} else if dataType == schemapb.DataType_Int16 {
			rowsNum = int64(len(fieldDataList[0].GetScalars().GetIntData().Data))
		} else if dataType == schemapb.DataType_Int32 {
			rowsNum = int64(len(fieldDataList[0].GetScalars().GetIntData().Data))
		} else if dataType == schemapb.DataType_Int64 {
			rowsNum = int64(len(fieldDataList[0].GetScalars().GetLongData().Data))
		} else if dataType == schemapb.DataType_Float {
			rowsNum = int64(len(fieldDataList[0].GetScalars().GetFloatData().Data))
		} else if dataType == schemapb.DataType_Double {
			rowsNum = int64(len(fieldDataList[0].GetScalars().GetDoubleData().Data))
		} else if dataType == schemapb.DataType_VarChar {
			rowsNum = int64(len(fieldDataList[0].GetScalars().GetStringData().Data))
		} else if dataType == schemapb.DataType_BinaryVector {
			rowsNum = int64(len(fieldDataList[0].GetVectors().GetBinaryVector())*8) / fieldDataList[0].GetVectors().GetDim()
		} else if dataType == schemapb.DataType_FloatVector {
			rowsNum = int64(len(fieldDataList[0].GetVectors().GetFloatVector().Data)) / fieldDataList[0].GetVectors().GetDim()
		} else if dataType == schemapb.DataType_JSON {
			rowsNum = int64(len(fieldDataList[0].GetScalars().GetJsonData().Data))
		}
	}
	if scores != nil && rowsNum != int64(len(scores)) {
		return nil, errors.New("search result invalid")
	}
	for i := int64(0); i < rowsNum; i++ {
		row := map[string]interface{}{}
		for j := 0; j < columnNum; j++ {
			dataType := fieldDataList[j].Type
			if dataType == schemapb.DataType_Bool {
				row[fieldDataList[j].FieldName] = fieldDataList[j].GetScalars().GetBoolData().Data[i]
			} else if dataType == schemapb.DataType_Int8 {
				row[fieldDataList[j].FieldName] = fieldDataList[j].GetScalars().GetIntData().Data[i]
			} else if dataType == schemapb.DataType_Int16 {
				row[fieldDataList[j].FieldName] = fieldDataList[j].GetScalars().GetIntData().Data[i]
			} else if dataType == schemapb.DataType_Int32 {
				row[fieldDataList[j].FieldName] = fieldDataList[j].GetScalars().GetIntData().Data[i]
			} else if dataType == schemapb.DataType_Int64 {
				row[fieldDataList[j].FieldName] = fieldDataList[j].GetScalars().GetLongData().Data[i]
			} else if dataType == schemapb.DataType_Float {
				row[fieldDataList[j].FieldName] = fieldDataList[j].GetScalars().GetFloatData().Data[i]
			} else if dataType == schemapb.DataType_Double {
				row[fieldDataList[j].FieldName] = fieldDataList[j].GetScalars().GetDoubleData().Data[i]
			} else if dataType == schemapb.DataType_VarChar {
				row[fieldDataList[j].FieldName] = fieldDataList[j].GetScalars().GetStringData().Data[i]
			} else if dataType == schemapb.DataType_BinaryVector {
				row[fieldDataList[j].FieldName] = fieldDataList[j].GetVectors().GetBinaryVector()[i*(fieldDataList[j].GetVectors().GetDim()/8) : (i+1)*(fieldDataList[j].GetVectors().GetDim()/8)]
			} else if dataType == schemapb.DataType_FloatVector {
				row[fieldDataList[j].FieldName] = fieldDataList[j].GetVectors().GetFloatVector().Data[i*fieldDataList[j].GetVectors().GetDim() : (i+1)*fieldDataList[j].GetVectors().GetDim()]
			} else if dataType == schemapb.DataType_JSON {
				data, ok := fieldDataList[j].GetScalars().Data.(*schemapb.ScalarField_JsonData)
				if ok && !fieldDataList[j].IsDynamic {
					row[fieldDataList[j].FieldName] = string(data.JsonData.Data[i])
				} else {
					var dataMap map[string]interface{}

					err := json.Unmarshal(fieldDataList[j].GetScalars().GetJsonData().Data[i], &dataMap)
					if err != nil {
						msg := fmt.Sprintf("[BuildQueryResp] Unmarshal error %s", err.Error())
						log.Error(msg)
						return nil, err
					}

					if containsString(dynamicOutputFields, "*") || containsString(dynamicOutputFields, "$meta") {
						for key, value := range dataMap {
							row[key] = value
						}
					} else {
						for _, dynamicField := range dynamicOutputFields {
							if _, ok := dataMap[dynamicField]; ok {
								row[dynamicField] = dataMap[dynamicField]
							}
						}
					}

				}

			}
		}
		if scores != nil {
			row["distance"] = scores[i]
		}
		querysResp = append(querysResp, row)
	}

	return querysResp, nil
}
