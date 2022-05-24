source = ["./dist/crc-macos_darwin_amd64/aws-resource-counter"]
bundle_id = "com.expel.aws-resource-counter"

apple_id { }

sign {
	application_identity = "71c685df24be2279027ca972134afd5c715ed841"
}

zip {
	output_path = "aws-resource-counter.zip"
}
