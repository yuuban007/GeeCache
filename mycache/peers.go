package mycache

import pb "mycache/mycachepb"

type PeerPicker interface {
	PickPeer(key string) (peer PeerGetter, ok bool)
}

type PeerGetter interface {
	Get(group *pb.Request, key *pb.Response) error
}
