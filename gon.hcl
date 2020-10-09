source = ["./dist/crc-macos_darwin_amd64/cloud-resource-counter"]
bundle_id = "com.expel.cloud-resource-counter"

apple_id { }

sign {
	application_identity = "56bab15156129308a6ecb290dc759bab3abe9666"
}

zip {
	output_path = "cloud-resource-counter.zip"
}
