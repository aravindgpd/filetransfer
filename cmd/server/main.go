package main

import (
	"bytes"
	"context"
	"io"
	"io/ioutil"
	"log"
	"net"
	"os"
	"sort"
	"strings"

	"github.com/aravindgpd/filetransfer/protobuf"

	"github.com/aravindgpd/filetransfer/service"

	"google.golang.org/grpc"
)

type servertype struct {
	protobuf.UnimplementedFileServiceServer
}

type WordCount struct {
	key   string
	value int
}

type wordCountlist []WordCount

func (wc wordCountlist) Len() int           { return len(wc) }
func (wc wordCountlist) Swap(i, j int)      { wc[i], wc[j] = wc[j], wc[i] }
func (wc wordCountlist) Less(i, j int) bool { return wc[i].value < wc[j].value }

// freq-words function
func (servertype) LeastWordCount(context context.Context, req *protobuf.LeastWordCountRequest) (*protobuf.LeastWordCountResponse, error) {
	folderLocation := service.GetFolderLocation()
	words := make(map[string]int)
	files, err := service.ReadDirectory(folderLocation)
	if err != nil {
		log.Fatal(err)
	}
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		data, err := ioutil.ReadFile(folderLocation + "/" + file.Name())
		if err != nil {
			return &protobuf.LeastWordCountResponse{}, err
		}
		for _, word := range strings.Fields(string(data)) {
			words[word]++
		}
	}
	wcl := make(wordCountlist, len(words))
	i := 0
	for k, v := range words {
		wcl[i] = WordCount{k, v}
		i++
	}
	// sorting of map by values
	sort.Sort(wcl)
	results := []string{}
	counter := 0

	if len(wcl) <= 10 {
		counter = 0
	} else {
		counter = len(wcl) - 10
	}

	for index := len(wcl) - 1; index >= counter; index-- {
		results = append(results, wcl[index].key)
	}

	res := &protobuf.LeastWordCountResponse{
		Result: results,
	}
	return res, nil

}

// word count function
func (servertype) WordCount(Context context.Context, req *protobuf.WordCountRequest) (*protobuf.WordCountResponse, error) {
	wordCount := 0
	folderLocation := service.GetFolderLocation()
	files, err := service.ReadDirectory(folderLocation)
	if err != nil {
		log.Fatal(err)
	}
	for _, fileinfo := range files {
		if fileinfo.IsDir() {
			continue
		}
		for _, file := range files {
			if file.IsDir() {
				continue
			}
			data, err := ioutil.ReadFile(folderLocation + "/" + file.Name())
			if err != nil {
				return &protobuf.WordCountResponse{}, err
			}
			wordCount = len(strings.Fields(string(data)))
		}
	}

	res := &protobuf.WordCountResponse{
		Result: int64(wordCount),
	}

	return res, nil
}

// delete of file mentioned in request
func (servertype) DeleteFiles(Context context.Context, req *protobuf.DeleteFileRequest) (*protobuf.DeleteFileResponse, error) {
	fileName := req.GetFileName()
	folderLocation := service.GetFolderLocation()
	err := os.Remove(folderLocation + "/" + fileName)
	if err != nil {
		log.Print("unable to remove file", err)
		return &protobuf.DeleteFileResponse{}, err
	}
	res := &protobuf.DeleteFileResponse{
		Result: fileName + " Successfully deleted",
	}
	return res, nil
}

// list of files present in the folder

func (servertype) GetFiles(context context.Context, req *protobuf.ListFileRequest) (*protobuf.ListFileResponse, error) {
	result := []string{}
	folderLocation := service.GetFolderLocation()
	file, err := os.Open(folderLocation)
	if err != nil {
		log.Fatalf("failed opening directory: %s", err)
		return &protobuf.ListFileResponse{}, err
	}
	defer file.Close()

	list, _ := file.Readdirnames(0) // 0 to read all files and folders
	result = append(result, list...)
	res := &protobuf.ListFileResponse{
		Result: result,
	}
	return res, nil
}

