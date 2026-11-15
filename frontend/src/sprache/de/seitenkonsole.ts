// **Die Skriptkonsole** — der Katalog `seiten.karten.konsole`, in eigener
// Datei, weil `seitenkarten.ts` an der Zeilengrenze stand (§2, Werkstück U13
// von Phase 5.1, #1424). Ein Katalog, mehrere Dateien.
export const konsole = {
  ueberschrift: 'Skriptkonsole',
  erklaerung:
    'Ein MSE-Script-Ausdruck im Kontext der aufgeschlagenen Karte. In der Umgebung stehen dieselben Namen wie beim Zeichnen: »card«, »set«, »stylesheet« und die Funktionen des Spielpakets. Sie ändert nichts.',
  ausdruck: 'Ausdruck',
  platzhalter: 'card.name',
  auswerten: 'Auswerten',
  ergebnis: 'Ergebnis',
  art: 'Art: {art}',
  keineKarte: 'Dafür muss eine Karte aufgeschlagen sein.',
  // **Und der Weg dorthin steht daneben** (Werkstück C von Phase 4.10).
  // Der Satz darüber nannte, was fehlt, und ließ offen, was zu tun ist —
  // §6 verlangt das Zweite.
  zuDenKarten: 'Zu den Karten',
  vorbehalte: 'Vorbehalte beim Aufbau der Umgebung',
  stelle: '(Zeile {zeile}, Spalte {spalte})',
  namen: 'Verfügbare Namen',
  namenErklaerung:
    'Was ein Ausdruck an dieser Karte nennen darf — die Bibliothek und was die Vorlage dazulegt. Ein Klick setzt den Namen ein.',
  beispiele: 'Beispiele',
  beispieleImmer: 'Laufen überall',
  beispieleImmerErklaerung:
    'Ein Klick setzt den Ausdruck ein. Diese hier brauchen nichts als die Sprache selbst.',
  beispieleVorlage: 'Brauchen eine Vorlage',
  beispieleVorlageErklaerung:
    '»card« und »set« legt das Stilpaket in die Umgebung. Ist für diese Karte keines eingebunden, sagt die Antwort es — die Beispiele darüber laufen trotzdem.',
  beispielListe: {
    rechnung: 'Rechnen',
    grossSchreiben: 'Eine Zeichenkette in Großbuchstaben',
    laenge: 'Wie viele Zeichen etwas hat',
    ausschnitt: 'Ein Stück aus einer Zeichenkette',
    bedingung: 'Eine Bedingung',
    name: 'Der Name dieser Karte',
    grossName: 'Ihr Name in Großbuchstaben',
    regeltext: 'Die Länge ihres Regeltextes',
    setangabe: 'Der Titel des Sets',
  },
  vervollstaendigung: 'Vorschläge',
  uebernehmen: 'Übernehmen',
  namenFiltern: 'Namen eingrenzen',
  namenGezeigt: 'Es stehen {gezeigt} von {ganz} da — tippe, um einzugrenzen.',
  namenLeer: 'Darauf passt kein Name.',

  // **Abbruch, Dauer, Verlauf und Baum** — Werkstück U13 von Phase 5.1
  // (#1424).
  kuerzel: '⌘↩ wertet aus.',
  abbrechen: 'Abbrechen',
  laeuft: 'Läuft …',
  dauer: '{ms} ms',
  abgebrochen:
    'Abgebrochen — der Lauf endete, bevor ein Ergebnis da war. Der Ausdruck steht noch im Feld.',
  baum: 'Ergebnis als Baum',
  verlauf: 'Verlauf dieser Sitzung',
  verlaufErklaerung:
    'Die letzten zwanzig Ausdrücke, nur in dieser Sitzung. Ein Klick lädt den Ausdruck zurück; was nach einem Token oder Kennwort klingt, steht hier redigiert und lässt sich nicht zurückladen.',
  verlaufLeer: 'Noch nichts ausgewertet.',
  verlaufLoeschen: 'Verlauf löschen',
  ausgangGelungen: 'gelungen',
  ausgangGescheitert: 'gescheitert',
  ausgangAbgebrochen: 'abgebrochen',
}
