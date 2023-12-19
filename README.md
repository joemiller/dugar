# gcp-iam-lookup

> it's *`du(1)` for Google Artifact Registry (GAR)*

Lists the total size of all image tags in a Google Artifact Registry docker repository.

## Install

`go install github.com/joemiller/dugar@latest`

## Usage

Requires Application Default Credentials. If running localy with `gcloud` you can run `gcloud auth application-default login` to setup local ADC.

example:

```console
$ dugar -project my-project -location us -repo my-repo

Analyzing projects/planetscale-registry/locations/us/repositories/my-repo ...
48.81 GiB       us-docker.pkg.dev/my-project/my-repo/web-app
6.30 GiB        us-docker.pkg.dev/my-project/my-repo/db-runtime
0.09 GiB        us-docker.pkg.dev/my-project/my-repo/worker-foo
0.03 GiB        us-docker.pkg.dev/my-project/my-repo/another-app
55.23 GiB       .
```

The `-format` flag can be used to change the output format. The default is `GiB`.

Use `-h` for help.