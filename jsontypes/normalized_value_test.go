// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package jsontypes_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

func TestNormalizedStringSemanticEquals(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		currentJson   jsontypes.Normalized
		givenJson     basetypes.StringValuable
		expectedMatch bool
		expectedDiags diag.Diagnostics
	}{
		"not equal - mismatched field values": {
			currentJson:   jsontypes.NewNormalizedValue(`{"hello": "dlrow", "nums": [3, 2, 1], "nested": {"test-bool": false}}`),
			givenJson:     jsontypes.NewNormalizedValue(`{"hello": "world", "nums": [1, 2, 3], "nested": {"test-bool": true}}`),
			expectedMatch: false,
		},
		"not equal - mismatched field names": {
			currentJson:   jsontypes.NewNormalizedValue(`{"Hello": "world", "Nums": [1, 2, 3], "Nested": {"Test-bool": true}}`),
			givenJson:     jsontypes.NewNormalizedValue(`{"hello": "world", "nums": [1, 2, 3], "nested": {"test-bool": true}}`),
			expectedMatch: false,
		},
		"not equal - object additional field": {
			currentJson:   jsontypes.NewNormalizedValue(`{"hello": "world", "nums": [1, 2, 3], "nested": {"test-bool": true}, "new-field": null}`),
			givenJson:     jsontypes.NewNormalizedValue(`{"hello": "world", "nums": [1, 2, 3], "nested": {"test-bool": true}}`),
			expectedMatch: false,
		},
		"not equal - array additional field": {
			currentJson:   jsontypes.NewNormalizedValue(`[{"hello": "world"}, {"nums":[1, 2, 3]}, {"nested": {"test-bool": true, "new-field": null}}]`),
			givenJson:     jsontypes.NewNormalizedValue(`[{"hello": "world"}, {"nums":[1, 2, 3]}, {"nested": {"test-bool": true}}]`),
			expectedMatch: false,
		},
		"not equal - array item order difference": {
			currentJson:   jsontypes.NewNormalizedValue(`[{"nums":[1, 2, 3]}, {"hello": "world"}, {"nested": {"test-bool": true}}]`),
			givenJson:     jsontypes.NewNormalizedValue(`[{"hello": "world"}, {"nums":[1, 2, 3]}, {"nested": {"test-bool": true}}]`),
			expectedMatch: false,
		},
		"semantically equal - object byte-for-byte match": {
			currentJson:   jsontypes.NewNormalizedValue(`{"hello": "world", "nums": [1, 2, 3], "nested": {"test-bool": true}}`),
			givenJson:     jsontypes.NewNormalizedValue(`{"hello": "world", "nums": [1, 2, 3], "nested": {"test-bool": true}}`),
			expectedMatch: true,
		},
		"semantically equal - array byte-for-byte match": {
			currentJson:   jsontypes.NewNormalizedValue(`[{"hello": "world"}, {"nums":[1, 2, 3]}, {"nested": {"test-bool": true}}]`),
			givenJson:     jsontypes.NewNormalizedValue(`[{"hello": "world"}, {"nums":[1, 2, 3]}, {"nested": {"test-bool": true}}]`),
			expectedMatch: true,
		},
		"semantically equal - object field order difference": {
			currentJson:   jsontypes.NewNormalizedValue(`{"hello": "world", "nums": [1, 2, 3], "nested": {"test-bool": true}}`),
			givenJson:     jsontypes.NewNormalizedValue(`{"nums": [1, 2, 3], "nested": {"test-bool": true}, "hello": "world"}`),
			expectedMatch: true,
		},
		"semantically equal - object whitespace difference": {
			currentJson: jsontypes.NewNormalizedValue(`{
				"hello": "world",
				"nums": [1, 2, 3],
				"nested": {
					"test-bool": true
				}
			}`),
			givenJson:     jsontypes.NewNormalizedValue(`{"hello":"world","nums":[1,2,3],"nested":{"test-bool":true}}`),
			expectedMatch: true,
		},
		"semantically equal - array whitespace difference": {
			currentJson: jsontypes.NewNormalizedValue(`[
				{
				  "hello": "world"
				},
				{
				  "nums": [
					1,
					2,
					3
				  ]
				},
				{
				  "nested": {
					"test-bool": true
				  }
				}
			  ]`),
			givenJson:     jsontypes.NewNormalizedValue(`[{"hello": "world"}, {"nums":[1, 2, 3]}, {"nested": {"test-bool": true}}]`),
			expectedMatch: true,
		},
		"error - invalid json": {
			currentJson:   jsontypes.NewNormalizedValue(`{"hello": "world", "nums": [1, 2, 3], "nested": {"test-bool": true}}`),
			givenJson:     jsontypes.NewNormalizedValue(`&#$^"hello": "world", "nums": [1, 2, 3], "nested": {"test-bool": true}}`),
			expectedMatch: false,
			expectedDiags: diag.Diagnostics{
				diag.NewErrorDiagnostic(
					"Semantic Equality Check Error",
					"An unexpected error occurred while performing semantic equality checks. "+
						"Please report this to the provider developers.\n\n"+
						"Error: invalid character '&' looking for beginning of value",
				),
			},
		},
		"error - not given normalized json value": {
			currentJson:   jsontypes.NewNormalizedValue(`{"hello": "world", "nums": [1, 2, 3], "nested": {"test-bool": true}}`),
			givenJson:     basetypes.NewStringValue(`{"hello": "world", "nums": [1, 2, 3], "nested": {"test-bool": true}}`),
			expectedMatch: false,
			expectedDiags: diag.Diagnostics{
				diag.NewErrorDiagnostic(
					"Semantic Equality Check Error",
					"An unexpected value type was received while performing semantic equality checks. "+
						"Please report this to the provider developers.\n\n"+
						"Expected Value Type: jsontypes.Normalized\n"+
						"Got Value Type: basetypes.StringValue",
				),
			},
		},
		// JSON Semantic equality uses (decoder).UseNumber to avoid Go parsing JSON numbers into float64. This ensures that Go
		// won't normalize the JSON number representation or impose limits on numeric range.
		"not equal - different JSON number representations": {
			currentJson:   jsontypes.NewNormalizedValue(`{"large": 12423434}`),
			givenJson:     jsontypes.NewNormalizedValue(`{"large": 1.2423434e+07}`),
			expectedMatch: false,
		},
		"semantically equal - larger than max float64 values": {
			currentJson:   jsontypes.NewNormalizedValue(`{"large": 1.79769313486231570814527423731704356798070e+309}`),
			givenJson:     jsontypes.NewNormalizedValue(`{"large": 1.79769313486231570814527423731704356798070e+309}`),
			expectedMatch: true,
		},
	}
	for name, testCase := range testCases {
		name, testCase := name, testCase
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			match, diags := testCase.currentJson.StringSemanticEquals(context.Background(), testCase.givenJson)

			if testCase.expectedMatch != match {
				t.Errorf("Expected StringSemanticEquals to return: %t, but got: %t", testCase.expectedMatch, match)
			}

			if diff := cmp.Diff(diags, testCase.expectedDiags); diff != "" {
				t.Errorf("Unexpected diagnostics (-got, +expected): %s", diff)
			}
		})
	}
}

