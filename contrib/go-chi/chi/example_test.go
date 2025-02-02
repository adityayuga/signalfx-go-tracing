package chi_test

import (
	"net/http"

	"github.com/go-chi/chi"

	chitrace "github.com/adityayuga/signalfx-go-tracing/contrib/go-chi/chi"
	"github.com/adityayuga/signalfx-go-tracing/ddtrace/tracer"
)

func handler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello World!\n"))
}

func Example() {
	// Start the tracer
	tracer.Start()
	defer tracer.Stop()

	// Create a chi Router
	router := chi.NewRouter()

	// Use the tracer middleware with the default service name "chi.router".
	router.Use(chitrace.Middleware())

	// Set up some endpoints.
	router.Get("/", handler)

	// And start gathering request traces
	http.ListenAndServe(":8080", router)
}

func Example_withServiceName() {
	// Start the tracer
	tracer.Start()
	defer tracer.Stop()

	// Create a chi Router
	router := chi.NewRouter()

	// Use the tracer middleware with your desired service name.
	router.Use(chitrace.Middleware(chitrace.WithServiceName("chi-server")))

	// Set up some endpoints.
	router.Get("/", handler)

	// And start gathering request traces
	http.ListenAndServe(":8080", router)
}
