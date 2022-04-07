# leetcode collect
## 简介
`Golang` 实现的 `Leetcode` 下载工具，将提交的题解整理保持在指定目录，并生成 `README` 清单文件，便于跳转查看。

具体效果请参考: [https://github.com/realzhangm/leetcode_sync](https://github.com/realzhangm/leetcode_sync)

参考了[LeetCode_Helper](https://github.com/KivenCkl/LeetCode_Helper), 这个项目代码很工整。致谢！！！

## 功能
- 下载 Leetcode 上的提交的代码。
- 汇总成一个清单文件，即 README.md。便于在 `github` 上查看，也可离线查看。
- 下载保存的路径可在 config.toml 文件中配置。
- Leetcode 账户/密码会保存在程序执行目录下的 `.password` 文件中。第一次使用会提示输入用户名和密码。

## 运行
```shell
$ go run cmd/collect_all.go
```
## TODO
- [ ] 增加命令行参数
- [ ] 英文题目描述支持
- [ ] 支持题目在清单文件中多种排序
- [ ] 支持通过 TAG 生成汇总清单文件

