# remote

`remote` is a package for opening remote files easily.

## Support remote source

### HTTP

``` go
f, _ := remote.Open("http://example.com/file.txt")
defer f.Close()
```

### GitHub

``` go
f, _, := remote.Open("github://k1LoW/tbls/README.md")
defer f.Close()
```

### Amazon S3

``` go
f, _, := remote.Open("s3://my-bucket/path/to/file.txt")
defer f.Close()
```

### Google Cloud Storage

``` go
f, _, := remote.Open("gs://my-bucket/path/to/file.txt")
defer f.Close()
```

``` go
f, _, := remote.Open("gcs://my-bucket/path/to/file.txt")
defer f.Close()
```
