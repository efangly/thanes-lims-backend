# 0009 - OCI Object Storage via its S3-compatible API

## Status
Accepted (2026-09-08)

## Context
Document files were stored on a self-hosted MinIO instance
(`siamatic.thddns.net:9000`). We want object storage on OCI alongside the rest
of the infrastructure (ADB, compute) for managed durability, monitoring and a
single access-control surface. Its long-lived secret key was also committed to
`.env` in the repo.

The existing adapter (`internal/adapters/objectstorage`) already speaks S3 via
`github.com/minio/minio-go/v7`, and only uses a small subset of operations:
`PutObject`, `GetObject`, `PresignedGetObject`, `RemoveObject`, `BucketExists`.

## Decision
Use **OCI Object Storage through its S3-compatible endpoint**, keeping
`minio-go` as the client. We do not adopt the native OCI Go SDK.

- Endpoint: `<namespace>.compat.objectstorage.<region>.<realm-domain>` - this
  tenancy is the **OC43 realm**, so the domain is `oci.thaiaiscloud.com`, e.g.
  `ax45vpor5h7p.compat.objectstorage.ap-samutprakan-1.oci.thaiaiscloud.com`
  (not `oraclecloud.com`).
- Credentials: an IAM user's **Customer Secret Key** (access key id + secret),
  created with `oci iam customer-secret-key create --user-id <ocid>`.
- The client must set `Region` (real OCI region id, e.g. `ap-samutprakan-1`) -
  OCI signs SigV4 with it - and `BucketLookup: minio.BucketLookupPath`, since
  OCI only serves path-style requests.
- Buckets are provisioned via `oci-cli`, not code. `EnsureBucket` now only
  verifies existence (OCI's S3 API does not support `CreateBucket`).
- Config keys renamed `MINIO_*` -> `STORAGE_*`, adding `STORAGE_REGION`. Package
  `internal/adapters/minio` renamed to `internal/adapters/objectstorage`.

## Consequences
- Minimal code change; domain/application layers untouched (port unchanged).
- Not locked to OCI - any S3-compatible provider still works by config alone.
- Presigned GET URLs now point at `*.oci.thaiaiscloud.com`; OCI supports up to 7-day
  expiry (we use 15 min). Frontend must open the URL directly rather than via
  `fetch()`/XHR - OCI's S3 compat layer has limited CORS support.
- Existing objects were copied MinIO -> OCI with `rclone` before cut-over;
  `StorageKey` values in Postgres are unchanged (same key layout both sides).
- The old MinIO access key should be rotated/revoked.
