// Die Texte der **Werkzeugseiten** — Konsole und Schlüsselwörter.
//
// Aus `seiten.ts` herausgelöst, als die Datei über 400 Zeilen wuchs (§2).
// **Ein Katalog darf sich auf mehrere Dateien verteilen**, solange er als
// **einer** ankommt — an der Zahl der Kataloge hängt der Vertragstest gegen
// die Go-Seite, nicht an der Zahl der Dateien.

export const werkzeugseiten = {
  pack: {
    titel: 'Pack aléatoire',
    erklaerung:
      'Un booster composé des cartes de ce set, selon les règles du paquet de jeu.',
  },
  symbol: {
    titel: 'Symbole',
    erklaerung: 'Le symbole de ce set — dessiné, non chargé.',
  },
  konsole: {
    titel: 'Console',
    erklaerung:
      'Une expression MSE Script dans le contexte de la carte ouverte.',
  },
  schluesselwoerter: {
    titel: 'Mots-clés',
    erklaerung:
      'Les mots-clés de ce set — et ceux que son jeu apporte avec lui.',
  },
}
