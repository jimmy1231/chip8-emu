package main

import (
	"fmt"
	"io"
	"math/rand/v2"
	"os"
)

/*
x init memory
x load program into memory
x load fonts

x FDE
clock speed
timing registers

x benchmark: IBM logo

IO - sound, keypress, display
*/

const SIZE_ADDR = 4096

// TODO: make this an input
const PROGRAM = "roms/ibm.ch8"

const ADDR_PROGRAM_START = 0x200
const ADDR_FONT_START = 0x050
const ADDR_FONT_END = 0x09F
const ADDR_DISPLAY_START = 0xf00

const DISPLAY_WIDTH = uint8(64)
const DISPLAY_HEIGHT = uint8(32)
const SPRITE_WIDTH = uint8(8)

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

var SP uint16

const INDEX_VF = 15

// hardware
var PRESSED_KEY uint8

func read_hbyte(value uint16, position int) uint8 {
	position = min(max(position, 0), 3)
	return uint8((value >> (position * 4)) & 0x000f)
}

func read_byte(value uint16, position int) uint8 {
	position = min(max(position, 0), 1)
	return uint8(value >> (position * 8))
}

func get_next(pc *uint16) uint16 {
	return *pc + 2
}

func next(pc *uint16) {
	*pc = get_next(pc)
}

func key() uint8 {
	return PRESSED_KEY & 0x0f
}

// blocks until key is pressed
func get_key() uint8 {
	// TODO
	return 0x0
}

func get_delay() uint8 {
	// TODO
	return 0x0
}

func set_delay(value uint8) {
	// TODO
}

func set_sound(value uint8) {
	// TODO
}

func font_addr(c uint8) uint16 {
	return ADDR_FONT_START + uint16(c&0x0f)*5
}

func get_bit(b uint8, offset uint8) uint8 {
	return (b >> offset) & 0x1
}

// x, y, w: in pixels
func get_display_addr(x uint8, y uint8, w uint8) (uint8, uint8) {
	return y*(w/8) + x/8, 7 - uint8(x%8)
}

func get_pixel(m []uint8, x uint8, y uint8, w uint8) uint8 {
	addr, offset := get_display_addr(x, y, w)
	return get_bit(m[addr], offset)
}

func set_pixel(m []uint8, x uint8, y uint8, w uint8, value bool) {
	addr, offset := get_display_addr(x, y, w)
	if !value {
		// unset
		m[addr] &= ^(uint8(1) << offset)
	} else {
		// set
		m[addr] |= (uint8(1) << offset)
	}
}

/*
Draws a sprite at coordinate start_x, start_y that has a width of 8 pixels
and a height of 'height' pixels.

Each row of 8 pixels is read as bit-coded starting from memory location 'sprite_addr';

Returns
- true if any screen pixels are flipped from set to unset when the sprite is drawn
- false otherwise
*/
func draw(mem []uint8, start_x uint8, start_y uint8, height uint8, sprite_addr uint16) bool {
	/*
		display
			y
			^
		5 |
		4 |
		3 |
		2 |
		1 |
		0 |----------------------P
			_ _ _ _ _ _ _ _ _ _ _  |_  _  _  _  _  _ > x
			0 1 2 3 4 5 6 7 8 9 10 11 12 13 14 15 16
			|-------------| |-------------------|  |---->
				b0                 b1                b2 ...

		P is at b1, with bit offset of 3

		memory
		0xfff
		...
		|_________________|
		|_________________|
		|_________________|
		|_________________|b2
		|_________________|b1
		|_________________|b0
		|                 |
		|                 |
		msb				  lsb
		...
		0x000
	*/
	display := mem[ADDR_DISPLAY_START:]
	sprite := mem[sprite_addr : sprite_addr+uint16(height)]
	start_x = start_x % DISPLAY_WIDTH
	start_y = start_y % DISPLAY_HEIGHT

	if IS_DEBUG {
		fmt.Println(sprite)
		print_bitmap(sprite, SPRITE_WIDTH, height)
		fmt.Printf("Drawing at (%d, %d) h=%d. Sprite_addr=0x%02x\n", start_x, start_y, height, sprite_addr)
	}

	is_collision := false

	x_s := uint8(0)
	y_s := height - 1

	for y := start_y; y < min(start_y+height, DISPLAY_HEIGHT); y++ {
		for x := start_x; x < min(start_x+8, DISPLAY_WIDTH); x++ {
			display_pixel := get_pixel(display, x, y, DISPLAY_WIDTH)

			addr, offset := get_display_addr(x_s, y_s, SPRITE_WIDTH)
			sprite_pixel := get_bit(sprite[addr], offset)

			// only modify display pixel if sprite pixel is set
			if sprite_pixel == 0x1 {
				if display_pixel == 0x1 {
					set_pixel(display, x, y, DISPLAY_WIDTH, false)
					is_collision = true
				} else {
					// set display pixel
					set_pixel(display, x, y, DISPLAY_WIDTH, true)
				}
			}

			// if IS_DEBUG {
			// 	fmt.Printf("(%d, %d): %d -> %d = %d\n",
			// 		x, y,
			// 		display_pixel, sprite_pixel,
			// 		get_pixel(display, x, y, DISPLAY_WIDTH))
			// }
			x_s++
		}

		x_s = 0
		y_s--
	}

	print_bitmap(display, DISPLAY_WIDTH, DISPLAY_HEIGHT)
	return is_collision
}

