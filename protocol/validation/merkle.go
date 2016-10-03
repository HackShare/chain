package validation

import (
	"math"

	"golang.org/x/crypto/sha3"

	"chain/protocol/bc"
)

var (
	leafPrefix     = []byte{0x00}
	interiorPrefix = []byte{0x01}
)

// CalcMerkleRoot creates a merkle tree from a slice of transactions
// and returns the root hash of the tree.
//
// A merkle tree is a tree in which every non-leaf node is the hash of its
// children nodes. The children nodes are the hashes of transactions themselves.
// A diagram depicting how this works for bitcoin transactions
// where h(x) is a double sha256 follows:
//
//	         root = h1234 = h(h12 + h34)
//	        /                           \
//	  h12 = h(h1 + h2)            h34 = h(h3 + h4)
//	   /            \              /            \
//	h1 = h(tx1)  h2 = h(tx2)    h3 = h(tx3)  h4 = h(tx4)
//
// The number of inputs is not always a power of two which results in a
// balanced tree structure as above. In that case, parent nodes with no
// children are also zero and parent nodes with only a single left node
// are calculated by concatenating the left node with itself before hashing.
// Since this function uses nodes that are pointers to the hashes, empty nodes
// will be nil.
func CalcMerkleRoot(transactions []*bc.Tx) bc.Hash {
	if len(transactions) == 0 {
		return bc.Hash{}
	}

	tree := buildMerkleTreeStore(transactions)
	return *tree[len(tree)-1]
}

// nextPowerOfTwo returns the next highest power of two from a given number if
// it is not already a power of two. This is a helper function used during the
// calculation of a merkle tree.
func nextPowerOfTwo(n int) int {
	// Return the number if it's already a power of 2.
	if n&(n-1) == 0 {
		return n
	}

	// Figure out and return the next power of two.
	exponent := uint(math.Log2(float64(n))) + 1
	return 1 << exponent // 2^exponent
}

func hashMerkleBranches(left, right *bc.Hash) *bc.Hash {
	var hash bc.Hash
	h := sha3.New256()
	h.Write(interiorPrefix)
	h.Write(left[:])
	h.Write(right[:])

	h.Sum(hash[:0])
	return &hash
}

// buildMerkleTreeStore creates a merkle tree from a slice of transactions,
// stores it using a linear array, and returns a slice of the backing array. A
// linear array was chosen as opposed to an actual tree structure since it uses
// about half as much memory.
//
// The above diagram on CalcMerkleRoot stored as a linear array is as follows:
//
// 	[h1 h2 h3 h4 h12 h34 root]
//
// As the above shows, the merkle root is always the last element in the array.
func buildMerkleTreeStore(transactions []*bc.Tx) []*bc.Hash {
	// Calculate how many entries are required to hold the binary merkle
	// tree as a linear array and create an array of that size.
	nextPoT := nextPowerOfTwo(len(transactions))
	arraySize := nextPoT*2 - 1
	merkles := make([]*bc.Hash, arraySize)

	// Create the base transaction shas and populate the array with them.
	for i, tx := range transactions {
		witHash := [32]byte(tx.WitnessHash())
		hash := bc.Hash(sha3.Sum256(append(leafPrefix, witHash[:]...)))
		merkles[i] = &hash
	}

	// Start the array offset after the last transaction and adjusted to the
	// next power of two.
	offset := nextPoT
	for i := 0; i < arraySize-1; i += 2 {
		switch {
		// When there is no left child node, the parent is nil too.
		case merkles[i] == nil:
			merkles[offset] = nil

		// When there is no right child, the left child is used as the parent.
		case merkles[i+1] == nil:
			merkles[offset] = merkles[i]

		// The normal case sets the parent node to the double sha256
		// of the concatentation of the left and right children.
		default:
			newHash := hashMerkleBranches(merkles[i], merkles[i+1])
			merkles[offset] = newHash
		}
		offset++
	}

	return merkles
}
