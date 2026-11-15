// Die Texte des **Paketinspectors und des Plans vor einer Paketänderung** —
// DOM-01 (#1696).
//
// Eine eigene Datei, weil `seiten.ts` an der Zeilengrenze aus §2 steht. Der
// Schnitt folgt der Frage: dort die Paketseite, hier, was sie über **ein**
// Paket sagt und was eine Änderung an ihm bewirkt.

export const paketplan = {
  // **Die drei Regionen des Arbeitsplatzes** — das Artboard `Pakete`.
  arbeitsplatz: 'Paketarbeitsplatz',
  navigator: 'Suche und Filter',
  inspector: 'Paket',
  zentrum: 'Pakete',

  // **Die Filter im Navigator** — das Artboard `Pakete` zeichnet die
  // Herkunft als Knöpfe; die vierte Art, übernommen, kommt aus dem Bestand.
  herkunft: 'Herkunft',
  herkunftAlle: 'Jede Herkunft',
  herkunftsarten: {
    mitgeliefert: 'mitgeliefert',
    bezogen: 'bezogen',
    uebernommen: 'übernommen',
    eigen: 'eigene',
  },

  spalteStand: 'Stand',
  standEingebunden: 'eingebunden',
  standGeaendert: 'lokal geändert',

  inspectorLeer:
    'Wähle ein Paket in der Liste — hier stehen dann Herkunft, Fassung, Abhängigkeiten und was eine Änderung bewirkt.',
  kopf: '{art} · Fassung {version}',
  kopfOhneFassung: '{art} · ohne Fassung',
  eingebunden: 'Eingebunden',
  keinSet: 'Kein Set dieses Profils benutzt es.',
  einSet: 'Ein Set benutzt es.',
  vieleSets: '{anzahl} Sets benutzen es.',
  setsUnbekannt: 'Welche Sets es brauchen, steht in keinem Setkopf.',
  geaendert:
    'Seit dem Einbinden lokal geändert. Ein Ersetzen fragt vorher; die Änderung bleibt als vorige Fassung erhalten.',
  woher: 'Woher',
  woherMitgeliefert: 'Mitgeliefert mit der Anwendung.',
  woherBezogen: 'Bezogen von {herkunft}.',
  woherUebernommen: 'Übernommen aus {herkunft}.',
  woherEigen: 'Hier angelegt oder von Hand ins Paketverzeichnis gelegt.',
  pruefsumme: 'Prüfsumme',
  // **Lizenz: nicht angegeben** — ein MSE-Paket hat kein Feld dafür
  // (`src/data/installer.hpp`, Fassung f9d9356 vom 2021-07-06). Was dort
  // nicht steht, wird hier nicht erfunden.
  lizenz: 'Lizenz',
  lizenzUnbekannt:
    'nicht angegeben — ein MSE-Paket trägt kein Lizenzfeld; die Inhalte gehören ihren Rechteinhabern',
  ort: 'Ort',
  haengtAb: 'Hängt ab von',
  gebrauchtVon: 'Gebraucht von',
  nichts: 'nichts',

  einbinden: 'Paket einbinden …',
  // **Der führende Knopf im Kopf** — UXR-17 (#1925), nach dem Artboard
  // `Pakete`: er öffnet die Schublade mit Bezug, Übernahme und Umschlag.
  beziehen: 'Paket beziehen …',
  beziehenTitel: 'Paket beziehen',
  loeschen: 'Löschen …',
  zurueck: 'Vorige Fassung zurückholen …',
  weitergeben: 'Als Installer weitergeben …',
  weitergegeben: 'Als Installer gesichert: {datei}',

  // **Der Plan** — DOM-01.AK-02: vor jeder Änderung, was sie bewirkt.
  titel: {
    loeschen: '{paket} löschen',
    einbinden: '{paket} einbinden',
    ersetzen: '{paket} ersetzen',
    zurueck: '{paket} auf die vorige Fassung zurücksetzen',
  },
  fassung: 'Fassung {von} → {zu}',
  ohneFassung: 'ohne Fassung',
  sets: 'Betroffene Sets',
  keineSets: 'Kein Set dieses Profils ist betroffen.',
  konflikte: 'Danach nicht mehr erfüllt',
  keineKonflikte: 'Keine Abhängigkeit bricht.',
  planGeaendert: 'Das Paket ist seit dem Einbinden lokal geändert.',
  ueberschreiben:
    'Die lokale Änderung ersetzen — sie bleibt als vorige Fassung erhalten',
  nichtMoeglich: {
    loeschen:
      'Dieses Paket liegt nicht im verwalteten Verzeichnis. Gelöscht wird hier nur, was diese Anwendung eingebunden hat.',
    einbinden: 'Dieses Paket lässt sich hier nicht einbinden.',
    ersetzen: 'Dieses Paket lässt sich hier nicht ersetzen.',
    zurueck:
      'Es gibt keine vorige Fassung. Eine entsteht, wenn du eine neue Fassung über dieses Paket einbindest.',
  },
  bestaetigen: {
    loeschen: 'Löschen',
    einbinden: 'Einbinden',
    ersetzen: 'Ersetzen',
    zurueck: 'Zurücksetzen',
  },
  verwerfen: 'Nichts ändern',
  // **Was einem Set fehlt** — DOM-01.AK-04: in den Setangaben, mit zwei
  // Wegen, es zu beheben.
  bedarf: {
    ueberschrift: 'Pakete, die diesem Set fehlen',
    erklaerung:
      'Ohne sie zeichnet die Vorschau einen Behelfsplan oder lässt aus, was das Paket beiträgt. Binde das Paket aus einem Verzeichnis ein — oder suche es bei den Paketen, um es zu beziehen oder aus MSE2 zu übernehmen.',
    suchen: 'Bei den Paketen suchen',
  },
  quelle: 'Aus {quelle}',
  // **Keine stillen Rechte** — DOM-01.AK-04: was ein Paket darf, steht vor
  // dem Einbinden da.
  keineRechte:
    'Ein Paket bringt Vorlagen und MSE-Script mit. Das Skript hat keinen Zugriff auf Dateien, Netz oder Programme, und was das Paket braucht, wird nicht mitgeholt — es steht oben als Konflikt.',
}
