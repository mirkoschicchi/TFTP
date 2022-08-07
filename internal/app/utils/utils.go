package utils

const (
	MAX_DATA_FIELD_LENGTH = 512
)

// CalculateNumberOfBlocks returns the number of blocks
// needed to transfer the file having its size
func CalculateNumberOfBlocks(dataSize int) int {
	// Get the number of full blocks
	var numberOfBlocks int = dataSize / MAX_DATA_FIELD_LENGTH

	// Add an additional not full block if needed
	if dataSize%MAX_DATA_FIELD_LENGTH > 0 {
		numberOfBlocks++
	}

	return numberOfBlocks
}
