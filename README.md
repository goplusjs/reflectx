# reflectx
Golang reflect package hack tools

***

[![Go1.14](https://github.com/goplusjs/reflectx/workflows/Go1.14/badge.svg)](https://github.com/goplusjs/reflectx/actions?query=workflow%3AGo1.14)
[![Go1.15](https://github.com/goplusjs/reflectx/workflows/Go1.15/badge.svg)](https://github.com/goplusjs/reflectx/actions?query=workflow%3AGo1.15)


**reflectx.CanSet**
```
type Point struct {
    x int
    y int
}

x := &Point{10, 20}
v := reflect.ValueOf(x).Elem()
sf := v.Field(0)

fmt.Println(sf.CanSet()) // output: false
// sf.SetInt(102)        // panic

sf = reflectx.CanSet(sf)
fmt.Println(sf.CanSet()) // output: true

sf.SetInt(102)           // x.x = 102
fmt.Println(x.x)         // output: 102
```

**reflectx.StructOf**
```
support more embedded field
```
