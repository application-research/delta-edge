package api

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strconv"
	"testing"

	"github.com/filecoin-project/go-data-segment/datasegment"
	"github.com/filecoin-project/go-data-segment/merkletree"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/ipfs/go-cid"
	"github.com/stretchr/testify/assert"
)

func TestInclustionProofResponses(t *testing.T) {
	type args struct {
		filePath string
	}

	tests := []struct {
		name string
		args args
	}{
		{
			name: "Response 1",
			args: args{
				filePath: "open_status_check_test_data/edge-ur-response1.json",
			},
		},
		{
			name: "Response 2",
			args: args{
				filePath: "open_status_check_test_data/edge-ur-response2.json",
			},
		},
		{
			name: "Response 3",
			args: args{
				filePath: "open_status_check_test_data/edge-ur-response3.json",
			},
		},
		{
			name: "Response 4",
			args: args{
				filePath: "open_status_check_test_data/edge-ur-response4.json",
			},
		},
		{
			name: "Response 5",
			args: args{
				filePath: "open_status_check_test_data/edge-ur-response5.json",
			},
		},
		{
			name: "Response 6",
			args: args{
				filePath: "open_status_check_test_data/edge-ur-response6.json",
			},
		},
		{
			name: "Response 7",
			args: args{
				filePath: "open_status_check_test_data/edge-ur-response7.json",
			},
		},
		{
			name: "Response 8",
			args: args{
				filePath: "open_status_check_test_data/edge-ur-response8.json",
			},
		},
		{
			name: "Response 9",
			args: args{
				filePath: "open_status_check_test_data/edge-ur-response9.json",
			},
		},
		{
			name: "Response 10",
			args: args{
				filePath: "open_status_check_test_data/edge-ur-response10.json",
			},
		},
		{
			name: "Response 11",
			args: args{
				filePath: "open_status_check_test_data/edge-ur-response11.json",
			},
		},
	}

	// Execute the test for each test case
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			// Read JSON data from file
			data, err := ioutil.ReadFile(tt.args.filePath)
			if err != nil {
				t.Fatalf("failed to read file: %v", err)
			}
			// Create an instance of StatusCheckBySubPieceCidResponse
			response := StatusCheckBySubPieceCidResponse{}

			// Unmarshal JSON data into the response object
			err = json.Unmarshal(data, &response)
			if err != nil {
				t.Fatalf("failed to unmarshal JSON: %v", err)
			}

			var proofIndexPath []merkletree.Node
			for _, n := range response.SubPieceInfo.InclusionProof.ProofIndex.Path {
				hexString, err := hex.DecodeString(n[2:])
				if err != nil {
					t.Fatalf("failed to decode path: %v", err)
				}

				// Create a [32]byte array and copy the byte data into it
				var byteArray [32]byte
				copy(byteArray[:], hexString)
				proofIndexPath = append(proofIndexPath, byteArray)
			}

			proofIndexIndex, err := strconv.ParseUint(response.SubPieceInfo.InclusionProof.ProofIndex.Index, 0, 64)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				return
			}

			var proofSubtreePath []merkletree.Node
			for _, n := range response.SubPieceInfo.InclusionProof.ProofSubtree.Path {
				hexString, err := hex.DecodeString(n[2:])
				if err != nil {
					t.Fatalf("failed to decode path: %v", err)
				}

				// Create a [32]byte array and copy the byte data into it
				var byteArray [32]byte
				copy(byteArray[:], hexString)
				proofSubtreePath = append(proofSubtreePath, byteArray)
			}
			proofSubtreeIndex, err := strconv.ParseUint(response.SubPieceInfo.InclusionProof.ProofSubtree.Index, 0, 64)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				return
			}

			inclusionProof := &datasegment.InclusionProof{
				ProofSubtree: merkletree.ProofData{
					Path:  proofSubtreePath,
					Index: proofSubtreeIndex,
				},

				ProofIndex: merkletree.ProofData{
					Path:  proofIndexPath,
					Index: proofIndexIndex,
				},
			}

			c, err := cid.Decode(response.SubPieceInfo.VerifierData.CommPc)
			if err != nil {
				t.Fatalf("failed to decode commPc: %v", err)
			}

			veriferData := datasegment.InclusionVerifierData{
				CommPc: c,
				SizePc: abi.PaddedPieceSize(response.SubPieceInfo.VerifierData.SizePc),
			}

			newAux, err := inclusionProof.ComputeExpectedAuxData(veriferData)
			assert.NoError(t, err)

			expectedCommPa, err := cid.Decode(response.SubPieceInfo.CommPa)
			if err != nil {
				t.Fatalf("failed to decode commPa1: %v", err)
			}
			expectedAux := &datasegment.InclusionAuxData{
				CommPa: expectedCommPa,
				SizePa: abi.PaddedPieceSize(response.SubPieceInfo.SizePa),
			}
			assert.Equal(t, expectedAux, newAux)
		})
	}
}
