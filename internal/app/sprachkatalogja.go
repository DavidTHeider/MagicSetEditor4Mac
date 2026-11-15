package app

// **Die Kataloge für Japanisch** — maschinell übersetzt (LOC-01, #1687), Ausgangssprache Deutsch.

import "github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"

var menuetexteJa = Menuetexte{
	Ablage:         "ファイル",
	Neu:            "新規",
	Oeffnen:        "開く…",
	Sichern:        "保存",
	SichernUnter:   "別名で保存…",
	Schliessen:     "閉じる",
	Drucken:        "プリント…",
	BestandSichern: "データベースをバックアップ…",
	Einstellungen:  "設定…",
	Zuletzt:        "最近使った項目",
	Hilfe:          "ヘルプ",
	HilfeOeffnen:   "MagicSetEditor4Mac ヘルプ",

	Set:              "セット",
	Rueckgaengig:     "元に戻す",
	AllesAuswaehlen:  "すべてのカードにチェックを付ける",
	Wiederherstellen: "やり直し",
	Suchen:           "検索…",
	Ersetzen:         "置換…",

	Ausgeben:      "書き出す",
	Exportvorlage: "書き出しテンプレートを使用…",
	Bilder:        "画像として…",
	Druckbogen:    "印刷シート…",

	Ueber: "MagicSetEditor4Macについて",

	Darstellung:            "表示",
	SeitenleisteEinblenden: "サイドバーを表示",
	SeitenleisteAusblenden: "サイドバーを隠す",
	NavigatorEinblenden:    "ナビゲータを表示",
	NavigatorAusblenden:    "ナビゲータを隠す",
	InspectorEinblenden:    "インスペクタを表示",
	InspectorAusblenden:    "インスペクタを隠す",
	SeiteZurueck:           "戻る",
	SeiteVor:               "進む",

	DarstellungZuruecksetzen: "外観をリセット",

	Anwendung:        "MagicSetEditor4Mac",
	Dienste:          "サービス",
	Ausblenden:       "MagicSetEditor4Macを非表示",
	AndereAusblenden: "ほかを非表示",
	AlleEinblenden:   "すべてを表示",
	Beenden:          "MagicSetEditor4Macを終了",

	Bearbeiten:      "編集",
	Widerrufen:      "取り消す",
	Wiederholen:     "やり直す",
	Ausschneiden:    "カット",
	Kopieren:        "コピー",
	Einsetzen:       "ペースト",
	Loeschen:        "削除",
	AllesMarkieren:  "すべてを選択",
	Sprachausgabe:   "スピーチ",
	SprechenStarten: "読み上げを開始",
	SprechenStoppen: "読み上げを停止",

	Fenster:       "ウインドウ",
	Minimieren:    "しまう",
	Zoomen:        "拡大/縮小",
	AlleNachVorne: "すべてを手前に移動",
}

