package gui

import (
	"fmt"
	"image/color"
	"net/url"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// 创建可选择的只读文本框
func newSelectableLabel(text string) *widget.Entry {
	entry := widget.NewMultiLineEntry()
	entry.SetText(text)
	entry.Disable()
	// 自定义样式：使用正常的文本颜色
	entry.TextStyle = fyne.TextStyle{}
	entry.Wrapping = fyne.TextWrapWord

	// 重写默认的禁用样式
	customEntry := entry
	customEntry.OnSubmitted = func(string) {} // 防止编辑
	customEntry.Disable()

	// 使用自定义资源设置正常的文本颜色
	customEntry.Refresh()
	return customEntry
}

// 创建可滚动到底部的多行文本框
func newScrollableLabel(text string) *widget.Entry {
	entry := widget.NewMultiLineEntry()
	entry.SetText(text)
	entry.Disable()
	entry.TextStyle = fyne.TextStyle{}
	entry.Wrapping = fyne.TextWrapWord

	// 使用 OnChanged 回调确保文本更新时滚动到底部
	entry.OnChanged = func(string) {
		// 使用 goroutine 确保在渲染完成后滚动
		go func() {
			time.Sleep(100 * time.Millisecond)
			entry.CursorRow = len(strings.Split(entry.Text, "\n")) - 1
			entry.Refresh()
		}()
	}

	// 触发一次 OnChanged 以执行初始滚动
	entry.OnChanged(entry.Text)
	return entry
}

func CreateWindow() {
	myApp := app.New()
	// 设置自定义主题以确保禁用状态下的文本仍然清晰可见
	myApp.Settings().SetTheme(&customTheme{theme.DefaultTheme()})
	myWindow := myApp.NewWindow("Fyne Components Demo")

	// 基础输入组件示例
	basicInputs := container.NewVBox(
		widget.NewLabel("Basic Input Components:"),
		widget.NewEntry(),
		widget.NewPasswordEntry(),
		widget.NewMultiLineEntry(),
		widget.NewCheck("Checkbox", func(checked bool) {
			fmt.Println("Checked:", checked)
		}),
	)

	// 按钮组件示例
	buttons := container.NewVBox(
		widget.NewLabel("Button Components:"),
		widget.NewButton("Standard Button", func() {
			fmt.Println("Button tapped")
		}),
		widget.NewButtonWithIcon("Icon Button", theme.HomeIcon(), func() {
			fmt.Println("Icon button tapped")
		}),
	)

	// 选择组件示例
	selections := container.NewVBox(
		widget.NewLabel("Selection Components:"),
		widget.NewSelect([]string{"Option 1", "Option 2", "Option 3"}, func(s string) {
			fmt.Println("Selected:", s)
		}),
		widget.NewRadioGroup([]string{"Radio 1", "Radio 2"}, func(s string) {
			fmt.Println("Radio selected:", s)
		}),
	)

	// 进度指示器示例
	progress := container.NewVBox(
		widget.NewLabel("Progress Indicators:"),
		widget.NewProgressBar(),
		widget.NewProgressBarInfinite(),
	)

	// 数据绑定示例
	data := binding.NewString()
	data.Set("Binding Demo")
	bindingDemo := container.NewVBox(
		widget.NewLabel("Data Binding:"),
		widget.NewEntryWithData(data),
		widget.NewLabelWithData(data),
	)

	// 文本组件示例
	scrollableText := newScrollableLabel(`This is a long multi-line text
Line 2
Line 3
Line 4
Line 5
Line 6
Line 7
Line 8
Line 9
Last line - This will be visible initially`)

	textComponents := container.NewVBox(
		widget.NewButton("Scroll To End", func() {
			// 手动触发滚动到底部
			lineCount := len(strings.Split(scrollableText.Text, "\n"))
			scrollableText.CursorRow = lineCount - 1
			scrollableText.Refresh()
		}),
		widget.NewLabel("Text Components:"),
		widget.NewLabel("Simple Label (Not Selectable)"),
		newSelectableLabel("This is a selectable text with better visibility"),
		scrollableText, // 使用已存储的引用
		widget.NewRichTextFromMarkdown("**Rich Text** with *Markdown* (Not Selectable)"),
		widget.NewHyperlink("Clickable Link", parseURL("https://fyne.io")),
		widget.NewTextGrid(),
		container.NewHBox(
			widget.NewIcon(theme.DocumentIcon()),
			widget.NewLabelWithStyle("Styled Label", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		),
	)

	// 使用选项卡组织所有组件
	tabs := container.NewAppTabs(
		container.NewTabItem("Basic Inputs", basicInputs),
		container.NewTabItem("Buttons", buttons),
		container.NewTabItem("Selections", selections),
		container.NewTabItem("Progress", progress),
		container.NewTabItem("Data Binding", bindingDemo),
		container.NewTabItem("Text", textComponents), // 添加新的文本组件选项卡
	)

	myWindow.SetContent(tabs)
	myWindow.Resize(fyne.NewSize(400, 500))

	// 启动一个goroutine来演示进度条
	go func() {
		if p, ok := progress.Objects[1].(*widget.ProgressBar); ok {
			for i := 0.0; i <= 1.0; i += 0.1 {
				time.Sleep(time.Second)
				p.SetValue(i)
			}
		}
	}()

	myWindow.ShowAndRun()
}

// 辅助函数：解析URL
func parseURL(urlStr string) *url.URL {
	link, _ := url.Parse(urlStr)
	return link
}

// 自定义主题来覆盖禁用状态的文本颜色
type customTheme struct {
	fyne.Theme
}

func (t *customTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	switch name {
	case theme.ColorNameDisabled:
		// Entry禁用状态显示浅灰色文字
		return color.NRGBA{R: 200, G: 200, B: 200, A: 255}
	case theme.ColorNameInputBackground:
		// Entry的背景色为黑色
		return color.NRGBA{R: 0, G: 0, B: 0, A: 255}
	case theme.ColorNamePlaceHolder, theme.ColorNameInputBorder:
		// Entry的占位符文字和边框颜色
		return color.NRGBA{R: 160, G: 160, B: 160, A: 255}
	default:
		// 其他所有颜色保持默认
		return t.Theme.Color(name, variant)
	}
}
