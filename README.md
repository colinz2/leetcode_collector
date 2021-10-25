# leetcode collect
## 简介
`Golang` 实现的 `Leetcode` 下载工具，将提交题解整理保持在指定目录，并生成 `README` 清单文件，便于跳转查看。

具体效果请参考: [https://github.com/realzhangm/leetcode_sync](https://github.com/realzhangm/leetcode_sync)

参考了，[LeetCode_Helper](https://github.com/realzhangm/LeetCode_Helper) ，
这个项目代码很工整。致谢！！！

## 功能
- 下载 Leetcode 上的提交的代码。
- 汇总成一个清单文件，即 README.md。便于在 `github` 上查看，也可离线查看。
- 密保会保存在程序执行目录下的 `.password` 文件中。

## 运行
```shell
$ go run cmd/main.go
```
## TODO
- [ ] 增加命令行参数
- [ ] 英文题目描述支持

