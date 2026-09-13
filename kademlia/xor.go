package kademlia

import (
	"log"
	"math/big"
)

type BitArray struct {
	bits []uint8
}

// Resets all bits to 0
func (b *BitArray) Reset() {
	for i := range b.bits {
		b.bits[i] = 0
	}
}

// Set a specific bit in bit array to 0 or 1
func (b *BitArray) SetBit(index int, val uint8) {
	if index < 0 || index >= len(b.bits) {
		log.Fatal("Index out of range in SetBit")
	}
	b.bits[index] = val
}

// Get a specific bit value from bit array
func (b *BitArray) GetBit(index int) uint8 {
	if index < 0 || index >= len(b.bits) {
		log.Fatal("Index out of range in SetBit")
	}
	return b.bits[index]
}

// Create a new BitArray with the specified number of bits initialized to 0
func MakeBitArray(n int) *BitArray {
	return &BitArray{
		bits: make([]uint8, n),
	}
}

// Convert a String to a node ID
func StringToNodeID(s string) NodeID {
	bits := MakeBitArray(160)
	for i := range 160 {
		if s[i] == '0' {
			bits.bits[i] = 0
		} else {
			bits.bits[i] = 1
		}
	}

	return BitArrayToBytes(*bits)
}

// Convert a NodeID to a BitArray
func BytesToBitArray(data NodeID) BitArray {
	bitArray := make([]uint8, 0, BucketLength)
	for _, b := range data {
		for i := 7; i >= 0; i-- {
			bit := (b>>i)&1 == 1
			var bitInt uint8
			if bit {
				bitInt = 1
			} else {
				bitInt = 0
			}
			bitArray = append(bitArray, bitInt)
		}
	}
	return BitArray{bits: bitArray}
}

// Convert a NodeID to a BitArray
func BitArrayToBytes(data BitArray) NodeID {
	var result NodeID
	for i := range data.bits {
		if data.bits[i] == 1 {
			result[i/8] |= (1 << (7 - (i % 8)))
		} else {
			result[i/8] &= ^(1 << (7 - (i % 8)))
		}
	}
	return result
}

// Calculate the XOR distance between two NodeIDs
func CalcDist(id1 *NodeID, id2 *NodeID) *big.Int {
	return new(big.Int).Xor(new(big.Int).SetBytes(id1[:]), new(big.Int).SetBytes(id2[:]))
}

// XOR operation on two NodeIDs to produce a BitArray.
func XorBits(id1 *NodeID, id2 *NodeID) BitArray {
	result := MakeBitArray(BucketLength)
	id1Bits := BytesToBitArray(*id1)
	id2Bits := BytesToBitArray(*id2)

	for i := range BucketLength {
		result.bits[i] = id1Bits.bits[i] ^ id2Bits.bits[i]
	}

	return *result
}
