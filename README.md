[![Build Status](https://travis-ci.org/SotirisAlfonsos/gocache.svg?branch=main)](https://travis-ci.org/SotirisAlfonsos/gocache)
[![GoDoc](https://godoc.org/github.com/SotirisAlfonsos/gocache?status.png)](https://godoc.org/github.com/SotirisAlfonsos/gocache)
[![Go Report Card](https://goreportcard.com/badge/github.com/SotirisAlfonsos/gocache)](https://goreportcard.com/report/github.com/SotirisAlfonsos/gocache)
[![codebeat badge](https://codebeat.co/badges/d47cd5fb-cb6c-4eea-9f3c-c414655dbe3a)](https://codebeat.co/projects/github-com-sotirisalfonsos-gocache-main)
[![codecov](https://codecov.io/gh/SotirisAlfonsos/gocache/branch/main/graph/badge.svg?token=pOexX69rp4)](https://codecov.io/gh/SotirisAlfonsos/gocache)

# Go cache
An in memory flexible cache, with lazy eviction, where both key and value are interfaces 

## Usage
As a first step you need to initialise the cache
```go
c := gocache.New(0)
```
or with one minute expiration
```go
c := gocache.New(1 * time.Minute)
```
<br/><br/>

And then all you have to do is implement the Key interface for the key of your cache
```go
type Key interface {
	Equals(key Key) bool
}
```

Example interface implementation
```go
type Key struct {
	Value1 string
	Value2 string
}

func (k Key) Equals(key gocache.Key) bool {
	if key == nil {
		return false
	}

	return k.Value1 == key.(Key).Value1 && k.Value2 == key.(Key).Value2
}
```
