package main

import (
	"context"
	"log"
	"net"

	"google.golang.org/grpc"

	recruitmentv1 "recruitment/api/gen/go/recruitment/v1"
	"recruitment/logic-grpc-service/internal/config"
	"recruitment/logic-grpc-service/internal/db"
	"recruitment/logic-grpc-service/internal/osssvc"
	"recruitment/logic-grpc-service/internal/service"
)

func main() {
	cfg := config.Load()
	gdb, err := db.Open(cfg.MySQLDSN)
	if err != nil {
		log.Fatalf("db: %v", err)
	}
	var ossClient *osssvc.Client
	if cfg.OSSBucket != "" {
		ossClient, err = osssvc.New(context.Background(), cfg)
		if err != nil {
			log.Printf("warn: oss init: %v (resume APIs will fail until OSS is configured)", err)
		}
	}
	svc, err := service.NewRecruitment(gdb, ossClient, cfg)
	if err != nil {
		log.Fatalf("service: %v", err)
	}
	lis, err := net.Listen("tcp", cfg.GRPCAddr)
	if err != nil {
		log.Fatalf("listen: %v", err)
	}
	s := grpc.NewServer()
	recruitmentv1.RegisterRecruitmentServiceServer(s, svc)
	log.Printf("logic-grpc-service listening on %s", cfg.GRPCAddr)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("serve: %v", err)
	}
}
