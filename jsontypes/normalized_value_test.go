// Copyright IBM Corp. 2023, 2025
// SPDX-License-Identifier: MPL-2.0

package jsontypes_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/attr/xattr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
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
		// JSON Semantic equality uses Go's encoding/json library, which replaces some characters to escape codes
		"semantically equal - HTML escape characters are equal": {
			currentJson:   jsontypes.NewNormalizedValue(`{"url_ampersand": "http://example.com?foo=bar&hello=world", "left-caret": "<", "right-caret": ">"}`),
			givenJson:     jsontypes.NewNormalizedValue(`{"url_ampersand": "http://example.com?foo=bar\u0026hello=world", "left-caret": "\u003c", "right-caret": "\u003e"}`),
			expectedMatch: true,
		},
	}
	for name, testCase := range testCases {

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

func TestNormalizedValidateAttribute(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		normalized    jsontypes.Normalized
		expectedDiags diag.Diagnostics
	}{
		"empty-struct": {
			normalized: jsontypes.Normalized{},
		},
		"null": {
			normalized: jsontypes.NewNormalizedNull(),
		},
		"unknown": {
			normalized: jsontypes.NewNormalizedUnknown(),
		},
		"valid json object": {
			normalized: jsontypes.NewNormalizedValue(`{"hello":"world", "array": [1, 2, 3]}`),
		},
		"valid json array": {
			normalized: jsontypes.NewNormalizedValue(`["hello", "world"]`),
		},
		"invalid json - bracket mismatch": {
			normalized: jsontypes.NewNormalizedValue(`{"hello":"world"`),
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("test"),
					"Invalid JSON String Value",
					"A string value was provided that is not valid JSON string format (RFC 7159).\n\n"+
						"Given Value: {\"hello\":\"world\"\n",
				),
			},
		},
		"invalid json - normal string": {
			normalized: jsontypes.NewNormalizedValue("notvalidjson123"),
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("test"),
					"Invalid JSON String Value",
					"A string value was provided that is not valid JSON string format (RFC 7159).\n\n"+
						"Given Value: notvalidjson123\n",
				),
			},
		},
	}
	for name, testCase := range testCases {

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			resp := xattr.ValidateAttributeResponse{}

			testCase.normalized.ValidateAttribute(
				context.Background(),
				xattr.ValidateAttributeRequest{
					Path: path.Root("test"),
				},
				&resp,
			)

			if diff := cmp.Diff(resp.Diagnostics, testCase.expectedDiags); diff != "" {
				t.Errorf("Unexpected diagnostics (-got, +expected): %s", diff)
			}
		})
	}
}

func TestNormalizedValidateParameter(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		normalized      jsontypes.Normalized
		expectedFuncErr *function.FuncError
	}{
		"empty-struct": {
			normalized: jsontypes.Normalized{},
		},
		"null": {
			normalized: jsontypes.NewNormalizedNull(),
		},
		"unknown": {
			normalized: jsontypes.NewNormalizedUnknown(),
		},
		"valid json object": {
			normalized: jsontypes.NewNormalizedValue(`{"hello":"world", "array": [1, 2, 3]}`),
		},
		"valid json array": {
			normalized: jsontypes.NewNormalizedValue(`["hello", "world"]`),
		},
		"invalid json - bracket mismatch": {
			normalized: jsontypes.NewNormalizedValue(`{"hello":"world"`),
			expectedFuncErr: function.NewArgumentFuncError(
				0,
				"Invalid JSON String Value: "+
					"A string value was provided that is not valid JSON string format (RFC 7159).\n\n"+
					"Given Value: {\"hello\":\"world\"\n",
			),
		},
		"invalid json - normal string": {
			normalized: jsontypes.NewNormalizedValue("notvalidjson123"),
			expectedFuncErr: function.NewArgumentFuncError(
				0,
				"Invalid JSON String Value: "+
					"A string value was provided that is not valid JSON string format (RFC 7159).\n\n"+
					"Given Value: notvalidjson123\n",
			),
		},
	}
	for name, testCase := range testCases {

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			resp := function.ValidateParameterResponse{}

			testCase.normalized.ValidateParameter(
				context.Background(),
				function.ValidateParameterRequest{
					Position: 0,
				},
				&resp,
			)

			if diff := cmp.Diff(resp.Error, testCase.expectedFuncErr); diff != "" {
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

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			diags := testCase.json.Unmarshal(testCase.target)

			if diff := cmp.Diff(diags, testCase.expectedDiags); diff != "" {
				t.Errorf("Unexpected diagnostics (-got, +expected): %s", diff)
			}
		})
	}
}
