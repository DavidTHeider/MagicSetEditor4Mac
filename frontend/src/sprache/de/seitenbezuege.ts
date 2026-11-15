// Die Texte der **bezogenen Stände** auf der Paketseite — G04
// (#2251).
//
// Eine eigene Datei, weil `seiten.ts` an der Zeilengrenze aus §2 steht.

export const bezuege = {
  ueberschrift: 'Bezogene Stände',
  erklaerung:
    'Ein bezogenes Paket steht an einem Commit und veraltet nicht von selbst. Frag die Quellen, ob es einen neueren gibt — erst dann geht etwas ins Netz.',
  fragen: 'Quellen fragen',
  leer: 'Noch kein Paket bezogen. Beziehe eines über „Paket beziehen …“ — danach steht hier, wie weit seine Quelle ist.',
  bezogen: 'Bezogen: {stand}',
  geprueft: 'Geprüft: {stand}',
  ohnePruefung: 'Nicht in der Liste geprüfter Quellen',
  pakete: '{zahl} Pakete',
  paketEins: 'ein Paket',
  keinPaket: 'kein Paket',
  lage: {
    ungefragt: 'Noch nicht gefragt.',
    aktuell: 'Auf dem neuesten Stand der Quelle.',
    neuerEins: 'Neuer Stand {spitze} vom {datum} — einen Commit voraus.',
    neuer: 'Neuer Stand {spitze} vom {datum} — {vorsprung} Commits voraus.',
    abweichend:
      'Die Quelle ist nicht voraus — zurückgesetzt oder verzweigt (Spitze {spitze}).',
    unbekannt: 'Die Quelle ließ sich nicht fragen: {grund}',
  },
  ansehenBezeichnung: 'Aktualisierung ansehen … ({name})',
  zurueckBezeichnung: 'Zum aufgehobenen Stand zurück ({name}, {stand})',
  ansehen: 'Aktualisierung ansehen …',
  aufgehoben: 'Aufgehoben: {stand}',
  zurueck: 'Zum aufgehobenen Stand zurück',
  vorschauTitel: 'Aktualisierung von {name}',
  vonZu: '{von} → {zu}: {vorsprung} voraus, {rueckstand} zurück',
  paketzeile: '{paket}: {neu} neu, {geaendert} geändert, {entfernt} entfernt',
  keinePakete: 'Kein Vorlagenpaket ändert sich.',
  uebrige: '{zahl} weitere Dateien gehören zu keinem Paket.',
  sets: 'Diese Sets nennen ein berührtes Paket: {sets}',
  keineSets: 'Kein Set nennt ein berührtes Paket.',
  aufheben: 'Der bisherige Stand bleibt aufgehoben und lässt sich zurückholen.',
  vorigeWeicht: 'Der schon aufgehobene Stand {stand} weicht dabei.',
  gruende: {
    setVerliert: 'Ein Paket, das ein Set nennt, verliert Dateien.',
    nichtVoraus:
      'Die Quelle ist nicht voraus — die Aktualisierung wäre ein Schritt zurück oder zur Seite.',
    gekuerzt:
      'Die Quelle nennt mehr als 300 geänderte Dateien; die Liste oben ist gekürzt.',
  },
  bestaetigen: 'Ich habe die Gründe gesehen und will trotzdem aktualisieren.',
  ausfuehren: 'Auf {zu} aktualisieren',
  verwerfen: 'Nicht aktualisieren',
}
