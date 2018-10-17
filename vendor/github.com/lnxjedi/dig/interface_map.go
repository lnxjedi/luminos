/*
  Copyright (c) 2013 José Carlos Nieto, http://xiam.menteslibres.org/
  Copyright (c) 2018 David L. Parsley, parsley@linuxjedi.org

  Permission is hereby granted, free of charge, to any person obtaining
  a copy of this software and associated documentation files (the
  "Software"), to deal in the Software without restriction, including
  without limitation the rights to use, copy, modify, merge, publish,
  distribute, sublicense, and/or sell copies of the Software, and to
  permit persons to whom the Software is furnished to do so, subject to
  the following conditions:

  The above copyright notice and this permission notice shall be
  included in all copies or substantial portions of the Software.

  THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND,
  EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF
  MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND
  NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE
  LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION
  OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION
  WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
*/

package dig

import (
	"fmt"
	"strconv"
	"strings"
)

// InterfaceMap is a type for arbitrary data structures that can be used with
// dig. The top-level of the value is always a map[string]interface, but
// values of any type can be stored and retrieved.
type InterfaceMap map[string]interface{}

// New returns an empty InterfaceMap
func New() InterfaceMap {
	return map[string]interface{}{}
}

// Set sets an arbitrary value in a struct given a route follwed by the
// value, or returns an error.
func (im InterfaceMap) Set(params ...interface{}) error {

	l := len(params)

	if l < 2 {
		return fmt.Errorf("missing value")
	}

	route := params[0 : l-1]
	value := params[l-1]

	if err := Dig(&im, route...); err != nil {
		return err
	}
	return Set(&im, value, route...)
}

// Get returns an arbitrary interface value.
func (im InterfaceMap) Get(route ...interface{}) interface{} {
	var i interface{}

	Get(&im, &i, route...)
	return i
}

func makeRoute(r string) []interface{} {
	elements := strings.Split(r, "/")
	route := make([]interface{}, len(elements))
	for i, v := range elements {
		if idx, err := strconv.Atoi(v); err == nil {
			route[i] = idx
		} else {
			route[i] = v
		}
	}
	return route
}

// PathGet takes a string route of the form path/to/item, breaks it up into
// route elements, and returns the value from the route. Numbers are converted
// to ints for indexing slices.
func (im InterfaceMap) PathGet(route string) interface{} {
	return im.Get(makeRoute(route)...)
}

// PathSet takes a string route of the form path/to/item, breaks it up into
// route elements, and set the value for the route. Numbers are converted
// to ints for indexing slices.
func (im InterfaceMap) PathSet(route string, value interface{}) error {
	params := append(makeRoute(route), value)
	return im.Set(params...)
}
