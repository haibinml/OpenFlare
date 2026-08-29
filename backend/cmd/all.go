// Copyright 2025 linux.do
// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

// Package cmd 提供 CLI 命令入口
package cmd

import (
	"log"
	"sync"

	"Wavelet/OpenFlare/plugins/server/infra/task/scheduler"
	"Wavelet/OpenFlare/plugins/server/infra/task/worker"
	"Wavelet/OpenFlare/plugins/server/platform/bootstrap"

	"github.com/spf13/cobra"
)

var allCmd = &cobra.Command{
	Use:   "all",
	Short: "以融合模式同时启动 API、Worker 和 Scheduler",
	Run: func(_ *cobra.Command, _ []string) {
		log.Println("[All] 融合模式启动")
		bootstrap.RegisterAll()
		runBootstrap(bootstrap.Options{API: true})

		var wg sync.WaitGroup

		// 启动 Asynq Worker 任务处理服务
		wg.Go(func() {
			log.Println("[All] 启动 Worker 服务")
			if err := worker.StartWorker(); err != nil {
				log.Printf("[All] Worker 启动失败: %v\n", err)
			}
		})

		// 启动 Asynq 定时任务调度器
		wg.Go(func() {
			log.Println("[All] 启动 Scheduler 服务")
			if err := scheduler.StartScheduler(); err != nil {
				log.Printf("[All] Scheduler 启动失败: %v\n", err)
			}
		})

		// API 服务持有前台阻塞与退出信号处理（runHTTPApp 返回后即已优雅退出）
		runHTTPApp("API + Worker + Scheduler")
		wg.Wait()
	},
}
