package rafTEE

import (
	"context"
	"google.golang.org/grpc"
	"log"
	"net"
)

type RafTEErpcService struct {
	UnimplementedRafTEErpcServiceServer
}

func (s *RafTEErpcService) RequestVoteRPC(ctx context.Context, in *VoteRequest) (*VoteReponse, error) {
	log.Println("Received: ", in.Body)
	return &VoteReponse{Body: "Hello From the RafTEErpcService!"}, nil
}

func (s *RafTEErpcService) AppendEntriesRPC(ctx context.Context, in *AppendEntriesRequest) (*AppendEntriesResponse, error) {
	log.Println("Received 2: ", in.Body)
	return &AppendEntriesResponse{Body: "Hello from the Server2!"}, nil
}

func StartServer() {
	lis, err := net.Listen("tcp", ":9000")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := RafTEErpcService{}
	grpcServer := grpc.NewServer()
	RegisterRafTEErpcServiceServer(grpcServer, &s)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
