package colors

import "fmt"

func Color(color, msg string) string {
	return fmt.Sprintf("\033[%sm%s\033[0m", color, msg)
}

func Black(msg string) string {
	return Color("0;30", msg)
}

func Red(msg string) string {
	return Color("0;31", msg)
}

func Green(msg string) string {
	return Color("0;32", msg)
}

func Orange(msg string) string {
	return Color("0;33", msg)
}

func Blue(msg string) string {
	return Color("0;34", msg)
}

func Purple(msg string) string {
	return Color("0;35", msg)
}

func Cyan(msg string) string {
	return Color("0;36", msg)
}

func LightGray(msg string) string {
	return Color("0;37", msg)
}

func DarkGray(msg string) string {
	return Color("1;30", msg)
}

func LightRed(msg string) string {
	return Color("1;31", msg)
}

func LightGreen(msg string) string {
	return Color("1;32", msg)
}

func Yellow(msg string) string {
	return Color("1;33", msg)
}

func LightBlue(msg string) string {
	return Color("1;34", msg)
}

func LightPurple(msg string) string {
	return Color("1;35", msg)
}

func LightCyan(msg string) string {
	return Color("1;36", msg)
}

func White(msg string) string {
	return Color("1;37", msg)
}
