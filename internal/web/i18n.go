package web

import "strings"

const (
	langZH = "zh-CN"
	langEN = "en"
)

var messages = map[string]map[string]string{
	langZH: {
		"nav.list":                 "列表",
		"nav.new":                  "发布",
		"share.kind.file":          "文件",
		"share.kind.text":          "文本",
		"share.kind.link":          "链接",
		"share.visibility.public":  "公开",
		"share.visibility.private": "私有",
		"share.pin":                "置顶",
		"share.unpin":              "取消置顶",
		"share.delete":             "删除",
		"share.delete.confirm":     "确认删除这个分享吗？",
		"share.expire.prefix":      "到期",
		"share.expire.never":       "永不过期",
		"list.title":               "公开分享列表",
		"list.description":         "支持文件下载、复制文本、复制链接和打开链接。置顶内容会优先展示。",
		"list.new":                 "新建分享",
		"list.empty.title":         "还没有公开内容",
		"list.empty.description":   "从发布页上传文件、分享文本或者链接。",
		"list.empty.cta":           "去发布",
		"list.action.download":     "下载文件",
		"list.action.copyDownload": "复制下载链接",
		"list.action.copyText":     "复制文本",
		"list.action.copyURL":      "复制 URL",
		"list.action.openURL":      "打开 URL",
		"list.action.detail":       "查看详情",
		"new.pageTitle":            "新建分享",
		"new.title":                "新建分享",
		"new.description":          "默认公开、置顶和过期时间来自启动参数。",
		"new.compose.title":        "发布内容",
		"new.compose.hint":         "输入文本，或把文件拖拽、粘贴到文本框中。文本 trim 后如果只剩一个 URL，会自动按链接分享。",
		"new.compose.hint.single":  "输入文本，或把文件拖拽、粘贴到文本框中。选择文件后会自动上传。文本 trim 后如果只剩一个 URL，会自动按链接分享。",
		"new.compose.hint.multi":   "输入文本，或把文件拖拽、粘贴到文本框中。可以选择多个文件一起上传。文本 trim 后如果只剩一个 URL，会自动按链接分享。",
		"new.compose.label":        "内容",
		"new.compose.placeholder":  "输入文本，或拖拽 / 粘贴文件到这里",
		"new.compose.pick":         "选择文件",
		"new.compose.upload":       "上传",
		"new.compose.public":       "Public",
		"new.compose.expire":       "过期时间",
		"detail.createdAt":         "创建时间：",
		"detail.path":              "详情地址：",
		"detail.originalName":      "原文件名：",
		"detail.size":              "大小：",
		"detail.mime":              "MIME：",
		"detail.action.download":   "下载文件",
		"detail.action.copyDownload": "复制下载链接",
		"detail.action.copyText":   "复制文本",
		"detail.action.copyURL":    "复制 URL",
		"detail.action.openURL":    "打开 URL",
		"detail.files.count":       "个文件",
		"detail.download.all":      "下载全部",
	},
	langEN: {
		"nav.list":                 "List",
		"nav.new":                  "New",
		"share.kind.file":          "File",
		"share.kind.text":          "Text",
		"share.kind.link":          "Link",
		"share.visibility.public":  "Public",
		"share.visibility.private": "Private",
		"share.pin":                "Pinned",
		"share.unpin":              "Unpin",
		"share.delete":             "Delete",
		"share.delete.confirm":     "Delete this share?",
		"share.expire.prefix":      "Expires",
		"share.expire.never":       "Never expires",
		"list.title":               "Public Shares",
		"list.description":         "Download files, copy text, copy links, or open shared URLs. Pinned items stay on top.",
		"list.new":                 "New Share",
		"list.empty.title":         "No public shares yet",
		"list.empty.description":   "Upload a file or share text or a link from the publish page.",
		"list.empty.cta":           "Create one",
		"list.action.download":     "Download",
		"list.action.copyDownload": "Copy download link",
		"list.action.copyText":     "Copy text",
		"list.action.copyURL":      "Copy URL",
		"list.action.openURL":      "Open URL",
		"list.action.detail":       "Details",
		"new.pageTitle":            "New Share",
		"new.title":                "New Share",
		"new.description":          "Default visibility, pinning, and expiration come from startup flags.",
		"new.compose.title":        "Share Content",
		"new.compose.hint":         "Type text or drag and paste a file into the box. If the trimmed text is a single URL, it is shared as a link.",
		"new.compose.hint.single":  "Type text or drag and paste a file into the box. Files are uploaded automatically when selected. If the trimmed text is a single URL, it is shared as a link.",
		"new.compose.hint.multi":   "Type text or drag and paste files into the box. You can select multiple files to upload together. If the trimmed text is a single URL, it is shared as a link.",
		"new.compose.label":        "Content",
		"new.compose.placeholder":  "Type text, or drag / paste a file here",
		"new.compose.pick":         "Choose file",
		"new.compose.upload":       "Upload",
		"new.compose.public":       "Public",
		"new.compose.expire":       "Expire",
		"detail.createdAt":         "Created at: ",
		"detail.path":              "Share path: ",
		"detail.originalName":      "Original name: ",
		"detail.size":              "Size: ",
		"detail.mime":              "MIME: ",
		"detail.action.download":   "Download",
		"detail.action.copyDownload": "Copy download link",
		"detail.action.copyText":   "Copy text",
		"detail.action.copyURL":    "Copy URL",
		"detail.action.openURL":    "Open URL",
		"detail.files.count":       "files",
		"detail.download.all":      "Download All",
	},
}

func detectLanguage(header string) string {
	for _, part := range strings.Split(header, ",") {
		lang := strings.TrimSpace(strings.SplitN(part, ";", 2)[0])
		if lang == "" {
			continue
		}
		lang = strings.ToLower(lang)
		if strings.HasPrefix(lang, "en") {
			return langEN
		}
		if strings.HasPrefix(lang, "zh") {
			return langZH
		}
	}
	return langZH
}

func translate(lang, key string) string {
	if catalog, ok := messages[lang]; ok {
		if value, ok := catalog[key]; ok {
			return value
		}
	}
	if value, ok := messages[langZH][key]; ok {
		return value
	}
	return key
}

func kindLabel(lang, kind string) string {
	switch kind {
	case "file":
		return translate(lang, "share.kind.file")
	case "text":
		return translate(lang, "share.kind.text")
	case "link":
		return translate(lang, "share.kind.link")
	default:
		return kind
	}
}
