// +build windows

package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
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

func getLogFileInfo() (os.FileInfo, error) {
	fileinfo, err := os.Stat(filepath.Join(logDir, logName))
	if err != nil {
		return nil, err
	}
	return fileinfo, nil
}

func getFileCTime(fi os.FileInfo) (*time.Time, error) {
	winData := fi.Sys().(*syscall.Win32FileAttributeData)
	if winData == nil {
		return nil, fmt.Errorf("unable to read attributes of '%s'", fi.Name())
	}
	cTime := time.Unix(0, winData.CreationTime.Nanoseconds())
	return &cTime, nil
}

func logRotate(cTime *time.Time) error {
	fn := strings.Split(logName, ".txt")[0]                                    // Is this robust at all?
	newFn := fmt.Sprintf("%s_%s.txt", fn, cTime.Format("2006-01-02T15-04-05")) // ISO8601'ish

	srcPath := filepath.Join(logDir, logName)
	destPath := filepath.Join(outDir, newFn)

	// Make all parent directories (if necessary)
	if err := os.MkdirAll(outDir, 0777); err != nil {
		return err
	}

	inFile, err := os.Open(srcPath)
	defer inFile.Close()
	if err != nil {
		return err
	}

	outFile, err := os.Create(destPath)
	defer outFile.Close() // https://www.joeshaw.org/dont-defer-close-on-writable-files/ Who knows?
	if err != nil {
		return err
	}

	// Copy log file to destination directory
	_, err = io.Copy(outFile, inFile)
	if err != nil {
		return err
	}

	// Close log file after copying, so we can delete it
	err = inFile.Close()
	if err != nil {
		return nil
	}

	// Delete original log file
	err = os.Remove(srcPath)
	if err != nil {
		return err
	}

	return nil
}

func init() {
	flag.Int64Var(&maxSizeMB, "s", maxSizeMB, "Max size of WoWCombatLog.txt (in MB)")
	flag.StringVar(&logDir, "i", logDir, "World of Warcraft Logs folder")
	flag.StringVar(&outDir, "o", outDir, "Folder to move expired log files to")
	flag.IntVar(&nDays, "n", nDays, "Maximum file age (in days)")
	flag.Parse()
}

func run() error {
	fileinfo, err := getLogFileInfo()
	if err != nil {
		return err
	}

	cTime, err := getFileCTime(fileinfo)
	if err != nil {
		return err
	}

	maxAge := time.Now().AddDate(0, 0, -nDays)
	if cTime.Before(maxAge) || fileinfo.Size() > (maxSizeMB*1e6) {
		err := logRotate(cTime)
		if err != nil {
			return err
		}
	}
	return nil
}

func main() {
	err := run()
	if err != nil {
		log.Fatal(err)
	}
}
