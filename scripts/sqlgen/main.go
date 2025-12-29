package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"text/template"
	"unicode/utf8"

	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

type Config struct {
	Group       string // 功能组名（如 user, file）
	Name        string // 功能名称（如 用户管理, 文件管理）
	OutputDir   string // 输出目录
	TemplateDir string // 模板目录
	ParentID    string // 父菜单ID（可选，不填则使用父目录路径查找或临时目录）
	ParentPath  string // 前端父目录路径（如 /system，默认 /temp）
}

type TemplateData struct {
	Group         string // 功能组名（如 user）
	Name          string // 功能名称（如 用户管理）
	GroupUpper    string // 大写的组名（如 User）
	GroupLower    string // 小写的组名（如 user，用于 API 对象名）
	GroupFuncName string // 函数名（首字母小写，如 fileList）
	Path          string // 前端路径（如 /system/user 或 /temp/user）
	Component     string // 前端组件路径（如 system/UserList 或 temp/UserList）
	APIBasePath   string // API 基础路径（如 /api/v1/users）
	ParentID      string // 父菜单ID（字符串形式，用于 SQL 模板）
	ParentPath    string // 前端父目录路径（如 /system 或 /temp）
}

// fixEncoding 修复 Windows 环境下的编码问题
// PowerShell 传递的中文参数可能是 GBK 编码，需要转换为 UTF-8
// 如果字符串已经是乱码（如 "婕旂ず鍔熻兘"），尝试反向转换
func fixEncoding(s string) string {
	if s == "" {
		return s
	}

	// 如果字符串不是有效的 UTF-8，尝试从 GBK/GB18030 转换
	if !utf8.ValidString(s) {
		decoder := simplifiedchinese.GB18030.NewDecoder()
		result, _, err := transform.String(decoder, s)
		if err == nil && utf8.ValidString(result) {
			return result
		}
		return s
	}

	// 如果字符串是有效的 UTF-8，检查是否包含中文字符
	hasChinese := false
	for _, r := range s {
		if r >= 0x4E00 && r <= 0x9FFF {
			hasChinese = true
			break
		}
	}

	// 如果包含中文，直接返回
	if hasChinese {
		return s
	}

	// 如果不包含中文但字符串看起来像是乱码（包含日文假名等），尝试反向转换
	// 将字符串的 UTF-8 字节序列当作 GBK 字节序列，然后解码
	// 注意：这种方法可能不总是有效，但可以处理常见的编码错误
	encoder := simplifiedchinese.GB18030.NewEncoder()
	gbkBytes, _, err := transform.Bytes(encoder, []byte(s))
	if err == nil {
		decoder := simplifiedchinese.GB18030.NewDecoder()
		result, _, err := transform.Bytes(decoder, gbkBytes)
		if err == nil {
			resultStr := string(result)
			// 检查转换后的结果是否包含中文
			for _, r := range resultStr {
				if r >= 0x4E00 && r <= 0x9FFF {
					return resultStr
				}
			}
		}
	}

	// 如果所有转换都失败，返回原字符串
	return s
}

func main() {
	var config Config
	flag.StringVar(&config.Group, "group", "", "功能组名（必需，如 user, file）")
	flag.StringVar(&config.Name, "name", "", "功能名称（必需，如 用户管理, 文件管理）")
	flag.StringVar(&config.OutputDir, "output", "", "输出目录（可选，默认 admin-server/db）")
	flag.StringVar(&config.TemplateDir, "template", "", "模板目录（可选，默认 scripts/sqlgen/templates）")
	flag.StringVar(&config.ParentID, "parent-id", "", "父菜单ID（可选，不填则根据父目录路径查找，默认使用临时目录）")
	flag.StringVar(&config.ParentPath, "parent-path", "", "前端父目录路径（如 /system，默认 /temp）")
	flag.Parse()

	// 在 Windows 环境下修复编码问题
	if runtime.GOOS == "windows" {
		originalName := config.Name
		config.Name = fixEncoding(config.Name)
		config.Group = fixEncoding(config.Group)
		config.OutputDir = fixEncoding(config.OutputDir)
		config.TemplateDir = fixEncoding(config.TemplateDir)
		config.ParentID = fixEncoding(config.ParentID)
		config.ParentPath = fixEncoding(config.ParentPath)

		// 如果编码修复后名称发生变化，输出提示
		if originalName != "" && originalName != config.Name {
			fmt.Fprintf(os.Stderr, "提示: 检测到编码问题，已自动修复参数编码\n")
			fmt.Fprintf(os.Stderr, "  原始参数: %q\n", originalName)
			fmt.Fprintf(os.Stderr, "  修复后参数: %q\n", config.Name)
		}
	}

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
		// 默认输出到 migrations 目录，便于区分增量 SQL
		config.OutputDir = filepath.Join(projectRoot, "admin-server", "db", "migrations")
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

	// 如果未指定父目录路径，默认使用临时目录 /temp
	if config.ParentPath == "" {
		config.ParentPath = "/temp"
	}

	// 准备模板数据
	data := prepareTemplateData(config.Group, config.Name, config.ParentID, config.ParentPath)

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
func prepareTemplateData(group, name, parentID, parentPath string) TemplateData {
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

	// 规范化父目录路径，默认 /temp
	if parentPath == "" {
		parentPath = "/temp"
	}
	if !strings.HasPrefix(parentPath, "/") {
		parentPath = "/" + parentPath
	}

	// 前端路径：{parentPath}/{group}，例如 /system/user 或 /temp/user
	path := fmt.Sprintf("%s/%s", strings.TrimRight(parentPath, "/"), group)

	// 前端组件路径：{baseDir}/{GroupUpper}List，其中 baseDir 为父路径的最后一段（如 system 或 temp）
	baseDir := strings.Trim(parentPath, "/")
	if strings.Contains(baseDir, "/") {
		parts := strings.Split(baseDir, "/")
		baseDir = parts[len(parts)-1]
	}
	if baseDir == "" {
		baseDir = "temp"
	}
	component := fmt.Sprintf("%s/%sList", baseDir, groupUpper)
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
		ParentID:      parentID,
		ParentPath:    parentPath,
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

	// 创建输出文件（使用 UTF-8 编码，不写入 BOM，避免编码问题）
	file, err := os.Create(outputFile)
	if err != nil {
		return fmt.Errorf("无法创建输出文件: %v", err)
	}
	defer file.Close()

	// 执行模板（直接写入，不添加 BOM）
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

	// 创建输出文件（使用 UTF-8 编码，不写入 BOM）
	file, err := os.Create(outputFile)
	if err != nil {
		return fmt.Errorf("无法创建输出文件: %v", err)
	}
	defer file.Close()

	// 执行模板（直接写入，不添加 BOM）
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

	// 创建输出文件（使用 UTF-8 编码，不写入 BOM）
	file, err := os.Create(outputFile)
	if err != nil {
		return fmt.Errorf("无法创建输出文件: %v", err)
	}
	defer file.Close()

	// 执行模板（直接写入，不添加 BOM）
	if err := tmpl.Execute(file, data); err != nil {
		return fmt.Errorf("模板执行失败: %v", err)
	}

	return nil
}
