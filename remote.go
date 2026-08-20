package main

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

// RemoteCheckPath is the single endpoint exposed by an instance running in remote mode.
const RemoteCheckPath = "/check"

var remoteClient = &http.Client{Timeout: 30 * time.Second}

// remoteEndpoint turns a config address such as "192.168.0.13:8080" into a full URL.
// The scheme defaults to http and the path defaults to RemoteCheckPath.
func remoteEndpoint(addr string) (string, error) {
	raw := strings.TrimSpace(addr)

	if raw == "" {
		return "", fmt.Errorf("the remote address is empty")
	}

	if !strings.Contains(raw, "://") {
		raw = "http://" + raw
	}

	parsed, err := url.Parse(raw)
	if err != nil {
		return "", fmt.Errorf("'%s' is not a valid remote address:\n%w", addr, err)
	}

	if parsed.Host == "" {
		return "", fmt.Errorf("'%s' is not a valid remote address: no host", addr)
	}

	if parsed.Path == "" || parsed.Path == "/" {
		parsed.Path = RemoteCheckPath
	}

	return parsed.String(), nil
}

// CallRemote asks an auto-suspend instance running in remote mode whether it agrees
// to suspend. It returns "yes" or "no". An error response code or any output other
// than "yes"/"no" is reported as an error so the step's on_error behavior applies.
func CallRemote(addr string) (string, error) {
	endpoint, err := remoteEndpoint(addr)
	if err != nil {
		return "", err
	}

	resp, err := remoteClient.Get(endpoint)
	if err != nil {
		return "", fmt.Errorf("failed to reach the remote at %s:\n%w", endpoint, err)
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read the response of the remote at %s:\n%w", endpoint, err)
	}

	output := strings.TrimSpace(string(body))

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("the remote at %s responded with '%s':\n%s", endpoint, resp.Status, output)
	}

	switch output {
	case "yes", "no":
		return output, nil
	default:
		return "", fmt.Errorf("the remote at %s returned a malformed response: '%s'", endpoint, output)
	}
}

// RunRemoteServer runs auto-suspend as a remote decision server. It never suspends
// this machine; it only evaluates the configured sequence on demand and reports the
// verdict as plain text ("yes" or "no").
func RunRemoteServer(cfg *Config, env *Environment, logger *DaemonLogger, addr string) error {
	mux := http.NewServeMux()

	// The sequence may run arbitrary scripts, so serialize concurrent requests.
	var mu sync.Mutex

	mux.HandleFunc("GET "+RemoteCheckPath, func(w http.ResponseWriter, r *http.Request) {
		logger.Log(fmt.Sprintf("[remote] GET %s from %s, running sequence", RemoteCheckPath, r.RemoteAddr))

		mu.Lock()
		suspend, seqErr := AutoRunSequence(cfg, env, logger)
		mu.Unlock()

		w.Header().Set("Content-Type", "text/plain; charset=utf-8")

		if seqErr != nil {
			seqErr.Print()

			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "%s - %s\n", seqErr.message, seqErr.details)

			return
		}

		answer := "no"
		if suspend {
			answer = "yes"
		}

		logger.Log(fmt.Sprintf("[remote] answering %s", logger.Bold(answer)))
		fmt.Fprintln(w, answer)
	})

	logger.BeginRemote(env, addr)

	return http.ListenAndServe(addr, mux)
}
