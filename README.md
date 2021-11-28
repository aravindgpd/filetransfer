# filetransfer
Our example is an simple file storage service that stores the plain-text files. The server would recieve request from clients to store,update,delete files and perform operations on the files stored.

# Supported Operations
1. To add files to the store.
   - Eg: store add file1.txt file2.txt
2. To list files in the store.
   - Eg: store ls
3. To remove a file in the store.
   - Eg: store rm file1.txt
4. To Update the content in the store.
   - Eg: store update file1.txt
5. To get the total count of the words available in all files combined in storage directory.
   - Eg: store wc
6. To get 10 most frequent words in all the files combined in storage directory in descending order.
   - Eg: store freq-words
    
# grpc file upload
gRPC is a great choice for client-server application development or good alternate for replacing traditional REST based inter-microservices communication. gRPC provides 4 different RPC types. One of them is Client streaming in which client can send multiple requests to the server as part of single RPC/connection. We are going to make use of this, upload a large file as small chunks into the server to implement this gRPC file upload functionality.
![image](https://user-images.githubusercontent.com/22413229/143763357-2c343bd7-643a-4a6d-83c2-787da635a831.png)

# ## Setup development and build environment

Prerequisites:

* [Go compiler](https://golang.org/dl/)

Install `protoc`:

Linux:
```bash
apt install -y protobuf-compiler
or 
Installing through pre-build binaries 
# Make sure you grab the latest version
curl -OL https://github.com/google/protobuf/releases/download/v3.5.1/protoc-3.5.1-linux-x86_64.zip
# Unzip
unzip protoc-3.5.1-linux-x86_64.zip -d protoc3
# Move protoc to /usr/local/bin/
sudo mv protoc3/bin/* /usr/local/bin/
# Move protoc3/include to /usr/local/include/
sudo mv protoc3/include/* /usr/local/include/
# Optional: change owner
sudo chown [user] /usr/local/bin/protoc
sudo chown -R [user] /usr/local/include/google

```

MacOs:
```bash
brew install protobuf
```
Windows:

Download the windows archive: https://github.com/google/protobuf/releases

Example: https://github.com/google/protobuf/releases/download/v3.5.1/protoc-3.5.1-win32.zip

Extract all to C:\proto3  

Your directory structure should now be
  - C:\proto3\bin 
  - C:\proto3\include 

Finally, add C:\proto3\bin to your PATH.


Install `protoc-gen-go` and `protoc-gen-go-grpc` binaries

```bash
go get google.golang.org/protobuf/cmd/protoc-gen-go
go get google.golang.org/grpc/cmd/protoc-gen-go-grpc
```

# Required Env Variables
In server side host address should be configured through environment variable `HOST_ADDRESS` and  default value will be `0.0.0.0:8080`.
In client side server listen address should be configured through environment variable `SERVER_ADDRESS` and default value will be `0.0.0.0:8080`.
And file storage location in server side should be configure through environment variable `STORAGE_LOCATION` and default value will be `\tmp`.


