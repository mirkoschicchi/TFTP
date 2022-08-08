package utils

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/mirkoschicchi/TFTP/internal/app/logger"
	"github.com/pkg/errors"
)

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

func ReadFileFromFS(filename string) ([]byte, error) {
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "The error is %+v\n", err)
		return []byte{}, errors.Wrapf(err, "cannot read file %s", filename)
	}
	logger.Debug("File %s has been read from file-system", filename)
	return content, nil
}

// CreateDataBlocks returns a list of bytes array splitted in blocks
// of size 512
func CreateDataBlocks(fileContent []byte) [][]byte {
	numberOfBlocks := CalculateNumberOfBlocks(len(fileContent))

	var dataBlocks [][]byte
	for i := 0; i < numberOfBlocks; i++ {
		if i == numberOfBlocks-1 {
			dataBlocks = append(dataBlocks, fileContent[512*i:])
			continue
		}
		dataBlocks = append(dataBlocks, fileContent[512*i:512*(i+1)])
	}

	return dataBlocks
}
