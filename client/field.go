package main

import (
	"fmt"
	"math/big"
	"strings"

	lib "github.com/vszholobov/tetrisLib"
)

const moveToTopASCII = "\033[22A"
const moveRightASCII = "\r\033[36C"
const moveRightEnemyFieldASCII = "\r\033[52C"
const moveDownOneLineASCII = "\r\033[1B"
const moveDownAllLinesASCII = "\r\033[17B"

var RepresentationByType = map[lib.PieceType][]string{
	lib.TShape:      {"      Ж     ", "     ЖЖЖ    "},
	lib.ZigZagRight: {"      ЖЖ    ", "     ЖЖ     "},
	lib.ZigZagLeft:  {"     ЖЖ     ", "      ЖЖ    "},
	lib.IShape:      {"    ЖЖЖЖ    "},
	lib.RightLShape: {"    ЖЖЖ     ", "    Ж       "},
	lib.LeftLShape:  {"    ЖЖЖ     ", "      Ж     "},
	lib.SquareShape: {"     ЖЖ     ", "     ЖЖ     "},
}

func PrintEnemyField(field *big.Int, speed string, score string, cleanCount string, nextPieceType lib.PieceType) {
	fieldStr := fmt.Sprintf("%b", field)
	fmt.Print(moveToTopASCII)
	fmt.Print(moveRightEnemyFieldASCII)
	for i := lib.FieldHeight - 1; i >= 0; i-- {
		line := fieldStr[i*lib.FieldWidth : i*lib.FieldWidth+lib.FieldWidth]
		line = strings.ReplaceAll(line, "1", " Ж ")
		line = strings.ReplaceAll(line, "0", "   ")
		fmt.Print(line)
		fmt.Print(moveDownOneLineASCII)
		fmt.Print(moveRightEnemyFieldASCII)
	}
	builder.Reset()
	builder.WriteString("Score: ")
	builder.WriteString(score)
	builder.WriteString(" | Speed: ")
	builder.WriteString(speed)
	builder.WriteString(" | Cleaned: ")
	builder.WriteString(cleanCount)
	fmt.Print(builder.String())
	fmt.Print(moveDownOneLineASCII)
}

var builder = strings.Builder{}

func PrintSelfField(
	field *big.Int,
	speed string,
	score string,
	cleanCount string,
	nextPieceType lib.PieceType,
	pingMs string,
) {
	fieldStr := fmt.Sprintf("%b", field)
	fmt.Print(moveToTopASCII)
	for i := lib.FieldHeight - 1; i >= 0; i-- {
		line := fieldStr[i*lib.FieldWidth : i*lib.FieldWidth+lib.FieldWidth]
		line = strings.ReplaceAll(line, "1", " Ж ")
		line = strings.ReplaceAll(line, "0", "   ")
		fmt.Print(line)
		fmt.Print(moveDownOneLineASCII)
	}
	builder.Reset()
	builder.WriteString("Score: ")
	builder.WriteString(score)
	builder.WriteString(" | Speed: ")
	builder.WriteString(speed)
	builder.WriteString(" | Cleaned: ")
	builder.WriteString(cleanCount)
	builder.WriteString(" | Ping: ")
	builder.WriteString(pingMs)
	builder.WriteString("    ")
	fmt.Print(builder.String())
	fmt.Print(moveDownOneLineASCII)
	printNextPiece(nextPieceType)
}

func printNextPiece(nextPieceType lib.PieceType) {
	fmt.Print(moveToTopASCII + moveRightASCII + " ##############")
	fmt.Printf(moveDownOneLineASCII + moveRightASCII + " #            #")
	pieceLines := RepresentationByType[nextPieceType]
	for i := 0; i < 2; i++ {
		curLine := "            "
		if i < len(pieceLines) {
			curLine = pieceLines[i]
		}
		fmt.Printf(moveDownOneLineASCII+moveRightASCII+" #%s#", curLine)
	}
	fmt.Printf(moveDownOneLineASCII + moveRightASCII + " #            #")
	fmt.Print(moveDownOneLineASCII + moveRightASCII + " ##############")
	fmt.Print(moveDownAllLinesASCII)
}
