package main

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

func main() {
	var (
		src string
		dst string
		cnt int
		rtl int
	)
	flag.StringVar(&src, "s", "", "Source folder")
	flag.StringVar(&dst, "d", "", "Destination folder")
	flag.IntVar(&cnt, "c", 1, "Count backup files per machine")
	flag.IntVar(&rtl, "r", 0, "Rate limit in Kb/sec (0 is no limit)")
	flag.Parse()

	if src == "" || dst == "" {
		fmt.Println("-s and -d parameters required")
		os.Exit(1)
	}

	if cnt < 1 {
		fmt.Println("the key -c value must be greater than zero")
		os.Exit(1)
	}

	err := copyLastBackup(src, dst, cnt, rtl)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

}

type BackupFile struct {
	FileName         string
	FileSize         int64
	DateTimeFromName time.Time
}

func (bf BackupFile) Copy(src, dst string, rtl int) error {
	if f, err := os.Stat(filepath.Join(dst, bf.FileName)); !os.IsNotExist(err) {
		if bf.FileSize == f.Size() {
			log.Printf("file %s exist\r\n", bf.FileName)
			return nil
		}
		log.Printf("file %s exist but size (%d != %d) not equal \r\n", bf.FileName, f.Size(), bf.FileSize)
		err = os.Remove(filepath.Join(dst, bf.FileName))
		if err != nil {
			return err
		}
	}
	fmt.Printf("copy %s\r\n", bf.FileName)
	r, err := os.Open(filepath.Join(src, bf.FileName))
	if err != nil {
		return err
	}
	defer r.Close()

	w, err := os.Create(filepath.Join(dst, bf.FileName))
	if err != nil {
		return err
	}
	defer w.Close()

	if rtl < 1 || rtl > 250000000 {
		_, err = io.Copy(w, r)
		if err != nil {
			return err
		}

	} else {
		// Rate limit copy
		buf := make([]byte, 1024*4) // 4K sector size?
		t := time.NewTicker(time.Second / time.Duration(rtl) * 4)
		defer t.Stop()
		defer fmt.Println("")
		for {
			n, err := r.Read(buf)
			if err != nil && err != io.EOF {
				return err
			}
			if n == 0 {
				break
			}
			if _, err := w.Write(buf[:n]); err != nil {
				return err
			}
			<-t.C
		}
	}
	return nil
}

func (bf BackupFile) Delete(dir string) error {
	log.Printf("delete file %s\r\n", bf.FileName)
	return os.Remove(filepath.Join(dir, bf.FileName))
}

func copyLastBackup(src, dst string, cnt, rtl int) error {
	srcBackups, err := findBackup(src)
	if err != nil {
		return err
	}
	for k := range srcBackups {
		fmt.Println(k)
		c := cnt
		if len(srcBackups[k]) < cnt {
			c = len(srcBackups[k])
		}
		for i := range srcBackups[k][:c] {
			err = srcBackups[k][i].Copy(src, dst, rtl)
			if err != nil {
				return err
			}
			old, err := findOldFiles(dst, k, cnt)
			if err != nil {
				return err
			}

			for o := range old {
				log.Printf("remove old file %s\r\n", old[o].FileName)
				err = old[o].Delete(dst)
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func findOldFiles(dir, id string, cnt int) ([]BackupFile, error) {
	bf, err := findBackup(dir)
	if err != nil {
		return []BackupFile{}, err
	}
	if find, ok := bf[id]; ok {
		if len(find) < cnt {
			return []BackupFile{}, nil
		}
		return find[cnt:], nil
	}
	return []BackupFile{}, nil
}

func findBackup(dir string) (map[string][]BackupFile, error) {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	backups := map[string][]BackupFile{}

	for _, f := range files {
		if f.IsDir() {
			continue
		}
		if valid, parsedID, parsedTime := ParseName(f.Name()); valid {
			backups[parsedID] = append(
				backups[parsedID],
				BackupFile{
					FileName:         f.Name(),
					DateTimeFromName: parsedTime,
					FileSize:         f.Size(),
				})
		}
	}

	for k := range backups {
		sort.Slice(backups[k], func(i, j int) bool {
			return backups[k][i].DateTimeFromName.After(backups[k][j].DateTimeFromName)
		})
	}

	return backups, nil
}

// ParseName get file name return is valid proxmox backup format, machine id and create time
func ParseName(name string) (bool, string, time.Time) {
	s := strings.Split(filepath.Base(name), ".")
	if len(s) < 2 || len(s) > 3 {
		return false, "", time.Time{}
	}

	if s[1] != "vma" {
		return false, "", time.Time{}
	}

	if len(s) == 3 {
		if s[2] != "gz" {
			if s[2] != "lzo" {
				return false, "", time.Time{}
			}
		}
	}

	splitName := strings.Split(s[0], "-")
	if len(splitName) != 5 || splitName[0] != "vzdump" || splitName[1] != "qemu" {
		return false, "", time.Time{}
	}

	machineID := splitName[2]

	backupTime, err := time.Parse("2006_01_02-15_04_05", splitName[3]+"-"+splitName[4])
	if err != nil {
		return false, "", time.Time{}
	}

	return true, machineID, backupTime
}
