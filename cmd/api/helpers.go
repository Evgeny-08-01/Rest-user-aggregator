package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"Rest-user-agregator/internal/cache"
	"Rest-user-agregator/internal/database"
	"Rest-user-agregator/internal/service"
	"Rest-user-agregator/pkg/logger"

	"google.golang.org/grpc"
)

// runServersWithContext — для тестов (завершается по ctx.Done())
//
//nolint:unused
func runServersWithContext(ctx context.Context, svc *service.SubscriptionService, authService *service.AuthService, templateSvc *service.TemplateService) error {
	restServer := createRESTServer(svc, authService)
	grpcServer, lis, err := createGRPCServer(svc, templateSvc)
	if err != nil {
		return fmt.Errorf("failed to start gRPC server: %w", err)
	}

	var wg sync.WaitGroup
	chErr := make(chan error, 2)

	// Запуск REST
	wg.Add(1)
	go func() {
		defer wg.Done()
		logger.Info("REST API server starting...")
		if err := restServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			chErr <- fmt.Errorf("REST API: %w", err)
		}
	}()

	// Запуск gRPC
	wg.Add(1)
	go func() {
		defer wg.Done()
		logger.Info("gRPC API server starting...")
		if err := grpcServer.Serve(lis); err != nil && err != grpc.ErrServerStopped {
			chErr <- fmt.Errorf("gRPC API: %w", err)
		}
	}()

	// Ждём сигнал ИЛИ отмену контекста
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-quit:
		logger.Info("Shutdown signal received")
	case <-ctx.Done():
		logger.Info("Context canceled, shutting down...")
	case err := <-chErr:
		logger.Error("Server error: %v", err)
	}

	// Graceful shutdown (как в runServers)
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	logger.Info("Stopping REST API...")
	if err := restServer.Shutdown(shutdownCtx); err != nil {
		logger.Error("REST shutdown error: %v", err)
	}

	logger.Info("Stopping gRPC API...")
	done := make(chan struct{})
	go func() {
		grpcServer.GracefulStop()
		close(done)
	}()

	select {
	case <-done:
		logger.Info("gRPC stopped gracefully")
	case <-shutdownCtx.Done():
		logger.Warn("gRPC timeout, forcing stop")
		grpcServer.Stop()
	}

	wg.Wait()
	logger.Info("All servers stopped")
	return nil
}

// runWithContext — для тестов (завершается по контексту, а не по сигналу)
//
//nolint:unused
func runWithContext(ctx context.Context) error {
	loadEnv()
	initLogger()
	initPprof()

	if err := initRedis(); err != nil {
		logger.Warn("Redis unavailable: %v", err)
	} else {
		// ✅ ЕСЛИ ОТКРЫЛСЯ — ЗАКРЫВАЕМ ПРИ ВЫХОДЕ
		defer func() {
			if err := cache.Close(); err != nil {
				logger.Warn("Redis close error: %v", err)
			}
		}()
	}

	if err := initDB(); err != nil {
		return fmt.Errorf("DB init: %w", err)
	}
	defer database.Close()

	if err := runMigrations(); err != nil {
		return fmt.Errorf("migrations: %w", err)
	}

	svc, authSvc, templateSvc := buildServices()
	return runServersWithContext(ctx, svc, authSvc, templateSvc)
}
