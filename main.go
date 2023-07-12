package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
)

type manifest struct {
	Config   string   `json:"Config"`
	RepoTags []string `json:"RepoTags"`
	Layers   []string `json:"Layers"`
}

// execCommand 执行命令
func execCommand(command string) (err error) {
	cmd := exec.Command("/bin/bash", "-c", command)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	return
}

// 获取docker images
func dockerPullImage(imagesName string) {
	err := execCommand("docker pull " + imagesName)
	if err != nil {
		log.Printf("get docker images failed, ImagesName: %v,err: %v", imagesName, err)
		os.Exit(1)
	}
}

// dockerSaveImage 保存docker image
func dockerSaveImage(imageName, tarName string) {
	//tarName := strings.Split(imageName, "/")
	//tarNamePackage := tarName[len(tarName)-1]
	command := fmt.Sprintf("docker save %v -o %v", imageName, tarName)
	err := execCommand(command)
	if err != nil {
		log.Printf("docker save Image failed, %v", err)
		os.Exit(1)
	}
	//return tarNamePackage
}

// 解压 docker 镜像到指定目录
func decompressionDockerImagesTar(tarName, dir string) {
	os.Mkdir(dir, 0777)

	command := fmt.Sprintf("tar -xf %v -C %v > /dev/null ", tarName, dir)
	err := execCommand(command)
	if err != nil {
		log.Printf("解压缩失败")
		os.Exit(1)
	}
}

// readManifest 读取我们 manifest 文件内容
func readManifest(manifest *[]manifest, config string) {
	configLine, _ := os.ReadFile(config)
	fmt.Println(string(configLine), config)
	//configLine = configLine[1:]
	//configLine = configLine[len(configLine)-1:]
	//fmt.Println(configLine)
	err := json.Unmarshal(configLine, manifest)
	if err != nil {
		fmt.Println("反解析错误，json 格式错误")
		os.Exit(1)
	}
}

func diffSlice(slice1, slice2 []string) []string {
	var diff []string
	for _, v1 := range slice1 {
		for _, v2 := range slice2 {
			if v1 == v2 {
				diff = append(diff, v1)
			}
		}
	}
	return diff
}

// differenceFile 对比文件差异，删除我们不需要的内容
func differenceFile(oldFile, newFile, newDir string) {
	var oldManifest, newManifest []manifest
	readManifest(&oldManifest, oldFile)
	readManifest(&newManifest, newFile)
	diff := diffSlice(oldManifest[0].Layers, newManifest[0].Layers)
	for _, v := range diff {
		command := fmt.Sprintf("cd %v && rm -rf %v", newDir, strings.Split(v, "/")[0])
		_ = execCommand(command)
	}
	_ = execCommand("cd ../")
}

func tagName(image string) string {
	imageSlice := strings.Split(image, "/")
	return strings.Replace(fmt.Sprintf("%v.tar", imageSlice[len(imageSlice)-1]), ":", "_", -1)

}

func diff(oldImage, newImage string) {
	//拉去docker images
	dockerPullImage(oldImage)
	dockerPullImage(newImage)
	oldImageTarName := tagName(oldImage)
	newImageTarName := tagName(newImage)

	//保存docker image
	dockerSaveImage(oldImage, oldImageTarName)
	dockerSaveImage(newImage, newImageTarName)

	// 解压
	decompressionDockerImagesTar(oldImageTarName, "old")
	decompressionDockerImagesTar(newImageTarName, "new")

	// 对比文件差异，删除我们不需要的内容
	differenceFile("old/manifest.json", "new/manifest.json", "new")

	//打包 docker 差异内容

	command := fmt.Sprintf("tar czf %v-upgrade -C %v . ", newImageTarName, "new")
	err := execCommand(command)
	if err != nil {
		log.Printf("打包失败%v", err)
	} else {
		log.Printf("docker load 差异性文件内容为 %v-upgrade\n", newImageTarName)
		//如果打包成功，则删除中间产物
		_ = os.RemoveAll("new")
		_ = os.RemoveAll("old")
		_ = os.RemoveAll(oldImageTarName)
		_ = os.RemoveAll(newImageTarName)
	}
}

func main() {
	// 获取命令行参数，获取到old Images and newImage
	oldImage := flag.String("o", "", "old Images")
	newImage := flag.String("n", "", "new Images")
	flag.Parse()
	// 获取docker images

	diff(*oldImage, *newImage)

}
