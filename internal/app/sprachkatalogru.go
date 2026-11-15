package app

// **Die Kataloge für Russisch** — maschinell übersetzt (LOC-01, #1687), Ausgangssprache Deutsch.

import "github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"

var menuetexteRu = Menuetexte{
	Ablage:         "Файл",
	Neu:            "Новый",
	Oeffnen:        "Открыть…",
	Sichern:        "Сохранить",
	SichernUnter:   "Сохранить как…",
	Schliessen:     "Закрыть",
	Drucken:        "Напечатать…",
	BestandSichern: "Создать резервную копию базы данных…",
	Einstellungen:  "Настройки…",
	Zuletzt:        "Открыть недавние",
	Hilfe:          "Справка",
	HilfeOeffnen:   "Справка MagicSetEditor4Mac",

	Set:              "Набор",
	Rueckgaengig:     "Откатить",
	AllesAuswaehlen:  "Отметить все карты",
	Wiederherstellen: "Вернуть",
	Suchen:           "Найти…",
	Ersetzen:         "Заменить…",

	Ausgeben:      "Экспорт",
	Exportvorlage: "С помощью шаблона экспорта…",
	Bilder:        "Как изображения…",
	Druckbogen:    "Печатный лист…",

	Ueber: "О программе MagicSetEditor4Mac",

	Darstellung:            "Вид",
	SeitenleisteEinblenden: "Показать боковую панель",
	SeitenleisteAusblenden: "Скрыть боковую панель",
	NavigatorEinblenden:    "Показать навигатор",
	NavigatorAusblenden:    "Скрыть навигатор",
	InspectorEinblenden:    "Показать инспектор",
	InspectorAusblenden:    "Скрыть инспектор",
	SeiteZurueck:           "Назад",
	SeiteVor:               "Вперёд",

	DarstellungZuruecksetzen: "Сбросить оформление",

	Anwendung:        "MagicSetEditor4Mac",
	Dienste:          "Службы",
	Ausblenden:       "Скрыть MagicSetEditor4Mac",
	AndereAusblenden: "Скрыть остальные",
	AlleEinblenden:   "Показать все",
	Beenden:          "Завершить MagicSetEditor4Mac",

	Bearbeiten:      "Правка",
	Widerrufen:      "Отменить",
	Wiederholen:     "Повторить",
	Ausschneiden:    "Вырезать",
	Kopieren:        "Скопировать",
	Einsetzen:       "Вставить",
	Loeschen:        "Удалить",
	AllesMarkieren:  "Выбрать все",
	Sprachausgabe:   "Речь",
	SprechenStarten: "Начать озвучивание",
	SprechenStoppen: "Остановить озвучивание",

	Fenster:       "Окно",
	Minimieren:    "Свернуть",
	Zoomen:        "Масштаб",
	AlleNachVorne: "Все окна — на передний план",
}

