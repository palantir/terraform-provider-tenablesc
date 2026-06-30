package fail

fail

/*
This is a non-compiling file that has been added to explicitly ensure that CI fails.
It also contains the command that caused the failure and its output.
Remove this file if debugging locally.

./godelw verify failed after updating godel plugins and assets

Command that caused error:
./godelw verify --skip-test --skip-lint

Output:
Running format...
Running mod...
Running generate...
rendering website for provider "terraform-provider-tenablesc" (as "terraform-provider-tenablesc")
copying any existing content to tmp dir
exporting schema from Terraform
compiling provider "tenablesc"
using Terraform CLI binary from PATH if available, otherwise downloading latest Terraform CLI binary
Error executing command: unable to generate website: error exporting provider schema from Terraform: unable to download Terraform binary: Unknown status: 502

exit status 1
main.go:26: running "go": exit status 1
Error: failed to run go generate in "/repo": exit status 1
Running license...
Running distgo-task...
Failed tasks:
	generate

*/
