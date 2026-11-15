// Die Texte der **Setliste** — öffnen, anlegen, einlesen.
//
// Aus `seiten.ts` herausgelöst, als die Datei über 400 Zeilen wuchs (§2). Der
// Anlass war das Anlegen eines neuen Sets
// (#891):
// die Seite führt inzwischen vier Wege in den Bestand — Datei, neues Set,
// MTG JSON und Scryfall.

export const sets = {
  // **Die Auswahl aus dem Überblick lösen** — AP-8 (#2365).
  auswahlLoesen: 'Alle Sets zeigen',
  erklaerung:
    'Jedes offene Set liegt im Bestand. Von hier aus geht es zu seinen Karten — oder zurück in eine Datei.',
  titel: 'Sets',
  leer: 'Noch kein Set geöffnet. Öffne eine .mse-set-Datei, um anzufangen.',
  setOeffnen: 'Set öffnen …',
  neuTitel: 'Ein neues Set',
  // **Die führende Handlung im Kopf der Setliste** — UXR-04 (#1912), nach
  // dem Artboard *Sets*. Sie tut, was ⌘N tut: zum Namen des neuen Sets.
  neuKopf: 'Neues Set',

  // **Der Titel des Aufklappkastens, in dem die Einfuhrwege stehen** —
  // Werkstück C von Phase 4.10
  // (#1325).
  // Offen über der Setliste schöben die Einfuhrwege das erste Set unter die
  // Falz.
  einlesenTitel: 'Ein Set von woanders einlesen',
  neuHinweis:
    'Ein leeres Set — Spiel und Stil bestimmen, wie seine Karten aussehen. Der Stil lässt sich später wechseln.',
  neuName: 'Name des Sets',
  neuStil: 'Stil des neuen Sets',
  neuAnlegen: 'Set anlegen',
  // **Der Assistent** — UX-05, S07.NAT-03 (#1677), nach dem Artboard *Sets*:
  // drei Schritte, überspringbar, und zurück geht nichts verloren.
  neuSchritte: 'Schritte zum neuen Set',
  neuSchrittArt: 'Art',
  neuSchrittAussehen: 'Aussehen',
  neuSchrittName: 'Name',
  neuArtFrage: 'Wofür sind die Karten? Den Paketnamen suchen wir für dich.',
  neuArtZusatz: '· {stile} verfügbar',
  neuEigenes: 'Etwas ganz Eigenes — ein Spielpaket anlegen',
  neuWeiter: 'Weiter',
  neuSchrittZurueck: 'Zurück',
  neuUeberspringen: 'Überspringen',
  neuWahl: '{spiel} · {stil}',
  neuOhneSpiele:
    'Es ist kein Spielpaket installiert. Ohne eines gibt es keine Kartenfelder — die Pakete kommen über die Paketseite.',
  neuOhneStile:
    'Zu {spiel} ist kein Stilpaket installiert. Ohne einen Stil lässt sich kein Set anlegen — die Pakete kommen über die Paketseite.',
  // **Der Weg dorthin, und zurück** — UX-05 S07.NAT-04: die Seite der Sets
  // bleibt stehen, und was eingegeben ist, mit ihr (UX-03.AK-02).
  neuZuDenPaketen: 'Zu den Paketen',
  neuZurueck:
    'Mit ⌘[ kommst du hierher zurück; was du eingegeben hast, steht dann noch da.',
  beispielsets: 'Zum Ausprobieren mitgeliefert',
  geoeffnet: 'geöffnet',
  schonOffen: 'Diese Datei ist schon offen — hier ist sie.',
  teilerfolg:
    'Die Datei wurde nicht zu Ende gelesen. {anzahl} Sets liegen trotzdem im Bestand — sie stehen unten in der Liste.',
  beimOeffnenBefunde:
    'Beim Öffnen nicht lesbar. Das Set ist trotzdem da; diese Stellen fehlen darin:',
  // **Die Vorschau vor dem Einlesen** — FMT-03.AK-01 (#1692). Sie erscheint
  // nur, wenn es etwas zu sagen gibt.
  vorschau: {
    ueberschrift: 'Vor dem Einlesen: {name}',
    umfang: '{karten} · Spiel {spiel} · Stil {stil}',
    umfangOhneStil: '{karten} · Spiel {spiel}',
    schonOffen:
      'Diese Datei liegt schon im Bestand — Aufschlagen zeigt das vorhandene Set, statt ein zweites anzulegen.',
    befunde: 'Diese Stellen lassen sich nicht lesen; das Set käme ohne sie:',
    bedarf:
      'Es fehlen Pakete. Einlesen geht trotzdem — die Karten zeichnen dann behelfsmäßig, bis die Pakete da sind:',
    einlesen: 'Einlesen',
    aufschlagen: 'Aufschlagen',
    verwerfen: 'Nicht einlesen',
  },
  // **Das Löschen** (#1218) — endgültig, mit Namensbestätigung.
  loeschen: 'Löschen …',
  loeschenTitel: 'Dieses Set löschen',
  loeschenWarnung:
    '{name} verschwindet aus dem Bestand, samt {karten}. Es gibt keinen Papierkorb und kein Zurück. Was in einer Datei liegt, bleibt dort — gelöscht wird der Stand im Bestand.',
  loeschenSammlung:
    'Deine Sammlung bleibt: ein Exemplar hängt an einem Druck, nicht an einem Set.',
  loeschenFeld: 'Zum Bestätigen den Namen des Sets eintippen',
  loeschenPasstNicht: 'Der Name stimmt noch nicht mit {name} überein.',
  loeschenJa: 'Endgültig löschen',
  loeschenNein: 'Abbrechen',
  loeschenBefunde: 'Das Set ist weg. Diese Reste ließen sich nicht wegräumen:',

  spalten: {
    name: 'Name',
    spiel: 'Spiel',
    karten: 'Karten',
    herkunft: 'Herkunft',
  },
  mtgjsonUeberschrift: 'MTG JSON',
  mtgjsonErklaerung:
    'Kartendefinitionen von mtgjson.com — als einzelne Datei oder als Verzeichnis. Die Zuordnung auf Kartenfelder hängt am Spiel.',
  mtgjsonSpiel: 'Spiel',
  mtgjsonDatei: 'Datei öffnen …',
  mtgjsonVerzeichnis: 'Verzeichnis öffnen …',
  mtgjsonKeinSpiel:
    'Für kein Spiel gibt es eine Zuordnung von MTG-JSON-Feldern. Ohne sie lässt sich nichts einlesen.',
  scryfallsatzUeberschrift: 'Ein Set von Scryfall',
  scryfallsatzErklaerung:
    'Das Satzkürzel bestimmt, welche Karten kommen — mit ihren Texten, ohne Bilder. Mitgelesen werden Name, Manakosten, Typen, Regeltext, Aromatext, Stärke, Widerstand, Seltenheit und Illustrator; die Bilder holt man je Karte. Den Stil wählt man danach.',
  scryfallsatzKeinSpiel:
    'Ohne ein Spiel mit Zuordnung lässt sich kein Satz einlesen.',
  scryfallsatzSpiel: 'Spiel für die Zuordnung',
  scryfallsatzCode: 'Satzkürzel bei Scryfall',
  scryfallsatzPlatzhalter: 'z. B. lea',
  scryfallsatzEinlesen: 'Satz einlesen',
  vorlageAusfuhr: 'Über eine Exportvorlage …',
  vorlageWaehlen: 'Exportvorlage',
  vorlageAusfuehren: 'Ausführen',
  vorlageSchliessen: 'Schließen',
  vorlageKeine:
    'Für dieses Spiel ist keine Exportvorlage installiert. Vorlagen kommen aus bezogenen Paketen.',
  vorlageGeschrieben: 'Geschrieben nach',
  vorlageBefunde:
    'Beim Ausführen ist etwas aufgefallen — die Datei ist trotzdem geschrieben:',
  vorlageGescheitert:
    'Diese Vorlage läuft nicht durch. Was ihr fehlt, steht darunter:',
  mtgjsonSichern: 'Als MTG JSON sichern …',
  mtgjsonGesichert: 'Karten gesichert nach',
  mtgjsonVerluste:
    'Beim Schreiben nicht abbildbar. Die Datei ist da; diese Angaben fehlen darin:',
  mtgjsonSpalte: 'Ausfuhr',
  speichern: 'Sichern',
  speichernUnter: 'Sichern unter …',
  speicherSpalte: 'Sichern',
  gespeichert: 'Gesichert nach',
  beimSpeichernBefunde:
    'Beim Sichern nicht mitgekommen. Die Setdatei ist geschrieben; diese Dateien daneben fehlen:',
  oeffnen: 'Karten dieses Sets zeigen',
  // **Das Menü an jeder Setkarte** — UXR-04 (#1912): der Name führt, das
  // Seltene steht dahinter. `handlungenKurz` ist das Zeichen auf dem Knopf,
  // sein Name steht in `handlungenAn`.
  handlungenAn: 'Handlungen — {name}',
  handlungenKurz: '…',
  vorbehalte: {
    mitZahl: 'Nicht gedeutet ({anzahl})',
    ueberschrift: 'Nicht gedeutet',
    erklaerung:
      'Diese Zeilen hat der Import nicht verstanden. Sie reisen unverändert mit und gehen beim Speichern nicht verloren.',
    keine: 'Nichts — dieses Set wurde vollständig gelesen.',
  },

  // **Der portable Umschlag** — Werkstück N4 von Phase 5.4
  // (#1457).
  umschlagAblegen: 'Als Umschlag ablegen',
  umschlagGeschnuert:
    'Abgelegt. Der Umschlag ist unterschrieben — wer ihn bekommt, kann nachsehen, dass er von hier stammt.',
  umschlagTitel: 'Einen Umschlag hereinnehmen',
  umschlagErklaerung:
    'Ein Umschlag ist ein Set, das jemand anderes abgelegt hat. Er wird erst angesehen und dann übernommen — vorher ändert sich im Bestand nichts.',
  umschlagWaehlen: 'Umschlag ansehen …',
  umschlagUrheber: 'Unterschrieben von',
  umschlagAbdruckHinweis:
    'Der Abdruck sagt, WER unterschrieben hat — nicht, ob man ihm trauen sollte. Wer die Person kennt, vergleicht ihn mit ihr.',
  umschlagInhalt: 'Darin liegt',
  umschlagUebernehmen: 'Übernehmen',
  umschlagVerwerfen: 'Doch nicht',
  umschlagUebernommen:
    'Übernommen. Das Set ist neu angelegt — es hat eigene Kennungen und überschreibt nichts.',
}
