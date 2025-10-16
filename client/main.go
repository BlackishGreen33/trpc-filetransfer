package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"os"

	"trpc.group/trpc-go/trpc-go"
	"trpc.group/trpc-go/trpc-go/client"
	pb "trpc.group/trpc-go/trpc-go/examples/features/stream/proto"

)

func main() {
	if err := initializeTrpc(); err != nil {
		log.Fatalf("Initialization failed: %v", err)
	}

	fmt.Println("Choose an option:")
	fmt.Println("1. Upload a file")
	fmt.Println("2. Download a file")

	choice := getInput()

	switch choice[0] {
	case '1':
		callUploadFile()
	case '2':
		callDownloadFile()
	default:
		fmt.Println("Invalid choice")
	}
}

func initializeTrpc() error {
	cfg, err := trpc.LoadConfig(trpc.ServerConfigPath)
	if err != nil {
		return fmt.Errorf("load config failed: %w", err)
	}
	trpc.SetGlobalConfig(cfg)
	if err := trpc.Setup(cfg); err != nil {
		return fmt.Errorf("setup plugin failed: %w", err)
	}
	return nil
}

func getInput() string {
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		log.Fatalf("Failed to read input: %v", err)
	}
	return input[:len(input)-1]
}

func callUploadFile() {
	proxy := createClientProxy()
	fileName := "test.zip"

	file, err := os.Open("./" + fileName)
	if err != nil {
		log.Fatalf("Failed to open file: %v", err)
	}
	defer file.Close()

	stream, err := proxy.UploadFileStream(context.Background())
	if err != nil {
		log.Fatalf("Could not upload file: %v", err)
	}

	buffer := make([]byte, 1024)
	for {
		n, err := file.Read(buffer)
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("Failed to read file: %v", err)
		}

		err = stream.Send(&pb.UploadFileReq{
			Content:  buffer[:n],
			Filename: fileName,
		})
		if err != nil {
			log.Fatalf("Failed to send chunk: %v", err)
		}
	}

	status, err := stream.CloseAndRecv()
	if err != nil {
		log.Fatalf("Failed to receive status: %v", err)
	}

	log.Printf("Upload status: %v, message: %s", status.GetSuccess(), status.GetMessage())
}

func callDownloadFile() {
	proxy := createClientProxy()
	fileName := "test.zip"

	stream, err := proxy.DownloadFileStream(context.Background(), &pb.DownloadFileReq{
		Filename: fileName,
	})
	if err != nil {
		log.Fatalf("Could not initiate download: %v", err)
	}

	outFile, err := os.Create("./" + fileName)
	if err != nil {
		log.Fatalf("Failed to create file: %v", err)
	}
	defer outFile.Close()

	for {
		chunk, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("Failed to receive chunk: %v", err)
		}

		if _, err := outFile.Write(chunk.GetContent()); err != nil {
			log.Fatalf("Failed to write chunk: %v", err)
		}
	}

	log.Printf("Download completed: %s", fileName)
}

func createClientProxy() pb.TestStreamClientProxy {
	return pb.NewTestStreamClientProxy(
		client.WithTarget("ip://47.74.41.12:8010"),
		client.WithProtocol("trpc"),
	)
}
