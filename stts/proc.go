package main

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strconv"
	"unicode"

	fp "path/filepath"
	str "strings"
)

func getProcInfo(st *sttsT, vars *varsT) {
	files, err := os.ReadDir("/proc")
	errExit(err)

	for _, file := range files {
		path := "/proc/" + file.Name()
		if !pathIsDir(path) || !hasOnlyDigits(file.Name()) {
			return
		}
		var p processT
		p.pid = file.Name()
		p.bin, _ = os.Readlink(fp.Join(path, "exe"))
		p.args = readString(fp.Join(path, "cmdline"))

		target, _ := os.Readlink(fp.Join(path, "cwd"))
		p.pwd = target

		p.readBytes, p.writeBytes = getProcessIo(p.pid)
		p.fdCount, p.files = getProcessFiles(p.pid)
		p.env = readStringSl(fp.Join(path, "environ"))
		p.stat = getPsStat(p.pid)

		st.ps = append(st.ps, p)
	}
}

func getProcessFiles(pid string) (int, []string) {
	path := fp.Join("/proc", pid, "fd")
	files, err := os.ReadDir(path)
	if err != nil {
		return 0, nil
	}

	var res []string

	for _, file := range files {
		target, _ := os.Readlink(fp.Join(path, file.Name()))
		if !elInSlice(res, target) {
			res = append(res, target)
		}
	}

	sort.Strings(res)

	return len(files), res
}

func getProcessIo(pid string) (int, int) {
	path := fp.Join("/proc", pid, "io")
	fd, err := os.Open(path)
	if err != nil {
		return 0, 0
	}
	defer fd.Close()

	var readBytes, writeBytes int64

	input := bufio.NewScanner(fd)
	for input.Scan() {
		line := str.TrimSpace(input.Text())
		if str.HasPrefix(line, "read_bytes: ") {
			rb := str.TrimPrefix(line, "read_bytes: ")
			readBytes, _ = strconv.ParseInt(rb, 10, 64)
		}
		if str.HasPrefix(line, "write_bytes: ") {
			wb := str.TrimPrefix(line, "write_bytes: ")
			writeBytes, _ = strconv.ParseInt(wb, 10, 64)
		}
	}

	return int(readBytes), int(writeBytes)
}

func getPsStat(pid string) psStatT {
	var psStat psStatT
	path := fp.Join("/proc", pid, "stat")
	fd, err := os.Open(path)
	if err != nil {
		return psStat
	}
	defer fd.Close()

	f := "%d %s %c %d %d %d %d %d %d %d "
	f += "%d %d %d %d %d %d %d %d %d %d "
	f += "%d %d %d %d %d %d %d %d %d %d "
	f += "%d %d %d %d %d %d %d %d %d %d "
	f += "%d %d %d %d %d %d %d %d %d %d "
	f += "%d %d"
	_, err = fmt.Fscanf(fd, f,
		&psStat.pid,
		&psStat.comm,
		&psStat.state,
		&psStat.ppid,
		&psStat.pgrp,
		&psStat.session,
		&psStat.tty_nr,
		&psStat.tpgid,
		&psStat.flags,
		&psStat.minflt,

		&psStat.cminflt,
		&psStat.majflt,
		&psStat.cmajflt,
		&psStat.utime,
		&psStat.stime,
		&psStat.cutime,
		&psStat.cstime,
		&psStat.priority,
		&psStat.nice,
		&psStat.num_threads,

		&psStat.itrealvalue,
		&psStat.starttime,
		&psStat.vsize,
		&psStat.rss,
		&psStat.rsslim,
		&psStat.startcode,
		&psStat.endcode,
		&psStat.startstack,
		&psStat.kstkesp,
		&psStat.kstkeip,

		&psStat.signal,
		&psStat.blocked,
		&psStat.sigignore,
		&psStat.sigcatch,
		&psStat.wchan,
		&psStat.nswap,
		&psStat.cnswap,
		&psStat.exit_signal,
		&psStat.processor,
		&psStat.rt_priority,

		&psStat.policy,
		&psStat.delayacct_blkio_ticks,
		&psStat.guest_time,
		&psStat.cguest_time,
		&psStat.start_data,
		&psStat.end_data,
		&psStat.start_brk,
		&psStat.arg_start,
		&psStat.arg_end,
		&psStat.env_start,

		&psStat.env_end,
		&psStat.exit_code,
	)
	errExit(err)
	return psStat
}

func readString(path string) string {
	tempBin := make([]byte, 8192)

	fd, _ := os.Open(path)
	fd.Read(tempBin)
	fd.Close()

	return str.TrimSpace(str.Replace(string(tempBin[:]), "\x00", " ", -1))
}

func readStringSl(path string) []string {
	tempBin := make([]byte, 8192)

	fd, _ := os.Open(path)
	fd.Read(tempBin)
	fd.Close()

	fields := str.Split(string(tempBin[:]), "\x00")

	var res []string
	for _, f := range fields {
		if f != "" {
			res = append(res, f)
		}
	}

	sort.Strings(res)

	return res
}

func pathIsDir(path string) bool {
	file, err := os.Open(path)
	if err != nil {
		return false
	}
	defer file.Close()
	fileInfo, err := file.Stat()
	if fileInfo.IsDir() {
		return true
	}
	return false
}

func hasOnlyDigits(s string) bool {
	for _, r := range s {
		if !unicode.IsNumber(r) {
			return false
		}
	}
	return true
}
