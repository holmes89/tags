package internal

import "math/rand"

type Color string

const (
Black          Color = "#000000"
Blue           Color = "#0000FF"
BlueViolet     Color = "#8A2BE2"
Brown          Color = "#A52A2A"
CadetBlue      Color = "#5F9EA0"
Chocolate      Color = "#D2691E"
CornflowerBlue Color = "#6495ED"
Crimson        Color = "#DC143C"
DarkBlue       Color = "#00008B"
DarkCyan       Color = "#008B8B"
DarkGoldenRod  Color = "#B8860B"
DarkGreen      Color = "#006400"
DarkMagenta    Color = "#8B008B"
DarkOliveGreen Color = "#556B2F"
DarkOrchid     Color = "#9932CC"
DarkRed        Color = "#8B0000"
DarkSlateBlue  Color = "#483D8B"
DarkSlateGray  Color = "#2F4F4F"
DarkTurquoise  Color = "#00CED1"
DarkViolet     Color = "#9400D3"
DeepPink       Color = "#FF1493"
DeepSkyBlue    Color = "#00BFFF"
DimGray        Color = "#696969"
DodgerBlue     Color = "#1E90FF"
FireBrick      Color = "#B22222"
ForestGreen    Color = "#228B22"
IndianRed      Color = "#CD5C5C"
Indigo         Color = "#4B0082"
MidnightBlue   Color = "#191970"
OliveDrab      Color = "#6B8E23"
OrangeRed      Color = "#FF4500"
)

var allColors = []Color{
Black,
Blue,
BlueViolet,
Brown,
CadetBlue,
Chocolate,
CornflowerBlue,
Crimson,
DarkBlue,
DarkCyan,
DarkGoldenRod,
DarkGreen,
DarkMagenta,
DarkOliveGreen,
DarkOrchid,
DarkRed,
DarkSlateBlue,
DarkSlateGray,
DarkTurquoise,
DarkViolet,
DeepPink,
DeepSkyBlue,
DimGray,
DodgerBlue,
FireBrick,
ForestGreen,
IndianRed,
Indigo,
MidnightBlue,
OliveDrab,
OrangeRed,
}

func GetRandomColor() Color {
	idx := rand.Intn(len(allColors))
	return allColors[idx]
}
