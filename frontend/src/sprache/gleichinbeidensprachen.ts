// **Die Schlüssel, die in beiden Sprachen gleich heißen dürfen** —
// herausgezogen für Werkstück A13.3 von Phase 5.3
// (#1452),
// als `sprache.test.tsx` an der Zeilengrenze aus §2 stand.
//
// **Der Schnitt folgt der Frage**: dort steht, *dass* übersetzt wird, hier,
// *wo nicht* — und jeder Eintrag trägt seinen Grund. Eine Ausnahme ohne
// Begründung wäre eine Ausrede (§10).

// Pfade, die in beiden Sprachen absichtlich gleich lauten.
//
// **Jeder Eintrag hier ist eine Behauptung, und sie ist zu begründen.** Die
// Liste ist die Stelle, an der eine vergessene Übersetzung sich verstecken
// könnte: wer sie aufnimmt, statt zu übersetzen, macht den Test blind.
//
// | Pfad | Warum |
// | --- | --- |
// | `allgemein.anwendungsname` | ein Eigenname |
// | `seiten.start.ueberschrift` | derselbe Eigenname |
// | `seiten.sets.titel` | „Sets" ist im Deutschen dasselbe Wort — MSE nennt sie so, und „Sammlungen" wäre eine Übersetzung, die kein Anwender wiedererkennt |
// | `seiten.decks.ausSetSet` | die Beschriftung der Setwahl beim Deck aus einem Set (FMT-04) — derselbe Grund wie `seiten.sets.titel` |
// | `seiten.sets.spalten.name` | „Name" ist in beiden Sprachen dasselbe Wort |
// | `seiten.bestandspflege.sets` | dasselbe Wort wie in `seiten.sets.titel` |
// | `seiten.bestandspflege.unbekannt` | ein Gedankenstrich, kein Wort |
// | `seiten.ruecksprung.gegenstaende.sets`, `…decks`, `…unbekannt` | „Sets" und „Decks" sind im Deutschen dieselben Wörter — MSE nennt sie so —, und der Gedankenstrich ist ein Zeichen (O7, #1431) |
// | `seiten.ruecksprung.unklarZeile`, `…ohneNamen` | zwei Platzhalter und ein Malzeichen, kein Satz — `4 × Lightning Bolt` liest sich in jeder Sprache gleich |
// | `seiten.datenpflege.bilanzZeile` | drei Platzhalter und zwei Satzzeichen, kein Satz (O8, #1432) |
// | `seiten.technik.betrieb.webTeil` | zwei Platzhalter und ein Leerzeichen, kein Satz — „Webinhalt 300,0 MB“ (OPS02, #2348) |
// | `seiten.ruecksetzung.gruppe` | zwei Platzhalter und ein Doppelpunkt, kein Satz (O4, #1428) |
// | `seiten.einstellungen.werte.feldtiefe.normal` | dasselbe Wort, dieselbe Sache — der Wert der Einstellung zur Tiefe |
// | `seiten.karten.tiefeNormal` | „Normal" ist in beiden Sprachen dasselbe Wort — und es benennt hier keinen Grad, sondern den **Normalfall** (UX-05, #1677) |
// | `seiten.spielen.standCommanderschaden` | ein Platzhalter, ein Doppelpunkt und ein zweiter Platzhalter — kein Wort. **Der Name darin ist ein Kartenname** und wird nicht übersetzt (§9), die Zahl ist eine Zahl (PLAY-03, #1711) |
// | `seiten.spielen.gespraechZeigt`, `…gespraechZeigtAuf` | zwei bzw. drei Platzhalter und ein Doppelpunkt, kein Satz — **die Wörter darin sind übersetzt** (`zeigerZeiger` und die drei daneben), und der Rahmen trägt keines (G5, #1462) |
// | `seiten.kosten.postenZeile` | drei Platzhalter und zwei Trennzeichen, kein Satz (O12, #1439) |
// | `seiten.adressbuch.wegZeile` | zwei Platzhalter und ein Doppelpunkt, kein Satz (ID-04, #1703) |
// | `seiten.einstellungen.erbeMitWert` | zwei Platzhalter und Klammern, kein Satz — der Erbwert kommt übersetzt aus den Werten (#1993) |
// | `seiten.spielen.ereignisSatz`, `…ereignisSatzOhneKarte` | drei bzw. zwei Platzhalter, ein Doppelpunkt und ein Gedankenstrich, kein Satz — die Handlung darin kommt übersetzt aus `handlungsname`, Sitz und Karte sind Namen (#1992) |
// | `seiten.start.aufmerksamkeit.amOrt` | zwei Platzhalter und ein Doppelpunkt, kein Satz — der Satz darin kommt übersetzt aus dem Katalog (#1963) |
// | `seiten.ruecksetzung.gruppen.sets`, `…decks` | dieselben Wörter wie bei den Rücksprungsgegenständen (O4, #1428) |
// | `seiten.bestandspflege.datei` | „File" und „Datei" — **nicht** gleich; steht hier nicht |
// | `seiten.sets.mtgjsonUeberschrift` | der Name des Formats — „MTG JSON" heißt es auch auf Deutsch so |
// | `seiten.bilder.format` | „Format" ist in beiden Sprachen dasselbe Wort |
// | `seiten.druckbogen.papierA4`, `…papierLetter` | Papiernamen — „A4" heißt in jeder Sprache A4 |
// | `seiten.druckbogen.dpi150`, `…dpi300`, `…dpi600` | eine Zahl und eine Einheit; beides ist international |
// | `seiten.bilder.formatPNG`, `…formatJPG`, `…formatBMP`, `…formatTIFF`, `…formatSVG` | Formatnamen. Sie stehen im Katalog und nicht im JSX, weil `keintextimjsx` jeden Textknoten abweist — auch einen, der nicht zu übersetzen ist |
// | `seiten.pakete.katalogName`, `…katalogVersion` | „Name" und „Version" sind in beiden Sprachen dasselbe Wort |
// | `seiten.pakete.katalogOrdnungName`, `…katalogOrdnungVersion` | dieselben beiden Wörter, als Eintrag der Ordnungsliste (#244) |
// | `seiten.einstellungen.werte.schriftgroesse.normal` | „Normal" ist in beiden Sprachen dasselbe Wort (#245) |
// | `hilfe.begriffe.set.name` | dasselbe Wort wie `seiten.sets.titel`, aus demselben Grund (#679) |
// | `seiten.karten.konsole.platzhalter` | ein **MSE-Script-Ausdruck**. §9: „Nicht übersetzt wird, was aus einem MSE-Paket kommt" — `card.name` heißt in jeder Sprache so, und ein übersetzter Platzhalter wäre ein Beispiel, das nicht läuft (#1070) |
// | `hilfe.seiten.start.ueberschrift`, `hilfe.seiten.sets.ueberschrift` | dieselben Wörter wie `seiten.start.ueberschrift` und `seiten.sets.titel`, aus denselben Gründen — die Überschrift einer Seitenhilfe ist der Name der Seite (#806) |
// | `seiten.karten.symbol.ueberschrift`, `hilfe.seiten.symbol.ueberschrift`, `seiten.symbol.titel` | „Symbol" ist in beiden Sprachen dasselbe Wort, und MSE nennt das Setfeld so — `set field: name: symbol` steht in jedem Spielpaket. Ein übersetzter Laschenname zeigte auf ein Feld, das anders heißt (#890) |
// | `seiten.sammlung.foil`, `…nonfoil`, `…etched` | Handelsbegriffe. Wer Karten kauft, liest „Foil" auch auf einer deutschen Seite; „veredelt" wiedererkennt niemand (#1165). **`etched` kam mit T2 dazu** (#1467) und aus demselben Grund — „geätzt" steht auf keinem Marktplatz |
// | `seiten.sammlung.zustaende.*` | die Skala von **Cardmarket** — ein fremder Begriff (§9). Wer eine Sammlung von dort einliest, findet dieselben sieben Wörter wieder |
// | `hilfe.ueber.quelleSatz` | eine **Adresse** und kein Satz. `github.com/…` ist in jeder Sprache dieselbe Zeichenkette; eine übersetzte führte nirgendwohin (#1168) |
// | `seiten.start.weiterarbeiten.arten.set`, `…arten.deck` | „Set" und „Deck" sind in beiden Sprachen dasselbe Wort — dieselben wie `seiten.sets.titel` und die Deckseite (#1422) |
// | `seiten.technik.einladung.eingabeBeispiel` | die **Gestalt** eines Einladungscodes — `mse4mac-kontakt:1:` ist eine Marke, die das Programm prüft, und ein übersetztes Beispiel zeigte einen Code, den es abweist (NET-01, #1705) |
// | `seiten.pakete.anlegenArtInclude` | „Include" ist MSE2s Name der Paketart — `IncludePackage::typeName` gibt `include`, und die Endung heißt `.mse-include`. Ein übersetzter Name zeigte auf eine Art, die niemand wiederfindet (FMT-02, #1691) |
// | `seiten.technik.schnittstellen.lauschwahl.arten.tunnel` | „Tunnel (VPN)" ist die Art einer Schnittstelle (`utun`) und heißt in beiden Sprachen so; „VPN" ist das Kürzel, das macOS in den Netzwerkeinstellungen selbst zeigt (G08, #2272) |
import { gleicheZeichen } from './gleichezeichen'

