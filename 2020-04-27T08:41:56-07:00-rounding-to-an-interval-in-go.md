---
date: 2020-04-27T08:41:56-07:00
title: Rounding To An Interval In Go
---

# Rounding To An Interval In Go

```go
// roundTo defines the value to round all prices to the nearest of
// For intance, roundTo: 50
//		487 -> 500
//		472 -> 450
//
func (price *Price) roundTo(val, roundVal float64) float64 {
	return math.Round(val/roundVal) * roundVal
}
```
