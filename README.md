# cloud-resource-counter

![Build, Lint and Test](https://github.com/expel-io/cloud-resource-counter/workflows/Build,%20Lint%20and%20Test/badge.svg?branch=master)

Go utility for counting the resources in use at a cloud infrastructure provider.

The cloud resource counter utility known as "cloud-resource-counter" inspects
a cloud deployment (for now, only Amazon Web Services) to assess the number of
distinct computing resources. The result is a CSV file that describes the counts
of each.

This command requires access to a valid AWS Account. For now, it is assumed that
this is stored in the user's `.aws` folder (located in `$HOME/.aws`).

A future version of this will allow the caller to supply credentials in more
flexible ways.

## Command Line

The following command line arguments are supported:

Argument         | Meaning
-----------------|----------------------------------
--all-regions    | View resource counts for all regions supported by the account.
--append         | Append (rather than overwrite) the output file.
--help           | Information on the command line options
--output-file OF | Write the results in Comma Separated Values format to file OF.
--profile PN     | Use the credentials associated with shared profile named PN.
--region RN      | View resource counts for the AWS region RN.
--trace-file TF  | Write a trace of all AWS calls to file TF.
--version        | Display version information and then exit.

## Installing

You can build this from source or use the precompiled binaries (see the [Releases](https://github.com/expel-io/cloud-resource-counter/releases) page for binaries). We provided binaries for Linux (x86_64 and i386) and MacOS. There is no installation process as this is simply a command line tool. To unzip from the command line, use:

```bash
$ tar -Zxvf cloud-resource-counter_<<RELEASE_VERSION>>_<<PLATFORM>>_<<ARCH>>.tar.gz
x README
x cloud-resource-counter
```

The result is a binary called `cloud-resource-counter` in the current directory.

These binaries run on Linux OSes (32- and 64-bit versions) and MacOS (Go 1.15 requires macOS 10.12 Sierra or later).

### MacOS Download

If you are using MacOS Catalina, there is a stricter process for running binaries produced by third party developers. You must allow "App Store and identified developers" for the binary to run. Here are the detailed steps:

1. From the Apple menu, click "System Preferences".
1. Select "Security and Privacy"
1. If the settings are locked, unlock them. This requires you to enter your password.
1. From the "Allow apps downloaded from:" section, choose "App Store and identified developers".
1. You can lock your settings if you like.

## Building from Source

You can also build this utility directly from source. We have built and tested this with the following Go versions:

* v1.14
* v1.15

To run from source, use the following command line:

```bash
// Assumes that you are inside the cloud-resource-counter folder
$ go run . --help
```

To run the unit tests, use the following command line:

```bash
// Assumes that you are inside the cloud-resource-counter folder
$ go test . -v
```

## Minimal IAM Policy

To use this utility, this minimal IAM Profile can be associated with a bare user account:

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
                "ec2:DescribeVolumes",
                "ecs:DescribeTaskDefinition",
                "ecs:ListTaskDefinitions",
                "lambda:ListFunctions",
                "lightsail:GetInstances",
                "lightsail:GetRegions",
                "rds:DescribeDBInstances"
                "s3:ListAllMyBuckets"
            ],
            "Resource": "*"
        }
    ]
}
```

## Resources Counted

The `cloud-resource-counter` examines the following resources:

1. **Account ID.** We use the Security Token Service to collect the account ID associated with the caller.

   * This is stored in the generated CSV file under the "Account ID" column.

2. **EC2**. We count the number of EC2 instances (both "normal" and Spot instances) across all regions.

   * For EC2 instances, we only count those _without_ an Instance Lifecycle tag (which is either `spot` or `scheduled`).
   * For Spot instance, we only count those with an Instance Lifecycle tag of `spot`.

   * This is stored in the generated CSV file under the "# of EC2 Instances" and "# of Spot Instances" columns.

3. **EBS Volumes.** We count the number of "attached" EBS volumes across all regions.

   * We only count those EBS volumes that are "attached" to an EC2 instance. 

   * This is stored in the generated CSV file under the "# of EBS Volumes" column.

4. **Unique ECS Containers.** We count the number of "unique" ECS containers across all regions.

   * We look at all task definitions and collect all of the `Image` name fields inside the Container Definitions. 
   * We then simply count the number of unique `Image` names _across all regions._ This is the only resource counted this way.
   * This is stored in the generated CSV file under the "# of Unique Containers" column.

5. **Lambda Functions.** We count the number of all Lambda functions across all regions.

   * We do not qualify the type of Lambda function.
   * This is stored in the generated CSV file under the "# of Lambda Functions" column.

6. **Lightsail Instances.** We count the number of Lightsail instances across all regions.

   * We do not qualify the type of Lightsail instance.
   * This is stored in the generated CSV file under the "# of Lightsail Instances" column.

7. **RDS Instances.** We count the number of RDS instance across all regions.

   * We do not qualify the type of RDS instance.
   * This is stored in the generated CSV file under the "# of RDS Instances" column.

8. **S3 Buckets.** We count the number of S3 buckets across all regions.

   * We do not qualify the type of S3 bucket.
   * This is stored in the generated CSV file under the "# of S3 Buckets" column.

## Alternative Means of Resource Counting

If you do not wish to use the `cloud-resource-counter` utility, you can use the AWS CLI to collect these same counts. For some of these counts, it will be easy to do. For others, the command line is a bit more complex.

==For the purposes of explaining these scripts, we are using a Bash command line on a Unix operating system. If you use another command line processor (or OS), please adapt the script appropriately.==

It is also assumed that you have configured the AWS CLI for your desired profile. ==For our purposes, we will assume that correct profile is called `my-profile`.==

### Account ID

To collect the account ID, use the AWS CLI `sts` command, as in:

```bash
$ aws sts get-caller-identity --profile my-profile --output text --query Account
123456789012
```

### EC2 Instances

#### EC2 Regions

To collect the total number of EC2 instances across all regions, we will need to run two AWS CLI commands. First, let's get the list of accessible regions where your EC2 instances are located:

```bash
$ aws ec2 describe-regions --profile my-profile --filters "Name=opt-in-status,Values=opt-in-not-required,opted-in" \
		--output text --query Regions[].RegionName
