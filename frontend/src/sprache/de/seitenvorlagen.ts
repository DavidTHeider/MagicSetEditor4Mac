// Die Texte der **Paketwerkstatt** (#1164) — eine Vorlage öffnen, ändern,
// speichern.
//
// **Eine eigene Datei**, weil `seiten.ts` an der Zeilengrenze aus §2 steht.
// Ein Katalog darf sich auf mehrere Dateien verteilen, solange er als **einer**
// ankommt (§9).
export const vorlagen = {
  titel: 'Vorlagen',
  erklaerung:
    'Was eine Karte kann und wie sie aussieht, steht in Textdateien im Vorlagenpaket — Felder, Skripte, Maße. Hier lassen sie sich ändern. Der vorige Stand bleibt als Sicherung neben dem Paket liegen.',
  paket: 'Paket',
  keinPaket: 'Noch kein Paket gewählt.',
  // **Der leere Zustand sagt, was zu tun ist** (§6, Werkstück C von Phase
  // 4.10). Die Seite stand zu 88 % leer, und das einzige, was darauf stand,
  // nannte, was **nicht** ist.
  waehleEines:
    'Wähle oben ein Paket, um seine Textdateien zu sehen und zu ändern — Felder, Skripte, Maße.',
  keineDateien: 'Dieses Paket hat keine Datei zum Bearbeiten.',
  dateien: 'Dateien',
  datei: 'Datei',
  inhalt: 'Inhalt',
  speichern: 'Sichern',
  gespeichert: 'Gesichert.',
  ungespeichert: 'Nicht gesichert.',
  warnung:
    'Diese Datei gilt für jedes Set, das dieses Paket benutzt — nicht nur für das offene.',
  konflikt:
    'Die Datei hat sich seit dem Aufschlagen geändert. Lade ihre Fassung — dein Text geht dabei verloren — oder behalte deine: dann zeigen die Änderungen, was du an ihr überschreibst.',
  neuLaden: 'Die Fassung der Datei laden',
  behalten: 'Meine behalten',
  vorschau: 'Vorschau',
  vorschauErklaerung:
    'Die erste Karte des ersten Sets, das diese Vorlage benutzt — gezeichnet mit dem Text, der hier steht. Gespeichert ist dabei nichts.',
  vorschauLaeuft: 'Wird gezeichnet …',
  vorschauKeineKarte:
    'Kein offenes Set benutzt diese Vorlage. Öffne unter „Sets“ eines, das sie nennt — dann ist hier zu sehen, was eine Änderung bewirkt.',

  // **Die Werkstatt** — Werkstück U7 von Phase 5.1 (#1418): Paketbaum,
  // Editor, Inspector.
  arbeitsplatz: 'Vorlagenwerkstatt',
  baumBezeichnung: 'Pakete und Dateien',
  inspectorBezeichnung: 'Vorschau und Diagnose',
  baumSuche: 'Pakete und Dateien durchsuchen',
  baumSuchePlatzhalter: 'Paket oder Datei …',
  baumLeer: 'Kein Paket und keine Datei passt zur Suche.',
  baumOhnePakete:
    'Noch kein Paket zum Bearbeiten. Lege unter „Pakete“ eines an oder beziehe eines — dann steht es hier.',
  nurLesbar:
    'Mitgeliefert — nur lesbar. Eine eigene Kopie lässt sich bearbeiten.',
  baumNurMitgeliefert:
    'Alle Pakete hier sind mitgeliefert. Lege eine Kopie an oder unter „Pakete“ ein neues — beide lassen sich bearbeiten und stehen dann oben.',
  dateienLaden: 'Dateien werden gelesen …',
  inspectorLeer:
    'Vorschau, Diagnose und Herkunft der offenen Datei stehen hier.',
  waehleImBaum:
    'Wähle links im Baum eine Datei — Felder, Skripte, Maße stehen dort als Text und lassen sich hier ändern.',
  verwerfenFrage:
    'Die offene Datei hat ungesicherte Änderungen. Sollen sie verworfen und {datei} geöffnet werden?',
  verwerfen: 'Verwerfen und öffnen',
  hierBleiben: 'Hier bleiben',
  testkarte: 'Testkarte',
  testkarteVorgabe: 'Erste Karte des ersten Sets',
  testkarteOhneNamen: 'Karte {nummer}',
  testkarteErklaerung:
    'Eine Vorlage bricht selten an der ersten Karte. Das Beispielset „Werkstatt-Testkarten“ bringt Langtext, Sonderlayout, eine zweite Seite, ein fehlendes Bild und Sonderzeichen mit — öffne es unter „Sets“ und wechsle seinen Stil auf dieses Paket.',
  veraltet:
    'Veraltet — die Datei hat einen Befund in Zeile {zeile}; gezeigt wird der letzte gültige Stand.',
  wirdNeuGezeichnet: 'Wird neu gezeichnet …',
  diagnose: 'Diagnose',
  diagnoseFrei: 'Kein Befund in dieser Datei.',
  feldartFehlt:
    'Diesem Feld fehlt »type:«. MSE2 bricht beim Lesen des Pakets ab; hier gilt es als Textfeld.',
  feldartUnbekannt:
    'Die Feldart »{0}« kennt MSE2 nicht und bricht beim Lesen des Pakets ab. Hier gilt das Feld als Text. Bekannt sind: {1}.',
  diagnoseStelle: 'Zeile {zeile}, Spalte {spalte}',
  zurStelle: 'Zur Stelle',
  herkunft: 'Herkunft eines Namens',
  herkunftKeine:
    'Setze den Schreibzeiger auf einen Namen — dann steht hier, woher er kommt und wer ihn benutzt.',
  herkunftUnbekannt:
    '„{name}“ definiert dieses Paket nicht — der Name kommt aus der Bibliothek oder aus einem eingebundenen Paket.',
  herkunftStelle: '{art} in {datei}, Zeile {zeile}',
  zurDefinition: 'Zur Definition',
  verwendungen: 'Verwendungen',
  verwendungenKeine: 'Keine weitere Stelle nennt diesen Namen.',
  verwendungLaden: 'Verwendungen werden gesucht …',
  symbolarten: {
    'card field': 'Kartenfeld',
    'set field': 'Setfeld',
    'styling field': 'Stiloption',
    'extra card field': 'Zusatzfeld',
    script: 'Skriptname',
  } as Record<string, string>,
  symbolartSonst: 'Definition',
  aenderungen: 'Änderungen vor dem Sichern',
  aenderungenZeigen: 'Änderungen zeigen',
  aenderungenVerbergen: 'Änderungen verbergen',
  aenderungenKeine: 'Keine Änderung gegenüber dem gelesenen Stand.',
  aenderungenZusammenfassung: '{neu} Zeile(n) neu, {weg} entfernt.',
  aenderungenAbschnitt: 'Änderung {nummer}',
  zeileWeg: 'entfernt',
  zeileNeu: 'neu',
  // **Die eigene Kopie** (FMT-02.AK-04, #1691).
  kopieHinweis: 'Herkunft des Pakets',
  kopieren: 'Eigene Kopie anlegen',
  kopieMitgeliefert:
    'Dieses Paket liefert die Anwendung mit; ändern lässt es sich nicht. Lege eine eigene Kopie an — sie steht danach unter deinen Paketen, und das Original bleibt.',
  kopieFremd:
    'Dieses Paket stammt von außerhalb. Eine Änderung hier gilt als lokale Änderung und steht beim nächsten Bezug gegen die neue Fassung. Lege eine eigene Kopie an, wenn sie bleiben soll.',
  kopieOriginal:
    'Das Original bleibt. Du arbeitest an einer Kopie — {original} selbst ist unverändert.',
  // **Der Dateibaum** (FMT-02.AK-01, #1691).
  dateiaktionen: 'Dateien des Pakets',
  dateiNeu: 'Neue Datei',
  dateiName: 'Name der neuen Datei',
  dateiAnlegen: 'Anlegen',
  dateiAbbrechen: 'Abbrechen',
  dateiHinzufuegen: 'Bild oder Schrift hinzufügen …',
  dateiUmbenennen: 'Umbenennen',
  dateiNeuerName: 'Neuer Name',
  dateiUmbenennenJa: 'Umbenennen',
  dateiLoeschen: 'Löschen …',
  dateiLoeschenFrage:
    'Löschen? Die Datei liegt danach in der Sicherung neben dem Paket.',
  dateiLoeschenJa: 'Ja, löschen',
  dateiErstSichern: 'Erst sichern oder verwerfen.',
  beiwerk: 'Beiwerk',
  beiwerkHinweis:
    'Kein Text — das Paket nimmt diese Datei, wie sie ist. Umbenennen und Löschen gehen hier.',
  // **Der große Baum faltet** (UXR-18, #1926): das Beiwerk steht hinter
  // einer Zeile, und keine Liste zeigt mehr als ein Stück auf einmal.
  beiwerkZeile: 'Beiwerk ({anzahl})',
  weitereZeigen: 'Weitere zeigen ({gezeigt} von {gesamt})',
} as const
