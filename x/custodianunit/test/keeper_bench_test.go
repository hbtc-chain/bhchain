package test

import (
	"testing"

	sdk "github.com/hbtc-chain/bhchain/types"
)

func BenchmarkAccountMapperGetAccountFound(b *testing.B) {
	input := setupTestInput()

	// assumes b.N < 2**24
	for i := 0; i < b.N; i++ {
		arr := []byte{byte((i & 0xFF0000) >> 16), byte((i & 0xFF00) >> 8), byte(i & 0xFF)}
		addr := sdk.CUAddress(arr)
		cu := input.ak.NewCUWithAddress(input.ctx, sdk.CUTypeUser, addr)
		input.ak.SetCU(input.ctx, cu)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		arr := []byte{byte((i & 0xFF0000) >> 16), byte((i & 0xFF00) >> 8), byte(i & 0xFF)}
		input.ak.GetCU(input.ctx, sdk.CUAddress(arr))
	}
}

func BenchmarkAccountMapperGetAccountFoundWithCoins(b *testing.B) {
	input := setupTestInput()

	// assumes b.N < 2**24
	for i := 0; i < b.N; i++ {
		arr := []byte{byte((i & 0xFF0000) >> 16), byte((i & 0xFF00) >> 8), byte(i & 0xFF)}
		addr := sdk.CUAddress(arr)
		cu := input.ak.NewCUWithAddress(input.ctx, sdk.CUTypeUser, addr)
		input.ak.SetCU(input.ctx, cu)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		arr := []byte{byte((i & 0xFF0000) >> 16), byte((i & 0xFF00) >> 8), byte(i & 0xFF)}
		input.ak.GetCU(input.ctx, sdk.CUAddress(arr))
	}
}

func BenchmarkAccountMapperSetAccount(b *testing.B) {
	input := setupTestInput()

	b.ResetTimer()

	// assumes b.N < 2**24
	for i := 0; i < b.N; i++ {
		arr := []byte{byte((i & 0xFF0000) >> 16), byte((i & 0xFF00) >> 8), byte(i & 0xFF)}
		addr := sdk.CUAddress(arr)
		cu := input.ak.NewCUWithAddress(input.ctx, sdk.CUTypeUser, addr)
		input.ak.SetCU(input.ctx, cu)
	}
}

func BenchmarkAccountMapperSetAccountWithCoins(b *testing.B) {
	input := setupTestInput()

	b.ResetTimer()

	// assumes b.N < 2**24
	for i := 0; i < b.N; i++ {
		arr := []byte{byte((i & 0xFF0000) >> 16), byte((i & 0xFF00) >> 8), byte(i & 0xFF)}
		addr := sdk.CUAddress(arr)
		cu := input.ak.NewCUWithAddress(input.ctx, sdk.CUTypeUser, addr)
		input.ak.SetCU(input.ctx, cu)
	}
}
