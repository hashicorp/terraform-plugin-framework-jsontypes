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
		"not equal - additional field": {
			currentJson:   jsontypes.NewNormalizedValue(`{"hello": "world", "nums": [1, 2, 3], "nested": {"test-bool": true}, "new-field": null}`),
			givenJson:     jsontypes.NewNormalizedValue(`{"hello": "world", "nums": [1, 2, 3], "nested": {"test-bool": true}}`),
			expectedMatch: false,
		},
		"semantically equal - byte-for-byte match": {
			currentJson:   jsontypes.NewNormalizedValue(`{"hello": "world", "nums": [1, 2, 3], "nested": {"test-bool": true}}`),
			givenJson:     jsontypes.NewNormalizedValue(`{"hello": "world", "nums": [1, 2, 3], "nested": {"test-bool": true}}`),
			expectedMatch: true,
		},
		"semantically equal - order difference": {
			currentJson:   jsontypes.NewNormalizedValue(`{"hello": "world", "nums": [1, 2, 3], "nested": {"test-bool": true}}`),
			givenJson:     jsontypes.NewNormalizedValue(`{"nums": [1, 2, 3], "nested": {"test-bool": true}, "hello": "world"}`),
			expectedMatch: true,
		},
		"semantically equal - whitespace difference": {
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
		"error - invalid json": {
			currentJson:   jsontypes.NewNormalizedValue(`{"hello": "world", "nums": [1, 2, 3], "nested": {"test-bool": true}}`),
			givenJson:     jsontypes.NewNormalizedValue(`"hello": "world", "nums": [1, 2, 3], "nested": {"test-bool": true}}`),
			expectedMatch: false,
			expectedDiags: diag.Diagnostics{
				diag.NewErrorDiagnostic(
					"Semantic Equality Check Error",
					"An unexpected error occurred while performing semantic equality checks. "+
						"Please report this to the provider developers.\n\n"+
						"Error: invalid character ':' after top-level value",
				),
			},
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
				t.Errorf("Unexpected diagnostics (+got, -expected): %s", diff)
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
				t.Errorf("Unexpected diagnostics (+got, -expected): %s", diff)
			}
		})
	}
}
