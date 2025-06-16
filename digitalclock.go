package main

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"time"
)

func clearScreen() {
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/c", "cls")
	} else {
		cmd = exec.Command("clear")
	}
	cmd.Stdout = os.Stdout
	cmd.Run()
}

type digit [5]string
type timeSlide [8]digit

func convertDigit(num string) digit {
	switch num {
	case "0":
		return digit{"* * *", "*   *", "*   *", "*   *", "* * *"}
	case "1":
		return digit{"* *  ", "  *  ", "  *  ", "  *  ", "* * *"}
	case "2":
		return digit{"* * *", "    *", "* * *", "*    ", "* * *"}
	case "3":
		return digit{"* * *", "    *", "* * *", "    *", "* * *"}
	case "4":
		return digit{"*   *", "*   *", "* * *", "    *", "    *"}
	case "5":
		return digit{"* * *", "*    ", "* * *", "    *", "* * *"}
	case "6":
		return digit{"* * *", "*    ", "* * *", "*   *", "* * *"}
	case "7":
		return digit{"* * *", "    *", "    *", "    *", "    *"}
	case "8":
		return digit{"* * *", "*   *", "* * *", "*   *", "* * *"}
	case "9":
		return digit{"* * *", "*   *", "* * *", "    *", "* * *"}
	case ":":
		return digit{"     ", "  *  ", "     ", "  *  ", "     "}
	}
	return digit{}
}

func printResult(in timeSlide) {
	var j = 0
nextrow:
	for i := range 8 {
		fmt.Print(in[i][j], "  ")
	}
	fmt.Println()
	j++
	if j < 5 {
		goto nextrow
	}
}

func logDigitalClocl() {
	for {
		clearScreen()
		var t = time.Now()
		var h, m, s = t.Hour(), t.Minute(), t.Second()

		var tot = fmt.Sprintf("%d%d:%d%d:%d%d", h/10, h%10, m/10, m%10, s/10, s%10)
		fmt.Println(tot)
		var res timeSlide
		for index, i := range tot {
			res[index] = convertDigit(string(i))
		}
		printResult(res)
		time.Sleep(1 * time.Second)
	}
}
