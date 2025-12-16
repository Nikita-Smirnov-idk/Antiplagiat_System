package storage

import (
	storagepb "github.com/Nikita-Smirnov-idk/storage-service/contracts/gen/go"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type StorageClient struct {
	client storagepb.StorageClient
	conn   *grpc.ClientConn
}

func NewStorageClient(addr string) (*StorageClient, error) {
	// Создаем соединение
	conn, err := grpc.Dial(addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	// Создаем клиента
	client := storagepb.NewStorageClient(conn)

	return &StorageClient{
		client: client,
		conn:   conn,
	}, nil
}

func (c *StorageClient) Close() {
	c.conn.Close()
}
