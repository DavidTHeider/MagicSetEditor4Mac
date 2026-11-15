// **Was in beiden Sprachen gleich steht, weil es kein Wort ist** — Zeichen,
// Zahlen, Platzhalter, Tastenkürzel und Funktionsnamen.
//
// # Warum eine eigene Datei
//
// **Die Zeilengrenze** (§2): `gleichinbeidensprachen.ts` stand bei 403
// Zeilen. Der Schnitt folgt der Frage — dort stehen **Wörter**, die im
// Englischen dieselben sind, hier steht, was gar keines ist: ein Strich, ein
// Prozentzeichen, `{zahl} ms`, `⌘K`, `sort_list`.
//
// **Der Unterschied trägt**: bei einem Wort ist „nicht übersetzt" eine
// Entscheidung, die man prüfen kann; bei einem Zeichen gäbe es gar nichts zu
// entscheiden.

export const gleicheZeichen = [
  // **Ein Strich, ein Minus und ein Plus sind Zeichen, keine Sprache** —
  // Werkstück U10 von Phase 5.1
  // (#1421).
  // Sie stehen trotzdem im Katalog: was auf einem Knopf steht, gehört
  // dorthin, wo die Texte stehen — der **Name** des Knopfes ist übersetzt
  // (`wenigerVon`, `mehrVon`), und den liest eine Vorlesehilfe vor.
  // **„Rulings" ist der Fachbegriff und kein englisches Wort** — Werkstück
  // A16 von Phase 5.3
  // (#1515).
  // Er steht so in den Scryfall-Daten, in MSE und in jedem deutschen
  // Regelforum; eine Übersetzung („Schiedsrichterentscheidungen") benannte
  // etwas anderes als das, was im Index liegt.
  'seiten.recherche.rechercheFallRulings',
  // **„Deck" ist im Englischen dasselbe Wort** — Werkstück G1 von Phase 5.5
  // (#1458).
  // Es steht so in MSE, in Scryfall und auf jedem Turniertisch; eine
  // „Übersetzung" wäre eine Erfindung.
  'seiten.spielen.sitzDeck',

  // **Und „Hand" ebenso** — dasselbe Wort, dieselbe Bedeutung, dieselbe
  // Herkunft. Eine erfundene Abweichung wäre schlechter als die Gleichheit.
  'seiten.spielen.zoneHand',
  'seiten.decks.keineZahl',
  'seiten.decks.weniger',
  'seiten.decks.mehr',
  // **Und am Lebensstand dieselben zwei** — P4 (#2372); der Name des Knopfes
  // ist übersetzt (`lebenWeniger`, `lebenMehr`).
  'seiten.spielen.lebenMinus',
  'seiten.spielen.lebenPlus',
  // **Die Anbieternamen sind Eigennamen** —
  // #1401.
  // GitHub schreibt sich in jeder Sprache so, und eine „Übersetzung" wäre
  // eine Erfindung. Dieselbe Überlegung wie bei `allgemein.anwendungsname`.
  'seiten.einstellungen.anbietername.github',
  'seiten.einstellungen.anbietername.google',
  'seiten.einstellungen.anbietername.microsoft',
  'seiten.einstellungen.anbietername.x',

  // **Ein Beispiel für eine Konfigurationsdatei** ist Code, keine Sprache —
  // Werkstück O2 von Phase 5.2 (#1426).
  'seiten.nutzer.tokenOnboardingBeispiel',

  // **Ein Tastenkürzel ist keine Sprache** — Werkstück U14 von Phase 5.1
  // (#1490): ⌥⌘S heißt auf jeder Tastatur so.
  'allgemein.navigation.seitenleisteKuerzel',
  // **Eine Zahl mit Einheit** — Werkstück U13 von Phase 5.1 (#1424): »ms«
  // ist in beiden Sprachen das Zeichen für die Millisekunde.
  'seiten.karten.konsole.dauer',

  // **»Set« ist der Begriff aus MSE** (§9: nicht übersetzt, was aus einem
  // Paket kommt) — als Gruppierung der Sammlung heißt er in beiden Sprachen
  // gleich (U8, #1419).
  'seiten.sammlung.gruppierungSet',
  'seiten.sammlung.spalteSet',
  // Eine Einheit — „12 s“ ist in beiden Sprachen dieselbe (#1422).
  'seiten.start.auftraege.dauer',
  // Die Kommandozentrale (#1423): „Text“, „JSON“ und „12 ms“ sind in beiden
  // Sprachen dieselben.
  'bausteine.kommandozentrale.alsText',
  'bausteine.kommandozentrale.alsJson',
  'bausteine.kommandozentrale.dauer',
  // **Prozent stehen hier nicht mehr** — sie galten als „in beiden
  // Sprachen gleich" und sind es nicht: im Englischen steht kein
  // Abstand vor dem Zeichen. `alsProzent` schreibt sie jetzt
  // (LOC-03.AK-01, #1689).

  // **„Bye" ist ein Lehnwort aus dem Englischen** — die Magic Tournament
  // Rules nennen es so, und „Freilos" erkennt am Turniertisch niemand
  // wieder. Dieselbe Überlegung wie bei „Sideboard".
  'seiten.turniere.bye',
  // **Der Platzhaltersatz des Durchgereichten** (#1178). Er lautet in beiden
  // Sprachen `{0}` und **ist** der Satz aus Go: die 48 Meldungen der
  // Standardbibliothek nennen Funktions- und Argumentnamen, und ein Eintrag je
  // Funktion wäre eine zweite Abschrift der Bibliothek in zwei Sprachen.
  'fehler.skript.durchgereicht',

  // **`sort_list` ist ein Funktionsname** und bleibt einer.
  'fehler.skript.keineListe',

  // **Und dieselben zwei in der Rückkehr** (O7, #1431): „Sets" ist im
  // Deutschen dasselbe Wort, und der Gedankenstrich ist ein Zeichen.
  'seiten.ruecksprung.gegenstaende.sets',
  'seiten.ruecksprung.gegenstaende.decks',
  'seiten.ruecksprung.unbekannt',
  'seiten.ruecksprung.unklarZeile',
  'seiten.ruecksprung.ohneNamen',

  // **Drei Platzhalter und zwei Satzzeichen** (O8, #1432) — kein Satz, und
  // `Feldwerte ohne ihre Karte: Weggeräumt (3)` liest sich in jeder Sprache
  // gleich.
  'seiten.datenpflege.bilanzZeile',

  // **Zwei Platzhalter und ein Doppelpunkt** (O4, #1428) — kein Satz, und
  // `Karten: 107` liest sich in jeder Sprache gleich.
  'seiten.ruecksetzung.gruppe',
  'seiten.karten.tiefeNormal',
  'seiten.einstellungen.werte.feldtiefe.normal',
  'seiten.spielen.standCommanderschaden',
  'seiten.spielen.gespraechZeigt',
  'seiten.spielen.gespraechZeigtAuf',

  // **Drei Platzhalter und zwei Trennzeichen** (O12, #1439) — kein Satz, und
  // `karte.text · abgerechnet · 12,00 €` liest sich in jeder Sprache gleich.
  'seiten.kosten.postenZeile',

  // **Eine Adresse oder Nummer mit ihrer Art** (ID-04, #1703):
  // `E-Mail: ada@example.org` — zwei Platzhalter und ein Doppelpunkt.
  'seiten.adressbuch.wegZeile',

  // **⇞ und ⇟ sind Zeichen und keine Wörter** (#1290). Sie stehen so auf der
  // Tastatur und in MSEs Menü (`data/en.mse-locale/locale`, Zeilen 48–49,
  // Fassung f9d9356 vom 2021-07-06); eine „Übersetzung" wäre eine zweite
  // Schreibweise für dieselbe Taste.
  'seiten.karten.blaettern.kuerzelVorige',
  'seiten.karten.blaettern.kuerzelNaechste',

  // **Der Platzhalter der Listeneinfuhr ist eine Kartenliste** (#1292), und
  // Kartennamen werden nicht übersetzt (§9): »4 Lightning Bolt (M11) 149«
  // steht in jedem Deckbaukasten so, gleich in welcher Sprache. Eine
  // „Übersetzung" hieße, ein Beispiel zu erfinden, das niemand einfügen würde.
  'seiten.sammlung.listePlatzhalter',

  // **Die Laschennamen der Deckseite** heißen in beiden Sprachen gleich:
  // »Decks« ist der Begriff des Spiels, »Formate« unterscheidet sich nur um
  // ein Zeichen — und ein erfundener Unterschied wäre keine Übersetzung.
  'seiten.decks.gebietDecks',

  // **»Name« heißt in beiden Sprachen so** — ein erfundener Unterschied wäre
  // keine Übersetzung, sondern ein anderes Wort für dieselbe Sache.
  'seiten.nutzer.neuerNameFuer',
  // **Ein Zeichen ist keine Sprache** — Werkstück U2 von Phase 5.1 (#1413):
  // das „…" auf dem Menüknopf einer Paketzeile steht in beiden Katalogen,
  // damit `keintextimjsx` es nicht als Text im Quelltext liest; der Name des
  // Knopfs (`handlungenAn`) ist übersetzt.
  'seiten.pakete.handlungenKurz',
  // **Dasselbe Zeichen am Menü einer Setkarte** — UXR-04 (#1912); der Name
  // des Knopfs (`seiten.sets.handlungenAn`) ist übersetzt.
  'seiten.sets.handlungenKurz',
  // **Dasselbe Zeichen am Menü eines Decks** — AP-6 (#2363); der Name des
  // Knopfs (`seiten.decks.handlungenAn`) ist übersetzt.
  'seiten.decks.handlungenKurz',
  // Dasselbe Zeichen am offenen Turnier — AP-6 (#2363), B-04.
  'seiten.turniere.handlungenKurz',
  // **„Deck“ ist im Englischen und Französischen dasselbe Wort** — der Name
  // des Menüs in der Leiste (Zweitblick über Phase 6, Befund 6).
  'seiten.decks.deckmenue',
  // **Dasselbe Zeichen am Menü einer Partie** — UXR-13 (#1921); der Name
  // (`partienHandlungen`) nennt die Partie und ist übersetzt.
  'seiten.spielen.partienHandlungenKurz',
  // **Eine Abkürzung ist keine Sprache** — Werkstück O9 von Phase 5.2
  // (#1433): »MCP« heißt in beiden Katalogen so, weil das Protokoll so
  // heißt. Die drei Nachbarn daneben — Fenster, Kommandozeile, Dienst —
  // sind übersetzt.
  'seiten.technik.flaechen.mcp',
  // **Ein Eigenname ist keine Sprache** — Werkstück A14 von Phase 5.3
  // (#1453): »Claude Code« und »Codex« heißen in beiden Katalogen so, weil
  // die Programme so heißen (0063, Entscheidung 5). Die Sätze daneben —
  // wohin das Bruchstück gehört, woher die Gestalt stammt — sind übersetzt.
  'seiten.technik.agentenNamen.claude',
  'seiten.technik.agentenNamen.codex',
  // **Ein Fachwort ist keine Sprache** — Werkstück A13.2b von Phase 5.3
  // (#1452): das »Format« einer Balanceanalyse heißt auf Englisch genauso,
  // und zwar in derselben Bedeutung — Commander, Pauper, Standard. Der Satz
  // daneben, der sagt, dass keines genannt wurde, ist übersetzt.
  'seiten.einstellungen.kiEinschaetzungFormat',
  // **Und dasselbe für das »Set«** — Werkstück A13.3 von Phase 5.3 (#1452).
  // Eine Kartenerweiterung heißt in beiden Sprachen so; MSE selbst nennt die
  // Datei `.mse-set`, und ein deutsches Wort daneben verschleierte, dass es
  // dasselbe ist (§9, „nicht übersetzt wird, was aus einem MSE-Paket
  // kommt"). Die fünf Nachbarn — Kartenname, Sammlernummer, Typzeile,
  // Bauart, Illustration — sind übersetzt.
  'seiten.einstellungen.kiErkennungArten.setcode',
  // **Eine Adresse ist keine Sprache** — Werkstück N2 von Phase 5.4
  // (#1455): `192.168.1.7:41234` und `:41234` sind Beispiele im leeren Feld
  // und werden nicht gelesen, sondern abgetippt. Wer sie übersetzte,
  // übersetzte eine Zahl. Die Sätze daneben — was das Feld bedeutet, was
  // der Port ohne Wirt heißt — sind übersetzt.
  'seiten.technik.kontakteAdresseBeispiel',
  'seiten.technik.horchenAdresseBeispiel',
  // **Und das Wort Mana heißt in beiden Sprachen so** — Werkstück M3 von
  // Phase 5.5 (#1605).
  // Das Regelwerk benutzt dasselbe Wort, die deutschen Karten ebenfalls, und
  // ein Ersatzwort daneben gäbe es nicht. Die Nachbarn an derselben Leiste —
  // Leben, Gift, Du — sind übersetzt.
  'seiten.spielen.standMana',
  // **„Zone" heißt in beiden Sprachen so** — Werkstück zu
  // #1623.
  // Das Regelwerk benutzt das Wort auf Englisch (CR 400), die deutsche
  // Übersetzung ebenfalls; ein Ersatzwort daneben gäbe es nicht. Die
  // Nachbarn in derselben Tabelle — Anzahl, Art, Stelle, Ziele — sind
  // übersetzt.
  'seiten.handlungen.handlungsfeld.zone',

  // **Zwei Platzhalter und ein Doppelpunkt, kein Satz** — UX-15.AK-04
  // (#1771).
  // Derselbe Grund wie bei `seiten.ruecksetzung.gruppe` und
  // `seiten.kosten.postenZeile`: `Proxy für https: http://proxy.firma:3128`
  // liest sich in jeder Sprache gleich, und **die Wörter darin sind
  // übersetzt** — `netzwegProxyHTTPS` steht daneben.
  'seiten.technik.netzwegAusUmgebung',

  // **Set und Nummer, und ein Auslassungszeichen** — P10
  // (#2378).
  // `M10 #146` liest sich in jeder Sprache gleich, und „…" steht an einer
  // Handelszeile, solange ihre Karte noch nicht nachgeschlagen ist. **Die
  // Wörter daneben sind übersetzt** — `druck.eigen`, `…ohneNamen`,
  // `…unbekannt`.
  'seiten.handeln.druck.bezug',
  'seiten.handeln.druck.laedt',

  // **Zwei Platzhalter und ein Malzeichen** — P11 (#2379), dieselbe Gestalt
  // wie `seiten.ruecksprung.unklarZeile`: `2 × Sturmrufer` liest sich in
  // jeder Sprache gleich.
  'seiten.handeln.sicht.zeile',

  // **Drei Platzhalter, Mittelpunkte und ein Malzeichen** — P12 (#2380):
  // Zustand, Sprache und Menge einer Karte im Fach; die Wörter darin kommen
  // übersetzt aus dem Katalog der Sammlung.
  'seiten.handeln.markttisch.karte',
]
