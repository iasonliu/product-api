# 
# golang microservices
- [Gorilla](https://www.gorillatoolkit.org/)
- [Validator](https://pkg.go.dev/github.com/go-playground/validator?utm_source=godoc)
- [redoc](https://github.com/Redocly/redoc)
- [Swagger](https://goswagger.io/)

## Auto-generating HTTP clients from Swagger

```
$ mkdir sdk && cd sdk
$ swagger generate client -f ../swagger.yaml -A product-api
```