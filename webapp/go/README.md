```
go get github.com/Masterminds/glide
go install github.com/Masterminds/glide
glide install
GO15VENDOREXPERIMENT=1 go run app.go -staticpath=$PWD/../static
```
