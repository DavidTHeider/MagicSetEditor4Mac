// Die Texte der **Symbolbibliothek** — DOM-02
// (#1697).
//
// Eine eigene Datei, weil `seitenkarten.ts` an der Zeilengrenze aus §2 steht.
// Der Schnitt folgt der Frage: dort das Symbol **eines** Sets, hier die, die
// sich in jedes einsetzen lassen.

export const symbolbibliothek = {
  ueberschrift: 'Symbolbibliothek',
  erklaerung:
    'Symbole dieses Profils, die sich in jedes Set einsetzen lassen. Eingesetzt wird eine Kopie: das Set trägt sie beim Sichern mit, auch wenn das Symbol hier später gelöscht wird.',

  name: 'Name für das Setsymbol',
  aufnehmen: 'Setsymbol aufnehmen',
  einlesen: 'Symboldatei einlesen …',
  suche: 'Symbole suchen',

  leer: 'Noch keine Symbole. Nimm das Symbol dieses Sets auf oder lies eine .mse-symbol-Datei ein — danach lässt es sich in jedes Set einsetzen.',
  keinTreffer:
    'Kein Symbol passt zu Suche und Markenfilter. Ändere eines von beiden.',
  liste: 'Symbole der Bibliothek',
  vorschau: 'Vorschau von {name}',
  verwendet: 'Getragen von {sets}',
  unverwendet: 'Kein Set trägt es.',

  einsetzen: 'In dieses Set einsetzen',
  eingesetzt: '{name} ist jetzt das Symbol dieses Sets.',
  duplizieren: 'Duplizieren',
  kopieName: '{name} (Kopie)',
  umbenennen: 'Umbenennen',
  neuerName: 'Neuer Name für {name}',
  namenUebernehmen: 'Namen übernehmen',
  umbenennenAbbrechen: 'Abbrechen',
  ausgeben: 'Ausgeben …',
  ausgegeben: 'Ausgegeben nach {pfad}',
  eingelesen: '{name} steht jetzt in der Bibliothek.',
  aufgenommen: '{name} steht jetzt in der Bibliothek.',

  loeschen: 'Löschen',
  loeschenFrage: '{name} aus der Bibliothek löschen?',
  loeschenFolgen:
    'Diese Sets tragen es und behalten ihre Kopie: {sets}. Verloren geht nur der Eintrag der Bibliothek — und mit ihm der Nachweis, dass sie dasselbe Symbol tragen.',
  loeschenOhneFolgen:
    'Kein Set trägt es. Das Löschen lässt sich nicht zurücknehmen.',
  loeschenJa: 'Löschen',
  loeschenNein: 'Abbrechen',

  schriften: 'Symbolschriften',
  schriftenErklaerung:
    'Symbolschriften sind Pakete dieses Rechners — Manasymbole und ähnliche Zeichen, die ein Stil in Textfelder setzt. Sie werden unter „Pakete“ eingebunden und stehen hier nur zur Ansicht.',
  keineSchriften:
    'Keine Symbolschrift eingebunden. Unter „Pakete“ lässt sich eine beziehen.',
  schriftMitFassung: '{name} · Fassung {fassung}',
  schriftenListe: 'Eingebundene Symbolschriften',
  // **Der Knopf über der eingeklappten Liste** — UXR-07 (#1915).
  schriftenZeigen: 'Eingebundene Symbolschriften ({zahl})',

  // **Die Bearbeitung** — DOM-02.AK-02.
  bearbeiten: 'Zeichnung bearbeiten',
  bearbeitungsueberschrift: '{name} aus der Bibliothek',
  bearbeitungserklaerung:
    'Du bearbeitest ein Symbol der Bibliothek. Gesichert wird eine neue Fassung; die bisherige bleibt als vorige erhalten. Die Sets behalten ihre Kopie, bis du sie übernimmst.',
  zumSetsymbol: 'Zurück zum Setsymbol',
  inBibliothekSichern: 'In der Bibliothek sichern',
  inBibliothekGesichert: 'Neue Fassung von {name} gesichert.',
  veraltet: 'Diese Sets tragen noch die vorige Fassung: {sets}.',
  keineVeraltet: 'Kein Set trägt die vorige Fassung.',
  nachziehen: 'In diese Sets übernehmen',
  nachgezogen: 'Übernommen in {sets}. Ihre Karten werden neu gezeichnet.',
  zuruecknehmen: 'Vorige Fassung zurückholen',
  zurueckgeholt: 'Die vorige Fassung von {name} ist zurück.',

  // **Die Fassungen** — die Größen- und Farbvorschau.
  fassungen: 'In den Fassungen dieses Sets',
  fassungenZu: 'Fassungen schließen',
  fassungsliste: 'Fassungen von {name}',
  keineFassungen:
    'Der Stil dieses Sets nennt keine Fassungen für das Setsymbol. Wähle einen Stil mit Fassungen, um das Symbol in ihnen zu sehen.',
  fassungsbild: '{name} in {fassung}, {groesse} Punkte',
}
