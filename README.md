Sematext Hook for Logrus
==

Installing
---

in your go.mod add:
```go.mod
require (
    github.com/fino-digital/sematextHook latest
    github.com/sirupsen/logrus v1.4.2
)
```

and execute `go list -m all`

now go ahead and configure your logger somewhat like this:

```go
package main

import (
    "os"
    
    "github.com/fino-digital/sematextHook"
    "github.com/go-resty/resty/v2"
    "github.com/sirupsen/logrus"
)

func main()  {    

    hook, err := sematextHook.NewSematextHook(
        resty.New(),
        "https://logsene-receiver.sematext.com/<SEMATEXT LOG APP TOKEN>/",
        "product",
        "service",
        "development",
    )
    if err != nil {
        logrus.WithError(err).Error("unable to initialize sematext hook")
    } else {
        // translate log levels to strings when sending messages 
        hook.WithLevelMapper(sematextHook.AsLogbackLevel)
        logrus.AddHook(hook)
    }

}

```
Parameters explained:

**sematextHook.NewSematextHook**
 * `client: *resty.Client` to use when send out log messages e.g. `resty.New()`,
 * `baseUrl: string`. The baseUrl from the integrations section in sematext, without the logsene_type argument. e.g. `"https://logsene-receiver.sematext.com/<SEMATEXT LOG APP TOKEN>/"`,
 * `group: string`. The logsene_type url param to use, e.g. `supershop`
 * `facility: string` the service (as in application) this logger is inside, e.g. `api`
 * `environment: string` Something like "development", "staging", "production"
 
 Use **sematextHook.WithLevelMapper()** to select the textual representation of log levels.
 This library comes with two iplementations: 
  * sematextHook.AsLogrusLevel (lower case, default logrus behavior)
  * sematextHook.AsLogbackLevel (upper case, default logback behavior)