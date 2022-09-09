/*
Copyright (c) 2022 RaptorML authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package sdk

import (
	"github.com/raptor-ml/raptor/api"
	coreApi "go.buf.build/raptor/api-go/raptor/core/raptor/core/v1alpha1"
)

func FromAPIPrimitive(p coreApi.Primitive) api.PrimitiveType {
	switch p {
	default:
		return api.PrimitiveTypeUnknown
	case coreApi.Primitive_PRIMITIVE_STRING:
		return api.PrimitiveTypeString
	case coreApi.Primitive_PRIMITIVE_INTEGER:
		return api.PrimitiveTypeInteger
	case coreApi.Primitive_PRIMITIVE_FLOAT:
		return api.PrimitiveTypeFloat
	case coreApi.Primitive_PRIMITIVE_TIMESTAMP:
		return api.PrimitiveTypeTimestamp
	case coreApi.Primitive_PRIMITIVE_STRING_LIST:
		return api.PrimitiveTypeStringList
	case coreApi.Primitive_PRIMITIVE_INTEGER_LIST:
		return api.PrimitiveTypeIntegerList
	case coreApi.Primitive_PRIMITIVE_FLOAT_LIST:
		return api.PrimitiveTypeFloatList
	case coreApi.Primitive_PRIMITIVE_TIMESTAMP_LIST:
		return api.PrimitiveTypeTimestampList
	}
}

func FromAPIAggrFunc(f coreApi.AggrFn) api.WindowFn {
	switch f {
	default:
		return api.WindowFnUnknown
	case coreApi.AggrFn_AGGR_FN_SUM:
		return api.WindowFnSum
	case coreApi.AggrFn_AGGR_FN_AVG:
		return api.WindowFnAvg
	case coreApi.AggrFn_AGGR_FN_MIN:
		return api.WindowFnMin
	case coreApi.AggrFn_AGGR_FN_MAX:
		return api.WindowFnMax
	case coreApi.AggrFn_AGGR_FN_COUNT:
		return api.WindowFnCount
	}
}

func FromAPIAggrFuncs(fs []coreApi.AggrFn) []api.WindowFn {
	wfs := make([]api.WindowFn, 0, len(fs))
	for _, f := range fs {
		wfs = append(wfs, FromAPIAggrFunc(f))
	}
	return wfs
}

func FromAPIMetadata(m *coreApi.Metadata) api.Metadata {
	return api.Metadata{
		FQN:           m.Fqn,
		Primitive:     FromAPIPrimitive(m.Primitive),
		Aggr:          FromAPIAggrFuncs(m.Aggr),
		Freshness:     m.Freshness.AsDuration(),
		Staleness:     m.Staleness.AsDuration(),
		Timeout:       m.Timeout.AsDuration(),
		Builder:       m.Builder,
		DataConnector: m.DataConnector,
	}
}

func fromScalar(scalar *coreApi.Scalar) any {
	if scalar == nil {
		return nil
	}

	switch scalar.Value.(type) {
	case *coreApi.Scalar_StringValue:
		return scalar.GetStringValue()
	case *coreApi.Scalar_FloatValue:
		return scalar.GetFloatValue()
	case *coreApi.Scalar_IntValue:
		return int(scalar.GetIntValue())
	case *coreApi.Scalar_TimestampValue:
		return scalar.GetTimestampValue().AsTime()
	}

	panic("unknown scalar type")
}

func FromValue(val *coreApi.Value) any {
	if val == nil {
		return nil
	}

	switch v := val.Value.(type) {
	case *coreApi.Value_ScalarValue:
		return fromScalar(v.ScalarValue)
	case *coreApi.Value_ListValue:
		list := v.ListValue
		ret := make([]any, len(list.Values))
		for i, v := range list.Values {
			ret[i] = fromScalar(v)
		}
		return ret
	}

	panic("unknown value type")
}
