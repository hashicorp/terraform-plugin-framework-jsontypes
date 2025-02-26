// Copyright (c) HashiCorp, Inc.
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

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
)

func TestExactValidateAttribute(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		exact         jsontypes.Exact
		expectedDiags diag.Diagnostics
	}{
		"empty-struct": {
			exact: jsontypes.Exact{},
		},
		"null": {
			exact: jsontypes.NewExactNull(),
		},
		"unknown": {
			exact: jsontypes.NewExactUnknown(),
		},
		"valid json object": {
			exact: jsontypes.NewExactValue(`{"hello":"world", "array": [1, 2, 3]}`),
		},
		"valid json array": {
			exact: jsontypes.NewExactValue(`["hello", "world"]`),
		},
		"invalid json - bracket mismatch": {
			exact: jsontypes.NewExactValue(`{"hello":"world"`),
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
			exact: jsontypes.NewExactValue("notvalidjson123"),
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

			testCase.exact.ValidateAttribute(
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

func TestExactValidateParameter(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		exact           jsontypes.Exact
		expectedFuncErr *function.FuncError
	}{
		"empty-struct": {
			exact: jsontypes.Exact{},
		},
		"null": {
			exact: jsontypes.NewExactNull(),
		},
		"unknown": {
			exact: jsontypes.NewExactUnknown(),
		},
		"valid json object": {
			exact: jsontypes.NewExactValue(`{"hello":"world", "array": [1, 2, 3]}`),
		},
		"valid json array": {
			exact: jsontypes.NewExactValue(`["hello", "world"]`),
		},
		"invalid json - bracket mismatch": {
			exact: jsontypes.NewExactValue(`{"hello":"world"`),
			expectedFuncErr: function.NewArgumentFuncError(
				0,
				"Invalid JSON String Value: "+
					"A string value was provided that is not valid JSON string format (RFC 7159).\n\n"+
					"Given Value: {\"hello\":\"world\"\n",
			),
		},
		"invalid json - normal string": {
			exact: jsontypes.NewExactValue("notvalidjson123"),
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

			testCase.exact.ValidateParameter(
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

func TestExactUnmarshal(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		json          jsontypes.Exact
		target        any
		output        any
		expectedDiags diag.Diagnostics
	}{
		"exact value is null ": {
			json: jsontypes.NewExactNull(),
			target: struct {
				Hello string `json:"hello"`
			}{},
			expectedDiags: diag.Diagnostics{
				diag.NewErrorDiagnostic(
					"Exact JSON Unmarshal Error",
					"json string value is null",
				),
			},
		},
		"exact value is unknown ": {
			json: jsontypes.NewExactUnknown(),
			target: struct {
				Hello string `json:"hello"`
			}{},
			expectedDiags: diag.Diagnostics{
				diag.NewErrorDiagnostic(
					"Exact JSON Unmarshal Error",
					"json string value is unknown",
				),
			},
		},
		"invalid target - not a pointer ": {
			json: jsontypes.NewExactValue(`{"hello": "world"}`),
			target: struct {
				Hello string `json:"hello"`
			}{},
			expectedDiags: diag.Diagnostics{
				diag.NewErrorDiagnostic(
					"Exact JSON Unmarshal Error",
					"json: Unmarshal(non-pointer struct { Hello string \"json:\\\"hello\\\"\" })",
				),
			},
		},
		"valid target ": {
			json: jsontypes.NewExactValue(`{"hello": "world", "nums": [1, 2, 3], "test-bool": true}`),
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
