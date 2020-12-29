package core

import (
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"strings"
)

func GetABI(abiStr string) (abi.ABI, error) {
	ab, err := abi.JSON(strings.NewReader(abiStr))
	if err != nil {
		return abi.ABI{}, err
	}
	return ab, nil
}

// UnpackLogIntoMap unpacks a retrieved log into the provided map.
func UnpackLogIntoMap(a abi.ABI, out map[string]interface{}, event string, log types.Log) error {
	if len(log.Data) > 0 {
		if err := a.UnpackIntoMap(out, event, log.Data); err != nil {
			return err
		}
	}
	var indexed abi.Arguments
	for _, arg := range a.Events[event].Inputs {
		if arg.Indexed {
			indexed = append(indexed, arg)
		}
	}
	return parseTopicsIntoMap(out, indexed, log.Topics[1:])
}

// parseTopicsIntoMap converts the indexed topic field-value pairs into map key-value pairs
func parseTopicsIntoMap(out map[string]interface{}, fields abi.Arguments, topics []common.Hash) error {
	// Sanity check that the fields and topics match up
	if len(fields) != len(topics) {
		return errors.New("topic/field count mismatch")
	}
	// Iterate over all the fields and reconstruct them from topics
	for _, arg := range fields {
		if !arg.Indexed {
			return errors.New("non-indexed field in topic reconstruction")
		}

		switch arg.Type.T {
		case abi.BoolTy:
			out[arg.Name] = topics[0][common.HashLength-1] == 1
		case abi.IntTy, abi.UintTy:
			out[arg.Name] = abi.ReadInteger(arg.Type, topics[0].Bytes())
		case abi.AddressTy:
			var addr common.Address
			copy(addr[:], topics[0][common.HashLength-common.AddressLength:])
			out[arg.Name] = addr
		case abi.HashTy:
			out[arg.Name] = topics[0]
		case abi.FixedBytesTy:
			array, err := abi.ReadFixedBytes(arg.Type, topics[0].Bytes())
			if err != nil {
				return err
			}
			out[arg.Name] = array
		case abi.StringTy, abi.BytesTy, abi.SliceTy, abi.ArrayTy:
			// Array types (including strings and bytes) have their keccak256 hashes stored in the topic- not a hash
			// whose bytes can be decoded to the actual value- so the best we can do is retrieve that hash
			out[arg.Name] = topics[0]
		case abi.FunctionTy:
			if garbage := binary.BigEndian.Uint64(topics[0][0:8]); garbage != 0 {
				return fmt.Errorf("bind: got improperly encoded function type, got %v", topics[0].Bytes())
			}
			var tmp [24]byte
			copy(tmp[:], topics[0][8:32])
			out[arg.Name] = tmp
		default: // Not handling tuples
			return fmt.Errorf("unsupported indexed type: %v", arg.Type)
		}

		topics = topics[1:]
	}

	return nil
}
