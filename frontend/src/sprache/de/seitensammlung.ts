// Die Texte der **Sammlung** — Werkstück 2 und 3 von Phase 3.4
// (#1165).
//
// **Ein eigener Baustein**, weil `seiten.ts` an der Zeilengrenze aus §2 steht
// — dieselbe Aufteilung wie bei `seitenkarten.ts`.
export const sammlung = {
  titel: 'Sammlung',
  erklaerung:
    'Was du wirklich besitzt — je Druck, Zustand, Sprache und Ausführung eine Zeile. Das hat mit den Sets nichts zu tun: eine entworfene Karte ist ein Entwurf, ein Exemplar ein Stück Papier.',

  // **Der Kopf und seine zwei Schubladen** — UXR-10
  // (#1918).
  // Bis hierher stand alles in einem Kasten im Navigator, bei leerer
  // Sammlung offen: achtzehn Felder in 240 Punkten.
  aufnehmenOeffnen: 'Aufnehmen …',
  einfuhrOeffnen: 'Einlesen und ausgeben …',
  einfuhrTitel: 'Einlesen und ausgeben',
  leerTitel: 'Noch nichts eingetragen',
  leer: 'Ein Exemplar braucht die Scryfall-Kennung seines Drucks — die UUID aus der Adresse einer Kartenseite. Eine Liste liest du in einem Zug ein.',

  druck: 'Druck',
  menge: 'Menge',
  zustand: 'Zustand',
  sprache: 'Sprache',
  ausfuehrung: 'Ausführung',
  foil: 'Foil',
  nonfoil: 'Nonfoil',

  // **Die dritte Ausführung** — Werkstück T2 von Phase 5.6 (#1467).
  // **Nicht übersetzt**, wie »Foil« und »Nonfoil« auch: wer Karten kauft,
  // liest den Begriff auf jedem Marktplatz so.
  etched: 'Etched',
  erwerbsdatum: 'Erwerbsdatum',
  erwerbspreis: 'Erwerbspreis',
  erwerbswaehrung: 'Währung',
  preisMehrdeutig:
    'Zwei Lesarten: {erste} oder {zweite}. Bitte eindeutig schreiben.',
  notiz: 'Notiz',

  // **Der lesbare Bezug und die Sicht darauf** (#1200).
  name: 'Kartenname',
  setcode: 'Set',
  nummer: 'Nummer',
  sicht: 'Sicht auf die Sammlung',
  suche: 'Suchen',
  suchePlatzhalter: 'Name, Set, Kennung oder Notiz',
  sortierung: 'Sortieren nach',
  sortierBezeichnung: 'Karte',
  sortierMenge: 'Menge',
  sortierErwerbsdatum: 'Erwerbsdatum',
  ausfuehrungAlle: 'Alle Ausführungen',
  nurFoil: 'Nur Foil',
  nurNonfoil: 'Nur Nonfoil',
  nurEtched: 'Nur Etched',

  // **Der Arbeitsplatz** — Werkstück U8 von Phase 5.1 (#1419).
  navigator: 'Sicht und Werkzeuge',
  inspector: 'Exemplar',
  // **Der Name des Menüs „Exemplar“** in der Leiste und am Rechtsklick
  // (AP-6, #2363). Ein eigener Schlüssel, weil „Copia“ in der Menüleiste
  // neben dem Befehl Kopieren stünde; englisch heißen beide „Card Copy“.
  exemplarmenue: 'Exemplar',
  // **Der Rückweg zu den Decks** — Werkstück U9 von Phase 5.1 (#1420).
  inDecks: 'In diesen Decks',

  inspectorLeer:
    'Wähle in der Liste ein Exemplar — hier stehen dann Menge, Zustand, Erwerb und Preis, und hier änderst du es.',
  inspectorAusserhalb:
    'Dieses Exemplar steht nicht mehr in der Sicht. Die Änderung trifft trotzdem dieses Exemplar.',
  arbeitsplatz: 'Sammlungsarbeitsplatz',
  listeBezeichnung: 'Exemplare',
  spalteWahl: 'Wahl',
  spalteBezeichnung: 'Bezeichnung',
  spalteMenge: 'Menge',
  spalteZustand: 'Zustand',
  spalteSprache: 'Sprache',
  spalteSet: 'Set',
  spalteLagerort: 'Lagerort',
  gruppeOhne: 'Ohne Angabe',
  gruppeSumme: '{anzahl} Exemplare · {menge} Stück',
  kachelWaehlen: '{bezeichnung} wählen',

  // **Ansicht und Gruppierung** — Werkstück U8 von Phase 5.1 (#1419).
  ansicht: 'Ansicht',
  ansichtTabelle: 'Tabelle',
  ansichtGalerie: 'Galerie',
  ansichtGruppen: 'Gruppen',
  gruppierung: 'Gruppieren nach',
  gruppierungKeine: 'Nicht gruppieren',
  gruppierungSet: 'Set',
  gruppierungLagerort: 'Lagerort',
  gruppierungZustand: 'Zustand',
  gruppierungSprache: 'Sprache',
  ohneTreffer: 'Kein Exemplar passt zu Suche und Filter.',
  ohneTrefferZuruecksetzen: 'Suche und Filter zurücksetzen',
  gezeigt: '{gezeigt} von {gesamt}',
  gekappt: '{gezeichnet} von {passend} gezeichnet',
  mehrZeigen: 'Mehr zeigen',

  // **Die Listeneinfuhr** (#1292).
  listeTitel: 'Eine Liste einfügen',
  listeErklaerung:
    'Eine Karte je Zeile, etwa »4 Lightning Bolt (M11) 149«. Leerzeilen, »Deck«, »Sideboard« und Zeilen mit # oder // werden übergangen.',
  listeFeld: 'Kartenliste',
  listePlatzhalter: '4 Lightning Bolt (M11) 149\n2x Counterspell',
  listeDeuten: 'Zuordnen',
  listeEinlesen: 'Aufnehmen',
  listeOffen:
    '{offen} von {gesamt} Zeilen brauchen eine Entscheidung — wähle den Druck, den du besitzt.',
  listeAlleEindeutig: 'Alle {gesamt} Zeilen sind eindeutig.',
  listeDruckWahl: 'Welcher Druck?',
  listeOhneWahl: '— noch nicht entschieden —',
  listeDrucke: '{drucke} Drucke',
  listeBericht: '{angelegt} neu, {erhoeht} erhöht, {stueck} Stück aufgenommen.',
  listeNichtsGewaehlt:
    'Keine Zeile entschieden — wähle mindestens einen Druck.',
  listeAusgelassen: '{ausgelassen} vorhandene Zeilen ausgelassen.',
  listeKeinDruck: 'Keine Karte heißt so — prüfe die Schreibweise.',
  listeZuViele:
    'Mehr als {drucke} Drucke heißen so — nenne das Set in Klammern, etwa »(M11)«, dann bleibt einer.',
  listeOhneKarte: '{anzahl} Zeilen finden keine Karte und werden übergangen.',
  listeOhneKarteEins: 'Eine Zeile findet keine Karte und wird übergangen.',

  // **Die Einfuhrvorschau** — DOM-03.AK-01 (#1698). Sie steht nur da, wenn
  // etwas zusammengeführt würde, und bis zur Wahl ist nichts geschrieben.
  vorschauTitel: 'Vor dem Einlesen',
  vorschauErklaerung:
    'Einige Zeilen treffen Zeilen, die deine Sammlung schon führt — gleicher Druck, Zustand, Sprache, Ausführung und Absicht. Wähle, was mit ihnen geschieht; bis dahin ist nichts geschrieben.',
  vorschauZaehlung:
    '{neu} neu · {zusammen} treffen vorhandene · {wiederholt} wiederholen eine Zeile dieser Einfuhr',
  vorschauZusammen: '{bezeichnung} ({zustand}): {vorher} → {nachher}',
  vorschauModus: 'Vorhandene Zeilen',
  vorschauAddieren: 'Mengen addieren',
  vorschauAuslassen: 'Vorhandene auslassen',
  vorschauEinlesen: 'Einlesen',
  vorschauAbbrechen: 'Nicht einlesen',

  // **Die Ausfuhr** (#1292).
  ausfuhrTitel: 'Ausgeben und sichern',
  ausfuhrErklaerung:
    '»CSV« trägt jede Angabe und lässt sich vollständig zurücklesen. »Liste« ist die einfügbare Gestalt für einen Deckbaukasten — sie trägt nur Menge, Karte, Druck und Ausführung.',
  ausfuhrAlsCSV: 'Als CSV',
  ausfuhrAlsListe: 'Als Liste',
  ausfuhrErgebnis: 'Ausgabe',
  ausfuhrZurueck: 'CSV zurücklesen',
  ausfuhrLeer: 'Noch nichts ausgegeben.',

  // **Die Ordnung** (#1292, Vermerk 0057).
  lagerort: 'Lagerort',
  wunsch: 'Wunsch, kein Besitz',

  // **Die Ordnungssicht** (#1292).
  ordnungTitel: 'Ordnung',
  ordnungLagerorte: 'Lagerorte',
  ordnungDubletten: 'Dubletten',
  ordnungWuensche: 'Wunschliste',
  ordnungBehalten: 'Behalten',
  ordnungLeerLagerorte:
    'Noch kein Exemplar — ohne eines gibt es keinen Lagerort.',
  ordnungLeerDubletten: 'Kein Druck liegt öfter als {behalten}-mal da.',
  ordnungLeerWuensche:
    'Nichts auf der Wunschliste. Ein Exemplar kommt als „Wunsch, kein Besitz“ darauf — oder an einem Deck über „Was fehlt, auf die Wunschliste“.',
  ordnungStueck: '{stueck} Stück in {zeilen} Zeilen',
  ordnungOhneOrt: 'ohne Angabe',
  ordnungUeberzaehlig: '{ueberzaehlig} zu viel',

  // **Besitze ich diese Karte?** (#1292).
  besitzJa: 'Du besitzt {stueck} davon.',
  besitzNein: 'Du besitzt diesen Druck nicht.',
  besitzGewuenscht: '({gewuenscht} auf der Wunschliste)',

  // **Die Massenvorgänge** (#1292).
  masseTitel: 'Massenvorgänge',
  masseWaehlen: 'Auswählen',
  masseGewaehlt: '{gewaehlt} von {gesamt} gewählt',
  masseLagerortSetzen: 'Lagerort setzen auf',
  masseZustandSetzen: 'Zustand setzen auf',
  masseAendern: 'Ausgewählte ändern',
  masseLoeschen: 'Ausgewählte löschen',
  masseLoeschenFrage: '{anzahl} Exemplare löschen?',
  masseLoeschenErklaerung:
    'Was hier weg ist, tippst du neu: Menge, Zustand, Lagerort und Erwerbspreis stehen nirgends sonst. Die Sammlung hat keinen Verlauf wie ein Set.',
  masseLoeschenJa: 'Löschen',
  masseBericht: '{geaendert} geändert.',
  // **Die Rücknahme** — DOM-03.AK-03 (#1698): der alte Lagerort an jede
  // Zeile zurück, gegen die neue Fassung.
  masseZuruecknehmen: 'Zurücknehmen',
  masseZuruecknehmenVon: 'Den Lagerort von {anzahl} Zeilen zurücknehmen',
  masseZurueckgenommen: '{anzahl} zurückgenommen.',
  massePreise: 'Preise für alle holen',
  massePreisschub: '{geholt} geholt, {offen} offen',
  massePreiseFehlgeschlagen: '{anzahl} fehlgeschlagen',
  massePreiseSteht:
    'Dieser Schub hat nichts bewegt — ein weiterer Versuch gäbe dasselbe.',
  massePreiseFertig: 'Alle Preise sind da.',

  aufnehmen: 'Aufnehmen',
  aenderungSichern: 'Sichern',
  loeschen: 'Löschen',
  loeschenFrage: 'Dieses Exemplar löschen?',
  loeschenErklaerung:
    'Menge, Zustand, Lagerort und Erwerbspreis stehen nirgends sonst — die Sammlung hat keinen Verlauf wie ein Set.',
  loeschenJa: 'Löschen',
  abbrechen: 'Abbrechen',
  bearbeiten: 'Bearbeiten',

  // **Die zwei Wege aus der Sammlung heraus** — UX-03.AK-01
  // (#1675).
  // Wörtlich aus dem Entwurf (`docs/prototyp/project/Sammlung.dc.html`).
  zumTausch: 'Zum Tausch',
  insDeck: 'Ins Deck',
  // **Eine Auswahl reist mit** — UX-03.AK-01 (#1675).
  auswahlTitel: 'Auswahl',
  auswahlInsDeck: '{anzahl} → Deck',

  neuesTitel: 'Ein Exemplar aufnehmen',
  // **Was ein Exemplar nicht ausmacht, eingeklappt** — UXR-10 (#1918):
  // Bezug, Erwerb und Notiz unter einem Knopf, im Formular wie im Inspector.
  weitereAngaben: 'Weitere Angaben',

  // **Derselbe Druck zweimal ist eine Entscheidung über die Menge** —
  // DOM-03, S04.NAT-01 bis 03 (#1698).
  dubletteFrage:
    '{bezeichnung} liegt schon in der Sammlung — {menge} Stück im selben Zustand, in derselben Sprache und Ausführung.',
  dubletteErklaerung:
    'Erhöhen legt die neue Menge dazu; die übrigen Angaben der vorhandenen Zeile bleiben. Abbrechen ändert nichts — deine Eingabe bleibt stehen.',
  dubletteErhoehen: 'Auf {summe} erhöhen',
  dubletteZurZeile: 'Zur vorhandenen Zeile',
  dubletteAbbrechen: 'Abbrechen',

  zustaende: {
    mint: 'Mint',
    nearMint: 'Near Mint',
    excellent: 'Excellent',
    good: 'Good',
    lightPlayed: 'Light Played',
    played: 'Played',
    poor: 'Poor',
  },

  // **Die Preise** — mit allem, was der Satz nicht sagt.
  preiseTitel: 'Preise',
  preiseHolen: 'Preise holen',
  // **Der Kasten um Marktplatz und fremden Verlauf im Inspector** — UXR-10
  // (#1918): zwei Abrufe ins Netz, eingeklappt, bis jemand sie braucht.
  marktUndVerlauf: 'Angebote und Verlauf',
  // **Der Marktplatz** — Werkstück A4 von Phase 5.3 (#1443).
  marktTitel: 'Angebote auf dem Marktplatz',
  marktErklaerung:
    'Was einzelne Verkäufer für diesen Druck verlangen — Preis, Menge, Zustand und Sprache. Der Abruf geht ins Netz. Es wird nichts gekauft.',
  marktHolen: 'Angebote holen',
  marktHolenBei: 'Angebote bei {quelle} holen',
  marktLaeuft: 'Wird geholt …',
  marktLeer: 'Zu diesem Druck steht dort gerade nichts zum Verkauf.',
  marktZugeordnet:
    'Dieser Druck ist jetzt der Kartenfassung {fassung} zugeordnet — der nächste Abruf findet sie sofort.',
  marktStueck: '{menge} Stück',
  marktBuendel: 'im Bündel zu {buendel}',
  marktImUrlaub: 'versendet gerade nicht',
  marktGradiert: 'bewertet',
  marktSigniert: 'signiert',
  marktVeraendert: 'verändert',
  marktBefunde: 'Nicht jeder Posten war zu deuten:',

  // **Die Begleitbefunde des Preisholens** (#1202).
  preisbefunde: 'Nicht jeder Preis war zu deuten. Diese Reihen fehlen:',
  preiseLaufen: 'Wird geholt …',
  keinePreise: 'Für diesen Druck liegt kein Preis vor.',
  preisAlter: 'vor {stunden} h · {quelle}',
  preisFrisch: 'gerade eben · {quelle}',
  preisVeraltet: 'veraltet, {tage} Tage alt · {quelle}',
  preisQuellstand: 'Stand der Quelle {tag}',

  // **Der Vorbehalt steht auf der Seite und nicht in einer Fußnote.**
  vorbehalt:
    'Preise sind Schätzwerte von Scryfall und dessen Partnern. Für keine Angabe wird eine Gewähr übernommen; verbindlich ist der Preis im Laden.',

  verlauf: 'Verlauf',
  keinVerlauf:
    'Für diese Reihe liegt nur ein Stand vor — ein Verlauf entsteht, sobald du wieder Preise holst.',

  bulkTitel: 'Preise einer ganzen Quelle holen',
  // **Die Leiste über der Sammlung** — AP-9 (#2366).
  leiste: {
    bezeichnung: 'Kennzahlen der Sammlung',
    exemplare: 'Exemplare',
    eintraege: '{zahl} Einträge',
    wert: 'Wert',
    keinPreis: 'kein Preis',
    ohnePreis: 'Ohne Preis',
    nichtGelesen: 'nicht gelesen',
  },
  // **Ein Druck über seinen Namen** — AP-10 (#2367).
  drucksuche: {
    karte: 'Karte',
    hinweis:
      'Name, wahlweise mit Set und Nummer — etwa Lightning Bolt (M11) 149',
    suchen: 'Suchen',
    drucke: 'Drucke',
    keinTreffer: 'Keine Karte dieses Namens gefunden.',
    weitere:
      'Es gibt weitere Drucke — mit Set und Nummer wird die Wahl kürzer.',
    treffer: '{name} · {set} ({code}) · Nr. {nummer}',
  },
  bulkErklaerung:
    'MTGJSON antwortet nicht je Karte, sondern mit einer Datei. »Heute« bringt den laufenden Tag, »Neunzig Tage« den Verlauf — und über ein Gigabyte.',
  bulkQuelle: 'Quelle',
  bulkUmfang: 'Umfang',
  bulkHeute: 'Heute',
  bulkNeunzigTage: 'Neunzig Tage',
  bulkHolen: 'Holen',
  bulkHolenBei: 'Preise bei {quelle} holen',
  bulkLaeuft: 'Wird geholt …',
  bulkBilanz: '{staende} Preise zu {drucke} Drucken abgelegt.',
  bulkOhneDrucke:
    'Kein Druck dieses Bestands trägt eine MTGJSON-Kennung — es wurde nicht gefragt.',
  uebersichtTitel: 'Übersicht',
  uebersichtLeer: 'Noch keine Preise geholt.',
  ohnePreis: 'ohne Preis: {zahl}',
  // **Ein Exemplar und nicht „1 Exemplare“.** Eine Zahl mit falscher
  // Form liest sich wie ein Fehler im Programm — gesehen im laufenden
  // Fenster.
  einExemplar: '{zahl} Exemplar',
  exemplare: '{zahl} Exemplare',

  // **Und der Grund, warum nicht umgerechnet wird.**
  keineUmrechnung:
    'Je Währung eine Summe. Umgerechnet wird nicht — dafür bräuchte es einen Wechselkurs, und den hat dieses Programm nicht.',
  // **Der Preisverlauf einer fremden Quelle** — A5, #1444.
  verlaufTitel: 'Preisverlauf einer anderen Quelle',
  verlaufErklaerung:
    'Holt eine Reihe über die Zeit und legt sie ab. Der Abruf geht ins Netz und kostet Kontingent; die Erhaltung steht so da, wie die Quelle sie nennt.',
  verlaufHolen: 'Verlauf holen',
  verlaufHolenBei: 'Verlauf bei {quelle} holen',
  verlaufLaeuft: 'Wird geholt …',
  verlaufBilanz: '{staende} Preisstände abgelegt — Kartenfassung {fassung}.',
} as const
