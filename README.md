# staticfiles-sync

Simple CLI command to sync a directory to a cloud bucket (aws, gcp, azure). It
uses a lockfile to indicate if a directory has been copied already.  This
command is especially helpful when run on docker startup to copy for example
static files to a target store

```bash
staticfiles-sync -s .next/static -t gcp://my-static-files-bucket/_next/static -l .locks/v1.0.0;
```


##  Usage

In your Dockerfile add the following to install:

```
ARG STATICFILES_SYNC_VERSION=2.0.0-beta-19df697

RUN curl -L https://github.com/labdigital-evolve/staticfiles-sync/releases/download/v${STATICFILES_SYNC_VERSION}/staticfiles-sync_${STATICFILES_SYNC_VERSION}_linux_amd64.zip -o /tmp/staticfiles-sync.zip \
	&& unzip /tmp/staticfiles-sync.zip -d /tmp/staticfiles-sync \
	&& install -m 0755 /tmp/staticfiles-sync/staticfiles-sync_v${STATICFILES_SYNC_VERSION} /usr/local/bin/staticfiles-sync \
	&& rm -rf /tmp/staticfiles-sync.zip /tmp/staticfiles-sync

```
