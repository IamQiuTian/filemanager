# 生成文件下载链接, 监听80端口
```
fileupdown -download -f demo.txt -p 80
```
# 生成文件下载链接,需要密码123456
```
fileupdown -download -f demo.txt -p 80 -pwd 123456 
```
# 生成目录浏览链接
```
fileupdown -download -d /data -p 80
```
# 生成目录浏览链接,需要密码123456
```
fileupdown -download -d /data -p 80 -pwd 123456
```
# 生成文件上传链接，文件上传后保存在/data目录下
```
fileupdown -upload -d /data -p 80
```
# 生成文件上传链接,需要密码123456
```
fileupdown -upload -d /data  -p 80 -pwd 123456
```
