// Die Texte der **Autoersetzung** —
// #1341,
// Vermerk [0060](../../../../docs/entscheidungen/0060-eigene-autoersetzungen-liegen-im-bestand.md).
//
// **Eine eigene Datei**, weil `seitenkarten.ts` mit ihr über die Zeilengrenze
// aus §2 ginge. Der Katalog bleibt **einer**: `seitenkarten.ts` setzt sie ein,
// und der Vertragstest gegen die Go-Seite sieht dieselbe Zahl von Katalogen.

export const autoersetzung = {
  ueberschrift: 'Auto replace',
  erklaerung:
    'Replaced while you type: what is on the left becomes what is on the right. The rules from the game package are fixed — what can be changed is what they insert and whether they apply. Your own rules hold in every set of this game.',
  schalter: 'Auto replace applies',
  muster: 'Match',
  ersatz: 'Replace with',
  aktiv: 'applies',
  ganzesWort: 'whole word',
  anlegen: 'Add rule',
  loeschen: 'Delete',
  zuruecksetzen: 'Forget mine',
  zuruecksetzenFrage: 'Forget all of your own rules?',
  zuruecksetzenErklaerung:
    'Afterwards only the rules the game brings apply. Your own do not come back.',
  zuruecksetzenJa: 'Forget mine',
  ausDemPaket: 'from the game package',
}