var dialogtexteJa = Dialogtexte{
	BildOeffnen: "画像を開く",
	BildFilter:  "画像 (*.png, *.jpg, *.gif)",

	PaketdateiEinfuegen: "パッケージにファイルを追加",
	PaketdateiFilter:    "画像とフォント (*.png, *.jpg, *.gif, *.ttf, *.otf)",

	PaketWeitergeben:  "パッケージを共有する名前",
	InstallerEinlesen: "インストーラを読み込む",
	InstallerFilter:   "MSEインストーラ (*.mse-installer)",

	TischabzugSpeichern: "テーブルの画像を別名で保存",
	TischabzugFilter:    "PNG画像 (*.png)",
	TischabzugVorschlag: "テーブル.png",

	TischabzugPDFSpeichern: "テーブルの画像をPDFとして保存",
	TischabzugPDFFilter:    "PDFファイル (*.pdf)",
	TischabzugPDFVorschlag: "テーブル.pdf",

	PartieSpeichern: "ゲームの書き出し先",
	PartieOeffnen:   "ゲームを読み込む",
	PartieFilter:    "MSE4Macゲーム (*.msepartie)",
	PartieVorschlag: "ゲーム.msepartie",

	TurnierSpeichern: "トーナメントの書き出し先",
	TurnierOeffnen:   "トーナメントを読み込む",
	TurnierFilter:    "MSE4Macトーナメント (*.mseturnier)",
	TurnierVorschlag: "トーナメント.mseturnier",
	BerichtSpeichern: "トーナメントレポートを別名で保存",
	BerichtFilter:    "Markdown (*.md)",
	BerichtVorschlag: "トーナメントレポート.md",

	SymbolAusgeben:      "シンボルの書き出し先",
	SymbolEinlesen:      "シンボルをライブラリに読み込む",
	SymbolFilter:        "MSEシンボル (*.mse-symbol)",
	SymbolVorschlag:     "シンボル.mse-symbol",
	SprachpaketAusgeben: "翻訳の書き出し先",
	SprachpaketEinlesen: "翻訳を読み込む",
	SprachpaketFilter:   "MSE4Mac翻訳 (*.json)",

	SetOeffnen:   "セットを開く",
	SetFilter:    "MSEセット (*.mse-set)",
	SetSpeichern: "セットを別名で保存",
	SetVorschlag: "セット.mse-set",

	UmschlagSchnueren: "セットをエンベロープとして保存",
	UmschlagOeffnen:   "エンベロープを確認",
	UmschlagFilter:    "MSE4Macエンベロープ (*.msepackage)",
	UmschlagVorschlag: "エンベロープ.msepackage",
	Uebernommen:       "読み込み済み",

	LeiheAblegen:   "貸し出しのエンベロープを保存",
	LeiheVorschlag: "貸し出し.msepackage",

	BestandSpeichern: "データベースをバックアップ",
	BestandFilter:    "MSE4Macバックアップ (*.msesicherung)",
	BestandVorschlag: "mse-バックアップ.msesicherung",
	BestandOeffnen:   "復元するバックアップを選択",
	KontakteOeffnen:  "vCardの連絡先を選択",
	KontakteFilter:   "vCard (*.vcf)",

	DiagnoseSpeichern: "診断情報を保存",
	DiagnoseFilter:    "テキストファイル (*.txt)",
	DiagnoseVorschlag: "mse-診断.txt",

	DiagnoseBestand:      "データベース",
	DiagnoseProtokollort: "ログ",
	DiagnoseKanaele:      "チャンネル",
	DiagnoseKanalDatei:   "ファイル",
	DiagnoseKanalSystem:  "Unified Log",
	DiagnoseKanalAusgabe: "標準エラー出力",
	DiagnoseKanalAn:      "オン",
	DiagnoseKanalAus:     "オフ",
	DiagnoseKanalFehlt:   "見つかりません",
	DiagnoseStufe:        "レベル",
	DiagnoseAusfuehrlich: "詳細ログがオンになっています。保持されるのは、1ファイルあたり最大1 MBと、" +
		"1つ前の世代のファイルです。問題が見つかったら、設定 › 診断で元に戻してください。",
	DiagnosepaketSpeichern: "診断パッケージを保存",
	DiagnosepaketVorschlag: "診断.msediagnose",
	DiagnosepaketFilter:    "診断パッケージ (*.msediagnose)",
	DiagnoseTeilUnbekannt: "この項目は不明です。アプリケーションは名前を把握していますが、" +
		"内容は把握していません。",
	DiagnoseProtokollVerdeckt: "ログの行はここには表示されません。ログの行は、このプロファイルではなく、" +
		"このプロセスに属しているためです。必要な場合は、ご自身のプロファイルを開くか、" +
		"上記の場所にあるファイルを読んでください。",
	DiagnoseSchema:          "スキーマ",
	DiagnoseWaisen:          "孤立した行",
	DiagnoseBau:             "ビルド",
	DiagnoseVeraendert:      "作業ツリーが変更されています",
	DiagnoseBindungsvertrag: "バインディング契約",
	DiagnoseMigrationen:     "%d件のマイグレーション、最新は%s",
	DiagnoseKeineSchema:     "マイグレーションの記録なし",
	DiagnoseNichtLesbar:     "読み取れません",
	DiagnosePakete:          "パッケージ",
	DiagnoseWeitere:         "… ほか%d件",
	DiagnoseKeinBestand:     "開いているデータベースはありません",
	DiagnoseKeinProtokoll:   "ログなし",
	DiagnoseKeineVerwaltung: "パッケージ管理は設定されていません",

	MTGJSONDatei:       "MTG JSONを開く",
	MTGJSONFilter:      "MTG JSON (*.json)",
	MTGJSONVerzeichnis: "MTG JSONのフォルダを開く",
	MTGJSONSchreiben:   "MTG JSONとして保存",
	MTGJSONVorschlag:   "set.json",

	PersonendatenSichern: "自分のデータを保存",
	JSONFilter:           "JSON (*.json)",

	EinstellungenSichern:   "設定を保存",
	EinstellungenOeffnen:   "設定を読み込む",
	EinstellungenVorschlag: "設定.json",

	VorlageSchreiben: "書き出しテンプレートで保存",
	VorlageFilter:    "テンプレートの出力",

	Bildordner:     "画像用のフォルダを選択",
	BogenSchreiben: "印刷シートを保存",
	BogenVorschlag: "印刷シート.pdf",
	PDFFilter:      "PDF (*.pdf)",

	PaketWaehlen: "テンプレートパッケージを選択",
	MSE2Waehlen:  "MSE2テンプレートのフォルダを選択",

	Planordner: "週ごとのバックアップを保存するフォルダを選択",
}

var berichttexteJa = domain.Berichttexte{
	Spiel: "ゲーム", Format: "フォーマット", Stand: "状態", Abgeschlossen: "完了日",
	Laufend: "進行中", Teilnehmende: "参加者", Ausgestiegen: "途中離脱", Runden: "ラウンド数",
	Rangliste: "順位表", Platz: "順位", Name: "名前", Punkte: "ポイント", Bilanz: "戦績",
	Gegnermatchquote: "OMW", Spielquote: "GW", Gegnerspielquote: "OGW",
	Runde: "ラウンド", Tisch: "テーブル", SeiteA: "プレイヤーA", SeiteB: "プレイヤーB", Ergebnis: "結果",
	Bye: "不戦勝", Offen: "未定",
}
