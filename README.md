# reflectx
Golang reflect package hack tools

***

**reflectx.CanSet**
```
type Point struct {
    x int
    y int
}

x := &Point{10, 20}
v := reflect.ValueOf(x).Elem()
sf := v.Field(0)
// sf.SetInt(102) // panic
sf = reflectx.CanSet(sf)
sf.SetInt(102) // x.x = 102
```
