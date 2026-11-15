// **Eine Datei ansehen** — UX-13 (#1685). Ein eigener Teil des Katalogs
// `bausteine`, weil der sonst über die Zeilengrenze liefe (§2, §9).
export const dateiansicht = {
  ueberschrift: 'Die Datei',
  laedt: 'Die Datei wird angesehen …',
  arten: {
    paket: 'MSE-Paket',
    bild: 'Bild',
    vektor: 'Vektorbild',
    pdf: 'PDF',
    text: 'Text',
    archiv: 'Archiv',
    unbekannt: 'Unbekannte Art',
  },
  name: 'Name',
  art: 'Art',
  groesse: 'Größe',
  geaendert: 'Geändert',
  ort: 'Ort',
  masse: '{breite} × {hoehe} Punkte',
  seiten: '{anzahl} Seiten',
  eintraege: '{anzahl} Einträge',
  gekuerzt: 'Gezeigt ist der Anfang — die Datei geht weiter.',
  unsicher:
    '{anzahl} Einträge führen aus der Datei hinaus oder sind Verweise. Sie werden gezählt und nicht betreten.',
  unsicherEintrag: 'nicht betreten',
  verschluesselt:
    'Eine Sicherung mit Kennwort. Was darin liegt, zeigt erst das Zurückspielen.',
  zuGross: 'Zu groß für eine Vorschau. Öffne sie mit ihrem Programm.',
  unlesbar: 'Der Inhalt lässt sich nicht lesen. Name, Größe und Zeit stimmen.',
  keineVorschau:
    'Für diese Art gibt es keine Vorschau: sie könnte ein Programm sein. Zeige sie im Finder, wenn du sie brauchst.',
  imFinder: 'Im Finder zeigen',
  oeffnen: 'Öffnen',
  imFinderFuer: '{name} im Finder zeigen',
  oeffnenFuer: '{name} mit seinem Programm öffnen',
  ansehen: 'Ansehen',
  ansehenFuer: '{name} ansehen',
  zuklappen: 'Ansicht schließen',
} as const
