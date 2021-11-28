package main_test

import (
	"bufio"
	"context"
	"fmt"
	"grpc-course/filetransfer/protobuf"
	"io"
	"net"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
)

type testserver struct {
	protobuf.UnimplementedFileServiceServer
}

func TestLeastWordCount(t *testing.T) {
	serverAddress := startTestFileServer(t)
	filestorage := newTestFileClient(t, serverAddress)
	req := &protobuf.LeastWordCountRequest{}
	_, err := filestorage.LeastWordCount(context.Background(), req)
	require.NoError(t, err)
}

func TestDeleteFile(t *testing.T) {
	serverAddress := startTestFileServer(t)
	filestorage := newTestFileClient(t, serverAddress)
	req := &protobuf.LeastWordCountRequest{}
	_, err := filestorage.LeastWordCount(context.Background(), req)
	require.NoError(t, err)
}
func TestWordCount(t *testing.T) {
	serverAddress := startTestFileServer(t)
	filestorage := newTestFileClient(t, serverAddress)
	req := &protobuf.WordCountRequest{}
	_, err := filestorage.WordCount(context.Background(), req)
	require.NoError(t, err)

}

func TestFileList(t *testing.T) {
	serverAddress := startTestFileServer(t)
	filestorage := newTestFileClient(t, serverAddress)
	req := &protobuf.ListFileRequest{}
	_, err := filestorage.GetFiles(context.Background(), req)
	require.NoError(t, err)
}

func TestFileUpdate(t *testing.T) {
	testFileFolder := "../../tmp"
	serverAddress := startTestFileServer(t)
	filestorage := newTestFileClient(t, serverAddress)
	filePath := fmt.Sprintf("%s/test.txt", testFileFolder)
	file, _ := os.Open(filePath)
	defer file.Close()
	stream, err := filestorage.UpdateFile(context.Background())
	require.NoError(t, err)
	req := &protobuf.UpdateFileRequest{
		Data: &protobuf.UpdateFileRequest_FileName{
			FileName: "test.txt",
		},
	}
	err = stream.Send(req)
	require.NoError(t, err)
	reader := bufio.NewReader(file)
	buffer := make([]byte, 1024)
	for {
		n, err := reader.Read(buffer)
		if err == io.EOF {
			break
		}

		require.NoError(t, err)

		req := &protobuf.UpdateFileRequest{
			Data: &protobuf.UpdateFileRequest_ChunkData{
				ChunkData: buffer[:n],
			},
		}

		err = stream.Send(req)
		require.NoError(t, err)
	}

	_, err = stream.CloseAndRecv()
	require.NoError(t, err)

	savedFilePath := fmt.Sprintf("%s/%s", testFileFolder, "test.txt")
	require.FileExists(t, savedFilePath)
	require.NoError(t, os.Remove(savedFilePath))
}

func TestFileUpload(t *testing.T) {
	t.Parallel()
	testFileFolder := "../../tmp"
	serverAddress := startTestFileServer(t)
	filestorage := newTestFileClient(t, serverAddress)
	filePath := fmt.Sprintf("%s/test.txt", testFileFolder)
	file, _ := os.Open(filePath)
	defer file.Close()
	stream, err := filestorage.UploadFile(context.Background())
	require.NoError(t, err)
	req := &protobuf.UploadFileRequest{
		Data: &protobuf.UploadFileRequest_FileName{
			FileName: "test.txt",
		},
	}
	err = stream.Send(req)
	require.NoError(t, err)
	reader := bufio.NewReader(file)
	buffer := make([]byte, 1024)
	for {
		n, err := reader.Read(buffer)
		if err == io.EOF {
			break
		}

		require.NoError(t, err)

		req := &protobuf.UploadFileRequest{
			Data: &protobuf.UploadFileRequest_ChunkData{
				ChunkData: buffer[:n],
			},
		}

		err = stream.Send(req)
		require.NoError(t, err)
	}

	_, err = stream.CloseAndRecv()
	require.NoError(t, err)

	savedFilePath := fmt.Sprintf("%s/%s", testFileFolder, "test.txt")
	require.FileExists(t, savedFilePath)
	require.NoError(t, os.Remove(savedFilePath))
}

func startTestFileServer(t *testing.T) string {

	grpcServer := grpc.NewServer()
	protobuf.RegisterFileServiceServer(grpcServer, testserver{})

	listener, err := net.Listen("tcp", ":0") // random available port
	require.NoError(t, err)

	go grpcServer.Serve(listener)

	return listener.Addr().String()
}

func newTestFileClient(t *testing.T, serverAddress string) protobuf.FileServiceClient {
	conn, err := grpc.Dial(serverAddress, grpc.WithInsecure())
	require.NoError(t, err)
	return protobuf.NewFileServiceClient(conn)
}
