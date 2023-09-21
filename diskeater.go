package main

import (
	crand "crypto/rand"
	"flag"
	"fmt"
	mrand "math/rand"
	"os"
	"strings"
)

type flags struct {
	flagHelpPtr        *bool
	flagPatternSizePtr *uint
	flagPrefixPtr      *string
	flagFileSizePtr    *uint
	flagPathPtr        *string
	flagRemoveOnExit   *bool
}

var (
	RND_PATTERN_SIZE uint   = 1024
	PREFIX           string = "KILLSSD"
	FILE_SIZE        uint   = 1073741824 //1 Gb
	PATH             string = "e:/"
	REMOVE_ON_EXIT   bool   = true
)

var pattern []byte
var bytes_count int64 = 0

func sprintf(s uint) string {
	i := fmt.Sprintf("%06d", s)
	return i
}

func path_free_space(path string) uint64 {
	usage := NewDiskUsage(path)
	return usage.Available()
}

func init_flags() flags {
	var f flags
	f.flagHelpPtr = flag.Bool("h", false, "Вывод информации об использовании программы.")
	f.flagPatternSizePtr = flag.Uint("b", RND_PATTERN_SIZE, "Размер блока случайных данных для записи в байтах")
	f.flagFileSizePtr = flag.Uint("s", FILE_SIZE, "Размер одного файла в байтах")
	f.flagPrefixPtr = flag.String("p", PREFIX, "Префикс")
	f.flagPathPtr = flag.String("path", PATH, "Путь")
	f.flagRemoveOnExit = flag.Bool("r", REMOVE_ON_EXIT, "Remove junk on exit")
	flag.Parse()
	return f
}

func init_params(f flags) {
	RND_PATTERN_SIZE = *f.flagPatternSizePtr
	if RND_PATTERN_SIZE > 1024*1024 {
		RND_PATTERN_SIZE = 1024 * 1024
	}

	PREFIX = *f.flagPrefixPtr
	FILE_SIZE = *f.flagFileSizePtr
	PATH = *f.flagPathPtr
}

func init_pattern() {
	pattern = make([]byte, RND_PATTERN_SIZE)
	_, err := crand.Read(pattern)
	if err != nil {
		fmt.Println("error:", err)
		return
	}
}

func write_pattern(fp *os.File, size uint) {
	if size > RND_PATTERN_SIZE {
		size = RND_PATTERN_SIZE
	}
	_, err := fp.Write(pattern[0:size])
	if err != nil {
		fmt.Println("Unable write to file:", err)
	} else {
		bytes_count += int64(size)
	}

}

func create_junk_file(filename string, size uint) {
	file, err := os.Create(filename)
	if err != nil {
		fmt.Println("Unable to create file:", err)
		os.Exit(1)
	}
	defer file.Close()

	var i uint
	for i = 0; i < size/RND_PATTERN_SIZE; i++ {
		write_pattern(file, RND_PATTERN_SIZE)
	}
	write_pattern(file, size%RND_PATTERN_SIZE)
}

func remove_all_junk() {
	if !REMOVE_ON_EXIT {
		return
	}
	files, err := os.ReadDir(PATH)
	if err != nil {
		fmt.Println("Unable list path:", err)
		os.Exit(1)
	}
	for _, file := range files {
		if !file.IsDir() {
			if strings.HasPrefix(file.Name(), PREFIX) {
				err := os.Remove(PATH + file.Name())
				if err != nil {
					fmt.Println("Unable remove path:", file.Name(), "  ", err)
				}

			}

		}
	}

}

func delete_rnd_file_with_prefix() {
	var junk_files []string
	files, err := os.ReadDir(PATH)

	if err != nil {
		fmt.Println("Unable list path:", err)
		os.Exit(1)
	}

	for _, file := range files {
		if !file.IsDir() {
			if strings.HasPrefix(file.Name(), PREFIX) {
				junk_files = append(junk_files, file.Name())
			}

		}
	}
	file_to_remove := junk_files[mrand.Intn(len(junk_files))]
	fmt.Println("Gonna delete...", file_to_remove)
	os.Remove(PATH + file_to_remove)
}

func main() {
	defer remove_all_junk()

	f := init_flags()
	init_params(f)

	if *f.flagHelpPtr {
		fmt.Println("Disk life Eater v0.1")
		fmt.Println("github.com/Komrakovaa/Disk-life-eater")
		flag.PrintDefaults()
		os.Exit(0)
	}
	init_pattern()

	quit := make(chan bool)
	go func() {
		var file_count uint = 0
		for {
			if path_free_space(PATH) < uint64(FILE_SIZE) {
				delete_rnd_file_with_prefix()
			}
			create_junk_file(PATH+PREFIX+sprintf(file_count), FILE_SIZE)
			fmt.Println("Total writed:", ByteCountDecimal(bytes_count))
			file_count++
			select {
			case <-quit:
				return
			default:
				// Do other stuff
			}
		}
	}()
	fmt.Println("Press enter key to exit")
	fmt.Scanln()

	// Quit goroutine
	quit <- true
}
