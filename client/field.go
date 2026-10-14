package main

import lib "github.com/vszholobov/tetrisLib"

var RepresentationByType = map[lib.PieceType][]string{
	lib.TShape:      {"      Ж     ", "     ЖЖЖ    "},
	lib.ZigZagRight: {"      ЖЖ    ", "     ЖЖ     "},
	lib.ZigZagLeft:  {"     ЖЖ     ", "      ЖЖ    "},
	lib.IShape:      {"    ЖЖЖЖ    "},
	lib.RightLShape: {"    ЖЖЖ     ", "    Ж       "},
	lib.LeftLShape:  {"    ЖЖЖ     ", "      Ж     "},
	lib.SquareShape: {"     ЖЖ     ", "     ЖЖ     "},
}
