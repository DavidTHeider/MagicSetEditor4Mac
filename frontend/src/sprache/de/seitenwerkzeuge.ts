// Die Texte der **Werkzeugseiten** — Konsole und Schlüsselwörter.
//
// Aus `seiten.ts` herausgelöst, als die Datei über 400 Zeilen wuchs (§2).
// **Ein Katalog darf sich auf mehrere Dateien verteilen**, solange er als
// **einer** ankommt — an der Zahl der Kataloge hängt der Vertragstest gegen
// die Go-Seite, nicht an der Zahl der Dateien.

export const werkzeugseiten = {
  pack: {
    titel: 'Zufallspack',
    erklaerung:
      'Ein Booster aus den Karten dieses Sets, nach den Regeln des Spielpakets.',
  },
  symbol: {
    titel: 'Symbol',
    erklaerung: 'Das Symbol dieses Sets — gezeichnet, nicht geladen.',
  },
  konsole: {
    titel: 'Konsole',
    erklaerung: 'Ein MSE-Script-Ausdruck im Kontext der aufgeschlagenen Karte.',
  },
  schluesselwoerter: {
    titel: 'Schlüsselwörter',
    erklaerung:
      'Die Schlüsselwörter dieses Sets — und die, die sein Spiel mitbringt.',
  },
}
