package main

import (
	"fmt"
	"os"

	fp "path/filepath"
	str "strings"
)

func main() {
	var (
		rootDir string
		newName string
		names   []string
		err     error
	)

	if len(os.Args[1:]) == 1 {
		rootDir = os.Args[1]
	} else {
		fmt.Println("just one argument plz")
		os.Exit(1)
	}

	rp := str.NewReplacer(
		" ", "_",
		"(", "_",
		"[", "_",
		"{", "_",
		")", "_",
		"]", "_",
		"}", "_",
		",", "_",
		"/", "_",
		"__", "_",

		"•", "-",
		"_-", "-",
		"-_", "-",
		"--", "-",
		"_~_", "-",

		"..", ".",
		"._", ".",
		"_.", ".",
		"_.", ".",

		"/_", "/",

		"&", "_and_",
		"&&", "_and_",
		"_and_and_", "_and_",
		"$", "S",

		"\"", "",
		"'", "",
		"’", "",
		"´", "",
		"`", "",
		"#", "",
		"!", "",
		"?", "",
	)

	names, err = walkDir(rootDir)
	if err != nil {
		panic(err)
	}

	// change the order of the files to bottom-up
	for i, j := 0, len(names)-1; i < j; i, j = i+1, j-1 {
		names[i], names[j] = names[j], names[i]
	}

	for _, name := range names {
		newName = rename(name, rp)

		if str.TrimLeft(newName, "./") == name || name == "." {
			continue

		} else if _, err := os.Stat(newName); err == nil {
			backupName := getBackupName(newName)
			msg := "target exists: %s => %s\n"
			fmt.Fprintf(os.Stderr, msg, name, backupName)
			os.Rename(name, backupName)

		} else if os.IsNotExist(err) {
			os.Rename(name, newName)

		} else {
			msg := "some other error occured: %s\n"
			fmt.Fprintf(os.Stderr, msg, newName)
			os.Exit(1)
		}
	}
}

func walkDir(rootDir string) ([]string, error) {
	var names []string
	err := fp.Walk(rootDir,
		func(path string, info os.FileInfo, err error) error {
			names = append(names, path)
			return nil
		})
	return names, err
}

func rename(s string, rp *str.Replacer) string {
	fn := fp.Base(s)
	path := fp.Dir(s)

	fn = str.ToLower(fn)
	fn = replaceEmojis(fn)
	fn = applyReplacer(fn, rp)
	fn = str.Trim(fn, "_")
	fn = str.Trim(fn, "-")

	return fp.Join(path, fn)
}

func applyReplacer(s string, rp *str.Replacer) string {
	var oldS string
	for {
		oldS = s
		s = rp.Replace(s)

		if s == oldS {
			break
		}
	}

	return s
}

func getBackupName(name string) string {
	name += "~"
	for {
		if _, err := os.Stat(name); err == nil {
			name += "~"
		} else {
			return name
		}
	}
	return name
}

func replaceEmojis(s string) string {
	var out str.Builder

	for _, r := range s {
		if isEmoji(r) {
			out.WriteRune('_')
		} else {
			out.WriteRune(r)
		}
	}

	return out.String()
}

func isEmoji(r rune) bool {
	return (r >= 0x1F600 && r <= 0x1F64F) ||
		(r >= 0x1F300 && r <= 0x1F5FF) ||
		(r >= 0x1F680 && r <= 0x1F6FF) ||
		(r >= 0x2600 && r <= 0x26FF) ||
		(r >= 0x2700 && r <= 0x27BF) ||
		(r >= 0xFE00 && r <= 0xFE0F) ||
		(r >= 0x1F900 && r <= 0x1F9FF) ||
		(r >= 0x1F1E6 && r <= 0x1F1FF) ||
		(r >= 0x2B50 && r <= 0x2B55) ||
		(r >= 0x231A && r <= 0x231B) ||
		(r >= 0x23E9 && r <= 0x23FA) ||
		(r >= 0x25AA && r <= 0x25FE) ||
		(r >= 0x2194 && r <= 0x2199) ||
		(r >= 0x21A9 && r <= 0x21AA) ||
		(r >= 0x2934 && r <= 0x2935) ||
		(r == 0x3030) ||
		(r == 0x303D) ||
		(r >= 0x3297 && r <= 0x3299) ||
		(r == 0x1F004) ||
		(r == 0x1F0CF) ||
		(r >= 0x1F170 && r <= 0x1F251)
}