func print_bitmap(bm []uint8, w uint8, h uint8) {
	fmt.Println("------------- print -------------")
	for y := int(h) - 1; y >= 0; y-- {
		fmt.Printf("%d:\t", y)
		for x := range w {
			display_pixel := get_pixel(bm, x, uint8(y), w)
			if display_pixel == 0x1 {
				fmt.Print(" * ")
			} else {
				fmt.Print("   ")
			}
		}
		fmt.Printf("\n")
	}
	fmt.Print("\t")
	for x := range w {
		fmt.Printf("%02d ", x)
	}
	fmt.Println("\n------------- end print -------------")
}

func disp_clear(mem []uint8) {
	for addr := 0xf00; addr <= 0xfff; addr++ {
		mem[addr] = 0
	}
}

func stack_push(mem []uint8, sp *uint16, value uint16) {
	*sp += 2
	b1, b0 := read_byte(value, 1), read_byte(value, 0)
	// follow big endian
	mem[*sp] = b1
	mem[*sp+1] = b0
}

func stack_pop(mem []uint8, sp *uint16) uint16 {
	b1, b0 := mem[*sp], mem[*sp+1]
	*sp -= 2
	// follow big endian
	return (uint16(b1) << 8) | uint16(b0)
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
		next(&R_PC)

		// instruction:
		//            b1                      b0
		// |---------------------|  |--------------------|
		// 15 14 13 12 11 10  9  8  7  6  5  4  3  2  1  0
		// |---------| |---------|  |--------|  |--------|
		//     h3  			h2          h1          h0
		//             |---------------------------------|
		//                             NNN
		_, b0 := read_byte(instruction, 1), read_byte(instruction, 0)
		NNN := instruction & 0x0fff

		h3, h2, h1, h0 :=
			read_hbyte(instruction, 3),
			read_hbyte(instruction, 2),
			read_hbyte(instruction, 1),
			read_hbyte(instruction, 0)

		if IS_DEBUG {
			fmt.Printf("0x%04x:\n", instruction)
		}

		if instruction == 0x00E0 {
			// 00E0
			disp_clear(mem)

		} else if instruction == 0x00EE {
			// 00E0
			R_PC = stack_pop(mem, &SP)

		} else if instruction != 0 && h3 == 0x0 {
			// 0NNN
			// set return addr to next instruction
			stack_push(mem, &SP, R_PC)
			R_PC = NNN

		} else if h3 == 0x1 {
			// 1NNN
			R_PC = NNN

		} else if h3 == 0x2 {
			// 2NNN
			if IS_DEBUG {
				fmt.Printf("Calling subroutine at 0x%03x\n", NNN)
			}
			stack_push(mem, &SP, R_PC)
			R_PC = NNN

		} else if h3 == 0x3 {
			// 3XNN
			NN := b0
			VX := R_V[h2]
			if VX == NN {
				next(&R_PC)
			}

		} else if h3 == 0x4 {
			// 4XNN
			NN := b0
			VX := R_V[h2]
			if VX != NN {
				next(&R_PC)
			}

		} else if h3 == 0x5 {
			// 5XY0
			VX := R_V[h2]
			VY := R_V[h1]
			if VX == VY {
				next(&R_PC)
			}

		} else if h3 == 0x6 {
			// 6XNN
			VX := &R_V[h2]
			NN := b0

			if IS_DEBUG {
				fmt.Printf("Set V%d to %d\n", h2, NN)
			}

			*VX = NN

		} else if h3 == 0x7 {
			// 7XNN
			VX := &R_V[h2]
			NN := b0

			if IS_DEBUG {
				fmt.Printf("Add V%d to %d. V%d=%d\n", h2, NN, h2, *VX)
			}

			*VX += NN

		} else if h3 == 0x8 && h0 == 0x0 {
			// 8XY0
			R_V[h2] = R_V[h1]

		} else if h3 == 0x8 && h0 == 0x1 {
			// 8XY1
			R_V[h2] |= R_V[h1]

		} else if h3 == 0x8 && h0 == 0x2 {
			// 8XY2
			R_V[h2] &= R_V[h1]

		} else if h3 == 0x8 && h0 == 0x3 {
			// 8XY3
			R_V[h2] ^= R_V[h1]

		} else if h3 == 0x8 && h0 == 0x4 {
			// 8XY4
			VX := &R_V[h2]
			VY := &R_V[h1]
			result := uint16(*VX) + uint16(*VY)

			// check for uint8 overflow
			if result > 0xff {
				// set VF to 1 to indicate carryover
				R_V[INDEX_VF] = 1
			}

			*VX = uint8(result)

		} else if h3 == 0x8 && h0 == 0x5 {
			// 8XY5
			VX := &R_V[h2]
			VY := &R_V[h1]
			VF := &R_V[INDEX_VF]

			// set VF to 1 if no underflow
			if *VX < *VY {
				*VF = 0
			} else {
				*VF = 1
			}

			*VX -= *VY
			fmt.Printf("SUBTRACT: vx: %d, vy: %d, vf: %d\n", *VX, *VY, *VF)

		} else if h3 == 0x8 && h0 == 0x6 {
			// 8XY6
			VX := &R_V[h2]
			VY := &R_V[h1]
			VF := &R_V[INDEX_VF]

			*VX = *VY
			*VF = *VX & 0x1
			*VX >>= 1

		} else if h3 == 0x8 && h0 == 0x7 {
			// 8XY7
			VX := &R_V[h2]
			VY := &R_V[h1]
			VF := &R_V[INDEX_VF]

			if *VY >= *VX {
				*VF = 1
			} else {
				*VF = 0
			}

			*VX = *VY - *VX

		} else if h3 == 0x8 && h0 == 0xE {
			// 8XYE
			VX := &R_V[h2]
			VY := &R_V[h1]
			VF := &R_V[INDEX_VF]

			*VX = *VY
			*VF = (*VX >> 7) & 0x1
			*VX <<= 1

		} else if h3 == 0x9 {
			// 9XY0
			VX := &R_V[h2]
			VY := &R_V[h1]

			if *VX != *VY {
				next(&R_PC)
			}

		} else if h3 == 0xA {
			// ANNN
			if IS_DEBUG {
				fmt.Printf("Set I to 0x%x\n", NNN)
			}
			R_I = NNN

		} else if h3 == 0xB {
			// BNNN
			R_PC = uint16(R_V[0]) + NNN

		} else if h3 == 0xC {
			// CXNN
			VX := &R_V[h2]
			NN := b0

			*VX = uint8(rand.IntN(255)) & NN

		} else if h3 == 0xD {
			// DXYN
			VX := &R_V[h2]
			VY := &R_V[h1]
			VF := &R_V[INDEX_VF]
			N := h0

			is_collision := draw(mem, *VX, *VY, N, R_I)
			if is_collision {
				*VF = 1
			} else {
				*VF = 0
			}

		} else if h3 == 0xE && b0 == 0x9E {
			// EX9E
			VX := &R_V[h2]

			if key() == *VX {
				next(&R_PC)
			}

		} else if h3 == 0xE && b0 == 0xA1 {
			// EXA1
			VX := &R_V[h2]

			if key() != *VX {
				next(&R_PC)
			}

		} else if h3 == 0xF && b0 == 0x0A {
			// FX0A
			VX := &R_V[h2]
			*VX = get_key()

		} else if h3 == 0xF && b0 == 0x1E {
			// FX1E
			VX := &R_V[h2]
			R_I += uint16(*VX)

		} else if h3 == 0xF && b0 == 0x07 {
			// FX07
			VX := &R_V[h2]
			*VX = get_delay()

		} else if h3 == 0xF && b0 == 0x15 {
			// FX15
			VX := &R_V[h2]
			set_delay(*VX)

		} else if h3 == 0xF && b0 == 0x18 {
			// FX18
			VX := &R_V[h2]
			set_sound(*VX)

		} else if h3 == 0xF && b0 == 0x29 {
			// FX29
			VX := &R_V[h2]
			R_I = font_addr(*VX)

		} else if h3 == 0xF && b0 == 0x33 {
			// FX33
			VX := &R_V[h2]

			hundreds := *VX / 100
			tens := (*VX % 100) / 10
			ones := *VX % 10

			mem[R_I] = hundreds
			mem[R_I+1] = tens
			mem[R_I+2] = ones

		} else if h3 == 0xF && b0 == 0x55 {
			// FX55
			VX := &R_V[h2]
			copy(mem[R_I:], R_V[:*VX+1])

		} else if h3 == 0xF && b0 == 0x65 {
			// FX65
			VX := &R_V[h2]
			copy(R_V[:*VX+1], mem[R_I:])

		} else {
			break
		}

		if IS_DEBUG {
			fmt.Printf("I=0x%03x, PC=0x%03x, SP=0x%03x\n", R_I, R_PC, SP)
			for i := range len(R_V) - 1 {
				fmt.Printf("V%d=%d, ", i, R_V[i])
			}
			fmt.Printf("VF=%d\n", R_V[15])
			fmt.Println("-------------------------")
		}
	}
}
