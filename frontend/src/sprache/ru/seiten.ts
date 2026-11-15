// The texts of the individual pages.
import { bestandspflege } from './seitenbestand'
import { datenpflege } from './seitenpflege'
import { kosten } from './seitenkosten'
import { ruecksetzung } from './seitenruecksetzung'
import { ruecksprung } from './seitenruecksprung'
import { einstellungen } from './seiteneinstellungen'
import { nutzer } from './seitennutzer'
import { karten } from './seitenkarten'
import { sprachstaende } from './seitensprachstaende'
import { handeln } from './seitenhandeln'
import { technik } from './seitentechnik'
import { recherche } from './seitenrecherche'
import { decks } from './seitendecks'
import { turniere } from './seitenturniere'
import { handlungen } from './seitenhandlungen'
import { spielen } from './seitenspielen'
import { sammlung } from './seitensammlung'
import { adressbuch } from './seitenadressbuch'
import { vorlagen } from './seitenvorlagen'
import { sets } from './seitensets'
import { ausgeben } from './seitenausgeben'
import { werkzeugseiten } from './seitenwerkzeuge'
import { start } from './seitenstart'
import { paketplan } from './seitenpaketplan'
import { bezuege } from './seitenbezuege'

export const seiten = {
  start,
  sets,
  setangaben: {
    kopf: 'Сведения о файле',
    kopfHinweis:
      'Что Magic Set Editor записал об этом наборе при его сохранении. Эти сведения переносятся дальше и не изменяются.',
    mseFassung: 'Записано в MSE',
    spielfassung: 'Версия игрового пакета',
    stilfassung: 'Версия стиля',
    versionskontrolle: 'Контроль версий',
    apprenticecode: 'Код Apprentice',
    titel: 'Сведения о наборе',
    leerTitel: 'Нет полей набора',
    leer: 'Игровой пакет этого набора не объявляет блоков `set field:`. Выбери игру, в которой они есть.',
  },
  karten,
  sprachstaende,
  decks,
  turniere,
  spielen,
  handlungen,
  sammlung,
  adressbuch,
  vorlagen,
  ...werkzeugseiten,
  bilder: {
    titel: 'Изображения',
    erklaerung:
      'Каждая карта отрисовывается и записывается отдельно. Процесс идёт здесь, в окне: он показывает, на каком этапе находится, и его можно остановить в любой момент.',
    ordnerWaehlen: 'Выбрать папку …',
    keinOrdner: 'Папка ещё не выбрана.',
    gemerkt: 'Выбрана в прошлый раз — одна папка для всех сетов.',
    nichtGemerkt: 'Папка действует для этого вывода, но не запомнена:',
    vorlage: 'Шаблон имени',
    konflikt: 'Если файл уже существует',
    konfliktHilfe:
      'Две карты с одинаковым названием дают одинаковое имя файла. «Нумеровать заново» — вариант Magic Set Editor 2 по умолчанию: счёт идёт в пределах одного запуска, и всё, что оставил предыдущий запуск, перезаписывается.',

    // Die Schlüssel kommen aus `internal/domain/einstellung.go`.
    konflikte: {
      altbehalten: 'Оставить существующий',
      ueberschreiben: 'Перезаписать',
      nummerieren: 'Нумеровать',
      nummerierenneu: 'Нумеровать заново',
    },
    vorlageHilfe:
      '{nummer} превращается в трёхзначный номер карты, {name} — в название карты. К одинаковым названиям добавляется номер.',
    format: 'Формат',
    formatPNG: 'PNG',
    formatJPG: 'JPG',
    formatBMP: 'BMP',
    formatTIFF: 'TIFF',
    formatSVG: 'SVG',
    guete: 'Качество',
    gueteFast: 'Быстро — для беглого просмотра',
    gueteMedium: 'По правилу Magic Set Editor 2',
    gueteHigh: 'Высокое — без мерцания на тонких линиях',
    gueteUltra: 'Максимальное — для печати в полном разрешении',
    breite: 'Ширина в пикселях',
    breiteHilfe:
      'Без dpi: ни один шаблон не задаёт ширину карты в миллиметрах. 750 пикселей на карте шириной 63 мм — это 300 dpi.',
    ausgeben: 'Записать изображения',
    abbrechen: 'Остановить',
    laeuft: 'записано из',
    fertig: 'изобр. записано в',
    abgebrochen: 'Остановлено.',
    uebersprungen:
      'Пропущено карт: {anzahl} — файл уже существовал, а выбран вариант «Оставить существующий»:',
    befunde:
      'При отрисовке получилось не всё. Изображения созданы; в них не хватает следующего:',
  },
  druckbogen: {
    titel: 'Лист печати',
    erklaerung:
      'Карты сохраняют свой размер: 63 × 88 мм — ровно так, как их режут. Метки реза находятся на краю бумаги, а не на карте.',
    papier: 'Бумага',
    papierA4: 'A4',
    papierLetter: 'Letter',
    anschnitt: 'Припуск под обрез в миллиметрах',
    anschnittHilfe:
      'Припуск под обрез — это то, насколько изображение выходит за линию реза; благодаря ему на разрезанной карте не видна белая кромка. Между картами оставляется двойной припуск.',
    aufloesung: 'Разрешение',
    dpi150: '150 dpi',
    dpi300: '300 dpi',
    dpi600: '600 dpi',
    schnittmarken: 'Метки реза',
    legen: 'Разместить лист печати …',
    fertig: '→',
    seiten: 'страницы:',
    befunde:
      'Поместилось не всё. Лист создан; следующие места нужно проверить:',
  },
  // The **technical area** (#1164) — commands to copy.
  handeln,
  technik,

  // **Die Recherche** (A16, #1515) — sie steht im Sucharbeitsplatz und
  // im Karteneditor, nicht auf einer eigenen Seite.
  recherche,
  pakete: {
    plan: paketplan,
    bezuege,
    sammeldaten: {
      ueberschrift: 'Массив данных Scryfall',
      erklaerung:
        'Файл со всеми данными карт. Когда он загружен, функции Scryfall работают без сети — с данными возрастом до суток.',
      fehlt: 'Ещё не загружено.',
      vorhanden: 'Загружено, {mb} МБ.',
      angeboten: 'Scryfall предлагает {mb} МБ по состоянию на {stand}.',
      geholt: 'Загружено {zeit}.',
      holen: 'Загрузить массив данных',
      laedt: 'Загрузка …',
    },

    titel: 'Пакеты',
    bezugUeberschrift: 'Загрузить шаблон',
    bezugErklaerung:
      'Шаблоны MSE хранятся в репозиториях. Загружается состояние — релиз или коммит, — но не ветка: ветка меняется, и та же запись завтра дала бы другой пакет.',
    bezugAdresse: 'Адрес репозитория',
    bezugStand: 'Состояние (релиз или коммит)',
    bezugUnterverzeichnis: 'Подпапка (корень пакета)',
    bezugName: 'Имя в папке пакетов',
    bezugPruefen: 'Показать источник …',
    bezugStarten: 'Загрузить',
    bezugAbbrechen: 'Не загружать',
    bezugGeholt: 'Загружено в',
    bezugPruefsumme: 'Контрольная сумма',
    bezugUebergangen: 'Не распаковано — исполняемые файлы не переносятся:',
    bezugUnvollstaendig:
      'Не указаны адрес и состояние. Без обоих загружать нечего.',
    bezugHerkunft: 'Будет загружено: {quelle}, состояние {stand}.',
    bezugLizenz:
      'Это содержимое не принадлежит MagicSetEditor4Mac и не подпадает под нашу лицензию. Рамки карт, символы и шрифты могут быть защищены; их использование — на твоей ответственности. Исполняемые файлы не распаковываются.',
    mse2Ueberschrift: 'Перенести из установки MSE2',
    mse2Erklaerung:
      'Тот, кто пользуется MSE2, уже имеет шаблоны на своём компьютере. Ниже перечислены обычные места — ни одно из них не читается, пока ты его не выберешь.',
    mse2Uebernehmen: 'Взять отсюда',
    mse2VonHier: 'Взять отсюда — {ort}',
    mse2Waehlen: 'Выбрать папку …',
    mse2KeinVorschlag:
      'В обычных местах установка MSE2 не найдена. Выбери папку с шаблонами самостоятельно — в MSE2 она называется «data».',
    installerErklaerung:
      'Или файл .mse-installer — так MSE2 делится шаблонами. Всё, что уже установлено, остаётся как есть.',
    installerWaehlen: 'Прочитать установщик …',
    mse2Ergebnis: 'Перенесено шаблонов: {anzahl}.',
    mse2Uebergangen:
      'Уже существуют и поэтому пропущены — ничего не перезаписано:',
    abhaengigUeberschrift: 'Зависимости',
    abhaengigAllesGut: 'Все зависимости выполнены.',
    abhaengigFehlt:
      '{paket} требует {braucht} — не установлено. Загрузи его или перенеси из установки MSE2.',
    abhaengigZuAlt:
      '{paket} требует {braucht} версии {verlangt}; установлена версия {da}. Без более новой версии шаблон молча нарисует что-то другое.',
    abhaengigZyklus: '{paket} и {braucht} по кругу требуют друг друга.',
    abhaengigUnlesbar:
      '{paket} требует {braucht}: версии нельзя сравнить — требуется {verlangt}, установлена {da}.',
    quellenErklaerung:
      'Четыре коллекции, по одному щелчку на каждую — адрес и ревизия уже подставлены; пока ты не согласишься, ничего не загружается. Шрифты шаблонов загружаются вместе с ними; без них ни одна карта не отрисовывается правильно. Ни у одной из четырёх нет лицензии: содержимое принадлежит правообладателям, а не нам.',
    quellenGroesse: 'около {megabyte} МБ',

    // Die Schlüssel kommen aus `internal/bezug/quellen.go`. Ein Wächter
    // in `internal/werkzeuge` hält beide Seiten zusammen: eine Quelle
    // ohne Satz hier macht die Pipeline rot.
    quellen: {
      basicM15: 'Обычные рамки M15. Кратчайший путь к отрисованной карте.',
      fullNonMagic: 'Шаблоны для игр, отличных от Magic.',
      m15: 'Публичные шаблоны M15 — 324 пакета. Правильное начало для большинства.',
      fullMagic:
        'Почти все публичные шаблоны Magic — 526 пакетов. Эта коллекция служит эталоном для проекта.',
    },
    katalogUeberschrift: 'Установленные пакеты',
    katalogLeer:
      'Ни один пакет не установлен. Без игры и стиля предпросмотр использует заглушку.',
    katalogName: 'Название',
    katalogArt: 'Вид',
    katalogVersion: 'Версия',
    katalogHandlung: 'Действие',
    katalogBefunde: 'При чтении папок пакетов:',
    anlegen: 'Создать пакет',
    anlegenArt: 'Вид нового пакета',
    anlegenArtSpiel: 'Игра (.mse-game)',
    anlegenArtStil: 'Стиль (.mse-style)',
    anlegenArtInclude: 'Include (.mse-include)',
    anlegenArtExport: 'Шаблон экспорта (.mse-export-template)',
    anlegenName: 'Краткое имя',
    anlegenNameHinweis:
      'Буквы, цифры, дефисы и подчёркивания — это станет именем папки',
    anlegenVollname: 'Полное имя',
    anlegenVollnameHinweis: 'Пусто означает: то же, что краткое имя',
    anlegenSpiel: 'Игра нового пакета',
    anlegenSpielHinweis:
      'Стиль размечает поля своей игры, шаблон экспорта их выводит',
    anlegenSpielLeer: 'Выбери',
    anlegenKnopf: 'Создать',
    anlegenGelungen: 'Создано: {name}',
    loeschen: 'Удалить',
    loeschenAn: 'Удалить — {name}',
    handlungenAn: 'Действия — {name}',
    handlungenKurz: '…',
    katalogSuche: 'Поиск',
    auswahlLoesen: 'Показать все пакеты',
    leiste: {
      bezeichnung: 'Показатели пакетов',
      installiert: 'Установлено',
      befund: 'С замечаниями',
      nichtGelesen: 'не прочитано',
      auswahl: '{pakete} с замечаниями',
    },
    katalogSucheHinweis: 'Название, вид, источник или группа',
    katalogArtAlle: 'Все виды',
    katalogOrdnung: 'Сортировать по',
    katalogOrdnungGruppen: 'Группам',
    katalogOrdnungName: 'Названию',
    katalogOrdnungVersion: 'Версии',
    katalogOrdnungHerkunft: 'Источнику',
    katalogGefunden: 'Пакеты: {gefunden} из {gesamt}',
    katalogNichtsGefunden:
      'Ни один пакет не подходит под этот поиск. Попробуй другое написание или сбрось фильтр по виду.',
    katalogLeerTitel: 'Пакетов пока нет',
    katalogNichtsTitel: 'Ничего не найдено',
    katalogOhneGruppe: 'Без группы',
    katalogGruppeAn: 'Группа {name}',
    katalogGruppe: 'Группа',
    katalogOhneVersion: 'без версии',
    katalogZustandFehlt: 'Нет зависимости',
    katalogZustandZuAlt: 'Зависимость устарела',
    katalogZustandUnlesbar: 'Версии несравнимы',
    katalogZustandZyklus: 'Циклическая зависимость',
    katalogEinVorbehalt: 'одно замечание',
    katalogVorbehalte: 'Замечаний: {anzahl}',
    umschlagTitel: 'Шаблон из конверта',
    umschlagErklaerung:
      'Конверт несёт **файлы** шаблона; имя ты выбираешь вместе с расширением, например `mygame.mse-game` или `mystyle.mse-style`. После этого шаблон хранится здесь под этим именем. Если такой пакет уже существует, ничего не перезаписывается: ты получаешь отказ и выбираешь другое имя.',
    umschlagName: 'Как назвать пакет',
    umschlagOhneNamen: 'Сначала укажи, как назвать пакет.',
    umschlagAnnehmen: 'Принять шаблон',
    umschlagAngenommen: 'Шаблон теперь находится в твоей папке пакетов.',
  },
  // **Suchen und Ersetzen** (#1168).
  ersetzen: {
    titel: 'Замена',
    erklaerung:
      'Заменяет текст в текстовых полях этого набора. Сначала предпросмотр, затем применение — и одна отмена возвращает всё действие целиком.',
    suchtext: 'Найти',
    ersatz: 'Заменить на',
    grossKlein: 'Учитывать регистр',
    ganzesWort: 'Только слова целиком',
    vorschauen: 'Показать, что изменится',
    ersetzen: 'Заменить сейчас ({treffer})',
    nichtsGefunden: 'Ничего не найдено. Попробуй другое слово?',
    zusammenfassung: 'Совпадений: {treffer} · затронуто карт: {karten}',
    getan: 'Заменено — совпадений: {treffer} · карт: {karten}',
    ruecknahme: 'Нажми ⌘Z, чтобы отменить всё действие.',
    stellen: 'Затронутые места',
    kartenfeld: 'Карта {nummer} · {feld}',
    weitere: 'Остальные места не перечислены — они всё равно заменяются.',
  },

  statistik: {
    worueber: 'Считать по',
    belegungWahl: 'Заполненность по полям',
    gruppeSpielpaket: 'По игровому пакету',
    gruppeFelder: 'По значению поля',
    verteilungVon: 'Распределение «{feld}» — доля карт, у которых оно есть',
    nachDimension: 'Статистика по «{dimension}»',
    titel: 'Статистика',
    karten: 'карт',
    felder: 'разных полей',
    leere: 'карт без единого значения',
    belegung: 'Заполненность по полям — доля от всех карт набора',
    belegungLeer: 'В этом наборе пока нет значений полей.',
    verteilungLeer: 'В этом наборе у этого поля нет значений.',
    anteil: 'Доля',
    dimensionenErklaerung:
      'Игровой пакет сам сообщает, что можно подсчитать — цветовую категорию, ' +
      'значение маны, редкость. Каждое измерение вычисляет своё значение для ' +
      'каждой карты, поэтому это занимает некоторое время.',
    dimensionLeer: 'Ни у одной карты нет значения по этому измерению.',
    gruppeHinweis: 'Выбери столбец, чтобы увидеть его карты в списке карт.',
    gruppeGewaehlt: 'Выбрано: «{gruppe}» · карт: {karten}',
    gruppeZeigen: 'Показать карты',
    gruppeAufheben: 'Снять выбор',
    gruppeLeer: 'Ни у одной карты нет этого значения.',
    filterBezeichnung: 'Фильтр из статистики',
    filterKopf: 'Из статистики: «{quelle}» = «{gruppe}»',
    filterZurStatistik: 'Назад к статистике',
    filterAufheben: 'Снять фильтр',
  },
  // Die Lasche „Style“: Stilpaket und Stiloptionen des Sets (#288).
  stil: {
    titel: 'Стиль',
    vorschauTitel: 'Предпросмотр',
    vorschauLeer:
      'В этом наборе ещё нет карт — создай одну, чтобы увидеть стиль.',
    eigeneFarbe: 'свой цвет',
    vorgabe: 'По умолчанию',
  },
  suche: {
    titel: 'Поиск',
    eingabe: 'Что ты ищешь?',
    recherche: 'Исследование …',
    feld: 'Только в этом поле',
    alleFelder: 'Все поля',
    set: 'Только в этом наборе',
    alleSets: 'Все наборы',
    leer: 'Пока ничего не искали. Введи что-нибудь, чтобы найти карты.',
    nichts: 'Ничего не найдено. Попробуй другое слово?',
    gekuerzt:
      'Результатов больше, чем показано здесь. Сузь поиск, чтобы увидеть их.',
    treffer: 'Результаты',
  },
  ausgeben,
  hilfe: {
    titel: 'Справка',
    erklaerung:
      'Что означают понятия, с которыми работает это приложение, и что делать при первом запуске.',
    glossar: 'Понятия',

    // **Die Beschriftung der Laschenleiste** — seit U11 sind es Laschen
    // und keine Sprungmarken mehr; der Name ist geblieben, weil er sagt,
    // was die Leiste tut.
    aufDieserSeite: 'На этой странице',
    seitenuebersicht: 'Что делает каждая страница',
  },

  bestandspflege,
  datenpflege,
  kosten,
  ruecksetzung,
  ruecksprung,

  einstellungen,

  nutzer,
} as const
