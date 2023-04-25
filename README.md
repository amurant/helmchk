# helmchk

> `helmchk` is a cli tool that can extract all the variables used in the templates of a
> Helm chart and compare them with the default values configured in the values.yaml file.

Install the latest version of helmchk with:

```console
$ go install github.com/amurant/helmchk
```

Run helmchk on a chart directory:

```console
$ helm pull jetstack/cert-manager --untar
$ helmchk ./cert-manager/
value missing from values.yaml: .$.acmesolver.image.tag
value missing from values.yaml: .$.automountServiceAccountToken
...
```

You can also learn what the allowed exceptions are by running helmchk with the `--exceptions` flag:

```console
# Learn what the allowed exceptions are
$ helmchk ./my-chart/ > exceptions.txt

# Run helmchk and ignore the exceptions
$ helmchk ./my-chart/ --exceptions exceptions.txt
```