var dialogtexteRu = Dialogtexte{
	BildOeffnen: "Открыть изображение",
	BildFilter:  "Изображения (*.png, *.jpg, *.gif)",

	PaketdateiEinfuegen: "Добавить файл в пакет",
	PaketdateiFilter:    "Изображения и шрифты (*.png, *.jpg, *.gif, *.ttf, *.otf)",

	PaketWeitergeben:  "Передать пакет как",
	InstallerEinlesen: "Прочитать установщик",
	InstallerFilter:   "Установщики MSE (*.mse-installer)",

	TischabzugSpeichern: "Сохранить изображение стола как",
	TischabzugFilter:    "Изображения PNG (*.png)",
	TischabzugVorschlag: "стол.png",

	TischabzugPDFSpeichern: "Сохранить изображение стола в PDF",
	TischabzugPDFFilter:    "Файлы PDF (*.pdf)",
	TischabzugPDFVorschlag: "стол.pdf",

	PartieSpeichern: "Экспортировать партию в",
	PartieOeffnen:   "Импортировать партию",
	PartieFilter:    "Партии MSE4Mac (*.msepartie)",
	PartieVorschlag: "партия.msepartie",

	TurnierSpeichern: "Экспортировать турнир в",
	TurnierOeffnen:   "Импортировать турнир",
	TurnierFilter:    "Турниры MSE4Mac (*.mseturnier)",
	TurnierVorschlag: "турнир.mseturnier",
	BerichtSpeichern: "Сохранить отчёт о турнире как",
	BerichtFilter:    "Markdown (*.md)",
	BerichtVorschlag: "отчёт-о-турнире.md",

	SymbolAusgeben:      "Экспортировать символ в",
	SymbolEinlesen:      "Импортировать символ в библиотеку",
	SymbolFilter:        "Символы MSE (*.mse-symbol)",
	SymbolVorschlag:     "символ.mse-symbol",
	SprachpaketAusgeben: "Экспортировать перевод в",
	SprachpaketEinlesen: "Импортировать перевод",
	SprachpaketFilter:   "Переводы MSE4Mac (*.json)",

	SetOeffnen:   "Открыть набор",
	SetFilter:    "Наборы MSE (*.mse-set)",
	SetSpeichern: "Сохранить набор как",
	SetVorschlag: "набор.mse-set",

	UmschlagSchnueren: "Сохранить набор как конверт",
	UmschlagOeffnen:   "Просмотреть конверт",
	UmschlagFilter:    "Конверты MSE4Mac (*.msepackage)",
	UmschlagVorschlag: "конверт.msepackage",
	Uebernommen:       "импортировано",

	LeiheAblegen:   "Сохранить конверт займа",
	LeiheVorschlag: "заём.msepackage",

	BestandSpeichern: "Создать резервную копию базы данных",
	BestandFilter:    "Резервные копии MSE4Mac (*.msesicherung)",
	BestandVorschlag: "mse-копия.msesicherung",
	BestandOeffnen:   "Выбери резервную копию для восстановления",
	KontakteOeffnen:  "Выбери контакты в формате vCard",
	KontakteFilter:   "vCard (*.vcf)",

	DiagnoseSpeichern: "Сохранить диагностику",
	DiagnoseFilter:    "Текстовый файл (*.txt)",
	DiagnoseVorschlag: "mse-диагностика.txt",

	DiagnoseBestand:      "База данных",
	DiagnoseProtokollort: "Журнал",
	DiagnoseKanaele:      "Каналы",
	DiagnoseKanalDatei:   "файл",
	DiagnoseKanalSystem:  "Unified Log",
	DiagnoseKanalAusgabe: "стандартный вывод ошибок",
	DiagnoseKanalAn:      "вкл.",
	DiagnoseKanalAus:     "выкл.",
	DiagnoseKanalFehlt:   "отсутствует",
	DiagnoseStufe:        "уровень",
	DiagnoseAusfuehrlich: "Подробный журнал включён — он хранит не более 1 МБ " +
		"на файл и одну предыдущую версию. Верни прежнее значение в разделе " +
		"«Настройки › Диагностика», когда ошибка будет найдена.",
	DiagnosepaketSpeichern: "Сохранить диагностический пакет",
	DiagnosepaketVorschlag: "диагностика.msediagnose",
	DiagnosepaketFilter:    "Диагностический пакет (*.msediagnose)",
	DiagnoseTeilUnbekannt: "Эта часть неизвестна — приложению известно " +
		"её название, но не содержимое.",
	DiagnoseProtokollVerdeckt: "Строки журнала здесь не показаны: " +
		"они принадлежат этому процессу, а не этому профилю. Чтобы " +
		"их увидеть, открой свой профиль или прочитай файл " +
		"по указанному выше адресу.",
	DiagnoseSchema:          "Схема",
	DiagnoseWaisen:          "Строки-сироты",
	DiagnoseBau:             "Сборка",
	DiagnoseVeraendert:      "рабочее дерево изменено",
	DiagnoseBindungsvertrag: "Контракт привязок",
	DiagnoseMigrationen:     "Миграций: %d, последняя: %s",
	DiagnoseKeineSchema:     "ни одной миграции не записано",
	DiagnoseNichtLesbar:     "не удаётся прочитать",
	DiagnosePakete:          "Пакеты",
	DiagnoseWeitere:         "… и ещё %d",
	DiagnoseKeinBestand:     "нет открытой базы данных",
	DiagnoseKeinProtokoll:   "Нет журнала",
	DiagnoseKeineVerwaltung: "управление пакетами не настроено",

	MTGJSONDatei:       "Открыть MTG JSON",
	MTGJSONFilter:      "MTG JSON (*.json)",
	MTGJSONVerzeichnis: "Открыть папку с MTG JSON",
	MTGJSONSchreiben:   "Сохранить как MTG JSON",
	MTGJSONVorschlag:   "set.json",

	PersonendatenSichern: "Сохранить мои данные",
	JSONFilter:           "JSON (*.json)",

	EinstellungenSichern:   "Сохранить настройки",
	EinstellungenOeffnen:   "Импортировать настройки",
	EinstellungenVorschlag: "настройки.json",

	VorlageSchreiben: "Сохранить с помощью шаблона экспорта",
	VorlageFilter:    "Результат шаблона",

	Bildordner:     "Выбери папку для изображений",
	BogenSchreiben: "Сохранить печатный лист",
	BogenVorschlag: "печатный-лист.pdf",
	PDFFilter:      "PDF (*.pdf)",

	PaketWaehlen: "Выбери пакет шаблона",
	MSE2Waehlen:  "Выбери папку с шаблонами MSE2",

	Planordner: "Выбери папку для еженедельной резервной копии",
}

var berichttexteRu = domain.Berichttexte{
	Spiel: "Игра", Format: "Формат", Stand: "Статус", Abgeschlossen: "завершён",
	Laufend: "идёт", Teilnehmende: "Участники", Ausgestiegen: "выбыл", Runden: "Туры",
	Rangliste: "Турнирная таблица", Platz: "Место", Name: "Имя", Punkte: "Очки", Bilanz: "Итог",
	Gegnermatchquote: "OMW", Spielquote: "GW", Gegnerspielquote: "OGW",
	Runde: "Тур", Tisch: "Стол", SeiteA: "Игрок A", SeiteB: "Игрок B", Ergebnis: "Результат",
	Bye: "Свободный тур", Offen: "открыт",
}
