package main

import (
	"fmt"
	"io"
	"os"

	trpc "trpc.group/trpc-go/trpc-go"
	pb "trpc.group/trpc-go/trpc-go/examples/features/stream/proto"
	"trpc.group/trpc-go/trpc-go/log"

)

func main() {
	s := trpc.NewServer()
	impl := &testStreamImpl{}
	pb.RegisterTestStreamService(s.Service("trpc.examples.stream.TestStream"), impl)

	if err := s.Serve(); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

type testStreamImpl struct {
	pb.UnimplementedTestStream
}

func (s *testStreamImpl) UploadFileStream(stream pb.TestStream_UploadFileStreamServer) error {
	var filename string
	var file *os.File

	defer func() {
		if file != nil {
			file.Close()
		}
	}()

	for {
		chunk, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to receive chunk: %w", err)
		}

		if file == nil {
			filename = chunk.GetFilename()
			file, err = os.Create(filename)
			if err != nil {
				return fmt.Errorf("failed to create file: %w", err)
			}
			defer file.Close()
		}

		if _, err = file.Write(chunk.GetContent()); err != nil {
			return fmt.Errorf("failed to write to file: %w", err)
		}
	}

	return stream.SendAndClose(&pb.UploadFileResp{
		Success: true,
		Message: fmt.Sprintf("File %s uploaded successfully", filename),
	})
}

func (s *testStreamImpl) DownloadFileStream(req *pb.DownloadFileReq, stream pb.TestStream_DownloadFileStreamServer) error {
	filePath := req.GetFilename()
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	buffer := make([]byte, 1024)
	for {
		n, err := file.Read(buffer)
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to read file: %w", err)
		}

		if err := stream.Send(&pb.DownloadFileResp{
			Content: buffer[:n],
		}); err != nil {
			return fmt.Errorf("failed to send file chunk: %w", err)
		}
	}

	return nil
}
