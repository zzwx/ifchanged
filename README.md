# ifchanged

A collection of functions to perform callbacks in case a file has changed (using sha256 hash) and (optionally) destination result is missing.

Use case: Generating `css` file only if `scss` file has changed or `css` file is missing.

## Dependencies

* go 1.13
* github.com/prologic/bitcask - for key/value db based sha256 