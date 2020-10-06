# cloud-resource-counter

![Build, Lint and Test](https://github.com/expel-io/cloud-resource-counter/workflows/Build,%20Lint%20and%20Test/badge.svg?branch=master)

Go utility for counting the resources in use at a cloud infrastructure provider.

The cloud resource counter utility known as "cloud-resource-counter" inspects
a cloud deployment (for now, only Amazon Web Services) to assess the number of
distinct computing resources. The result is a CSV file that describes the counts
of each.

This command requires access to a valid AWS Account. For now, it is assumed that
this is stored in the user's ".aws" folder (located in $HOME/.aws).

A future version of this will allow the caller to supply credentials in more
flexible ways.

## Command Line

The following command line arguments are supported:

Argument         | Meaning
-----------------|----------------------------------
--output-file OF | Write the results in Comma Separated Values format to file OF.
--profile PN     | Use the credentials associated with shared profile PN.
--region RN      | View resource counts for the AWS region RN.
--all-regions    | View resource counts for all regions supported by the account.
--trace-file TF  | Write a trace of all AWS calls to file TF.
--version        | Display version information and then exit.

## Installing

You can build this from source or use the precompiled binaries (see the [Releases](https://github.com/expel-io/cloud-resource-counter/releases) page for binaries). We provided binaries for Linux (x86_64 and i386) and MacOS. There is no installation process as this is simply a command line tool. To unzip from the command line, use:

```Bash
$ tar -Zxvf cloud-release-counter_<<RELEASE_VERSION>>_<<PLATFORM>>_<<ARCH>>.tar.gz
x README
x cloud-resource-counter
```

The result is a binary called `cloud-resource-counter` in the current directory.

### MacOS Download

If you are using MacOS Catalina, there is a stricter process for allowing third party developers to run binaries. You must allow "App Store and identified developers" for the binary to run. Here are the detailed steps:

1. From the Apple menu, click "System Preferences".
1. Select "Security and Privacy"
1. If the settings are locked, unlock them. This requires you to enter your password.
1. From the "Allow apps downloaded from:" section, choose "App Store and identified developers".
1. You can lock your settings if you like.

## Minimal IAM Policy

To use this utility, a bare minimal IAM Profile can be associated with anotherwise bare user account:

```JSON
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Sid": "VisualEditor0",
            "Effect": "Allow",
            "Action": [
                "ec2:DescribeInstances",
                "ec2:DescribeRegions",
                "lambda:ListFunctions",
                "rds:DescribeDBInstances"
                "s3:ListAllMyBuckets",
            ],
            "Resource": "*"
        }
    ]
}
```
