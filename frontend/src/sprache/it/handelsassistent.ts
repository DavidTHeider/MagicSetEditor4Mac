// **A trade in four phases** — TRADE-02
// (#1716).
// Maschinell übersetzt.
export const handelsassistent = {
  schrittleiste: 'A che punto è questo scambio',
  phasen: {
    gegenueber: 'Con chi',
    angebot: 'Che cosa scambiate',
    zusagen: 'Riserve',
    verbuchen: 'Registrazione',
  },
  erledigt: 'fatto',
  jetzt: 'adesso',
  kommt: 'a seguire',
  schritte: 'Che cosa puoi fare adesso',
  keinSchritt: 'Al momento qui non c’è nulla da fare per te.',
  neu: 'Un nuovo scambio: metti insieme ciò che va e viene, poi avvialo.',
  abgleich: 'Che cosa potreste scambiare',
  unterwegs: 'Che cosa è in viaggio',

  // deinGegenueber setzt sich in „Wartet auf {name}" ein — der Raum
  // gehört ohnehin einem Gegenüber (TRADE-02).
  deinGegenueber: 'risposta della controparte',

  // Das Protokoll eines Handels (TRADE-03.AK-03).
  verlauf: 'Cronologia',
  verlaufWechsel: 'Revisione {revision}: {von} → {nach} · {zeit}',
  verlaufNeu: 'Revisione {revision}: {nach} · {zeit}',
  verlaufLeer: 'Ancora nessun cambiamento.',
  verlaufOhneHandel:
    'Scegli uno scambio — qui vedrai come si è arrivati a questo punto.',

  // Die zwei Spalten eines Tauschs mit Preisen (TRADE-02b, #1951).
  zeileIn: 'Aggiungi una riga: {seite}',
  zeileKurz: '+ Riga',
  keinPreis: 'senza prezzo',
  differenzDir: '⇆ {betrag} a tuo favore',
  differenzGegenueber: '⇆ {betrag} a favore della tua controparte',
  ausgeglichen: '⇆ Bilanciato',
  differenzTeil: '{zahl} senza prezzo — si calcola solo ciò che ne ha uno.',
  preisHinweis:
    'Prezzi del {datum}, da {quelle}. Sono un’indicazione, non un vincolo.',
  keinePreise: 'Per queste carte non ci sono ancora prezzi.',
  zeileVerwerfen: 'Annulla',
}
