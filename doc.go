/*
mpool is very basic implementation for controlled pool

For our projects we needed controlled pool (e.g. pool of connections)
with possibility to manage pool size and the way of reallocating and destroying items.

TODO: add more info

Example:
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

*/
package mpool
