// Die Texte der **Werkzeugseiten** — Konsole und Schlüsselwörter.
//
// Aus `seiten.ts` herausgelöst, als die Datei über 400 Zeilen wuchs (§2).
// **Ein Katalog darf sich auf mehrere Dateien verteilen**, solange er als
// **einer** ankommt — an der Zahl der Kataloge hängt der Vertragstest gegen
// die Go-Seite, nicht an der Zahl der Dateien.

export const werkzeugseiten = {
  pack: {
    titel: 'Random pack',
    erklaerung:
      'A booster from the cards of this set, following the rules of the game package.',
  },
  symbol: {
    titel: 'Symbol',
    erklaerung: "This set's symbol — drawn, not loaded.",
  },
  konsole: {
    titel: 'Console',
    erklaerung: 'An MSE script expression in the context of the open card.',
  },
  schluesselwoerter: {
    titel: 'Keywords',
    erklaerung:
      'The keywords of this set — and the ones its game brings along.',
  },
}
