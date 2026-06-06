package main

import (
	"fmt"
	"net/http"
)

func RunAPIServer(logger *DaemonLogger, env *Environment, addr string) error {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /suspend", func(w http.ResponseWriter, r *http.Request) {
		logger.Log("[API] POST /suspend received, suspending now")

		doSuspend(logger, env)
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "suspend initiated")
	})

	logger.Log(fmt.Sprintf("[API] Server listening on %s", addr))

	return http.ListenAndServe(addr, mux)
}
