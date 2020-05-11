---
date: 2020-05-11T10:42:26-07:00
title: Go Table Test Structure
tags: go
---

# Go Table Test Structure

Because I can never remember how to structure the `for` loop with
`t.Run` inside it, and have to look at other projects every single time:

```
func Test_NewFunction(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

		})
	}
}
```
