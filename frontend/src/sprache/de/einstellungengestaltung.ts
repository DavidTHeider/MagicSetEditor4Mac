// **Die Gestaltung auf der Einstellungsseite** — UX-09
// (#1681).
//
// **Eine eigene Datei**, weil `seiteneinstellungen.ts` mit ihnen über die
// Zeilengrenze aus §2 ginge — dieselbe Gestalt wie `einstellungenhandel.ts`.
// **Ein Katalog bleibt es trotzdem**: die Teile werden dort hineingestreut.

// gestaltungsnamen sind die Beschriftungen der vier Einstellungen.
export const gestaltungsnamen = {
  farbwelt: 'Farbwelt',
  schriftbereich: 'Anzeigeschrift gilt für',
  hintergrund: 'Fensterbild',
  gestaltung: 'Eigene Farben',
}

// gestaltungswerte sind die Stellungen — und die drei neuen Schriften.
export const gestaltungswerte = {
  'farbwelt.pergament': 'Pergament — warm, mit Faser',
  'farbwelt.klar': 'Klar — neutral, ohne Faser',
  'farbwelt.schiefer': 'Schiefer — kühles Blaugrau',
  'farbwelt.moos': 'Moos — gedämpftes Grün',
  'schriftbereich.ueberschriften': 'Nur die Überschriften',
  'schriftbereich.ganz': 'Die ganze Oberfläche',
  'hintergrund.aus': 'Aus',
  'hintergrund.fuellen': 'Füllen',
  'hintergrund.einpassen': 'Einpassen',
  'hintergrund.kacheln': 'Kacheln',
  'schriftart.buch': 'Buch — New York',
  'schriftart.rund': 'Rund — SF Pro Rounded',
  'schriftart.serifenlos': 'Serifenlos — die Systemschrift',
}

// gestaltungshinweise sagen, was sie bewirken.
export const gestaltungshinweise = {
  farbwelt:
    'Die Farben der Oberfläche, hell wie dunkel, jede nachgerechnet. Bei hohem Kontrast gilt die Tinte, gleich was hier steht.',
  schriftbereich:
    'Die ganze Oberfläche bekommt nur eine Textschrift; eine Schmuckschrift bleibt an den Überschriften. Die Karte behält die Schriften ihrer Vorlage.',
  hintergrund:
    'Legt dein eigenes Bild unter den Grund, mit einem Schleier, der Text lesbar hält. Das Bild wählst du darunter; **Aus** nimmt es sofort weg.',
  gestaltung:
    'Eigene Farben für die zehn Merkmale der Oberfläche, eine je Zeile. Nur Farben als Hexzahl — keine Adressen, keine Schriften. Was nicht zu lesen wäre, gilt nicht.',
}

// gestaltungsbeiwerk sind die Texte unter den Einstellungen der Gruppe.
export const gestaltungsbeiwerk = {
  fensterbild: 'Dein Fensterbild',
  vorschau: 'Vorschau des Fensterbilds',
  keinBild:
    'Noch kein eigenes Bild. Wähle eines — es erscheint dann hier und, sobald das Fensterbild nicht aus ist, hinter der Oberfläche.',
  aus: 'Das Bild liegt bereit, das Fensterbild steht aber auf **Aus**.',
  waehlen: 'Bild wählen …',
  entfernen: 'Bild entfernen',
  merkmale: 'Eigene Farben',
  merkmaleHilfe:
    'Eine Angabe je Zeile, etwa »--mse-marke: #4e2f16;«. Erlaubt sind {merkmale}.',
  uebernehmen: 'Übernehmen',
  zeilenfehler:
    'Zeile {zeile}: erwartet wird eine Angabe wie »--mse-grund: #1e1b16;« — eines der erlaubten Merkmale und eine Farbe als Hexzahl.',
  nichtLesbar: 'Nicht angewandt — so wäre es nicht zu lesen:',
  paar: '{vorn} auf {hinten}: {verhaeltnis} : 1, verlangt {schwelle} : 1',
  // **Was kein Merkmal ist, bekommt einen Namen** — die Farben der
  // Meldungen und der Grund über dem Fensterbild (`kontrast.ts`).
  paarnamen: {
    'status:warnung': 'Warnungen',
    'status:erfolg': 'Erfolgsmeldungen',
    'fensterbild:dunkel': '--mse-grund über dunklem Fensterbild',
    'fensterbild:hell': '--mse-grund über hellem Fensterbild',
  } as Record<string, string>,
  angewandt: 'Angewandt.',
  ruhtBeiKontrast:
    'Gespeichert — bei hohem Kontrast gelten eigene Farben nicht; sie kommen wieder, sobald er aus ist.',
  schmuckNurOben:
    'Die gewählte Anzeigeschrift ist eine Schmuckschrift — sie bleibt an den Überschriften.',
  zuruecksetzenGescheitert: 'Die Darstellung ließ sich nicht zurücksetzen:',
  zurVorgabe: 'Eigene Farben auf die Vorgabe zurücksetzen',
}
