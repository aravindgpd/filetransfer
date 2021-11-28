package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/aravindgpd/filetransfer/protobuf"

	"google.golang.org/grpc"
)

func main() {

	operation := os.Args[1]
	fileName := ""
	server := "0.0.0.0:8080"
	if os.Getenv("SERVER_ADDRESS") != "" {
		server = os.Getenv("SERVER_ADDRESS")
	}
	conn, err := grpc.Dial(server, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("unable to connect to server: %v", err)
	}
	defer conn.Close()
	clientConnect := protobuf.NewFileServiceClient(conn)

	if operation == "add" {
		if len(os.Args) >= 4 {
			log.Fatalf(" Only two files can be added in the single command")
		}
		filelist := os.Args[2:len(os.Args)]
		fileUpload(clientConnect, filelist)
	} else if operation == "ls" {
		getFileList(clientConnect)
	} else if operation == "rm" {
		fileName = os.Args[2]
		deleteFile(clientConnect, fileName)
	} else if operation == "update" {
		fileName = os.Args[2]
		updateFile(clientConnect, fileName)
	} else if operation == "wc" {
		wordCount(clientConnect)
	} else if operation == "freq-words" {
		leastWordCount(clientConnect)
	} else {
		fmt.Println(" Allowed operations are 'add','ls','update','wc,'freq-words' ")
	}
}

// 10 Most frequent words occurred in all files
func leastWordCount(con protobuf.FileServiceClient) {
	req := &protobuf.LeastWordCountRequest{}
	res, err := con.LeastWordCount(context.Background(), req)
	if err != nil {
		log.Fatal("Error while counting leasr number of word in the file: ", err)
	}
	for _, value := range res.Result {
		fmt.Println(value)
	}
}

// sum of total  words available in the all files
func wordCount(con protobuf.FileServiceClient) {
	req := &protobuf.WordCountRequest{}
	res, err := con.WordCount(context.Background(), req)
	if err != nil {
		log.Fatal("Error while counting number word in the file: ", err)
	}
	fmt.Print(res.GetResult())

}

// update the existing file in the storage folder
// if file not present create new
func updateFile(con protobuf.FileServiceClient, fileName string) {
	file, err := os.Open(fileName)
	if err != nil {
		log.Fatal("cannot open file: ", err)
	}
	defer file.Close()
	req := &protobuf.UpdateFileRequest{
		Data: &protobuf.UpdateFileRequest_FileName{
			FileName: fileName,
		},
	}
	stream, err := con.UpdateFile(context.Background())
	if err != nil {
		log.Fatal("cannot upload file: ", err)
	}

	err = stream.Send(req)
	if err != nil {
		log.Fatal("cannot send image info to server: ", err, stream.RecvMsg(nil))
	}

	reader := bufio.NewReader(file)
	buffer := make([]byte, 1024)

	for {
		n, err := reader.Read(buffer)
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal("cannot read chunk to buffer: ", err)
		}

		req := &protobuf.UpdateFileRequest{
			Data: &protobuf.UpdateFileRequest_ChunkData{
				ChunkData: buffer[:n],
			},
		}

		err = stream.Send(req)
		if err != nil {
			log.Fatal("cannot send chunk to server: ", err, stream.RecvMsg(nil))
		}
	}

	res, err := stream.CloseAndRecv()
	if err != nil {
		log.Fatal("cannot receive response: ", err)
	}

	log.Printf("file %v upload is successful ", res)

}

// Deletion of file present in the storage folder
func deleteFile(con protobuf.FileServiceClient, fileName string) {
	req := &protobuf.DeleteFileRequest{
		FileName: fileName,
	}
	res, err := con.DeleteFiles(context.Background(), req)
	if err != nil {
		log.Fatalf("unable to delete the file: %v", err)
	}
	fmt.Println(res)
}

// list of files available in the storage folder
func getFileList(con protobuf.FileServiceClient) {
	req := &protobuf.ListFileRequest{}
	res, err := con.GetFiles(context.Background(), req)
	if err != nil {
		log.Fatal("unable to retrieve the file list: ", err)
	}
	if len(res.Result) == 0 {
		fmt.Println("storage folder contains 0 files")
	}
	for _, value := range res.Result {
		fmt.Println(value)
	}
}

// uploading new files into the storage folder
// maximum of two files can be uploaded into the storage folder
func fileUpload(con protobuf.FileServiceClient, filelist []string) {
	for _, filename := range filelist {
		file, err := os.Open(filename)
		if err != nil {
			log.Fatal("cannot open file: ", err)
		}
		defer file.Close()
		req := &protobuf.UploadFileRequest{
			Data: &protobuf.UploadFileRequest_FileName{
				FileName: filename,
			},
		}
		stream, err := con.UploadFile(context.Background())
		if err != nil {
			log.Fatal("cannot upload file: ", err)
		}

		err = stream.Send(req)
		if err != nil {
			log.Fatal("cannot send image info to server: ", err, stream.RecvMsg(nil))
		}

		reader := bufio.NewReader(file)
		buffer := make([]byte, 1024)

		for {
			n, err := reader.Read(buffer)
			if err == io.EOF {
				break
			}
			if err != nil {
				log.Fatal("cannot read chunk to buffer: ", err)
			}

			req := &protobuf.UploadFileRequest{
				Data: &protobuf.UploadFileRequest_ChunkData{
					ChunkData: buffer[:n],
				},
			}

			err = stream.Send(req)
			if err != nil {
				log.Fatal("cannot send chunk to server: ", err, stream.RecvMsg(nil))
			}
		}

		res, err := stream.CloseAndRecv()
		if err != nil {
			log.Fatal("cannot receive response: ", err)
		}

		fmt.Printf("file %v upload is successful ", res)
	}
}
