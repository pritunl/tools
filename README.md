# pritunl/tools: Pritunl tools package

Common libraries used in Pritunl golang codebases

### Errors

```go
func test() (err error) {
    _, err = os.Open("test.json")
    if err != nil {
        err = &errortypes.ReadError{
            errors.Wrap(err, "tools: Failed to read file"),
        }
        return
    }
    return
}
```

### Commander

```go
ret, err := commander.Exec(&commander.Opt{
    Name: "cat",
    Args: []string{
        "--number",
    },
    Env: map[string]string{
        "TEST_VAR": "test",
    },
    Input:   "test_input",
    Timeout: 5 * time.Second,
    Dir:     "/",
    PipeOut: true,
    PipeErr: true,
})
if err != nil {
    return
}

println(ret.ExitCode)
println(string(ret.Output))
```

### Logger

```go
logger.Init(
    logger.SetMaxLimit(2*time.Hour),
    logger.SetIcons(true),
)

logger.AddHandler(func(record *logger.Record) {
    fmt.Print(record.StringColor())
})

logger.WithFields(logger.Fields{
    "string":  "test",
    "number":  1,
    "boolean": true,
}).Panic("tools: Test panic")

logger.WithFields(logger.Fields{
    "string":  "test",
    "number":  1,
    "boolean": true,
}).Crit("tools: Test crit")

logger.WithFields(logger.Fields{
    "string":  "test",
    "number":  1,
    "boolean": true,
}).Error("tools: Test error")

logger.WithFields(logger.Fields{
    "string":  "test",
    "number":  1,
    "boolean": true,
}).Warn("tools: Test warn")

logger.WithFields(logger.Fields{
    "string":  "test",
    "number":  1,
    "boolean": true,
}).Info("tools: Test info")

logger.WithFields(logger.Fields{
    "string":  "test",
    "number":  1,
    "boolean": true,
}).Debug("tools: Test debug")

logger.WithFields(logger.Fields{
    "string":  "test",
    "number":  1,
    "boolean": true,
}).Trace("tools: Test trace")

e := fmt.Errorf("test error")

testErr := &errortypes.ParseError{
    errors.Wrapf(e, "tools: Test parse error '%s'", "type"),
}

testErrData := &errortypes.ErrorData{
    Error:   "test_error_key",
    Message: "Test error data value",
}

logger.WithFields(logger.Fields{
    "string":  "test",
    "number":  1,
    "boolean": true,
    "error":   testErr,
}).Error("tools: Test error")

logger.WithFields(logger.Fields{
    "string":     "test",
    "number":     1,
    "boolean":    true,
    "error_data": testErrData,
}).Error("tools: Test error data")

logger.WithFields(logger.Fields{
    "string":     "test",
    "number":     1,
    "boolean":    true,
    "error":      testErr,
    "error_data": testErrData,
}).Error("tools: Test error and error data")

for i := 0; i < 12; i++ {
    logger.WithFields(logger.Fields{
        "test": true,
    }).Limit(1 * time.Second).Info("tools: Test limit")
    time.Sleep(100 * time.Millisecond)
}

logr := logger.New(
    logger.SetTimeFormat("[15:04:05]"),
    logger.SetMaxLimit(2*time.Hour),
    logger.SetIcons(false),
)

logr.AddHandler(func(record *logger.Record) {
    fmt.Print(record.String())
})

logr.WithFields(logger.Fields{
    "string":  "test",
    "number":  1,
    "boolean": true,
}).Info("tools: Test info")
```
