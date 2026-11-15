// Die Texte der **Decks** — Werkstück 7 von Phase 4.6
// (#1293),
// nach Vermerk 0061.
//
// **Ein eigener Baustein**, weil `seiten.ts` an der Zeilengrenze aus §2 steht
// — dieselbe Aufteilung wie bei `seitensammlung.ts`.
export const decks = {
  titel: 'Decks',
  erklaerung:
    'Zusammenstellungen aus eigenen Karten und echten Drucken. Ein Deck verweist auf Karten, statt sie zu kopieren — eine berichtigte Karte im Ursprungsset erreicht es damit sofort.',

  leer: 'Noch kein Deck. „Deck anlegen“ darüber beginnt eines. Ein Deck gehört zu einem Spiel; ohne das gäbe es kein Format zum Prüfen.',

  // Die Anlage.
  neu: 'Deck anlegen',
  name: 'Name',
  spiel: 'Spiel',
  notiz: 'Notiz',
  anlegen: 'Anlegen',
  // **Aus einem Set ein Deck** — FMT-04 (#1693).
  ausSetTitel: 'Oder aus einem Set',
  ausSetErklaerung:
    'Jede Karte des Sets wird ein Platz im Hauptdeck — verwiesen, nicht kopiert: wer sie im Set berichtigt, berichtigt sie im Deck.',
  ausSetSet: 'Set',
  ausSetMenge: 'Menge je Karte',
  ausSetAnlegen: 'Aus dem Set anlegen',
  ausSetKeinSet:
    'Noch kein Set mit Karten — ein Deck aus einem Set braucht eines.',

  // Die Übersicht.
  hauptdeck: 'Hauptdeck',
  sideboard: 'Sideboard',
  zaehlung: '{haupt} im Hauptdeck, {side} im Sideboard',
  waehlen: 'Öffnen',
  offenMarke: 'offen',
  umbenennen: 'Umbenennen',
  // **Der neue Name und sonst nichts** — die Notiz reist unverändert mit.
  neuerName: 'Neuer Name',
  umbenennenTitel: '{name} umbenennen',
  nameSichern: 'Sichern',
  umbenennenAbbrechen: 'Nicht umbenennen',
  // **Die Kopie** — DOM-03.AK-02. *Duplizieren* und nicht *Kopieren*: das
  // Wort des Katalogs für „ein zweites davon" (⌘D), wie bei den Partien;
  // *Kopieren* ist die Zwischenablage. Der Vorschlag ist ein Vorschlag: im
  // Spiel ist der Name einmalig, und wer eine zweite Kopie zieht, ändert ihn.
  duplizieren: 'Duplizieren',
  kopieTitel: '{name} duplizieren',
  kopieName: 'Name der Kopie',
  kopieVorschlag: '{name} (Kopie)',
  kopieAnlegen: 'Kopie anlegen',
  kopieAbbrechen: 'Nicht kopieren',
  // **Was fehlt, auf die Wunschliste** — ein schon gewünschter Druck bleibt,
  // wie er ist, und wird mitgezählt.
  aufWunschliste: 'Was fehlt, auf die Wunschliste',
  wunschlisteErgebnis: 'Neu auf der Wunschliste: {neu} · schon darauf: {schon}',
  loeschen: 'Deck löschen',
  // **Das Menü am offenen Deck** — AP-6 (#2363): Umbenennen, Duplizieren,
  // Teilen und Löschen hinter einem Knopf. `handlungenKurz` ist sein
  // Zeichen, sein Name steht in `handlungenAn`.
  handlungenAn: 'Handlungen — {name}',
  handlungenKurz: '…',
  // Das Menü in der Leiste — Zweitblick über Phase 6, Befund 6.
  deckmenue: 'Deck',
  loeschenFrage: 'Dieses Deck löschen? Die Karten selbst bleiben.',
  loeschenJa: 'Löschen',
  loeschenNein: 'Abbrechen',

  // Die Plätze.
  plaetze: 'Plätze',
  plaetzeLeer: 'Noch keine Karte im Deck.',
  bezug: 'Karte oder Druck',
  menge: 'Menge',
  bereich: 'Bereich',
  setWaehlen: 'Set',
  karteWaehlen: 'Karte',
  karteSuchen: 'Karte suchen',
  karteKeine: 'Keine Karte passt zu »{suche}«.',
  setLeer:
    'Kein Set offen — Karten aus eigenen Sets lassen sich erst wählen, wenn eines da ist.',
  ohneSet: 'Set wählen …',
  ohneKarte: 'Karte wählen …',
  setOhneKarten: 'Dieses Set hat noch keine Karte.',
  kurve: 'Manakurve',
  kurveHinweis:
    'Wie viele Karten je Manawert. Gezählt wird, wessen Wert bekannt ist: bei einem Druck von Scryfall der dort geführte, sonst der aus den Kosten gerechnete. Länder tragen keine Kosten und stehen deshalb nicht darin — eine eigene Karte mit leerem Kostenfeld ebenso wenig.',
  farben: 'Farben',
  farbeWeiss: 'Weiß',
  farbeBlau: 'Blau',
  farbeSchwarz: 'Schwarz',
  farbeRot: 'Rot',
  farbeGruen: 'Grün',
  farbenHinweis:
    'Eine zweifarbige Karte zählt in beiden Farben — die Summe ist deshalb größer als das Deck.',
  karten: 'Karten',
  gebietFormate: 'Formate',
  formateKnopf: 'Formate …',

  // **Ein Deck aus einem fremden Profil holen** — Werkstück 12 von Phase 4.12
  // (#1351).
  //
  // **Die Erklärung sagt, was NICHT mitkommt.** Wer ein Deck holt und zwölf
  // Karten weniger bekommt, soll es vorher gelesen haben und nicht hinterher
  // zählen (§6).
  uebernahme: 'Ein Deck aus einem fremden Profil holen',
  auswahlLoesen: 'Alle Decks zeigen',
  // **Die Leiste über dem offenen Deck** — AP-9 (#2366).
  leiste: {
    bezeichnung: 'Kennzahlen des Decks',
    sideboard: 'Sideboard {zahl}',
    pruefung: 'Prüfung',
    ohneFormat: 'ohne Format',
    nichtGeprueft: 'nicht geprüft',
    fehlt: 'Fehlt in der Sammlung',
    unbekannt: '{zahl} unbekannt',
    nichtAbgeglichen: 'nicht abgeglichen',
  },
  uebernahmeErklaerung:
    'Holen darf, wer in das andere Profil wechseln darf. Hat dessen Eigner einen zweiten Faktor, braucht es dort eine Freigabe — »Leser« genügt. Nicht mit kommen: Plätze, die auf eine eigene Karte des anderen Profils zeigen, und das Format — beides sind Zeilen dort drüben. Was nicht mitkam, steht danach hier.',
  uebernahmeProfil: 'Kennung des fremden Profils',
  uebernahmeDeck: 'Kennung des Decks',
  // **Gewählt und nicht getippt** — UX-08.AK-04 (#1680).
  uebernahmeProfilwahl: 'Aus welchem Profil?',
  uebernahmeProfilWaehlen: 'Ein Profil wählen …',
  uebernahmeProfilEintrag: '{nutzer} · {profil}',
  uebernahmeKeinProfil: 'Es gibt kein anderes Profil auf diesem Rechner.',
  uebernahmeDeckwahl: 'Welches Deck?',
  uebernahmeDeckWaehlen: 'Ein Deck wählen …',
  uebernahmeKeinDeck: 'In diesem Profil liegt kein Deck.',
  uebernahmeFachleute: 'Kennungen selbst nennen',
  uebernahmeFachleuteErklaerung:
    'Profil und Deck lassen sich auch über ihre Kennungen nennen — so, wie sie die Kommandozeile und der Dienst führen. Die Erlaubnis ist dieselbe.',
  uebernahmeHolen: 'Holen',
  uebernahmeGeschrieben: '„{name}“ steht jetzt bei deinen Decks — {karten}.',
  uebernahmeNichtMit: 'Nicht mitgekommen:',

  gebietKarten: 'Karten',
  gebietAuswertung: 'Auswertung',
  gebietPruefung: 'Prüfung',
  gebieteDeck: 'Gebiete dieses Decks',
  legen: 'Ins Deck legen',
  nehmen: 'Herausnehmen',
  druck: 'Druck (Scryfall-Kennung)',

  // **Die drei Wege, die vom Deck wegführen** — UX-03.AK-01
  // (#1675),
  // wörtlich aus dem Entwurf (`docs/prototyp/project/Decks.dc.html`).
  // **Und derselbe Satz, solange kein Deck offensteht.** Ohne ihn stünde die
  // Karte da und niemand sähe, dass sie auf etwas wartet (§6).
  anschlussOhneDeck:
    '„{name}“ wartet — schlag ein Deck auf oder leg eines an, dann steht sie in seiner Kartenwahl.',

  holenAus: 'Karten holen aus',
  holenSammlung: 'Meiner Sammlung',
  holenSet: 'Einem Set',
  // **Die seltenen Wege in einem Kasten** — AP-6 (#2363), Entwurf
  // `Bereiche09`: Übernahme, Umschlag und „Karten holen aus“.
  wegeTitel: 'Holen und weitergeben',
  mitspielen: 'Mit diesem Deck spielen',

  // **Was aus der Sammlung mitgereist ist** — UX-03.AK-01
  // (#1675).
  //
  // **Er sagt, was noch fehlt.** Der Druck steht schon im Feld; Menge und
  // Bereich sind eine Entscheidung, und ein Satz, der nur meldet, dass etwas
  // angekommen ist, ließe den Menschen suchen, was jetzt dran ist (§6).
  anschlussWartet:
    '„{name}“ ist aus der Sammlung mitgereist — Menge und Bereich wählen, dann liegt sie im Deck.',

  // Das Format.
  format: 'Format',
  formatWaehlen: 'Format wählen',
  ohneFormat: 'Kein Format — ungeprüft',
  formatHinweis:
    'Ein Deck ohne Format ist ungeprüft, nicht falsch. Das Format muss zum Spiel des Decks gehören.',

  // Die Prüfung.
  pruefen: 'Gegen das Format prüfen',
  zulaessig: 'Zulässig.',
  unzulaessig: 'Nicht zulässig.',
  ungeprueftZahl: '{zahl} Karten sind ungeprüft.',
  ungepruefterklaerung:
    'Ungeprüft heißt nicht unzulässig: eine selbst entworfene Karte ist in keinem Format der Welt gelistet. Auch »restricted« steht dem Deck nicht entgegen — wie oft eine eingeschränkte Karte vorkommen darf, steht in den Turnierregeln, und die hat dieses Programm nicht gelesen.',

  // Die fünf Wörter, die ein Urteil trägt.
  legal: 'zulässig',
  not_legal: 'nicht gelistet',
  restricted: 'eingeschränkt',
  banned: 'verboten',
  ungeprüft: 'ungeprüft',

  // Der Abgleich.
  bestand: 'Was davon da ist',
  vorhanden: 'Vorhanden',
  gebraucht: 'Gebraucht',
  fehlend: 'Fehlend',
  unbekannt: 'Unbekannt',
  bestandVollstaendig: 'Alles da.',
  bestandFehlt: '{zahl} Karten fehlen.',

  // **Die Zuordnung** — Werkstück U9 von Phase 5.1
  // (#1420).
  gebietZuordnung: 'Zuordnung',
  zuordnungen: 'Bedarf und Zuordnung',
  zuordnungenLaden: 'Wird geladen …',
  zuordnungenLeer:
    'Noch keine Karte im Deck. Was hier steht, entsteht aus den Plätzen.',
  zuordnungHinweis:
    'Eine Deckliste reserviert nichts: zwei Decks mit denselben drei Karten sind beide vollständig. Zugeordnet ist, was du ausdrücklich für dieses Deck festlegst — frei ist, was danach noch übrig bleibt.',
  zuordnungszahlen:
    '{vorhanden} von {gebraucht} da · {frei} frei · {zugeordnet} zugeordnet · {reserviert} reserviert',
  zuordnungPlatz: 'Bereich',
  zuordnungExemplar: 'Exemplar',
  zuordnungExemplarWaehlen: 'Exemplar wählen',
  zuordnungMenge: 'Menge',
  zuordnen: 'Zuordnen',
  zuordnungLoesen: 'Lösen',
  abweichend: 'anderer Druck',
  abweichendErklaerung:
    'Dieses Exemplar ist ein anderer Druck als der Platz. Ausdrücklich so gewählt — derselbe Name ist nicht dieselbe Karte.',
  fehlkarten: 'Fehlkarten als Liste',
  fehlkartenliste: 'Die fehlenden Karten',
  fehlkartenKeine: 'Es fehlt nichts.',

  // **Die Kennzahlen** — Werkstück U10 von Phase 5.1
  // (#1421).
  typen: 'Typen',
  kennzahlen: 'Kennzahlen',
  laender: 'Länder',
  durchschnitt: 'Manawert im Schnitt',
  keineZahl: '—',
  quellen: 'Farbquellen',
  quellenHinweis:
    'Gezählt werden Länder mit einem der fünf Grundtypen — sie geben ihr Mana ohne Regeltext (CR 305.6). Ein Land, das seine Farbe im Text nennt, steht hier nicht: es zu deuten wäre geraten.',

  // **Die Plätze am Deckbauplatz** — Werkstück U10 von Phase 5.1 (#1421).
  weniger: '−',
  wenigerVon: 'Eine weniger',
  mehr: '+',
  mehrVon: 'Eine mehr',
  insSideboard: 'Ins Sideboard',
  insHauptdeck: 'Ins Hauptdeck',
  suche: 'Karten suchen',
  deckbereich: 'Das Deck',
  analyse: 'Analyse',
  analyseLeer:
    'Auswertung, Prüfung und Zuordnung des aufgeschlagenen Decks stehen hier.',
  // **Stapel und Wunsch** — UX-03.AK-01 (#1675).
  stapelTitel: 'Aus der Sammlung',
  stapelWartet: '{name} aus der Sammlung, noch in keinem Deck: {liste}.',
  stapelLegen: 'Alle ins Deck legen',
  stapelTeilweise: '{gelegt} von {gesamt} liegen im Deck, der Rest nicht:',
  stapelOhneDruck:
    '„{name}“ hat keine Scryfall-Kennung — ein Deck nimmt nur Drucke mit einer auf.',
  stapelGrund: '„{name}“: {grund}',
  eintauschen: 'Die {anzahl} eintauschen',
  deckplatz: 'Deckbauplatz',
  pruefungOhneFormat:
    'Ohne Format wird nicht geprüft. Wähle im Deck ein Format — bis dahin ist es ungeprüft, nicht falsch.',
  pruefungLaeuft: 'Wird geprüft …',
  postenUnzulaessig: 'nicht zulässig',
  postenUnzulaessigErklaerung:
    'Diese Karte ist im gewählten Format nicht erlaubt. Was genau, sagt die Lasche »Prüfung«.',
  keinDeckOffen:
    'Kein Deck aufgeschlagen. Wähle links eines aus oder lege ein neues an.',

  // Die Formate.
  formate: 'Formate',
  formateLeer:
    'Noch kein Format. Ein Format ist hier ein Datensatz und keine eingebaute Liste — es wird angelegt, nicht mitgeliefert.',
  bauart: 'Bauart',
  bauarten: 'Nachgelesene Bauarten',
  bauartHinweis:
    'Die Zahlen stehen in den Comprehensive Rules. Sie sind ein Vorschlag; jede lässt sich danach ändern.',
  scryfallname: 'Scryfall-Name',
  scryfallnameHinweis:
    'Der Schlüssel in Scryfalls legalities, etwa »standard«. Leer heißt: Scryfall kennt dieses Format nicht, und jede Karte bleibt ungeprüft.',
  mindestgroesse: 'Mindestens',
  hoechstgroesse: 'Höchstens',
  sideboardgroesse: 'Sideboard',
  sideboardOffen: 'Sideboard ohne Grenze',
  hoechstzahlGleich: 'Gleiche Karten',
  grenzenHinweis:
    'Null heißt: keine Grenze. Beim Sideboard heißt Null: kein Sideboard.',
  formatNeu: 'Format anlegen',
  formatSichern: 'Format sichern',
  formatLoeschen: 'Format löschen',
  formatLoeschenFrage: '{name} löschen?',
  formatLoeschenJa: 'Löschen',
  formatLoeschenHinweis:
    'Die Decks, die darauf standen, bleiben — sie sind danach ungeprüft.',
  umschlagTitel: 'Als Umschlag weitergeben',
  umschlagErklaerung:
    'Ein Umschlag trägt ein Deck an einen anderen Rechner — signiert, mit Prüfsumme und ohne Konto. Was hereinkommt, wird ein neues Deck; überschrieben wird nichts. **Mit** geht der Name, das Spiel und jeder Platz auf einen Druck; **hier bleiben** die Notiz, die Herkunft und Plätze auf eigene Karten — deren Kennung gilt nur in diesem Bestand.',
  umschlagSchnueren: 'Offenes Deck ablegen',
  umschlagAnnehmen: 'Deck aus einem Umschlag',
  umschlagOhneDeck: 'Schlag zuerst ein Deck auf.',
  umschlagGeschnuert: 'Abgelegt.',
  umschlagAngenommen: 'Angenommen',
  einfuhrTitel: 'Eine Deckliste einlesen',
  einfuhrMenue: 'Eine Deckliste einlesen …',
  einfuhrErklaerung:
    'Füge eine Deckliste ein — eine Karte je Zeile, „4 Lightning Bolt (M11) 149“. Überschriften gliedern: unter „Sideboard“ steht das Sideboard. Die Vorschau schreibt nichts; erst das Übernehmen legt die Zeilen ins offene Deck, und es nimmt nichts heraus.',
  einfuhrFeld: 'Deckliste',
  einfuhrVorschau: 'Vorschau',
  einfuhrUebernehmen: 'Deckliste übernehmen',
  einfuhrLeer: 'Setze zuerst eine Liste ein.',
  einfuhrOffen: 'offen',
  einfuhrBilanz: 'Hauptdeck / Sideboard',
  einfuhrGelegt: 'Übernommen',
  einfuhrStueck: 'Stück',
  einfuhrNichtsOffen: 'Keine Zeile ist eindeutig — nichts zu legen.',
} as const
