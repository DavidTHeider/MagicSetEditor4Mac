// Die Texte der **Turniere** — Werkstück 5 von Phase 4.7
// (#1294),
// nach Vermerk 0062.
//
// **Ein eigener Baustein**, wie `seitendecks.ts` und aus demselben Grund:
// `seiten.ts` steht an der Zeilengrenze aus §2.
export const turniere = {
  alleTurniere: 'Alle Turniere ({anzahl})',
  titel: 'Turniere',
  erklaerung:
    'Ein Turnier ist ein Verlauf über die Zeit: Runden, Paarungen, Ergebnisse. Die Rangliste wird daraus gerechnet und nirgends abgelegt — wer ein Ergebnis berichtigt, sieht sofort den neuen Stand.',

  leer: 'Noch kein Turnier. „Turnier anlegen …“ oben beginnt eines, „Turnier einlesen …“ holt eines aus einer Datei.',

  // Die Anlage.
  neu: 'Turnier anlegen',

  // **Der Knopf im Kopf der Seite** — UXR-12 (#1920): er öffnet den
  // Assistenten in einer Schublade, und die Auslassungspunkte sagen es.
  neuOeffnen: 'Turnier anlegen …',
  name: 'Name',
  spiel: 'Spiel',
  notiz: 'Notiz',
  einzelspiel: 'Ein Match aus einem Spiel',
  einzelspielHinweis:
    'Dann sagen Spielquote und Gegnerspielquote nichts — sie sind Kopien der Matchzahlen. Gerechnet werden sie trotzdem.',
  anlegen: 'Anlegen',

  // **Der Assistent** — DOM-04.AK-01 (#1699): Format, Regeln, Teilnehmende
  // und erste Runde, in dieser Reihenfolge. Die Pflichtangaben sagen, warum.
  assistent: 'Ein neues Turnier in vier Schritten',
  schrittTurnier: 'Turnier',
  schrittRegeln: 'Regeln',
  schrittTeilnehmende: 'Teilnehmende',
  schrittErsteRunde: 'Erste Runde',
  weiter: 'Weiter',
  zurueck: 'Zurück',
  verlassen: 'Später weitermachen',
  vorlage: 'Vorlage',
  ohneVorlage: 'Ohne Vorlage',
  vorlageHinweis:
    'Eine Vorlage bringt Spiel, Format, Regeln und Notiz mit. Auf Wunsch auch die Namen, nie die Decks: eine eingereichte Liste gilt für ein Turnier.',
  mitTeilnehmenden: 'Die Namen der Teilnehmenden übernehmen',
  nameWarum:
    'Pflicht — damit du es in der Liste wiederfindest. Zwei Turniere eines Spiels heißen verschieden.',
  spielWarum: 'Pflicht — es bestimmt, welche Formate und Decks passen.',
  format: 'Format',
  ohneFormat: 'Kein Format',
  formatWarum: 'Wahlfrei — zur Wahl stehen die Formate dieses Spiels.',
  regelnAusVorlage: 'Die Regeln kommen aus der Vorlage.',
  schweizerSystem:
    'Gepaart wird nach Schweizer System: wer gleich viele Punkte hat, spielt gegeneinander, und niemand zweimal gegen denselben. Wie viele Runden es werden, schlägt das Programm nach der Zahl der Teilnehmenden vor.',
  zuWenige:
    'Eine Runde braucht zwei, die gegeneinander spielen — melde mindestens zwei an.',
  ersteRundeHinweis:
    'So fiele die erste Runde aus. Ausgelost wird erst mit dem Knopf.',
  ersteRundeAuslosen: 'Runde 1 auslosen',

  // Die Übersicht.
  waehlen: 'Öffnen',
  loeschen: 'Turnier löschen',
  // **Ausgeben, Bericht und Löschen hinter „…“** — AP-6 (#2363), B-04.
  handlungenAn: 'Handlungen — {name}',
  handlungenKurz: '…',
  // Das Menü in der Leiste — Zweitblick über Phase 6, Befund 6.
  turniermenue: 'Turnier',
  loeschenFrage:
    'Dieses Turnier löschen? Runden, Paarungen und Ergebnisse gehen mit. Die Decks der Teilnehmer bleiben.',
  loeschenJa: 'Löschen',
  loeschenNein: 'Abbrechen',

  // Die Teilnehmer — **„Teilnehmende" wie im Artboard und im Assistenten**
  // (UXR-12, #1920): eine Lasche, die anders hieß als der Schritt, der
  // dieselben Menschen anmeldet, waren zwei Namen für eine Sache.
  teilnehmer: 'Teilnehmende',
  teilnehmerLeer: 'Noch niemand angemeldet.',
  anmelden: 'Anmelden',
  gebiete: 'Gebiete dieses Turniers',
  gebietTeilnehmer: 'Teilnehmende',
  gebietRunden: 'Runden',
  gebietRangliste: 'Rangliste',

  // **Abschluss und Protokoll** — DOM-04.AK-04 (#1699), Vermerk 0139.
  abschliessen: 'Abschließen',
  abgeschlossenAm:
    'Abgeschlossen am {datum}. Teilnehmende, Runden und Ergebnisse stehen fest.',
  wiedereroeffnen: 'Wiedereröffnen',
  nichtWiedereroeffnen: 'Nicht wiedereröffnen',
  wiedereroeffnenFrage: 'Dieses Turnier wiedereröffnen?',
  wiedereroeffnenErklaerung:
    'Danach lassen sich Teilnehmende, Runden und Ergebnisse wieder ändern. Das Protokoll hält fest, dass es geschah.',
  gebietProtokoll: 'Protokoll',
  protokollLeer: 'Noch nichts geschehen.',
  protokollAngelegt: 'Angelegt als »{angaben}«',
  protokollAngemeldet: '{angaben} angemeldet',
  protokollAusgestiegen: '{angaben} ausgestiegen',
  protokollEingestiegen: '{angaben} wieder angemeldet',
  protokollRunde: 'Runde {angaben} gelegt',
  protokollErgebnis: 'Ergebnis {paarung}: {angaben}',
  protokollZurueckgenommen: 'Ergebnis {paarung} zurückgenommen',
  protokollAbgeschlossen: 'Abgeschlossen',
  protokollWiedereroeffnet: 'Wiedereröffnet',
  protokollEingelesen: 'Eingelesen als »{angaben}«',

  // **Datei und Bericht** — DOM-04.AK-03 (#1699), Vermerk 0140.
  ausgeben: 'Ausgeben …',
  berichtSichern: 'Bericht sichern …',
  einlesen: 'Turnier einlesen …',
  gesichertUnter: 'Gesichert unter {pfad}.',
  eingelesen: 'Eingelesen und aufgeschlagen.',
  einlesebefunde: 'Beim Einlesen fiel auf:',

  // **Kopf und Verknüpfung nach dem Artboard** — DOM-04.AK-02 (#1699).
  kopfzeile:
    'Schweizer System · Runde {runde} von {von} · {anzahl} Teilnehmende',
  kopfzeileOhneRunde: 'Schweizer System · {anzahl} Teilnehmende',
  kopfzeileOhneEmpfehlung:
    'Schweizer System · Runde {runde} · {anzahl} Teilnehmende',
  gebietDecklisten: 'Decklisten',
  decklistenLeer:
    'Noch hat niemand eine Liste eingereicht — wer beim Anmelden ein Deck wählt, reicht dessen Liste ein.',
  deckAufschlagen: 'Deck aufschlagen',
  partieOeffnen: 'Partie öffnen',
  deck: 'Deck',
  ohneDeck: 'Ohne Deck',
  deckHinweis:
    'Wird ein Deck genannt, wird seine Liste abgezogen und nicht verwiesen: eine Deckliste im Turnier ist ein eingereichtes Dokument.',
  deckabzug: 'Eingereichte Liste',
  deckabzugLeer: 'Kein Deck eingereicht.',
  deckabzugUnbenannt: 'Karte',
  aussteigen: 'Abmelden',
  einsteigen: 'Wieder anmelden',
  ausgestiegen: 'ausgestiegen',
  ausgestiegenHinweis:
    'Er wird nicht mehr gepaart und bleibt in der Rangliste: seine Quote geht weiter in die seiner Gegner ein.',

  // Der Verlauf.
  runden: 'Runden',
  rundenLeer: 'Noch keine Runde gepaart.',
  runde: 'Runde {nummer}',
  empfehlung:
    '{gelegt} von {empfohlen} Runden — die Turnierregeln empfehlen {empfohlen} für {spieler} Teilnehmer. Ein Vorschlag, keine Vorschrift.',
  vorschlag: 'Nächste Runde ansehen',
  vorschlagTitel: 'So sähe die nächste Runde aus',
  vorschlagHinweis:
    'Angesehen, nicht gelegt. Erst der Knopf darunter legt sie an.',
  rundeLegen: 'Runde legen',
  bye: 'Bye (2–0)',
  gegen: 'gegen',
  offen: 'noch nicht gespielt',
  spiele: 'Spiele',
  spieleA: 'Spiele für den Ersten',
  spieleB: 'Spiele für den Zweiten',
  unentschiedenSpiele: 'Unentschieden',
  eintragen: 'Eintragen',
  zuruecknehmen: 'Zurücknehmen',

  // **Der Knopf an einer Paarung** — UXR-12 (#1920), wie im Artboard
  // *Turniere*: er klappt das Eintragen an dieser Zeile auf.
  ergebnis: 'Ergebnis',

  // **Die Folgen einer Berichtigung** — DOM-04.AK-02 (#1699).
  folgenFrage: 'Diese Berichtigung verschiebt die Rangliste.',
  folgenPlatz:
    '{name}: Platz {vorher} → {nachher}, {punkteVorher} → {punkteNachher} Punkte',
  folgenSpaeter:
    'Runde {runden} wurde nach dem alten Stand gepaart und bleibt, wie sie ist.',
  folgenEintragen: 'Trotzdem eintragen',
  folgenNicht: 'Nicht eintragen',

  // **Die drei Sätze eines Spielstands** — Werkstück U3 von Phase 5.1
  // (#1414).
  //
  // Sie sind aus `domain.Ergebnis.Pruefe` gelesen und nicht erfunden: keine
  // negativen Zahlen, kein Ergebnis ohne ein Spiel, und höchstens so viele
  // Spiele, wie ein Match hier hat (`service.hoechstensSpiele`: bei einem
  // Einzelspielturnier eines, sonst keine Grenze).
  ergebnisZahlen: 'Spiele zählt man in ganzen Zahlen ab null.',
  ergebnisOhneSpiel:
    'Ein Ergebnis ohne ein einziges Spiel ist kein Ergebnis — wer nicht gespielt hat, hat noch keines.',
  ergebnisZuViele: 'Ein Match geht hier über höchstens {hoechstens} Spiele.',

  // Die Rangliste.
  rangliste: 'Rangliste',
  ranglisteLeer: 'Noch nichts gespielt.',
  platz: 'Platz',
  bilanz: 'Bilanz',
  matchpunkte: 'Punkte',
  gegnermatchquote: 'OMW',
  spielquote: 'GW',
  gegnerspielquote: 'OGW',
  quotenHinweis:
    'Vier Tiebreaker in der Reihenfolge der Magic Tournament Rules: Punkte, Gegnermatchquote, Spielquote, Gegnerspielquote. Jede Quote hat eine Untergrenze von 0,33.',

  // **Eine Paarung, an eine Partie gebunden** — Werkstück G7 von Phase 5.5
  // (#1464).
  match: 'Hier gespielt',
  matchErklaerung:
    'Wer diese Begegnung in diesem Programm spielt, bindet sie an die Partie. Das Ergebnis wird dann von beiden Seiten gemeldet und erst eingetragen, wenn beide dasselbe sagen.',
  matchNichtGebunden: 'An dieser Paarung hängt keine Partie.',
  matchAnsehen: 'Nachsehen',
  matchPartie: 'Kennung der Partie',
  matchSitzA: 'Sitz des ersten Teilnehmers',
  matchSitzB: 'Sitz des zweiten Teilnehmers',
  matchBinden: 'An diese Partie binden',
  matchGebunden: 'Gebunden an die Partie',
  matchSitz: 'Sitz',
  matchDeckabdruck: 'Deckabdruck',
  matchKeineMeldung: 'Noch hat niemand gemeldet.',
  matchMeldungVon: 'gemeldet von',
  matchMelden: 'Für diesen Sitz melden',
  matchAusDerPartie: 'Ergebnis aus der Partie melden',
  matchAusDerPartieHilfe:
    'Spielstand und Nachweis kommen aus der beendeten Partie — getippt wird nur, wer meldet.',
  matchNachweis: 'Nachweis (letztes Kettenglied)',
  matchUebernehmen: 'Ergebnis übernehmen',
  matchUebernommen: 'Das Ergebnis steht im Turnier.',
  matchGrenze:
    'Der Nachweis sagt, dass beide Seiten denselben Verlauf gesehen haben. Er sagt nicht, dass richtig gespielt wurde — und eine private Runde ist keine offizielle Turnierzulassung.',
}
