package main

import (
	"flag"
	"fmt"
	"math"
	"os"
	"os/user"
	"strconv"
	"syscall"
	"time"

	str "strings"

	"github.com/emersion/go-imap/v2"
	"github.com/emersion/go-imap/v2/imapclient"
)

type sessionT struct {
	user       string
	pass       string
	mailServer string
	folders    []folderT

	c     *imapclient.Client
	outFd *os.File

	authTries  int
	configFile string
	resCount   int
}

type folderT struct {
	path  string
	mnemo string
}

var (
	DEBUG   bool
	OUTFILE string
)

func main() {
	var configFile string
	var privDrop bool

	flag.StringVar(&configFile, "c", "", "path to a config")
	flag.StringVar(&OUTFILE, "o", "", "path to an output file")
	flag.BoolVar(&privDrop, "p", false, "drop privileges")
	flag.BoolVar(&DEBUG, "D", false, "print debug info")
	flag.Parse()

	var s sessionT
	s.configFile = configFile

	if configFile != "" {
		s.parseConfig(configFile)
	} else {
		s.user = os.Getenv("MAIL_USER")
		s.pass = os.Getenv("MAIL_PASS")
		s.mailServer = os.Getenv("MAIL_SERVER")
		s.folders = []folderT{folderT{path: "INBOX", mnemo: "in"}}
	}

	// oneshot info fetch
	if OUTFILE == "" {
		s.login()
		for _, f := range s.folders {
			num, err := getUnseen(s.c, f.path)
			errExit(err)

			print("%-8s%-24s%-2d unseen", f.mnemo, f.path, num)
		}
		s.c.Close()
		return
	}

	var err error
	flags := os.O_CREATE | os.O_TRUNC | os.O_WRONLY
	oldUmask := syscall.Umask(0)
	s.outFd, err = os.OpenFile(OUTFILE, flags, 0644)
	syscall.Umask(oldUmask)
	errExit(err)

	if privDrop {
		dropPrivileges("nobody")
	}

	// main loop to write to an output file
	for {
		s.login()
		var res []string

		for _, f := range s.folders {
			num, err := getUnseen(s.c, f.path)
			if err != nil {
				if s.c != nil {
					s.c.Close()
				}
				s.outFd.Truncate(0)
				break
			}
			if num > 0 {
				res = append(res, f.mnemo)
			}
			print("%-8s%-24s%-2d unseen", f.mnemo, f.path, num)
		}

		if len(res) != s.resCount {
			s.resCount = len(res)

			s.outFd.Seek(0, 0)
			s.outFd.Truncate(0)

			if len(res) > 0 {
				_, err = fmt.Fprintf(s.outFd, "mail: %s",
					str.Join(res, ","))
				errExit(err)
			}
		}

		print("end of loop\n")
		time.Sleep(10 * time.Second)
	}
}

func (s *sessionT) login() {
	var err error

	if s.c != nil && s.c.State().String() == "authenticated" {
		print("already logged in.")
		return
	}

	print("connecting...")
	s.c, err = imapclient.DialTLS(s.mailServer+":993", nil)
	if err != nil {
		print("dial tls error %s", err)

		s.wait()
		s.login()

		return
	}

	print("logging in...")
	err = s.c.Login(s.user, s.pass).Wait()
	if err != nil {
		print("login error %s", err)

		s.wait()
		s.login()

		return
	}
	s.authTries = 0
}

func (s *sessionT) wait() {
	print("waiting %dm...", s.authTries)

	if s.c != nil {
		s.c.Close()
	}

	mul := time.Duration(math.Min(float64(s.authTries), 120))
	time.Sleep(mul * time.Minute)

	s.authTries++
}

func getUnseen(c *imapclient.Client, folder string) (int, error) {
	statusItems := []imap.StatusItem{imap.StatusItemNumUnseen}

	data, err := c.Status(folder, statusItems).Wait()
	if err != nil {
		return 0, err
	}

	return int(*data.NumUnseen), nil
}

func dropPrivileges(newUser string) {
	userInfo, err := user.Lookup(newUser)
	errExit(err)

	uid, err := strconv.Atoi(userInfo.Uid)
	errExit(err)

	gid, err := strconv.Atoi(userInfo.Gid)
	errExit(err)

	// unset supplementary group ids
	err = syscall.Setgroups([]int{})
	errExit(err)

	// set group id (real and effective)
	err = syscall.Setgid(gid)
	errExit(err)

	// set user id (real and effective)
	err = syscall.Setuid(uid)
	errExit(err)
}

func print(format string, a ...any) {
	if DEBUG {
		fmt.Printf("debug: "+format+"\n", a...)
		return
	}

	if OUTFILE == "" {
		fmt.Printf(format+"\n", a...)
	}
}

func errExit(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}
