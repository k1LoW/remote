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
