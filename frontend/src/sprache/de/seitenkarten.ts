// Die Texte der **Kartenseite** — die größte Seite der Anwendung.
//
// Aus `seiten.ts` herausgelöst, als die Datei über 400 Zeilen wuchs (§2).
// Der Schnitt folgt der Frage: dort stehen die Texte **aller** Seiten, hier
// die der einen, die am meisten trägt — und weiter wachsen wird
// (#674).

import { verlaufDe } from './seitenverlauf'
import { massenvorschauDe } from './seitenmassen'
import { autoersetzung } from './seitenautoersetzung'
import { pack } from './seitenpack'
import { schluesselwoerter } from './seitenschluesselwoerter'
import { werkstatt } from './seitenwerkstatt'
import { konsole } from './seitenkonsole'
import { symbolbibliothek } from './seitensymbolbibliothek'

export const karten = {
  ...werkstatt,
  autoersetzung,
  schluesselwoerter,
  symbolbibliothek,
  anlegen: 'Karte anlegen',
  nachOben: 'Nach oben',
  nachUnten: 'Nach unten',
  // **Duplizieren, nicht »kopieren«** — Werkstück 1 von Phase 3.7
  // (#1168).
  //
  // Seit die Zwischenablage daneben steht, gibt es zwei Knöpfe mit »kopieren«
  // im Namen, und sie tun Verschiedenes: dieser hängt eine zweite Karte ans
  // **selbe** Set, jener legt sie auf das Klemmbrett. Der englische Katalog
  // sagt seit jeher `Duplicate card`; das deutsche Wort dafür ist das, was
  // der Finder benutzt.
  duplizieren: 'Karte duplizieren',
  loeschen: 'Karte löschen',
  loeschenFrage: 'Diese Karte löschen? Mit ⌘Z lässt es sich zurücknehmen.',
  loeschenJa: 'Löschen',
  loeschenNein: 'Abbrechen',
  loeschenTrotzdem: 'Trotzdem löschen',
  symbole: 'Symbole einfügen',
  hergeleitet: 'Hergeleitet',
  hergeleitetWert: 'Hergeleitet: {wert}',
  bildWaehlen: 'Bild wählen',
  stilfelder: 'Felder des Stils',
  stilfelderHilfe:
    'Diese Felder bringt der Stil der Karte mit. Ihr Wert gilt nur für diesen Stil.',
  symbolZeichnen: 'Symbol zeichnen',
  kartenskripteTitel: 'Karten aus einem Skript',
  kartenskripteErklaerung:
    'Fertige Züge aus dem Spielpaket — ein Klick legt mehrere Karten auf einmal an. ' +
    'Ein Zurücknehmen holt sie zusammen zurück.',
  kartenskriptEine: 'Eine Karte angelegt.',
  kartenskriptMehrere: '{anzahl} Karten angelegt.',
  bildEinsetzen: 'Bild einsetzen',
  pack,
  symbol: {
    ueberschrift: 'Symbol',
    erklaerung:
      'Das Symbol dieses Sets — gezeichnet, nicht geladen. Es steht sofort in der Vorschau und kommt beim Sichern in die Datei. Jede Form lässt sich mit der Tastatur setzen und verschieben.',

    grundform: 'Grundform hinzufügen',

    verlauf: 'Verlauf der Zeichnung',
    nachzeichnen: 'Bild nachzeichnen',
    groesser: 'Größer',
    nachzeichnenHilfe:
      'Ein PNG, JPEG oder GIF mit einer klaren Silhouette wird zu Formen mit Bézierpunkten, die sich hier weiterbearbeiten lassen.',
    zuruecknehmen: 'Handgriff zurücknehmen',
    wiederherstellen: 'Handgriff wiederherstellen',
    kreis: 'Kreis',
    rechteck: 'Rechteck',
    dreieck: 'Dreieck',
    stern: 'Stern',

    teile: 'Teile',
    keineTeile:
      'Noch nichts gezeichnet. Eine Grundform ist der Anfang — sie lässt sich danach an jedem Punkt verändern.',
    teilLoeschen: 'Teil löschen',
    teilMitPunkten: '{name} — {zahl} Punkte',
    teilMitKopien: '{name} — Symmetrie, {zahl} Kopien',
    gruppieren: 'Gruppieren',
    gruppierenHilfe:
      'Zwei oder mehr gewählte Teile werden eine Gruppe. Mit gedrückter ⌘-Taste wählst du mehrere.',
    aufloesen: 'Gruppe auflösen',
    nachOben: 'Nach oben',
    nachUnten: 'Nach unten',
    nachObenFuer: '{name} nach oben',
    nachUntenFuer: '{name} nach unten',
    teilLoeschenFuer: '{name} löschen',
    teilMitKindern: '{name} — {zahl} Teile',
    teilKind: '{name}, in einer Gruppe',
    wahlHilfe:
      'Mit gedrückter ⌘-Taste wählst du mehrere Teile. Die Reihenfolge ist die Zeichenreihenfolge: was oben steht, liegt im Bild oben.',

    zeichnung: 'Zeichenfläche',
    punkt: 'Punkt {nummer} von {name}',
    griffDavor: 'Griff vor Punkt {nummer}',
    griffDanach: 'Griff nach Punkt {nummer}',
    punktHinzufuegen: 'Punkt einfügen',
    punktLoeschen: 'Punkt löschen',

    tastatur: 'Mit der Tastatur',
    tastaturhilfe:
      'Tabulator wechselt zwischen den Punkten. Pfeiltasten verschieben um ein Hundertstel, mit Umschalt um ein Zehntel. Eingabetaste fügt hinter dem gewählten Punkt einen neuen ein, Rückschritt löscht ihn.',

    kombination: 'Verrechnung',
    ueberdecken: 'Überdecken',
    vereinen: 'Vereinen',
    abziehen: 'Abziehen',
    schneiden: 'Schneiden',
    unterschied: 'Unterschied',
    umranden: 'Umranden',

    symmetrie: 'Symmetrie',
    symmetrieAnlegen: 'Symmetrie anlegen',
    symmetrieAufloesen: 'Symmetrie auflösen',
    symmetrieErklaerung:
      'Eine Symmetrie vervielfältigt, was in ihr liegt. Die Kopien stehen blass auf der Zeichenfläche; verändert wird immer nur das Original.',
    // **Geduzt wie die ganze Anwendung** (Werkstück C von Phase 4.10);
    // `ansprache.test.ts` hält es.
    symmetrieHinweis: 'Wähle ein Teil, um eine Symmetrie daraus zu machen.',
    symmetrieart: 'Art',
    drehung: 'Drehung',
    achse: 'Spiegelung',
    kopien: 'Kopien',
    kopienHinweis:
      'Eine Spiegelung braucht Paare — eine ungerade Zahl lässt eine Kopie allein, und MSE rundet dann ab.',
    mitteX: 'Mitte waagerecht',
    mitteY: 'Mitte senkrecht',
    griffX: 'Griff waagerecht',
    griffY: 'Griff senkrecht',
    mitte: 'Mitte der Symmetrie',
    griff: 'Griff der Symmetrie',

    sichern: 'Ins Set legen',
    gesichert:
      'Als {name} abgelegt. Beim Sichern des Sets kommt es in die Datei.',
  },
  konsole,

  wortlisten: {
    ueberschrift: 'Auswahl',
    liste: 'Auswahl aus {liste}',
    keine: 'nichts davon',
    gruppe: '{name} (Gruppe)',
  },
  regeln: {
    ueberschrift: 'Regelhinweise von Scryfall',
    holen: 'Hinweise holen',
    keine: 'Zu dieser Karte gibt es keine Hinweise.',
    offiziell: 'offiziell',
    anmerkung: 'Anmerkung von Scryfall',
    anmerkungErklaerung:
      'Anmerkungen stammen vom Scryfall-Team und erklären oft ein inoffizielles Format. Sie sind keine Regel.',
    quelle: 'Diese Karte bei Scryfall ansehen',
    alter: 'Stand: {stunden} Stunden alt',
    manaPruefen: 'Manakosten prüfen',
    manaGut: 'Der Manakosten ist in Ordnung.',
    manaSchlecht: 'Unbekannt: {symbole}',
    manaLeer: 'Diese Karte hat keinen Manakosten.',
  },
  abgleich: {
    ueberschrift: 'Kartendaten von Scryfall',
    erklaerung:
      'Zeigt, was sich ändern würde. Geschrieben wird erst, was du anhakst.',
    holen: 'Vorschlag holen',
    uebernehmen: 'Angehakte übernehmen',
    nichtsZuTun: 'Nichts zu ändern — die Karte stimmt mit Scryfall überein.',
    alt: 'bisher',
    neu: 'neu',
    leer: '(leer)',
    alter: 'Stand: {stunden} Stunden alt',
    frisch: 'Stand: gerade geholt',
    quelle: 'Diese Karte bei Scryfall ansehen',
  },
  scryfall: {
    ueberschrift: 'Bild von Scryfall',
    erklaerung:
      'Der Kartenname bestimmt das Bild. Der Künstler wird mit in die Karte geschrieben — Scryfall verlangt, dass er neben dem Bild steht.',
    name: 'Kartenname bei Scryfall',
    art: 'Bildart',
    artKlein: 'Klein (Vorschau)',
    artNormal: 'Ganze Karte',
    artAusschnitt: 'Nur die Illustration',
    holen: 'Bild holen',
    ausschnittHinweis:
      'Der Ausschnitt zeigt nur die Illustration. Er wird abgewiesen, wenn Scryfall den Künstler nicht nennt.',
    quelle: 'Diese Karte bei Scryfall ansehen',
  },
  scryfallkarte: {
    name: 'Kartenname bei Scryfall',
    platzhalter: 'z. B. Lightning Bolt',
    mitBild: 'mit Illustration',
    anlegen: 'Aus Scryfall anlegen',
  },
  zuschnitt: {
    ueberschrift: 'Ausschnitt wählen',
    erklaerung:
      'Der Rahmen sagt, welcher Teil des Bildes auf die Karte kommt. Ziehen verschiebt ihn, die Ecken ändern seine Größe.',
    quelle: 'Ursprungsbild mit dem gewählten Ausschnitt',
    vorschau: 'Vorschau in Feldgröße',
    links: 'Links',
    oben: 'Oben',
    breite: 'Breite',
    hoehe: 'Höhe',
    seitenverhaeltnis: 'Seitenverhältnis des Feldes wahren',
    uebernehmen: 'Einsetzen',
    abbrechen: 'Abbrechen',
    masse: '{breite} × {hoehe} Bildpunkte',
  },
  feldWaehlen: 'Dieses Feld bearbeiten',
  leerTitel: 'Noch keine Karte',
  listeBezeichnung: 'Karten des Sets',
  mehr: 'Mehr',
  // **Die zwei Gruppen der Zeile über dem Arbeitsplatz** — P03 (#2174):
  // was eine Karte entstehen lässt, und wie der Arbeitsplatz steht.
  gruppeAnlage: 'Karte',
  gruppeAnsicht: 'Ansicht',
  ausScryfallAufklapp: 'Aus Scryfall anlegen …',
  ausScryfallErklaerung:
    'Statt eine leere Karte anzulegen: Namen eintippen oder bei Scryfall suchen — die Felder kommen dann von dort.',
  dieseKarte: 'Diese Karte',
  // **Der Titel des Menüs in der Leiste** — GESTE (#2358), kurz wie macOS.
  kartenmenue: 'Karte',
  auswahlLoesen: 'Alle Karten zeigen',
  // **Der Bezug zu einem Druck** — seit UXR-06 (#1914) ein Eintrag im Menü
  // „Diese Karte", der einen Kasten aufschlägt; vorher mit der Ablage in
  // einem Aufklapper daneben.
  bezugAufklapp: 'Bezug zu einem Druck …',
  dieseKarteLeer: 'Schlag eine Karte auf, um sie zu bearbeiten.',
  keineKarteTitel: 'Keine Karte gewählt',
  titel: 'Karten',
  leer: 'Lege eine an oder füge kopierte Karten ein.',
  ohneName: 'Ohne Namen',
  ohneFelder: 'Diese Karte hat keinen einzigen Feldwert.',
  nummer: 'Nummer',
  vorschau: 'Vorschau der Karte',
  keineKarte:
    'Wähle eine aus der Liste — ihre Felder und ihre Vorschau stehen dann hier.',
  bearbeiten: 'Bearbeiten',

  notizen: {
    ueberschrift: 'Notizen zur Karte',
    platzhalter: 'Nur für dich — steht auf keiner Karte.',
  },

  // **Was das Spielpaket nicht erklärt** (#1271).
  buchfuehrung: 'Was die Anwendung dazu weiß',

  // **Die drei Tiefen des Kartenformulars** — UX-05.AK-01
  // (#1677).
  //
  // **Sie liegen ineinander**: wer tiefer geht, sieht mehr — nie anderes.
  // Deshalb heißen sie nach dem, was sie **dazunehmen**, und nicht nach dem,
  // was sie verbergen.
  tiefe: 'Wie viel du siehst',
  tiefeNormal: 'Normal',
  tiefeErweitert: 'Erweitert',
  tiefeExperte: 'Experte',

  tiefeErweitertTitel: 'Was ein Skript füllt',
  tiefeErweitertHinweis:
    'Diese Felder rechnet das Spielpaket aus — sie sagen, warum etwas auf der Karte steht. Ändern lassen sie sich nicht.',

  // **Der Befund unter einer verdeckten Tiefe** — AK-02 verlangt die
  // „direkte Öffnungsaktion“: ohne sie stünde eine Karte da, die sich nicht
  // speichern lässt, und niemand sähe, woran es liegt.
  befundVerdeckt: 'Ein Befund liegt unter „{tiefe}“.',
  befundSpringen: 'Dorthin springen',
  buchfuehrungHinweis:
    'Diese Angaben gehören nicht zur Karte, sondern zu ihrer Verwaltung — welche Vorlage sie benutzt und was daran eingestellt ist. Ändere sie nur, wenn du weißt, warum.',

  // **Notiz und Marken zusammen, eingeklappt** (#1271). Sie standen unter
  // der Liste und nahmen ihr die Höhe: gemessen blieben vier Zeilen.
  arbeitsstand: 'Arbeitsstand dieser Karte',
  arbeitsstandHinweis:
    'Notiz und Marken — sie gehören zu deiner Arbeit an der Karte, nicht zu ihr. Eine weitergegebene Datei trägt sie nicht mit.',

  marken: {
    ueberschrift: 'Marken',
    erklaerung:
      'Wörter für deine Arbeit — sie stehen neben der Karte und nicht in ihr. Eine weitergegebene Datei trägt sie nicht mit.',
    platzhalter: 'z. B. fertig',
    setzen: 'Marke setzen',
    entfernen: 'Marke {marke} abnehmen',
    keine: 'Noch keine Marke an dieser Karte.',
    schonGesetzt:
      'Die Karte trägt »{marke}« inzwischen schon — gesetzt anderswo.',
    filter: 'Nach Marke filtern',
    alle: 'Alle',
  },

  scryfallsuche: {
    frage: 'Bei Scryfall suchen',
    platzhalter: 'z. B. Lightning Bolt — oder t:creature c:r cmc<=3',
    erklaerung:
      'Ein Kartenname genügt. Wer Scryfalls Suchsprache kennt, kann sie hier benutzen.',
    suchen: 'Suchen',
    nichtsGefunden: 'Nichts gefunden. Ein anderes Wort probieren?',
    mehr: 'Scryfall hat mehr Treffer. Die Frage enger fassen — nachgeladen wird nicht.',
    alter: 'Stand: {stunden} Stunden alt',
    frisch: 'Stand: gerade geholt',
    befund: 'Befund: {text}',
  },

  // **Der Verweis auf einen echten Druck** (#1167).
  verweis: {
    aufschrift: 'Bezogen auf',
    keiner:
      'Diese Karte ist auf keinen Druck bezogen. Der Abgleich mit Scryfall nimmt dann irgendeinen — such oben eine Karte und beziehe sie.',
    druck: 'aus {setcode} {nummer}',
    gesetzt: '— bezogen am {wann}',
    setzen: 'Auf {name} ({setcode} {nummer}) beziehen',
    loesen: 'Bezug lösen',
  },
  // **Die Zwischenablage** (#1168).
  ablage: {
    kopieren: 'Kopieren ({zahl})',
    ausschneiden: 'Ausschneiden ({zahl})',
    einfuegen: 'Einfügen',
  },

  // **Die Spalten der Kartenliste** (#1168).
  listenspalten: {
    ueberschrift: 'Spalten der Liste',
    jeSpiel: 'Die Wahl gilt für alle Sets desselben Spiels.',
    filtern: 'Spalte suchen',
    gezeigt: '{gezeigt} von {alle} Spalten',
    ausgeblendet:
      '{anzahl} weitere sind ausgeblendet — such sie über das Feld darüber.',
    genauer: 'Es passen mehr — such genauer, um sie zu sehen.',
    nachOben: 'weiter nach vorn',
    nachUnten: 'weiter nach hinten',
    sortieren: 'Sortieren nach',
    garNicht: 'Reihenfolge des Sets',
    aufsteigend: 'Aufsteigend',
    absteigend: 'Absteigend',
    ohneVorlage:
      'Das Spielpaket ist nicht zu lesen. Die Spalten kommen aus den Karten selbst — ohne Aufschriften und ohne die Reihenfolge des Pakets.',
    spalteAusblenden: 'Spalte ausblenden',
  },

  verlauf: verlaufDe,
  massen: {
    ...massenvorschauDe,
    keinFeld: '— Feld wählen —',
    keinWert: '— Wert wählen —',
    ueberschrift: 'Massenänderung',
    erklaerung:
      'Setzt ein Feld auf allen angehakten Karten — in einem Vorgang. Geht etwas schief, wird keine einzige geändert.',
    auswaehlen: 'Für die Massenänderung anhaken',
    mehrereWaehlen: 'Mehrere wählen',
    ausgewaehlt: 'angehakt',
    keineAuswahl:
      'Keine Karte angehakt. Hake in der Liste mehrere an, um sie zusammen zu ändern.',
    feld: 'Feld',
    wert: 'Neuer Wert',
    anwenden: 'Anwenden',
    betroffen:
      'An diesen Karten ist die Änderung gescheitert. Geändert wurde keine einzige:',
    gescheitert: 'gescheitert',
    ohneWert: 'noch kein Wert',
  },
  stilTitel: 'Stil',
  stilWaehlen: 'Stil dieses Sets',
  stilLeer:
    'Kein Stil installiert. Ohne Stil zeigt die Vorschau einen Behelfsplan.',
  stilFrage: 'Diesen Stil übernehmen?',
  stilSpielwechsel: 'Der neue Stil gehört zu einem anderen Spiel.',
  stilUngedeutet:
    'Diese Felder kennt das neue Spiel nicht. Ihre Werte bleiben im Bestand — gezeichnet werden sie nicht mehr:',
  stilVerloreneOptionen:
    'Diese Stiloptionen führt der neue Stil nicht. Ihre Werte gehen verloren:',
  stilUebernehmen: 'Stil übernehmen',
  stilAbbrechen: 'Stil behalten',
  stilKeiner:
    'Das Set nennt einen Stil, der nicht installiert ist — wähle einen der installierten.',
  stiloptionen: 'Stiloptionen',
  // **„4 von 11“ und der Rest dahinter** — das Artboard *Eigenschaften*
  // (UXR-07, #1915).
  stiloptionenSichtbar: '{zahl} von {gesamt}',
  stiloptionenWeitere: 'Weitere Optionen ({zahl})',
  marktTitel: 'Wo es diese Karte gibt',
  marktErklaerung:
    'Diese Wege entstehen auf deinem Rechner — aus dem Kartennamen oder einer Anbieterkennung. Hinaus geht erst dein Klick, im Browser deiner Wahl. Ein Schlüssel ist dafür nicht nötig; ob dort etwas zu kaufen ist, entscheidet der Anbieter.',
  marktGenau: 'führt auf genau diese Karte',
  marktSuche: 'führt auf eine Suche',
  marktNurAnbieter: 'führt nur zum Anbieter',
  umschlagAnnehmen: 'Karte aus einem Umschlag',
  umschlagAngenommen: 'Die Karte liegt jetzt in diesem Set.',
}
