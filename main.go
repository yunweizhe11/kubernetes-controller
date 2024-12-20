package main

import (
	"fmt"
	kubeService "kubernetes-controller/pkg"
)

func main() {
	kubeService.Logger("info", "starting...")
	defer func() {
		if r := recover(); r != nil {
			// 记录 panic 信息到日志
			kubeService.Logger("panic", fmt.Sprintf("panic: %v", r))
		}
	}()
	kubeService.Service()
}
