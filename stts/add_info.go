package main

import (
	"bufio"
)

func readAddInfo(st *sttsT, vars *varsT) {
	var addInfo []string

	for _, fd := range vars.addInfoFd {
		fileInfo, err := fd.Stat()
		if err != nil || fileInfo.Size() == 0 {
			continue
		}

		rd := bufio.NewReaderSize(fd, 32)
		lineBin, _, _ := rd.ReadLine()
		fd.Seek(0, 0)

		addInfo = append(addInfo, string(lineBin))
	}

	st.addInfo = addInfo
}