func TestNormalizedUnmarshal(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		json          jsontypes.Normalized
		target        any
		output        any
		expectedDiags diag.Diagnostics
	}{
		"normalized value is null ": {
			json: jsontypes.NewNormalizedNull(),
			target: struct {
				Hello string `json:"hello"`
			}{},
			expectedDiags: diag.Diagnostics{
				diag.NewErrorDiagnostic(
					"Normalized JSON Unmarshal Error",
					"json string value is null",
				),
			},
		},
		"normalized value is unknown ": {
			json: jsontypes.NewNormalizedUnknown(),
			target: struct {
				Hello string `json:"hello"`
			}{},
			expectedDiags: diag.Diagnostics{
				diag.NewErrorDiagnostic(
					"Normalized JSON Unmarshal Error",
					"json string value is unknown",
				),
			},
		},
		"invalid target - not a pointer ": {
			json: jsontypes.NewNormalizedValue(`{"hello": "world"}`),
			target: struct {
				Hello string `json:"hello"`
			}{},
			expectedDiags: diag.Diagnostics{
				diag.NewErrorDiagnostic(
					"Normalized JSON Unmarshal Error",
					"json: Unmarshal(non-pointer struct { Hello string \"json:\\\"hello\\\"\" })",
				),
			},
		},
		"valid target ": {
			json: jsontypes.NewNormalizedValue(`{"hello": "world", "nums": [1, 2, 3], "test-bool": true}`),
			target: &struct {
				Hello   string `json:"hello"`
				Numbers []int  `json:"nums"`
				Test    bool   `json:"test-bool"`
			}{},
			output: &struct {
				Hello   string `json:"hello"`
				Numbers []int  `json:"nums"`
				Test    bool   `json:"test-bool"`
			}{
				Hello:   "world",
				Numbers: []int{1, 2, 3},
				Test:    true,
			},
		},
	}
	for name, testCase := range testCases {
		name, testCase := name, testCase
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			diags := testCase.json.Unmarshal(testCase.target)

			if diff := cmp.Diff(diags, testCase.expectedDiags); diff != "" {
				t.Errorf("Unexpected diagnostics (-got, +expected): %s", diff)
			}
		})
	}
}
