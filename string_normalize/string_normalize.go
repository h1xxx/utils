package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"unicode"

	str "strings"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

type argsS struct {
	printOrig *bool
	dropPunct *bool
	dropDigit *bool
	dropBoth  *bool
	lowerAll  *bool
}

var args argsS

func init() {
	args.printOrig = flag.Bool("d", false, "duplicate - print orig line")
	args.dropPunct = flag.Bool("p", false, "add line w/o punctuation")
	args.dropDigit = flag.Bool("n", false, "add line w/o digit")
	args.dropBoth = flag.Bool("b", false, "add line w/o digit and punct")
	args.lowerAll = flag.Bool("l", false, "add line lowercase")
}

func main() {
	flag.Parse()

	punct := `!"#$%&'()*+,-./:;?@[\]^_{|}~€¹²³°┻┬╲╱┗┛`
	punct += `–•‐○—′◦°″●“„©‘’®”…´·‚™¨¸¿¡¢£¤¥¦§«¯±»÷×=<>`
	punct += "`"
	punct += "😭😂"
	punct += "🪙️°‼↗⌐╯■□◕☮☺♂❤⬆⬇︵）🌝🍻🏳🏻🏼"
	punct += "👀👁👄👆👍👏💀💩📈📱🔥😁😂😅😆😎😱😺🙏🚀🚨🤑🤛🤜🤡🤣🤦🦀₮™♀️💎♂️"
	punct += "🤔🤷💪😉💚👌🙌😍😀🤞😊😄🙃😔😏💯✅🙄😳🐻🤝🎉"

	var replSlice []string
	for _, c := range punct {
		replSlice = append(replSlice, string(c))
		replSlice = append(replSlice, "")
	}
	punct_repl := str.NewReplacer(replSlice...)

	numbers := "0123456789"
	replSlice = nil
	for _, c := range numbers {
		replSlice = append(replSlice, string(c))
		replSlice = append(replSlice, "")
	}
	numbers_repl := str.NewReplacer(replSlice...)

	t := transform.Chain(norm.NFD,
		runes.Remove(runes.In(unicode.Mn)),
		norm.NFC)

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		res := map[string]struct{}{}

		line := scanner.Text()
		lineNorm, _, _ := transform.String(t, line)

		lineNorm = str.Map(func(r rune) rune {
			if r == 'ł' {
				return 'l'
			} else if r == 'Ł' {
				return 'L'
			}
			return r
		}, lineNorm)

		res[line] = struct{}{}
		res[lineNorm] = struct{}{}

		if *args.dropPunct {
			lineReplaced := punct_repl.Replace(line)
			res[lineReplaced] = struct{}{}

			lineReplaced = punct_repl.Replace(lineNorm)
			res[lineReplaced] = struct{}{}
		}

		if *args.dropDigit {
			lineReplaced := numbers_repl.Replace(line)
			res[lineReplaced] = struct{}{}

			lineReplaced = numbers_repl.Replace(lineNorm)
			res[lineReplaced] = struct{}{}
		}

		if *args.dropBoth {
			lineReplaced := punct_repl.Replace(line)
			lineReplaced = numbers_repl.Replace(lineReplaced)
			res[lineReplaced] = struct{}{}

			lineReplaced = numbers_repl.Replace(lineNorm)
			lineReplaced = numbers_repl.Replace(lineReplaced)
			res[lineReplaced] = struct{}{}
		}

		if *args.lowerAll {
			for l, _ := range res {
				res[str.ToLower(l)] = struct{}{}
			}
		}

		for l, _ := range res {
			fmt.Println(l)
		}
	}

	err := scanner.Err()
	errExit(err)
}

func errExit(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
