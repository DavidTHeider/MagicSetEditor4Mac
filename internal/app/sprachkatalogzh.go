package app

// **Die Kataloge für Chinesisch (vereinfacht)** — maschinell übersetzt (LOC-01, #1687), Ausgangssprache Deutsch.

import "github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"

var menuetexteZh = Menuetexte{
	Ablage:         "文件",
	Neu:            "新建",
	Oeffnen:        "打开…",
	Sichern:        "存储",
	SichernUnter:   "存储为…",
	Schliessen:     "关闭",
	Drucken:        "打印…",
	BestandSichern: "备份数据库…",
	Einstellungen:  "设置…",
	Zuletzt:        "打开最近使用的项目",
	Hilfe:          "帮助",
	HilfeOeffnen:   "MagicSetEditor4Mac 帮助",

	Set:              "系列",
	Rueckgaengig:     "还原",
	AllesAuswaehlen:  "勾选所有卡牌",
	Wiederherstellen: "恢复",
	Suchen:           "查找…",
	Ersetzen:         "替换…",

	Ausgeben:      "导出",
	Exportvorlage: "使用导出模板…",
	Bilder:        "导出为图像…",
	Druckbogen:    "打印版面…",

	Ueber: "关于 MagicSetEditor4Mac",

	Darstellung:            "显示",
	SeitenleisteEinblenden: "显示边栏",
	SeitenleisteAusblenden: "隐藏边栏",
	NavigatorEinblenden:    "显示导航器",
	NavigatorAusblenden:    "隐藏导航器",
	InspectorEinblenden:    "显示检查器",
	InspectorAusblenden:    "隐藏检查器",
	SeiteZurueck:           "返回",
	SeiteVor:               "前进",

	DarstellungZuruecksetzen: "重置外观",

	Anwendung:        "MagicSetEditor4Mac",
	Dienste:          "服务",
	Ausblenden:       "隐藏 MagicSetEditor4Mac",
	AndereAusblenden: "隐藏其他",
	AlleEinblenden:   "全部显示",
	Beenden:          "退出 MagicSetEditor4Mac",

	Bearbeiten:      "编辑",
	Widerrufen:      "撤销",
	Wiederholen:     "重做",
	Ausschneiden:    "剪切",
	Kopieren:        "拷贝",
	Einsetzen:       "粘贴",
	Loeschen:        "删除",
	AllesMarkieren:  "全选",
	Sprachausgabe:   "语音",
	SprechenStarten: "开始朗读",
	SprechenStoppen: "停止朗读",

	Fenster:       "窗口",
	Minimieren:    "最小化",
	Zoomen:        "缩放",
	AlleNachVorne: "前置全部窗口",
}

