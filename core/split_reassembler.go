package core

// function to split a file into chunks
import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"github.com/ipfs/go-cid"
	"os"
	"sort"
)

type SplitReassembler struct {
	SplitReassemblerParam
}
type SplitReassemblerParam struct {
	LightNode *LightNode
}

func NewSplitReassembler(param SplitReassemblerParam) SplitReassembler {
	return SplitReassembler{
		SplitReassemblerParam: param,
	}
}

//	cid of the splitupload file
func (c SplitReassembler) ReassembleFileFromCid(cidStr string) (*os.File, error) {

	cidDecode, err := cid.Decode(cidStr)
	fmt.Println(cidDecode)
	if err != nil {
		return nil, fmt.Errorf("Error decoding cid: %v", err)
	}
	node, err := c.LightNode.Node.Get(context.Background(), cidDecode)

	if err != nil {
		return nil, fmt.Errorf("Error getting node: %v", err)
	}
	var splits []UploadSplits
	json.Unmarshal(node.RawData(), &splits)
	fmt.Println(splits[0].Cid)
	keys := make([]int, 0)
	for _, split := range splits {
		keys = append(keys, split.Index)
	}
	sort.Ints(keys)
	fmt.Println(keys)
	file, err := os.Create("output")
	w := bufio.NewWriter(file)
	for _, k := range keys {
		fmt.Println(splits[k].Cid)
		cidGet, err := cid.Decode(splits[k].Cid)
		if err != nil {
			return nil, fmt.Errorf("Error decoding cid: %v", err)
		}
		splitNode, err := c.LightNode.Node.Get(context.Background(), cidGet)
		if err != nil {
			return nil, fmt.Errorf("Error getting node: %v", err)
		}

		if _, err := w.Write(splitNode.RawData()); err != nil {
			return nil, fmt.Errorf("Error writing to file: %v", err)
		}
	}

	defer file.Close()
	if err := w.Flush(); err != nil {
		return nil, fmt.Errorf("Error flushing writer: %v", err)
	}

	return file, nil
}
