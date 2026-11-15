// **A trade in four phases** — TRADE-02
// (#1716).
export const handelsassistent = {
  schrittleiste: 'Where this trade stands',
  phasen: {
    gegenueber: 'With whom',
    angebot: 'What you swap',
    zusagen: 'Commitments',
    verbuchen: 'Booking',
  },
  erledigt: 'done',
  jetzt: 'now',
  kommt: 'still to come',
  schritte: 'What you can do now',
  keinSchritt: 'There is nothing for you to do here right now.',
  neu: 'A new trade: put together what goes back and forth, then open it.',
  abgleich: 'What you could swap',
  unterwegs: 'What is on its way',

  // deinGegenueber setzt sich in „Wartet auf {name}" ein — der Raum
  // gehört ohnehin einem Gegenüber (TRADE-02).
  deinGegenueber: 'the other side',

  // Das Protokoll eines Handels (TRADE-03.AK-03).
  verlauf: 'History',
  verlaufWechsel: 'Revision {revision}: {von} → {nach} · {zeit}',
  verlaufNeu: 'Revision {revision}: {nach} · {zeit}',
  verlaufLeer: 'No change yet.',
  verlaufOhneHandel: 'Choose a trade — then this shows how it got here.',

  // Die zwei Spalten eines Tauschs mit Preisen (TRADE-02b, #1951).
  zeileIn: 'Add row: {seite}',
  zeileKurz: '+ Row',
  keinPreis: 'no price',
  differenzDir: '⇆ {betrag} in your favour',
  differenzGegenueber: '⇆ {betrag} in the other side’s favour',
  ausgeglichen: '⇆ Even',
  differenzTeil: '{zahl} without a price — only what has one is counted.',
  preisHinweis:
    'Prices from {datum}, from {quelle}. They are a guide, not a rule.',
  keinePreise: 'There are no prices for these cards yet.',
  zeileVerwerfen: 'Cancel',
}
