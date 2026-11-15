// **A trade in four phases** — TRADE-02
// (#1716).
// Maschinell übersetzt.
export const handelsassistent = {
  schrittleiste: 'На каком этапе эта сделка',
  phasen: {
    gegenueber: 'С кем',
    angebot: 'Предмет обмена',
    zusagen: 'Удержания',
    verbuchen: 'Учёт',
  },
  erledigt: 'выполнено',
  jetzt: 'сейчас',
  kommt: 'впереди',
  schritte: 'Что можно сделать сейчас',
  keinSchritt: 'Сейчас тебе здесь делать нечего.',
  neu: 'Новый обмен: собери то, что идёт туда и обратно, и открой его.',
  abgleich: 'Что вы могли бы обменять',
  unterwegs: 'Что в пути',

  // deinGegenueber setzt sich in „Wartet auf {name}" ein — der Raum
  // gehört ohnehin einem Gegenüber (TRADE-02).
  deinGegenueber: 'партнёр',

  // Das Protokoll eines Handels (TRADE-03.AK-03).
  verlauf: 'История',
  verlaufWechsel: 'Ревизия {revision}: {von} → {nach} · {zeit}',
  verlaufNeu: 'Ревизия {revision}: {nach} · {zeit}',
  verlaufLeer: 'Изменений пока нет.',
  verlaufOhneHandel:
    'Выбери сделку — здесь будет видно, как она к этому пришла.',

  // Die zwei Spalten eines Tauschs mit Preisen (TRADE-02b, #1951).
  zeileIn: 'Добавить строку: {seite}',
  zeileKurz: '+ Строка',
  keinPreis: 'без цены',
  differenzDir: '⇆ {betrag} в твою пользу',
  differenzGegenueber: '⇆ {betrag} в пользу партнёра',
  ausgeglichen: '⇆ Равноценно',
  differenzTeil: '{zahl} без цены — считается только то, у чего она есть.',
  preisHinweis:
    'Цены на {datum}, источник: {quelle}. Это ориентир, а не предписание.',
  keinePreise: 'Для этих карт пока нет цен.',
  zeileVerwerfen: 'Отмена',
}
