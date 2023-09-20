package main

import (
	"crypto/rand"
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
}

var RND_PATTERN_SIZE uint = 1024
var PREFIX string = "KILLSSD"
var FILE_SIZE uint = 1073741824 //1 Gb
var PATH string = "e:/"

var pattern []byte
var file_count uint = 0
var bytes_count uint = 0

var (
	flagHelpPtr        *bool
	flagPatternSizePtr *uint
	flagPrefixPtr      *string
	flagFileSizePtr    *uint
	flagPathPtr        *string
)

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
	flag.Parse()
	return f
}

func init_params(f flags) {
	RND_PATTERN_SIZE = *f.flagPatternSizePtr
	if RND_PATTERN_SIZE > 1024*1024 {
		RND_PATTERN_SIZE = 1024 * 1024
	}

	PREFIX = *f.flagPrefixPtr
}

func init_pattern() {
	pattern = make([]byte, RND_PATTERN_SIZE)
	_, err := rand.Read(pattern)
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
		bytes_count += size
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

func delete_rnd_file_with_prefix() {
	files, err := os.ReadDir(PATH)
	var junk_files []string

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
	f := init_flags()
	init_params(f)

	if *f.flagHelpPtr {
		flag.PrintDefaults()
	}
	init_pattern()
	go func() {
		for {
			if path_free_space(PATH) < uint64(FILE_SIZE) {
				delete_rnd_file_with_prefix()
			}
			create_junk_file(PATH+PREFIX+sprintf(file_count), FILE_SIZE)
			file_count++
		}
	}()
	fmt.Println("Press any key to exit")
	fmt.Scanln()
}
