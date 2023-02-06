package core

// function to split a file into chunks
import (
	"bufio"
	"context"
	"fmt"
	"github.com/ipfs/go-merkledag"
	"io"
	"os"
)

var defaultChuckSize int64 = 1024 * 1024

type FileSplitter struct {
	SplitterParam
}
type SplitterParam struct {
	ChuckSize int64
	LightNode *LightNode
}

type SplitChunk struct {
	Cid   string
	Chunk []byte `json:"Chunk,omitempty"`
	Size  int
	Index int
}

func NewFileSplitter(param SplitterParam) FileSplitter {
	if param.ChuckSize == 0 {
		param.ChuckSize = defaultChuckSize
	}
	return FileSplitter{
		SplitterParam: param,
	}
}

func (c FileSplitter) SplitFileFromReaderIntoBlockstore(fileFromReader io.Reader) ([]SplitChunk, error) {
	// Read the file into a buffer
	buf := make([]byte, c.ChuckSize)
	//var chunks [][]byte
	var splitChunks []SplitChunk
	var i = 0
	for {

		n, err := fileFromReader.Read(buf)
		if n == 0 {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("Error reading file: %v", err)
		}

		rawNode := merkledag.NewRawNode(buf[:n])
		c.LightNode.Node.Add(context.Background(), rawNode)
		splitChunks = append(splitChunks, SplitChunk{
			//Chunk: buf[:n],
			Index: i,
			Size:  n,
			Cid:   rawNode.Cid().String(),
		})
		i++
	}
	return splitChunks, nil
}
func (c FileSplitter) SplitFileFromReader(fileFromReader io.Reader) ([][]byte, error) {

	// Read the file into a buffer
	buf := make([]byte, c.ChuckSize)
	var chunks [][]byte
	for {
		n, err := fileFromReader.Read(buf)
		if n == 0 {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("Error reading file: %v", err)
		}
		chunks = append(chunks, buf[:n])
	}
	return chunks, nil
}

func (c FileSplitter) SplitFile(filePath string) ([][]byte, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("Error opening file: %v", err)
	}
	defer file.Close()

	// Read the file into a buffer
	buf := make([]byte, c.ChuckSize)
	var chunks [][]byte
	for {
		n, err := file.Read(buf)
		if n == 0 {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("Error reading file: %v", err)
		}
		chunks = append(chunks, buf[:n])
	}
	return chunks, nil
}

func (c FileSplitter) ReassembleFile(filePath string, chunks [][]byte) error {
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("Error creating file: %v", err)
	}
	defer file.Close()

	w := bufio.NewWriter(file)
	for _, chunk := range chunks {
		if _, err := w.Write(chunk); err != nil {
			return fmt.Errorf("Error writing to file: %v", err)
		}
	}
	if err := w.Flush(); err != nil {
		return fmt.Errorf("Error flushing writer: %v", err)
	}
	return nil
}
