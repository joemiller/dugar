# dugar

> it's *`du(1)` for Google Artifact Registry (GAR)*

Get the total size of images in a Google Artifact Registry docker repository.

## Install

`go install github.com/joemiller/dugar@latest`

## Usage

Requires Application Default Credentials. If running localy with `gcloud` you can run `gcloud auth application-default login` to setup local ADC.

example:

```console
$ dugar -project my-project -location us -repo my-repo

Analyzing projects/my-project/locations/us/repositories/my-repo ...
6.30 GiB        us-docker.pkg.dev/my-project/my-repo/db-runtime
48.81 GiB       us-docker.pkg.dev/my-project/my-repo/web-app
0.03 GiB        us-docker.pkg.dev/my-project/my-repo/another-app
0.09 GiB        us-docker.pkg.dev/my-project/my-repo/worker-foo
55.23 GiB       .
```

Pipe through `sort(1)` to sort by size:

```console
$ dugar -project my-project -location us -repo my-repo | sort -h
Analyzing projects/my-project/locations/us/repositories/my-repo ...
48.81 GiB       us-docker.pkg.dev/my-project/my-repo/web-app
6.30 GiB        us-docker.pkg.dev/my-project/my-repo/db-runtime
0.09 GiB        us-docker.pkg.dev/my-project/my-repo/worker-foo
0.03 GiB        us-docker.pkg.dev/my-project/my-repo/another-app
55.23 GiB       .


The `-format` flag can be used to change the output format. The default is `GiB`.

Use `-h` for help.

Images and tags can be explicitly included or excluded with `--include-image`, `--exclude-image`, `--include-tag`, and `--exclude-tag` flags. The arguments must be valid RE2 regexes.
