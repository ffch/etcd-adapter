package extend

import (
	"context"
)

type Extend interface {
	LeaseGrant(ctx context.Context, id string, ttl int64) (*GrantRes, error)
}

type GrantRes struct {
	ID             string         `json:"ID,omitempty"`
	TTL            int64          `json:"TTL,omitempty"`
	Error          string         `json:"error,omitempty"`
	GrantHeaderRes GrantHeaderRes `json:"header,omitempty"`
}

type GrantHeaderRes struct {
	ClusterId uint64 `protobuf:"varint,1,opt,name=cluster_id,json=clusterId,proto3" json:"cluster_id,omitempty"`
	MemberId  uint64 `protobuf:"varint,2,opt,name=member_id,json=memberId,proto3" json:"member_id,omitempty"`
	Revision  int64  `protobuf:"varint,3,opt,name=revision,proto3" json:"revision,omitempty"`
	RaftTerm  uint64 `protobuf:"varint,4,opt,name=raft_term,json=raftTerm,proto3" json:"raft_term,omitempty"`
}
