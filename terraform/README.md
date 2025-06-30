# Terraform

This directory holds the infrastructure to run Eskimo in AWS environment. Once `bootstrap` and `aws` are created,
one can trigger the scanner using AWS CLI:

```sh
aws events put-events \
  --entries '[{
    "Source":"eskimo.manual",
    "DetailType":"manual trigger",
    "Detail":"{}"
  }]'
```
