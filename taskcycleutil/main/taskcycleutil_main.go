package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/taskcycleutil"
)

// ========== 下载任务配置 ==========
// DownloadTaskConfig 下载任务的自定义参数
type DownloadTaskConfig struct {
	SaveDir    string        `json:"saveDir"`    // 文件保存目录
	Timeout    time.Duration `json:"timeout"`    // 单个URL下载超时（秒）
	RetryCount int           `json:"retryCount"` // 重试次数
}

// ========== 核心下载函数 ==========
// downloadFile 单个URL下载实现
func downloadFile(ctx context.Context, url, saveDir string, timeout time.Duration) (string, error) {
	// 创建保存目录
	if err := os.MkdirAll(saveDir, 0755); err != nil {
		return "", fmt.Errorf("create save dir failed: %w", err)
	}

	// 设置HTTP客户端超时
	client := &http.Client{
		Timeout: timeout,
	}

	// 发起HTTP请求
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", fmt.Errorf("create request failed: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// 检查响应状态
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("invalid status code: %d", resp.StatusCode)
	}

	// 生成保存文件名（基于URL和时间戳）
	fileName := fmt.Sprintf("%s_%d%s",
		filepath.Base(url),
		time.Now().UnixNano(),
		filepath.Ext(url),
	)
	savePath := filepath.Join(saveDir, fileName)

	// 创建文件并写入内容
	file, err := os.Create(savePath)
	if err != nil {
		return "", fmt.Errorf("create file failed: %w", err)
	}
	defer file.Close()

	// 拷贝响应内容到文件
	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return "", fmt.Errorf("write file failed: %w", err)
	}

	return savePath, nil
}

// ========== 框架执行函数适配 ==========
// downloadTaskExecuteFunc 适配框架的执行函数
// 功能：遍历URL数组，逐个下载，全部成功则返回true，否则返回false
func downloadTaskExecuteFunc(ctx context.Context, task *taskcycleutil.Task) (bool, error) {

	// 校验配置
	if len(task.Key) == 0 {
		return false, errors.New("empty key")
	}
	url := task.Key
	downloadTaskConfig, ok := task.Data.Params["downloadConfig"].(*DownloadTaskConfig)
	if !ok {
		return false, errors.New("downloadTaskConfig is empty")
	}
	if downloadTaskConfig == nil {
		return false, errors.New("downloadTaskConfig is nil")
	}

	if downloadTaskConfig.SaveDir == "" {
		downloadTaskConfig.SaveDir = "/tmp" // 默认保存目录
	}
	if downloadTaskConfig.Timeout <= 0 {
		downloadTaskConfig.Timeout = 70 * time.Minute // 默认30秒超时
	}
	if downloadTaskConfig.RetryCount < 0 {
		downloadTaskConfig.RetryCount = 1 // 默认重试2次
	}

	// 遍历URL执行下载（带重试）
	successCount := 0
	failedURLs := make([]string, 0)

	var savePath string
	var err error

	// 重试逻辑
	for i := 0; i <= downloadTaskConfig.RetryCount; i++ {
		savePath, err = downloadFile(ctx, url, downloadTaskConfig.SaveDir, downloadTaskConfig.Timeout)
		if err == nil {
			belogs.Info(fmt.Sprintf("download success: %s -> %s (retry: %d)", url, savePath, i))
			successCount++
			break
		}
		belogs.Warn(fmt.Sprintf("download failed: %s (retry: %d): %v", url, i, err))

		// 最后一次重试失败，记录失败URL
		if i == downloadTaskConfig.RetryCount {
			failedURLs = append(failedURLs, url)
		}

		// 非最后一次重试，等待1秒后重试
		if i < downloadTaskConfig.RetryCount {
			time.Sleep(1 * time.Second)
		}
	}

	// 全部URL下载成功则返回true，否则返回false
	if len(failedURLs) == 0 {
		return true, nil
	}

	return false, fmt.Errorf("download failed for url: %v (success: %d/%d)",
		failedURLs, successCount, len(failedURLs)+successCount)
}

