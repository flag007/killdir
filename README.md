# killdir
别问 问就是不知道用来干啥的

用法:
```
-c 设置线程
-t 设置阈值
-v 输出详细信息
```

```
▶ cat dir.out | killdir
```

```
▶ killdir -f dir.json
```

200
```
▶ cat */* | grep -oE "^200.*" | killdir
```

403
```
▶ cat */* | grep -oE "^403.*" | killdir
```

安装:

```
▶ go get -u -v github.com/flag007/killdir
```
