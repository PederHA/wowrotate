// +build windows

package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"syscall"
	"time"
)

var (
	logDir    = "C:/World of Warcraft 5.4.8/Logs" // These are my preferred paths, if you (which is unlikely) see this
	outDir    = "F:/Logs"                         // then you should change these paths or use the -i and -o CLI options
	logName   = "WoWCombatLog.txt"
	nDays     = 7
	maxSizeMB = int64(1000)
)

func getLogFileInfo() os.FileInfo {
	fileinfo, err := os.Stat(logDir + logName)
	if err != nil {
		log.Fatal(err)
	}
	return fileinfo
}

func getFileCTime(fi os.FileInfo) time.Time {
	winData := fi.Sys().(*syscall.Win32FileAttributeData)
	return time.Unix(0, winData.CreationTime.Nanoseconds())
}

func logRotate(cTime time.Time) error {
	fn := strings.Split(logName, ".txt")[0]                                    // Is this robust at all?
	newFn := fmt.Sprintf("%s_%s.txt", fn, cTime.Format("2006-01-02T15-04-05")) // ISO8601'ish
	srcPath := logDir + logName
	destPath := outDir + newFn

	// Make all parent directories (if necessary)
	if err := os.MkdirAll(outDir, 0777); err != nil {
		return err
	}

	inputFile, err := os.Open(srcPath)
	if err != nil {
		return err
	}

	outputFile, err := os.Create(destPath)
	if err != nil {
		return err
	}

	defer outputFile.Close() // Kinda like try..finally: f.close()
	_, err = io.Copy(outputFile, inputFile)
	inputFile.Close()
	if err != nil {
		return err
	}

	err = os.Remove(srcPath)
	if err != nil {
		return err
	}

	return nil
}

func fixPathSuffix(p *string) {
	// Both forward- and backslashes are fine, but they should not be mixed.
	if strings.Contains(*p, "/") && !strings.HasSuffix(*p, "/") {
		*p = *p + "/"
	} else if strings.Contains(*p, "\\") && !strings.HasSuffix(*p, "\\") {
		*p = *p + "\\"
	}
}

func init() {
	// TODO: Parse args
	flag.Int64Var(&maxSizeMB, "s", maxSizeMB, "Max size of WoWCombatLog.txt (in MB)")
	flag.StringVar(&logDir, "i", logDir, "World of Warcraft Logs folder")
	flag.StringVar(&outDir, "o", outDir, "Folder to move expired log files to")
	flag.IntVar(&nDays, "n", nDays, "Maximum file age (in days)")
	flag.Parse()
	fixPathSuffix(&logDir)
	fixPathSuffix(&outDir)
}

func main() {
	fileinfo := getLogFileInfo()
	cTime := getFileCTime(fileinfo)

	maxAge := time.Now().AddDate(0, 0, -nDays)
	if cTime.Before(maxAge) || fileinfo.Size() > (maxSizeMB*1e6) {
		err := logRotate(cTime)
		if err != nil {
			log.Fatal(err)
		}
	}
}
