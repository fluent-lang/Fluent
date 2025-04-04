/*
   The Fluent Programming Language
   -----------------------------------------------------
   This code is released under the GNU GPL v3 license.
   For more information, please visit:
   https://www.gnu.org/licenses/gpl-3.0.html
   -----------------------------------------------------
   Copyright (c) 2025 Rodrigo R. & All Fluent Contributors
   This program comes with ABSOLUTELY NO WARRANTY.
   For details type `fluent l`. This is free software,
   and you are welcome to redistribute it under certain
   conditions; type `fluent l -f` for details.
*/

package relocate

import (
	"fluent/ir/pool"
	"fluent/ir/tree"
)

var blockEnd = "__block_end__"

// Remaining relocates the remaining blocks in the queue.
// It returns the address of the remaining block.
//
// Parameters:
// - appendedBlocks: A pointer to the BlockPool where new blocks are appended.
// - blockQueue: A pointer to a slice of BlockMarshalElement pointers representing the block queue.
// - queueElement: A pointer to the current BlockMarshalElement being processed.
//
// Returns:
// - A pointer to a string representing the address of the remaining block.
func Remaining(
	appendedBlocks *pool.BlockPool,
	blockQueue *[]*tree.BlockMarshalElement,
	queueElement *tree.BlockMarshalElement,
) *string {
	// If this is the last child, redirect directly to the remaining
	if queueElement.IsLast || len(*queueElement.Element.Children) == 0 {
		if queueElement.RemainingAddr == nil {
			return &blockEnd
		}

		return queueElement.RemainingAddr
	}

	// Request an address for a new block that will hold the
	// rest of the code
	remainingAddr, remainingBuilder := appendedBlocks.RequestAddress()

	// Relocate the rest of the block
	for _, el := range *blockQueue {
		if el.Id != queueElement.Id {
			continue
		}

		el.Representation = remainingBuilder
	}

	return remainingAddr
}
