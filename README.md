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
--output-file OF | Write the results to file OF
--profile PN     | Use the credentials associated with shared profile PN
--region RN      | View resource counts for the AWS region RN
--version        | Display version information
