##### 生成文件下载链接, 监听80端口
```
filemanager -download -f demo.txt -p 80
```
##### 生成文件下载链接,需要密码123456
```
filemanager -download -f demo.txt -p 80 -pwd 123456 
```
##### 生成目录浏览链接
```
filemanager -download -d /data -p 80
```
##### 生成目录浏览链接,需要密码123456
```
filemanager -download -d /data -p 80 -pwd 123456
```
##### 生成文件上传链接，文件上传后保存在/data目录下
```
filemanager -upload -d /data -p 80
```
##### 生成文件上传链接,需要密码123456
```
filemanager -upload -d /data  -p 80 -pwd 123456
```
