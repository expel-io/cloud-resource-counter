# cloud-resource-counter

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
--trace-file TF  | Write a trace of all AWS calls to file TF.
--version        | Display version information and then exit.

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
                "rds:DescribeDBInstances"
            ],
            "Resource": "*"
        }
    ]
}
```
