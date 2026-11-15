// Die Texte der **Turniere** — Werkstück 5 von Phase 4.7
// (#1294),
// nach Vermerk 0062.
//
// **Ein eigener Baustein**, wie `seitendecks.ts` und aus demselben Grund:
// `seiten.ts` steht an der Zeilengrenze aus §2.
export const turniere = {
  alleTurniere: 'Все турниры ({anzahl})',
  titel: 'Турниры',
  erklaerung:
    'Турнир — это ход событий во времени: раунды, пары, результаты. Турнирная таблица вычисляется из него и нигде не хранится — исправь результат, и ты сразу увидишь новую таблицу.',

  leer: 'Турниров пока нет. «Создать турнир …» вверху начинает новый, «Импортировать турнир …» берёт его из файла.',

  neu: 'Создать турнир',
  neuOeffnen: 'Создать турнир …',
  name: 'Название',
  spiel: 'Игра',
  notiz: 'Заметка',
  einzelspiel: 'Одна партия в матче',
  einzelspielHinweis:
    'Тогда процент выигранных партий и процент выигранных партий соперниками ничего не говорят — они повторяют показатели матчей. Они всё равно вычисляются.',
  anlegen: 'Создать',

  // **The wizard** — DOM-04.AK-01 (#1699).
  assistent: 'Новый турнир за четыре шага',
  schrittTurnier: 'Турнир',
  schrittRegeln: 'Правила',
  schrittTeilnehmende: 'Участники',
  schrittErsteRunde: 'Первый раунд',
  weiter: 'Далее',
  zurueck: 'Назад',
  verlassen: 'Продолжить позже',
  vorlage: 'Шаблон',
  ohneVorlage: 'Без шаблона',
  vorlageHinweis:
    'Шаблон приносит игру, формат, правила и заметку. При желании и имена, но никогда не колоды: поданный список принадлежит одному турниру.',
  mitTeilnehmenden: 'Перенести имена игроков',
  nameWarum:
    'Обязательно — чтобы снова найти его в списке. У двух турниров одной игры разные названия.',
  spielWarum:
    'Обязательно — от этого зависит, какие форматы и колоды подходят.',
  format: 'Формат',
  ohneFormat: 'Без формата',
  formatWarum: 'Необязательно — можно выбрать из форматов этой игры.',
  regelnAusVorlage: 'Правила берутся из шаблона.',
  schweizerSystem:
    'Пары составляются по швейцарской системе: играют друг с другом игроки с одинаковым числом очков, и никто не встречается с одним и тем же соперником дважды. Приложение предлагает число раундов исходя из числа игроков.',
  zuWenige:
    'Чтобы сыграть раунд, нужны два человека — зарегистрируй не меньше двух.',
  ersteRundeHinweis:
    'Так выглядел бы первый раунд. Пока ты не нажмёшь кнопку, пары не составляются.',
  ersteRundeAuslosen: 'Составить пары раунда 1',

  waehlen: 'Открыть',
  loeschen: 'Удалить турнир',
  handlungenAn: 'Действия — {name}',
  handlungenKurz: '…',
  turniermenue: 'Турнир',
  loeschenFrage:
    'Удалить этот турнир? Раунды, пары и результаты будут удалены вместе с ним. Колоды игроков остаются.',
  loeschenJa: 'Удалить',
  loeschenNein: 'Отмена',

  teilnehmer: 'Участники',
  teilnehmerLeer: 'Пока никто не зарегистрирован.',
  anmelden: 'Зарегистрировать',
  gebiete: 'Разделы этого турнира',
  gebietTeilnehmer: 'Участники',
  gebietRunden: 'Раунды',
  gebietRangliste: 'Турнирная таблица',

  // **Closing and log** — DOM-04.AK-04 (#1699).
  abschliessen: 'Завершить',
  abgeschlossenAm:
    'Завершён {datum}. Игроки, раунды и результаты окончательны.',
  wiedereroeffnen: 'Открыть заново',
  nichtWiedereroeffnen: 'Не открывать',
  wiedereroeffnenFrage: 'Открыть этот турнир заново?',
  wiedereroeffnenErklaerung:
    'После этого игроков, раунды и результаты снова можно менять. В протоколе записывается, что это произошло.',
  gebietProtokoll: 'Протокол',
  protokollLeer: 'Пока ничего не происходило.',
  protokollAngelegt: 'Создан как «{angaben}»',
  protokollAngemeldet: 'Регистрация: {angaben}',
  protokollAusgestiegen: 'Выбыл: {angaben}',
  protokollEingestiegen: 'Снова в игре: {angaben}',
  protokollRunde: 'Пары раунда {angaben} составлены',
  protokollErgebnis: 'Результат {paarung}: {angaben}',
  protokollZurueckgenommen: 'Результат {paarung} отозван',
  protokollAbgeschlossen: 'Завершён',
  protokollWiedereroeffnet: 'Открыт заново',
  protokollEingelesen: 'Импортирован как «{angaben}»',

  // **File and report** — DOM-04.AK-03 (#1699).
  ausgeben: 'Экспорт …',
  berichtSichern: 'Сохранить отчёт …',
  einlesen: 'Импортировать турнир …',
  gesichertUnter: 'Сохранено в {pfad}.',
  eingelesen: 'Импортировано и открыто.',
  einlesebefunde: 'Замечено при импорте:',

  // **Header and links after the artboard** — DOM-04.AK-02 (#1699).
  kopfzeile: 'Швейцарская система · Раунд {runde} из {von} · игроков: {anzahl}',
  kopfzeileOhneRunde: 'Швейцарская система · игроков: {anzahl}',
  kopfzeileOhneEmpfehlung:
    'Швейцарская система · Раунд {runde} · игроков: {anzahl}',
  gebietDecklisten: 'Списки колод',
  decklistenLeer:
    'Никто ещё не подал список — выбор колоды при регистрации подаёт её список.',
  deckAufschlagen: 'Открыть колоду',
  partieOeffnen: 'Открыть партию',
  deck: 'Колода',
  ohneDeck: 'Без колоды',
  deckHinweis:
    'Если колода указана, её список копируется, а не связывается: список колоды в турнире — это поданный документ.',
  deckabzug: 'Поданный список',
  deckabzugLeer: 'Колода не подана.',
  deckabzugUnbenannt: 'Карта',
  aussteigen: 'Выбыть',
  einsteigen: 'Вернуться',
  ausgestiegen: 'выбыл',
  ausgestiegenHinweis:
    'Его больше не включают в пары, но он остаётся в турнирной таблице: его процент по-прежнему учитывается у его соперников.',

  runden: 'Раунды',
  rundenLeer: 'Ни один раунд ещё не составлен.',
  runde: 'Раунд {nummer}',
  empfehlung:
    'Раундов: {gelegt} из {empfohlen} — правила турниров рекомендуют {empfohlen} при числе игроков {spieler}. Это предложение, а не правило.',
  vorschlag: 'Показать следующий раунд заранее',
  vorschlagTitel: 'Так выглядел бы следующий раунд',
  vorschlagHinweis:
    'Предварительный просмотр, раунд не создан. Создаёт его только кнопка ниже.',
  rundeLegen: 'Создать раунд',
  bye: 'Бай (2–0)',
  gegen: 'против',
  offen: 'ещё не сыграно',
  spiele: 'Партии',
  spieleA: 'Партии первого',
  spieleB: 'Партии второго',
  unentschiedenSpiele: 'Ничьи',
  eintragen: 'Записать',
  zuruecknehmen: 'Отменить',
  ergebnis: 'Результат',

  // **The effects of a correction** — DOM-04.AK-02 (#1699).
  folgenFrage: 'Это исправление меняет турнирную таблицу.',
  folgenPlatz:
    '{name}: место {vorher} → {nachher}, очки {punkteVorher} → {punkteNachher}',
  folgenSpaeter:
    'Раунд {runden} был составлен по старой таблице и остаётся как есть.',
  folgenEintragen: 'Всё равно записать',
  folgenNicht: 'Не записывать',
  ergebnisZahlen: 'Партии считаются целыми числами, начиная с нуля.',
  ergebnisOhneSpiel:
    'Результат без единой партии — не результат: у кого не сыграно, у того его ещё нет.',
  ergebnisZuViele: 'Максимум партий в матче: {hoechstens}.',

  rangliste: 'Турнирная таблица',
  ranglisteLeer: 'Пока ничего не сыграно.',
  platz: 'Место',
  bilanz: 'Итог',
  matchpunkte: 'Очки',
  gegnermatchquote: 'OMW',
  spielquote: 'GW',
  gegnerspielquote: 'OGW',
  quotenHinweis:
    'Четыре тай-брейка в порядке Magic Tournament Rules: очки, процент выигранных матчей соперниками, процент выигранных партий, процент выигранных партий соперниками. У каждого процента нижняя граница 0,33.',

  // **Eine Paarung, an eine Partie gebunden** (G7, #1464).
  match: 'Сыграно здесь',
  matchErklaerung:
    'Если эта пара играется в этой программе, привяжи её к партии. Тогда результат сообщают обе стороны, и он записывается, только когда обе сообщают одно и то же.',
  matchNichtGebunden: 'К этой паре не привязана ни одна партия.',
  matchAnsehen: 'Проверить',
  matchPartie: 'Идентификатор партии',
  matchSitzA: 'Место первого игрока',
  matchSitzB: 'Место второго игрока',
  matchBinden: 'Привязать к этой партии',
  matchGebunden: 'Привязано к партии',
  matchSitz: 'Место',
  matchDeckabdruck: 'Отпечаток колоды',
  matchKeineMeldung: 'Пока никто не сообщил.',
  matchMeldungVon: 'сообщено:',
  matchMelden: 'Сообщить за это место',
  matchAusDerPartie: 'Сообщить результат из партии',
  matchAusDerPartieHilfe:
    'Счёт и доказательство берутся из завершённой партии — тебе нужно лишь указать, кто сообщает.',
  matchNachweis: 'Доказательство (последнее звено цепочки)',
  matchUebernehmen: 'Записать результат',
  matchUebernommen: 'Результат записан в турнир.',
  matchGrenze:
    'Доказательство говорит, что обе стороны видели одну и ту же историю. Оно не говорит, что партия сыграна правильно, — а частный раунд не является официальным турнирным санкционированием.',
}
