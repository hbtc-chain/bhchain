package types

import "fmt"

type Epoch struct {
	Index             uint64      `json:"index" yaml:"index"`
	StartBlockNum     uint64      `json:"start_block_num" yaml:"start_block_num"` // inclusive
	EndBlockNum       uint64      `json:"end_block_num" yaml:"end_block_num"`     // inclusive, 0 means this epoch is not finished yet
	KeyNodeSet        []CUAddress `json:"key_node_set" yaml:"key_node_set"`
	MigrationFinished bool        `json:"migration_finished" yaml:"migration_finished"`
}

func NewEpoch(index uint64, startBlockNum uint64, endBlockNum uint64, keyNodeSet []CUAddress, finished bool) Epoch {
	return Epoch{
		Index:             index,
		StartBlockNum:     startBlockNum,
		EndBlockNum:       endBlockNum,
		KeyNodeSet:        keyNodeSet,
		MigrationFinished: finished,
	}
}

func (e Epoch) String() string {
	return fmt.Sprintf(`Epoch
  Index:              %d
  Start Block Num:    %d
  End Block Num:      %d
  KeyNodes:           %v
  Migration Finished: %t`, e.Index, e.StartBlockNum, e.EndBlockNum, e.KeyNodeSet, e.MigrationFinished)
}