export const gleichInBeidenSprachen = new Set([
  'seiten.einstellungen.erbeMitWert',
  'seiten.spielen.ereignisSatz',
  'seiten.spielen.ereignisSatzOhneKarte',
  ...gleicheZeichen,
  'seiten.technik.betrieb.webTeil',
  'seiten.technik.einladung.eingabeBeispiel',
  'seiten.technik.schnittstellen.lauschwahl.arten.tunnel',

  // **Nur Platzhalter und ein Mittelpunkt** — Dateiname und Größe setzt
  // das Fenster ein (NET-04, #1708).
  'bausteine.teilen.datei',

  // **Die Kartenzustände sind englische Fachwörter** — Werkstück 9 von
  // Phase 4.12 (#1351), Punkt 3. „Near Mint" und „Light Played" stehen so auf
  // jeder Handelsseite und in jeder Sammlungsliste; „fast neuwertig" erkennt
  // niemand wieder, der Karten tauscht. Dieselbe Überlegung wie bei
  // `seiten.decks.sideboard`.
  'seiten.einstellungen.werte.vorgabe.zustand.mint',
  'seiten.einstellungen.werte.vorgabe.zustand.nearMint',
  'seiten.einstellungen.werte.vorgabe.zustand.excellent',
  'seiten.einstellungen.werte.vorgabe.zustand.good',
  'seiten.einstellungen.werte.vorgabe.zustand.lightPlayed',
  'seiten.einstellungen.werte.vorgabe.zustand.played',
  'seiten.einstellungen.werte.vorgabe.zustand.poor',

  // **„Code" ist in beiden Sprachen dasselbe Wort** — Werkstück 7 von
  // Phase 4.12 (#1351). Dieselbe Überlegung wie bei `seiten.decks.name`:
  // jede Authenticator-App nennt ihn so, und „Kennzahl" erkennt niemand
  // wieder.
  'bausteine.nachweis.beschriftung',

  // **„PDF", „Text" und „Name" sind in beiden Sprachen dieselben Wörter** —
  // UX-13 (#1685). Die Art einer Datei heißt so im Finder beider Sprachen,
  // und „Bezeichnung" statt „Name" wäre eine Übersetzung ins Ungewohnte.
  'bausteine.dateiansicht.arten.pdf',
  'bausteine.dateiansicht.arten.text',
  'bausteine.dateiansicht.name',
  // **Name und Profil in einer Zeile** (AP-11, #2368) — nur Platzhalter
  // und ein Mittelpunkt.
  'bausteine.kontext.zeile',

  // **Ein Posten der Markenpflege ist nur Platzhalter** — UX-12 (#1684):
  // Marke und Zahl, in beiden Sprachen gleich gesetzt.
  'bausteine.objektmarken.pflegePosten',

  // **Der Ort vor dem Grund ist nur Platzhalter** (#1963): die Datei eines
  // Skriptbefunds und der übersetzte Satz, getrennt durch einen Doppelpunkt
  // — im Französischen mit Leerzeichen davor, deshalb steht er im Katalog.
  'seiten.start.aufmerksamkeit.amOrt',

  // **„Decks" ist im Englischen dasselbe Wort** — und „Stapel" wäre eine
  // Übersetzung, die kein Spieler wiedererkennt. Dieselbe Überlegung wie bei
  // `seiten.sets.titel`.
  'seiten.decks.titel',
  'hilfe.seiten.decks.ueberschrift',

  // **„Name" und „Format" sind in beiden Sprachen dasselbe Wort** — dieselbe
  // Begründung wie bei `seiten.sets.spalten.name`.
  'seiten.decks.name',
  'seiten.decks.format',
  'seiten.turniere.format',
  'seiten.spielen.format',

  // **„Sideboard" ist ein Lehnwort aus dem Englischen** — MSE, Scryfall und
  // jede Deckliste nennen es so, und „Nebenbank" erkennt kein Spieler wieder.
  'seiten.decks.sideboard',
  // **Dasselbe Wort in der Leiste über dem Deck** — AP-9 (#2366).
  'seiten.decks.leiste.sideboard',
  'seiten.decks.sideboardgroesse',

  // **Dieselben Wörter bei den Turnieren** (#1294): „Name" ist in beiden
  // Sprachen dasselbe, und „OMW", „GW" und „OGW" sind die Abkürzungen aus
  // den Magic Tournament Rules — sie zu übersetzen hieße, sie unlesbar zu
  // machen für jeden, der schon einmal eine Rangliste gesehen hat.
  'seiten.turniere.name',
  'seiten.turniere.gegnermatchquote',
  'seiten.turniere.spielquote',
  'seiten.turniere.gegnerspielquote',

  'seiten.pakete.anlegenArtInclude',

  'hilfe.ueber.quelleSatz',
  'seiten.start.weiterarbeiten.arten.set',
  'seiten.start.weiterarbeiten.arten.deck',
  'seiten.karten.konsole.platzhalter',
  'seiten.sammlung.foil',
  'seiten.sammlung.nonfoil',
  'seiten.sammlung.etched',
  'seiten.sammlung.zustaende.mint',
  'seiten.sammlung.zustaende.nearMint',
  'seiten.sammlung.zustaende.excellent',
  'seiten.sammlung.zustaende.good',
  'seiten.sammlung.zustaende.lightPlayed',
  'seiten.sammlung.zustaende.played',
  'seiten.sammlung.zustaende.poor',
  'seiten.karten.symbol.ueberschrift',
  'seiten.karten.verlauf.urheberServer',
  'hilfe.seiten.symbol.ueberschrift',
  'seiten.symbol.titel',
  'seiten.bestandspflege.sets',
  'seiten.bestandspflege.unbekannt',

  // **»Sets« und »Decks« sind im Deutschen dieselben Wörter** — MSE nennt sie
  // so, und dieselbe Überlegung wie bei `seiten.ruecksprung.gegenstaende.sets`
  // (O4, #1428).
  'seiten.ruecksetzung.gruppen.sets',
  'seiten.ruecksetzung.gruppen.decks',
  'allgemein.anwendungsname',

  // **„Start" heißt in beiden Sprachen „Start"** — UX-02
  // (#1674).
  // Dasselbe Wort, dieselbe Bedeutung; `seiten.start.ueberschrift` steht aus
  // demselben Grund weiter unten.
  'hilfe.begriffe.set.name',
  'seiten.start.ueberschrift',
  'hilfe.seiten.start.ueberschrift',
  'hilfe.seiten.sets.ueberschrift',
  'seiten.sets.titel',
  'seiten.decks.ausSetSet',
  'seiten.sets.spalten.name',
  'seiten.sets.mtgjsonUeberschrift',
  'seiten.bilder.format',
  'seiten.bilder.formatPNG',
  'seiten.bilder.formatJPG',
  'seiten.bilder.formatBMP',
  'seiten.bilder.formatTIFF',
  'seiten.bilder.formatSVG',
  'seiten.druckbogen.papierA4',
  'seiten.druckbogen.papierLetter',
  'seiten.druckbogen.dpi150',
  'seiten.druckbogen.dpi300',
  'seiten.druckbogen.dpi600',
  'seiten.pakete.katalogOrdnungName',
  'seiten.pakete.katalogOrdnungVersion',
  'seiten.einstellungen.werte.schriftgroesse.normal',
  'seiten.pakete.katalogName',
  'seiten.pakete.katalogVersion',

  // **„Set" heißt in beiden Sprachen Set** — es ist der Begriff von MSE und
  // von Scryfall, und eine deutsche Übersetzung („Edition") stünde neben dem
  // Setcode, den das Paket englisch führt (§9,
  // #1200).
  'seiten.sammlung.setcode',

  // **»Set« heißt in beiden Sprachen so** — es ist der Begriff des Formats
  // (`.mse-set`) und nicht ein deutsches Wort, das zufällig gleich aussieht
  // (§9: was aus einem MSE-Paket kommt, wird nicht übersetzt). Die Aufschrift
  // der Setwahl im Deck steht deshalb hier.
  'seiten.decks.setWaehlen',

  // **»Deck« heißt in beiden Sprachen so.** Es ist der Begriff des Spiels und
  // steht in jedem englischen wie deutschen Regeltext gleich.
  'seiten.turniere.deck',

  // **Die Namen der Anbieter sind Eigennamen** — Werkstück 4 von Phase 4.12
  // (#1351).
  //
  // GitHub heißt GitHub, X heißt X. Ein „Übersetzen" wäre hier kein
  // Übersetzen, sondern ein Umbenennen — und wer nach „X" sucht, fände es
  // nicht mehr. Dieselbe Regel wie bei Setnamen aus einem MSE-Paket (§9).
  'seiten.nutzer.anbieternamen.github',
  'seiten.nutzer.anbieternamen.google',
  'seiten.nutzer.anbieternamen.microsoft',
  'seiten.nutzer.anbieternamen.x',

  // **Ein Emblem heißt in beiden Sprachen so** — CR 114.1 nennt es
  // „emblem“, und das Deutsche hat das Wort übernommen. Eine Übersetzung
  // erkennte niemand wieder, der Magic spielt.
  'seiten.spielen.tischArt.emblem',

  // **Und „Name" ist in beiden Sprachen dasselbe Wort** — PLAY-03.AK-05
  // (#1711).
  // Derselbe Grund wie bei `seiten.sets.spalten.name` und
  // `seiten.pakete.katalogOrdnungName`: es benennt dieselbe Spalte, nach der
  // geordnet wird. Die Nachbarn in derselben Liste — „Zuletzt geändert" und
  // „Zug" — sind übersetzt.
  'seiten.spielen.partienOrdnungName',

  // **Papiernamen** — PLAY-05.AK-03
  // (#1713).
  // Derselbe Grund wie bei `seiten.druckbogen.papierA4` und `…papierLetter`:
  // „A4" heißt in jeder Sprache A4, und „Letter" ist der Name des Formats
  // und kein englisches Wort.
  'seiten.spielen.abzugBlattA4',
  'seiten.spielen.abzugBlattLetter',

  // **„Mana" und „Commander" sind in beiden Sprachen dasselbe Wort** —
  // PLAY-02.AK-02
  // (#1710).
  // Beides sind Begriffe des Spiels (CR 106 und CR 903) und keine deutschen
  // Sätze; die elf Bereiche daneben sind übersetzt.
  'seiten.spielen.regelbereichMana',
  'seiten.spielen.regelbereichCommander',

  // **„Mulligan" ebenso** — #1839.
  // Ein Begriff des Spiels (CR 103.5) und kein deutsches Wort; die übrigen
  // Wörter des Beginns sind übersetzt.
  'seiten.spielen.standMulligans',

  // **Zwei Platzhalter und ein Pfeil, kein Satz** — „Aschenwächter →
  // Hallenwart" liest sich in jeder Sprache gleich (Artboard *Tisch*).
  'seiten.spielen.kampfGegen',
  'seiten.handlungen.handlungsname.spieler.mulligan',

  // **„Sanskrit" und „Quenya" heißen in beiden Sprachen so** — UX-06
  // (#1678).
  // Die übrigen achtzehn Sprachen eines Drucks sind übersetzt.
  'bausteine.drucksprachen.namen.sa',
  'bausteine.drucksprachen.namen.qya',

  // **„Set" und „Deck" heißen in beiden Sprachen so, und der Rest ist ein
  // Platzhalter** — der Ort einer Zeile in der Befehlspalette (UX-02.AK-02,
  // #1674). Die Zahl darin kommt übersetzt aus `zaehlwort`.
  'allgemein.palette.alsSet',
  'allgemein.palette.alsDeck',

  // **Jede Sprache heißt in ihrer eigenen Schrift** — LOC-01
  // (#1687).
  // Wer die Oberfläche nicht lesen kann, findet seine Sprache trotzdem; so
  // steht es auch in der Sprachwahl von macOS.
  'allgemein.sprache.de',
  'allgemein.sprache.en',
  'allgemein.sprache.es',
  'allgemein.sprache.fr',
  'allgemein.sprache.it',
  'allgemein.sprache.ja',
  'allgemein.sprache.ru',
  'allgemein.sprache.zh',
  'seiten.einstellungen.werte.sprache.de',
  'seiten.einstellungen.werte.sprache.en',
  'seiten.einstellungen.werte.sprache.es',
  'seiten.einstellungen.werte.sprache.fr',
  'seiten.einstellungen.werte.sprache.it',
  'seiten.einstellungen.werte.sprache.ja',
  'seiten.einstellungen.werte.sprache.ru',
  'seiten.einstellungen.werte.sprache.zh',

  // **„Name" heißt in beiden Sprachen so** — das Feld einer eigenen
  // Übersetzung (LOC-02, #1688), wie `seiten.sets.spalten.name`.
  'seiten.einstellungen.eigenuebersetzung.name',

  // **Ein Gedankenstrich, kein Wort** — die freie Ziffer in der Belegung
  // von ⌘0 bis ⌘9 (UX-02.AK-03, #1674).
  'seiten.einstellungen.belegungFrei',

  // **Zwei Platzhalter und ein Punkt dazwischen** — die Kennung eines Stils
  // mit seinem fremden Spiel, und Spiel und Stil eines neuen Sets (UX-05,
  // #1677). Übersetzt kommen die Teile, nicht der Punkt.
  'bausteine.stilliste.mitZusatz',
  'seiten.sets.neuWahl',
  // **Die Zusagen eines Handels** (UX-08.AK-04, #1680) — Stückzahl und Name,
  // die Zeile aus Wörtern, die jede Sprache selbst einsetzt.
  'seiten.handeln.zusagenPosition',
  'seiten.handeln.zusagenZeile',
  // **Der Verlauf eines Handels** (TRADE-03.AK-03, #1717) — „Revision" ist
  // in beiden Sprachen dasselbe Wort, der Rest sind eingesetzte Zustände und
  // eine Zeit.
  'seiten.handeln.assistent.verlaufWechsel',
  'seiten.handeln.assistent.verlaufNeu',

  // **„Name" heißt auf Englisch so** — der letzte Schritt des Assistenten.
  'seiten.sets.neuSchrittName',

  // **Zwei Platzhalter und ein Punkt dazwischen** — Name und Fingerabdruck
  // eines Gegenübers zur Wahl (UX-08.AK-04, #1680). Der Name ist der, den
  // jemand vergeben hat, der Abdruck eine Zahl; übersetzt wird keines.
  'bausteine.kontaktwahlEintrag',
  // **Und Mensch und Profil zur Wahl** — dieselbe Gestalt, zwei Namen, die
  // jemand vergeben hat.
  'seiten.decks.uebernahmeProfilEintrag',
])
