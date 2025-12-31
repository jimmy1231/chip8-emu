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
}
