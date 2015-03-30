/**
 *	generate sticker_package.json by images in base folder
 *
 *	TODO:read exist sticker_package.json and fix and update info
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
	"strconv"

	"path/filepath"
)

var base_path string

var package_title_str string

var image_map map[string][]string

const PACKAGE_NAME string = "sticker_package.json"

var support_img_format_list = []string{".png", "jpg", "gif", ".bmp"}

type StickerPackage struct {
	Title     string            `json:"title"`
	SizeCheck string            `json:"size_check"`
	Stickers  map[string]string `json:"stickers"`
}

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
	list_image()

	size_check := ""
	sticker_package := new(StickerPackage)
	sticker_package.Title = os.Args[1]
	sticker_package.Stickers = map[string]string{}
	for f_name, md5_list := range image_map {
		sticker_package.Stickers[f_name] = md5_list[1]
		size_check += md5_list[0]
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

func list_image() {
	list, err := ioutil.ReadDir(base_path)
	if err != nil {
		log.Fatal(err)
	}

	image_map = map[string][]string{}
	for _, f := range list {
		if !f.IsDir() {
			ext_str := filepath.Ext(f.Name())
			for _, tmp := range support_img_format_list {
				if ext_str == tmp {
					fi, err := os.Open(f.Name())
					if err != nil {
						log.Fatal(err)
					}
					if img_format(fi) == ext_str {
						file_md5 := md5.New()
						io.Copy(file_md5, fi)
						fild_md5_str := fmt.Sprintf("%x", file_md5.Sum(nil))
						image_map[f.Name()] = []string{md5str(strconv.FormatInt(f.Size(), 10)), fild_md5_str}
					} else {
						fmt.Println(f.Name() + " format may error")
					}
				}
			}
		}

	}
}

func img_format(file *os.File) string {
	bytes := make([]byte, 4)
	n, _ := file.ReadAt(bytes, 0)
	if n < 4 {
		return ""
	}
	if bytes[0] == 0x89 && bytes[1] == 0x50 && bytes[2] == 0x4E && bytes[3] == 0x47 {
		return ".png"
	}
	if bytes[0] == 0xFF && bytes[1] == 0xD8 {
		return ".jpg"
	}
	if bytes[0] == 0x47 && bytes[1] == 0x49 && bytes[2] == 0x46 && bytes[3] == 0x38 {
		return ".gif"
	}
	if bytes[0] == 0x42 && bytes[1] == 0x4D {
		return ".bmp"
	}
	return ""
}

func md5str(data string) string {
	t := md5.New()
	io.WriteString(t, data)

	return fmt.Sprintf("%x", t.Sum(nil))
}
