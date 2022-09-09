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

package pyexp

import (
	"fmt"
	"time"

	"github.com/raptor-ml/raptor/api"
	sTime "go.starlark.net/lib/time"
	"go.starlark.net/starlark"
)

func starToGo(val any) (any, error) {
	if val == nil {
		return nil, nil
	}

	switch v := val.(type) {
	case starlark.String:
		return string(v), nil
	case starlark.Int:
		return int(v.BigInt().Int64()), nil
	case starlark.Float:
		return float64(v), nil
	case *starlark.List:
		iter := v.Iterate()
		defer iter.Done()
		var elems []any
		var x starlark.Value
		for iter.Next(&x) {
			nv, err := starToGo(x)
			if err != nil {
				return nil, err
			}
			elems = append(elems, nv)
		}
		return api.NormalizeAny(elems)
	case sTime.Time:
		return time.Time(v), nil
	case starlark.NoneType:
		return nil, nil
	default:
		if v, ok := v.(fmt.Stringer); ok {
			return v.String(), nil
		}
		return nil, api.ErrUnsupportedPrimitiveError
	}
}

func headersToStarDict(h map[string][]string) starlark.Value {
	headers := starlark.NewDict(len(h))
	for k, v := range h {
		vls := make([]starlark.Value, 0, len(v))
		for _, vl := range v {
			vls = append(vls, starlark.String(vl))
		}
		_ = headers.SetKey(starlark.String(k), starlark.NewList(vls))
	}
	return headers
}
