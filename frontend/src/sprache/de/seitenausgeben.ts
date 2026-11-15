// Die Texte des **Ausgabeassistenten** — wohin, und was jedes Ziel
// braucht.
//
// Aus `seiten.ts` herausgelöst, als die Datei mit dem Assistenten an die
// Zeilengrenze aus §2 kam (UXR-09,
// #1917).

export const ausgeben = {
  titel: 'Ausgeben',
  erklaerung:
    'Fünf Wege, ein Set aus dieser Anwendung herauszubekommen — und wofür jeder gedacht ist.',

  // **Der Ausgabeassistent** (UXR-09, #1917) — die erste Frage, der
  // leere Zustand daneben und je Ziel ein Satz für die Wahl (`kurz`) und
  // einer für seine Fläche (`lang`). Bilder und Druckbogen bringen den
  // für ihre Fläche als Seite mit, und ihr Name ist der der Seite.
  wohin: 'Wohin?',
  keinZiel: 'Noch kein Ziel gewählt',
  nichtGemerkt:
    'Das Ziel gilt, gemerkt wurde es nicht — beim nächsten Start ist es nicht vorgewählt:',
  keinZielErklaerung:
    'Wähle unter „Wohin?“, in welcher Form das Set hinaus soll — dann steht hier, was dafür einzustellen ist.',
  ohneKarten:
    'Bilder und Druckbogen hätten nichts zu zeigen. Karten entstehen unter „Karten“ — angelegt, eingefügt oder eingelesen.',
  ziele: {
    setdatei: {
      name: 'Setdatei',
      kurz: 'Das Set selbst, wieder aufmachbar — eine `.mse-set`-Datei, die auch Magic Set Editor 2 liest.',
      lang: 'Die Taste dafür ist ⌘S. **Nicht darin** stehen Decks, Exemplare, Preise, Marken und die Herkunft einzelner Werte — sie bleiben im Bestand, denn MSE2 verwirft fremde Angaben beim Speichern.',
    },
    bilder: {
      kurz: 'Je Karte eine Bilddatei zum Weiterverwenden — für ein Forum, eine Tabelle, ein Spielprogramm.',
    },
    druckbogen: {
      kurz: 'Die Karten nebeneinander auf Papier, mit Schnittmarken.',
    },
    vorlage: {
      kurz: 'Text statt Bild: eine Liste, eine Tabelle, eine Datei für ein anderes Programm.',
      lang: 'Welche Vorlagen es gibt, entscheidet das bezogene Paket.',
    },
    mtgjson: {
      kurz: 'Das Kartenformat von mtgjson.com — für Werkzeuge, die es lesen.',
      lang: 'Nicht zum Weiterarbeiten in dieser Anwendung: was sich in MTG JSON nicht abbilden lässt, nennt die Ausgabe danach.',
    },
  },
  vorlagenEingrenzen: 'Vorlagen eingrenzen',
  weitereOptionen: 'Weitere Optionen ({zahl})',
  // **„Wie es wird"** — der Schritt vor dem Schreiben (UXR-F7, #2015).
  probelauf: {
    ueberschrift: 'Wie es wird',
    rechnet: 'Wird berechnet …',
    hinweis: 'Geschrieben wird erst mit dem Knopf darunter.',
    bogen: {
      jeSeite: '{zahl} je Seite · {seiten}',
      masse: '{karten} · Papier {breite} × {hoehe} mm · {dpi} dpi',
      letzte: 'Auf der letzten Seite: {karten}.',
    },
    vorlage: {
      datei: 'Datei: {name} ({groesse})',
      karten: 'Ausgewertet über {karten}.',
      beigaben: 'Dateien daneben: {zahl} (Ordner „{ordner}“, {groesse})',
      mindestens: 'mindestens {groesse}',
      befunde: 'Bei der Auswertung aufgefallen:',
    },
  },
}