var dialogtexteZh = Dialogtexte{
	BildOeffnen: "打开图像",
	BildFilter:  "图像 (*.png, *.jpg, *.gif)",

	PaketdateiEinfuegen: "将文件添加到软件包",
	PaketdateiFilter:    "图像和字体 (*.png, *.jpg, *.gif, *.ttf, *.otf)",

	PaketWeitergeben:  "将软件包共享为",
	InstallerEinlesen: "读取安装程序",
	InstallerFilter:   "MSE 安装程序 (*.mse-installer)",

	TischabzugSpeichern: "将牌桌图像存储为",
	TischabzugFilter:    "PNG 图像 (*.png)",
	TischabzugVorschlag: "牌桌.png",

	TischabzugPDFSpeichern: "将牌桌图像存储为 PDF",
	TischabzugPDFFilter:    "PDF 文件 (*.pdf)",
	TischabzugPDFVorschlag: "牌桌.pdf",

	PartieSpeichern: "导出对局到",
	PartieOeffnen:   "导入对局",
	PartieFilter:    "MSE4Mac 对局 (*.msepartie)",
	PartieVorschlag: "对局.msepartie",

	TurnierSpeichern: "导出赛事到",
	TurnierOeffnen:   "导入赛事",
	TurnierFilter:    "MSE4Mac 赛事 (*.mseturnier)",
	TurnierVorschlag: "赛事.mseturnier",
	BerichtSpeichern: "将赛事报告存储为",
	BerichtFilter:    "Markdown (*.md)",
	BerichtVorschlag: "赛事报告.md",

	SymbolAusgeben:      "导出符号到",
	SymbolEinlesen:      "将符号导入资源库",
	SymbolFilter:        "MSE 符号 (*.mse-symbol)",
	SymbolVorschlag:     "符号.mse-symbol",
	SprachpaketAusgeben: "导出翻译到",
	SprachpaketEinlesen: "导入翻译",
	SprachpaketFilter:   "MSE4Mac 翻译 (*.json)",

	SetOeffnen:   "打开系列",
	SetFilter:    "MSE 系列 (*.mse-set)",
	SetSpeichern: "将系列存储为",
	SetVorschlag: "系列.mse-set",

	UmschlagSchnueren: "将系列存储为信封",
	UmschlagOeffnen:   "查看信封",
	UmschlagFilter:    "MSE4Mac 信封 (*.msepackage)",
	UmschlagVorschlag: "信封.msepackage",
	Uebernommen:       "已导入",

	LeiheAblegen:   "保存借用的信封",
	LeiheVorschlag: "借用.msepackage",

	BestandSpeichern: "备份数据库",
	BestandFilter:    "MSE4Mac 备份 (*.msesicherung)",
	BestandVorschlag: "mse-备份.msesicherung",
	BestandOeffnen:   "选择要恢复的备份",
	KontakteOeffnen:  "选择 vCard 联系人",
	KontakteFilter:   "vCard (*.vcf)",

	DiagnoseSpeichern: "存储诊断信息",
	DiagnoseFilter:    "文本文件 (*.txt)",
	DiagnoseVorschlag: "mse-诊断.txt",

	DiagnoseBestand:      "数据库",
	DiagnoseProtokollort: "日志",
	DiagnoseKanaele:      "通道",
	DiagnoseKanalDatei:   "文件",
	DiagnoseKanalSystem:  "Unified Log",
	DiagnoseKanalAusgabe: "标准错误输出",
	DiagnoseKanalAn:      "开",
	DiagnoseKanalAus:     "关",
	DiagnoseKanalFehlt:   "缺失",
	DiagnoseStufe:        "级别",
	DiagnoseAusfuehrlich: "详细日志已开启。每个文件最多保留 1 MB，另保留一个更早的文件。" +
		"找到问题后，请在“设置 › 诊断”中将其恢复原状。",
	DiagnosepaketSpeichern: "存储诊断包",
	DiagnosepaketVorschlag: "诊断.msediagnose",
	DiagnosepaketFilter:    "诊断包 (*.msediagnose)",
	DiagnoseTeilUnbekannt: "此部分未知：应用程序知道它的名称，" +
		"但不知道它的内容。",
	DiagnoseProtokollVerdeckt: "日志行不在此处显示：它们属于此进程，而不属于此用户档案。" +
		"如需查看，请打开你自己的用户档案，或读取上述位置的文件。",
	DiagnoseSchema:          "架构",
	DiagnoseWaisen:          "孤立行",
	DiagnoseBau:             "构建",
	DiagnoseVeraendert:      "工作树已修改",
	DiagnoseBindungsvertrag: "绑定契约",
	DiagnoseMigrationen:     "%d 次迁移，最新为 %s",
	DiagnoseKeineSchema:     "未记录任何迁移",
	DiagnoseNichtLesbar:     "无法读取",
	DiagnosePakete:          "软件包",
	DiagnoseWeitere:         "… 另有 %d 项",
	DiagnoseKeinBestand:     "未打开任何数据库",
	DiagnoseKeinProtokoll:   "无日志",
	DiagnoseKeineVerwaltung: "未设置软件包管理",

	MTGJSONDatei:       "打开 MTG JSON",
	MTGJSONFilter:      "MTG JSON (*.json)",
	MTGJSONVerzeichnis: "打开包含 MTG JSON 的文件夹",
	MTGJSONSchreiben:   "存储为 MTG JSON",
	MTGJSONVorschlag:   "set.json",

	PersonendatenSichern: "存储我的数据",
	JSONFilter:           "JSON (*.json)",

	EinstellungenSichern:   "存储设置",
	EinstellungenOeffnen:   "导入设置",
	EinstellungenVorschlag: "设置.json",

	VorlageSchreiben: "通过导出模板存储",
	VorlageFilter:    "模板输出",

	Bildordner:     "选择图像所在的文件夹",
	BogenSchreiben: "存储打印版面",
	BogenVorschlag: "打印版面.pdf",
	PDFFilter:      "PDF (*.pdf)",

	PaketWaehlen: "选择模板软件包",
	MSE2Waehlen:  "选择 MSE2 模板所在的文件夹",

	Planordner: "选择每周备份的文件夹",
}

var berichttexteZh = domain.Berichttexte{
	Spiel: "游戏", Format: "赛制", Stand: "状态", Abgeschlossen: "结束于",
	Laufend: "进行中", Teilnehmende: "参赛者", Ausgestiegen: "已退出", Runden: "轮数",
	Rangliste: "排名", Platz: "名次", Name: "名称", Punkte: "积分", Bilanz: "战绩",
	Gegnermatchquote: "OMW", Spielquote: "GW", Gegnerspielquote: "OGW",
	Runde: "轮次", Tisch: "桌号", SeiteA: "选手 A", SeiteB: "选手 B", Ergebnis: "结果",
	Bye: "轮空", Offen: "待定",
}
