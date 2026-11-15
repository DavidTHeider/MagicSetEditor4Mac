// Die Texte des Bereichs **Handeln** — Werkstück T9 von Phase 5.6
// (#1474).
//
// **Ein eigener Katalog und keine Zeilen mehr in `seitentechnik.ts`.** Der
// Abgleich, die Post und die Verbuchung standen dort, weil es den Bereich
// noch nicht gab — ihre eigenen Kommentare sagen es. Jetzt gibt es ihn, und
// ein Text über das Handeln unter „Technik“ wäre eine Fundstelle, die
// niemand sucht.

import { handelsliste } from './handelsliste'
import { handelsassistent } from './handelsassistent'

export const handeln = {
  liste: handelsliste,
  assistent: handelsassistent,
  titel: 'Handeln',
  erklaerung:
    'Tauschen und handeln mit einem Gegenüber, das du bestätigt hast. Gerechnet wird auf diesem Rechner; was hinausgeht, entscheidest du.',
  gebiete: 'Bereiche des Handelns',
  gebietPartner: 'Mit wem',
  gebietMarkt: 'Was es wert ist',
  neuerTausch: 'Neuer Tausch …',
  neuerTauschTitel: 'Neuer Tausch',
  partnerTitel: 'Bestätigte Gegenüber',
  partnerErklaerung:
    'Gehandelt wird nur mit einem Gegenüber, mit dem du einen Code verglichen hast. Wähl eines, dann geht euer Tausch auf; einladen und bestätigen geht im Kasten darunter.',
  partnerLeer:
    'Noch kein bestätigtes Gegenüber. Tauscht im Kasten darunter eure Einladungscodes aus und vergleicht die Ziffern.',
  partnerWaehlen: 'Aufschlagen',
  partnerWaehlenName: '{name} aufschlagen',
  partnerOffen: 'Aufgeschlagen',
  partnerOffenName: '{name} aufgeschlagen',
  partnerKeiner:
    'Wähl links einen Handel — oder beginn mit „Neuer Tausch …“ einen neuen.',
  marktTitel: 'Was deine Sammlung wert ist',
  marktErklaerung:
    'Eine Marktbeobachtung ist keine Zusage: sie sagt, was anderswo verlangt wurde, nicht was du bekommst. Woher jede Zahl stammt und wie alt sie ist, steht an ihr — und wo keine steht, steht keine.',
  vorgangTitel: 'Ein Handel von der Eröffnung bis zur Übergabe',
  vorgangErklaerung:
    'Hier läuft der Vertrag: eröffnen, anbieten, zustimmen, vorbereiten, übergeben. Es bewegt sich dabei **kein Bestand** — verbucht wird in „Verbuchen“, und dafür fragt dieses Programm noch einmal.',
  // **Die Wahl eines Handels** — UX-08.AK-04 (#1680): gewählt statt getippt.
  wahlTitel: 'Welcher Handel',
  wahlNeu: 'Ein neuer Handel',
  wahlNeuZusatz: 'Seine Kennung vergibt das Programm beim Eröffnen.',
  wahlEintrag: '{zustand} · Revision {revision} · {positionen}',
  wahlLeer: 'Mit diesem Gegenüber gibt es noch keinen Handel.',
  wahlFachleute: 'Einen Handel über seine Kennung öffnen',
  wahlFachleuteErklaerung:
    'Jeder Handel hat eine **Kennung**. Das Programm vergibt sie beim Eröffnen, und jede Nachricht über ihn trägt sie. Steht ein Handel oben noch nicht — etwa, weil seine erste Nachricht noch unterwegs ist —, öffnest du ihn hier über sie.',
  wahlKennung: 'Kennung des Handels',
  wahlOeffnen: 'Öffnen',
  wahlGewaehlt: 'Gewählt ist der Handel {kennung}.',
  vorgangOhneKennung: 'Wähl oben einen Handel — oder eröffne einen neuen.',
  vorgangAnsehen: 'Stand ansehen',
  vorgangFachleute: 'Den Stand mit Revision und Abdruck ansehen',
  vorgangEroeffnen: 'Eröffnen',
  vorgangVorschlagen: 'Neu vorschlagen',
  vorgangRevision: 'Revision',
  vorgangAbdruck: 'Abdruck',
  vorgangParteien: 'Parteien',
  vorgangZustimmungen: 'Zugestimmt haben',
  vorgangKeine: 'Noch niemand.',
  vorgangDu: 'du',
  vorgangEnde: 'Hier geht es nicht weiter.',
  vorgangArt: 'Art',
  arten: {
    teilen: 'Teilen',
    deckleihe: 'Deckleihe',
    tausch: 'Tausch',
    markt: 'Markt',
  },
  artOhneBestand:
    'Dieser Handel bewegt keinen Bestand — verbucht wird nur ein Tausch.',
  vorgangVereinbarung:
    'Ihr seid euch einig. Das ist eine **Vereinbarung**: die Karten wechseln erst bei der Übergabe, und gezählt wird erst, wenn beide verbucht haben.',
  vorgangSchritte: 'Schritte',
  schrittAnbieten: 'Anbieten',
  schrittGegenangebot: 'Gegenangebot',
  schrittZustimmen: 'Zustimmen',
  schrittVorbereiten: 'Vorbereiten',
  schrittUebergeben: 'Übergabe bestätigen',
  schrittAblehnen: 'Ablehnen',
  schrittAbbrechen: 'Abbrechen',
  schrittAblaufen: 'Abgelaufen',
  schrittKlaeren: 'Zur Klärung',
  positionTitel: 'Was hin und her geht',
  positionVon: 'Von wem',
  positionVonIch: 'Von dir',
  positionVonGegenueber: 'Vom Gegenüber',
  positionPrinting: 'Druck',
  kennungVonHand: 'Kennung von Hand',
  positionMenge: 'Stück',
  positionGuete: 'Güte',
  positionSprache: 'Sprache',
  positionFinish: 'Ausführung',
  positionHinzufuegen: 'Zeile hinzufügen',

  // **Was aus der Sammlung mitgereist ist** — UX-03.AK-01
  // (#1675).
  anschlussWartet:
    '„{name}“ ist aus der Sammlung mitgereist — Menge und Seite prüfen, dann steht sie im Angebot.',

  // **Und derselbe Satz, solange noch kein Gegenüber gewählt ist.** Ohne ihn
  // stünde die Karte da und niemand sähe, dass sie auf etwas wartet (§6).
  // **Ein Wunsch aus einem Deck** — UX-03.AK-01 (#1675).
  wunschTitel: 'Wunschliste aus einem Deck',
  wunschWartet: '{name} aus einem Deck, noch nicht eingetragen: {liste}.',
  wunschEintragen: 'Als Wunsch vom Gegenüber eintragen',
  anschlussOhnePartner:
    '„{name}“ wartet — wähl unter „Neuer Tausch …“ ein Gegenüber, dann steht sie in seiner Zeile.',
  // **Die Karte einer Zeile** — P10 (#2378): Name, Set und Nummer statt der
  // Kennung; ein eigener Druck hat nie ein Bild.
  // **Preise holen, wo getauscht wird** — P10 (#2378), Befund C-17.
  preise: {
    holen: 'Preise holen',
    abbrechen: 'Preisabruf abbrechen',
    laeuft: 'Hole Preise … {fertig} von {gesamt}',
    fertig: 'Preise für {zahl} Drucke geholt.',
    teils: '{zahl} von {gesamt} Drucken ohne Preis: {meldung}',
    abgebrochen:
      'Abgebrochen nach {fertig} von {gesamt} — was geholt ist, bleibt.',
  },
  // **Die Sicht des Gegenübers über die Leitung** — P11 (#2379), C-15.
  // **Der Markttisch** — P12 (#2380), Entwurf `Markttisch09`.
  markttisch: {
    neuTitel: 'Neuer Tausch mit',
    neuMit: 'Neuer Tausch mit {name}',
    meinFach: 'Mein Fach',
    sammlung: 'Meine Sammlung',
    wunschliste: 'Wunschliste',
    seinFach: 'Fach von {name}',
    freigegeben: 'Was {name} zum Handeln freigibt — gefragt beim Aufschlagen.',
    meinLeer:
      'Hier liegt noch nichts. Exemplare kommen über die Sammlung herein.',
    wunschLeer: 'Keine Wünsche. Eine Karte wird zum Wunsch in der Sammlung.',
    seinLeer: 'Nichts freigegeben.',
    neuFragen: 'Neu fragen',
    hinweis: 'Ziehen oder Leertaste, Pfeile, Eingabe. Mehrfachwahl mit ⌘.',
    laedt: 'Frage …',
    karte: '{zustand} · {sprache} · {menge}×',
    ansicht: 'Ansicht',
    alsTisch: 'Als Tisch',
    alsListe: 'Als Liste',
    // **Ein großes Fach baut nicht alles auf** (#2538).
    fachSuchen: 'Im Fach suchen',
    fachZaehlung: '{gezeigt} von {gesamt}',
    fachMehr: 'Mehr zeigen',
    fachKeinTreffer: 'Keine Karte in diesem Fach heißt so.',
  },
  // **Ein Zug je Entscheidung** — P13 (#2381), Vermerk 0200.
  zug: {
    veraltet:
      'Inzwischen hat sich der Handel geändert. Sieh dir den neuen Stand an und entscheide noch einmal.',
    vorbereitenUebergeben: 'Vorbereiten und übergeben',
    uebergebenVerbuchen: 'Übergeben und verbuchen',
    anbietenZustimmen: 'Anbieten und zustimmen',
    buchungHinweis:
      '„Verbuchen“ meldet danach den Abgleich und bestätigt deinen Abschluss. Die Stimme von {name} trägst du eigens ein.',
    festhaltenOffen:
      'Welches Exemplar weggeht, steht nicht fest. Wähle es unter „Zusagen“ und drück dann noch einmal.',
    gescheitert: '{schritt} ging nicht: {meldung}',
    lesen:
      'Eingetragen ist es, nur der Stand ließ sich danach nicht lesen: {meldung}',
    ausgang:
      'Bei dir ist es eingetragen, in den Ausgang kam es nicht: {meldung}',
    unterwegs:
      'Bei dir ist es eingetragen, {name} hat es noch nicht: {meldung}',
    hinaus: 'An {name} hinausgegangen.',
  },
  sicht: {
    titel: 'Was dein Gegenüber freigibt',
    anfragen: 'Freigegebene Sicht anfragen',
    laeuft: 'Frage über die Leitung …',
    zeilen: '{zahl} freigegebene Zeilen, davon {wuensche} gesucht.',
    keinVorschlag:
      'Dazu passt nichts: kein Wunsch hier trifft ein Angebot dort.',
    gibst: 'Du gibst',
    bekommst: 'Du bekommst',
    zeile: '{menge} × {name}',
    uebernehmen: 'Vorschlag übernehmen',
    uebernommen: 'Übernommen: {zahl} Zeilen stehen im Angebot.',
  },
  druck: {
    eigen: 'eigener Druck · ohne Bild',
    ohneNamen: 'Druck ohne Namen',
    unbekannt: 'Unbekannter Druck',
    bezug: '{code} #{nummer}',
    laedt: '…',
  },
  positionEntfernen: 'Zeile entfernen',
  positionEntfernenName: 'Zeile {nummer} entfernen',
  positionLeer:
    'Noch keine Zeile. Ein Angebot ohne Position sagt nicht, worum es geht.',
  positionVorbereitung:
    'Welches deiner Exemplare weggeht, sagst du in „Zusagen“: ein Angebot nennt einen **Druck** und keine Zeile deines Bestands.',
  zusagenTitel: 'Welche Exemplare weggehen',
  zusagenErklaerung:
    'Ein Angebot nennt einen **Druck** und keine Zeile deines Bestands. Sag hier, **welches** Exemplar weggeht — sonst weiß beim Verbuchen niemand, wovon abzuziehen ist.',
  zusagenWahl: 'Welches deiner Exemplare weggeht',
  zusagenPosition: '{menge} × {name}',
  zusagenZeile: '{lagerort} · {zustand} · {sprache} · {finish} · {stueck}',
  zusagenOhneLagerort: 'ohne Lagerort',
  zusagenKeinExemplar: 'Davon ist nichts in deiner Sammlung.',
  zusagenKeinePassende:
    'Keine Zeile deiner Sammlung passt zu Zustand, Sprache und Ausführung dieser Zeile. Ergänze die Sammlung, oder lege die Zeile im Tausch mit passenden Angaben neu an.',
  zusagenFachleute: 'Eine Besitzzeile über ihre Kennung festhalten',
  zusagenFachleuteErklaerung:
    'Eine **Besitzzeile** ist eine Zeile deiner Sammlung, ihre **Fassung** zählt jede Änderung mit. Beides steht oben schon in der Wahl — hier trägst du es von Hand ein, etwa für ein Exemplar, das oben nicht steht.',
  zusagenExemplar: 'Welche Besitzzeile',
  zusagenMenge: 'Wie viele',
  zusagenFassung: 'Gesehene Fassung',
  zusagenFesthalten: 'Festhalten',
  zusagenNachsehen: 'Zusagen anzeigen',
  zusagenOhneVorgang:
    'Wähl oben einen Handel; dann steht hier, was du dafür hergibst.',
  zusagenLeer: 'Für diesen Handel ist nichts festgehalten.',
  zusagenOffen: 'festgehalten',
  zusagenUebergeben: 'übergeben',
  ausgangTitel: 'Was zum Gegenüber hinausgeht',
  ausgangErklaerung:
    'Der Stand wird unterschrieben und in den **Ausgang** gelegt. **Jetzt übertragen** baut eine Leitung zum Gegenüber und bringt hinaus, was dort liegt — liegen bleibt es trotzdem, bis das Gegenüber es bestätigt.',
  ausgangSchicken: 'In den Ausgang legen',
  ausgangNachricht: 'Kennung der Nachricht',
  ausgangWegraeumen: 'Ist angekommen — wegräumen',
  ausgangWegraeumenName: 'Ist angekommen — wegräumen: {nachricht}',
  ausgangFachleute: 'Eine Nachricht über ihre Kennung wegräumen',
  ausgangFachleuteErklaerung:
    'Jede Nachricht im Ausgang hat eine **Kennung**. Was nach einem Abriss noch aussteht, steht oben mit einem Knopf zum Wegräumen; hier räumst du eine weg, die dort nicht steht.',
  ausgangBestaetigen: 'Als angekommen wegräumen',
  ausgangOhneVorgang: 'Wähl oben einen Handel, der hinausgehen soll.',
  ausgangOhneNachricht: 'Trag zuerst die Kennung der Nachricht ein.',
  ausgangWiederaufnahme: 'Nach einem Abriss nachsehen',
  ausgangZustellen: 'Jetzt übertragen',
  ausgangOhneGegenueber: 'Wähl zuerst, zu wem übertragen werden soll.',
  ausgangGetragen: 'Hinausgegangen sind {zahl}.',
  ausgangKlaerung:
    'Zwei Verläufe stehen da. Das lässt sich nicht ausrechnen — sprecht miteinander.',
  ausgangNachzuliefern: 'Muss noch einmal hinaus',
  ausgangLeer: 'Nichts liegt im Ausgang.',
  ausgangFolge: 'Folge',
  tauschTitel: 'Was ihr tauschen könntet',
  tauschErklaerung:
    'Setz hier ein, was dein Gegenüber dir geschickt hat — seine Dubletten und seine Wünsche, so wie sein Programm sie freigegeben hat. Gerechnet wird auf diesem Rechner; hinaus geht dabei nichts.',
  tauschStand: 'Was dein Gegenüber geschickt hat',
  tauschAbgleichen: 'Abgleichen',
  tauschAusUmschlag: 'Aus einem Umschlag lesen',
  tauschUnlesbar: 'Das ist kein freigegebener Stand.',
  tauschLeer: 'Dazu passt nichts.',
  tauschGebe: 'Du gibst',
  tauschBekomme: 'Du bekommst',
  tauschVorschlag: 'Ein Vorschlag',
  tauschOhneVorschlag:
    'Für einen Vorschlag muss auf beiden Seiten etwas stehen — ein Tausch, bei dem einer nichts gibt, ist ein Geschenk.',
  tauschBefunde: 'Was sich nicht vergleichen ließ',
  postLeer: 'Nichts unterwegs.',
  postOffen: 'Unterwegs',
  postVersuche: 'Versuche',
  buchungTitel: 'Einen Handel verbuchen',
  buchungErklaerung:
    'Wenn die Karten wirklich den Besitzer gewechselt haben, trag den Handel hier ein. Dein Bestand ändert sich dabei, und das Journal hält fest, was wohin ging.',
  buchungVerbuchen: 'Verbuchen',
  buchungAbgleichen: 'Abgleich melden',
  buchungEigeneStimme: 'Abschluss bestätigen',
  buchungBestaetigen: 'Gegenseite bestätigt',
  buchungNochNicht:
    'Gerade gibt es nichts zu verbuchen: verbucht wird, wenn die Übergabe bestätigt ist.',
  buchungZustand: 'Stand',
  buchungStandEntwurf: 'Entwurf',
  buchungStandAngeboten: 'Angeboten',
  buchungStandVerhandlung: 'In Verhandlung',
  buchungStandAngenommen: 'Angenommen',
  buchungStandVorbereitet: 'Vorbereitet',
  buchungStandUebergabeOffen: 'Übergabe bestätigt',
  buchungStandLokalVerbucht: 'Bei dir verbucht',
  buchungStandAbgleichOffen: 'Abgleich läuft',
  buchungStandBeidseitigVerbucht: 'Beidseitig verbucht',
  buchungStandAbgelehnt: 'Abgelehnt',
  buchungStandAbgebrochen: 'Abgebrochen',
  buchungStandAbgelaufen: 'Abgelaufen',
  buchungStandKlaerung: 'Zu klären',
  buchungBeleg: 'Beleg',
  buchungEinseitig: 'Nur du hast unterschrieben.',
  buchungBeidseitig: 'Beide haben unterschrieben.',
  buchungOhneGegenbeleg:
    'Von der Gegenseite liegt noch kein geprüfter Beleg vor.',
  buchungGegenbeleg: 'Geprüfter Beleg von {wer}, angekommen {wann}.',
  buchungGegenbelegAbgewiesen:
    'Ein Beleg von {wer} kam {wann} an, passt aber nicht zu deinem — {mangel}',
  buchungGegenbelegRat:
    'Gleiche den Handel mit {wer} ab; ein passender Beleg ersetzt diese Meldung.',
  gegenbelegMaengel: {
    unterschrift: 'die Unterschrift stammt nicht vom Absender.',
    vorgang: 'er gehört zu einem anderen Handel.',
    revision: 'er gehört zu einem anderen Stand dieses Handels.',
    parteien: 'er nennt andere Beteiligte.',
    zeilen: 'die verbuchten Karten weichen ab.',
  },
  buchungJournal: 'Was sich bewegt hat',
  buchungJournalLeer: 'Zu diesem Handel steht noch nichts im Journal.',
  kontoTitel: 'Dein eigenes Konto bei einem Marktplatz',
  kontoErklaerung:
    'Was **du** dort anbietest und was **du** dort verkauft oder gekauft hast. Es wird nur gelesen: kein Angebot wird angelegt, nichts in einen Warenkorb gelegt, nichts gekauft und nichts versandt. Der Weg fängt **ausgeschaltet** an — unter Einstellungen → Dienste, und zwar zusätzlich zum Marktplatz selbst.',
  kontoNachsehen: 'Nachsehen',
  kontoAngebote: 'Angebote',
  kontoBestellungen: 'Bestellungen',
  kontoBewegt: 'hat Bestand bewegt',
  kontoBewegtNicht: 'hat keinen Bestand bewegt',
} as const
