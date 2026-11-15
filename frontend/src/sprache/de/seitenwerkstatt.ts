// **Die Werkstatt der Kartenseite** — Werkstück U5 von Phase 5.1
// (#1416):
// Arbeitsplatz, Schnellfilter, Zoom, Überlagerungen und die Problemliste.
// Ein Teil des Katalogs `karten`, in eigener Datei, weil `seitenkarten.ts`
// an der Zeilengrenze steht (§2, §9).
export const werkstatt = {
  arbeitsplatz: 'Kartenarbeitsplatz',
  inspectorBezeichnung: 'Felder und Prüfung',
  // **Marken an Karten** (#1208).
  // **Der Schnellfilter** — Werkstück U5 von Phase 5.1 (#1416), die Sprache
  // von MSE2.
  schnellfilter: {
    bezeichnung: 'Karten filtern',
    platzhalter: 'Filtern — z. B. type:creature -goblin',
    leeren: 'Filter leeren',
    hilfe:
      'Leerraum trennt Teile, „…“ hält zusammen, feld:wert bindet an ein Feld, ein führendes - verneint. Groß und Klein sind egal.',
    treffer: '{gezeigt} von {gesamt} Karten',
  },
  zoom: {
    bezeichnung: 'Größe der Vorschau',
    anpassen: 'Anpassen',
    druck: 'Druckgröße',
    ohneDruck:
      'Die Vorlage nennt keine Auflösung; eine Druckgröße gibt es dann nicht.',
  },
  // **Das Kartenbild in den Finder ziehen** — AP-12 (#2369).
  ziehen: {
    beschriftung: 'Bild ziehen',
    hinweis:
      'Ziehe das Bild in den Finder, um die Karte als PNG abzulegen. Dasselbe geht unter {weg}.',
  },
  overlays: {
    bezeichnung: 'Überlagerungen',
    bildgrenzen: 'Bildgrenzen',
    beschnitt: 'Beschnitt',
    diagnose: 'Diagnose',
    beschnittHinweis:
      'Innerhalb der Linie bleibt, was ein Schnitt von 3 mm nicht erfasst.',
  },
  probleme: {
    ueberschrift: 'Probleme',
    keine: 'Keine Probleme an dieser Karte.',
    zumFeld: 'Zum Feld {feld}',
    ohneFeld: 'ohne Feldbezug',
  },
  // **Blättern** (#1290).
  // Portiert aus MSEs Kartenmenü: *Select Previous Card* / *Select Next Card*
  // auf ⇞ und ⇟ (`data/en.mse-locale/locale`, Zeilen 48–49, Fassung f9d9356
  // vom 2021-07-06).
  blaettern: {
    vorige: 'Vorige Karte',
    naechste: 'Nächste Karte',
    stelle: '{stelle} von {gesamt}',
    kuerzelVorige: '⇞',
    kuerzelNaechste: '⇟',
  },
}
