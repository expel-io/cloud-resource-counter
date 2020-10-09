# CI/CD Process

This page describes the Continuous Integration/Continuous Delivery configuration for this project. It is implemented using Github Actions (stored in `.github/workflows`). It is composed of two workflows:

* **Build, Lint and Test.** This is run for every pull request that is merged into master. It is described in `build.yml`. It performs the following steps on the "cross product" of Go versions (1.14, 1.15) and target OS (Linux, MacOS):

  1. Setup Go (for the specific version)
  1. Get all of the Go source code for this repository
  1. Collect all of the dependent Go modules
  1. Build the project
  1. Run Lint
  1. Run the automated unit test suite

  If any step fails, the overall process fails.

* **Release.** This is run whenever a new tag is pushed. This indicates the completion of a phase of development which can be shared as a specific version. It is described in `release.yml`. This workflow is executed on a MacOS runner and performs the following steps:

  1. Setup Go (currently: 1.15)
  1. Get all of the Go source code for this repository
  1. Collect all of the dependent Go modules
  1. Import the Code Signing Certificates into the Keychain (for details, see the section on "Code Signing and Notarization")
  1. Install the `gon` tool via HomeBrew. This does the actual signing and notarization. The configuration file for `gon` is `gon.hcl`.
  1. Invoke `goreleaser` to construct the binary and invoke `gon`. The configuration file for `goreleaser` is `.goreleaser.yml`.

  Of course, if any step fails, the overall workflow fails.

## Code Signing and Notarization

The MacOS operating system (starting with Catalina) is much more strict as it relates to allowing binaries from the Internet to be downloaded and run. To allow the `cloud-resource-counter` binary to be run on customer's machine, the customer must allow binaries from "identified developers" and our binary needs to be "signed and notarized".

The process of code signing involves Apple Services to complete (along with an Apple ID account and an Apple Developer ID Application certificate).

For details, see [Notarizing macOS Software Before Distribution](https://developer.apple.com/documentation/xcode/notarizing_macos_software_before_distribution).

Here is how we implement this process in our Release workflow:

1. We use an open source tool called `gon` ([Github](https://github.com/mitchellh/gon)) which is a command line tool for MacOS signing and notarization.

1. This tool is invoked as part of an open source tool called `goreleaser` ([website](https://goreleaser.com/)). This tool has the ability to compile Go source code for a number of different target operating systems.

1. Rather than put any sensitive data in our open source files, we store all this information as Github Secrets, which are only accessible to administrators of this repository. There are currently four relevant secrets:

   * `AC_USERNAME` and `AC_PASSWORD`: These are the user name and password to be used in the signing process. These credentials refer to a register Apple ID account.
   * `APPLE_DEVELOPER_CERTIFICATE_P12_BASE64` and `APPLE_DEVELOPER_CERTIFICATE_PASSWORD`: These are the base64 certificate and password for the associated Apple Developer ID Application certificate.

1. The SHA-1 hash for the Apple Developer ID Application certificate is stored in the `gon.hcl` file as part of the `application_identity` object.
