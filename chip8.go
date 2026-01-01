package main

import (
	"fmt"
	"io"
	"os"
)

/*
x init memory
x load program into memory
x load fonts

FDE
clock speed
timing registers

benchmark: IBM logo

IO - sound, keypress, display
*/

const SIZE_ADDR = 4096

// TODO: make this an input
const PROGRAM = "roms/zero-demo.ch8"

const ADDR_PROGRAM_START = 0x1FF
const ADDR_FONT_START = 0x050
const ADDR_FONT_END = 0x09F

const IS_DEBUG = true

// == registers ==
// index register: points to location in memory
// only use first 12 bits as address space is 2^12 = 4096 bytes
var R_I uint16

// program counter: points at the location of the current instruction in memory
var R_PC uint16

// 16 8-bit data (general purpose) registers named V0 to VF
// VF:
// - VF register doubles as a flag for some instructions
// - VF is the carry flag, while in subtraction, it is the "no borrow" flag
// - In the draw instruction VF is set upon pixel collision
var R_V [16]uint8

func read_hbyte(value uint16, position int) uint16 {
	return value >> (max(min(position, 0), 3) * 4)
}

func read_byte(value uint16, position int) uint16 {
	return value >> (max(min(position, 0), 1) * 8)
}

func main() {
	mem := make([]uint8, SIZE_ADDR)

	fi, err := os.Open(PROGRAM)
	if err != nil {
		panic(err)
	}

	// close fi on exit and check for its returned error
	defer func() {
		if err := fi.Close(); err != nil {
			panic(err)
		}
	}()

	print_mem := func() {
		if !IS_DEBUG {
			return
		}

		len := len(mem)
		fmt.Println("================ START MEM ================")
		for addr := len - 1; addr >= 0; addr-- {
			var newline string
			if addr%4 == 0 {
				newline = "\n"
			} else {
				newline = ", "
			}

			fmt.Printf("0x%03x: %02x"+newline, addr, mem[addr])
		}
		fmt.Println("================ END MEM ================")
	}

	// make a buffer to keep chunks that are read
	mem_load_addr := ADDR_PROGRAM_START
	for {
		mem_slice := mem[mem_load_addr:]

		// read a chunk
		n, err := fi.Read(mem_slice)
		if err != nil && err != io.EOF {
			panic(err)
		}
		if n == 0 {
			break
		}

		mem_load_addr += n
	}

	fonts := []uint8{
		0xF0, 0x90, 0x90, 0x90, 0xF0, // 0
		0x20, 0x60, 0x20, 0x20, 0x70, // 1
		0xF0, 0x10, 0xF0, 0x80, 0xF0, // 2
		0xF0, 0x10, 0xF0, 0x10, 0xF0, // 3
		0x90, 0x90, 0xF0, 0x10, 0x10, // 4
		0xF0, 0x80, 0xF0, 0x10, 0xF0, // 5
		0xF0, 0x80, 0xF0, 0x90, 0xF0, // 6
		0xF0, 0x10, 0x20, 0x40, 0x40, // 7
		0xF0, 0x90, 0xF0, 0x90, 0xF0, // 8
		0xF0, 0x90, 0xF0, 0x10, 0xF0, // 9
		0xF0, 0x90, 0xF0, 0x90, 0x90, // A
		0xE0, 0x90, 0xE0, 0x90, 0xE0, // B
		0xF0, 0x80, 0x80, 0x80, 0xF0, // C
		0xE0, 0x90, 0x90, 0x90, 0xE0, // D
		0xF0, 0x80, 0xF0, 0x80, 0xF0, // E
		0xF0, 0x80, 0xF0, 0x80, 0x80, // F
	}
	copy(mem[ADDR_FONT_START:], fonts[:])

	print_mem()

	R_PC = ADDR_PROGRAM_START
	for {
		// fetch
		instruction := (uint16(mem[R_PC]) << 8) | uint16(mem[R_PC+1])
		R_PC += 2

		// instruction:
		//            b1                      b0
		// |---------------------|  |--------------------|
		// 15 14 13 12 11 10  9  8  7  6  5  4  3  2  1  0
		// |---------| |---------|  |--------|  |--------|
		//     h3  			h2          h1          h0
		_, b0 := read_byte(instruction, 1), read_byte(instruction, 0)

		h3, _, _, h0 :=
			read_hbyte(instruction, 3),
			read_hbyte(instruction, 2),
			read_hbyte(instruction, 1),
			read_hbyte(instruction, 0)

		if instruction == 0x00E0 {
			// 00E0

		} else if instruction == 0x00EE {
			// 00E0

		} else if h3 == 0x0 {
			// 0NNN

		} else if h3 == 0x1 {
			// 1NNN

		} else if h3 == 0x2 {
			// 2NNN

		} else if h3 == 0x3 {
			// 3XNN

		} else if h3 == 0x4 {
			// 4XNN

		} else if h3 == 0x5 {
			// 5XY0

		} else if h3 == 0x6 {
			// 6XNN

		} else if h3 == 0x7 {
			// 7XNN

		} else if h3 == 0x8 && h0 == 0x0 {
			// 8XY0

		} else if h3 == 0x8 && h0 == 0x1 {
			// 8XY1

		} else if h3 == 0x8 && h0 == 0x2 {
			// 8XY2

		} else if h3 == 0x8 && h0 == 0x3 {
			// 8XY3

		} else if h3 == 0x8 && h0 == 0x4 {
			// 8XY4

		} else if h3 == 0x8 && h0 == 0x5 {
			// 8XY5

		} else if h3 == 0x8 && h0 == 0x6 {
			// 8XY6

		} else if h3 == 0x8 && h0 == 0x7 {
			// 8XY7

		} else if h3 == 0x8 && h0 == 0xE {
			// 8XYE

		} else if h3 == 0x9 {
			// 9XY0

		} else if h3 == 0xA {
			// ANNN

		} else if h3 == 0xB {
			// BNNN

		} else if h3 == 0xC {
			// CXNN

		} else if h3 == 0xD {
			// DXYN

		} else if h3 == 0xE {
			// EX9E

		} else if h3 == 0xE && b0 == 0x9E {
			// EXA1

		} else if h3 == 0xE && b0 == 0xA1 {
			// FX0A

		} else if h3 == 0xF && b0 == 0x0A {
			// FX1E

		} else if h3 == 0xF && b0 == 0x1E {
			// FX07

		} else if h3 == 0xF && b0 == 0x07 {
			// FX15

		} else if h3 == 0xF && b0 == 0x15 {
			// FX18

		} else if h3 == 0xF && b0 == 0x18 {
			// FX29

		} else if h3 == 0xF && b0 == 0x29 {
			// FX33

		} else if h3 == 0xF && b0 == 0x33 {
			// FX55

		} else if h3 == 0xF && b0 == 0x55 {
			// FX65

		} else if h3 == 0xF && b0 == 0x65 {
			panic("Incorrect opcode!")
		}
	}
}
