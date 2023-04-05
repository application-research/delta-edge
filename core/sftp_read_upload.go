package core

//
//import (
//	"fmt"
//	"github.com/pkg/sftp"
//	"golang.org/x/crypto/ssh"
//	"io"
//	"log"
//	"os"
//)
//
//func main() {
//	// Connect to the SFTP server
//	sshConfig := &ssh.ClientConfig{
//		User: "username",
//		Auth: []ssh.AuthMethod{
//			ssh.Password("password"),
//		},
//		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
//	}
//	conn, err := ssh.Dial("tcp", "host:port", sshConfig)
//	if err != nil {
//		log.Fatal("Failed to dial: ", err)
//	}
//	defer conn.Close()
//
//	// Open a new SFTP session
//	sftp, err := sftp.NewClient(conn)
//	if err != nil {
//		log.Fatal("Failed to start SFTP session: ", err)
//	}
//	defer sftp.Close()
//
//	// Open the directory
//	srcDir, err := sftp.Open("/path/to/directory")
//	if err != nil {
//		log.Fatal("Failed to open directory: ", err)
//	}
//	defer srcDir.Close()
//
//	// Read the directory
//	entries, err := srcDir.Readdir(0)
//	if err != nil {
//		log.Fatal("Failed to read directory: ", err)
//	}
//
//	// Download each file in the directory
//	for _, entry := range entries {
//		srcPath := "/path/to/directory/" + entry.Name()
//		dstPath := "./" + entry.Name()
//
//		srcFile, err := sftp.Open(srcPath)
//		if err != nil {
//			log.Fatal("Failed to open file: ", err)
//		}
//		defer srcFile.Close()
//
//		dstFile, err := os.Create(dstPath)
//		if err != nil {
//			log.Fatal("Failed to create file: ", err)
//		}
//		defer dstFile.Close()
//
//		_, err = io.Copy(dstFile, srcFile)
//		if err != nil {
//			log.Fatal("Failed to copy file: ", err)
//		}
//
//		fmt.Printf("Downloaded %s\n", srcPath)
//	}
//}
