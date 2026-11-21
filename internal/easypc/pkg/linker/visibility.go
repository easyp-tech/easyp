package linker

// Visibility maintains an O(1) bitset matrix representing file visibility.
// For N files, callerFileID can see targetFileID if bit (callerFileID, targetFileID) is set.
type Visibility struct {
	numFiles int
	numWords int
	bits     []uint64
}

// NewVisibility initializes a bitset matrix for numFiles files.
func NewVisibility(numFiles int) *Visibility {
	words := (numFiles + 63) / 64
	if words == 0 {
		words = 1
	}
	return &Visibility{
		numFiles: numFiles,
		numWords: words,
		bits:     make([]uint64, numFiles*words),
	}
}

// SetVisible marks targetFileID as visible to callerFileID.
func (v *Visibility) SetVisible(callerFileID, targetFileID uint32) {
	if int(callerFileID) >= v.numFiles || int(targetFileID) >= v.numFiles {
		return
	}
	idx := int(callerFileID)*v.numWords + int(targetFileID)/64
	v.bits[idx] |= uint64(1) << (targetFileID % 64)
}

// IsVisible returns true if targetFileID is visible to callerFileID.
func (v *Visibility) IsVisible(callerFileID, targetFileID uint32) bool {
	if int(callerFileID) >= v.numFiles || int(targetFileID) >= v.numFiles {
		return false
	}
	idx := int(callerFileID)*v.numWords + int(targetFileID)/64
	return (v.bits[idx] & (uint64(1) << (targetFileID % 64))) != 0
}

// Merge copies all visible files from srcFileID into destFileID.
func (v *Visibility) Merge(destFileID, srcFileID uint32) {
	if int(destFileID) >= v.numFiles || int(srcFileID) >= v.numFiles {
		return
	}
	destStart := int(destFileID) * v.numWords
	srcStart := int(srcFileID) * v.numWords
	for w := 0; w < v.numWords; w++ {
		v.bits[destStart+w] |= v.bits[srcStart+w]
	}
}

// MergeBits merges the given words into destFileID.
func (v *Visibility) MergeBits(destFileID uint32, words []uint64) {
	if int(destFileID) >= v.numFiles {
		return
	}
	destStart := int(destFileID) * v.numWords
	limit := len(words)
	if limit > v.numWords {
		limit = v.numWords
	}
	for w := 0; w < limit; w++ {
		v.bits[destStart+w] |= words[w]
	}
}

// NumWords returns the number of 64-bit words per file in the bitset matrix.
func (v *Visibility) NumWords() int {
	return v.numWords
}
