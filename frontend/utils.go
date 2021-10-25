// Copyright 2020 ConsenSys AG
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package frontend

import (
	"math/big"
	"reflect"
)

type toBigIntInterface interface {
	ToBigIntRegular(res *big.Int) *big.Int
}

// FromInterface converts an interface to a big.Int element
// interface must implement ToBigIntRegular(res *big.Int) *big.Int
// (which is the case for field generated by goff)
// or be uint64, int, string, []byte or big.Int
// it panics if the input is invalid
func FromInterface(i1 interface{}) big.Int {
	var val big.Int

	switch c1 := i1.(type) {
	case big.Int:
		val.Set(&c1)
	case *big.Int:
		val.Set(c1)
	case uint64:
		val.SetUint64(c1)
	case uint:
		val.SetUint64(uint64(c1))
	case int:
		val.SetInt64(int64(c1))
	case string:
		if _, ok := val.SetString(c1, 10); !ok {
			panic("unable to set big.Int from base10 string")
		}
	case []byte:
		val.SetBytes(c1)
	default:
		if v, ok := i1.(toBigIntInterface); ok {
			v.ToBigIntRegular(&val)
			return val
		} else if reflect.ValueOf(i1).Kind() == reflect.Ptr {
			vv := reflect.ValueOf(i1).Elem()
			if vv.CanInterface() {
				if v, ok := vv.Interface().(toBigIntInterface); ok {
					v.ToBigIntRegular(&val)
					return val
				}
			}
		}
		panic(reflect.TypeOf(i1).String() + " to big.Int not supported")
	}

	return val
}
