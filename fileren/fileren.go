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

	names, err = walkDir(rootDir)
	if err != nil {
		panic(err)
	}

	// change the order of the files to bottom-up
	for i, j := 0, len(names)-1; i < j; i, j = i+1, j-1 {
		names[i], names[j] = names[j], names[i]
	}

	for _, name := range names {
		newName = rename(name)

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

func rename(s string) string {
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
		"#", "",
		"!", "",
		"?", "",
	)

	name := fp.Base(s)
	path := fp.Dir(s)

	name = str.ToLower(name)
	name = applyReplacer(name, rp)
	name = str.Trim(name, "_")
	name = str.Trim(name, "-")

	return fp.Join(path, name)
}

func applyReplacer(s string, rp *str.Replacer) string {
	var oldS string
	for {
		oldS = s
		s = rp.Replace(s)
		//fmt.Println(oldS, s)
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
