## goa plugin

### Use [slogdriver](https://github.com/kitagry/slogdriver)

```go
package design

import (
	slogdriver "github.com/kitagry/goaplugin/slogdriver/dsl"
	. "goa.design/goa/v3/dsl"
)

...

var _ = Service("calc", func() {
    slogdriver.HealthCheckPaths("/healthz")
})
```
