# cloudstaticfiles

Simple CLI command to sync a directory to a cloud bucket (aws, gcp, azure). It
uses a lockfile to indicate if a directory has been copied already.  This
command is especially helpful when run on docker startup to copy for example
static files to a target store

```bash
staticfiles-sync -s .next/static -t gcp://my-static-files-bucket/_next/static -l .locks/v1.0.0;
```
