// The texts of the **set list** — open, create, import.
//
// Split out of `seiten.ts` when the file grew past 400 lines (§2).

export const sets = {
  auswahlLoesen: 'Показать все наборы',
  erklaerung:
    'Каждый открытый набор хранится в хранилище. Отсюда ты попадаешь к его картам — или записываешь его обратно в файл.',
  titel: 'Наборы',
  leer: 'Ни один набор ещё не открыт. Открой файл .mse-set, чтобы начать.',
  setOeffnen: 'Открыть набор …',
  neuTitel: 'Новый набор',
  neuKopf: 'Новый набор',
  einlesenTitel: 'Прочитать набор из другого места',
  neuHinweis:
    'Пустой набор — игра и стиль определяют, как выглядят его карты. Стиль можно сменить позже.',
  neuName: 'Название набора',
  neuStil: 'Стиль нового набора',
  neuAnlegen: 'Создать набор',
  neuSchritte: 'Шаги к новому набору',
  neuSchrittArt: 'Вид',
  neuSchrittAussehen: 'Внешний вид',
  neuSchrittName: 'Название',
  neuArtFrage: 'Для чего эти карты? Название пакета мы подберём сами.',
  neuArtZusatz: '· доступно стилей: {stile}',
  neuEigenes: 'Что-то совершенно своё — создать игровой пакет',
  neuWeiter: 'Далее',
  neuSchrittZurueck: 'Назад',
  neuUeberspringen: 'Пропустить',
  neuWahl: '{spiel} · {stil}',
  neuOhneSpiele:
    'Ни один игровой пакет не установлен. Без него нет полей карт — пакеты берутся на странице пакетов.',
  neuOhneStile:
    'Для {spiel} не установлен ни один пакет стиля. Без стиля набор создать нельзя — пакеты берутся на странице пакетов.',
  neuZuDenPaketen: 'К пакетам',
  neuZurueck: 'Нажми ⌘[, чтобы вернуться сюда; введённое сохранится.',
  beispielsets: 'Поставляются для пробы',
  geoeffnet: 'открыт',
  schonOffen: 'Этот файл уже открыт — вот он.',
  teilerfolg:
    'Файл прочитан не до конца. Тем не менее в базе данных оказалось наборов: {anzahl} — они в списке ниже.',
  beimOeffnenBefunde:
    'При открытии не всё удалось прочитать. Набор тем не менее создан; в нём не хватает следующего:',
  // **The preview before reading in** — FMT-03.AK-01 (#1692). It appears
  // only when there is something to say.
  vorschau: {
    ueberschrift: 'Перед чтением: {name}',
    umfang: '{karten} · игра {spiel} · стиль {stil}',
    umfangOhneStil: '{karten} · игра {spiel}',
    schonOffen:
      'Этот файл уже есть в твоей библиотеке — «Открыть» покажет существующий набор вместо добавления второго.',
    befunde: 'Эти места не читаются; набор был бы создан без них:',
    bedarf:
      'Не хватает пакетов. Прочитать файл всё равно можно — карты рисуются с заглушкой, пока пакетов нет:',
    einlesen: 'Прочитать',
    aufschlagen: 'Открыть',
    verwerfen: 'Не читать',
  },
  // **Deletion** (#1218) — final, confirmed by name.
  loeschen: 'Удалить …',
  loeschenTitel: 'Удалить этот набор',
  loeschenWarnung:
    '{name} исчезает из хранилища; вместе с ним удаляется: {karten}. Корзины и отмены нет. То, что лежит в файле, остаётся там — удаляется копия в хранилище.',
  loeschenSammlung:
    'Твоя коллекция остаётся: экземпляр относится к изданию, а не к набору.',
  loeschenFeld: 'Введи название набора для подтверждения',
  loeschenPasstNicht: 'Имя пока не совпадает с «{name}».',
  loeschenJa: 'Удалить безвозвратно',
  loeschenNein: 'Отмена',
  loeschenBefunde: 'Набор удалён. Не удалось убрать следующие остатки:',

  spalten: {
    name: 'Название',
    spiel: 'Игра',
    karten: 'Карты',
    herkunft: 'Источник',
  },
  mtgjsonUeberschrift: 'MTG JSON',
  mtgjsonErklaerung:
    'Описания карт с mtgjson.com — один файл или папка. С какими полями карты они сопоставляются, зависит от игры.',
  mtgjsonSpiel: 'Игра',
  mtgjsonDatei: 'Открыть файл …',
  mtgjsonVerzeichnis: 'Открыть папку …',
  mtgjsonKeinSpiel:
    'Ни у одной игры нет сопоставления с полями MTG JSON. Без него ничего прочитать нельзя.',
  scryfallsatzUeberschrift: 'Набор из Scryfall',
  scryfallsatzErklaerung:
    'Код набора определяет, какие карты придут — с текстами, без изображений. Приходят название, мановая стоимость, типы, текст правил, художественный текст, сила, выносливость, редкость и иллюстратор; изображения загружаются для каждой карты отдельно. Стиль выбери потом.',
  scryfallsatzKeinSpiel:
    'Без игры, у которой есть сопоставление, набор импортировать нельзя.',
  scryfallsatzSpiel: 'Игра для сопоставления',
  scryfallsatzCode: 'Код набора на Scryfall',
  scryfallsatzPlatzhalter: 'например, lea',
  scryfallsatzEinlesen: 'Импортировать набор',
  vorlageAusfuhr: 'Через шаблон экспорта …',
  vorlageWaehlen: 'Шаблон экспорта',
  vorlageAusfuehren: 'Выполнить',
  vorlageSchliessen: 'Закрыть',
  vorlageKeine:
    'Для этой игры не установлен ни один шаблон экспорта. Шаблоны приходят из загруженных пакетов.',
  vorlageGeschrieben: 'Записано в',
  vorlageBefunde: 'При выполнении что-то произошло — файл всё равно записан:',
  vorlageGescheitert:
    'Этот шаблон не выполняется до конца. Чего ему не хватает, указано ниже:',
  mtgjsonSichern: 'Сохранить как MTG JSON …',
  mtgjsonGesichert: 'шт. сохранено в',
  mtgjsonVerluste:
    'Не всё можно выразить при экспорте. Файл создан; в нём не хватает следующих деталей:',
  mtgjsonSpalte: 'Экспорт',
  speichern: 'Сохранить',
  speichernUnter: 'Сохранить как …',
  speicherSpalte: 'Сохранение',
  gespeichert: 'Сохранено в',
  beimSpeichernBefunde:
    'При сохранении не всё перенесено. Файл набора записан; рядом с ним не хватает следующих файлов:',
  oeffnen: 'Показать карты этого набора',
  handlungenAn: 'Действия — {name}',
  handlungenKurz: '…',
  vorbehalte: {
    mitZahl: 'Не понято ({anzahl})',
    ueberschrift: 'Не истолковано',
    erklaerung:
      'Импорт не понял эти строки. Они переносятся дальше без изменений и при сохранении не теряются.',
    keine: 'Ничего — этот набор прочитан полностью.',
  },

  // **The portable envelope** — work item N4 of phase 5.4
  // (#1457).
  umschlagAblegen: 'Сохранить как конверт',
  umschlagGeschnuert:
    'Сохранено. Конверт подписан — получатель может проверить, что он отсюда.',
  umschlagTitel: 'Принять конверт',
  umschlagErklaerung:
    'Конверт — это набор, сохранённый кем-то другим. Сначала его проверяют, потом принимают — до этого в базе данных ничего не меняется.',
  umschlagWaehlen: 'Проверить конверт …',
  umschlagUrheber: 'Подписал',
  umschlagAbdruckHinweis:
    'Отпечаток говорит, КТО подписал, — но не о том, стоит ли ему доверять. Если ты знаешь этого человека, сверь отпечаток с ним.',
  umschlagInhalt: 'Содержимое',
  umschlagUebernehmen: 'Принять',
  umschlagVerwerfen: 'Не надо',
  umschlagUebernommen:
    'Принято. Набор создан заново — у него собственные идентификаторы, и ничего не перезаписывается.',
}
