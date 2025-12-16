package storage

import (
	storagepb "github.com/Nikita-Smirnov-idk/storage-service/contracts/gen/go"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Storage struct {
	Client storagepb.StorageClient
	conn   *grpc.ClientConn
}

func NewStorageClient(addr string) (*Storage, error) {
	conn, err := grpc.NewClient(addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	client := storagepb.NewStorageClient(conn)

	return &Storage{
		Client: client,
		conn:   conn,
	}, nil
}

func (c *Storage) Close() {
	c.conn.Close()
}
