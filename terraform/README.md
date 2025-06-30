# Terraform

This directory holds the infrastructure to run Eskimo in AWS environment. Once `bootstrap` and `aws` are created,
the scan schedule can be adjusted using the `scan_schedule_expression` variable, which holds the raw cron expression. One can also trigger the scanner manually using AWS CLI:

```sh
aws events put-events \
  --entries '[{
    "Source":"eskimo.manual",
    "DetailType":"manual trigger",
    "Detail":"{}"
  }]'
```