// update / create of file in the storage folder

func (servertype) UpdateFile(stream protobuf.FileService_UpdateFileServer) error {
	req, err := stream.Recv()
	if err != nil {
		log.Print("cannot receive image info")
		return err
	}
	fileName := req.GetFileName()

	fileData := bytes.Buffer{}

	for {
		log.Print("waiting to receive more data")

		req, err := stream.Recv()
		if err == io.EOF {
			log.Print("no more data")
			break
		}
		if err != nil {
			log.Printf("cannot receive chunk data: %v", err)
			return err
		}

		chunk := req.GetChunkData()
		size := len(chunk)

		log.Printf("received a chunk with size: %d", size)
		_, err = fileData.Write(chunk)
		if err != nil {
			log.Printf("cannot write chunk data: %v", err)
			return err
		}
	}
	folderLocation := service.GetFolderLocation()
	err = service.CheckFolderExists(folderLocation)
	if err != nil {
		return err
	}
	err = service.CheckFileExists(folderLocation, fileName)
	if err != nil {
		return err
	}

	// writing to the file
	file, err := os.OpenFile(folderLocation+"/"+fileName, os.O_RDWR, 0644)

	if err != nil {
		log.Printf("failed opening file: %s", err)
		return err
	}
	defer file.Close()

	_, err = file.WriteAt(fileData.Bytes(), 0) // Write at 0 beginning
	if err != nil {
		log.Printf("failed writing to file: %s", err)
		return err
	}
	res := &protobuf.UpdateFileResponse{
		Result: fileName,
	}
	if err = stream.SendAndClose(res); err != nil {
		log.Printf("cannot send response : %v", err)
		return err
	}
	return nil
}

// create of file in the storage location
func (servertype) UploadFile(stream protobuf.FileService_UploadFileServer) error {
	req, err := stream.Recv()
	if err != nil {
		log.Print("cannot receive data info")
		return err
	}
	fileName := req.GetFileName()
	fileData := bytes.Buffer{}

	for {
		log.Print("waiting to receive more data")

		req, err := stream.Recv()
		if err == io.EOF {
			log.Print("no more data")
			break
		}
		if err != nil {
			log.Printf("cannot receive chunk data: %v", err)
			return err
		}

		chunk := req.GetChunkData()
		size := len(chunk)

		log.Printf("received a chunk with size: %d", size)
		_, err = fileData.Write(chunk)
		if err != nil {
			log.Printf("cannot write chunk data: %v", err)
			return err
		}
	}
	folderLocation := service.GetFolderLocation()
	_, err = os.Stat(folderLocation)
	if os.IsNotExist(err) {
		errDir := os.MkdirAll(fileName, 0644)
		if errDir != nil {
			return err
		}
	}
	file, err := os.Create(folderLocation + "/" + fileName)
	if err != nil {
		log.Printf("cannot create file : %v", err)
		return err
	}

	_, err = fileData.WriteTo(file)
	if err != nil {
		log.Printf("cannot create file : %v", err)
		return err
	}
	res := &protobuf.UploadFileResponse{
		Result: fileName,
	}
	if err = stream.SendAndClose(res); err != nil {
		log.Printf("cannot send response : %v", err)
		return err
	}
	return nil
}

func main() {
	hostAddress := "0.0.0.0:8080"
	if os.Getenv("HOST_ADDRESS") != "" {
		hostAddress = os.Getenv("HOST_ADDRESS")
	}
	lis, err := net.Listen("tcp", hostAddress)
	if err != nil {
		log.Fatalf("unable to listen : %v", err)
	}
	server := grpc.NewServer()
	protobuf.RegisterFileServiceServer(server, &servertype{})

	if err = server.Serve(lis); err != nil {
		log.Fatalf("unable to serve : %v", err)
	}

}
