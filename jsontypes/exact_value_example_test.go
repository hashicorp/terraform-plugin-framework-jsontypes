package jsontypes_test

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

type ExactResourceModel struct {
	Json jsontypes.Exact `tfsdk:"json"`
}

type ExactJson struct {
	Hello   string `json:"hello"`
	Numbers []int  `json:"numbers"`
}

func ExampleExact_Unmarshal() {
	var diags diag.Diagnostics

	// For example purposes, typically the data model would be populated automatically by Plugin Framework via Config, Plan or State.
	// https://developer.hashicorp.com/terraform/plugin/framework/handling-data/accessing-values
	data := ExactResourceModel{
		Json: jsontypes.NewExactValue("{\"hello\":\"world\", \"numbers\": [1, 2, 3]}"),
	}

	// Check that the JSON data is known and able to be unmarshalled
	if !data.Json.IsNull() && !data.Json.IsUnknown() {
		var jsonStruct ExactJson

		diags.Append(data.Json.Unmarshal(&jsonStruct)...)
		if diags.HasError() {
			return
		}

		// Output: {world [1 2 3]}
		fmt.Printf("%v\n", jsonStruct)
	}
}
