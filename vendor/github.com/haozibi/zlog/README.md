![logo](logo.png)


# zlog

Just Log Basic On [zerolog](https://github.com/rs/zerolog)

## Install

```shell
$ go get -u github.com/haozibi/zlog
```

## Demo

```go
package main

import (
	"os"

	"github.com/haozibi/zlog"
)

func init() {

	zlog.NewBasicLog(os.Stdout)
	// zlog.NewJSONLog(os.Stdout)
}

func main() {
	zlog.ZInfo().
		Int("z", 100-1).
		Msg("just do it")

	zlog.ZDebug().
		Float64("f", 3.1415926).
		Msgf("hello %s", "zlog")
}
```
