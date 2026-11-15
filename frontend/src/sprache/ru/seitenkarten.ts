// Die Texte der **Kartenseite** — die größte Seite der Anwendung.
//
// Aus `seiten.ts` herausgelöst, als die Datei über 400 Zeilen wuchs (§2).
// Der Schnitt folgt der Frage: dort stehen die Texte **aller** Seiten, hier
// die der einen, die am meisten trägt — und weiter wachsen wird
// (#674).

import { verlaufEn } from './seitenverlauf'
import { massenvorschauEn } from './seitenmassen'
import { autoersetzung } from './seitenautoersetzung'
import { schluesselwoerter } from './seitenschluesselwoerter'
import { werkstatt } from './seitenwerkstatt'
import { konsole } from './seitenkonsole'
import { symbolbibliothek } from './seitensymbolbibliothek'

export const karten = {
  ...werkstatt,
  autoersetzung,
  schluesselwoerter,
  symbolbibliothek,
  anlegen: 'Добавить карту',
  nachOben: 'Переместить выше',
  nachUnten: 'Переместить ниже',
  duplizieren: 'Дублировать карту',
  loeschen: 'Удалить карту',
  loeschenFrage: 'Удалить эту карту? Удаление можно отменить сочетанием ⌘Z.',
  loeschenJa: 'Удалить',
  loeschenNein: 'Отмена',
  loeschenTrotzdem: 'Всё равно удалить',
  symbole: 'Вставить символы',
  hergeleitet: 'Вычислено',
  hergeleitetWert: 'Вычислено: {wert}',
  bildWaehlen: 'Выбрать изображение',
  stilfelder: 'Поля стиля',
  stilfelderHilfe:
    'Эти поля приходят вместе со стилем карты. Их значение действует только для этого стиля.',
  symbolZeichnen: 'Нарисовать символ',
  kartenskripteTitel: 'Карты из скрипта',
  kartenskripteErklaerung:
    'Готовые действия из игрового пакета — один щелчок добавляет сразу несколько карт. ' +
    'Одна отмена возвращает их все.',
  kartenskriptEine: 'Добавлена одна карта.',
  kartenskriptMehrere: 'Добавлено карт: {anzahl}.',
  bildEinsetzen: 'Вставить изображение',
  pack: {
    ueberschrift: 'Случайный пак',
    erklaerung:
      'Бустер из карт этого набора по правилам игрового пакета. Одно и то же начальное значение даёт ту же выборку — так результат можно воспроизвести.',
    laedtGrund:
      'Каждый тип пака спрашивает каждую карту, подходит ли она, — в первый раз это занимает мгновение, потом всё готово.',
    art: 'Тип пака',
    same: 'Начальное значение',
    sameZahl: 'Начальное значение — целое число от 0 до 18446744073709551615.',
    ziehen: 'Вытянуть',
    neuerSame: 'Новое начальное значение',
    gezogen: 'Вытянуто',
    keineArten:
      'Игра этого набора не содержит типов паков — или её пакет не установлен.',
    keineKarten: 'Этот тип пака не дал ни одной карты.',
    artMitZahl: '{name} (карт: {zahl})',
    nurBaustein: 'Составной блок — используется другими типами паков',
    vorbehalte: 'Типы паков, чей фильтр не удалось вычислить',
  },
  symbol: {
    ueberschrift: 'Символ',
    erklaerung:
      'Символ этого набора — нарисованный, а не загруженный. Он сразу появляется в предпросмотре и попадает в файл при сохранении набора. Каждую фигуру можно разместить и переместить с клавиатуры.',

    grundform: 'Добавить базовую фигуру',

    verlauf: 'История рисования',
    nachzeichnen: 'Обвести изображение',
    groesser: 'Крупнее',
    nachzeichnenHilfe:
      'PNG, JPEG или GIF с чётким силуэтом превращается в фигуры с точками Безье, которые можно дальше править здесь.',
    zuruecknehmen: 'Отменить жест',
    wiederherstellen: 'Повторить жест',
    kreis: 'Круг',
    rechteck: 'Прямоугольник',
    dreieck: 'Треугольник',
    stern: 'Звезда',

    teile: 'Части',
    keineTeile:
      'Пока ничего не нарисовано. Начни с базовой фигуры — каждую точку можно потом изменить.',
    teilLoeschen: 'Удалить часть',
    teilMitPunkten: '{name} — точек: {zahl}',
    teilMitKopien: '{name} — симметрия, копий: {zahl}',
    gruppieren: 'Сгруппировать',
    gruppierenHilfe:
      'Две или более выбранные части становятся группой. Удерживай ⌘, чтобы выбрать несколько.',
    aufloesen: 'Разгруппировать',
    nachOben: 'Выше',
    nachUnten: 'Ниже',
    nachObenFuer: 'Поднять «{name}» выше',
    nachUntenFuer: 'Опустить «{name}» ниже',
    teilLoeschenFuer: 'Удалить «{name}»',
    teilMitKindern: '{name} — частей: {zahl}',
    teilKind: '{name}, внутри группы',
    wahlHilfe:
      'Удерживай ⌘, чтобы выбрать несколько частей. Порядок — это порядок рисования: то, что стоит первым, лежит сверху.',

    zeichnung: 'Область рисования',
    punkt: 'Точка {nummer} — {name}',
    griffDavor: 'Маркер перед точкой {nummer}',
    griffDanach: 'Маркер после точки {nummer}',
    punktHinzufuegen: 'Вставить точку',
    punktLoeschen: 'Удалить точку',

    tastatur: 'С клавиатуры',
    tastaturhilfe:
      'Tab переходит между точками. Стрелки смещают на одну сотую, с Shift — на одну десятую. Enter вставляет новую точку после выбранной, Backspace удаляет её.',

    kombination: 'Комбинировать',
    ueberdecken: 'Наложение',
    vereinen: 'Объединение',
    abziehen: 'Вычитание',
    schneiden: 'Пересечение',
    unterschied: 'Разность',
    umranden: 'Контур',

    symmetrie: 'Симметрия',
    symmetrieAnlegen: 'Добавить симметрию',
    symmetrieAufloesen: 'Снять симметрию',
    symmetrieErklaerung:
      'Симметрия размножает всё, что лежит внутри неё. Копии показываются на холсте бледно; редактируется всегда только оригинал.',
    symmetrieHinweis: 'Выбери часть, чтобы превратить её в симметрию.',
    symmetrieart: 'Вид',
    drehung: 'Поворот',
    achse: 'Отражение',
    kopien: 'Копии',
    kopienHinweis:
      'Отражению нужны пары — при нечётном числе одна копия остаётся без пары, а MSE округляет вниз.',
    mitteX: 'Центр, по горизонтали',
    mitteY: 'Центр, по вертикали',
    griffX: 'Маркер, по горизонтали',
    griffY: 'Маркер, по вертикали',
    mitte: 'Центр симметрии',
    griff: 'Маркер симметрии',

    sichern: 'Поместить в набор',
    gesichert: 'Сохранено как {name}. Попадёт в файл при сохранении набора.',
  },
  konsole,

  wortlisten: {
    ueberschrift: 'Варианты выбора',
    liste: 'Выбрать из {liste}',
    keine: 'ничего из этого',
    gruppe: '{name} (группа)',
  },
  regeln: {
    ueberschrift: 'Разъяснения правил от Scryfall',
    holen: 'Загрузить разъяснения',
    keine: 'Для этой карты нет разъяснений.',
    offiziell: 'официальное',
    anmerkung: 'Примечание Scryfall',
    anmerkungErklaerung:
      'Примечания исходят от команды Scryfall и часто поясняют неофициальный формат. Правилами они не являются.',
    quelle: 'Открыть эту карту на Scryfall',
    alter: 'Возраст данных: {stunden} ч',
    manaPruefen: 'Проверить мановую стоимость',
    manaGut: 'Мановая стоимость в порядке.',
    manaSchlecht: 'Неизвестно: {symbole}',
    manaLeer: 'У этой карты нет мановой стоимости.',
  },
  abgleich: {
    ueberschrift: 'Данные карты из Scryfall',
    erklaerung:
      'Показывает, что изменится. Записывается только то, что отмечено.',
    holen: 'Загрузить предложение',
    uebernehmen: 'Применить отмеченное',
    nichtsZuTun: 'Менять нечего — карта совпадает со Scryfall.',
    alt: 'сейчас',
    neu: 'новое',
    leer: '(пусто)',
    alter: 'Состояние: возраст {stunden} ч',
    frisch: 'Состояние: только что загружено',
    quelle: 'Открыть эту карту на Scryfall',
  },
  scryfall: {
    ueberschrift: 'Изображение со Scryfall',
    erklaerung:
      'Изображение выбирается по названию карты. Художник тоже записывается в карту — Scryfall требует указывать имя рядом с иллюстрацией.',
    name: 'Название карты на Scryfall',
    art: 'Тип изображения',
    artKlein: 'Маленькое (предпросмотр)',
    artNormal: 'Вся карта',
    artAusschnitt: 'Только иллюстрация',
    holen: 'Загрузить изображение',
    ausschnittHinweis:
      'Вариант «Только иллюстрация» показывает одну иллюстрацию. Он отклоняется, если Scryfall не называет художника.',
    quelle: 'Открыть эту карту на Scryfall',
  },
  scryfallkarte: {
    name: 'Название карты на Scryfall',
    platzhalter: 'например, Lightning Bolt',
    mitBild: 'с иллюстрацией',
    anlegen: 'Создать из Scryfall',
  },
  zuschnitt: {
    ueberschrift: 'Выбор кадрирования',
    erklaerung:
      'Рамка показывает, какая часть изображения попадёт на карту. Перетащи её, чтобы сдвинуть, или потяни за угол, чтобы изменить размер.',
    quelle: 'Исходное изображение с выбранным кадрированием',
    vorschau: 'Предпросмотр в размере поля',
    links: 'Слева',
    oben: 'Сверху',
    breite: 'Ширина',
    hoehe: 'Высота',
    seitenverhaeltnis: 'Сохранять соотношение сторон поля',
    uebernehmen: 'Вставить',
    abbrechen: 'Отмена',
    masse: '{breite} × {hoehe} пикселей',
  },
  feldWaehlen: 'Редактировать это поле',
  leerTitel: 'Карт пока нет',
  listeBezeichnung: 'Карты набора',
  mehr: 'Ещё',
  gruppeAnlage: 'Карта',
  gruppeAnsicht: 'Вид',
  ausScryfallAufklapp: 'Создать из Scryfall …',
  ausScryfallErklaerung:
    'Вместо пустой карты: введи название или найди карту на Scryfall — поля тогда берутся оттуда.',
  dieseKarte: 'Эта карта',
  kartenmenue: 'Карта',
  auswahlLoesen: 'Показать все карты',
  bezugAufklapp: 'Связать с печатью …',
  dieseKarteLeer: 'Открой карту, чтобы её редактировать.',
  keineKarteTitel: 'Карта не выбрана',
  titel: 'Карты',
  leer: 'Создай карту или вставь скопированные.',
  ohneName: 'Без названия',
  ohneFelder: 'У этой карты вообще нет значений полей.',
  nummer: 'Номер',
  vorschau: 'Предпросмотр карты',
  keineKarte: 'Выбери карту в списке — здесь появятся её поля и предпросмотр.',
  bearbeiten: 'Редактировать',

  notizen: {
    ueberschrift: 'Заметки к карте',
    platzhalter: 'Только для тебя — на карте это никогда не появляется.',
  },

  // **What the game package does not declare** (#1271).
  buchfuehrung: 'Что приложению известно об этом',
  tiefe: 'Сколько ты видишь',
  tiefeNormal: 'Обычный',
  tiefeErweitert: 'Расширенный',
  tiefeExperte: 'Эксперт',
  tiefeErweitertTitel: 'Что заполняет скрипт',
  tiefeErweitertHinweis:
    'Игровой пакет вычисляет эти значения — они объясняют, почему что-то есть на карте. Их нельзя редактировать.',
  befundVerdeckt: 'Замечание скрыто на уровне «{tiefe}».',
  befundSpringen: 'Перейти',
  buchfuehrungHinweis:
    'Эти записи относятся не к самой карте, а к её учёту — какой шаблон она использует и что на ней задано. Меняй их, только если знаешь, зачем.',

  // **Note and marks together, collapsed** (#1271). They sat below the list
  // and took its height: four rows were left.
  arbeitsstand: 'Ход работы над этой картой',
  arbeitsstandHinweis:
    'Заметка и метки относятся к твоей работе над картой, а не к самой карте. В файле, который ты передаёшь дальше, их нет.',

  marken: {
    ueberschrift: 'Метки',
    erklaerung:
      'Слова для твоей собственной работы — они находятся рядом с картой, а не внутри неё. В файле, который ты передаёшь дальше, их нет.',
    platzhalter: 'например, готово',
    setzen: 'Добавить метку',
    entfernen: 'Удалить метку {marke}',
    keine: 'На этой карте ещё нет меток.',
    schonGesetzt:
      'У карты уже есть метка «{marke}» — её поставили в другом месте.',
    filter: 'Фильтр по метке',
    alle: 'Все',
  },

  scryfallsuche: {
    frage: 'Поиск на Scryfall',
    platzhalter: 'например, Lightning Bolt — или t:creature c:r cmc<=3',
    erklaerung:
      'Достаточно названия карты. Если ты знаешь язык поиска Scryfall, можно использовать его здесь.',
    suchen: 'Найти',
    nichtsGefunden: 'Ничего не найдено. Попробуй другое слово?',
    mehr: 'На Scryfall результатов больше. Уточни запрос — автоматически ничего не загружается.',
    alter: 'Состояние: возраст {stunden} ч',
    frisch: 'Состояние: только что загружено',
    befund: 'Замечание: {text}',
  },

  // **Der Verweis auf einen echten Druck** (#1167).
  verweis: {
    aufschrift: 'Связано с',
    keiner:
      'Эта карта не связана ни с одним изданием. Синхронизация со Scryfall тогда выберет любое — найди его выше и свяжи.',
    druck: 'из {setcode} {nummer}',
    gesetzt: '— связано {wann}',
    setzen: 'Связать с {name} ({setcode} {nummer})',
    loesen: 'Разорвать связь',
  },
  // **Die Zwischenablage** (#1168).
  ablage: {
    kopieren: 'Копировать ({zahl})',
    ausschneiden: 'Вырезать ({zahl})',
    einfuegen: 'Вставить',
  },

  // **Die Spalten der Kartenliste** (#1168).
  listenspalten: {
    ueberschrift: 'Столбцы списка',
    jeSpiel: 'Выбор действует для всех наборов одной игры.',
    filtern: 'Найти столбец',
    gezeigt: '{gezeigt} из {alle} столбцов',
    ausgeblendet: 'Ещё {anzahl} скрыто — найди их через поле выше.',
    genauer: 'Подходит больше — уточни поиск, чтобы их увидеть.',
    nachOben: 'сдвинуть левее',
    nachUnten: 'сдвинуть правее',
    sortieren: 'Сортировать по',
    garNicht: 'Порядок набора',
    aufsteigend: 'По возрастанию',
    absteigend: 'По убыванию',
    ohneVorlage:
      'Игровой пакет не удаётся прочитать. Столбцы берутся из самих карт — без заголовков и без порядка, заданного пакетом.',
    spalteAusblenden: 'Скрыть столбец',
  },

  verlauf: verlaufEn,
  massen: {
    ...massenvorschauEn,
    keinFeld: '— выбери поле —',
    keinWert: '— выбери значение —',
    ueberschrift: 'Массовое изменение',
    erklaerung:
      'Задаёт одно поле у каждой отмеченной карты — одной операцией. Если что-то пойдёт не так, не изменится ни одна.',
    auswaehlen: 'Отметить для массового изменения',
    mehrereWaehlen: 'Выбрать несколько',
    ausgewaehlt: 'отмечено',
    keineAuswahl:
      'Ни одна карта не отмечена. Отметь несколько в списке, чтобы изменить их вместе.',
    feld: 'Поле',
    wert: 'Новое значение',
    anwenden: 'Применить',
    betroffen: 'Изменение не удалось для этих карт. Ни одна карта не изменена:',
    gescheitert: 'не удалось',
    ohneWert: 'значения пока нет',
  },
  stilTitel: 'Стиль',
  stilWaehlen: 'Стиль этого набора',
  stilLeer: 'Стиль не установлен. Без него предпросмотр использует заглушку.',
  stilFrage: 'Переключиться на этот стиль?',
  stilSpielwechsel: 'Новый стиль относится к другой игре.',
  stilUngedeutet:
    'Новая игра не знает этих полей. Их значения остаются в базе данных, но больше отрисовываться не будут:',
  stilVerloreneOptionen:
    'В новом стиле нет таких параметров оформления. Их значения будут потеряны:',
  stilUebernehmen: 'Сменить стиль',
  stilAbbrechen: 'Оставить стиль',
  stilKeiner:
    'Набор указывает стиль, который не установлен, — выбери один из установленных.',
  stiloptionen: 'Параметры оформления',
  stiloptionenSichtbar: '{zahl} из {gesamt}',
  stiloptionenWeitere: 'Другие параметры ({zahl})',
  marktTitel: 'Где найти эту карту',
  marktErklaerung:
    'Эти ссылки формируются на твоём компьютере — из названия карты или идентификатора поставщика. Пока ты не щёлкнешь, ничего не отправляется, а открывается ссылка в браузере по твоему выбору. Ключ не нужен; продаётся ли там что-нибудь, зависит от поставщика.',
  marktGenau: 'ведёт прямо на эту карту',
  marktSuche: 'ведёт на поиск',
  marktNurAnbieter: 'ведёт только на сайт поставщика',
  umschlagAnnehmen: 'Карта из конверта',
  umschlagAngenommen: 'Карта теперь в этом наборе.',
}
