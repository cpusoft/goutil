package ginserver

import (
	"fmt"
	"mime/multipart"
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestRunTLSEx(t *testing.T) {
	engine := gin.New()
	engine.GET("/user/:name/*action", func(c *gin.Context) {
		name := c.Param("name")
		action := c.Param("action")
		message := name + " is " + action
		fmt.Println(message)
		c.String(http.StatusOK, message)
	})

	// change to actual path
	certFile := `..\conf\cert\server.crt`
	keyFile := `..\conf\cert\server.key`
	fmt.Println(certFile, keyFile)
	err := RunTLSEx(engine, ":7771", certFile, keyFile)
	fmt.Println(err)
}

func TestUploadFile(t *testing.T) {
	r := gin.Default()
	r.MaxMultipartMemory = 50 << 20 // 50MB

	// 调试接口，查看所有上传的字段
	r.POST("/debug-upload", debugUploadHandler)

	// 动态处理任何文件字段
	r.POST("/upload", flexibleUploadHandler)

	r.POST("/uploadfile", uploadFile)

	fmt.Println("服务器启动在 :8080")
	r.Run(":8070")
	select {}
}

// 调试处理器，查看所有上传的字段
func debugUploadHandler(c *gin.Context) {
	form, err := c.MultipartForm()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "解析表单失败: " + err.Error(),
		})
		return
	}

	// 打印所有文件字段
	fileFields := make(map[string]interface{})
	for fieldName, files := range form.File {
		fmt.Println("fieldName:", fieldName, "  files:", files)
		fileInfo := make([]gin.H, 0)
		for _, file := range files {
			fileInfo = append(fileInfo, gin.H{
				"filename": file.Filename,
				"size":     file.Size,
			})
			fmt.Println("file.Filename:", file.Filename, "  file.Size:", file.Size)
		}
		fileFields[fieldName] = fileInfo
	}

	// 打印所有普通表单字段
	textFields := make(map[string]interface{})
	for key, values := range form.Value {
		fmt.Println("key:", key, "  values:", values)
		if len(values) > 0 {
			textFields[key] = values
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message":     "调试信息",
		"file_fields": fileFields,
		"text_fields": textFields,
	})
}

// 灵活的文件上传处理器，自动检测文件字段
func flexibleUploadHandler(c *gin.Context) {
	form, err := c.MultipartForm()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "解析表单失败: " + err.Error(),
		})
		return
	}

	// 查找所有文件字段
	if len(form.File) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":            "没有找到任何文件字段",
			"available_fields": getAvailableFields(form),
		})
		return
	}

	uploadDir := "./uploads"
	var results []gin.H

	// 遍历所有文件字段
	for fieldName, files := range form.File {
		for _, file := range files {
			// 保存文件
			filename := fmt.Sprintf("%s_%s", fieldName, file.Filename)
			fullPath := fmt.Sprintf("%s/%s", uploadDir, filename)

			if err := c.SaveUploadedFile(file, fullPath); err != nil {
				results = append(results, gin.H{
					"field":    fieldName,
					"filename": file.Filename,
					"status":   "失败",
					"error":    err.Error(),
				})
				continue
			}

			results = append(results, gin.H{
				"field":    fieldName,
				"filename": file.Filename,
				"status":   "成功",
				"size":     file.Size,
				"saved_as": filename,
			})
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message":     "文件上传完成",
		"results":     results,
		"total_files": len(results),
	})
}

func getAvailableFields(form *multipart.Form) []string {
	var fields []string
	for fieldName := range form.File {
		fields = append(fields, fieldName)
	}
	return fields
}

func uploadFile(c *gin.Context) {

	receiveFile, err := ReceiveFile(c, "/tmp")
	fmt.Println("ParseValidateFile(): ReceiveFile, receiveFile:", receiveFile, err)

}
