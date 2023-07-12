# 说明
## 介绍
因为公司涉及离线环境的交付，我们有一个服务的镜像大小有12G；每次离线升级的时候同步镜像都是一件很蛋疼的事情。
因为离线环境没有办法直接使用docker pull 来拉取镜像的；每次都需要给这个镜像save 保存下来，然后传到离线环境的机器上去。
所以我们先减轻一下每次升级传包的工作量；我们就需要获取到每次升级的docker images增量的内容

## 使用介绍
编译
```bash
GO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a 
```
因为我是放在Linux 机器上使用的，所以我这边直接编译的是linux机器的程序

运行
```bash
./dockerImageIncrementTalExport -o OldImage -n NewImage
```

## 原理介绍
我们可以发现在每次升级的时候docker pull 会自动的只拉去变更的镜像层之下的所有内容，然后它是知道需要拉去哪些层呢？
是因为在我们的docker Images 镜像中有一个文件`manifest.json`
我们观察这个文件可以发现，每次升级的时候他拉去的就是新旧镜像里面不相同的层；所以我们就利用这个来进行增量打包;
对比两个文件内相同的层，然后在新的docker images tag里面删除它