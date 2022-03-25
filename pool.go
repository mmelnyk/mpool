package mpool

import "errors"

type Pool[T any] interface {
	Get() (T, bool)
	Put(T)
}

var (
	ErrorInvalidParameters = errors.New("Invalid Parameters")
)
