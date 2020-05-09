---
date: 2020-05-07T10:14:06-07:00
title: IP Address Int to Dotted Quad
tags: go
---

# IP Address Int To Dotted Quad

```go
package main

import (
	"fmt"
)

func main() {
	a := 3628584131

	fmt.Printf("%d.%d.%d.%d\n", byte(a>>24), byte(a>>16), byte(a>>8), byte(a))
}

>> 216.71.204.195
```
