# mpool
[![License][license-img]][license] [![Actions Status][action-img]][action] [![GoDoc][godoc-img]][godoc] [![Go Report Card][goreport-img]][goreport] [![Coverage Status][codecov-img]][codecov]

mpool is very basic implementation for controlled pool

For our projects we needed controlled pool (e.g. pool of connections)
with possibility to manage pool size and the way of reallocating and destroying items.

TODO: add more info

Example:
```
	...
	pool, err := NewPool(2, 5, func() *object {
		fmt.Println("Shared object allocated")
		return &object{}
	}, func(i *object) {
		fmt.Println("Shared released")
	}, func(i *object) bool {
		fmt.Println("Validate shared object")
		return true
	})

	if err!=nil {
		panic("Error during pool creating:"+err)
	}

	for j := 1; j < 10; j++ {
		go func() {
			if v1,ok := pool.Get(); ok {
				defer pool.Put(v1)
				v1.print()
			}
		}()
	}
	...
```

## Development Status: In active development
All APIs are in active development and not finalized, and breaking changes will be made in the 0.x series.


[license-img]: https://img.shields.io/badge/license-MIT-blue.svg
[license]: https://github.com/mmelnyk/mpool/blob/master/LICENSE
[action-img]: https://github.com/mmelnyk/mpool/workflows/Test/badge.svg
[action]: https://github.com/mmelnyk/mpool/actions
[godoc-img]: https://godoc.org/go.melnyk.org/mpool?status.svg
[godoc]: https://godoc.org/go.melnyk.org/mpool
[goreport-img]: https://goreportcard.com/badge/go.melnyk.org/mpool
[goreport]: https://goreportcard.com/report/go.melnyk.org/mpool
[codecov-img]: https://codecov.io/gh/mmelnyk/mpool/branch/master/graph/badge.svg
[codecov]: https://codecov.io/gh/mmelnyk/mpool