// ========== 递归生成新任务函数（可选） ==========
// generateNewDownloadTasks 从成功的下载任务生成新任务（示例：下载关联文件）
// 注：仅在AddTaskModeRecursive模式下生效
func generateNewDownloadTasks(completedTask *taskcycleutil.Task) []*taskcycleutil.Task {
	// 示例逻辑：从已下载文件的URL生成新的下载任务（可自定义业务逻辑）
	// 这里简化处理：生成一个空任务（实际需根据业务场景扩展）
	newTask := &taskcycleutil.Task{
		Key: fmt.Sprintf("download_task_%s_%d", completedTask.Key, time.Now().UnixNano()),
		Data: taskcycleutil.TaskData{
			Content: `{"urls":["https://example.com/关联文件.txt"], "saveDir":"./downloads关联"}`,
		},
	}
	return []*taskcycleutil.Task{newTask}
}

// ========== 主函数：框架使用示例 ==========
func main() {
	// ========== 1. 初始化框架配置 ==========
	// 选择模式：
	// - AddTaskModeRecursive：递归模式（从成功任务生成新任务）
	// - AddTaskModeExternal：外部模式（仅外部添加任务，待周期执行）
	mode := taskcycleutil.AddTaskModeExternal
	config := taskcycleutil.NewConfig(mode)
	// 可自定义配置（默认30分钟周期、70分钟超时）
	config.CycleInterval = 30 * time.Minute
	config.MaxTimeout = 70 * time.Minute
	config.CheckInterval = 10 * time.Minute

	// ========== 2. 创建框架实例 ==========
	framework, err := taskcycleutil.NewTaskFramework(config)
	if err != nil {
		belogs.Error("create task framework failed: ", err)
		os.Exit(1)
	}
	defer framework.Stop() // 程序退出时停止框架

	// ========== 3. （可选）设置递归生成任务函数（仅递归模式需要） ==========
	if mode == taskcycleutil.AddTaskModeRecursive {
		framework.SetGenerateTasksFunc(generateNewDownloadTasks)
	}

	// ========== 4. 添加禁止执行的URL/任务Key（可选） ==========
	// 示例：禁止下载特定域名的任务
	forbiddenKeys := []string{
		"download_task_forbidden_1", // 任务Key
	}
	framework.AddForbiddenKeys(forbiddenKeys...)

	// ========== 5. 构建下载任务列表 ==========
	// 待下载的URL数组
	downloadURLs := []string{
		"https://speedtest.xjtu.edu.cn/",
		"https://fast.com/",
		"https://www.speedtest.net/",
		"https://www.speedtest.cn",
		"https://suyan.baidu.com",
		"https://wangsu.360.cn",
		"https://rrdp.afrinic.net/notification.xml",
		"https://rrdp.afrinic.net/8fe3109e-2561-4627-8850-83ab94b9bb91/73121/snapshot.xml",
		"https://rrdp.afrinic.net/8fe3109e-2561-4627-8850-83ab94b9bb91/73120/delta.xml",
		"https://rrdp.afrinic.net/8fe3109e-2561-4627-8850-83ab94b9bb91/73119/delta.xml",
		"https://rrdp.afrinic.net/8fe3109e-2561-4627-8850-83ab94b9bb91/73118/delta.xml",
		"https://rrdp.afrinic.net/8fe3109e-2561-4627-8850-83ab94b9bb91/73117/delta.xml",
		"https://rrdp.afrinic.net/8fe3109e-2561-4627-8850-83ab94b9bb91/73116/delta.xml",
		"https://rrdp.afrinic.net/8fe3109e-2561-4627-8850-83ab94b9bb91/73115/delta.xml",
		"https://rrdp.afrinic.net/8fe3109e-2561-4627-8850-83ab94b9bb91/73114/delta.xml",
		"https://rrdp.apnic.net/notification.xml",
		"https://rrdp.arin.net/notification.xml",
		"https://rrdp.lacnic.net/rrdp/notification.xml",
		"https://rrdp.lacnic.net/rrdpas0/notification.xml",
		"https://rrdp.ripe.net/notification.xml",
		"https://rrdp-as0.apnic.net/notification.xml",
		"https://0.sb/rrdp/notification.xml",
		"https://ca.nat.moe/rrdp/notification.xml",
		"https://ca.rg.net/rrdp/notify.xml",
		"https://chloe.sobornost.net/rpki/news.xml",
		"https://cloudie.rpki.app/rrdp/notification.xml",
		"https://dev.tw/rpki/notification.xml",
		"https://krill.accuristechnologies.ca/rrdp/notification.xml",
		"https://krill.ca-bc-01.ssmidge.xyz/rrdp/notification.xml",
		"https://krill.rayhaan.net/rrdp/notification.xml",
		"https://krill.stonham.uk/rrdp/notification.xml",
		"https://magellan.ipxo.com/rrdp/notification.xml",
		"https://pub.krill.ausra.cloud/rrdp/notification.xml",
		"https://pub.rpki.win/rrdp/notification.xml",
		"https://repo.kagl.me/rpki/notification.xml",
		"https://repo-rpki.idnic.net/rrdp/notification.xml",
		"https://rov-measurements.nlnetlabs.net/rrdp/notification.xml",
		"https://rpki.0i1.eu/rrdp/notification.xml",
		"https://rpki.admin.freerangecloud.com/rrdp/notification.xml",
		"https://rpki.akrn.net/rrdp/notification.xml",
		"https://rpki.apernet.io/rrdp/notification.xml",
		"https://rpki.berrybyte.network/rrdp/notification.xml",
		"https://rpki.caramelfox.net/rrdp/notification.xml",
		"https://rpki.cc/rrdp/notification.xml",
		"https://rpki.cernet.edu.cn/rrdp/notification.xml",
		"https://rpki.cnnic.cn/rrdp/notify.xml",
		"https://rpki.ezdomain.ru/rrdp/notification.xml",
		"https://rpki.co/rrdp/notification.xml",
		"https://rpki.folf.systems/rrdp/notification.xml",
		"https://rpki.komorebi.network:3030/rrdp/notification.xml",
		"https://rpki.luys.cloud/rrdp/notification.xml",
		"https://rpki.multacom.com/rrdp/notification.xml",
		"https://rpki.nap.re:3030/rrdp/notification.xml",
		"https://rpki.netiface.net/rrdp/notification.xml",
		"https://rpki.owl.net/rrdp/notification.xml",
		"https://rpki.pedjoeang.group/rrdp/notification.xml",
		"https://rpki.pudu.be/rrdp/notification.xml",
		"https://rpki.qs.nu/rrdp/notification.xml",
		"https://rpki.rand.apnic.net/rrdp/notification.xml",
		"https://rpki.roa.net/rrdp/notification.xml",
		"https://rpki.sailx.co/rrdp/notification.xml",
		"https://rpki.sn-p.io/rrdp/notification.xml",
		"https://rpki.ssmidge.xyz/rrdp/notification.xml",
		"https://rpki.telecentras.lt/rrdp/notification.xml",
		"https://rpki.tools.westconnect.ca/rrdp/notification.xml",
		"https://rpki.xindi.eu/rrdp/notification.xml",
		"https://rpki.zappiehost.com/rrdp/notification.xml",
		"https://rpki-01.pdxnet.uk/rrdp/notification.xml",
		"https://rpki1.rpki-test.sit.fraunhofer.de/rrdp/notification.xml",
		"https://rpkica.mckay.com/rrdp/notify.xml",
		"https://rpki-publication.haruue.net/rrdp/notification.xml",
		"https://rpki-repo.registro.br/rrdp/notification.xml",
		"https://rpki-repository.nic.ad.jp/rrdp/ap/notification.xml",
		"https://rpki-rrdp.mnihyc.com/rrdp/notification.xml",
		"https://rpki-rrdp.us-east-2.amazonaws.com/rrdp/08c2f264-23f9-49fb-9d43-f8b50bec9261/notification.xml",
		"https://rpki-rrdp.us-east-2.amazonaws.com/rrdp/16f1ffee-7461-4674-bb05-fddefa9a02c6/notification.xml",
		"https://rpki-rrdp.us-east-2.amazonaws.com/rrdp/20aa329b-fc52-4c61-bf53-09725c042942/notification.xml",
		"https://rpki-rrdp.us-east-2.amazonaws.com/rrdp/2f059a21-d41b-4846-b7ae-7ea38c32fd4c/notification.xml",
		"https://rpki-rrdp.us-east-2.amazonaws.com/rrdp/42582c67-dd3f-4bc5-ba60-e97e552c6e35/notification.xml",
		"https://rpki-rrdp.us-east-2.amazonaws.com/rrdp/517f3ed7-58b5-4796-be37-14d62e48f056/notification.xml",
		"https://rpki-rrdp.us-east-2.amazonaws.com/rrdp/54602fb0-a9d4-4f9f-b0ca-be2a139ea92b/notification.xml",
		"https://rpki-rrdp.us-east-2.amazonaws.com/rrdp/602a26e5-4a9e-4e5e-89f0-ef891490d9c9/notification.xml",
		"https://rpki-rrdp.us-east-2.amazonaws.com/rrdp/708aafaf-00b4-485b-854c-0b32ca30f57b/notification.xml",
		"https://rpki-rrdp.us-east-2.amazonaws.com/rrdp/71e5236f-c6f1-4928-a1b9-8def09c06085/notification.xml",
		"https://rpki-rrdp.us-east-2.amazonaws.com/rrdp/967a255c-d680-42d3-9ec3-ecb3f9da088c/notification.xml",
		"https://rpki-rrdp.us-east-2.amazonaws.com/rrdp/a841823c-a10d-477c-bfdf-4086f0b1594c/notification.xml",
		"https://rpki-rrdp.us-east-2.amazonaws.com/rrdp/b3f6b688-cff4-402f-97d5-02f6f1886b7e/notification.xml",
		"https://rpki-rrdp.us-east-2.amazonaws.com/rrdp/b68a32ee-455d-483a-943d-1a5be748bfea/notification.xml",
		"https://rpki-rrdp.us-east-2.amazonaws.com/rrdp/b8a1dd25-c313-4f25-ac21-bf55514d9c7d/notification.xml",
		"https://rpki-rrdp.us-east-2.amazonaws.com/rrdp/bd48a1fa-3471-4ab2-8508-ad36b96813e4/notification.xml",
		"https://rpki-rrdp.us-east-2.amazonaws.com/rrdp/c3cd7c24-12cb-4abc-8fd2-5e2bcbb85ae6/notification.xml",
		"https://rpki-rrdp.us-east-2.amazonaws.com/rrdp/db9a372a-09bc-4a32-bfe4-8c48e5dbd219/notification.xml",
		"https://rpki-rrdp.us-east-2.amazonaws.com/rrdp/dba8f01c-9669-44a3-ac6e-db2edb099b84/notification.xml",
		"https://rpki-rrdp.us-east-2.amazonaws.com/rrdp/dfd7f6d3-e6e9-4987-9ae7-d052c5353898/notification.xml",
		"https://rpki-rrdp.us-east-2.amazonaws.com/rrdp/e72d8db0-4728-4fc1-bdd8-471129866362/notification.xml",
		"https://rpki-rrdp.us-east-2.amazonaws.com/rrdp/e7518af5-a343-428d-bf78-f982b6e60505/notification.xml",
		"https://rpki-rrdp.us-east-2.amazonaws.com/rrdp/f703696e-e47b-4c20-bd93-6f80904e42d2/notification.xml",
		"https://rpki-rrdp.us-east-2.amazonaws.com/rrdp/ff9fa84e-9783-4a0b-a58d-6dc8e2433d33/notification.xml",
		"https://rrdp.afrinic.net/notification.xml",
		"https://rrdp.apnic.net/notification.xml",
		"https://rrdp.arin.net/notification.xml",
		"https://rrdp.krill.cloud/notification.xml",
		"https://rrdp.lacnic.net/rrdp/notification.xml",
		"https://rrdp.lacnic.net/rrdpas0/notification.xml",
		"https://rrdp.paas.rpki.ripe.net/notification.xml",
		"https://rrdp.ripe.net/notification.xml",
		"https://rrdp.roa.tohunet.com/rrdp/notification.xml",
		"https://rrdp.rp.ki/notification.xml",
		"https://rrdp.rpki.co/rrdp/notification.xml",
		"https://rrdp.rpki.tianhai.link/rrdp/notification.xml",
		"https://rrdp.sub.apnic.net/notification.xml",
		"https://rrdp.taaa.eu/rrdp/notification.xml",
		"https://rrdp.twnic.tw/rrdp/notify.xml",
		"https://rrdp-as0.apnic.net/notification.xml",
		"https://rrdp-rps.arin.net/notification.xml",
		"https://sakuya.nat.moe/rrdp/notification.xml",
		"https://x-0100000000000011.p.u9sv.com/notification.xml",
		"https://x-8011.p.u9sv.com/notification.xml",
		"https://repo.rpki.space/rrdp/notification.xml",
		"https://rpki.uz/rrdp/notification.xml",
	}

	// 构建自定义下载配置
	downloadConfig := &DownloadTaskConfig{
		SaveDir:    "/tmp/taskcycleutil_downloads",
		Timeout:    70 * time.Minute,
		RetryCount: 1,
	}

	// 构建框架任务
	tasks := make([]*taskcycleutil.Task, 0, len(downloadURLs))
	for _, url := range downloadURLs {
		task := &taskcycleutil.Task{
			Key: url, // 任务唯一Key
			Data: taskcycleutil.TaskData{
				Params: map[string]interface{}{
					"downloadConfig": downloadConfig,
				},
			},
		}
		tasks = append(tasks, task)
	}
	// 可添加更多任务...
	// {
	// 	Key: "download_task_2",
	// 	Data: taskcycleutil.TaskData{
	// 		Content: `{"urls":["https://example.com/file4.zip"], "saveDir":"./downloads"}`,
	// 	},
	// },

	// ========== 6. 启动框架 ==========
	framework.Start()
	belogs.Info("task framework started, waiting for cycle execution...")

	// ========== 7. 批量添加任务 ==========
	successCount, failedTasks := framework.AddTasks(tasks, downloadTaskExecuteFunc)
	belogs.Info(fmt.Sprintf("add tasks result: success=%d, failed=%d", successCount, len(failedTasks)))

	// 打印失败任务信息
	if len(failedTasks) > 0 {
		for _, task := range failedTasks {
			belogs.Error(fmt.Sprintf("add task failed: %s, reason: %s", task.Key, task.FailReason))
		}
	}

	// ========== 8. 保持程序运行 ==========
	// 示例：运行1小时后退出
	runDuration := 1 * time.Hour
	belogs.Info(fmt.Sprintf("framework will run for %v, press Ctrl+C to exit", runDuration))
	time.Sleep(runDuration)

	// ========== 9. 停止框架 ==========
	framework.Stop()
	belogs.Info("task framework stopped")

	/* ========== 10. 输出任务执行结果（可选） ==========
	framework.TasksMu.RLock() // 注：需将框架的tasksMu改为导出字段，或添加获取任务的方法
	defer framework.TasksMu.RUnlock()

	for key, task := range framework.Tasks {
		belogs.Info(fmt.Sprintf("task %s result: %s, success count: %d, fail count: %d",
			key, task.Result, task.SuccessCount, task.FailCount))
	}
	*/
}
