# 5. S3 Bucket Logging for Terraform State

## Status
Accepted

## Context
The bootstrap module created an encrypted S3 bucket to store Terraform state but did not record access to it. Capturing server access logs helps track any reads or writes to the state files.

## Decision
We introduced a separate bucket `eskimo-tf-logs` managed by the same S3 module. Logging is enabled on the state bucket with logs delivered to this new bucket. The log bucket is versioned, encrypted and allows log delivery from S3 only.

## Consequences
- Access to the Terraform state bucket is now auditable via standard S3 logs.
- A small additional bucket must be managed and retained for compliance.
