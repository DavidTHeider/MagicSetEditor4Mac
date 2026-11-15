// The **collection** page — piece 2 and 3 of phase 3.4
// (#1165).
export const sammlung = {
  titel: 'Коллекция',
  erklaerung:
    'То, чем ты действительно владеешь, — одна строка на издание, состояние, язык и отделку. К твоим наборам это не имеет отношения: разработанная карта — это черновик, а экземпляр — кусок картона.',

  aufnehmenOeffnen: 'Добавить …',
  einfuhrOeffnen: 'Импорт и экспорт …',
  einfuhrTitel: 'Импорт и экспорт',
  leerTitel: 'Пока ничего не записано',
  leer: 'Для экземпляра нужен Scryfall ID его издания — UUID из адреса страницы карты. Целый список загружается за один раз.',

  druck: 'Издание',
  menge: 'Количество',
  zustand: 'Состояние',
  sprache: 'Язык',
  ausfuehrung: 'Отделка',
  foil: 'Foil',
  nonfoil: 'Nonfoil',
  etched: 'Etched',
  erwerbsdatum: 'Дата приобретения',
  erwerbspreis: 'Заплачено',
  erwerbswaehrung: 'Валюта',
  preisMehrdeutig:
    'Возможны два прочтения: {erste} или {zweite}. Запиши однозначно.',
  notiz: 'Заметка',

  // **Der lesbare Bezug und die Sicht darauf** (#1200).
  name: 'Название карты',
  setcode: 'Набор',
  nummer: 'Номер',
  sicht: 'Вид коллекции',
  suche: 'Поиск',
  suchePlatzhalter: 'название, набор, id или заметка',
  sortierung: 'Сортировать по',
  sortierBezeichnung: 'Карте',
  sortierMenge: 'Количеству',
  sortierErwerbsdatum: 'Дате приобретения',
  ausfuehrungAlle: 'Все отделки',
  nurFoil: 'Только Foil',
  nurNonfoil: 'Только Nonfoil',
  nurEtched: 'Только Etched',

  navigator: 'Вид и инструменты',
  inspector: 'Экземпляр',
  exemplarmenue: 'Экземпляр',
  inDecks: 'В этих колодах',

  inspectorLeer:
    'Выбери экземпляр в списке — здесь появятся количество, состояние, покупка и цена, и здесь же ты их меняешь.',
  inspectorAusserhalb:
    'Этого экземпляра больше нет в текущем виде. Изменение всё равно затронет именно этот экземпляр.',
  arbeitsplatz: 'Рабочее пространство коллекции',
  listeBezeichnung: 'Экземпляры',
  spalteWahl: 'Выбор',
  spalteBezeichnung: 'Название',
  spalteMenge: 'Количество',
  spalteZustand: 'Состояние',
  spalteSprache: 'Язык',
  spalteSet: 'Набор',
  spalteLagerort: 'Хранение',
  gruppeOhne: 'Не указано',
  gruppeSumme: 'Экземпляров: {anzahl} · штук: {menge}',
  kachelWaehlen: 'Выбрать: {bezeichnung}',

  ansicht: 'Вид',
  ansichtTabelle: 'Таблица',
  ansichtGalerie: 'Галерея',
  ansichtGruppen: 'Группы',
  gruppierung: 'Группировать по',
  gruppierungKeine: 'Без группировки',
  gruppierungSet: 'Набор',
  gruppierungLagerort: 'Хранение',
  gruppierungZustand: 'Состояние',
  gruppierungSprache: 'Язык',
  ohneTreffer: 'Ни один экземпляр не соответствует поиску и фильтру.',
  ohneTrefferZuruecksetzen: 'Сбросить поиск и фильтр',
  gezeigt: '{gezeigt} из {gesamt}',
  gekappt: 'Показано {gezeichnet} из {passend}',
  mehrZeigen: 'Показать больше',

  // **Die Listeneinfuhr** (#1292).
  listeTitel: 'Вставить список',
  listeErklaerung:
    'Одна карта в строке, например «4 Lightning Bolt (M11) 149». Пустые строки, «Deck», «Sideboard» и строки, начинающиеся с # или //, пропускаются.',
  listeFeld: 'Список карт',
  listePlatzhalter: '4 Lightning Bolt (M11) 149\n2x Counterspell',
  listeDeuten: 'Сопоставить',
  listeEinlesen: 'Добавить',
  listeOffen:
    'Строк, требующих решения: {offen} из {gesamt} — выбери издание, которым ты владеешь.',
  listeAlleEindeutig: 'Все строки ({gesamt}) однозначны.',
  listeDruckWahl: 'Какое издание?',
  listeOhneWahl: '— пока не выбрано —',
  listeDrucke: 'Изданий: {drucke}',
  listeBericht:
    'Новых: {angelegt}, увеличено: {erhoeht}, добавлено экземпляров: {stueck}.',
  listeNichtsGewaehlt:
    'Ни одна строка не выбрана — выбери хотя бы одно издание.',
  listeAusgelassen: 'Пропущено существующих строк: {ausgelassen}.',
  listeKeinDruck: 'Карты с таким названием нет — проверь написание.',
  listeZuViele:
    'Изданий с таким названием больше {drucke} — укажи сет в скобках, например «(M11)», и останется одно.',
  listeOhneKarte: 'Строк без найденной карты: {anzahl}; они пропускаются.',
  listeOhneKarteEins: 'Одна строка не находит карту и пропускается.',
  vorschauTitel: 'Перед импортом',
  vorschauErklaerung:
    'Некоторые строки совпадают со строками, которые в твоей коллекции уже есть, — то же издание, состояние, язык, отделка и назначение. Выбери, что с ними делать; до тех пор ничего не записывается.',
  vorschauZaehlung:
    'Новых: {neu} · совпадают с существующими строками: {zusammen} · повторяют строку этого импорта: {wiederholt}',
  vorschauZusammen: '{bezeichnung}, {zustand}: {vorher} → {nachher}',
  vorschauModus: 'Существующие строки',
  vorschauAddieren: 'Сложить количества',
  vorschauAuslassen: 'Пропустить существующие строки',
  vorschauEinlesen: 'Импортировать',
  vorschauAbbrechen: 'Не импортировать',

  // **Die Ausfuhr** (#1292).
  ausfuhrTitel: 'Экспорт и резервное копирование',
  ausfuhrErklaerung:
    '«CSV» содержит все поля и читается обратно полностью. «Список» — вид для вставки в конструктор колод: он содержит только количество, карту, издание и отделку.',
  ausfuhrAlsCSV: 'В виде CSV',
  ausfuhrAlsListe: 'В виде списка',
  ausfuhrErgebnis: 'Результат',
  ausfuhrZurueck: 'Прочитать CSV обратно',
  ausfuhrLeer: 'Пока ничего не экспортировано.',

  // **Die Ordnung** (#1292, Vermerk 0057).
  lagerort: 'Место хранения',
  wunsch: 'Желаемое, но не в наличии',

  // **Die Ordnungssicht** (#1292).
  ordnungTitel: 'Порядок',
  ordnungLagerorte: 'Места хранения',
  ordnungDubletten: 'Дубликаты',
  ordnungWuensche: 'Список желаемого',
  ordnungBehalten: 'Оставить',
  ordnungLeerLagerorte: 'Экземпляров пока нет — без них нет и мест хранения.',
  ordnungLeerDubletten: 'Ни один выпуск не лежит больше {behalten} раз.',
  ordnungLeerWuensche:
    'Список желаний пуст. Экземпляр попадает туда как «Желаемое, но не в наличии» — или из колоды через «Добавить недостающее в список желаний».',
  ordnungStueck: 'Экземпляров: {stueck}, строк: {zeilen}',
  ordnungOhneOrt: 'не указано',
  ordnungUeberzaehlig: 'Лишних: {ueberzaehlig}',

  // **Besitze ich diese Karte?** (#1292).
  besitzJa: 'У тебя есть таких: {stueck}.',
  besitzNein: 'У тебя нет этого издания.',
  besitzGewuenscht: '(в списке желаемого: {gewuenscht})',

  // **Die Massenvorgänge** (#1292).
  masseTitel: 'Массовые действия',
  masseWaehlen: 'Выбрать',
  masseGewaehlt: 'Выбрано {gewaehlt} из {gesamt}',
  masseLagerortSetzen: 'Задать место хранения',
  masseZustandSetzen: 'Задать состояние',
  masseAendern: 'Изменить выбранное',
  masseLoeschen: 'Удалить выбранное',
  masseLoeschenFrage: 'Удалить экземпляры ({anzahl})?',
  masseLoeschenErklaerung:
    'То, что удаляется, тебе придётся вводить заново: количество, состояние, место хранения и цена покупки больше нигде не сохранены. У коллекции нет истории, как у набора.',
  masseLoeschenJa: 'Удалить',
  masseBericht: 'Изменено: {geaendert}.',
  masseZuruecknehmen: 'Отменить',
  masseZuruecknehmenVon:
    'Отменить изменение места хранения для строк: {anzahl}',
  masseZurueckgenommen: 'Отменено: {anzahl}.',
  massePreise: 'Загрузить цены для всех',
  massePreisschub: 'Загружено: {geholt}, осталось: {offen}',
  massePreiseFehlgeschlagen: 'Не удалось: {anzahl}',
  massePreiseSteht:
    'Эта порция ничего не сдвинула — ещё одна попытка даст то же самое.',
  massePreiseFertig: 'Все цены получены.',

  aufnehmen: 'Добавить',
  aenderungSichern: 'Сохранить',
  loeschen: 'Удалить',
  loeschenFrage: 'Удалить этот экземпляр?',
  loeschenErklaerung:
    'Количество, состояние, место хранения и цена покупки больше нигде не сохранены — у коллекции нет истории, как у набора.',
  loeschenJa: 'Удалить',
  abbrechen: 'Отмена',
  bearbeiten: 'Изменить',
  zumTausch: 'В обмен',
  insDeck: 'В колоду',
  auswahlTitel: 'Выбор',
  auswahlInsDeck: '{anzahl} → колода',

  neuesTitel: 'Добавить экземпляр',
  weitereAngaben: 'Другие сведения',

  dubletteFrage:
    '{bezeichnung} уже есть в коллекции — {menge} в том же состоянии, на том же языке и с той же отделкой.',
  dubletteErklaerung:
    'Увеличение прибавляет новое количество; остальные данные существующей строки остаются. Отмена ничего не меняет — твой ввод сохраняется.',
  dubletteErhoehen: 'Увеличить до {summe}',
  dubletteZurZeile: 'Перейти к существующей строке',
  dubletteAbbrechen: 'Отмена',

  zustaende: {
    mint: 'Mint — идеальное',
    nearMint: 'Near Mint — почти идеальное',
    excellent: 'Excellent — отличное',
    good: 'Good — хорошее',
    lightPlayed: 'Light Played — слегка игранное',
    played: 'Played — игранное',
    poor: 'Poor — плохое',
  },

  preiseTitel: 'Цены',
  preiseHolen: 'Загрузить цены',
  marktUndVerlauf: 'Предложения и история',
  // **The marketplace** — work item A4 of phase 5.3 (#1443).
  marktTitel: 'Предложения торговой площадки',
  marktErklaerung:
    'Что отдельные продавцы просят за это издание — цена, количество, состояние и язык. Запрос уходит в сеть. Ничего не покупается.',
  marktHolen: 'Загрузить предложения',
  marktHolenBei: 'Загрузить предложения из {quelle}',
  marktLaeuft: 'Загрузка …',
  marktLeer: 'Сейчас там нет ничего в продаже для этого издания.',
  marktZugeordnet:
    'Это издание теперь связано с версией карты {fassung} — следующий запрос найдёт его сразу.',
  marktStueck: 'В наличии: {menge}',
  marktBuendel: 'комплектами по {buendel}',
  marktImUrlaub: 'сейчас не отправляет',
  marktGradiert: 'оценена (graded)',
  marktSigniert: 'с автографом',
  marktVeraendert: 'альтерированная',
  marktBefunde: 'Не все предложения удалось прочитать:',

  // **Die Begleitbefunde des Preisholens** (#1202).
  preisbefunde: 'Не все цены удалось прочитать. Отсутствуют эти ряды:',
  preiseLaufen: 'Загрузка …',
  keinePreise: 'Для этого издания нет записанной цены.',
  preisAlter: '{stunden} ч назад · {quelle}',
  preisFrisch: 'только что · {quelle}',
  preisVeraltet: 'устарело: {tage} дн. · {quelle}',
  preisQuellstand: 'данные источника на {tag}',

  vorbehalt:
    'Цены — это оценки Scryfall и его партнёров. Гарантий по ним не даётся; решающей является цена в магазине.',

  verlauf: 'История',
  keinVerlauf:
    'Для этого ряда есть только один снимок — история накапливается по мере повторной загрузки цен.',

  bulkTitel: 'Загрузить цены из целого источника',
  leiste: {
    bezeichnung: 'Показатели коллекции',
    exemplare: 'Экземпляры',
    eintraege: 'записей: {zahl}',
    wert: 'Стоимость',
    keinPreis: 'нет цены',
    ohnePreis: 'Без цены',
    nichtGelesen: 'не прочитано',
  },
  drucksuche: {
    karte: 'Карта',
    hinweis:
      'Название, при желании с выпуском и номером — напр. Lightning Bolt (M11) 149',
    suchen: 'Искать',
    drucke: 'Издания',
    keinTreffer: 'Карта с таким названием не найдена.',
    weitere: 'Есть и другие издания — выпуск и номер сужают выбор.',
    treffer: '{name} · {set} ({code}) · № {nummer}',
  },
  bulkErklaerung:
    'MTGJSON отвечает не по каждой карте, а файлом. «Сегодня» приносит текущий день, «Девяносто дней» — историю, и это больше гигабайта.',
  bulkQuelle: 'Источник',
  bulkUmfang: 'Охват',
  bulkHeute: 'Сегодня',
  bulkNeunzigTage: 'Девяносто дней',
  bulkHolen: 'Загрузить',
  bulkHolenBei: 'Загрузить цены из {quelle}',
  bulkLaeuft: 'Загрузка …',
  bulkBilanz: 'Сохранено цен: {staende} для изданий: {drucke}.',
  bulkOhneDrucke:
    'Ни у одного издания в этой коллекции нет идентификатора MTGJSON — ничего не запрашивалось.',
  uebersichtTitel: 'Обзор',
  uebersichtLeer: 'Цены пока не загружены.',
  ohnePreis: 'без цены: {zahl}',
  // **Ein Exemplar und nicht „1 Exemplare“.** Eine Zahl mit falscher
  // Form liest sich wie ein Fehler im Programm — gesehen im laufenden
  // Fenster.
  einExemplar: '{zahl} экземпляр',
  exemplare: '{zahl} экз.',

  keineUmrechnung:
    'Одна сумма на валюту. Без пересчёта — для него нужен обменный курс, а у этой программы его нет.',
  // **A price history from another source** — A5, #1444.
  verlaufTitel: 'История цен из другого источника',
  verlaufErklaerung:
    'Загружает ряд по времени и сохраняет его. Запрос уходит в сеть и расходует квоту; состояние показано так, как его называет источник.',
  verlaufHolen: 'Загрузить историю',
  verlaufHolenBei: 'Загрузить историю из {quelle}',
  verlaufLaeuft: 'Загрузка …',
  verlaufBilanz: 'Сохранено ценовых точек: {staende} — версия карты {fassung}.',
} as const