eu-north-1    ap-south-1    eu-west-3 ...
```

It filters the list of regions to just those that are either "opted in" or where "opt in" is not required.

We output the query as text and just extract the `RegionName` field from the JSON structure. 

(If your profile does not have a default region, then you should `--region us-east-1` to the above command line. You need to direct this request to a valid region.)

We will be using the results of this command to "iterate" over all regions.

#### EC2 Instances

Here is the command to count the number of _normal_ EC2 instances (those that are _not_ Spot nor Scheduled instances) for a given region:

```bash
$ aws ec2 describe-instances --profile my-profile --region us-east-1 \
		--query 'length(Reservations[].Instances[?!not_null(InstanceLifecycle)].InstanceId[])'
4
```

The number 4 above means that there were 4 EC2 instances found.

By default, the EC2 `describe-instances` "normal" EC2 instances as well as those that are "spot" instances. As such, the query argument does the following:

1. Find all Instances (in all Reservations) and qualify each:
   * `InstanceLifecycle` attribute is not (`!`) non-null (`not_null`).
     * This is effectively saying, where `InstanceLifecycle` is null.
     * The language specification (JMESPath) does not have a `null()` function.
   * Get the `InstanceId` for each of these matching Instances and form into a flattened array.
2. Get the length of that array.

We will need to run this command over all regions. Here is what it looks like:

```bash
$ for reg in $(aws ec2 describe-regions --profile my-profile --filters "Name=opt-in-status,Values=opt-in-not-required,opted-in" \
	--output text --query Regions[].RegionName); do \
		aws ec2 describe-instances --profile my-profile --region $reg \
				--filters "length(Reservations[].Instances[?!not_null(InstanceLifecycle)].InstanceId[])" \
  done | paste -s -d+ - | bc
 23
```

The first two lines allow us to loop over all regions (using the variable `reg` to hold the current value).

We then paste all of the values into a long addition and use `bc` to sum the values.

