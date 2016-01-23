```
go get github.com/Masterminds/glide
go install github.com/Masterminds/glide
glide install
GO15VENDOREXPERIMENT=1 go run app.go -staticpath=$PWD/../static
```

or

```
go get github.com/codegangsta/gin
GO15VENDOREXPERIMENT=1 gin -p 8001 -a 8000
```
