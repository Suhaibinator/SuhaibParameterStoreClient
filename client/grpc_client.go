package client

import (
	"context"
	"fmt"
	"log"

	pb "github.com/Suhaibinator/SuhaibParameterStoreClient/proto"

	"google.golang.org/grpc"
)

func GrpcimpleRetrieve(ServerAddress string, AuthenticationPassword string, key string) (val string, err error) {
	// Dial to the server
	conn, err := grpc.Dial(ServerAddress, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	// Create a new client
	client := pb.NewParameterStoreClient(conn)

	// Retrieve the stored value
	retrieveReq := &pb.RetrieveRequest{
		Key:      key,
		Password: AuthenticationPassword,
	}
	retrieveResp, err := client.Retrieve(context.Background(), retrieveReq)
	if err != nil {
		log.Printf("could not retrieve value: %v", err)
	}
	return retrieveResp.GetValue(), err
}

func GrpcSimpleStore(ServerAddress string, AuthenticationPassword string, key string, value string) (err error) {
	// Dial to the server
	conn, err := grpc.Dial(ServerAddress, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	// Create a new client
	client := pb.NewParameterStoreClient(conn)

	// Store a value
	storeReq := &pb.StoreRequest{
		Key:      key,
		Value:    value,
		Password: AuthenticationPassword,
	}
	storeResp, err := client.Store(context.Background(), storeReq)
	if err != nil {
		log.Printf("could not store value: %v", err)
	}
	fmt.Println("Store response:", storeResp.GetMessage())
	return err
}
