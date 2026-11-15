// **Der Handel in vier Phasen** — TRADE-02
// (#1716),
// nach dem Artboard *Handeln*.
export const handelsassistent = {
  schrittleiste: 'Wo dieser Handel steht',
  phasen: {
    gegenueber: 'Mit wem',
    angebot: 'Was tauscht ihr',
    zusagen: 'Zusagen',
    verbuchen: 'Verbuchen',
  },
  erledigt: 'erledigt',
  jetzt: 'jetzt',
  kommt: 'kommt noch',
  schritte: 'Was du jetzt tun kannst',
  keinSchritt: 'Gerade ist hier nichts von dir zu tun.',
  neu: 'Ein neuer Tausch: stell zusammen, was hin und her geht, und eröffne ihn.',
  abgleich: 'Was ihr tauschen könntet',
  unterwegs: 'Was unterwegs ist',

  // deinGegenueber setzt sich in „Wartet auf {name}" ein — der Raum
  // gehört ohnehin einem Gegenüber (TRADE-02).
  deinGegenueber: 'dein Gegenüber',

  // Das Protokoll eines Handels (TRADE-03.AK-03).
  verlauf: 'Verlauf',
  verlaufWechsel: 'Revision {revision}: {von} → {nach} · {zeit}',
  verlaufNeu: 'Revision {revision}: {nach} · {zeit}',
  verlaufLeer: 'Noch kein Wechsel.',
  verlaufOhneHandel:
    'Wähle einen Handel — dann steht hier, wie er dorthin kam.',

  // Die zwei Spalten eines Tauschs mit Preisen (TRADE-02b, #1951).
  zeileIn: 'Zeile hinzufügen: {seite}',
  zeileKurz: '+ Zeile',
  keinPreis: 'kein Preis',
  differenzDir: '⇆ {betrag} zu deinen Gunsten',
  differenzGegenueber: '⇆ {betrag} zugunsten deines Gegenübers',
  ausgeglichen: '⇆ Ausgeglichen',
  differenzTeil: '{zahl} ohne Preis — gerechnet ist nur, was einen hat.',
  preisHinweis:
    'Preise vom {datum}, aus {quelle}. Sie sind ein Anhalt, keine Vorgabe.',
  keinePreise: 'Für diese Karten liegen noch keine Preise vor.',
  zeileVerwerfen: 'Abbrechen',
}
