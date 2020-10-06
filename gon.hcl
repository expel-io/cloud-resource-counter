source = ["./dist/crc-macos_darwin_amd64/cloud-resource-counter"]
bundle_id = "com.expel.cloud-resource-counter"

apple_id { }

sign {
	application_identity = "34A9AF032F4967131A552BAA92FC2E2E3556AA84"
}

zip {
	output_path = "cloud-resource-counter.zip"
}
