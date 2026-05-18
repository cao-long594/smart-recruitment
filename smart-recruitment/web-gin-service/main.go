package main

import (
	"log"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	recruitmentv1 "recruitment/api/gen/go/recruitment/v1"
	"recruitment/web-gin-service/internal/config"
	"recruitment/web-gin-service/internal/handler"
)

func main() {
	cfg := config.Load()
	conn, err := handler.DialLogic(cfg.LogicGRPCAddr)
	if err != nil {
		log.Fatalf("grpc dial %s: %v", cfg.LogicGRPCAddr, err)
	}
	defer conn.Close()
	h := &handler.Handler{
		Cfg:    cfg,
		Client: recruitmentv1.NewRecruitmentServiceClient(conn),
	}
	r := gin.Default()
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173", "http://localhost:5174", "http://127.0.0.1:5173", "http://127.0.0.1:5174"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))
	handler.RegisterRoutes(r, h)
	log.Printf("web-gin-service listening on %s", cfg.HTTPAddr)
	if err := r.Run(cfg.HTTPAddr); err != nil {
		log.Fatal(err)
	}
}
