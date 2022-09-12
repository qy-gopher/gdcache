package gdcache

import "github.com/qy-gopher/gdcache/pb"

// PeerPicker 节点选择接口
type PeerPicker interface {
	PickPeer(key string) (peer PeerGetter, ok bool)
}

// PeerGetter 节点接口
type PeerGetter interface {
	// Get(group string, key string) ([]byte, error)
	Get(in *pb.Request, out *pb.Response) error
}
