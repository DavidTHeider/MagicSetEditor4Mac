// **Die Wörter der Partienliste** — herausgelöst für PLAY-03.AK-05
// (#1711).
//
// **Ein Katalog, zwei Dateien** (§9): `seitenspielen.ts` stand mit 399 Zeilen
// einen Schritt vor der Grenze aus §2. Der Schnitt folgt der Frage — dort der
// Tisch, hier die Liste davor.
export const partienwoerter = {
  // **Der Weg zurück in die eigene Partie** (#1600).
  partienTitel: 'Deine Partien',
  partienErklaerung:
    'Eine Partie bleibt im Bestand, auch wenn du das Fenster schließt oder das Profil wechselst.',
  partienLeer:
    'Du hast keine Partie. Fang eine an — oder setze dich an einen fremden Tisch.',
  partienEinsteigenAls: 'Weiterspielen als {name}',
  partienObjekte: 'Karten im Spiel',
  partienZug: 'Zug',

  // **Die Marke an einer Partie, auf die man wartet** — UXR-F16
  // (#2031):
  // so kurz wie im Artboard *Spielen*. Die anderen Sprachen nehmen die Wörter
  // der Handelsliste („Du bist dran“, `auswahlDran`) — dasselbe Wort für
  // dieselbe Auskunft.
  partienDran: 'dran',

  // **Suchen, Ordnen und Filtern** — PLAY-03.AK-05 (#1711).
  partienSuche: 'Partie suchen',
  partienSuchePlatzhalter: 'Name oder Spiel',
  partienSpiel: 'Spiel',
  partienStand: 'Stand',
  partienStandAlle: 'Jeder Stand',
  partienOrdnung: 'Ordnung',
  partienOrdnungZuletzt: 'Zuletzt geändert',
  partienOrdnungName: 'Name',
  partienOrdnungZug: 'Zug',

  // **Die Zahl steht nur, wo es etwas zu zählen gibt**: die Leiste erscheint
  // ab zwei Partien, und darunter wäre „1 Partien“ ohnehin kein Deutsch.
  partienZahl: '{zahl} Partien',
  partienZahlGefiltert: '{gezeigt} von {alle} Partien',
  partienNichtsGefunden:
    'Keine deiner Partien passt dazu. Ändere die Suche oder setze sie zurück.',
  partienZuruecksetzen: 'Auswahl zurücksetzen',

  // **Die Zeile der Liste** — UXR-13
  // (#1921):
  // zwei Menüs statt eines Knopfs je Sitz und dreier daneben. Der Name eines
  // Menüs nennt die Partie, die Einträge heißen, was sie tun (WCAG 2.4.6).
  partienWeiterspielen: 'Weiterspielen',
  partienWeiterspielenIn: 'Weiterspielen: {name}',
  partienZuschauen: 'Zuschauen (ohne Sitz)',
  partienHandlungen: 'Mit dieser Partie: {name}',
  partienHandlungenKurz: '…',

  // **Die Verwaltung einer Partie als Ganzem** — PLAY-03.AK-05 (#1711).
  partienDuplizieren: 'Duplizieren',
  partienAusfuhr: 'Ausführen',
  partienLoeschen: 'Löschen',
  partienEinlesen: 'Partie einlesen',
  partienEinlesenErklaerung:
    'Eine ausgeführte Datei wird zu einer neuen Partie. Sie bekommt eine neue Kennung und fängt bei Fassung 1 an; die Partie, aus der sie stammt, wird damit nicht fortgesetzt. In der Datei steht der volle Zustand — auch die Hände der anderen.',
  partienDupliziert: 'Dupliziert. Die neue Partie steht in der Liste.',
  partienAbgelegt: 'Abgelegt nach {pfad}.',
  partienEingelesen: 'Eingelesen. Die neue Partie steht in der Liste.',
  partienGeloescht: 'Gelöscht.',
}
