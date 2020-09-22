# ifchanged

Package `ifchanged` is a collection of Go functions to perform callbacks in case of file changes (using sha256 hash)
and / or missing files.

Use case: Generating `css` file only if `scss` file has changed or `css` file is missing.

## Example

```go
err = ifchanged.NewIf().
    Changed(fileName, fileName+".sha256").
    Missing("somefile.txt").
    Execute(func() error {
        fmt.Printf("This has been called because \"somefile.txt\" is missing or %v has changed\n", fileName)
        return nil
    })
```

## Doc

[pkg.go.dev/github.com/zzwx/ifchanged](https://pkg.go.dev/github.com/zzwx/ifchanged)

## Dependencies

* go 1.13

## `DB` 

`ifchanged`, additionally, has a way to provide checksums using a `DB` interface.

* `linetuple.go` contains a simple file-based implementation of `DB` (not really optimised yet)
* [github.com/prologic/bitcask](github.com/prologic/bitcask) implements it.
