package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/go-openapi/runtime/middleware"
	gohandlers "github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/iasonliu/env"
	protos "github.com/iasonliu/golang-microservices/currency/protos/currency"
	"github.com/iasonliu/golang-microservices/product-api/data"
	"github.com/iasonliu/golang-microservices/product-api/handlers"
	"google.golang.org/grpc"
)

var bindAddress = env.String("BIND_ADDRESS", false, ":9090", "Bind address for the server")
var currencyServiceAddress = env.String("CURRENCY_SERVICE_ADDRESS", false, "localhost:9092", "currency service address")

func main() {
	env.Parse()

	logger := log.New(os.Stdout, "[product-api] ", log.LstdFlags)
	v := data.NewValidation()

	conn, err := grpc.Dial("localhost:9092", grpc.WithInsecure())
	if err != nil {
		panic(err)
	}

	defer conn.Close()

	// create client
	cc := protos.NewCurrencyClient(conn)

	// create the handlers
	ph := handlers.NewProducts(l, v, cc)

	// create a new server mux and register hanndlers

	sm := mux.NewRouter()
	getR := sm.Methods(http.MethodGet).Subrouter()
	getR.HandleFunc("/products", ph.ListAll)
	getR.HandleFunc("/products/{id:[0-9]+}", ph.ListSingle)

	putR := sm.Methods(http.MethodPut).Subrouter()
	putR.HandleFunc("/products", ph.Update)
	putR.Use(ph.MiddlewareValidateProduct)

	postR := sm.Methods(http.MethodPost).Subrouter()
	postR.HandleFunc("/products", ph.Create)
	postR.Use(ph.MiddlewareValidateProduct)

	deleteR := sm.Methods(http.MethodDelete).Subrouter()
	deleteR.HandleFunc("/products/{id:[0-9]+}", ph.Delete)

	// * handler for documentation
	opts := middleware.RedocOpts{SpecURL: "/swagger.yaml"}
	sh := middleware.Redoc(opts, nil)
	getR.Handle("/docs", sh)
	getR.Handle("/swagger.yaml", http.FileServer(http.Dir("./")))

	// CORS
	// https://pkg.go.dev/github.com/gorilla/handlers
	// AllowedOrigins([]string{"http://localhost:3000"})
	// "*" will allow everywhere access
	ch := gohandlers.CORS(gohandlers.AllowedOrigins([]string{"*"}))

	// config http server
	httpServer := &http.Server{
		Addr:         *bindAddress,
		Handler:      ch(sm),
		IdleTimeout:  120 * time.Second,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	// start the server
	go func() {
		logger.Println("Starting server Listen on", httpServer.Addr)
		err := httpServer.ListenAndServe()
		if err != nil {
			logger.Fatalln(err)
		}
	}()

	// trap sigterm or interupt and gracefully shotdown the http server
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)
	signal.Notify(sigChan, os.Kill)
	sig := <-sigChan
	logger.Println("Recieved terminate, graceful shutdown", sig)

	timeoutCtx, err := context.WithTimeout(context.Background(), 30*time.Second)
	if err != nil {
		logger.Println(err)
	}
	logger.Fatal(httpServer.Shutdown(timeoutCtx))
}
