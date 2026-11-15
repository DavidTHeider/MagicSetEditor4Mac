// **Run times, consumption and the performance profile** — OPS-05
// (#1722).

export const betrieb = {
  ueberschrift: 'Время выполнения и потребление ресурсов',
  zeitraum: 'Собрано с {seit} по {bis}.',
  aufwand: 'Затраты на сбор: {mikros} мкс',
  speicher: 'Память',
  speicherWert:
    'занимает {fussabdruck} — из них куча Go {heap}, выделено системой {system}',
  speicherOhneFussabdruck:
    'куча Go {heap}, выделено системой {system} — занимаемая память не измерена',
  web: 'Веб-процессы',
  webWert: 'всего {summe} — {teile}',
  webTeil: '{art} {groesse}',
  webArten: { webinhalt: 'Веб-содержимое', gpu: 'Графика', netz: 'Сеть' },
  ohneWeb: 'У этого процесса нет веб-представления.',
  webUngemessen: 'Не измерено: {grund}',
  goroutinen: 'Параллельные процедуры',
  rechenzeit: 'Процессорное время с запуска',
  datenbank: 'Подключения к библиотеке',
  datenbankWert:
    'открыто: {offen}, используется: {inBenutzung}, ожиданий: {gewartet}, общая длительность: {dauer}',
  ohneBestand: 'Этот процесс не держит библиотеку.',
  zwischenspeicher: 'Кеши сторонних сервисов',
  zwischenspeicherWert:
    'попаданий: {treffer}, загружено: {geholt}, выдано устаревших: {veraltet}, сбоев: {fehlgeschlagen}',
  laufzeiten: 'Время выполнения по интерфейсам',
  laufzeitenLeer:
    'Здесь вызовы не измеряются — измеряются только сервис и сервер инструментов.',
  ungemessen: 'Не измеряется: {liste}.',
  ungemessenNamen: {
    fenster: 'привязки окна (у них нет общего места, где вызов завершается)',
    kommandozeile: 'командная строка (она живёт одну команду)',
    speicher: 'занимаемая этим процессом память (система её не сообщила)',
  },
  spalteArt: 'Тип',
  spalteAnzahl: 'Вызовы',
  spalteFehler: 'Доля ошибок',
  spalteMittel: 'Среднее',
  spalteVerteilung: 'Распределение',
  arten: { lesend: 'чтение', schreibend: 'запись' },
  eimer: 'до {grenze} мс',
  verteilung: '{eimer}: вызовов — {anzahl}',
  eimerOffen: 'выше',
  spurAus:
    'Трассировка выключена. Включи её в разделе Настройки › Диагностика — она вступит в силу со следующего запуска.',
  spurAn:
    'Трассировка включена: {aufgenommen} из {gesehen} вызовов, выборка {abtastung} %. Не более {grenze} записей на {minuten} мин — вытеснено: {verdraengt}, истекло: {abgelaufen}.',
}

export const leistungsprofil = {
  ueberschrift: 'Профиль производительности',
  erklaerung:
    'Профиль ЦП этого окна за выбранное время в формате «go tool pprof». Он хранится рядом с библиотекой; сколько их остаётся, задаётся в разделе Настройки › Диагностика.',
  sekunden: 'Длительность в секундах',
  grenze: 'Целое число от 1 до 60.',
  aufzeichnen: 'Записать профиль',
  laeuft: 'Идёт запись — окно остаётся доступным.',
  fertig: 'Записано: {datei} ({groesse})',
}

export const zwischenspeicher = {
  ueberschrift: 'Кеши',
  erklaerung:
    'Что приложение сохраняет, чтобы не загружать и не вычислять это дважды — на диске и в памяти. Это не твои данные: после очистки каждый кеш снова заполняется при следующем использовании.',
  spalteSpeicher: 'Кеш',
  spalteOrt: 'Место',
  spalteQuote: 'Доля попаданий',
  spalteEintraege: 'Записи',
  spalteGroesse: 'Размер',
  spalteGrenze: 'Ограничение',
  keineQuote: 'обращений пока не было',
  keineGrenze: 'один файл, который заменяется',
  namen: {
    'scryfall.antworten': 'Ответы Scryfall',
    'scryfall.bilder': 'Изображения карт из Scryfall',
    'scryfall.sammeldaten': 'Сводный файл Scryfall',
    'render.ebenen': 'Слои после выборки',
    'pakete.bilder': 'Изображения пакетов шаблонов',
    'skript.baeume': 'Разобранные скрипты',
    'vorschau.bilder': 'Предпросмотры карт',
    'tisch.texturen': 'Изображения карт на столе',
  },
  orte: { platte: 'Диск', arbeitsspeicher: 'Память' },
  leerenHinweis:
    'Очистка удаляет всё. Без сети ответы и изображения Scryfall вернутся только после её появления.',
  leeren: 'Очистить всё',
  geleert: 'Очищено — кеши снова заполнятся при следующем использовании.',
}
