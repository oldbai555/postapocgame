package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

type Config struct {
	Group       string // 功能组名（如 user, file）
	Name        string // 功能名称（如 用户管理, 文件管理）
	OutputDir   string // 输出目录
	TemplateDir string // 模板目录
}

type TemplateData struct {
	Group         string // 功能组名（如 user）
	Name          string // 功能名称（如 用户管理）
	GroupUpper    string // 大写的组名（如 User）
	GroupLower    string // 小写的组名（如 user，用于 API 对象名）
	GroupFuncName string // 函数名（首字母小写，如 fileList）
	Path          string // 前端路径（如 /temp/user）
	Component     string // 前端组件路径（如 temp/UserList）
	APIBasePath   string // API 基础路径（如 /api/v1/users）
}

func main() {
	var config Config
	flag.StringVar(&config.Group, "group", "", "功能组名（必需，如 user, file）")
	flag.StringVar(&config.Name, "name", "", "功能名称（必需，如 用户管理, 文件管理）")
	flag.StringVar(&config.OutputDir, "output", "", "输出目录（可选，默认 admin-server/db）")
	flag.StringVar(&config.TemplateDir, "template", "", "模板目录（可选，默认 scripts/sqlgen/templates）")
	flag.Parse()

	if config.Group == "" || config.Name == "" {
		fmt.Fprintf(os.Stderr, "错误: 必须提供 -group 和 -name 参数\n")
		fmt.Fprintf(os.Stderr, "用法: %s -group <group> -name <name>\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "示例: %s -group user -name 用户管理\n", os.Args[0])
		os.Exit(1)
	}

	// 获取当前工作目录
	workDir, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "错误: 无法获取工作目录: %v\n", err)
		os.Exit(1)
	}

	// 获取项目根目录（从当前目录向上查找，直到找到包含 admin-server 和 admin-frontend 的目录）
	projectRoot := workDir
	for {
		// 检查是否包含 admin-server 和 admin-frontend 目录
		if _, err := os.Stat(filepath.Join(projectRoot, "admin-server")); err == nil {
			if _, err := os.Stat(filepath.Join(projectRoot, "admin-frontend")); err == nil {
				break
			}
		}
		// 如果已经到达根目录，退出
		parent := filepath.Dir(projectRoot)
		if parent == projectRoot {
			fmt.Fprintf(os.Stderr, "错误: 无法找到项目根目录（应包含 admin-server 和 admin-frontend）\n")
			os.Exit(1)
		}
		projectRoot = parent
	}

	// 设置默认输出目录
	if config.OutputDir == "" {
		config.OutputDir = filepath.Join(projectRoot, "admin-server", "db")
	}

	// 设置默认模板目录
	if config.TemplateDir == "" {
		config.TemplateDir = filepath.Join(projectRoot, "scripts", "sqlgen", "templates")
	}

	// 确保输出目录存在
	if err := os.MkdirAll(config.OutputDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "错误: 无法创建输出目录: %v\n", err)
		os.Exit(1)
	}

	// 准备模板数据
	data := prepareTemplateData(config.Group, config.Name)

	// 生成建表 SQL 文件
	createTableFile := filepath.Join(config.OutputDir, fmt.Sprintf("create_table_%s.sql", config.Group))
	if err := generateCreateTableSQL(config.TemplateDir, data, createTableFile); err != nil {
		fmt.Fprintf(os.Stderr, "错误: 生成建表 SQL 文件失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✓ 建表 SQL 文件生成成功: %s\n", createTableFile)

	// 生成初始化 SQL 文件
	outputFile := filepath.Join(config.OutputDir, fmt.Sprintf("init_%s.sql", config.Group))
	if err := generateSQL(config.TemplateDir, data, outputFile); err != nil {
		fmt.Fprintf(os.Stderr, "错误: 生成初始化 SQL 文件失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✓ 初始化 SQL 文件生成成功: %s\n", outputFile)

	// 生成 .api 文件
	apiOutputDir := filepath.Join(projectRoot, "admin-server", "api")
	if err := os.MkdirAll(apiOutputDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "错误: 无法创建 API 输出目录: %v\n", err)
		os.Exit(1)
	}

	apiOutputFile := filepath.Join(apiOutputDir, fmt.Sprintf("%s.api.temp", config.Group))
	if err := generateAPIFile(config.TemplateDir, data, apiOutputFile); err != nil {
		fmt.Fprintf(os.Stderr, "错误: 生成 .api 文件失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✓ .api 文件生成成功: %s\n", apiOutputFile)
	fmt.Printf("  请将内容追加到 admin-server/api/admin.api\n")

	// 生成前端 Vue 页面
	vueOutputDir := filepath.Join(projectRoot, "admin-frontend", "src", "views", "temp")
	if err := os.MkdirAll(vueOutputDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "错误: 无法创建 Vue 页面输出目录: %v\n", err)
		os.Exit(1)
	}

	vueOutputFile := filepath.Join(vueOutputDir, fmt.Sprintf("%sList.vue", data.GroupUpper))
	if err := generateVuePage(config.TemplateDir, data, vueOutputFile); err != nil {
		fmt.Fprintf(os.Stderr, "错误: 生成 Vue 页面失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✓ Vue 页面生成成功: %s\n", vueOutputFile)
}

// prepareTemplateData 准备模板数据
func prepareTemplateData(group, name string) TemplateData {
	// 将 group 转换为首字母大写（用于组件名）
	groupUpper := strings.ToUpper(group[:1]) + group[1:]
	// 如果 group 包含下划线，需要处理（如 user_role -> UserRole）
	if strings.Contains(group, "_") {
		parts := strings.Split(group, "_")
		var upperParts []string
		for _, part := range parts {
			upperParts = append(upperParts, strings.ToUpper(part[:1])+part[1:])
		}
		groupUpper = strings.Join(upperParts, "")
	}

	// 前端路径：/temp/{group}（临时目录下）
	path := fmt.Sprintf("/temp/%s", group)
	// 前端组件路径：temp/{GroupUpper}List
	component := fmt.Sprintf("temp/%sList", groupUpper)
	// API 基础路径：/api/v1/{group}s（复数形式）
	apiBasePath := fmt.Sprintf("/api/v1/%ss", group)
	// API 对象名（小写，如 fileApi）
	groupLower := strings.ToLower(group)
	// 函数名（首字母小写，如 fileList）
	groupFuncName := strings.ToLower(group[:1]) + group[1:]

	return TemplateData{
		Group:         group,
		Name:          name,
		GroupUpper:    groupUpper,
		GroupLower:    groupLower,
		GroupFuncName: groupFuncName,
		Path:          path,
		Component:     component,
		APIBasePath:   apiBasePath,
	}
}

// generateCreateTableSQL 生成建表 SQL 文件
func generateCreateTableSQL(templateDir string, data TemplateData, outputFile string) error {
	// 读取模板文件
	templateFile := filepath.Join(templateDir, "create_table.sql.tpl")
	tmpl, err := template.New("create_table.sql.tpl").ParseFiles(templateFile)
	if err != nil {
		return fmt.Errorf("无法读取模板文件: %v", err)
	}

	// 创建输出文件（使用 UTF-8 编码，不写入 BOM，避免 goctl 解析问题）
	file, err := os.Create(outputFile)
	if err != nil {
		return fmt.Errorf("无法创建输出文件: %v", err)
	}
	defer file.Close()

	// 执行模板
	if err := tmpl.Execute(file, data); err != nil {
		return fmt.Errorf("模板执行失败: %v", err)
	}

	return nil
}

// generateSQL 生成 SQL 文件
func generateSQL(templateDir string, data TemplateData, outputFile string) error {
	// 读取模板文件
	templateFile := filepath.Join(templateDir, "init_module.sql.tpl")
	tmpl, err := template.New("init_module.sql.tpl").ParseFiles(templateFile)
	if err != nil {
		return fmt.Errorf("无法读取模板文件: %v", err)
	}

	// 创建输出文件（使用 UTF-8 编码）
	file, err := os.Create(outputFile)
	if err != nil {
		return fmt.Errorf("无法创建输出文件: %v", err)
	}
	defer file.Close()

	// 写入 UTF-8 BOM（可选，确保某些编辑器正确识别编码）
	file.WriteString("\xEF\xBB\xBF")

	// 执行模板
	if err := tmpl.Execute(file, data); err != nil {
		return fmt.Errorf("模板执行失败: %v", err)
	}

	return nil
}

// generateAPIFile 生成 .api 文件
func generateAPIFile(templateDir string, data TemplateData, outputFile string) error {
	// 读取模板文件
	templateFile := filepath.Join(templateDir, "init_module.api.tpl")
	tmpl, err := template.New("init_module.api.tpl").ParseFiles(templateFile)
	if err != nil {
		return fmt.Errorf("无法读取模板文件: %v", err)
	}

	// 创建输出文件（使用 UTF-8 编码）
	file, err := os.Create(outputFile)
	if err != nil {
		return fmt.Errorf("无法创建输出文件: %v", err)
	}
	defer file.Close()

	// 写入 UTF-8 BOM（可选，确保某些编辑器正确识别编码）
	file.WriteString("\xEF\xBB\xBF")

	// 执行模板
	if err := tmpl.Execute(file, data); err != nil {
		return fmt.Errorf("模板执行失败: %v", err)
	}

	return nil
}

// generateVuePage 生成 Vue 页面文件
func generateVuePage(templateDir string, data TemplateData, outputFile string) error {
	// 读取模板文件
	templateFile := filepath.Join(templateDir, "list_page.vue.tpl")
	tmpl, err := template.New("list_page.vue.tpl").ParseFiles(templateFile)
	if err != nil {
		return fmt.Errorf("无法读取模板文件: %v", err)
	}

	// 创建输出文件（使用 UTF-8 编码）
	file, err := os.Create(outputFile)
	if err != nil {
		return fmt.Errorf("无法创建输出文件: %v", err)
	}
	defer file.Close()

	// 写入 UTF-8 BOM（可选，确保某些编辑器正确识别编码）
	file.WriteString("\xEF\xBB\xBF")

	// 执行模板
	if err := tmpl.Execute(file, data); err != nil {
		return fmt.Errorf("模板执行失败: %v", err)
	}

	return nil
}
