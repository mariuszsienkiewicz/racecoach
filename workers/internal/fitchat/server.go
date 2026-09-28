package fitchat

import (
	"context"
	"crypto/subtle"
	"errors"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/racecoach/workers/internal/config"
	"github.com/racecoach/workers/internal/llm"
	"github.com/racecoach/workers/internal/storage"
)

func Serve(ctx context.Context, cfg config.Config) error {
	log.Printf("checking config")
	if cfg.ChatServiceToken == "" {
		return errors.New("CHAT_SERVICE_TOKEN is required")
	}
	log.Printf("config checked")
	log.Println("starting fit chat server")
	mux := http.NewServeMux()

	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	storageClient := storage.NewClient(cfg.S3Region, cfg.S3AccessKey, cfg.S3SecretKey, cfg.S3Endpoint, cfg.S3UsePathStyle)
	llmClient := llm.NewClient(cfg.LLMBaseURL, cfg.LLMModel, cfg.LLMAPIKey)

	mux.HandleFunc("/v1/chat", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		if err := authenticateChatRequest(r, cfg); err != nil {
			log.Printf("failed to authenticate chat request: %v", err)
			writeJSONError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		if err := handleChat(r.Context(), w, r, storageClient, llmClient); err != nil {
			log.Printf("failed to handle chat: %v", err)
			writeJSONError(w, http.StatusServiceUnavailable, "failed to handle chat")
			return
		}
	})

	server := &http.Server{
		Addr:    cfg.ChatHTTPAddr,
		Handler: mux,
	}

	errChan := make(chan error, 1)
	go func() {
		err := server.ListenAndServe()
		errChan <- err
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		return server.Shutdown(shutdownCtx)
	case err := <-errChan:
		if errors.Is(err, http.ErrServerClosed) {
			log.Println("fit chat server closed gracefully")
			return nil
		}
		return err
	}
}

func authenticateChatRequest(r *http.Request, cfg config.Config) error {
	const prefix = "Bearer "
	auth := r.Header.Get("Authorization")
	if !strings.HasPrefix(auth, prefix) {
		return errors.New("invalid Authorization header")
	}

	token := strings.TrimSpace(strings.TrimPrefix(auth, prefix))
	if token == "" {
		return errors.New("Authorization header is required")
	}

	if subtle.ConstantTimeCompare([]byte(token), []byte(cfg.ChatServiceToken)) != 1 {
		return errors.New("invalid Authorization header")
	}

	return nil
}
