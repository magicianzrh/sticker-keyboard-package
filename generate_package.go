/**
 *	generate sticker_package.json by images in base folder
 *
 *	TODO:read exist sticker_package.json and fix and update info
 *	TODO:comment
 *
 *	author:magicianzrh
 *
 * 	KEEP SIMPLE
 * 	simple package
 * 	{
 * 		"title":"Same Sticker",
 * 		"size_check":"md5(concat(md5(file size)))",
 * 		"stickers":{
 * 						//only support one level folder
 * 						"1.png":"md5 value",
 * 						"2.png":"md5 value"
 * 					}
 * 	}
 */
package main

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"runtime"
	"sort"
	"strconv"
	"sync"

	"path/filepath"
)

var base_path string

var package_title_str string

const PACKAGE_NAME string = "sticker_package.json"

type StickerPackage struct {
	Title     string            `json:"title"`
	SizeCheck string            `json:"size_check"`
	Stickers  map[string]string `json:"stickers"`
}

type StickerFileMd5 struct {
	FileName    string
	FileSizeMd5 string
	FileMd5     string
}

var image_map map[string]StickerFileMd5

func main() {

	if len(os.Args) < 2 {
		fmt.Println("NEED input package title string")
		os.Exit(0)
	}
	var err error
	base_path, err = filepath.Abs("./")

	if err != nil {
		log.Fatal(err)
	}
	finish_tag := make(chan int, 1)
	image_map = map[string]StickerFileMd5{}
	for v := range list_image(finish_tag) {
		image_map[v.FileName] = v
	}

	<-finish_tag
	to_json()
}

func list_image(finish_tag chan int) <-chan StickerFileMd5 {
	list, err := ioutil.ReadDir(base_path)
	if err != nil {
		log.Fatal(err)
	}

	file_path_chan := make(chan string, 0)
	result_chan := make(chan StickerFileMd5, 0)
	go func() {
		defer close(file_path_chan)
		for _, f := range list {
			if !f.IsDir() {
				file_path_chan <- f.Name()
			}
		}
	}()

	MAXNUM := runtime.NumCPU()
	cpu_chan := make(chan int, MAXNUM)
	for i := 0; i < MAXNUM; i++ {
		cpu_chan <- 1
	}

	go func() {
		var file_finish_tag sync.WaitGroup
		for fn := range file_path_chan {
			<-cpu_chan
			file_finish_tag.Add(1)
			go func(fn string) {
				fi, err := os.Open(fn)
				defer fi.Close()
				defer file_finish_tag.Done()
				if err != nil {
					log.Fatal(err)
				}
				f, err := fi.Stat()
				if err != nil {
					log.Fatal(err)
				}

				ext_str := filepath.Ext(fn)
				if file_ext, file_md5 := img_format_calc_md5(fi); file_ext == ext_str && len(ext_str) > 0 {
					r := new(StickerFileMd5)
					r.FileMd5 = file_md5
					r.FileSizeMd5 = md5str(strconv.FormatInt(f.Size(), 10))
					r.FileName = fn
					result_chan <- *r
				}

				cpu_chan <- 1
			}(fn)
		}
		go func() {
			file_finish_tag.Wait()
			close(result_chan)
			finish_tag <- 1
		}()

	}()
	return result_chan
}

func img_format_calc_md5(file *os.File) (string, string) {
	bytes := make([]byte, 4)
	n, _ := file.ReadAt(bytes, 0)
	ext_str := ""
	if n < 4 {
		ext_str = ""
	}
	if bytes[0] == 0x89 && bytes[1] == 0x50 && bytes[2] == 0x4E && bytes[3] == 0x47 {
		ext_str = ".png"
	}
	if bytes[0] == 0xFF && bytes[1] == 0xD8 {
		ext_str = ".jpg"
	}
	if bytes[0] == 0x47 && bytes[1] == 0x49 && bytes[2] == 0x46 && bytes[3] == 0x38 {
		ext_str = ".gif"
	}
	if bytes[0] == 0x42 && bytes[1] == 0x4D {
		ext_str = ".bmp"
	}

	file_md5 := md5.New()
	io.Copy(file_md5, file)
	file_md5_str := fmt.Sprintf("%x", file_md5.Sum(nil))
	return ext_str, file_md5_str
}

func md5str(data string) string {
	t := md5.New()
	io.WriteString(t, data)

	return fmt.Sprintf("%x", t.Sum(nil))
}

func to_json() {
	size_check := ""
	sticker_package := new(StickerPackage)
	sticker_package.Title = os.Args[1]
	sticker_package.Stickers = map[string]string{}

	keys := make([]string, len(image_map))
	i := 0
	for f_name, _ := range image_map {
		keys[i] = f_name
		i++
	}
	sort.Strings(keys)

	for _, f_name := range keys {
		md5_list := image_map[f_name]
		sticker_package.Stickers[f_name] = md5_list.FileMd5
		size_check += md5_list.FileSizeMd5
	}

	size_check = md5str(size_check)
	sticker_package.SizeCheck = size_check

	b, err := json.Marshal(&sticker_package)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(b))

	fi, err := os.Create(base_path + "/" + PACKAGE_NAME)
	if err != nil {
		log.Fatal(err)
	}
	fi.WriteString(string(b))
	fi.Sync()
	fi.Close()
}
