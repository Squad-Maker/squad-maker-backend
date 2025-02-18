package squad

import (
	pb "squad-maker/generated/squad"
)

type SquadServiceServer struct {
	pb.UnimplementedSquadServiceServer
}

func NewSquadServiceServer() *SquadServiceServer {
	return &SquadServiceServer{}
}
