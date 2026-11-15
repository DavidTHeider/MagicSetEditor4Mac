// Die Texte der **Werkzeugseiten** — Konsole und Schlüsselwörter.
//
// Aus `seiten.ts` herausgelöst, als die Datei über 400 Zeilen wuchs (§2).
// **Ein Katalog darf sich auf mehrere Dateien verteilen**, solange er als
// **einer** ankommt — an der Zahl der Kataloge hängt der Vertragstest gegen
// die Go-Seite, nicht an der Zahl der Dateien.

export const werkzeugseiten = {
  pack: {
    titel: 'Bustina casuale',
    erklaerung:
      'Una bustina con le carte di questo set, secondo le regole del pacchetto di gioco.',
  },
  symbol: {
    titel: 'Simbolo',
    erklaerung: 'Il simbolo di questo set — disegnato, non caricato.',
  },
  konsole: {
    titel: 'Console',
    erklaerung: 'Un’espressione MSE Script nel contesto della carta aperta.',
  },
  schluesselwoerter: {
    titel: 'Parole chiave',
    erklaerung:
      'Le parole chiave di questo set — e quelle che porta con sé il suo gioco.',
  },
}
