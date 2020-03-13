package pcscommand

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/iikira/BaiduPCS-Go/baidupcs"
	"github.com/iikira/BaiduPCS-Go/pcsutil/converter"
	"io/ioutil"
	"os"
	"strings"
)

const (
	folderPrefix   = "├=="
)

type node struct {
	Name string `json:"name"`
	DisplaySize string `json:"displaySize"`
	Size int64  `json:"size"`
	Children []*node `json:"children"`
}

func JSONMarshal(t interface{}) ([]byte, error) {
	buffer := &bytes.Buffer{}
	encoder := json.NewEncoder(buffer)
	encoder.SetEscapeHTML(false)
	err := encoder.Encode(t)
	return buffer.Bytes(), err
}


func getDu(pcspath string, depth int, parent *node) int64 {
	var (
		err   error
		files baidupcs.FileDirectoryList
	)
	if depth == 0 {
		err := matchPathByShellPatternOnce(&pcspath)
		if err != nil {
			fmt.Println(err)
			return 0
		}
	}

	files, err = GetBaiduPCS().FilesDirectoriesList(pcspath, baidupcs.DefaultOrderOptions)
	if err != nil {
		fmt.Println(err)
		return 0
	}
	
	var (
		prefix          = pathPrefix
		fN              = len(files)
		indentPrefixStr = strings.Repeat(indentPrefix, depth)
		folderSize      = int64(0)
	)
	for i, file := range files {
		var fNode = node{}
		(*parent).Children = append((*parent).Children, &fNode)
		fNode.Name = file.Filename
		fNode.Size = file.Size
		fNode.DisplaySize = converter.ConvertFileSize(file.Size,2)
		if file.Isdir {
			fmt.Printf("%v%v %v/\n", indentPrefixStr, pathPrefix, file.Filename)
			var subFolderSize = getDu(file.Path, depth+1,&fNode)
			folderSize = folderSize + subFolderSize
			fNode.Size = subFolderSize
			fNode.DisplaySize = converter.ConvertFileSize(subFolderSize,2)
			continue
		}

		if i+1 == fN {
			prefix = lastFilePrefix
		}
		
		fmt.Printf("%v%v %v \t %s  \n", indentPrefixStr, prefix, file.Filename,converter.ConvertFileSize(file.Size, 2))
		folderSize = file.Size + folderSize
	}
	
	var (
		parentDepth = depth - 1
		parentIndentPrefixStr = ""
		parentPrefix = ""
	)
	
	if parentDepth > 0  {
		parentIndentPrefixStr = strings.Repeat(indentPrefix,parentDepth )
		parentPrefix = folderPrefix
	}else{
		parentIndentPrefixStr = strings.Repeat(indentPrefix, 0)
		parentPrefix = folderPrefix
	}
	
	fmt.Printf("%s%s %s \t %s \n" , parentIndentPrefixStr,parentPrefix, pcspath,converter.ConvertFileSize(folderSize,2))
	
	(*parent).Size = folderSize
	(*parent).DisplaySize = converter.ConvertFileSize((*parent).Size,2)
	return folderSize
}

// RunTree 列出树形图
func RunDu(path string) {
	var root = node{}
	root.Name = path
	root.Children = make([]*node,0)
	root.Size = getDu(path, 0, &root)
	root.DisplaySize = converter.ConvertFileSize(root.Size,2)
	var s, _ = JSONMarshal(root)
	formattedString := strings.ReplaceAll(string(s),`,"subData":null`,`,"subData":[]`)
	filename := strings.ReplaceAll(fmt.Sprintf("baidu-%s-*.json",path),"/","-")
	saveFile, _ := ioutil.TempFile(os.TempDir(), filename)
	_, writeErr := saveFile.WriteString(string(formattedString))
	if writeErr != nil {
		fmt.Printf("写入文件失败: %s\n", writeErr)
		return // 直接返回
	}
	fmt.Printf("百度网盘空间使用分析成功导出:[%s] \n请复制导出内容到 http://ialaddin.github.io/daisybaidu/ 进行查看", saveFile.Name())
}
