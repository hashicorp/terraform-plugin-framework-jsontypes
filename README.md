[![PkgGoDev](https://pkg.go.dev/badge/github.com/hashicorp/terraform-plugin-framework-jsontypes)](https://pkg.go.dev/github.com/hashicorp/terraform-plugin-framework-jsontypes)

# Terraform Plugin Framework JSON Types

terraform-plugin-framework-jsontypes is a Go module containing common [Custom Type](https://developer.hashicorp.com/terraform/plugin/framework/handling-data/custom-types) implementations for [terraform-plugin-framework](https://github.com/hashicorp/terraform-plugin-framework). It aims to provide RFC-based validation and semantic equality for types related to JSON strings ([RFC7159](https://www.rfc-editor.org/rfc/rfc7159.html)).

## Terraform Plugin Framework Compatibility

This Go module is typically kept up to date with the latest `terraform-plugin-framework` releases to ensure all Custom Type functionality is available.

## Go Compatibility

This Go module follows `terraform-plugin-framework` Go compatibility.

Currently that means Go **1.22** must be used when developing and testing code.

## Contributing

See [`.github/CONTRIBUTING.md`](.github/CONTRIBUTING.md)

## License

[Mozilla Public License v2.0](LICENSE)
