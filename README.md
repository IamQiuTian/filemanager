### 网页端文件上传下载目录浏览

```
easy -download -f demo.txt -p 80  #生成文件下载链接
easy -download -f demo.txt -p 80 -pwd 123456 #生成文件下载链接,需要密码校验
easy -download -d /data -p 80  #生成目录下载链接
easy -download -d /data -p 80 -pwd 123456 #生成目录下载链接,需要密码校验
easy -update -d /data -p 80  #生成上传链接，文件上传后存在/data目录下
easy -update -d /data  -p 80 -pwd 123456 #生成上传链接,需要密码校验
```
