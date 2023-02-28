# xerr - error utilities library

## Installation

```bash
go get github.com/rprtr258/xerr
```

## Example usage
```go
err := xerr.New(
    xerr.WithMessage("user not found"),
    xerr.WithErrs(sql.NoRows),
    xerr.WithValue(404),
    xerr.WithField("trace_id", "abcabcabcabc"),
    xerr.WithFields(map[string]any{
        "user_id": 1234,
        "user_page": "/posts",
    }),
)
```

## Why
Multiple libraries are around in go to help handling errors. But each one of them tackles only one task, e.g. [adding caller metadata, adding stack metadata](https://github.com/ztrue/tracerr), [formatting error messages](https://pkg.go.dev/fmt#Errorf), [wrapping error](https://github.com/pkg/errors), [wrapping multiple errors](go.uber.org/multierr), [structured errors](https://github.com/Southclaws/fault).

None of them is widely used to support all listed features and/or so is not supported actively. So I wrote all utils funcs I want to use while doing error handling making this lib.

# THIS IS MY PERSONAL LIBRARY, EXPECT BREAKING CHANGES AND NO STABILITY. NOT READY AND WON'T BE READY FOR ANY USAGE BESIDES MY PERSONAL PROJECTS.
