# xerr - error utilities library

## Installation

```bash
go get github.com/rprtr258/xerr
```

## Example usage
```go
err := xerr.New(
    xerr.Message("user not found"),
    xerr.Errors(sql.NoRows),
    xerr.Value(404),
    xerr.Field("trace_id", "abcabcabcabc"),
    xerr.Fields(map[string]any{
        "user_id": 1234,
        "user_page": "/posts",
    }),
)
```

Trace id can be added as option from context in following way:
```go
func WithTrace(ctx) xerr.Option {
    return WithFields(map[string]any{"trace_id": fromCtx(ctx)})
}
```

## Why
Multiple libraries are around in go to help handling errors. But each one of them tackles only one task, e.g. [adding caller metadata, adding stack metadata](https://github.com/ztrue/tracerr), [formatting error messages](https://pkg.go.dev/fmt#Errorf), [wrapping error](https://github.com/pkg/errors), [wrapping multiple errors](https://go.uber.org/multierr), [structured errors](https://github.com/Southclaws/fault).

None of them is widely used to support all listed features and/or so is not supported actively. So I wrote all utils funcs I want to use while doing error handling making this lib.

# THIS IS MY PERSONAL LIBRARY, EXPECT BREAKING CHANGES AND NO STABILITY. NOT READY AND WON'T BE READY FOR ANY USAGE BESIDES MY PERSONAL PROJECTS.
