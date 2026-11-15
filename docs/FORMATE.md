# Die MSE-Dateiformate

**Jede Feststellung hier nennt ihre Quelle, und die Quelle sagt, woher sie
kommt** — aus einem **Datenpaket** erschlossen oder aus dem **MSE2-Quelltext**
übernommen.

**Hier stand bis zum 2026-08-20 das Gegenteil**: „aus Datenpaketen erschlossen —
nicht aus dem MSE2-Quelltext", als Bedingung dafür, dass dieses Projekt unter
MIT stehen kann. Diese Lage gibt es seit dem 2026-08-16 nicht mehr: das Projekt
steht unter **GPL-2.0-or-later**, und aus MSE2 darf portiert werden
([0019](entscheidungen/0019-gpl-2-0-statt-mit.md), `AGENTS.md` §11). Der
Wächter über diese Datei weiß das seit damals; ihr Kopf wusste es nicht.

Die vier Angaben bleiben, und ihr Zweck hat sich verschoben: sie **unterscheiden**,
was erschlossen und was übernommen ist. Wer später etwas nachprüft, muss wissen,
welches von beidem vor ihm liegt.

---

## Schlüsselwörter — `keyword:`, `keyword mode:`, `keyword parameter type:`

Ein Spiel bringt Schlüsselwörter mit; ein Set kann eigene führen.

```
keyword:
	keyword: Flying
	match: Flying
	mode: core
	reminder: {this_or_that("creature", upper:true)} can’t be blocked except by creatures with flying or reach.
```

| Schlüssel | Heißt |
| --- | --- |
| `keyword` | der Name, „only for human use" |
| `match` | der Text, auf den es passt — mit `<atom-param>art</atom-param>` an den Stellen der Parameter |
| `reminder` | der Erinnerungstext. **Ein Skript**, kein fester Text (`StringScript`) |
| `rules` | eine Erläuterung für Menschen; erscheint nicht auf der Karte |
| `mode` | der Name eines `keyword mode:` |

**Mehr steht nicht in der Datei.** Alles Übrige an `Keyword` — der reguläre
Ausdruck, die Parameterliste, `valid` — wird beim Vorbereiten **gerechnet**.

**In den zehn kanonischen Sets steht kein einziger Block**: dort kommen die
Schlüsselwörter aus dem Spiel. Die Datei wird über
`include localized file: keywords` eingebunden und liegt je Sprache vor
(`keywords`, `keywords_de`, `keywords_fr`, `keywords_jp`, …).

**Quelle:** `magic.mse-game/keywords` — 334 `keyword:`-Blöcke, davon 134 mit
`<atom-param>`, dazu 6 `keyword mode:` und mehrere
`keyword parameter type:`; `magic.mse-game/game`, Zeile 28 (die Einbindung);
die zehn kanonischen Sets (je 0 Blöcke). Dazu MSE2:
`src/data/keyword.cpp:86–93` (`IMPLEMENT_REFLECTION(Keyword)`) und
`src/data/keyword.hpp:88–105`, Fassung f9d9356 vom 2021-07-06.
**Erschlossen am:** 2026-08-24
**Sicherheit:** belegt — die Feldnamen stehen im **Quelltext** von MSE2, nicht
nur in einem Paket; gegengeprüft an 334 Blöcken eines echten Packs.
**Testfall:** TestDieSchluesselwoerterEinesSpielsWerdenGelesen, TestDieSchluesselwortmodiWerdenGelesen, TestDieSchluesselwortparameterartenWerdenGelesen

---

### `is default` und `optional` stehen ohne Angabe auf `false`

Anders als `whole word` und `enabled` einer Autoersetzung, die ohne Angabe auf
**`true`** stehen. **Die Vorgabe ist nicht überall dieselbe**, und wer sie rät,
rät falsch: mit der falschen wäre **jeder** Schlüsselwortmodus die Vorgabe
gewesen.

**Quelle:** `src/data/keyword.cpp:20–27` (`KeywordParam::KeywordParam`),
`src/data/keyword.hpp:81` (`KeywordMode() : is_default(false)`), Fassung
f9d9356 vom 2021-07-06 — gegen `src/data/word_list.cpp:38–42` für die andere
Vorgabe.
**Erschlossen am:** 2026-08-24
**Sicherheit:** belegt — beide Vorgaben stehen als Initialisierung im
Quelltext, nicht als Verhalten erschlossen.
**Testfall:** TestDieSchluesselwortmodiWerdenGelesen, TestDieSchluesselwortparameterartenWerdenGelesen

---

### Ein Schlüsselwort ohne `match` wird abgeschaltet

`Keyword::prepare` endet mit `valid = !match_re.matches("")` — ein
Schlüsselwort, dessen Ausdruck auf den leeren Text passt, wird abgeschaltet,
„that would make MSE hang". Ein fehlendes `match` ist genau dieser Fall, und
zwar der einfachste.

**Quelle:** `src/data/keyword.cpp:206–208`, Fassung f9d9356 vom 2021-07-06.
**Erschlossen am:** 2026-08-24
**Sicherheit:** belegt — der Satz steht als Kommentar im Quelltext.
**Testfall:** TestEinSchluesselwortOhneMusterWirdUebergangen

---

### `include localized file:` lädt eine Sprachfassung

`magic.mse-game/game` bindet seine Schlüsselwörter und seine Auszählungsachsen
so ein:

```
include localized file: keywords
include localized file: statistics
```

Neben `keywords` liegen im `M15-Magic-Pack` **achtzehn** Fassungen —
`keywords_de`, `keywords_fr`, `keywords_ja`, `keywords_ru`, `keywords_ch-s`
und so fort, dazu `keywords_custom` ohne Sprachcode. Bei `statistics` dasselbe
Bild.

Die naheliegende Bedeutung: `<name>_<sprachcode>`, falls es die Datei gibt,
sonst `<name>`.

**Und das war zu kurz gedacht.** Am 2026-08-28 nachgezählt in
`full-magic-pack/data/magic.mse-game`: **vierzehn** Fassungen von `keywords`
und **dreizehn** von `statistics` — `_ch-s`, `_ch-t`, `_da`, `_de`, `_en`,
`_es`, `_fr`, `_it`, `_jp`, `_ko`, `_pl`, `_pt-br`, `_ru`, dazu
`keywords_custom` und `statistics_script`.

**Eine Datei ohne Zusatz gibt es zu keinem von beiden.** Der Rückfall auf
`<name>` suchte damit etwas, das kein Pack mitbringt — und weil er still
scheiterte, standen im Kanonlauf **zwei Befunde je Karte**, über den ganzen
Kanon 1.360.

Die Reihenfolge lautet deshalb: **gewählte Sprache, dann `_en`, dann der bloße
Name.** `_en` ist bei jedem lokalisierten Einschub dieses Packs dabei.

**Ohne diesen Schlüssel kommen 334 Schlüsselwörter nie an** — der Einbinder
kannte bis zum 2026-08-24 nur `include file:`, und `keyword:` stand damit in
keinem Knoten des Kopfes.

**Quelle:** `magic.mse-game/game`, Zeilen 28 und 33; die 18 Dateien
`keywords*` und 17 Dateien `statistics*` daneben. **Nicht** aus MSE2:
`mse2-quelle` f9d9356 vom 2021-07-06 kennt den Schlüssel nicht —
`Reader::unknownKey` (`src/util/io/reader.hpp`, Zeilen 173–186) behandelt
`include_file` und sonst keinen, und eine Suche über den ganzen Baum nach
„localized file" ist leer. Das Pack trägt `mse version: 2.5.8`.
**Erschlossen am:** 2026-08-24, **nachgeschärft am 2026-08-28**
**Sicherheit:** Vermutung — der Schlüssel steht in **einem** Paket aller
vorliegenden Packs: nachgezählt am 2026-10-10 über `full-magic-pack/data`
(ein Treffer, `magic.mse-game/game`) und `M15-Magic-Pack` (dieselbe Datei,
also kein zweiter Beleg). Der Quelltext unserer Referenzfassung kennt ihn
nicht. Gezählt sind die Dateien; dass `_en` die Vorgabe ist, bleibt erschlossen
(F05, #2247).

**Was sich am 2026-08-28 geändert hat, ist die Zahl der Belege, nicht ihre
Art**: dass es zur Grundfassung *keine* Datei gibt, ist gezählt; dass `_en` die
Vorgabe ist und nicht etwa die erste in alphabetischer Ordnung, bleibt
erschlossen.
**Testfall:** TestEinSprachverweisWirdAufgeloest, TestEineFehlendeSprachfassungFaelltZurueck, TestOhneSpracheGiltEnglisch, TestEineUnbekannteSpracheFaelltAufEnglisch

---

### `pack type:` beschreibt einen Booster als Baum

`magic.mse-game/packs` führt **30** `pack type:`-Blöcke. Ein Blatt trägt einen
Filter, eine Gruppe trägt `item:`-Posten:

```
pack type:
	name: common
	filter: card.rarity == "common" and not is_token_card() …
pack type:
	name: booster pack
	item:
		name: mythic rare or rare
		amount: 1
	item:
		name: uncommon sometimes shifted
		amount: 3
```

| Schlüssel | Heißt |
| --- | --- |
| `name` | der Name, unter dem ein Posten die Art nennt |
| `select` | die Auswahlart — zehn Möglichkeiten, `auto` ist die Vorgabe |
| `filter` | ein MSE-Script-Ausdruck über `card` |
| `item:` | ein Posten mit `name`, `amount` (Vorgabe 1) und `weight` (Vorgabe 1) |

**`auto` ist keine Auswahlart, sondern eine Ableitung**: mit Filter heißt es
`no replace`, mit Posten `all`, ohne beides bleibt es stehen und zieht nichts.

Benutzt werden im Pack `equal` (5×), `first` (4×), `no replace`, `nonempty`
und `proportional` (je 1×); der Rest steht auf `auto`.

**Quelle:** `magic.mse-game/packs` (30 Blöcke), `magic.mse-game/game`
Zeile 27 (`include file: packs`). Dazu MSE2: `src/data/pack.hpp`, Zeilen
27–76, und `src/data/pack.cpp`, Zeilen 21–46 und 75–79, Fassung f9d9356 vom
2021-07-06.
**Erschlossen am:** 2026-08-24
**Sicherheit:** belegt — die Auswahlarten und die Vorgaben stehen im
**Quelltext** von MSE2, nicht nur in einem Paket.
**Testfall:** TestDiePacksEinesSpielsWerdenGelesen, TestEinPackOhneAllesBleibtAuto

---

### `word list:` beschreibt eine Auswahlliste

`magic.mse-game/word_lists` führt **51** Listen. Eine Liste hat einen Namen und
Wörter; ein Wort ist entweder ein bloßer Wert oder ein Block:

```
word list:
	name: type-en
	word:
		name: Legendary 
		is prefix: true
		line below: true
	word: Creature
	word:
		name: Nichtstandard
		word: Plane
		word: Scheme
	word:
		script: all_sub_types(setting: "is_land")
```

| Schlüssel | Heißt |
| --- | --- |
| `name` | das Wort, wie es eingesetzt wird — **mit** seinem Leerzeichen am Ende |
| `is prefix` | ein Kästchen statt eines Punktes: es wird vorangestellt und schaltet um (Vorgabe falsch) |
| `line below` | ein Strich unter dem Eintrag (Vorgabe falsch) |
| `word:` darin | ein Untermenü — im Pack vier Ebenen tief (`race-de`) |
| `script` | rechnet die Wörter aus; das Ergebnis ist eine Zeichenkette mit Kommas |

**Ein bloßer Wert ist der Name**: `word: Creature` und ein Block mit
`name: Creature` bedeuten dasselbe.

**Welche Stelle eines Feldwerts zu welcher Liste gehört, steht im Wert** — als
`<word-list-NAME>…</word-list-NAME>`. Die Marken kommen aus dem **Wertskript**
des Feldes (`script: super_type_filter(value, …)`), nicht aus der Datei:
gespeichert steht in `Vanguard.mse-set` schlicht `Vanguard`, in der
Skriptumgebung `<word-list-type-en>Vanguard</word-list-type-en>`.

Gezählt: 432 Wörter auf oberster Ebene, 445 in Untergruppen, 100 Präfixe,
77 mit Skript, 60 mit Strich.

**Quelle:** `magic.mse-game/word_lists` (51 Blöcke), `magic.mse-game/game`
Zeile 29 (`include file: word_lists`), `magic.mse-game/card_fields` Zeilen
877 und 924. Dazu MSE2: `src/data/word_list.cpp` Zeilen 14–35 und
`src/gui/value/text.cpp` Zeilen 109–300, Fassung f9d9356 vom 2021-07-06.
**Erschlossen am:** 2026-08-24
**Sicherheit:** belegt — Vorgaben und Verhalten stehen im **Quelltext** von
MSE2; die Marken im gerechneten Wert sind an `Vanguard.mse-set` nachgemessen.
**Testfall:** TestDieWortlistenEinesSpielsWerdenGelesen,
TestDieMarkenKommenAusDemWertskript

---

## Wie eine Feststellung hier aussieht

```markdown
### <Was>

<Beobachtung.>

**Quelle:** <welche Paketdatei, welche Stelle>
**Erschlossen am:** JJJJ-MM-TT
**Sicherheit:** belegt | Vermutung
**Testfall:** <Name eines Go-Tests> | —
```

| Grad | Heißt |
| --- | --- |
| **belegt** | an mindestens zwei unabhängigen Paketen gesehen — oder im Quelltext von MSE2 nachgelesen |
| **Vermutung** | an **einem** Paket gesehen, oder gar nicht, sondern aus Verwendung oder Zweck geschlossen |

**„plausibel" gibt es seit dem 2026-10-10 nicht mehr** (F05,
#2247). Der
Grad hieß „an einem Paket gesehen, nicht gegengeprüft", und er war der bequeme
dazwischen: er sah aus wie ein Beleg und verlangte keinen Testfall, der die
Annahme benennt. Zwölf Feststellungen standen so da; nachgeprüft wurden acht
belegt — sechs davon am Quelltext, zwei an den kanonischen Sets — und vier
zur Vermutung.
`TestKeineFeststellungBleibtBeiEinemPaket` weist den Grad seither ab.

**Was nicht „belegt" ist und trotzdem in Code steht, braucht einen Testfall**,
der die Vermutung benennt — damit ein späterer Fund sie rot macht, statt sie zu
erben.

`internal/werkzeuge/formate_test.go` zählt das nach: eine Feststellung ohne die
vier Angaben, ein „Testfall", den es im Baum nicht gibt, oder eine Vermutung
ohne einen — jedes davon macht die Pipeline rot. **Damit ist diese Datei kein
Text über den Code, sondern ein Teil von ihm.**

Nur bei **belegt** darf `—` stehen. Eine Feststellung, die noch gar nicht in
Code übergegangen ist, schreibt `— (noch nicht in Code)` — und zwar in die
Angabe selbst, nicht in die Prosa daneben: eine Ausnahme, die der Wächter nicht
lesen kann, ist keine Ausnahme, sondern ein roter Lauf mit Erklärung.

Das ist keine Nachlässigkeit, sondern die Grenze: wo keine Annahme in Code
steht, braucht es keinen Wächter über sie.

---

## Paketarten

Ein MSE-„Paket" ist ein **Verzeichnis oder ein ZIP-Archiv** mit einer
Namensendung, die seine Art angibt.

| Endung | Enthält |
| --- | --- |
| `.mse-set` | ein Set: Karten und Setangaben |
| `.mse-game` | ein Spiel: Felder, Schlüsselwörter, Wörterbücher |
| `.mse-style` | einen Stil: Aussehen und Skripte für die Kartenausgabe |
| `.mse-symbol-font` | eine Symbolschrift (Mana- und andere Symbole) |
| `.mse-include` | wiederverwendbare Bestandteile für andere Pakete |
| `.mse-locale` | Übersetzungen von Feld- und Oberflächenbezeichnungen |
| `.mse-export-template` | eine Vorlage für den Export (HTML und anderes) |

**Quelle:** Verzeichnisliste von `MagicSetEditorPacks/Full-Magic-Pack`,
Verzeichnis `data/` — 528 Einträge, davon **526 Paketverzeichnisse** mit einer
dieser sieben Endungen.
**Erschlossen am:** 2026-08-08, **nachgezählt am 2026-08-30** (Nachlauf zu
#1291)
**Sicherheit:** belegt

**Hier stand „528 Einträge, sämtlich mit einer dieser sieben Endungen", und
das war eine rohe Zählung.** Drei Einträge tragen keine: `dictionaries` (ein
Verzeichnis), `packages` (eine Datei) und `.DS_Store`. Dieselbe Gestalt wie die
698 Auszählungsdimensionen, die 685 waren — **eine Zeilenzahl ist keine Zahl
fachlicher Einheiten** (§10).
**Testfall:** —

---

## Was die Anwendung mit jeder Paketart tun kann

**Gemessen am 2026-09-02**, Werkstück C von Phase 4.9
(#1324).

Bis dahin gab es dazu nur eine Zählung von **Nennungen der Endung** im
Quelltext — eine Zahl, die nichts darüber sagt, ob eine Art gelesen,
geschrieben oder bearbeitet werden kann. Hier steht die Fähigkeit, und
daneben die Bindung, über die sie läuft.

| Art | lesen | ändern | schreiben | neu anlegen |
| --- | --- | --- | --- | --- |
| `.mse-set` | `SetOeffnen`, `Karten` | `FeldSetzen`, `SetwertSetzen` | `SetSpeichern` | `SetAnlegen` |
| `.mse-game` | `Pakete`, `Paketdateien`, `PaketdateiLesen` | `PaketdateiSchreiben` | dieselbe | `PaketAnlegen` |
| `.mse-style` | dieselben | dieselbe | dieselbe | `PaketAnlegen` — auf einem Spiel |
| `.mse-include` | dieselben | dieselbe | dieselbe | `PaketAnlegen` — der Paketkopf |
| `.mse-symbol-font` | dieselben | dieselbe | dieselbe | **nein** |
| `.mse-locale` | dieselben | dieselbe | dieselbe | **nein** |
| `.mse-export-template` | dieselben | dieselbe | dieselbe | `PaketAnlegen` — auf einem Spiel |

**Vier Arten lassen sich anlegen, zwei nicht** — Stand 2026-10-07, nachgelesen
an `docs/FAEHIGKEITEN.md` („Ein Paket aus einem Gerüst anlegen"): Spiel und
Stil seit #1350,
Include und Exportvorlage seit FMT-02
(#1691). Hier
stand bis WART-01 (#2172)
für alle sechs „nein" — die Tabelle war vom 2026-09-02 und wurde beim Bau nicht
nachgezogen.

**Ändern und schreiben fallen für sechs Arten zusammen.** `PaketdateiSchreiben`
setzt eine Datei im Paket zurück und ist artunabhängig — was der Katalog führt,
lässt sich bearbeiten, und der Katalog nimmt jede der sieben Endungen
(`Paketverwalter.Installieren`).

### „Neu anlegen" ist keine Lücke

**Auch MSE2 kann es nicht.** Sein Menü legt genau zwei Dinge neu an — ein Set
(`new set`) und ein Symbol (`new symbol`); für `.mse-game`, `.mse-style` und
die übrigen gibt es dort keinen Befehl. Nachgesehen in
[`MSE2BESTAND.md`](MSE2BESTAND.md).

**Wer eine neue Vorlage will, kopiert eine vorhandene** — hier über
`PaketInstallieren`, dort über den Dateimanager.

### `.mse-symbol` ist eine Datei, kein Paket

Sie trägt keine der sieben Endungen im Sinne von `msepaket.Art`. Am Fenster
hängt sie am **Set** und nicht an einer eigenen Datei — gelesen über `Symbol`,
geschrieben über `SymbolSetzen`. Das ist eine Abweichung von MSE2 und steht als
solche in [`MSE2BESTAND.md`](MSE2BESTAND.md).

**Quelle:** die Bindungen in `internal/app` und der Katalog in
`internal/service/paketverwalter.go`; dazu [`MSE2BESTAND.md`](MSE2BESTAND.md)
für die Frage, was MSE2 selbst neu anlegt.
**Erschlossen am:** 2026-09-02, Werkstück C von Phase 4.9
**Sicherheit:** belegt für das Vorhandensein der Bindung; **nicht** dafür, dass
sie für jede Art dasselbe leistet — `Paketdateien` listet Dateien, und was der
Inhalt einer Sprachdatei darin bedeutet, weiß nur, wer sie öffnet.
**Testfall:** TestJedePaketartHatIhreZeile, TestJedeGenannteBindungGibtEs

---

## Der Paketkopf

Jede Hauptdatei beginnt mit Kopfangaben in `schlüssel: wert`-Form.

- Schlüssel sind **kleingeschrieben und dürfen Leerzeichen enthalten**
  (`short name`, `position hint`).
- Getrennt wird mit `: `.
- **Einrückung mit Tabulator** öffnet einen Unterblock.
- Derselbe Schlüssel **wiederholt sich** (`depends on:` mehrfach). Ein Leser,
  der in eine Abbildung schreibt, verliert Einträge.
- Versionen sind **Datumsangaben** (`2024-05-26`), keine Zahlenfolgen.
- Die Datei beginnt mit einer **BOM** (`U+FEFF`).

**Quelle:** `magic-m15.mse-style/style`, Zeilen 1–14;
`yugioh-series10.mse-style/style`, Zeilen 1–12 (BOM, `short name`,
`installer group`, `position hint`, Version `2017-05-01`, Tabulatoreinrückung
unter `depends on:`); `kamen-decade.mse-style/style` führt `depends on:`
**zweimal**.
**Erschlossen am:** 2026-08-08
**Sicherheit:** belegt — drei Pakete aus **drei verschiedenen Spielen** von
verschiedenen Urhebern (Magic, Yu-Gi-Oh!, Kamen Rider). Das ist der Grund, warum
gerade diese gewählt wurden: zwei Dateien aus derselben Sammlung belegen die
Sammlung, nicht das Format.
**Testfall:** TestWiederholteSchluesselGehenNichtVerloren, TestDieBOMUeberlebtDenRundlauf

---

## `depends on:` steht als Block **oder** in einer Zeile

- Als Block mit `package:` und `version:` darunter — die übliche Form.
- **In einer Zeile**: `depends on: magic.mse-game 2008-08-08`. Die Fassung
  steht hinter dem ersten Leerzeichen, **beliebig viele davor**
  (`magic.mse-game                     2009-08-10`); ohne Fassung gilt keine.
- **Eine Leerzeile oder ein Kommentar dahinter macht keinen Block** — der
  Zerteiler hängt sie an (siehe „Leerzeilen und Kommentare schließen keinen
  Block"), ein Block ist erst, was einen **Eintrag** unter sich hat.
- Ein Installer schreibt **nur** die zweite Form, mit Unterstrich
  (`depends_on:`).

**Quelle:** MSE2 — `Reader::handle(PackageDependency&)` und sein Gegenstück
(`src/util/io/package.cpp`, Zeilen 515–535, Fassung f9d9356 vom 2021-07-06),
die Fassung über `sscanf` (`Version::fromString`, `src/util/version.cpp`,
Zeilen 33–38). Datenpaket — **elf** Paketdateien in `full-magic-pack` tragen
die kurze Form, darunter `magic-crystal.mse-style` und fünf Exportvorlagen;
sechs davon mit Leerzeile dahinter.
**Erschlossen am:** 2026-09-29
**Sicherheit:** belegt — Quelltext und Zählung über alle 526 Pakete: vorher
1.904 Abhängigkeiten und 27 falsche Befunde „ohne `package`", danach 1.931
und keiner.
**Testfall:** TestAbhaengigkeitenInBeidenFormen, TestEineLeereAbhaengigkeitIstEinBefund

---

## Die Hauptdatei heißt wie die Endung ohne `mse-`

`magic.mse-game` enthält `game`, `magic-m15.mse-style` enthält `style`.

**Quelle:** `magic.mse-game/game`, `magic-m15.mse-style/style`,
`kamen.mse-game/game`, `kamen-decade.mse-style/style`,
`yugioh-series10.mse-style/style`.
**Erschlossen am:** 2026-08-08
**Sicherheit:** belegt **für `.mse-game` und `.mse-style`** — fünf Pakete, drei
Spiele, verschiedene Urheber. Für die übrigen fünf Arten geschlossen und nicht
gesehen. Der Paketleser nennt in seiner Meldung die gesuchte Datei; das ist
besser, als still eine falsche zu lesen.
**Testfall:** TestEinPaketMitEndungAberOhneHauptdatei

---

## Ein Paket ist ein Verzeichnis **oder** ein ZIP-Archiv

**Quelle:** `Full-Magic-Pack/data/` führt **526 Paketverzeichnisse** (von 527
Verzeichnissen insgesamt; das siebenundzwanzigste ist `dictionaries`); die
Endungen sind dieselben, die MSE für einzelne Dateien verwendet. Nachgezählt am
2026-08-30.
**Erschlossen am:** 2026-08-08
**Sicherheit:** Vermutung — dass ein Set als Archiv vorliegt, ist aus der
Sammlung nicht belegt: dort liegen keine Sets.
**Testfall:** TestVerzeichnisUndArchivGebenDasselbeHer

---

## Ein `.mse-installer` ist ein ZIP mit `installer` und je Paket einem Ordner

- Jede Datei eines Pakets liegt unter `<paket>/<datei>`, etwa
  `herospath.mse-game/stats/rarity.png` — **ohne** Ordnereinträge.
- **Die Hauptdatei `installer` steht zuletzt** und beginnt mit einer BOM.
  MSE2 schreibt ihre Schlüssel **mit Unterstrich** (`mse_version: 0.3.7`,
  `short_name:`), den Kopf leer bis auf `position_hint: 100000`,
  `version: 0.0.0` und `compatible_version: 0.0.0`.
- **Je Paket ein Block `package:`** mit `name`, `version`, `short_name`,
  `full_name`, `icon_url`, `installer_group`, `position_hint`,
  `description` und je Abhängigkeit `depends_on: paket fassung`.
- **Ein leerer Kurzname wird der Dateiname ohne Endung**, ein leerer
  voller Name der kurze; eine leere Gruppe wird aus dem Namen gebildet —
  bis zum letzten Bindestrich, Bindestriche als Schrägstriche, darunter
  „Game files" oder der Kurzname (`unclassified/herospath/herospath-watermarks`).
- **Was nicht mitkommt**: Unterordner mit Punkt vorn und `Thumbs.db`.

**Quelle:** MSE2 — `Installer::addPackage`, `PackageDescription` und ihre
Reflexion (`src/data/installer.cpp`, Zeilen 172–244), `Package::openSubdir`
(`src/util/io/package.cpp`, Zeilen 372–392), `ignore_file`
(`src/util/file_utils.cpp`, Zeilen 33–37), alle Fassung f9d9356 vom
2021-07-06. **Und ein Lauf des Orakels** am 2026-09-29:
`magicseteditor --create-installer herospath-watermarks.mse-include
herospath.mse-game`.
**Erschlossen am:** 2026-09-29
**Sicherheit:** belegt — dieses Programm schreibt für dieselben zwei Pakete
**Byte für Byte** dieselbe Hauptdatei und dieselben 35 Einträge.
**Testfall:** TestEinInstallerTraegtDiePaketeWieMSE2, TestEinInstallerAusMSE2WirdGelesen

---

## `include_file:` und `include file:` sind derselbe Verweis

MSE2 macht **jeden** Schlüssel kanonisch, bevor es ihn vergleicht
(`key = canonical_name_form(trim(key))`), und fragt dann nach
`include_file`. Ein Stil, der `include_file: /magic-modules.mse-include/scripts`
schreibt, bindet dort genauso ein wie einer mit Leerzeichen. Das Pack schreibt
durchweg die Form mit Leerzeichen (2.781 Zeilen, keine mit Unterstrich) —
die andere kommt aus Händen, nicht aus MSE2s Schreiber, denn der schreibt
keine Verweise.

**Quelle:** MSE2, `src/util/io/reader.cpp` Zeile 217 und
`src/util/io/reader.hpp` Zeile 175, Fassung f9d9356 vom 2021-07-06; gezählt
in `full-magic-pack/data` am 2026-10-01. Gefunden von `FuzzEinbinden`
(QA-06b, #1967).
**Erschlossen am:** 2026-10-01
**Sicherheit:** belegt
**Testfall:** TestEinVerweisInKanonischerSchreibweiseWirdEingebunden

---

## Jeder Schlüssel des Formats ist kanonisch — `z_index:` ist `z index:`

Was für `include_file:` gilt, gilt für jeden Schlüssel: MSE2 vergleicht ihn
nach `canonical_name_form`. Ein Feldstil mit `z_index:`, `render_style:` oder
`always_symbol:` und eine Schrift mit `italic_name:` oder `scale_down_to:`
lesen sich dort wie mit Leerzeichen. Das Pack schreibt durchweg Leerzeichen
(7.780-mal `z index:`, keinmal `z_index:`, gezählt am 2026-10-01).

**Was das nicht sagt**: dass ein **Name**, der als Schlüssel steht —
`card_name:` unter `card style:`, eine Wahl unter `choice images:` —, ebenso
gelesen wird. Das ist eine eigene Frage, und sie steht im nächsten Abschnitt.

**Quelle:** MSE2, `src/util/io/reader.cpp` Zeile 217, Fassung f9d9356 vom
2021-07-06; #1972.
**Erschlossen am:** 2026-10-01
**Sicherheit:** belegt
**Testfall:** TestEinFeldstilLiestSeineSchluesselKanonisch

---

## `show statistics:` und `choice colors:` — was ein Feld der Statistik mitgibt

Jedes Kartenfeld wird in MSE eine **eigene Dimension der Statistik**. Sie
steht vor den erklärten (`statistics dimension:`), solange das Feld es nicht
mit `show statistics: false` abschaltet.
- **Ohne Angabe gilt ja**, bei einem Bildfeld nein.
- Ein Auswahlfeld zählt nach seinen Wahlen der obersten Ebene und rechnet `primary_choice(card.<feld>)`.
- Seine Balken färbt `choice colors:`, Wahl auf Farbe.

Gemessen am 2026-10-01 über `full-magic-pack/data`:
- `magic.mse-game` schaltet es an allen 249 Kartenfeldern ab.
- `archenemy`, `herospath`, `planechase`, `saga`, `space` und `vanguard` bekommen je zwei bis fünf automatische Dimensionen.
- `choice colors:` steht in fünf Spielen, `choice colors cardlist:` nirgends.

**Quelle:** MSE2 (Fassung f9d9356 vom 2021-07-06):
- `src/data/field.cpp` Zeilen 28 und 55;
- `src/data/field/image.hpp` Zeile 27;
- `src/data/field/choice.cpp` Zeile 37;
- `src/data/statistics.cpp` Zeilen 27–56;
- `src/data/game.cpp` Zeilen 68–77.

Siehe #1985.
**Erschlossen am:** 2026-10-01
**Sicherheit:** belegt
**Testfall:** TestEinFeldSagtObUndWieEsInDieStatistikGeht, TestEineAutomatischeDimensionZaehltGefaerbt

---

## Ein Name als Schlüssel ist ebenso kanonisch — `rule_text:` trifft `rule text`

MSE2 macht **beide** Seiten kanonisch: den Feldnamen in `Field::after_reading`
und jeden Schlüssel beim Lesen. Deshalb trifft `card style: rule_text:` das Feld
`rule text`, und `styling style: rahmen_zeigen:` die Stiloption `rahmen zeigen`.
Ebenso kanonisch schlägt MSE nach:
- Wahlbilder (`choice images:`);
- Sprachpakete (`SubLocale::tr`).

Gemessen am 2026-10-01 über `full-magic-pack/data`:

| Was | gezählt | davon mit Unterstrich |
| --- | --- | --- |
| Feldnamen (`name:` eines Felds) | 8.535 | 3: `fuse_bridge` (zweimal, `save value: false`) und `mainframe_rarity_name` (eine Stiloption) |
| Schlüssel unter einem Stilblock | 8.925 | 6, darunter `rule_text` (`magic-baseball-1980-topps`) sowie `flavor_text`, `level_11_text` und `level_12_text` (`magic-m15-secret-lair-gatewatch`) |
| Schlüssel unter `choice images:` | 6.980 | 137 |

**Eine Ausnahme in MSE2 selbst**: die Farbtafel einer Statistik (`colors:`).
- MSE2 liest ihre Schlüssel kanonisch.
- `GraphData::GraphData` schlägt aber mit dem **rohen** Gruppennamen nach. Eine Gruppe `multi color` findet dort nie ihre Farbe.
- Hier wird auf beiden Seiten kanonisch verglichen. Im Pack trägt keine Statistikdimension eine eigene `colors:`-Tafel.

**Wo es hier ankommt**:
- Der **Stil** bekommt an der Naht zum Spiel die erklärten Namen (`service.mitErklaertenNamen`). Danach trägt jeder rohe Vergleich, auch die Feldfläche, über die das Fenster ein Feld auf der Karte findet.
- Der Feldname selbst bleibt, wie er erklärt ist: er ist der Schlüssel der Kartenwerte im Bestand.

**Quelle:** MSE2 (Fassung f9d9356 vom 2021-07-06):
- `src/data/field.cpp` Zeile 68;
- `src/util/io/reader.cpp` Zeile 217;
- `src/gui/value/choice.cpp` Zeilen 50, 156, 174 und 188;
- `src/render/value/multiple_choice.cpp` Zeile 59;
- `src/data/locale.cpp` Zeilen 82 und 90;
- `src/gui/control/graph.cpp` Zeile 138.

Siehe #1980.
**Erschlossen am:** 2026-10-01
**Sicherheit:** belegt
**Testfall:** TestEinFeldstilTraegtDenErklaertenNamen, TestEinOptionsstilTraegtDenErklaertenNamen, TestEinSprachpaketTrifftBeideSchreibweisen, TestEineFarbtafelTrifftBeideSchreibweisen

---

## Ein Bildwert behält seinen Leerraum am Ende

`image: image1 ` nennt die Datei `image1 `: MSE2 schneidet einen Wert nur
**links** (`value = trim_left(…)`), und der Name geht so an `openIn`. Der
Verweis auf das Set-Paket (`msepaket.SetpaketVerweis`) trägt ihn deshalb
unverändert und schneidet beim Zerlegen ebenfalls nur links.

**Quelle:** MSE2, `src/util/io/reader.cpp` Zeile 224, Fassung f9d9356 vom
2021-07-06. Gefunden von `FuzzPaketpfad` (QA-06b, #1967).
**Erschlossen am:** 2026-10-01
**Sicherheit:** belegt
**Testfall:** TestEinSetbildMitLeerraumAmEndeIstSeineDatei

---

## Leerzeilen und Kommentare schließen keinen Block

Eine Leerzeile trägt keine Einrückung, steht aber **zwischen** zwei eingerückten
Einträgen. Wer ihre Einrückung als Ebene nimmt, schließt jeden offenen Block.
Dasselbe gilt für eine Zeile, die mit `#` beginnt.

Sie hängen am **tiefsten offenen** Knoten. Ein Kommentar hinter einem Feld wird
damit dessen Kind, nicht das der Karte — das war beim Schreiben des Importeurs
nicht offensichtlich und steht deshalb hier.

**Quelle:** `magic-m15.mse-style/style` — `init script:` enthält Leerzeilen
zwischen eingerückten Zuweisungen und die Trennzeile `##### Extra scripts`.
**Erschlossen am:** 2026-08-08
**Sicherheit:** Vermutung für `#` als Kommentarzeichen auf Baumebene.
**Testfall:** TestLeerzeileUndKommentarSchliessenKeinenBlock

---

## Die letzte Zeile darf ohne Zeilenende enden

**Quelle:** keine — nicht beobachtet, angenommen. Nicht jede Datei endet mit
einem Umbruch, und eine, die es nicht tut, wäre deswegen nicht kaputt.
**Erschlossen am:** 2026-08-08
**Sicherheit:** Vermutung
**Testfall:** TestDieLetzteZeileDarfOhneZeilenendeEnden

---

## Drei Zeilenenden, und nur UTF-8

**Eine Zeile endet an `\n`, an `\r\n` und an einem einzelnen `\r`.** So liest
MSE2 selbst: `read_utf8_line` bricht an allen dreien — ein `\r`, dem kein `\n`
folgt, beendet die Zeile, und das nächste Zeichen gehört zur nächsten. Hier
brach der Zerteiler bis FMT-06 nur an `\n`, und eine Datei mit den
Zeilenenden des alten Mac OS war eine einzige Zeile. **Der Rundlauf hält**:
jede Zeile schreibt ihr Ende so zurück, wie sie es gelesen hat.

**Eine Datei ist UTF-8, und nichts sonst.** MSE2 wirft bei einer ungültigen
Folge `Invalid UTF-8 sequence on line N` und liest nicht weiter; dieses
Programm nennt dieselbe Stelle samt Spalte (`ErrKodierung`). Eine Datei in
UTF-16 — erkannt an ihrer Byte-Reihenfolge-Marke — wird ebenso benannt, statt
ihre Nullbytes als Inhalt zu lesen. Geraten wird keine Kodierung.

**Im vollen Pack kommt keines davon vor** (2026-09-23, 807 Textdateien): kein
einzelnes `\r`, kein UTF-16, und ungültiges UTF-8 nur in zwei Wortlisten einer
Exportvorlage (`magic-namecheck.mse-export-template/*.txt`), die nicht durch
den Zerteiler gehen. Die zehn kanonischen Sätze sind gültiges UTF-8.

**Quelle:** MSE2, `src/util/io/reader.cpp`, Zeilen 159–186 und 188–195,
Fassung f9d9356 vom 2021-07-06 — nachgelesen, nicht erschlossen.
**Erschlossen am:** 2026-09-23 (FMT-06, #1695)
**Sicherheit:** belegt
**Testfall:** TestEinEinzelnesWagenruecklaufBeendetEineZeile, TestUngueltigesUTF8WirdMitZeileBenannt, TestUTF16WirdAlsSolchesBenannt

---

## Wie weit ein Paket reicht

**Ein Paket kommt von außen, und was darin steht, wird nicht geglaubt.**
Die Grenzen stehen als benannte Konstanten in `internal/msepaket/paketgrenzen.go`,
jede **gemessen** am vollen Pack und an den kanonischen Sätzen:

| Grenze | Wert | größtes Echtes |
| --- | ---: | --- |
| eine Datei eines Pakets, ausgepackt | 256 MiB | 10,7 MB (ein Wörterbuch) |
| ein Bild, je Seite | 32.768 Bildpunkte | 4.157 |
| ein Bild, zusammen | 100 Millionen Bildpunkte | 12,4 Millionen |
| Einbindungen je Lauf | 1.000 | 87 (`magic.mse-game`) |
| Tiefe der Einbindungen | 16 | 5 (`magic-m15-meld-3in1.mse-style`) |
| eine Schriftdatei | 256 MiB | 192 MB (`Apple Color Emoji.ttc`), im Pack 1,6 MB |
| Einträge eines Archivpakets | 100.000 | 3.838 (`Black.mse-set`); der ganze Pack als ein Archiv hätte 34.683 |
| Verzeichnis eines Archivpakets | 16 MiB | 210 KB (`Black.mse-set`); der ganze Pack als ein Archiv etwa 3,5 MB |

**Der Kopf eines Zip-Eintrags wird nicht geglaubt**: eine Größe darüber weist
ab, eine darunter teilt keinen Speicher zu — gelesen wird, was wirklich kommt.
Ein Bild wird **vermessen, bevor es ausgepackt wird**, an genau einer Stelle
(`msepaket.DekodiereBild`, ein Wächter hält die Stelle). Ein Paketverzeichnis
wird als `os.Root` geöffnet: ein Symlink, der hinausführt, wird nicht gelesen,
einer, der im Paket bleibt, schon. Einen Archivnamen wie `../../x` bereinigt
`archive/zip` in seiner Dateisicht selbst zu `x`; er liegt damit im Paket.

**Das Verzeichnis eines Archivs wird nicht geglaubt**
(#1886):
`archive/zip` liest es **ganz**, bevor es etwas zurückgibt — Kopf um Kopf,
bis einer nicht passt, und die Zahl aus dem Ende bestimmt vorher die Kapazität
der Liste (Go 1.27.1, `src/archive/zip/reader.go`, `(*Reader).init`). Deshalb
liest `msepaket` zuerst das Ende (`LiesArchivende`, dieselbe Stelle wie die
Dateiansicht) und weist ab, was mehr Einträge oder ein größeres Verzeichnis
nennt; und beim Lesen zählt ein Leser mit und bricht ab, sobald das
Verzeichnis **länger ist als angegeben**. Das gilt beim Öffnen, beim
Schreiben ins Archiv und beim Installer. **Die Pakete des vollen Packs sind
Verzeichnisse** und damit nicht betroffen; geprüft ist die Regel an den zehn
kanonischen Sets, und keines fällt darunter (Lauf hinter `fremdpakete`).

**Was über eine Grenze geht, ist ein Befund oder ein Fehler mit Namen** — bei
Einbindungen ein Befund, und das Paket öffnet mit dem, was davor lag.

**Und dieselbe Grenze gilt an jedem Weg, der eine Paketdatei ganz liest** —
nicht nur beim Öffnen. Bis zum Zweitblick über Phase 6.1 (Befund 4) lasen der
Schreibweg (die Datei, die er ersetzen soll), die Symbolzeichnung und die
`game`-Datei beim Anlegen eines Stils ohne Grenze; gemessen waren 256, 572 und
768 MiB Zuteilung, bevor ein Konflikt kam. Jetzt gehen sie über
`msepaket.LeseDateiBegrenzt`, und ein Wächter
(`internal/werkzeuge/lesebudget_test.go`) verlangt von **jedem** ganzen Lesen
im Produktivcode eine Grenze oder einen Grund, der sagt, wessen Datei es ist.

**Quelle:** gemessen am vollen Pack und an den zehn kanonischen Sätzen, nicht
nachgelesen; MSE2 kennt keine dieser Grenzen.
**Erschlossen am:** 2026-09-23 (FMT-06)
**Sicherheit:** belegt
**Testfall:** TestEineAufgeblaseneDateiImArchivWirdAbgewiesen, TestDasSchreibenInsArchivLiestKeineAufgeblaseneDatei, TestDasSchreibenInsVerzeichnisLiestKeineUebergrosseDatei, TestEineUebergrosseSymbolzeichnungWirdAbgewiesen, TestEinUebergrossesSpielWirdNichtGanzGelesen, TestJedesGanzeLesenIstBegrenztOderBegruendet, TestEinZipKopfWirdNichtGeglaubt, TestEinBildWirdVermessenBevorEsAusgepacktWird, TestEineRauteImEinbindenBleibtImBudget, TestEineLangeKetteEndetAnDerTiefengrenze, TestEinPaketverzeichnisFolgtKeinemSymlinkHinaus, TestEinArchivpfadFuehrtNichtHinaus, TestEineUebergrosseSchriftWirdNichtGelesen, TestEinArchivUeberDerGrenzeWirdBenanntAbgewiesen, TestEinVerzeichnisLaengerAlsAngegebenWirdAbgewiesen, TestKeinVorhandenesPaketFaelltUnterDieRessourcenregel

---

## Was beim Zurückschreiben normalisiert wird

**Was niemand angefasst hat, geht Byte für Byte hinaus** — Einrückung,
Abstände, Kommentare, Leerzeilen, das BOM, das Zeilenende **jeder** Zeile und
jeder Block, den dieses Programm nicht versteht. Eine Zeile trägt ihre Rohform
bei sich ([0002](entscheidungen/0002-sqlite-als-arbeitsformat.md), #48).

**Eine geänderte Zeile verliert ihre Rohform** und wird gesetzt, wie MSE2
jede Zeile setzt: Einzug mit einem Tab je Ebene, `schlüssel: wert` mit genau
einem Leerzeichen. **Ihr Zeilenende behält sie** — `\r\n` bleibt `\r\n`. Ein
Wert **mit Umbruch oder mit führendem Leerraum** steht als Block eine Ebene
tiefer unter seinem Schlüssel: hinter dem Doppelpunkt schnitte das Lesen den
Leerraum ab, bei MSE2 wie hier. Bis FMT-06 stand ein Wert mit führendem
Leerzeichen hinter dem Doppelpunkt und kam beim Wiederlesen ohne es zurück.

**Was hinzukommt**, bekommt ein Zeilenende nach seinem Vorbild: ein neues
Feld steht hinter der Zeile seines Vorgängers im Bestand (sonst am Anfang des
Kartenblocks), eine neue Karte hinter der letzten Karte mit deren Zeilenende;
was ohne Vorbild entsteht, endet auf `\n`.

**Was wegfällt, nimmt seinen Unterblock mit** — auch einen Kommentar
**unter** der Zeile. Ein Kommentar **darüber** bleibt: der Zerteiler hängt ihn
an die Zeile davor.

**Wieder gelesen ändert sich nichts mehr**: ein zweiter Rundlauf über das
Geschriebene gibt dieselben Bytes. Bilder und Symbole neben der Hauptdatei
gehen mit; was fehlt, wird als Befund genannt.

**MSE2 selbst tut weniger**: es schreibt jede Datei neu aus seinem
Datenmodell, immer mit BOM und `\n`, und Kommentare liest es gar nicht erst
(`src/util/io/reader.cpp`, Zeilen 202–205). Die Form einer gesetzten Zeile
ist seine; dass der Rest bleibt, ist die Entscheidung dieses Programms.

**Quelle:** MSE2, `src/util/io/writer.cpp`, Zeilen 69–104, und
`src/util/io/reader.cpp`, Zeile 223, Fassung f9d9356 vom 2021-07-06 —
nachgelesen; die Rohzeilen sind dieses Programms eigene Regel.
**Erschlossen am:** 2026-09-23 (FMT-06, #1695)
**Sicherheit:** belegt
**Testfall:** TestEinGeaendertesSetBehaeltWasEsNichtVersteht, TestEinWertMitFuehrendemLeerraumUebersteht, TestEinMehrzeiligerWertKommtZurueck, TestEinDurchgereichtesSetIstByteFuerByteDasselbe, TestEinEingesetztesBildKommtMit

---

## Beide Programme lesen dieselbe Datei gleich — gemessen am Orakel

Der Setdatei-Rundlauf fährt zwei Richtungen über eine Kohorte und fragt dabei
das gebaute MSE2 nach **jedem Wert jedes Felds jeder Karte**, die gerechneten
eingeschlossen:

1. **Diese Anwendung schreibt, MSE2 liest**: einlesen, den Namen der ersten
   Karte auf einen Unicode-Text setzen, eine setweite Stiloption setzen,
   sichern. MSE2 liest Original und Abschrift gleich — bis auf die beiden
   Änderungen, und die stehen so da, wie sie gesetzt wurden —, und es warnt
   über die Abschrift nichts, was es über das Original nicht schon warnte.
2. **MSE2 schreibt, diese Anwendung liest**: MSE2 schreibt dasselbe Set mit
   `write_set_file`, und diese Anwendung liest jeden Wert so, wie MSE2 ihn aus
   derselben Datei liest.

| Set | Karten | Werte | davon Unicode | von MSE2 gerechnet | Dateien neben `set` |
| --- | ---: | ---: | ---: | ---: | ---: |
| eigene Saat (`magic`) | 2 | 11 | 3 | 5 | 1 |
| eigene Saat (`saga`) | 2 | 8 | 2 | 30 | 1 |
| Archenemy | 70 | 1.260 | 45 | 0 | 70 |
| Planechase | 86 | 1.626 | 42 | 8 | 86 |
| Vanguard | 107 | 1.926 | 51 | 0 | 107 |
| White | 3.761 | 68.591 | 2.955 | 130.821 | 3.801 |

**In keiner Richtung eine Abweichung dieser Anwendung**, kein verlorener
Schlüssel (oberste und zweite Ebene), keine fehlende oder veränderte Datei,
keine neue Warnung — und die gesetzte Stiloption liest MSE2 in allen sechs.
Der Beleg samt Eingaben steht in `docs/belege/setrundlauf.md`.

**Quelle:** `internal/flaechen/setrundlauf_fremd_test.go` gegen das Orakel
(`magicseteditor --cli`, Fassung f9d9356 vom 2021-07-06, gebaut nach 0026),
die Pakete aus `full-magic-pack/data`; vier kanonische Sets
(`archive.org/details/mse-canon-cards-2022-10-24`) und zwei eigene, die der
Lauf selbst schreibt. Gemessen am 2026-10-10 (F05, #2247).
**Erschlossen am:** 2026-10-10
**Sicherheit:** belegt — sechs Sets, zwei Spiele, beide Richtungen.
**Testfall:** TestSetrundlaufMitMSE2

---

## MSE2 speichert einen Wert **nach** seinem Feldskript

Was MSE2 schreibt, ist nicht, was es las: aus `"…"` wird `“…”`, aus `Scheme`
wird `<word-list-type-en>Scheme</word-list-type-en>`, aus `[1][W]` wird `1W`,
und ein Wort, das sein Wörterbuch nicht kennt, bekommt eine Unterringelung
(nächste Feststellung). Gezählt über die Kohorte oben: **330** solcher Werte
in Archenemy, **503** in Planechase, **434** in Vanguard, **24.293** in White.

Diese Anwendung speichert, was in der Datei steht, und rechnet das Skript
beim Zeichnen. **Für das Bild ist es dasselbe**; für den gespeicherten Wert
nicht — eine Datei, die MSE2 einmal gespeichert hat, unterscheidet sich vom
Original auch ohne jede Änderung.

**Quelle:** der Rundlauf oben; MSE2 rechnet `script:` eines Felds bei jeder
Änderung und legt das Ergebnis ab (`TextValue::update` ruft
`field().script.invokeOn(ctx, value)`, `src/data/field/text.cpp`, Zeilen
174–183, Fassung f9d9356 vom 2021-07-06).
**Erschlossen am:** 2026-10-10
**Sicherheit:** belegt — an vier kanonischen Sets und zwei Spielen gezählt.
**Testfall:** —

---

## MSE2 schreibt seine Rechtschreibmarken in die Datei — sie sind kein Wert

`<error-spelling:en_US>Ein</error-spelling:en_US>`, mit Zusatzwörterbuch
`<error-spelling:en_US:/magic.mse-game/dictionary/magic-words>`. MSE2 nimmt sie
vor jeder neuen Prüfung heraus (`check_spelling`,
`src/script/functions/spelling.cpp`, Zeilen 57–71), und die Vorlagen nehmen
sie vor dem Zeichnen heraus (`remove_tag@(tag: "<error-spelling")`).

**Diese Anwendung liest sie nicht** — sonst fand der Volltext „spelling",
„dictionary" und „magic" in jeder Karte, die MSE2 unterringelt hatte. Ein MSE2,
das die Abschrift öffnet, prüft neu und setzt sie wieder.

**Quelle:** der Rundlauf oben, an beiden eigenen Saaten (deutsche Wörter gegen
das englische Wörterbuch); MSE2 `src/script/functions/spelling.cpp`, Fassung
f9d9356 vom 2021-07-06.
**Erschlossen am:** 2026-10-10
**Sicherheit:** belegt — zwei Spiele, dazu der Quelltext.
**Testfall:** TestRechtschreibmarkenAusMSE2WerdenNichtGelesen

---

## Setweite Stiloptionen stehen im Block `styling:`, nie unter `set info:`

```
styling:
	magic-m15-altered:
		text_box_mana_symbols: magic-mana-small.mse-symbol-font
		chop_top: 7
```

Je Stilpaket, unter seinem Namen **mit** Spiel davor (`Set::stylingDataFor`,
`src/data/set.cpp`, Zeilen 99–101). MSE2 liest sie dort als `styling.chop_top`.

**Bis zum 2026-10-10 schrieb diese Anwendung beides falsch**: eine hier
gesetzte Option kam gar nicht in die Datei, und der eingelesene Block stand
zusätzlich als Zeile unter `set info:`, wo MSE2 ihn mit „Unexpected key:
'styling'" abwies — jedes Set, das diese Anwendung schrieb, öffnete in MSE2
mit einer Warnung. Eine solche Zeile aus einer früheren Abschrift wird beim
Einlesen übergangen und beim Schreiben entfernt — erkannt an ihrer flachen
Gestalt `paket/option: wert`, die MSE2 nie schreibt.

**Quelle:** MSE2 `src/data/set.cpp`, Fassung f9d9356 vom 2021-07-06; alle
zehn kanonischen Sets führen den Block (nachgesehen am 2026-08-20); der
Rundlauf oben fragt `styling.<option>` an sechs Sets ab.
**Erschlossen am:** 2026-10-10
**Sicherheit:** belegt — Quelltext, zehn kanonische Sets und das Orakel.
**Testfall:** TestSetweiteStiloptionenStehenImBlockStyling, TestSichernSchreibtDieSetweitenStiloptionen, TestEineVerirrteStilzeileWirdNichtEingelesen

**Geschrieben wird unter dem Namen des aufgelösten Pakets und nur dort**
(Nachlauf über Phase G, zweiter Teil, Befund 1): ein Block unter einem
anderen Namen — `magic-magic-m15`, wie ihn dieser Schreiber vor F05 anlegte —
bleibt, wie er steht. **Und gelesen gilt eine Reihenfolge**: Vorgabe des
Stils, Block der Datei, was hier gesetzt wurde, die Karte — für die
Optionsliste wie für die Rasterung. Bis zum 2026-10-11 las die Rasterung nur
den Block und die Liste nur das hier Gesetzte; eine gesetzte Option änderte
die Karte nicht, Byte für Byte gemessen.
**Testfall:** TestEinAufgeloesterStilSchreibtNurInSeinenBlock, TestHierGesetztLiegtUeberDemBlockDerDatei, TestEineGesetzteStiloptionAendertDieKarte, TestDieListeNenntDenWertAusDerDatei, TestEineStiloptionWirktNachSichernUndOeffnen

---

## MSE2 schreibt unbekannte Schlüssel nicht mit — diese Anwendung schon

Ein Schlüssel, den das Spiel nicht erklärt, liest MSE2 mit einer Warnung und
lässt ihn beim Schreiben weg (`eigener setschluessel`, `eigener
kartenschluessel` in beiden eigenen Saaten). Diese Anwendung schreibt ihn
zurück — er ist Teil der Datei, auch wenn niemand ihn deutet.

**Quelle:** der Rundlauf oben, an beiden eigenen Saaten; MSE2 meldet ihn über
`Reader::unknownKey` (`src/util/io/reader.hpp`, Zeilen 173–186, Fassung
f9d9356 vom 2021-07-06).
**Erschlossen am:** 2026-10-10
**Sicherheit:** belegt — zwei Spiele, dazu der Quelltext.
**Testfall:** TestSetrundlaufMitMSE2

---

## Kartenfelder stehen als `card field:`-Blöcke im Spielpaket

Jeder Block trägt `type:` und `name:`; eine Auswahl nennt ihre Möglichkeiten als
wiederholtes `choice:`, ein Standardwert steht als `default:`.

Beobachtete Werte für `type`: `text`, `choice`, `image`, `color`, `boolean`,
`info`.

**Quelle:** `magic.mse-game/card_fields` (rund 104 KB).
**Erschlossen am:** 2026-08-08
**Sicherheit:** Vermutung für die **Vollständigkeit** der Artenliste — geprüft
wurde nicht, ob es weitere gibt. Eine unbekannte Art wird deshalb als Text
behandelt statt abgewiesen.
**Testfall:** TestEineUnbekannteFeldartWirdText

---

## Eine mehrseitige Karte ist **eine** Karte mit numerierten Feldreihen

Vorder- und Rückseite stehen **in derselben Karte**: die zweite Seite trägt
dieselben Feldnamen mit dem Zusatz ` 2` — `name` und `name 2`, `casting cost`
und `casting cost 2`. Eine Beziehung zwischen zwei Kartenblöcken gibt es
nicht.

**Die Regel, an der eine Reihe erkannt wird**: ein Feld `F N` gehört zur Reihe
von `F`, **wenn es das Feld `F` gibt**. Ohne die zweite Hälfte wäre `loyalty
cost 1` bis `loyalty cost 16` eine sechzehnstellige Reihe — es sind die Zeilen
eines Planeswalkers, und ein Feld `loyalty cost` gibt es nicht.

Gezählt an `magic.mse-game` über den Leser dieses Projekts: 249 Kartenfelder,
**43 Reihen**, höchste Stelle **8**, und **40** Felder mit Zahl ohne
Grundfeld.

**Was eine Stelle bedeutet, entscheidet das Stilpaket** und nicht der
Feldname: `magic-m15-mainframe-dfc.mse-style` zeichnet `name`/`name 2` als
Vorder- und Rückseite, `magic-m15-flip.mse-style` als oben und unten, und
`additional credit 2` bis `additional credit 5` sind fünf Zeilen einer
Danksagung.

**Quelle:** MSE2 (`src/data/card.hpp`, Fassung `f9d9356` vom 2021-07-06 — `Card`
führt keinen Bezug auf eine andere Karte), `magic.mse-game/card_fields`,
`magic-m15-mainframe-dfc.mse-style/style` und `fenhl/magic-set-generator`
(`src/mse.rs`, Makro `push_alt!`).
**Erschlossen am:** 2026-08-27
**Sicherheit:** belegt — drei unabhängige Quellen, zwei davon fremder Code
**Testfall:** TestEinNumeriertesGeschwisterBildetEineReihe

---

## Ein Feld der Art `image` trägt einen Dateinamen **im Set-Paket**

Der Wert einer Karte ist dort kein Pfad und kein Paketpfad, sondern der Name
einer Datei im Set selbst — MSE öffnet sie über das **lokale** Paket:

```cpp
if (!value().filename.empty()) {
  auto image_file = getLocalPackage().openIn(value().filename);
  if (image_load_file(image, *image_file)) {
    image.Rescale(w, h);
  }
}
// nice placeholder
if (!image.Ok() && style().default_image.isReady()) { … }
```

**Drei Dinge stehen darin.** Der Name ist **lokal** zum Set; das Bild wird
**gestreckt** (`Rescale`, kein `ASPECT_FIT` — anders als bei einem Wahlbild,
siehe `render style:`); und das `default:` des Feldstils gilt **nur**, wenn kein
eigenes Bild da ist.

Für dieses Programm heißt das: der Ort des Sets muss bis zur Bildquelle reisen.
Ein Stilpaket wird über die Paketwurzeln **gesucht**, ein Set liegt dort, wo
jemand es abgelegt hat — deshalb trägt der Verweis den ganzen Ort
(`msepaket.SetpaketVerweis`) und nicht den Paketnamen.

**Quelle:** MSE2, `src/render/value/image.cpp`, Zeilen 30–46, Fassung f9d9356
vom 2021-07-06.
**Erschlossen am:** 2026-08-17
**Sicherheit:** belegt
**Testfall:** TestDasBildDerKarteStehtInDerVorschau

---

## Eine Symbolschrift führt `symbol:`-Blöcke mit `code` und `image`

`code` ist das Zeichen, wie es im Kartentext steht (`W`, `U`); `image` der
Paketpfad des Bildes.

**Quelle:** `magic-mana-large.mse-symbol-font/symbol-font` aus `M15-Magic-Pack`,
gelesen am 2026-08-16 — 167 `symbol:`-Blöcke.
**Erschlossen am:** 2026-08-08, **belegt am 2026-08-16**
**Sicherheit:** belegt
**Testfall:** TestSymbolmodellLiestZeichenUndBilder

**Bis zum 2026-08-16 stand hier „Vermutung — die Hauptdatei einer Symbolschrift
ist nicht gelesen worden".** Sie ist es seit
#275:
`Symbolmodell` liest `symbol:`-Blöcke aus dem Paketkopf. Die Angabe blieb
trotzdem stehen — der Zweitblick über #274/#275/#277/#278 hat sie angestrichen.

**Und `code` und `image` sind nicht alles.** Dieselbe Datei führt denselben Code
mehrfach — `T` und `Q` je dreimal, `V/W`, `W/U`, `W/B`, `V/U` je zweimal — und
unterscheidet die Fassungen über 54 `enabled:`-Bedingungen mit skriptberechnetem
`image:`. Unser Modell nimmt davon nur `code` und `image`, und bei doppeltem
Code gewinnt der **erste** Eintrag. Was das kostet, steht in
#293; hier
steht es, damit niemand die Feststellung oben für vollständig hält.

---

## Eine Symbolschrift zerlegt einen Text in **Dateireihenfolge**, nicht nach Länge

Hier stand bis zum 2026-08-17 „der längste Code zuerst", als **plausibel**
geführt und aus den Daten erschlossen. Der Blick in MSE2 hat die Regel
widerlegt: `SymbolFont::split` geht die Symbole in der Reihenfolge durch, in
der die Datei sie führt, und nimmt den **ersten**, dessen Bedingung gilt und
dessen Code am Anfang des Restes steht.

```cpp
FOR_EACH_CONST(sym, symbols) {
  if (!sym->code.empty() && sym->enabled) {
    …
    if (is_substr(text, pos, sym->code)) { … pos += sym->code.size(); goto next_symbol; }
```

Trifft an einer Stelle **kein** Code, rückt MSE um ein Zeichen weiter
(`pos += 1`) und zeichnet dort nichts.

Dass die alte Regel nie auffiel, hat einen Grund: in `magic-mana-large` stehen
die zusammengesetzten Codes vorn — `V/W` in Zeile 21, die einstelligen `W U B
R G` erst ab Zeile 779. Für dieses Paket liefern beide Regeln dasselbe. Für
eine Schrift, die es andersherum hält, nicht.

**Die Fassungen trennt `enabled:`.** Derselbe Code steht mehrfach da — `T` und
`Q` je dreimal —, und die Bedingung entscheidet, welche gilt:

```
symbol:
	code: T
	image: mana_t_older.png
	enabled: { mana_t() == "older" }
```

`enabled` ist ein `Scriptable<bool>` mit dem Vorgabewert `true`: ohne Angabe
gilt die Fassung, und eine Bedingung, die sich nicht ausrechnen lässt, lässt
den alten Wert stehen — sie gilt also ebenfalls.

**Ein Code darf ein regulärer Ausdruck sein** (`regex: yes`), und dann wird er
als solcher geprüft. `magic-mana-large` nutzt das 32-mal, im ganzen Pack sind
es 73 verschiedene Muster — mit **Vorausschau**: `10(?!/)`,
`[0-9](?!/[WUBRGCSTQ2])|.`.

**Ohne sie fehlten die Ziffern.** In `magic-mana-large` hängen die
Ziffernsymbole an `enabled: {ancestral_mana()}`; ist das falsch, deckt sie
allein der Auffangausdruck am Dateiende. Ein `1` in den Manakosten hat dann
kein eigenes Symbol mehr.

Gos `regexp` kennt keine Vorausschau (RE2 lässt sie zugunsten linearer
Laufzeit weg). **Alle 73 Muster haben dieselbe Bauform**: eine Alternative,
deren Zweige auf eine negative Vorausschau enden dürfen. Dieses Programm
trennt den Zweig deshalb an seiner `(?!…)`, prüft den Kopf mit `regexp` und
stellt die Vorausschau als eigene Frage an den Rest. Was diese Form nicht
trifft, wird übersprungen statt halb gelesen.

Gewählt wird der **längste** Zweig: MSE prüft mit `wxRegEx`, einem
POSIX-Ausdruck, und dort gilt „leftmost-longest".

**Und zerlegt wird ein Feld nur, wenn der Stil `always symbol: true` sagt**
(`src/render/text/element.cpp`, Zeile 244) — 133-mal im Pack, an den
Manakosten und nie am Regeltext. Ohne die Ansage ist der Wert Text, in dem nur
`<sym>`-Stellen Symbole sind. Sie ist zusammen mit dem Auffangausdruck zu
lesen: der deckt **jedes** Zeichen, und ohne die Ansage wäre jeder Regeltext
eine Reihe von Bildern.

**Quelle:** MSE2 `src/data/symbol_font.cpp`, Zeilen 103 und 191–230, sowie
`src/render/text/element.cpp`, Zeile 244 — Fassung f9d9356 vom 2021-07-06; dazu
`magic-mana-large.mse-symbol-font/symbol-font` aus `M15-Magic-Pack`
**Erschlossen am:** 2026-08-17 (ersetzt die Annahme vom 2026-08-16)
**Sicherheit:** belegt
**Testfall:** TestDerErsteCodeDerSchriftGewinnt,
TestDieFassungMitDerErfuelltenBedingungGewinnt,
TestDieDateireihenfolgeEntscheidetUeberDieZerlegung,
TestEinUnlesbarerAusdruckTrifftNicht, TestEinAuffangcodeMitVorausschauTrifft,
TestOhneAnsageBleibtEinFeldText, TestEinFeldstilSagtObSeinWertImmerAusSymbolenBesteht

---

## Ein `choice images:`-Block wird zu `[…][input]() or else nil`

`src/data/field/choice.cpp:171` baut aus

```
choice images:
     a: {uvw}
     b: {xyz}
```

das Skript `[a: {uvw}, b: {xyz}][input]() or else nil` — eine Abbildung von
Wahlname auf Bildskript, indiziert mit dem **Wert des Feldes**, gerufen.

**Jeder Name steht zweimal darin:** wie geschrieben und in der unkanonischen
Form, in der Unterstriche zu Leerzeichen werden (`util/string.cpp:127`).

**Quelle:** MSE2, `src/data/field/choice.cpp:171` und `src/util/string.cpp:127`,
Fassung f9d9356 vom 2021-07-06.
**Erschlossen am:** 2026-08-16
**Sicherheit:** belegt
**Testfall:** TestEinWahlbildWirdNachDemFeldwertGewaehlt

**Der Block steht im Stil, nicht im Spiel** — `magic-m15.mse-style:162`. Bis
zum 2026-08-16 meldete dieses Programm „ein Bild aus den Wahlbildern des
Spiels"; das war falsch.

**Ein Eintrag ist ein Pfad oder ein `script:`-Block.** Beide Formen stehen im
Pack: `magic-m15.mse-style:163` schreibt den Pfad hin,
`magic-modules.mse-include/rarities/choice_images` schachtelt ein Skript. Ein
Leser, der nur die erste Form kennt, lässt genau die 342 Seltenheitsfelder leer.

---

## `shrink-overflow` **staucht waagerecht**, es verkleinert nicht

`src/util/alignment.cpp:46` macht aus `shrink` ein `ALIGN_STRETCH` und aus
`overflow` ein `ALIGN_IF_OVERFLOW`. Der Faktor steht in
`src/data/field/text.cpp:73`:

```cpp
double factor = (width - padding_left - padding_right) / layout->width;
if (!(alignment() & ALIGN_IF_OVERFLOW) || factor < 1.0) return factor;
```

**Die Schriftgröße bleibt; der Zug wird in der Breite gestaucht.** Bei
`overflow` nur, wenn er sonst überliefe; ohne `overflow` wird auch gestreckt.

**Und ein gestauchter Zug wird nicht mehr ausgerichtet.** Im Stretch-Zweig
lässt MSE die Stellen stehen, wo sie sind — kein `align_delta_x`
(`src/render/text/viewer.cpp`, Zeilen 863–865); der Zug beginnt an der linken
Kante des Kastens. Passt er dagegen hinein, greift die gewöhnliche
Ausrichtung: 180 Feldstile des Packs tragen `center middle shrink-overflow`
und stehen mittig, solange sie nicht überlaufen.

**Gelesen wird ein Teilwort, kein Wort:** `s.find(_("shrink"))`. Im Pack steht
676-mal `shrink-overflow` und zweimal bloß `shrink`.

**Was hier fehlt:** die **Polsterung** (`padding_left`, `padding_right`) geht
in den Faktor ein, und `msepaket` liest sie noch nicht. Der Faktor fällt damit
zu groß aus — bei `magic-m15` um 10 Punkte auf 314.

**Quelle:** MSE2 `src/util/alignment.cpp`, Zeilen 46–49;
`src/data/field/text.cpp`, Zeilen 73–81; `src/render/text/viewer.cpp`, Zeilen
837–872; `src/render/value/text.cpp`, Zeile 26 — Fassung f9d9356 vom
2021-07-06
**Erschlossen am:** 2026-08-16, gebaut am 2026-08-17
**Sicherheit:** belegt
**Testfall:** TestEinZuBreiterZugWirdGestauchtUndNichtVerkleinert,
TestGedehntWirdNurOhneUeberlaufbedingung, TestEinGestauchterZugStehtLinks,
TestEinGestauchterZugBleibtInSeinemFeld

---

## Der zweistufige Satz ist **ein** zusätzlicher Durchgang, kein Festpunkt

MSE zeichnet eine Karte so (`src/render/card/viewer.cpp:44`):

```cpp
FOR_EACH(v, viewers) { if (v->prepare(dc)) changed_content_properties = true; }
if (changed_content_properties) updateStyles(true);   // nur inhaltsabhängige
FOR_EACH(v, viewers) { drawViewer(dc, *v); }
```

1. **`prepare`** setzt je Feld `content_width` und `content_height` und meldet,
   ob sich etwas geändert hat.
2. Hat es das, werden **nur die inhaltsabhängigen Stile** neu gerechnet
   (`s->content_dependent`, `script_manager.cpp:229`).
3. Dann wird gezeichnet.

**Genau ein zusätzlicher Durchgang** — kein wiederholtes Annähern.

**Und vor dem ersten steht `updateStyles(false)`.** Bevor irgendetwas gesetzt
wird, sind **alle** Stilangaben ausgerechnet: `Style::update`
(`src/data/field.cpp:141`) wertet `left`, `width`, `right`, `top`, `height`,
`bottom` aus, `TextStyle::update` (`src/data/field/text.cpp:83`) zusätzlich
`font` und `symbol_font`, und `Font::update` (`src/data/font.cpp:27`) Name und
Größe der Schrift.

Roh gelesen wird bei MSE also **keine** dieser Angaben. Wer misst, misst mit
ausgerechneten Angaben — auch dann, wenn eine davon selbst ein Skript ist. Ein
Knoten entsteht daraus nicht: liest so ein Skript `card_style`, bekommt es im
ersten Durchgang den Stand von vorher, und der zweite stellt es richtig
(#420).

**Quelle:** MSE2, `src/render/card/viewer.cpp:44`,
`src/script/script_manager.cpp:208–240`, `src/data/field.cpp:141`,
`src/data/field/text.cpp:83` und `src/data/font.cpp:27`, Fassung f9d9356 vom
2021-07-06.
**Erschlossen am:** 2026-08-16, ergänzt am 2026-08-18
**Sicherheit:** belegt
**Testfall:** TestDieInhaltsbreiteEinesSymbolfeldsWirdGemessen,
TestGerechneteSchriftangabenWerdenGemessen

---

## `type` und `card color` werden **abgeleitet**, nicht eingetippt

`magic.mse-game/card_fields` führt beide als Kartenfeld — und beide rechnen:

```
name: type
save value: false
script:
	combined_editor(
		field1: card.super_type,
		separator: language().type_separator,
		field2: card.sub_type,
		soft_before_empty: true,
		hide_when_empty:   true
	)
```

Zeilen 963–991. **`save value: false` ist der stärkste Beleg**: ein Feld, das
nie gespeichert wird, muss gerechnet werden.

```
name: card color
default: stylesheet; card_color(casting_cost: card.casting_cost, rules_text: card.rule_text, type: …)
script: card_color_filter(value)
```

**Der Unterschied ist die Stufe, nicht die Sache.** Ein `script:` rechnet
**immer**; ein `default:` nur, wenn die Karte nichts führt. MSE schreibt
`card color` als `default:`, weil der Anwender die Farbe überschreiben können
soll — `type` dagegen als `script:` mit `save value: false`.

**Für eine importierte Karte heißt das**: `super type` und `sub type` genügen,
`type` entsteht daraus. Wer beim Import `type` **und** die Teiltypen setzt,
setzt einen Wert, den MSE ohnehin überschreibt.

**Quelle:** `magic.mse-game/card_fields`, Zeilen 963–991 und der Block
`card color`, Stand 622160f des Packs.
**Erschlossen am:** 2026-08-18
**Sicherheit:** belegt — beide Felder tragen die Angabe im Paket
**Testfall:** TestDerTypWirdAusOberUndUntertypGerechnet,
TestDieKartenfarbeKommtAusIhremVorgabeskript

---

## `content_width` ist die Breite des **gesetzten** Textes, nach dem Umbruch

Für ein Textfeld ist es `layout->width` (`src/data/field/text.cpp:118`); daneben
stehen `content_height` und `content_lines`. Für ein Auswahlfeld ist es die
Größe des **erzeugten Bildes** (`src/render/value/choice.cpp:52`).

**Quelle:** MSE2, `src/data/field/text.cpp:118` und
`src/render/value/choice.cpp:52`, Fassung f9d9356 vom 2021-07-06.
**Erschlossen am:** 2026-08-16
**Sicherheit:** belegt
**Testfall:** TestDasNamensfeldRechnetMitDerInhaltsbreite

**Was dieses Programm heute misst, ist etwas anderes:** die Breite des Zugs
**ohne** Umbruch. Bei einem einzeiligen Feld — den Manakosten etwa — ist das
dasselbe; bei einem umgebrochenen Regeltext nicht. Die Abweichung steht hier,
statt unbemerkt zu bleiben
(#291).

---

## `always symbol: true` mischt Symbole **in den Text**, ohne `<sym>`

`src/render/text/element.cpp:244`: ist `always_symbol` gesetzt und eine
Symbolschrift da, sucht MSE mit `recognizePrefix` an **jeder Stelle** nach
Symbolcodes und macht daraus abwechselnd `FontTextElement` und
`SymbolTextElement`.

`recognizePrefix` ist dasselbe wie unser „längster Code zuerst" — nur inline
statt für das ganze Feld.

**Quelle:** MSE2, `src/render/text/element.cpp:244`, Fassung f9d9356 vom
2021-07-06.
**Erschlossen am:** 2026-08-16
**Sicherheit:** belegt
**Testfall:** TestDerErsteCodeDerSchriftGewinnt

**Dieses Programm entscheidet alles oder nichts** (`ganzAusSymbolen`) — ein
Regeltext bekommt deshalb keine Symbole
(#312).

---

## `<atom-param>` und `<kw-…>` zeichnen **sichtbare** Klammern

`src/render/text/element.cpp:238`:

```cpp
if (kwpph > 0 || param > 0) {
  content = String(LEFT_ANGLE_BRACKET) + content + RIGHT_ANGLE_BRACKET;
```

Der Inhalt bekommt ein Klammerpaar **dazu**.

**MSE2 deutet 22 Auszeichnungen**, nicht fünf: `align`, `atom`, `atom-kwpph`,
`atom-param`, `b`, `bullet`, `code`, `code-kw`, `code-str`, `color`, `error`,
`font`, `i`, `li`, `line`, `margin`, `ref-param`, `sep-soft`, `size`, `soft`,
`soft-line`, `sym`. Erkannt wird über ein **Präfix** (`is_tag`), deshalb deckt
`<i` auch `<i-flavor>` und `<i-auto>`.

**Fünf davon tragen einen Parameter** hinter einem Doppelpunkt — `<color:black>`,
`<font:Beleren Bold>`, `<size:…>`, `<margin:…>`, `<align:…>`.

**Auch hier stand bis zum 2026-08-16 eine falsche Zahl** („überhaupt nur fünf"),
und auch sie trug `Sicherheit: belegt`. Sie kam aus einer Suche nach
`_("…")`-Zeichenketten in zwei Dateien statt aus der Liste selbst.

**Quelle:** MSE2, `src/render/text/element.cpp:230–244`, Fassung f9d9356 vom
2021-07-06; gezählt in `magic.mse-game`: 2.165 `<atom-param>`.
**Erschlossen am:** 2026-08-16
**Sicherheit:** belegt
**Testfall:** TestAuszeichnungenStehenNichtRohAufDerKarte

**Dieses Programm entfernt sie ersatzlos** und verliert damit das Klammerpaar
(#312).

---

## Ein Bildskript gibt ein **Bild** zurück, keinen Dateinamen

`src/gfx/generated_image.hpp` von MSE2 führt 21 Klassen, die von
`GeneratedImage` erben. **Fünf davon erzeugen ein Bild aus dem Nichts** —
`BlankImage`, `PackagedImage` (eine Datei im Paket), `BuiltInImage`,
`SymbolToImage`, `ImageValueToImage` —, die übrigen sechzehn machen Bilder aus
Bildern.

**Bis zum 2026-08-16 stand hier „ein Blatt und zwanzig Verknüpfungen".** Das war
aus #270
übernommen und **nicht nachgezählt**, obwohl daneben `Sicherheit: belegt`
stand — angestrichen vom Nachlauf des Zweitblicks. Eine falsche Zahl unter einer
richtigen Quellenangabe ist schlimmer als gar keine.

**Quelle:** MSE2, `src/gfx/generated_image.hpp` und `src/gfx/blend_image.cpp`,
Fassung f9d9356 vom 2021-07-06 — geklont und gelesen am 2026-08-16.
**Erschlossen am:** 2026-08-16
**Sicherheit:** belegt
**Testfall:** TestRecolorImageGibtEinenBildbaum

Gezählt in `M15-Magic-Pack`: `masked_blend` 2.734, `recolor_image` 579,
`set_alpha` 572, `set_combine` 555, `linear_blend` 451 — die fünf häufigsten
machen 96 Prozent aus.

---

## `masked_blend` mischt **je Farbkanal**, nicht über die Helligkeit

`blend_image.cpp:63` läuft über Teilbildpunkte:

```cpp
data1[i] = (data1[i] * dataM[i] + data2[i] * (255 - dataM[i])) / 255;
```

`data1` ist `light`, `data2` ist `dark`, `dataM` die Maske. Bei rot wirkt der
**rote** Maskenkanal, bei grün der grüne.

**Drei Einzelheiten, die nur der Quelltext zeigt:** es sind `w * h * 3` Kanäle,
also **kein Alpha**; die Bilder müssen **gleich groß** sein, sonst wirft MSE
`Images used for blending must have the same size`; und RGB liegt in `wxImage`
**ungewichtet** vor, ein durchsichtiger Punkt trägt also trotzdem eine Farbe.

**Quelle:** MSE2, `src/gfx/blend_image.cpp:63`, Fassung f9d9356 vom 2021-07-06.
**Erschlossen am:** 2026-08-16
**Sicherheit:** belegt
**Testfall:** TestMaskedBlendMischtUeberDieMaske

**Der erste Anlauf nahm die mittlere Helligkeit** und begründete das damit, eine
Maske sei kein Foto. Die Begründung war erfunden — sie stand da, bevor jemand
den Quelltext gelesen hatte.

---

## `linear_blend` blendet entlang einer Geraden, in Anteilen

```cpp
int mult = x * xm + y * ym + d;   // 0 bei (x1,y1), fixed bei (x2,y2)
data1[0] = data1[0] + mult * (data2[0] - data1[0]) / fixed;
```

Die Koordinaten sind **Anteile** der Breite und Höhe; MSE rechnet sie über
`x1 * width` in Bildpunkte um. Gerechnet wird mit dem **Punktindex**, nicht mit
seiner Mitte. Fallen (x1,y1) und (x2,y2) zusammen, wirft MSE
`coordinates for blending overlap`.

**Quelle:** MSE2, `src/gfx/blend_image.cpp:18`, Fassung f9d9356 vom 2021-07-06.
**Erschlossen am:** 2026-08-16
**Sicherheit:** belegt
**Testfall:** TestLinearBlendBlendetEntlangEinerGeraden

---

## Ein `default:` eines Kartenfelds ist ein **Skript**

Alle 142 Vorgabewerte von `magic.mse-game/card_fields` beginnen mit
`stylesheet; ` und sind **nicht** geklammert:

```
default: stylesheet; card_color(casting_cost: card.casting_cost, …)
```

`a; b` ist eine Folge; das vorangestellte `stylesheet` ist eine
**Abhängigkeitsangabe**. Die Gegenprobe steht in `set_fields`:
`default: "Type custom types here; …"` — in Anführungszeichen, also eine
Zeichenkette **in Skriptsyntax**. Beide Formen sind Skript.

**Quelle:** `magic.mse-game/card_fields` und `set_fields` aus `M15-Magic-Pack`,
gelesen am 2026-08-16 — 142 Vorkommen, keines geklammert, keines ohne Präfix.
**Erschlossen am:** 2026-08-16
**Sicherheit:** belegt
**Testfall:** TestEinVorgabewertIstEinSkriptAuchOhneKlammern

---

## `stylesheet` trägt die Maße der Vorlage

`stylesheet.card_width` ist mit 499 Zugriffen der meistgelesene Wert des Packs,
`.card_height` folgt mit 506. Sie stehen **neben** den Styling-Optionen.

**Quelle:** `magic.mse-game/script:2891` rechnet
`faces_coordinates := { [[ … width: stylesheet.card_width … ]] }`; gezählt über
`M15-Magic-Pack` am 2026-08-16.
**Erschlossen am:** 2026-08-16
**Sicherheit:** belegt
**Testfall:** TestStylesheetTraegtDieKartenmasse

**Nicht gesetzt sind** `stylesheet.card_style` (36 Zugriffe), `.png` (24),
`.short_name` (20) und `.folder_name` (1): was sie bedeuten, ist nicht
erschlossen, und ein geratener Wert sähe aus wie ein gelesener.

**In der Ausfuhr ist der Stil ein Gegenstand an `set` und `card`**
(#2039):
`set.stylesheet` mit `short_name`, `full_name`, `card_width`, `card_height`,
`card_dpi` und `card_regions`, `card.stylesheet` der eigene Stil der Karte oder
`nil`. **`card_regions` gibt es erst in MSE 2.5** — eine Liste von Regionen
(`name`, `x`, `y`, `width`, `height`), leer, solange der Stil keine erklärt.
**Erklärt wird je Region ein `card region:`** im Kopf des Stils — acht Stile
des Packs tun es, etwa `magic-m15-mainframe-dfc` mit `front` (x 0) und `back`
(x 377) bei 752 × 523.

**Quelle:** MSE2 (Orakel über `Multicolor.mse-set`, Fassung f9d9356 vom
2021-07-06; `src/data/stylesheet.cpp` 94–126) und die Fortführung
`G-e-n-e-v-e-n-s-i-S/MagicSetEditor2` (`src/data/stylesheet.hpp` 44,
`src/data/card_region.cpp` 19–24, Fassung a4b79c9 vom 2026-09-04); gelesen
von `magic-cockatrice-v2`, `magic-lackeyccg`, `magic-lackeybot`.
**Erschlossen am:** 2026-10-03
**Sicherheit:** belegt (Orakel und Quelltext); wie ein Stil in 2.5 Regionen
erklärt, ist offen — das Pack enthält kein Beispiel
**Testfall:** TestDieAusfuhrKenntDenStilAlsGegenstand

---

## Ein `stylesheet:` trägt die Kurzform, das Paket den Namen mit Spiel davor

In einem `.mse-set` steht unter `stylesheet:` **nicht** der Paketname, sondern
der Name ohne das Spiel davor. `game: magic` mit `stylesheet: m15-altered`
meint das Paket `magic-m15-altered.mse-style`.

**Quelle:** `fenhl/magic-set-generator` (MIT), `src/mse.rs`, Fassung 1b8fb7f vom 2021-08-09
— es *schreibt*
`.mse-set`-Dateien und ist damit die einzige belegte Auskunft darüber, was in
einem echten steht. Es legt ein Set mit `("stylesheet", … "m15-altered")` an
und benennt den zugehörigen Stylingblock über
`format!("{}-{}", mse_game, stylesheet.render())` als `magic-m15-altered`.
**Erschlossen am:** 2026-08-16
**Sicherheit:** belegt
**Testfall:** TestEinStylesheetInKurzformFindetSeinPaket

**Der volle Name wird trotzdem weiter aufgelöst.** Ob jedes Werkzeug die
Kurzform schreibt, ist damit nicht gesagt, und §8 macht fremde Daten zum
Normalfall; unsere eigene Saat schreibt ohnehin den vollen. Gesucht wird die
zusammengesetzte Form **zuerst**, denn sie ist die belegte.

**Was das gekostet hätte:** bis zum 2026-08-16 wurde nur der rohe Wert gesucht.
Jedes echte Set fand damit kein Stilpaket, und die Anwendung zeigte einen
Behelfsplan mit den rohen Feldwerten darauf — sichtbar geworden ist es erst,
als ein Wächter zum ersten Mal durch die ganze Naht ging
(#292).

**Und geschrieben wird die Kurzform** — seit dem 2026-10-11 (F05,
#2247,
Zweitblick über Phase G, zweiter Teil, Befund 2). MSE2 bestätigt die Lesart
aus dem eigenen Quelltext: `StyleSheet::stylesheetName` nimmt dem Paketnamen
das Spiel samt Bindestrich, `StyleSheet::byGameAndName` sucht
`<spiel>-<wert>.mse-style`, und den Block `styling:` ordnet
`Set::stylingDataFor` nach dem **vollen** Namen (`src/data/stylesheet.cpp`
26–60, `src/data/set.cpp` 99, Fassung f9d9356 vom 2021-07-06). Bis dahin
übernahm der Schreiber `game:` und `stylesheet:` aus der Quelldatei — ein
gewechselter Stil stand im Bestand und **nicht in der Datei**, und ein hier
angelegtes Set trug den vollen Namen, den MSE2 nicht findet. Der Schreiber
löst den Wert jetzt über die Paketwurzeln auf und schreibt die Kurzform;
**ohne installiertes Paket bleibt der Wert, wie er im Bestand steht**, denn
welche Form gemeint ist, sagt nur das Paket.
**Quelle:** MSE2, wie oben.
**Erschlossen am:** 2026-10-11
**Sicherheit:** belegt
**Testfall:** TestDerKopfFolgtDemBestand, TestEinGewechselterStilStehtInDerGesichertenDatei

---

## Ein Sprachpaket bildet Spiel → Feld → Beschriftung ab

Dazu ein `package:`-Block für die Paketnamen selbst.

**Quelle:** `Full-Magic-Pack/data` führt dreizehn `.mse-locale`-Pakete; die
innere Form ist aus dem Zweck geschlossen.
**Erschlossen am:** 2026-08-08
**Sicherheit:** Vermutung
**Testfall:** TestSprachmodellLiestFeldbeschriftungen

---

## Ein Wert kann ein Skript sein — geklammert oder eingerückt

Zwei Schreibweisen, und sie bedeuten dasselbe:

```
top: { if styling.imagetop != "" then to_real(styling.imagetop) else 420 }

script:
	if foil_is_full() then 0
	else if is_transparent() then 14
	else 29
```

Die erste steht als Wert hinter dem Doppelpunkt, die zweite als eingerückter
Block darunter. Ein Ausdruck darf dabei über **mehrere Zeilen** laufen: die
Einrückung hält ihn zusammen, ein Umbruch beendet ihn nicht.

Das ist die Naht zwischen `msepaket` und `msescript`: der Paketleser sieht hier
Text, der Auswerter macht daraus einen Wert. Wer die Klammern schon im Zerteiler
auflöst, hat die beiden Aufgaben vermischt.

**Quelle:** `kamen-decade.mse-style/style`, Zeilen 105–108 (geklammert);
`yugioh-series10.mse-style/style`, Zeilen 152–158 (eingerückt);
`magic-m15.mse-style/style`, `init script:`.
**Erschlossen am:** 2026-08-08
**Sicherheit:** belegt — drei Spiele, verschiedene Urheber.
**Testfall:** TestGeklammertUndEingerueckt

---

## Eine Schleife **addiert** ihre Ergebnisse, sie sammelt sie nicht

```cpp
// the loop body should have a net stack effect of 0, but the entire expression of +1
// solution: add all results from the body, start with nil
script.addInstruction(I_PUSH_CONST, script_nil);      //    push nil
…
parseOper(input, script, PREC_SET, I_BINARY, I_ADD);  // EEE;  add
```

`src/script/parser.cpp`, Zeilen 554–608. Was herauskommt, entscheidet also `+`:

| Rumpf | Ergebnis |
| --- | --- |
| `[wert]` | eine **flache** Liste — zwei Sammlungen verketten sich |
| eine Zahl | die **Summe** |
| ein Text | der aneinandergesetzte Text |

**Und `+` verkettet zwei Sammlungen**, es verschmilzt sie nicht:

```cpp
} else if (at == SCRIPT_COLLECTION && bt == SCRIPT_COLLECTION) {
    a = make_intrusive<ScriptConcatCollection>(a, b);
```

`src/script/context.cpp`, Zeile 465 ff. Zahlenschlüssel werden dabei
umnummeriert (`ScriptConcatCollection::getMember`, `src/script/value.cpp`,
Zeilen 551–562); benannte Schlüssel bleiben, wie sie sind.

**Was das gekostet hat:** `for x from 1 to 9 do [wert]` ergab eine Liste von
Listen, und der nächste Zugriff darauf rechnete mit einer Liste statt mit einer
Zahl. Über die 190 Stile von `M15-Magic-Pack` hingen **43 der 165 Befunde**
daran (#390).

**Quelle:** MSE2, `src/script/parser.cpp:554`, `src/script/context.cpp:465`,
`src/script/value.cpp:551`, Fassung f9d9356 vom 2021-07-06.
**Erschlossen am:** 2026-08-17
**Sicherheit:** belegt
**Testfall:** TestEineSchleifeAddiertIhreErgebnisse

---

## Ein Auswahlstil richtet **ohne Angabe** auf `stretch` aus

```cpp
ChoiceStyle::ChoiceStyle(…)
  : alignment(ALIGN_STRETCH)
```

`src/data/field/choice.cpp`, Zeile 167. Gelesen wird es beim Zeichnen:

```cpp
opts.preserve_aspect = (style.alignment & ALIGN_STRETCH) ? ASPECT_STRETCH : ASPECT_FIT;
```

`src/render/value/choice.cpp`, Zeilen 111–124.

**Ein Auswahlfeld mit `render style: image` und ohne `alignment:` wird also
gestreckt**, nicht eingepasst. Erst eine ausdrückliche Ausrichtung ohne
`stretch` macht daraus eine Einpassung — so beim Seltenheitssymbol
(`alignment: middle right`).

**Und die Maske reist mit dem eingepassten Bild**, nicht mit dem Feld:

```cpp
if (mask) {
  GeneratedImage::Options mask_opts(options);
  mask_opts.width  = cached_i.GetWidth();
  mask_opts.height = cached_i.GetHeight();
```

`src/script/image.cpp`, Zeilen 144–149. Wer also fälschlich einpasst, staucht
die Maske mit: beim Feld `corners` (5 × 5 großes Bild, 375 × 523 große Maske)
landete deren untere Rundung 148 Zeilen zu hoch und riss zwei weiße Kerben in
den Kartenrand (#388).

**Quelle:** MSE2, `src/data/field/choice.cpp:167`, `src/render/value/choice.cpp:111`,
`src/script/image.cpp:144`, Fassung f9d9356 vom 2021-07-06.
**Erschlossen am:** 2026-08-17
**Sicherheit:** belegt
**Testfall:** TestEinAuswahlbildOhneAusrichtungWirdGestreckt

---

## Ein Feldwert darf ein **eingerückter Block** sein

```
rule_text:
	When you set this scheme in motion, until your next turn, whenever a
	player taps a land for mana, that player adds one mana of any type.
name: A Display of My Dark Power
```

Hinter dem Doppelpunkt steht dann **nichts**; der Wert sind die eingerückten
Zeilen darunter. Gemessen an unserem Zerteiler:

```
rule_text | Wert="" | Kinder=2 | Skripttext="Erste Zeile\nZweite Zeile"
name      | Wert="Feuerball" | Kinder=0
```

Es ist dieselbe Form, in der auch ein `script:`-Block steht — „was steht in dem
Block unter dieser Zeile" ist eine Frage, keine zwei.

**Was das gekostet hat:** Regeltext und Flavortext einer kanonischen Karte kamen
nie an, und weil ein Textfeld ohne Wert gar nicht gezeichnet wird, ohne Befund
(#398).

**Quelle:** `Archenemy.mse-set` (kanonisch, MSE 2.1.2), jede Karte mit
Regeltext; gemessen am 2026-08-17.
**Erschlossen am:** 2026-08-17
**Sicherheit:** belegt
**Testfall:** TestEinMehrzeiligerFeldwertKommtAn

---

## Ein Feldname in einer Datei ist **kanonisch**: `_` steht für ein Leerzeichen

```cpp
void canonical_name_form_in_place(String& str) {
  for (String::iterator it = str.begin(); it != str.end(); ++it) {
    if (*it == ' ') *it = '_';
  }
}
void uncanonical_name_form_in_place(String& str) {
  for (String::iterator it = str.begin(); it != str.end(); ++it) {
    if (*it == '_') *it = ' ';
  }
}
```

`src/util/string.cpp`, Zeilen 121–133. **MSE hält es innen kanonisch**: sein
Leser wandelt jeden Schlüssel um (`key = canonical_name_form(trim(key))`,
`src/util/io/reader.cpp`, Zeile 217), sein Skriptzerteiler jeden Namen
(`parser.cpp`, Zeile 200), und ein Feldname wird beim Lesen ebenso umgewandelt
(`src/data/field.cpp`, Zeile 68). Die Leerzeichen gibt es nur in den Dateien.

**Dieses Projekt hält es umgekehrt** — Feldnamen mit Leerzeichen, wie die
Vorlage sie erklärt — und wandelt an den Stellen um, an denen ein Schlüssel ein
**Feldname** ist: beim Lesen der Kartenwerte und der Setangaben, beim Schreiben
zurück. **Nicht am Zerteiler**: dort träfe es auch `mana_t: …` in einer
Skriptdatei, und da ist der Unterstrich Teil des Namens.

**Was das gekostet hat:** ein von MSE2 gespeichertes Set schreibt `rule_text:`,
und der Feldstil heißt `rule text`. Der Regeltext kam nie an — und weil ein
Textfeld ohne Wert gar nicht gezeichnet wird, **ohne Befund**
(#396).
Unsere eigenen Sets merkten nichts davon: `saat/` und `testdaten/` führen nur
Felder ohne Leerzeichen.

**Quelle:** MSE2, `src/util/string.cpp:121`, `src/util/io/reader.cpp:217`,
`src/data/field.cpp:68`, Fassung f9d9356 vom 2021-07-06; dazu
`Archenemy.mse-set` (kanonisch, MSE 2.1.2).
**Erschlossen am:** 2026-08-17
**Sicherheit:** belegt
**Testfall:** TestEinFeldnameWirdEntkanonisiert

---

## Auch der **Schriftname** ist oft ein Skript

Dieselbe Form wie bei der Farbe, am selben `font:`-Block. Ausgezählt über alle
190 `.mse-style`-Pakete von `M15-Magic-Pack`:

| | |
| ---: | --- |
| 3.119 | Feldstile nennen `font: name:` |
| **664** | davon sind Skripte |
| 581 | nennen `italic name:` |
| **439** | davon sind Skripte |

```
name: { body_font() }
name: { if styling.beleren then "Beleren" else "Matrix" }
```

**Reichweite** (§10): gezählt am Feldstil, nicht an gezeichneten Ebenen.

**Was das kostet, wenn man es übersieht:** die Schriftsuche bekommt den rohen
Text und sucht nach einer Schrift dieses Namens. Sie findet keine, und die Karte
bleibt ohne Text — gesehen am ersten Lauf über ein kanonisches MSE2-Set
(#393).

**Quelle:** `M15-Magic-Pack`, alle 190 `.mse-style`-Pakete, ausgezählt am
2026-08-17; `Archenemy.mse-set` (kanonisch, `stylesheet: archenemy`).
**Erschlossen am:** 2026-08-17
**Sicherheit:** belegt
**Testfall:** TestEinSchriftnameDarfEinSkriptSein

---

## Eine Schriftfarbe ist im Regelfall ein **Skript**, kein Farbname

`font: color:` steht selten als Literal da. Ausgezählt über **alle 190
`.mse-style`-Pakete von `M15-Magic-Pack`**, über die 5.950 Feldstile ihrer
`card style:`-Blöcke:

| | |
| ---: | --- |
| 3.011 | Feldstile nennen eine `font: color:` |
| **2.321** | davon sind **Skripte** |
| 6 | davon in der eingerückten `script:`-Form (alle in `magic-m15-future`) |

Was dort steht, sind drei Formen:

```
color: {styling.artist_color}
color: { font_color(card.card_color) }
color: { styling.copyright_text_color or else set.copyright_text_color }
```

**Ausgerechnet ergibt das entweder einen Farbwert oder einen Farbnamen.** Beides
kommt vor; `{ "white" }` ist ein Text, `{ rgb(255,255,255) }` eine Farbe.

**Was das kostet, wenn man es übersieht:** der rohe Text ist keine Farbangabe,
und eine Deutung als Literal fällt auf Schwarz zurück. Der **Fuß der Karte** —
Illustrator, Copyright, Setcode — stand damit als schwarze Schrift auf dem
schwarzen Kartenrand: gezeichnet und unsichtbar
(#382).

**Reichweite der Zahlen** (§10): gezählt am **Feldstil**, nicht an gezeichneten
Ebenen. Wie viele Ebenen einer Karte betroffen sind, hängt an `visible:` und
daran, welche Felder die Karte führt.

**Quelle:** `M15-Magic-Pack`, alle 190 `.mse-style`-Pakete, ausgezählt am
2026-08-17; `magic-modules.mse-include/information/card_fields` für den Fuß.
**Erschlossen am:** 2026-08-17
**Sicherheit:** belegt — 46 Pakete verschiedener Urheber führen die Form.
**Testfall:** TestEineSchriftfarbeDarfEinSkriptSein

---

## Eine geschweifte Klammer ist ein Funktionswert

Das war die teuerste offene Frage der Sprache, und sie ist entschieden:

```
get_pendulum_mask := {card.pendulum}
is_nopendulum     := {get_pendulum_mask() == "none"}
is_big_pend       := {get_pendulum_mask() == "big"}
```

Zeile 1 weist eine Klammer einem Namen zu. Zeile 2 **ruft** sie mit `()`. Damit
ist `{ … }` kein Block, dessen Wert sofort entsteht, sondern ein **Wert, der
gerufen werden kann** — und Funktionen sind in dieser Sprache gewöhnliche Werte.

Und sie deklariert **keine Parameter**. `card` in Zeile 1 ist ein freier Name; er
muss beim Aufruf aus der Umgebung kommen, nicht aus dem Ort der Definition.
Anders gesagt: eine Funktion **schließt ihre Umgebung nicht ein, sie bekommt sie
gereicht**.

Hier stand vorher „Vermutung für die Semantik von `{ … }`", erschlossen an
einem Paket, in dem nur `template` so aussah. Der Beleg kam aus einem Spiel,
das mit Magic nichts zu tun hat.

**Warum das mehr ist als eine Feinheit:** steht es falsch im Auswerter, ist
jede Vorlage falsch — und zwar nicht sichtbar falsch, sondern an einzelnen
Feldern, die einen Namen aus der falschen Umgebung nehmen.

**Quelle:** `yugioh-series10.mse-style/style`, Zeilen 102–121 und ihre Aufrufe
ab Zeile 153; `magic-m15.mse-style/style`, `template` als Funktion von `type`
und `input` ohne Parameterliste.
**Erschlossen am:** 2026-08-08
**Sicherheit:** belegt für **Zuweisung und Aufruf**. Dass die freien Namen aus
der *Aufruf*umgebung kommen und nicht aus der Definitionsumgebung, ist damit
nicht bewiesen — beide Umgebungen sind in diesen Dateien dieselbe. Der
Unterschied wird an einem Beispiel gemessen, dessen Ergebnis man im laufenden
MSE2 nachsehen kann (#23).
**Testfall:** TestEineKlammerSiehtDieUmgebungDesAufrufs

---

## Ausdrücke: Zuweisung, Bedingung, Verknüpfung, Rechnung

| Form | Gesehen als |
| --- | --- |
| Zuweisung | `:=`, nie `=` |
| Bedingung | `if … then … else if … then … else …` als **Ausdruck** |
| Verknüpfung | `and`, `or`, `not` — Wörter, keine Zeichen |
| Vergleich | `==`, `!=` |
| Rechnung | `length(remove_tags(card.type_1)) * 17 + 2` |
| Kommentar | `#` bis Zeilenende, auch hinter Code |
| Zugriff | `card.pendulum`, `styling.name_color` |

`+` ist überladen: es rechnet, es verkettet Zeichenketten, und es vereinigt
Abbildungen. Welche Bedeutung gilt, entscheidet sich an den Werten — nicht an
der Schreibweise.

**Keine Schleife gesehen.** In fünf Dateien aus drei Spielen kommt weder `for`
noch `while` vor; der eine Treffer war das englische Wort „for" in einer
Beschreibung. Das heißt **nicht**, dass es keine gibt — es heißt, dass hier
keine steht.

**Quelle:** `yugioh-series10.mse-style/style` (232 `if`, 12 `:=`, `rgba`,
`length`, `remove_tags`); `magic.mse-game/game_script` (`and` 35, `or` 124,
`not` 1, `nil`); `kamen-decade.mse-style/style` (Zeile 63, `if` ohne `else`;
`to_real`).
**Erschlossen am:** 2026-08-08
**Sicherheit:** belegt — jede Zeile der Tabelle an mindestens zwei Spielen.
**Testfall:** TestAuswertenRechnetDieFormenAusDenPaketen

---

## `else` ist weglassbar, und `nil` gibt es

```
visible: { if card.background == "Normal" then true }
if card.stylesheet == nil then …
```

Getrennt von der Tabelle darüber, und nicht aus Ordnungsliebe: beides steht in
je **einem** Paket. Stünde es in derselben Feststellung, trüge es deren Grad —
und ein Wächter, der den ersten Grad liest, hielte eine Vermutung für belegt.
Genau dafür gibt es `TestKeineFeststellungNenntZweiSicherheitsgrade`.

Für den Zerteiler ist die erste Zeile die unangenehmere: ein `if` ohne `else`
muss einen Wert ergeben, wenn die Bedingung falsch ist. Welchen, ist hier nicht
zu sehen.

**Quelle:** `kamen-decade.mse-style/style`, Zeile 63 (`if` ohne `else`);
`magic.mse-game/game_script` (`nil`).
**Erschlossen am:** 2026-08-08
**Sicherheit:** belegt — im Quelltext: ein `if` ohne `else` legt `nil` auf den
Stapel (`src/script/parser.cpp`, Zeilen 505–510), und `nil` ist ein Wort der
Sprache (Zeile 494), Fassung f9d9356 vom 2021-07-06; dazu die beiden Pakete aus
verschiedenen Spielen. Nachgelesen am 2026-10-10 (F05, #2247).
**Testfall:** TestEinIfOhneElseErgibtNichts

---

## Benannte Argumente, und `@` bindet sie im Voraus

```
sort_text@(order: "XYZI[0123456789]HSCA(WUBRG)")()
```

`@` bindet benannte Argumente an eine Funktion und ergibt **wieder eine
Funktion**; das folgende `()` ruft sie. Ohne `@` stehen benannte Argumente
direkt im Aufruf (`chosen(choice: "artifact")`), Stellungsargumente ebenso
(`rgb(156, 130, 90)`).

**Quelle:** `magic-m15.mse-style/style`, Zeile 75; `magic.mse-game/game_script`
(`chosen(choice: …)`).
**Erschlossen am:** 2026-08-08
**Sicherheit:** belegt — im Quelltext: `@(…)` liest benannte Argumente und
erzeugt `I_CLOSURE` (`src/script/parser.cpp`, Zeilen 831–841, Fassung f9d9356
vom 2021-07-06). Die beiden Fundstellen stammen aus demselben Spiel; der
Quelltext macht sie unabhängig davon. Nachgelesen am 2026-10-10 (F05, #2247).
**Testfall:** TestBenannteUndStellungsargumenteNebeneinander

---

## Abbildungen mit `[ … ]`, und `+` **verkettet** sie

`[ schlüssel: wert, … ]`, indiziert über `[…]`. `karte + [stylesheet: "…"]`
ergibt eine Abbildung mit dem zusätzlichen Eintrag.

**Es überschreibt nicht, es hängt an.** Beide Glieder bleiben stehen, und
gelesen wird das **linke**:

```cpp
ScriptValueP ScriptConcatCollection::getMember(const String& name) const {
  ScriptValueP member = a->getMember(name);
  if (member->type() != SCRIPT_ERROR) return member;
  …
  return b->getMember(name);
}
```

`src/script/value.cpp`, Zeilen 551–562. Am Orakel nachgefragt
([0026](entscheidungen/0026-mse2-rendert-die-referenz-selbst.md)):

| Ausdruck | MSE2 |
| --- | --- |
| `([a: 1] + [a: 2]).a` | **1** |
| `length([a: 1] + [a: 2])` | **2** |
| `([1, 2] + [3]).1` | 2 |

**Hier stand bis zum 2026-08-18 „vereinigt sie", mit Vorrang für rechts.** Der
Satz war aus dem Paketverhalten erschlossen — `map := map + [stylesheet: …]`
sieht aus wie ein Überschreiben — und nie am Original geprüft. Über den
Packlauf (190 Stile, 5.950 Feldstile) treffen **6.464** solcher Verbindungen
auf mindestens einen doppelten Schlüssel
(#271).

**Stellenschlüssel werden dabei umnummeriert**, benannte nicht: zwei Listen
verketten sich zu einer längeren.

**Quelle:** MSE2 `src/script/value.cpp`, Zeilen 551–562, Fassung f9d9356 vom
2021-07-06, dazu das Orakel
**Erschlossen am:** 2026-08-08, berichtigt am 2026-08-18
**Sicherheit:** belegt
**Testfall:** TestDiePlusVerkettungVonAbbildungen,
TestAbbildungsplusVerkettet

---

## Paketpfade beginnen mit `/` und sind keine Dateisystempfade

`/kamen-icons.mse-include/Tachibana.png` verweist in ein **anderes Paket**, nicht
auf eine Wurzel des Dateisystems. Dasselbe gilt für `include file:`.

Für den Leser heißt das: ein solcher Pfad darf nie an `os.Open` gereicht werden.
`internal/msepaket` löst ihn über den Paketnamen auf — dieselbe Trennung, die
schon `Paket.Datei` gegen `path` statt `filepath` führt.

**Quelle:** `kamen-decade.mse-style/style` (73 Verweise auf
`/kamen-icons.mse-include/`); `magic-m15.mse-style/style` (`include file:` auf
`/magic-modules.mse-include/`, 13 Verweise).
**Erschlossen am:** 2026-08-08
**Sicherheit:** belegt — zwei Spiele, verschiedene Urheber.
**Testfall:** TestZerlegePaketpfad

---

## Ein Rückstrich flieht das nächste Zeichen

```
replace@(match: "\\", replace: "")      # ein Rückstrich als Muster
contains(match: "\"", card.rule_text)   # ein Anführungszeichen mitten im Text
```

**Hier stand zweimal etwas anderes**, und beide Male aus gutem Grund: zuerst
„Zeichenketten kennen keine Maskierung" (geschlossen aus einem Fehlen, das ein
Suchfehler war), dann „der Lexer reicht durch" — weil **ein** Vorkommen von
`"\n"` nichts entscheidet und Durchreichen die umkehrbare Richtung ist.

**Die Exportvorlagen entscheiden es.** Fünf der neunzehn schreiben `\"` mitten
in einer Zeichenkette; ohne Fluchtfolge endet sie dort, und diese fünf Dateien
wären nicht lesbar. Das ist kein Abwägen mehr, sondern ein Beleg.

| Folge | Ergibt | Grad |
| --- | --- | --- |
| `\"` | ein Anführungszeichen | **belegt** — sonst sind fünf Vorlagen unlesbar |
| `\\` | ein Rückstrich | **belegt** — folgt unmittelbar aus `\"` |
| `\n`, `\t` | Umbruch, Tabulator | **plausibel** — dieselben Vorlagen schreiben `"\\n"`, wo sie die zwei Zeichen eines Musters meinen; das ergibt nur einen Sinn, wenn ein einfaches `\n` etwas anderes ist |
| alles andere | Rückstrich **und** Zeichen | Entscheidung — `"!promo\\b"` meint das Wortende eines Musters, und wer den Rückstrich verschluckte, machte ein anderes daraus |

**Der Weg ist damit nicht mehr umkehrbar**, und das gehört gesagt: wer `\n`
übersetzt, kann aus dem Umbruch nicht mehr ablesen, wie er dastand. Das ist der
Preis dafür, dass fünf Dateien überhaupt lesbar werden.

**Quelle:** `magic-cockatrice-v2.mse-export-template/export-template` und vier
weitere Exportvorlagen aus `M15-Magic-Pack@3f7f8ce/data`;
`magic-m15.mse-style/style` Zeile 289 für das ursprüngliche `"\n"`.
**Erschlossen am:** 2026-08-08, entschieden am 2026-08-15
**Sicherheit:** belegt für `\"` und `\\`, plausibel für `\n` und `\t`.
**Testfall:** TestEinRueckstrichFliehtDasNaechsteZeichen

---

## Zwei Vorrangbeziehungen erzwingen die Pakete

| Beobachtung | Was sie ausschließt |
| --- | --- |
| `card.rarity=="a" or card.rarity=="b"` | bände `or` stärker, stünde da `card.rarity == ("a" or card.rarity) == "b"` |
| `not hide_foil() and card.rarity=="rare"` | bände `not` schwächer, hieße die Zeile „nicht (verborgen und selten)" — und die Nachbarzeilen, die sich nur in der Seltenheit unterscheiden, ergäben keinen Sinn |

Also: Vergleich vor `or`, und `not` vor `and`. Mehr erzwingen die gesichteten
Dateien nicht.

**Quelle:** `yugioh-series10.mse-style/style`, Zeilen 116 und 224.
**Erschlossen am:** 2026-08-08
**Sicherheit:** belegt — beide Beziehungen folgen daraus, dass die andere
Lesart einen Ausdruck ergäbe, den niemand geschrieben haben kann.
**Testfall:** TestVorrangAnBelegtenBeispielen

---

## Die übrige Stufung ist übernommen, nicht beobachtet

Von schwach nach stark: `:=` · `if…then…else` · `or` · `and` · Vergleich ·
`+ -` · `* /` · `not -` · Nachsatz (`.`, Aufruf, `@`).

Fünf der sieben Beziehungen darin stehen in keinem Paket. Dass `*` stärker
bindet als `+`, steht in `length(…)*17 + 2` — aber `l*(17+2)` wäre auch eine
Zahl. Was dafür spricht, ist die Gewohnheit jeder anderen Sprache, und das ist
ein Argument über den Urheber, nicht über MSE.

**Die eine Zeile, die mehr entschieden hätte, entscheidet nichts:**
`no_shadow() and is_spell_card() or is_trap_card()` liest sich unter „`and` vor
`or`" genauso wie unter „gleich stark, nach links" — beide ergeben
`(a and b) or c`. Nur „`or` vor `and`" wäre anders, und dafür spricht nichts.

**Quelle:** `yugioh-series10.mse-style/style`, Zeilen 256 und 732;
`magic-m15.mse-style/style`, Zeilen 288–291 — sämtlich Stellen, die **mehrere**
Lesarten zulassen.
**Erschlossen am:** 2026-08-08
**Sicherheit:** Vermutung
**Testfall:** TestVorrangAnBelegtenBeispielen

---

## Eine Folge trennt am Zeilenumbruch

Ein `init script:` trägt mehrere Zuweisungen untereinander, getrennt allein
durch den Umbruch. Zugleich läuft **ein** Ausdruck über mehrere Zeilen: in
`yugioh` steht eine `else`-Kette über sechs, in `magic-m15` eine `and`-Kette
über vier.

Beides zusammen heißt: der Umbruch trennt nur dort, wo der Ausdruck ohnehin zu
Ende ist. `a b` auf einer Zeile ist deshalb keine Folge, sondern ein Ausdruck,
den jemand nicht zu Ende geschrieben hat.

**Quelle:** `yugioh-series10.mse-style/style`, Zeilen 102–121 (Folge) und
153–158 (ein Ausdruck über sechs Zeilen); `magic-m15.mse-style/style`,
Zeilen 288–293.
**Erschlossen am:** 2026-08-08
**Sicherheit:** belegt — zwei Spiele.
**Testfall:** TestEinSkriptblockIstEineFolge

---

## Freie Namen kommen aus der Umgebung des Aufrufs

```
mana_sort := {
    if styling.unsorted_casting_cost then input
    else sort_text@(order: "…")()
}
```

`input` steht in keiner Zuweisung und in keiner Argumentliste. Es kann nur von
dort kommen, wo `mana_sort` **benutzt** wird — MSE setzt es je Feld. Dasselbe
gilt für `sort_text@(…)()`: der Aufruf übergibt nichts, und sortiert trotzdem
etwas.

Eine geschweifte Klammer schließt ihre Umgebung also **nicht** ein, sie bekommt
sie gereicht. Das ist die teuerste Annahme des Auswerters: stünde sie falsch,
nähme jedes Feld einen Namen aus der falschen Umgebung — nicht sichtbar falsch,
sondern an einzelnen Feldern.

**Was daraus nicht folgt:** dass eine Zuweisung im Block nach außen wirkt. Das
Gegenteil ist nötig, damit zwei Felder, die dieselbe Vorlage benutzen, einander
nicht sehen; belegt ist es nicht.

**Quelle:** `magic-m15.mse-style/style`, Zeilen 54–55 (`type`, `input` in
`template`) und 71–76 (`mana_sort`).
**Erschlossen am:** 2026-08-08
**Sicherheit:** belegt — im Quelltext: ein Aufruf öffnet einen neuen
`LocalScope` **auf dem Kontext des Aufrufers**, setzt dort die Argumente und
wertet den Rumpf in diesem Kontext aus (`I_CALL`, `src/script/context.cpp`,
Zeilen 148–168, Fassung f9d9356 vom 2021-07-06); der Bereich wird danach
zurückgesetzt — eine Zuweisung im Block wirkt also **nicht** nach außen.
Nachgelesen am 2026-10-10 (F05, #2247).

**Und der Test belegt die Entscheidung, nicht das Verhalten von MSE2.** Er
zeigt, dass *dieser* Auswerter die Umgebung des Aufrufs nimmt — nicht, dass MSE
es auch tut. Der Zweitblick über Phase 3 hat angestrichen, dass die
Codekommentare hier „Beleg" sagten, wo die Datei „plausibel" führt; maßgeblich
ist diese Angabe.
**Testfall:** TestEineKlammerSiehtDieUmgebungDesAufrufs

---

## `and` und `or` rechnen die rechte Seite nur bei Bedarf

In `magic-m15` steht eine `and`-Kette, deren spätere Glieder voraussetzen, dass
die früheren zutrafen. Wer beide Seiten immer rechnet, bekommt dort einen
Fehler, den die Vorlage nicht meint.

**Quelle:** `magic-m15.mse-style/style`, Zeilen 288–292.
**Erschlossen am:** 2026-08-08
**Sicherheit:** Vermutung — die Kette **läuft** unter dieser Lesart, das ist
kein Beweis, dass MSE es so hält.
**Testfall:** TestAuswertenRechnetDieFormenAusDenPaketen

---

## `or else` gibt einen Ersatzwert

```
map["stylesheet"] or else "m15-altered"
```

Der linke Wert gilt, wenn er da ist; sonst der rechte. Er steht hinter einem
Index, dessen Schlüssel fehlen darf — und verlangt deshalb **keinen**
Wahrheitswert.

`or` und `else` stehen unmittelbar nebeneinander. Ein Umbruch dazwischen kommt
nirgends vor, und ihn zuzulassen erzeugte eine Mehrdeutigkeit: `if a then b or`
gefolgt von `else c` wäre nicht mehr von `b or else c` zu unterscheiden.

**Quelle:** `magic.mse-game/game_script`, Zeile 62 — das einzige Vorkommen.
**Erschlossen am:** 2026-08-08
**Sicherheit:** belegt — im Quelltext: auf `or` folgt `else`, und daraus wird
`I_OR_ELSE` (`src/script/parser.cpp`, Zeilen 766–771, Fassung f9d9356 vom
2021-07-06; dort auch `orelse` als ein Wort, mit dem Vermerk, `or else` solle
ihm weichen). Nachgelesen am 2026-10-10 (F05, #2247).
**Testfall:** TestDieBibliothekRechnetWasDieVorlagenRufen

---

## Welche Funktionen die Sprache mitbringt, ist an den Dateien nicht zu sehen

In `yugioh-series10` werden `hide_foil`, `is_link`, `is_jp` und zwei Dutzend
weitere gerufen und **nirgends im Paket definiert**. Sie kommen aus
`yugioh.mse-game`, das als Abhängigkeit geführt wird.

Damit ist an einer Stildatei allein nicht zu erkennen, was MSE mitbringt und was
ein Spielpaket beisteuert. Nur `rgb` erscheint in **drei verschiedenen Spielen**
und ist damit sicher keines von beidem.

| Gebaut | Fundstelle |
| --- | --- |
| `rgb`, `rgba` | 95 Aufrufe, `magic`, `yugioh`, `kamen` |
| `to_lower` | `magic.mse-game/game_script` Zeile 62 |
| `to_real` | `kamen-decade.mse-style/style`, Zeilen 105–108 |
| `length` | `yugioh-series10.mse-style/style` Zeile 732 |
| `max` | `magic-m15.mse-style/style` |

**Nicht gebaut:** alles, was nach einem Spiel klingt. Eine erfundene Funktion
ist schlimmer als eine fehlende — sie behauptet, MSE habe sie, und niemand
könnte das je belegen.

**Quelle:** die Aufrufe in `magic-m15.mse-style/style`,
`magic.mse-game/game_script`, `yugioh-series10.mse-style/style`,
`kamen-decade.mse-style/style`.
**Erschlossen am:** 2026-08-08
**Sicherheit:** Vermutung für jede einzelne Zuordnung. Sicher ist nur, dass
diese Namen gerufen werden — nicht, wer sie bereitstellt.
**Testfall:** TestDieBibliothekRechnetWasDieVorlagenRufen

---

## `auto replace` vergleicht feste Zeichenketten, keine Muster

```
auto replace:
	match: (C)
	replace: ©
auto replace:
	match: **
	replace: AUTOREM>
```

Als reguläre Ausdrücke wäre das erste eine Gruppe um ein `C` und das zweite gar
nicht übersetzbar. In 1541 Zeilen kommt **kein** Rückstrich vor, kein `[^`, kein
`.*`, kein `\d`.

`whole word:` steht zweimal, beide Male `false` (an `AE` und `TM`, die auch
mitten im Wort greifen sollen). Die übrigen rund 500 Regeln nennen das Feld
nicht; die Vorgabe ist damit „wahr".

**Daraus folgt die Antwort auf eine Frage, die #25 als Anforderung stellte:** ein
bösartiger regulärer Ausdruck aus einer fremden Vorlage kann die Anwendung nicht
anhalten, **weil es keinen Ausdruck gibt**. Eine Zeitschranke um eine
Mustererkennung, die nicht stattfindet, wäre eine Vorkehrung gegen eine
erfundene Gefahr.

**Quelle:** `magic.mse-game/auto_replace`, gesamte Datei (1541 Zeilen).
**Erschlossen am:** 2026-08-08
**Sicherheit:** belegt — nicht aus einem Fehlen geschlossen, sondern aus zwei
Mustern, die als Ausdruck etwas anderes bedeuteten als das, was danebensteht.
**Testfall:** TestErsetztWirdAnDerSchreibmarke

---

## `whole word:` ist schwächer, als es aussieht — die Grenze sieht auf das Muster

`TextValueEditor::isWordBoundary` prüft an einer Stelle **entweder** das
Zeichen dahinter **oder** das davor, nicht beide. Und `tryAutoReplace` ruft es
mit dem **Anfang des Musters** auf — also mit einer Stelle, deren Zeichen das
**erste Zeichen des Musters selbst** ist.

Daraus folgt: **ein Muster, das mit einem Satzzeichen beginnt, erfüllt seine
linke Grenze immer.** Für `--` heißt das, dass `whole word: true` gar nichts
ändert; beide Striche sind Satzzeichen, also gilt links wie rechts eine Grenze,
wo immer das Muster steht.

```
a--b   →   a—b      (mit whole word: true)
```

Grenzzeichen sind `" ,.:;()\n"` **plus** alles, was `isPunct` sagt. Und
Auszeichnungen werden in beide Richtungen überlesen — mit einem Unterschied,
der zählt: nach **vorn** endet die Suche bei einem `<` ohne sein `>` am
Textende und meldet damit eine Grenze; nach **hinten** bricht ein `>` ohne sein
`<` mit „keine Grenze" ab.

Das sieht nach einem Fehler der Vorlage aus. Portiert wird trotzdem das
Verhalten (§15) — ein „korrigiertes" `whole word` machte jede vorhandene Regel
anders wirksam als in MSE2.

**Quelle:** MSE2, `src/gui/value/text.cpp`, Zeilen 1033–1052 und 1116–1165,
Fassung f9d9356 vom 2021-07-06.
**Erschlossen am:** 2026-08-23
**Sicherheit:** belegt — aus dem Quelltext, nicht aus Beobachtung.
**Testfall:** TestDieWortgrenzeSiehtAufDasMusterSelbst

---

## Der Rückstrich, noch einmal — die Frage ist entschieden

Hier stand ein Argument **gegen** die damalige Wahl: wenn `match:` eine feste
Zeichenkette ist, dann meint `contains(match:"\n", …)` „enthält einen Umbruch",
und dafür müsste `"\n"` in der Zeichenkette schon einer sein.

**Das Argument hat recht behalten, und der Beleg kam von woanders**: aus den
Exportvorlagen, die `\"` mitten in einer Zeichenkette schreiben. Der Abschnitt
„Ein Rückstrich flieht das nächste Zeichen" führt ihn.

Was offen bleibt, ist die kleinere Frage darunter: ob `contains`’ `match:`
dasselbe meint wie das von `auto replace`. Beide heißen gleich und tun
dasselbe; gesehen ist die Gleichsetzung nicht. Sie ändert am Lexer nichts mehr.

**Quelle:** `magic-m15.mse-style/style` Zeile 289 gegen
`magic.mse-game/auto_replace`.
**Erschlossen am:** 2026-08-08, entschieden am 2026-08-15
**Sicherheit:** Vermutung — für die Gleichsetzung der beiden `match:`
**Testfall:** TestEinRueckstrichFliehtDasNaechsteZeichen

---

## Ein Feld liest andere Felder, und die Abhängigkeit ist ablesbar

`if card.rarity == "rare" then "gold" else "silber"` hängt an `card.rarity` —
nicht an `card`. Wer die ganze Abbildung als Abhängigkeit führte, rechnete bei
jeder Änderung an irgendeinem Feld jedes Feld neu; bei 300 Karten ist das der
Unterschied zwischen einer Vorschau und einer Wartezeit.

Ein Kreis zwischen Feldern wird **benannt**: „a → b → c → a" ist zu finden,
„irgendwo im Set ist ein Kreis" nicht.

**Quelle:** `yugioh-series10.mse-style/style`, Zeilen 102–121 — dort lesen
`is_pendulum`, `is_big_pend` und `is_nopendulum` einander über
`get_pendulum_mask`.
**Erschlossen am:** 2026-08-08
**Sicherheit:** belegt — dass Felder einander lesen, ist gesehen, und **dass
MSE2 feinkörnig nachrechnet**, steht im Quelltext: es meldet beim Laden je
Skriptfeld an, was es liest, und rechnet bei einer Änderung nur die
abhängigen neu (`src/script/dependency.cpp`; `updateAllDependend`,
`src/script/script_manager.cpp`, Zeilen 96–200, Fassung f9d9356 vom
2021-07-06; nachgelesen am 2026-08-30, `docs/PARITAET.md`). Eingestuft am
2026-10-10 (F05, #2247).
**Testfall:** TestEineAenderungBetrifftNurDieAbhaengigenFelder

---

## Ein Foil ist eine Bildebene, keine Mischart

In `yugioh-series10.mse-style/style` ist `foils:` ein Block mit `left`, `top`,
`width` und `height` — jedes aus einem Skript, keines eine Rechenvorschrift für
Farben. Das Foil ist also ein **Bild**, das über der Karte liegt.

**Ein Wort für „multiplizieren", „überlagern" oder „abdunkeln" kommt in keiner
gesichteten Datei vor.** `overlay` erscheint ausschließlich als Feldname:
`atkoverlay` und `defoverlay` in Yu-Gi-Oh!, `overlay:` als Bildfeld in Kamen
Rider, und in `magic-m15` als Styling-Feld mit der Beschreibung „Should there be
an overlay applied, such as foil?".

Der Renderer kennt deshalb **keine** Mischart. Eine zu bauen, die kein Paket
nennt, wäre dieselbe Erfindung wie eine Funktion, die keine Vorlage ruft.

**Quelle:** `yugioh-series10.mse-style/style`, Zeilen 150–178;
`magic-m15.mse-style/style`, Zeilen 156–157; `kamen-decade.mse-style/style`,
Zeile 55.
**Erschlossen am:** 2026-08-08
**Sicherheit:** Vermutung — aus einem **Fehlen** geschlossen, und das ist so
viel wert wie die Suche, die nichts fand. Diese hier ging über vier Dateien und
drei Spiele; in Phase 3 ist genau diese Art Schluss schon einmal umgefallen.
**Testfall:** TestEinFoilIstEineBildebeneUeberDerKarte

---

## MSE liest MTG JSON selbst — der Pfad steht im Spielpaket

`magic.mse-game` bindet eine Datei `json_paths` ein. Darin stehen drei
`json path:`-Blöcke; einer davon lautet

```
json path:
	MTGJSON file // data:cards:
```

Damit ist zweierlei belegt: dass MSE MTG JSON als Kartenquelle kennt, und dass
die Karten unter `data` → `cards` liegen — also **unter dem Umschlag**, den
MTGJSON seit Fassung 5 schreibt. Die beiden anderen Blöcke nennen Scryfall.

**Quelle:** `magic.mse-game/json_paths` (204 Bytes) und `magic.mse-game/game`
(`include file: json_paths`), Repositorium
`MagicSetEditorPacks/M15-Magic-Pack`, Zweig `main`, Blob
`8c567ce7c22ff16e24d0ff940d9ff4a7f47cb892`.
**Erschlossen am:** 2026-08-15
**Sicherheit:** Vermutung — `json path:` steht in **einem** Paket aller
vorliegenden Packs (`magic.mse-game/json_paths`; die übrigen Treffer über
`full-magic-pack/data` sind Übersetzungen in `.mse-locale`), und der Quelltext
unserer Referenzfassung kennt den Schlüssel nicht (Suche über `src/` am
2026-10-10 leer). Was der Pfad in MSE bewirkt, ist erschlossen (F05, #2247).
**Testfall:** TestJedeGestaltWirdGelesen

---

## Die Kartenfelder des Spiels `magic`

Aus `card_fields` gelesen, jeweils als `card field:`-Block mit `name:`:

| Feld | Art | Alternativnamen |
| --- | --- | --- |
| `name` | text | — |
| `casting cost` | text | `mana cost`, `manacost`, `cost`, `cc` |
| `super type` | text | `supertype` |
| `sub type` | text | `subtype` |
| `type` | text | `typeline`, `type line` |
| `rarity` | choice | — |
| `rule text` | text | `rules text`, `rules`, `oracle text` |
| `flavor text` | text | `flavortext` |
| `power`, `toughness` | text | — |
| `pt` | text | `p/t`, `power/toughness` |
| `loyalty` | text | — |
| `illustrator` | text | `artist` |
| `card number` | text | — |
| `copyright` | text | — |

Die Möglichkeiten von `rarity`: `basic land`, `common`, `uncommon`, `rare`,
`mythic rare`, `special`, `masterpiece`.

**Quelle:** `magic.mse-game/card_fields` (rund 104 KB), Repositorium
`MagicSetEditorPacks/M15-Magic-Pack`, Zweig `main`, Blob
`f070ed1782a75e5f7f9b70688e952986072383db`.
**Erschlossen am:** 2026-08-15
**Sicherheit:** belegt — die Tafel ist aus **einem** Spielpaket gelesen und an
**zehn** unabhängig geschriebenen Sets gegengeprüft: jeder Kartenschlüssel der
zehn kanonischen Sets (`archive.org/details/mse-canon-cards-2022-10-24`) ist
ein Feld aus `card_fields` oder eine Angabe der Karte selbst (`stylesheet`,
`styling data`, `notes`, `time created`, …) — nachgezählt am 2026-10-10,
18 bis 62 Schlüssel je Set, keiner außerhalb (F05, #2247). Ein anderes Spiel führt
andere Felder; die Zuordnung in `internal/mtgjson` gilt deshalb je Spiel.
**Testfall:** TestZuordnungMagicBildetKartenfelderAb

---

## Manakosten stehen in **eckigen** Klammern, und das Phyrexianische heißt `H`

```
casting cost: [2][R]
casting cost: [2/R][2/R][2/R]
casting cost: [H/R]
```

**Drei Schreibweisen, und alle drei sind echt.** Wer Manakosten deutet, deutet
alle drei — und keine davon ist die, die man zuerst vermutet:

| Schreibweise | Wer sie schreibt |
| --- | --- |
| `[2][R]` | die **Sets** — eckige Klammern je Symbol |
| `{2}{R}` | Scryfall und MTG JSON |
| `2R` | MSEs `add_cards_scripts` beim Anlegen einer Karte |

**Das Phyrexianische steht vorn und heißt `H`**: `[H/R]` ist das, was CR 107.4f
als `{R/P}` schreibt — CR 107.4g nennt `{H}` als das Symbol ohne
Farbhintergrund. `[2/R]` ist das einfarbige Hybridsymbol aus CR 107.4e.

**Warum das zählt, ist eine Zahl**: beim Manawert gilt die **größere** Hälfte
eines Hybridsymbols (CR 202.3f). Wer die eckigen Klammern nicht kennt, zerlegt
`[2/R]` in fünf Stücke — und Flame Javelin kostet dann neun statt sechs. **Die
Farbe stimmt dabei die ganze Zeit**, und genau deshalb fällt es nicht als
Fehler auf.

**Quelle:** `Red.mse-set` aus
`~/Library/Application Support/MagicSetEditor4Mac/kanonische-sets/` — 3.730
Karten, Scryfall-Daten mit Stand 2022-10-24, Herkunft in
[0026](entscheidungen/0026-mse2-rendert-die-referenz-selbst.md). Gelesen über
`mse oeffnen` in einen eigenen Bestand.
**Erschlossen am:** 2026-09-17, bei der ersten Sichtprüfung des Spielkerns mit
einem echten Deck.
**Sicherheit:** belegt für **dieses** Set und die Sets desselben Bestands; die
Zeile über den Kartenfeldern darüber stammt aus dem **Spielpaket** und sagt
nichts über die Schreibweise in einem Set.
**Testfall:** TestManakostenInEckigenKlammernWerdenGedeutet

---

## Ein mehrzeiliger Wert steht unter seinem Schlüssel, nicht dahinter

```
card:
	notes:
		Sets: [pmoa];

		Formats: ;
	border_color: rgb(0,0,0)
```

Hinter dem Doppelpunkt steht **nichts**; der Wert folgt eine Ebene tiefer
eingerückt, und er endet an der ersten Zeile, die flacher steht.

```cpp
} else if (value.empty()) {
  // a multiline string
  readLine(true);
  while (indent >= expected_indent && !input.Eof()) {
    previous_value += line.substr(expected_indent);
```

**Wer ihn hinter den Doppelpunkt schreibt, zerlegt die Datei**: die zweite
Zeile steht dann ohne Einrückung und beendet den Block, in dem sie steht. Über
das kanonische `Vanguard` kostete das 16 von 18 Feldwerten je Karte — bei
grünem Rundlauf über die Saat, die keinen mehrzeiligen Wert trug
(#704).

**Quelle:** MSE2, `src/util/io/reader.cpp` Zeilen 258–272, Fassung f9d9356 vom
2021-07-06; und `Vanguard.mse-set` aus den kanonischen Sets.
**Erschlossen am:** 2026-08-21
**Sicherheit:** belegt.
**Testfall:** TestEinMehrzeiligerWertWirdMehrzeiligGeschrieben

**Und ein Wert, der eine Klammer offen lässt, steht ebenso als Block**
(C02, #2227,
gefunden von `FuzzSymbolzeichnung` am 2026-10-11). Für MSE2 ist das nicht
nötig — es liest einen Wert bis zum Zeilenende —, wohl aber für diesen
Zerteiler: er zählt Klammern, um Skriptausdrücke über mehrere Zeilen zu lesen,
und eine Zeile `name: Feuer [` machte jede folgende derselben Tiefe zu ihrer
Fortsetzung. Eine Karte verlor beim Wiederöffnen jedes Feld nach ihrem Namen.
Der Block ist die Form, die beide Leser gleich verstehen; gelesen wird er
überall, wo freier Text steht — Kartenfelder, Settitel, Paketnamen und der Name
eines Symbolteils.

**Und gelesen trägt eine Klammer, die ein Eintrag öffnet, nur tiefere Zeilen**
(#2293) —
ein Eintrag derselben Tiefe ist ihr Geschwister, wie bei MSE2. Bis dahin
verschluckte eine von MSE2 geschriebene Zeile mit offener Klammer ihre
Nachbarn: im vollen Pack 16-mal in `magic.mse-game`, wo
`match: [^(.,”":;\n—]+` (ein ASCII-`"` in der Zeichenklasse) das
`optional: false` des Parameters »Präfix« in zehn Sprachen und ein
`example:` in dreien zur Fortsetzung machte. **Gemessen am 2026-10-11** über
`full-magic-pack/data` und die zehn kanonischen Sets, Baum gegen Baum: 265
Zeilen wechseln die Art, keine davon in einem Set; alle übrigen liegen in
Skripten, und über **176.321** Blöcke mit leerem Wert ist `Skripttext`
vorher und nachher gleich.
**Quelle:** MSE2, `src/util/io/reader.cpp` Zeilen 207–223 und 258–272,
Fassung f9d9356 vom 2021-07-06.
**Erschlossen am:** 2026-10-11
**Sicherheit:** belegt.
**Testfall:** TestEinWertMitOffenerKlammerKommtZurueck, TestEineOffeneKlammerUeberstehtSichernUndOeffnen, TestEinEintragDerselbenTiefeIstEinGeschwister

---

## Die Aufschrift eines Kartenfeldes steht nicht unter `caption:`

Gezählt an `magic.mse-game/card_fields` des bezogenen `M15-Magic-Pack`:

| Angabe | Zahl |
| --- | --- |
| Kartenfelder | 249 |
| `caption:` | **0** |
| `localized caption:` | **0** |
| `card list name:` | 243 |
| `localized card list name:` | 25 — **jede davon mit `de:`** |
| `description:` | 249 |
| `localized description:` | **0** |

**Das einzige Deutsch, das ein Kartenfeld dieses Packs mitbringt, steht unter
`localized card list name:`** — „Name", „Manakosten", „Typ", „Seltenheit",
„S/W", „MV". Die dreizehn `.mse-locale`-Pakete daneben helfen nicht: sie
übersetzen MSEs **eigenes Fenster**, und keines von ihnen führt einen
`game:`-Abschnitt, unter dem Feldnamen stünden.

```cpp
if (caption.default_.empty()) caption.default_ = name_to_caption(name);
if (card_list_name.default_.empty()) card_list_name.default_ = capitalize(caption.default_);
```

`src/data/field.cpp`, Zeilen 69–70: bei MSE hängt der **Listenname an der
Aufschrift**, nicht umgekehrt. Ein deutsches MSE beschriftet sein Formularfeld
deshalb mit „Type" und schreibt „Typ" nur über die Spalte.

**Quelle:** `magic.mse-game/card_fields` des bezogenen Packs und MSE2,
`src/data/field.cpp` Zeilen 62–70, Fassung f9d9356 vom 2021-07-06.
**Erschlossen am:** 2026-08-21
**Sicherheit:** belegt.
**Testfall:** TestEinKartenfeldTraegtSeinenListennamen

---

## Zwei Zeichen in einem Kartenwert sind keine Zeichen

Ein Feldwert trägt neben seinen Marken **zwei Steuerzeichen**, und beide sehen
in einer Datei aus wie nichts:

| Zeichen | Heißt | Wo es herkommt |
| --- | --- | --- |
| `U+0001` | ein **getipptes** `<` | wer `3 < 4` schreibt; sonst begänne beim nächsten Lesen eine Marke |
| `U+EB00` | ein **verbindendes** Leerzeichen | das Auflösen eines Schlüsselworts |

```cpp
/// Escaped '<'
const Char ESCAPED_LANGLE = _('\1');
#define CONNECTION_SPACE _('\xEB00') // in private use area, untags to ' '

wxUniChar untag_char(wxUniChar c) {
  if (c == ESCAPED_LANGLE) return _('<');
  else if (c == CONNECTION_SPACE) return _(' ');
  else return c;
}
```

**Wer sie anzeigt, zeigt ein unsichtbares Zeichen** an der Stelle eines
gewöhnlichen; wer ein getipptes `<` **nicht** schützt, bricht den Text beim
nächsten Lesen ab.

**Quelle:** MSE2, `src/util/tagged_string.hpp` Zeile 20, `src/util/string.hpp`
Zeile 85, `src/util/tagged_string.cpp` Zeilen 14–22, Fassung f9d9356 vom
2021-07-06.
**Erschlossen am:** 2026-08-21
**Sicherheit:** belegt.
**Testfall:** TestGeschuetzteZeichenWerdenGezeichnet

Im Frontend prüfen `ohneauszeichnung.test.ts` und `mitauszeichnung.test.ts`
denselben Weg in beide Richtungen — der Wächter über diese Datei kennt nur
Go-Tests und kann sie nicht nennen.

---

## `save value: false` heißt: das Feld wird abgeleitet, nicht gespeichert

Drei Felder des Spiels `magic` tragen diese Angabe — `type`, `pt` und
`card number` —, und alle drei haben daneben ein `script:`, das sie aus anderen
Feldern zusammensetzt:

```
	name: pt
	save value: false
	script:
		combined_editor(field1: card.power, …, field2: card.toughness, …)
```

Wer beim Import dorthin schriebe, schriebe einen Wert, den MSE beim nächsten
Öffnen überschreibt — und ließe die Felder leer, aus denen er entsteht. Die
Zuordnung führt sie deshalb mit einem **Grund** statt einem Ziel.

**Quelle:** `magic.mse-game/card_fields`, Blöcke `type` (Zeile 961), `pt`
(Zeile 2304) und `card number` (Zeile 2959), Blob
`f070ed1782a75e5f7f9b70688e952986072383db`.
**Erschlossen am:** 2026-08-15
**Sicherheit:** belegt — im Quelltext entschieden, siehe den Nachtrag vom
2026-08-21 unten; hier stand bis zum 2026-10-10 noch „plausibel" (F05, #2247).
**Testfall:** TestUebergangenesFeldGibtKeinenBefund

Nachtrag vom 2026-08-21: **nachgelesen und weiter, als es hier stand.** Der
Quelltext hat es entschieden — `bool save_value; ///< Should the value of this
field be written to the file?` (`src/data/field.hpp`, Fassung f9d9356 vom
2021-07-06), im Konstruktor auf `true`. Die Sicherheit oben lautete
„plausibel"; sie ist jetzt belegt.

Und die Zahl war zu klein: über `magic.mse-game` tragen **46** der 249
Kartenfelder die Angabe, nicht drei — gezählt am eingebundenen Baum, nicht am
Kopf der Datei. Vier davon nennen ihre Teile ablesbar, darunter das wichtigste:

```
	name: text
	save value: false
	script:
		if mainframe_walker() then mainframe_walker_text_script()
		else if alt_text() then alt_text_script()
		else combined_editor(field1: card.rule_text,
		                     separator: "<line>\n</line>", field2: card.flavor_text)
```

**Der Aufruf steht im dritten Zweig**, nicht als ganzer Ausdruck — wer ihn nur
an der Wurzel suchte, fände ihn an keinem einzigen Feld. `pt` und `type`
schreiben `separator: language().pt_separator`; ihr Trenner steht erst nach dem
Rechnen fest, und ihr Rückweg bleibt deshalb zu.

**Quelle:** MSE2 (`src/data/field.hpp`) und `magic.mse-game/card_fields` des
bezogenen `M15-Magic-Pack`.
**Nachgelesen am:** 2026-08-21
**Sicherheit:** belegt.
**Testfall:** TestEinFeldSagtObEsGespeichertWird, TestVerbundAus

Nachtrag vom 2026-08-15: an einer **bezogenen** Datei (*Limited Edition Alpha*,
295 Karten) führt eine MTGJSON-Karte **34** Schlüssel, die die Zuordnung für
`magic` nicht kennt — `uuid`, `identifiers`, `legalities`, `purchaseUrls`,
`foreignData` und weitere. Die meisten haben in MSE kein Gegenstück; sie reisen
roh mit.

---

## Manakosten stehen in MSE ohne geschweifte Klammern

MTG JSON schreibt `{2}{R}`; MSE schreibt `2R`.

```
new_card([casting_cost: "W", rarity: "uncommon"])
```

**Quelle:** `magic.mse-game/add_cards_scripts`, Blöcke *Common Cycle*,
*Uncommon Cycle*, *Rare Cycle*, Blob
`4ef298ee92d96df363f5ef0926c60a7539a97bbd`; dieselbe Datei belegt mit
`super_type: "Basic Land", sub_type: "Plains"`, dass `super type` Ober- **und**
Kartentyp zusammen trägt.
**Erschlossen am:** 2026-08-15
**Sicherheit:** belegt — an geschriebenen `.mse-set` nachgezählt am 2026-10-10:
über sieben kanonische Sets (White, Blue, Black, Red, Green, Multicolor,
Colorless) stehen **25.885** `casting_cost`-Werte, und **einer** trägt eine
geschweifte Klammer (`{HW}` in White). Hybride und phyrexianische Kosten
stehen in **eckigen** Klammern oder mit Schrägstrich (`[1][H/W]`, `[2/U]`,
`3W/U`, `UU/RR`) — siehe „Manakosten stehen in **eckigen** Klammern"
(F05, #2247).
**Testfall:** TestZuordnungMagicBildetKartenfelderAb

---

## MTG JSON hat kein maschinenlesbares Schema

Nachgesehen am 2026-08-15: weder [mtgjson.com](https://mtgjson.com) noch
`mtgjson/mtgjson` auf GitHub veröffentlichen eine Schemadatei. Die Datenmodelle
stehen als Dokumentation da (`mtgjson.com/data-models/card/card-set/`), die
Suche nach `*.schema.json` im Erzeugerrepositorium bringt null Treffer.

**Was daraus folgt, steht in `internal/mtgjson/gueltigkeit.go`:** eine Tabelle
der bekannten Kartenschlüssel samt ihrer JSON-Art, und **derselbe Prüfer läuft
hinter dem Build-Tag `fremdpakete` gegen eine bezogene Datei**. Wäre die Tabelle
erfunden, fiele die echte Datei durch. An *Limited Edition Alpha* (295 Karten)
sind 35 der 36 Einträge bestätigt.

**Quelle:** `mtgjson.com/data-models/card/card-set/`, `mtgjson.com/data-models/set/`,
GitHub-Codesuche `repo:mtgjson/mtgjson filename:*.schema.json` (0 Treffer), und
eine bezogene Datei `AllPrintings`-Einzelset.
**Erschlossen am:** 2026-08-15
**Sicherheit:** Vermutung — dass es **kein** Schema gibt, ist nachgesehen; dass
die Tafel in `gueltigkeit.go` die Schlüssel trifft, ist an **einer** bezogenen
Datei geprüft, und ein Set mit Planeswalkern oder doppelseitigen Karten führt
Schlüssel, die dort nicht vorkommen. Eine zweite Datei liegt hier nicht vor
(F05, #2247).
**Testfall:** TestDerPrueferWeistUngueltigesAb

---

---

## Wie ein Stil seine Zeichenanweisungen ausdrückt

Ein `.mse-style` sagt, **wo** auf der Karte ein Feld steht und **wie** es
gezeichnet wird. Der Kopf nennt die Kartengröße, `card style:` die Felder.

### Die Kartengröße steht im Kopf — samt Auflösung

```
card width: 375
card height: 523
card dpi: 150
```

**Damit ist die physische Größe berechenbar**: 375 ÷ 150 = 2,5 Zoll = 63,5 mm;
523 ÷ 150 = 3,487 Zoll = 88,6 mm. Eine Karte misst also nicht 63 × 88, sondern
63,5 × 88,6 — die Größe, die auch das echte Kartenspiel hat.

Das ist der Punkt, an dem
[0012](entscheidungen/0012-aufloesung-als-zielbreite.md) sich selbst
widerlegt: er lehnt eine dpi-Angabe ab, weil keine Vorlage sagt, wie groß eine
Karte in Millimetern ist. **Sie sagt es.**

### `card style:` bildet Feldnamen auf Darstellungsangaben ab

```
card style:
	name:
		left: { if card.card_symbol == "none" then 32 else 50 }
		top: 30
		right:  { 341 - card_style.casting_cost.content_width }
		height: 23
		alignment: bottom shrink-overflow
		padding bottom: 0
		z index: 1
		font:
			name: Beleren Bold
			size: 16
			color: black
```

- Der Schlüssel ist der **Feldname des Spiels** (`magic.mse-game/card_fields`),
  nicht ein eigener Name.
- Stelle und Größe stehen als `left`, `top`, `width`, `height` — **oder** als
  `right` und `bottom`. `name:` oben führt `left`, `right` und `height`: die
  Breite ergibt sich. Ein Leser, der `width` voraussetzt, liest dieses Feld
  falsch.
- **Jede Zahl darf ein Skript sein.** `left: 0` und `left: { … }` stehen
  nebeneinander in derselben Datei.
- `z index` gibt die Zeichenreihenfolge; sie ist **nicht** die Reihenfolge im
  Text (`card color` hat 0 und steht vor `name` mit 1, aber `border color` hat 4
  und steht ganz oben).
- `alignment` trägt mehrere Wörter (`bottom shrink-overflow`, `middle right`).
- `font:` ist ein Unterblock mit `name`, `size`, `color`.
- `mask`, `render style`, `image`, `default` und `symbol font` — siehe den
  eigenen Abschnitt darunter.

**Quelle:** `M15-Magic-Pack@3f7f8ce/data/magic-m15.mse-style/style`, Zeilen
27–29 (Kartengröße und dpi) und 188–245 (`card style:`);
`…/magic-m15-planeswalker.mse-style/style`, Zeilen 23–25 und 170–197 — dieselben
Schlüssel in einem anderen Stil desselben Packs.
**Erschlossen am:** 2026-08-15
**Sicherheit:** belegt für **Magic**. Alle Belege stammen aus einem Pack für ein
Spiel; ob Yu-Gi-Oh! oder Kamen Rider dieselben Schlüssel führen, ist damit
**nicht** gezeigt. Die Kopfzeilen (`card width`, `card height`, `card dpi`)
stehen in beiden geprüften Stilen an derselben Stelle und in derselben Form.
**Testfall:** TestDieKartengroesseKommtAusDemKopf, TestDerFeldstilNenntStelleUndGroesse, TestFremdeStile

### `render style:`, `image:`, `default:`, `mask:` und `symbol font:`

Ausgezählt über **alle 190 `style`-Dateien** von `M15-Magic-Pack`, getrennt
nach dem Block, in dem eine Angabe steht, und nach ihrer Einrückungstiefe.
Beides ist nötig: derselbe Schlüssel bedeutet an zwei Stellen zwei
verschiedene Dinge.

**`render style:` steht in `card style:` und in `styling style:` — und die
Werte sind nicht dieselben.**

| Block | Wert | Vorkommen |
| --- | --- | --- |
| `card style` | `image` | 485 |
| `card style` | `hidden` | 4 |
| `styling style` | `both` | 230 |
| `styling style` | `checklist` | 28 |

Im **Kartenstil** gibt es also nur zwei Werte. `image` heißt „dieses Feld
zeichnet ein Bild", `hidden` heißt „gar nicht zeichnen" — die vier Fälle sind
`card color` von Flip-Karten, deren Grund die Gegenseite malt. `both` und
`checklist` gehören zum **Stilmenü**: dort steht neben `render style: both`
regelmäßig ein `choice images:`-Block, und beides zusammen sagt, wie eine
Auswahl im Menü aussieht — Bild **und** Beschriftung. Wer `both` für einen
Kartenstilwert hält, sucht eine Zeichenregel, die es nicht gibt.

**`image:` steht nie ohne `render style: image`.** Über alle Felder aller
Stile ausgezählt:

| `image:` | `mask:` | `render style: image` | `symbol font:` | Felder |
| --- | --- | --- | --- | --- |
| — | — | — | — | 835 |
| — | — | — | ✓ | 435 |
| ✓ | — | ✓ | — | 268 |
| — | ✓ | — | — | 241 |
| — | — | ✓ | — | 127 |
| ✓ | ✓ | ✓ | — | 88 |
| — | ✓ | — | ✓ | 20 |
| — | ✓ | ✓ | — | 2 |

Die Zeile, die es **nicht** gibt, ist die Aussage: kein Feld trägt `image:`
ohne `render style: image`. Umgekehrt schon — 129 Felder sagen `image` und
nennen keins. Es sind `rarity`, `type symbol` und ihresgleichen: ihr Bild kommt
aus den **Wahlbildern des Spiels** (`choice images:` im `.mse-game`), nicht aus
dem Stil.

**Beide Formen stehen nebeneinander, in jeder Angabe.** Ein Wert ist entweder
ein Dateiname oder ein Skript:

| Angabe | Skript `{ … }` | einfacher Name | absoluter Paketpfad |
| --- | --- | --- | --- |
| `mask:` | 174 | 173 | 5 |
| `image:` | 354 | 2 | — |
| `default:` | 63 | 8 | — |

Ein einfacher Name meint eine Datei **im eigenen Paket**
(`mask: border_mask.png` neben `style` in `magic-m15.mse-style`); ein Name mit
führendem `/` meint ein anderes
(`mask: /magic-genevensis-00-main.mse-style/pride_mask.png`). Wer die erste Form
roh weiterreicht, sucht sie in der Paketwurzel und findet sie nie.

**Und eine dritte Form: `{` in der Mitte.** MSE2 fragt nicht, ob der Wert
mit einer Klammer **beginnt**, sondern ob er **irgendwo** eine trägt — dann
liest es ihn als Zeichenkette mit eingebetteten Ausdrücken, ohne
Anführungszeichen, und `\{` ist eine Klammer im Namen:

```
		mask: border-mask{pt()}.png
```

ist `border-mask` + `pt()` + `.png`, und `pt` ist in `magic-crystal`
`{ if card.pt != "" then "-pt" else "" }` — die Datei heißt `border-mask.png`
oder `border-mask-pt.png`. Gemessen am 2026-09-29 über `full-magic-pack`: von
12.739 Bildwerten tragen 70 eine Klammer in der Mitte, 68 davon innerhalb
eines Skriptaufrufs (benannte Argumente von `masked_blend(…)`) und **zwei**
als eigener Schlüssel, beide in `magic-crystal`. Bis
#1934
suchte dieses Programm die Datei `border-mask{pt()}.png`, und der schwarze
Rand der Karte fehlte.

Ein **ganz** geklammerter Wert ist hier ein Skript; MSE2 läse auch ihn im
Zeichenkettenmodus, und bei einem einzigen Ausdruck ist das Ergebnis
dasselbe. Gemessen: 5.191 Bildwerte beginnen mit `{` und enden auf `}`, keiner
besteht aus mehr als einem Ausdruck. Die einzeilige Form `image: script: …`
steht nirgends; der `script:`-Unterblock kommt geklammert an.

**Quelle:** MSE2, `Reader::handle(ScriptableImage&)`, `src/script/image.cpp`
Zeilen 67–78, und der Zeichenkettenmodus des Zerteilers,
`src/script/parser.cpp` Zeilen 139–143 und 249–260, Fassung f9d9356 vom
2021-07-06; Zählung über `full-magic-pack`.

**Ein leeres Ergebnis ist die Regel, nicht der Fehler.**
`image: {if styling.overlay == "" then nil else styling.overlay + "/overlay.png"}`
steht 99-mal: keine Überlagerung gewählt heißt keine Ebene. Ein Befund dafür
machte aus jeder zweiten Karte eine Karte mit Vorbehalten.

**`default:` ist nicht `image:`, und es steht nur an einer Art Feld.** Über
alle 190 Stile trägt es **genau zwei** Feldnamen — `image` und `image 2` — und
**nie** ein Feld, das auch `image:` führt. Es ist damit kein zweites
Bildskript, sondern der Marker für ein Feld, an dem die **Karte** ihr Bild
mitbringt: `default: {default_image(card.card_color)}` ist der Ersatz für eine
Karte, die noch keines hat.

Wo eine Karte eines trägt, liegt die Datei im **Set**-Paket, nicht im
Stilpaket. Dieser Weg ist nicht erschlossen.

**`symbol font:` ist ein eigener Unterblock neben `font:`, mit denselben
Schlüsseln.** 455 Felder führen ihn; `casting cost` in `magic-m15.mse-style`
trägt beide untereinander:

```
		font:
			name: MPlantin
			size: 15
		symbol font:
			name: magic-mana-large
			size: 15
			alignment: middle right
```

Wer beide in dasselbe Feld liest, zeichnet den Text in `magic-mana-large` —
einer Schrift, die es auf keinem Rechner gibt. Denn der Name ist **kein
Schriftname, sondern ein Paket**: `magic-mana-large.mse-symbol-font` ist ein
Verzeichnis aus PNG-Dateien, je eine je Symbol, mit einer `symbol-font`-Datei,
die `code:` auf `image:` abbildet. `alignment:` hat in `font:` keine
Entsprechung; es sagt, wo ein Symbol im Textfluss sitzt.

**Quelle:** alle 190 `*.mse-style/style` aus `M15-Magic-Pack@3f7f8ce/data`,
maschinell ausgezählt; die Beispiele aus `…/magic-m15.mse-style/style`
(`card color`, `casting cost`, `image`, `indicator`) und
`…/magic-mana-large.mse-symbol-font/symbol-font`.
**Erschlossen am:** 2026-08-15
**Sicherheit:** belegt für **Magic**. Die Auszählung deckt ein Pack für ein
Spiel; dass ein Yu-Gi-Oh!-Stil dieselben Schlüssel führt, ist damit nicht
gezeigt. **Nicht erschlossen** ist, wie eine Symbolschrift den Text in Symbole
zerlegt — welche Zeichenfolge welchen `code:` trifft und was mit unbekannten
geschieht.
**Testfall:** TestEinFeldstilNenntStandardbildUndSymbolschrift,
TestEinBildfeldKommtAlsBildebeneInDenPlan, TestEineMaskeKommtAnDieEbene,
TestEinVerborgenesFeldWirdNichtGezeichnet, TestWasNichtGedeutetIstWirdBenannt

**TestFremdeStile** steht hinter dem Build-Tag `fremdpakete`: er liest die
bezogenen Pakete selbst, läuft nicht in der Pipeline und zählt nicht zur
Abdeckung (§3). Er hält die Feststellung gegen **190** Stile statt gegen zwei.

---

## Die Ordnungszeichenkette von `sort_text`

`sort_text(order: "…", text)` ordnet die **Zeichen** eines Textes. Die Ordnung
ist eine kleine Sprache:

| Form | Bedeutung |
| --- | --- |
| `WUBRG` | jedes Zeichen ein eigener Rang, in der genannten Reihenfolge |
| `[0123456789]` | **ein gemeinsamer** Rang: die Eingabereihenfolge bleibt |
| `<WUBRG>` | filtern und **entdoppeln**: nur diese Zeichen, jedes einmal |
| `\?` | `?` als gewöhnliches Zeichen |
| `(WUBRG)`, `cycle(WUBRG)` | ein **Kreis**: die Drehung mit dem kürzesten Weg |
| `ordered(…)`, `mixed(…)`, `once(…)`, `compound(…)`, `pattern(…)`, `in_place(…)`, `any()` | kennt MSE, hier **nicht gebaut** und gemeldet |

Entschieden hat das nicht die Plausibilität, sondern was die Vorlagen selbst
über ihre Aufrufe sagen:

- **`<…>` filtert und entdoppelt.** `color_filter := sort_text@(order: "<WUBRG>")`
  steht unter dem Kommentar `#### color based on mana cost, input == a mana
  cost`; `tap_filter := sort_text@(order: "<TQ>")` unter `#### only allow
  tickets`; `color_filterH := sort_text@(order: "</>")` fragt nach dem
  Schrägstrich einer Hybridkosten. Den Ausschlag gibt `mana_is_wedge`: es zählt
  das Ergebnis von `sort_text(order:"<WUBRG>")` und vergleicht mit **drei** —
  bei `WWU` dürfen das nicht drei sein.
- **`[…]` ist ein gemeinsamer Rang.** `mana_unsort := sort_text@(order:
  "[/\?XYZI0123456789VHSCAIEKPLOMWUBRG]")` — der **Name** entscheidet: eine
  Ordnung, in der alles in einer Klammer steht, sortiert nicht. Dazu der
  Zweck von `[0123456789]`: ohne gemeinsamen Rang würde aus `10` ein `01`.

**Die runde Klammer ist ein Kreis.** Bis zum 2026-08-16 stand hier „nicht
erschlossen, wird gemeldet statt geraten" — mit der Vermutung, es könnte eine
zyklische Gruppe sein, und der Feststellung, dass eine gewöhnliche Folge die
Daten genauso erklärt. Die Vermutung war richtig, und **jetzt ist sie
nachgelesen**.

Die genannten Zeichen bilden einen Kreis im Uhrzeigersinn. Ausgegeben wird die
Drehung, die

1. den kürzesten Weg über den Kreis nimmt,
2. ihre **Lücken früh** hat — eine Lücke ist ein Zeichen der Ordnung, das in der
   Eingabe fehlt,
3. bei Gleichstand am weitesten vorn im Ausdruck beginnt.

MSE nennt selbst drei Beispiele mit `XYZ<0123456789>cycle(WUBRG)`: `W1G` →
`1GW`, `GRBUWWUG` → `WWUUBRGG`, `WUR` → `RWU`. `(…)` und `cycle(…)` sind
dieselbe Sache.

**Das war der Rahmen.** `magic-m15.mse-style` kommt über `card color → image`
und drei Ecken auf `sort_text` mit `(wubrg)`; ohne diese Ordnung meldete die
Angabe einen Befund, und das Feld, das den Rahmen zeichnet, fiel aus
(#240).

**Was MSE sonst noch kennt, ist nicht gebaut** — `ordered(`, `mixed(`, `once(`,
`compound(`, `pattern(`, `in_place(`, `any()`. Keines steht in
`M15-Magic-Pack`; sie zu bauen wäre Vorrat ohne Fundstelle. Sie werden gemeldet,
denn als Buchstabenfolge gelesen ergäben sie eine Ordnung, die plausibel
aussieht und falsch ist.

**Quelle:** `src/util/spec_sort.hpp` Zeilen 15–43 und `src/util/spec_sort.cpp`
Zeilen 135–167 von twanvl/MagicSetEditor2 — **übernommen, nicht erschlossen**.
Die Fundstellen in den Paketen: `magic.mse-game/script` 509–546 und 639–640,
`magic-m15-extra-udelude.mse-style/game_scripts` 24–51,
`magic-m15-planeswalker-extra-purple.mse-style/game_scripts` 21–27
**Erschlossen am:** 2026-08-16
**Sicherheit:** belegt
**Testfall:** TestSortTextOrdnetNachDerAngegebenenReihe, TestSortTextMitSpitzerKlammerFiltertUndEntdoppelt, TestEinZyklusOrdnetNachDemKuerzestenWeg, TestEinUngebautesSchluesselwortWirdGemeldet

---

## Was in keiner Ordnung steht, **fällt weg**

`sort_text` mit einer Ordnung behält **nur**, was die Ordnung nennt.

```cpp
} else { // single char
  size_t count = count_and_remove(it.value, input);
  ret.append(count, it.value);
}
…
return ret;
```

`spec_sort`, `src/util/spec_sort.cpp`, Zeilen 266–352; der Aufrufer in Zeile 355
gibt `ret` zurück und wirft `input` weg. Was die Schleife nicht angefasst hat,
ist damit fort.

**Hier stand bis zum 2026-08-18 das Gegenteil**, und es stand ausdrücklich als
**Vermutung** da: „gewählt ist die Lesart, die nichts verliert". Die Wahl war
falsch, und sie war sichtbar — ein kanonisches MSE2-Set speichert Manakosten als
`[3][W]`, die Klammern liefen durch `mana_filter → sort_text`, blieben stehen
und wurden als Symbole gezeichnet: **vier leere Kästen** neben den Manakosten,
auf jeder gewöhnlichen Karte
(#411).

**Die Begründung von damals war nicht falsch, nur unbelegt**: „ein stiller
Verlust ist der teurere Irrtum" — das gilt weiter als Faustregel und ersetzt
kein Nachlesen (§15, [0024](entscheidungen/0024-portierung-statt-neuerfindung.md)).

**Quelle:** MSE2, `src/util/spec_sort.cpp`, Zeilen 266–352 und 355, Fassung
f9d9356 vom 2021-07-06
**Erschlossen am:** 2026-08-18
**Sicherheit:** belegt
**Testfall:** TestSortTextVerwirftWasDieOrdnungNichtNennt
---

## Nicht jede Datei in einem Paket ist ein Baum

Die **Hauptdatei** eines Pakets ist ein `schlüssel: wert`-Baum. Was daneben
liegt, ist, was es ist — und das ist oft kein Baum:

| Datei | Inhalt |
| --- | --- |
| `magic.mse-game/script` | ein Skriptrumpf: `crop_safe :=`, darunter `{` auf der **Wurzelebene** |
| `magic.mse-game/unit_tests` | freistehende Anweisungen: `assert(card_color_color_count("") == 0)` |
| `magic-spoiler.mse-export-template/script js` | JavaScript |
| `magic-spoiler.mse-export-template/style css` | CSS |

Eine Zeile ohne Doppelpunkt ist deshalb **auch auf der Wurzelebene Text** und
kein Fehler. Ihre Bytes bleiben erhalten, und der Rundlauf trägt sie unverändert
wieder hinaus.

**Wann eine Datei überhaupt ein Baum ist, entscheidet nicht sie selbst, sondern
wer sie liest.** Es gibt kein Kennzeichen im Inhalt: `include file:` holt eine
Nebendatei in einen Skriptblock, und dort ist ihr Inhalt Skript. Die Wahl der
Namen belegt es — `script`, `unit_tests`, `new-blends`, `masterpiece_script`,
`style css` — sie folgen keiner Regel, die man ablesen könnte.

**Quelle:** `M15-Magic-Pack@3f7f8ce/data` — 324 Pakete, 567 Textdateien im
Rundlauf. Zwölf davon scheiterten an der alten Regel, und keine war eine
Hauptdatei: `magic.mse-game/{script,language,statistics_script,unit_tests}`,
`magic-modules.mse-include/loyalty/{init_script,init_script_dfc}`,
`magic-m15-{invocation,sketch}.mse-style/masterpiece_script`,
`magic-blends.mse-include/new-blends`,
`magic-lackeybot.mse-export-template/script`,
`magic-spoiler.mse-export-template/{script js,style css}`.
**Erschlossen am:** 2026-08-15
**Sicherheit:** belegt — vier Paketarten (`mse-game`, `mse-include`, `mse-style`,
`mse-export-template`) von verschiedenen Urhebern.
**Testfall:** TestEineWurzelzeileOhneDoppelpunktIstText

Die Entscheidung samt ihrem Preis steht in
[0014](entscheidungen/0014-eine-mse-datei-ist-nicht-notwendig-ein-baum.md): der
Zerteiler kann eine Hauptdatei mit verlorenem Doppelpunkt nicht mehr abweisen.

---

## Formen von MSE-Script, die echte Pakete schreiben

Erschlossen beim Auswerten der `init script:`-Blöcke aller 190 Stile von
`M15-Magic-Pack`. Jede dieser Formen hat den Zerteiler abgewiesen, und jede
steht in echten Vorlagen:

| Form | Beispiel | Betroffene Initskripte |
| --- | --- | ---: |
| Ein `if` im **`then`**-Zweig | `if color_count == 0 then if land then l else c` | 53 |
| Ein **Semikolon** als Trenner | `{ a := 1; b := 2 }` | ~60 |
| Eine **Zahl** hinter dem Punkt | `default_images[colors.0]` | 53 |
| Eine **Liste ohne Namen** | `[[left: 0, top: 0], […]]` | 33 |
| Eine **Zuweisung** in einem Zweig | `if card.loyalty != "" then output := output + […]` | 31 |

**Eine Liste ist eine Abbildung mit Zahlen als Namen** — das ist die
Feststellung hinter zwei Zeilen dieser Tabelle: `colors.0` greift auf das erste
Glied zu, und `[a, b]` legt die Namen `0` und `1` an. Wer beide Formen getrennt
behandelt, braucht zwei Regeln für eine Sache.

**Das `else` bindet an das nächstgelegene `if`.** Sonst läse sich die erste
Zeile der Tabelle als „wenn `color_count == 0`, dann (wenn `land` dann `l`),
sonst `c`", und die `else if`-Kette darunter hinge am falschen Zweig.

**Quelle:** die `init script:`-Blöcke von `M15-Magic-Pack@3f7f8ce/data/*.mse-style`,
maschinell ausgewertet; die Beispiele aus `magic-DayNight.mse-style`,
`magic-genevensis-90-counter.mse-style`,
`magic-m15-compleated-planeswalkers.mse-style` und `magic-m15.mse-style`.
**Erschlossen am:** 2026-08-15
**Sicherheit:** belegt — jede Form an mehreren Paketen verschiedener Urheber.
**Nicht erschlossen** ist, was MSE bei einer Zuweisung in einem Zweig mit dem
Gültigkeitsbereich tut; hier gilt die Umgebung des Aufrufers.
**Testfall:** TestDerThenZweigDarfSelbstEineBedingungSein,
TestEinSemikolonTrenntWieEinUmbruch, TestNachDemPunktDarfEineZahlStehen,
TestEineListeIstEineAbbildungMitZahlen, TestEinZweigDarfEineZuweisungSein

---

## Ein Stil bringt seine eigene Skriptumgebung mit

`init script:` im Kopf eines `.mse-style` legt an, was die Feldskripte benutzen
— `template`, `mana_sort`, `module_card`. Es beginnt in aller Regel mit
`include file:`-Zeilen, die Module aus `.mse-include`-Paketen einsetzen.

**Ein Modul bringt mehr als Skript mit.** Nach dem Einbinden trägt der Kopf
eines Stils auch die `card style:`-Blöcke seiner Module: aus 2016 Feldstilen
über alle 190 Stile wurden 5950.

**Quelle:** `magic-m15.mse-style/style`, Zeilen 30–79 (`init script:` mit
`include file: /magic-default-image.mse-include/scripts`);
`magic-modules.mse-include/loyalty/init_script`.
**Erschlossen am:** 2026-08-15
**Sicherheit:** belegt — an allen 190 Stilen des Packs ausgeführt; 80 von 189
Initskripten laufen vollständig durch, der Rest scheitert an Namen aus dem
**Spiel**paket (#195).
**Testfall:** TestEinStilBringtSeinInitskriptMit, TestEinModulLandetImInitskript

---

## Eine Exportvorlage ist ein Programm mit Wahlmöglichkeiten

```
game: magic
file type: *.xml|*.xml|*.*|*.*
create directory: true
option field:
	type: choice
	name: export Images
	description: Export images along with the XML?
	choice: no
	choice: JPG
	initial: no
script:
	write_cards := to_text(for each card in cards do write_card())
	to_string(write_cards)
```

**Der Wert des letzten Ausdrucks ist die Ausgabe.** Alle 19 Vorlagen von
`M15-Magic-Pack` schreiben so.

| Schlüssel | Bedeutung | Vorkommen |
| --- | --- | ---: |
| `script:` | das Programm | 19 |
| `game:` | für welches Spiel | 19 |
| `option field:` | eine Eingabe, als `options.<name>` lesbar | 78 |
| `file type:` | Filter des Dateidialogs, `*.xml\|*.xml\|*.*\|*.*` | 18 |
| `create directory:` | Verzeichnis statt einer Datei | 16 |
| `option style:` | Darstellung einer Wahlmöglichkeit | 2 |

**Ein `option field:` schreibt seine Vorgabe zweimal verschieden**: eine
Auswahl als `initial:`, ein Text als `default:`. Beide sagen dasselbe.

**Der Name wird mit Unterstrichen gelesen** — `options.export_Images` für
`name: export Images`. Dieselbe Regel wie bei den Kartenfeldern.

**Eine Vorlage kann nicht aus dem Zielverzeichnis ausbrechen**, und zwar aus
einem stärkeren Grund als einer Prüfung: **die Sprache kann keine Datei
anfassen.** Sie gibt Text zurück; wohin er geschrieben wird, entscheidet der
Aufrufer. **Gehalten wird das strukturell**: `internal/msescript` importiert
nichts, womit sich eine Datei, ein Prozess oder eine Verbindung anfassen ließe
(`schichten.TestDieSpracheKannNichtsAnfassen`, #227). Der Namensfilter in
`service` daneben ist ein **Hinweis** und keine Grenze — er sagt nur, dass die
Namen der Bibliothek nicht so tun, als könnten sie etwas anfassen.

**Quelle:** alle 19 `*.mse-export-template/export-template` aus
`M15-Magic-Pack@3f7f8ce/data`, maschinell ausgezählt; die Beispiele aus
`magic-cockatrice-v2` und `magic-text`.
**Erschlossen am:** 2026-08-15
**Sicherheit:** belegt — 19 Vorlagen verschiedener Urheber.
**Noch nicht lauffähig:** keine der 19 kommt bis zu ihrer Ausgabe; was der
Sprache dafür fehlt, ist gezählt in
#202.
**Testfall:** TestEineExportvorlageWirdGelesen, TestEineExportvorlageNenntIhreWahlmoeglichkeiten, TestEineExportvorlageErzeugtEineAusgabe

---

## MSE-Script kennt zwei Schleifen

```
for each card in sort_list(cards, order_by: {input.card_number}) do write_card()
for x from 0 to length(cmds)-1 do (…)
```

Die erste geht eine Liste durch, die zweite zählt — **einschließlich der oberen
Grenze**: `to length(cmds)-1` sagt es, denn wer bis `length` zählte und die
letzte Stelle ausließe, schriebe das `-1` nicht dazu.

**Beide ergeben eine Liste**, also eine Abbildung mit Zahlen als Namen; `to_text`
darüber setzt die Glieder aneinander, und genau so benutzen die Vorlagen sie.

**Der Schleifenname gilt nur im Durchgang.** Sonst sähe die zweite Schleife den
letzten Wert der ersten.

**Quelle:** `magic-text.mse-export-template/export-template`,
`magic-field-test.mse-export-template/export-template` Zeile 252,
`magic.mse-game/statistics_script` Zeile 16.
**Erschlossen am:** 2026-08-15
**Sicherheit:** belegt — drei Dateien, zwei Paketarten.
**Testfall:** TestForEachGehtEineListeDurch, TestForFromZaehltEinschliesslich

---

## Noch offen

Erkannt und **nicht** erschlossen:

| Punkt | Wo |
| --- | --- |
| Zeichenkodierung im Detail — die BOM sagt UTF-8, aber alte Pakete? | **beantwortet am 2026-09-23**: nur UTF-8, wie MSE2 — siehe „Drei Zeilenenden, und nur UTF-8" |
| Mehrzeilige Werte und ihre Fortsetzungsregel | offen |
| Aufbau von `set` in einem **echten** `.mse-set` | offen — **kein einziges Set gesichtet**, alle fünf Belegpakete sind Spiele und Stile |
| Ob es Schleifen gibt | offen — in fünf Dateien keine gesehen |
| Welche Auszeichnungen es in Kartentexten gibt | **teilweise beantwortet am 2026-08-16** — gezählt in `magic.mse-game`: 2.165 `<atom-param>`, 108 `<sym>`, 16 `<i-auto>`, dazu `<atom-sep>` und `<atom-autorem>`; im Pack außerdem `<i-flavor>`. **Was sie bedeuten, ist offen** (#312) — entfernt werden sie, gezeichnet nicht |
| ~~Was die **runde Klammer** in der Ordnung von `sort_text` bedeutet~~ | **beantwortet am 2026-08-16** — sie ist ein Kreis, nachgelesen in MSE2; der Abschnitt „Die runde Klammer ist ein Kreis" weiter oben führt die Regel samt der drei Beispiele. Die Zeile stand danach weiter als offen hier und ist vom Zweitblick über #274/#275/#277/#278 angestrichen worden |
| Die vollständige Standardbibliothek von MSE-Script | #25 |
| Ob `match:` in `contains` ein Muster ist oder eine feste Zeichenkette | #25 — hängt am Rückstrich-Befund oben |
| ~~Was `to_real` bei einem unlesbaren Wert tut~~ | **beantwortet am 2026-10-11** (F02, #2244) — ein **verzögerter** Fehler, den `or else` fängt; `""` und `nil` sind 0, `true` ist 1, eine Farbe das Mittel ihrer Kanäle. Nachgelesen in `src/script/functions/basic.cpp`, Zeilen 135–159, Fassung f9d9356 vom 2021-07-06, und am Orakel gefragt. Bis dahin brach dieses Programm mit Stelle ab |
| ~~Wie eine **Symbolschrift** einen Text in Symbole zerlegt~~ | **beantwortet am 2026-08-16** (#275, #173 geschlossen) — längster Code zuerst, der Abschnitt oben führt die Regel; ein unbekanntes Zeichen wird **benannt**, nicht übersprungen |
| Wie eine Symbolschrift **Fassungen desselben Codes** unterscheidet | #293 — `magic-mana-large` führt `T` und `Q` je dreimal und trennt sie über `enabled:`; unser Modell nimmt den ersten Treffer |
| Woher ein Feld mit `render style: image` **ohne** `image:` sein Bild nimmt | offen — 129 Felder sind so (`rarity`, `type symbol`); die Spur führt zu `choice images:` im `.mse-game`, nachgemessen ist sie nicht |
| Ob andere Spiele dieselben Stilschlüssel führen | offen — alle Belege stammen aus einem Magic-Pack |
| ~~Ob `stylesheet:` den vollen Paketnamen trägt oder eine Kurzform~~ | **beantwortet am 2026-08-16** — die Kurzform; der Abschnitt „Ein `stylesheet:` trägt die Kurzform" führt die Regel samt Quelle. Sie stand hier keine Stunde lang offen |
| Der Aufbau einer `.mse-export-template` | #51 |
| Wie MSE hybride und phyrexianische Manakosten schreibt (`{W/U}`, `{U/P}`) | offen — die Zuordnung lässt den Klammerinhalt zusammen (`W/U`), gesehen ist das nicht; beim Schreiben geht er als **eine** Klammer hinaus |
| Ob ein MTG-JSON-Untertyp aus mehreren Wörtern bestehen kann (`Time Lord`) | offen — beim Schreiben wird am Leerzeichen getrennt; der Rundlauf hält, die Bedeutung wäre falsch |
| Was ein `json path:` in MSE genau auslöst | offen — der Pfad ist gelesen, seine Wirkung nicht |

**Der Punkt mit dem größten Gewicht steht unscheinbar in der Mitte:** unsere
Testpakete sind **nachgebaut**, nicht bezogen. Sie prüfen, dass der Leser tut,
was wir für richtig halten — nicht, dass er echte Pakete liest. Beides ist nötig,
und nur das erste ist da.

Der Gegenlauf gehört hinter den Build-Tag `fremdpakete` (`AGENTS.md` §3), läuft
nicht in der Pipeline und ist noch nicht geschrieben. Bis dahin ist jede
„plausibel"-Angabe oben genau so viel wert, wie ihr Testfall sie macht.

## `include file:` — zwei Formen, und der Schrägstrich entscheidet

**Mit führendem Schrägstrich** nennt der Verweis ein fremdes Paket:
`include file: /magic-modules.mse-include/corners/card_fields`.

**Ohne ihn** meint er eine Datei im **eigenen** Paket:
`magic.mse-game` schreibt `include file: script` und meint
`/magic.mse-game/script`. Tiefer greifen ist erlaubt
(`watermarks/watermarks` in `magic-m15-extra.mse-style`), hinausgreifen nicht.

**Quelle:** ausgezählt über die 327 Pakete von `M15-Magic-Pack` — von 161
verschiedenen Angaben tragen **134** einen Schrägstrich und **27** nicht. Die
eigenen stehen in Stilen (`magic-m15-invocation.mse-style` → `invocation`), in
Einschüben (`magic-blends.mse-include` → `blends_scripts`), in einer
Exportvorlage und im Spielpaket (`magic.mse-game/game`, Zeile 18).
**Erschlossen am:** 2026-08-15
**Sicherheit:** belegt
**Testfall:** TestEinVerweisOhneSchraegstrichMeintDasEigenePaket, TestEinVerweisDarfDerWertEinerZeileSein

**Und ein Verweis darf der *Wert* einer Zeile sein.** `magic.mse-game` schreibt
seinen Einstieg einzeilig:

```
init script: include file: script
```

Ohne diese Form ist das Initskript des Spielpakets 74 Zeichen lang und besteht
aus einer Leerzeile und einem Kommentar — nachgemessen.

## „Das eigene Paket" ist das der Datei, nicht das des Laufs

Ein Verweis ohne Schrägstrich meint das Paket der **Datei, in der die Zeile
steht** — nicht das, in dem die Auflösung begonnen hat. Wer eine eingebundene
Datei erreicht, wechselt damit die Heimat für alles, was darin steht.

Die Kette, an der es sichtbar wird, geht über zwei Pakete:

```
magic.mse-game/script           Zeile 34:  include file: /magic-blends.mse-include/new-blends
magic-blends.mse-include/new-blends  Zeile 10:  include file: masks
```

Gemeint ist `magic-blends.mse-include/masks` — eine gewöhnliche Textdatei mit 44
Zeilen, die `id_mask` und die Maskennamen setzt. In `magic.mse-game` gibt es
keine Datei dieses Namens.

**Hier stand die Vermutung, `masks` sei ein Verzeichnis**
(#241). Sie kam
aus der Fehlermeldung des eigenen Einbinders — der suchte im falschen Paket, fand
nichts, und die Deutung „dann ist es wohl ein Verzeichnis" war naheliegend und
falsch. Ein Verzeichnis hinter `include file:` kommt in `M15-Magic-Pack` **nicht
vor**; nachgezählt sind alle 161 Angaben, und jede nennt eine Datei.

**Quelle:** `magic.mse-game/script` Zeile 34 und
`magic-blends.mse-include/new-blends` Zeile 10; die Zieldatei ist
`magic-blends.mse-include/masks`. Ausgezählt über alle Pakete stehen **49**
Verweise ohne Schrägstrich in `magic-modules.mse-include` und **15** in
`magic.mse-game` — jeder von ihnen wird aus einem anderen Paket heraus erreicht.
**Erschlossen am:** 2026-08-16
**Sicherheit:** belegt
**Testfall:** TestEinRelativerVerweisMeintDasPaketSeinerDatei, TestNachEinemTiefenIncludeGiltWiederDasEigenePaket

## Eine Zeichenkette in MSE-Script darf über Zeilen gehen

Hier stand bis zum Zweitblick über Phase 6 das Gegenteil, und die Begründung
war, dass **kein Beleg** für eine mehrzeilige vorlag. Ein Schluss aus dem Fehlen
eines Belegs ist kein Beleg.

**Quelle:** von 429 Skriptdateien in `M15-Magic-Pack` führen **24** eine
Zeichenkette über die Zeilengrenze. `magic.mse-game/statistics_script` beginnt
bei `trace("` und schließt Dutzende Zeilen später; die `keywords_*`-Dateien
halten über 2.000 Zeilen offen.
**Erschlossen am:** 2026-08-15
**Sicherheit:** belegt
**Testfall:** TestEineZeichenketteDarfUeberZeilenGehen, TestNachEinerMehrzeiligenZeichenketteStimmtDieZeile

**Ein Rückstrich flieht das nächste Zeichen, auch einen Umbruch.** Die
Zeilenzählung läuft dabei mit — sonst zeigt jeder Fehler danach auf eine falsche
Zeile.

## Ein Schlüssel einer Abbildung darf eine Zeichenkette sein

`magic.mse-game/language` schreibt seine Sprachtabelle als
`languages := [ "Deutsch": [ code: "de", … ] ]`.

Hier stand einmal, die Schlüssel seien „Namen, keine Ausdrücke […] Ein Ausdruck
als Schlüssel kommt nirgends vor". Ein **Ausdruck** ist es weiterhin nicht:
Name oder Zeichenkette, und beides steht fest in der Datei.

**Quelle:** `magic.mse-game/language`, Zeile 33 ff.
**Erschlossen am:** 2026-08-15
**Sicherheit:** belegt
**Testfall:** TestEinAbbildungsschluesselDarfEineZeichenketteSein

## Hinter `name:` darf der Wert auf der nächsten Zeile beginnen

`magic.mse-game` schreibt seine Wortlisten so:

```
word_list_artifact:
  "Altmetall,"
+ "Attraction,"
+ "Ausrüstung,"
```

Der Doppelpunkt steht am Zeilenende, der Wert folgt darunter — und geht mit
führendem `+` weiter. Dieselbe Regel gilt seit jeher hinter `:=` und hinter
einem Operator; hinter `name:` fehlte sie.

**Der Wert darf umbrechen, nicht fehlen**: am Quellende bleibt es bei „hier
fehlt ein Wert".

**Quelle:** `magic.mse-game`, Zeile 196 des eingebundenen Initskripts. **Diese
eine Lücke hielt alle 190 Stile des Packs auf** — jeder lädt dasselbe
Spielskript.
**Erschlossen am:** 2026-08-16
**Sicherheit:** belegt
**Testfall:** TestNachEinemDoppelpunktDarfDerWertUmbrechen, TestEinDoppelpunktOhneWertBleibtEinFehler

## `mod` ist ein Operator und bindet wie `*`

`input mod 10`, `(to_int(offset) + 0) mod 4`, `if (input mod 2) == 0 then …`.

Die Klammern in den echten Vorlagen sagen die Vorrangstufe: um eine Addition
steht eine, um `mod` gegen `==` auch — beides passt zur Stufe von `*` und `/`
und zu keiner anderen. Gerechnet wird auf **ganzen** Zahlen; MSE-Script kennt
nur einen Zahltyp, und im Pack steht `mod` ausschließlich hinter ganzen Werten.

**Was nicht gebaut ist:** `div`. Die 38 Vorkommen im selben Pack sind sämtlich
HTML — `document.createElement("div")` in Exportvorlagen. Was nicht belegt ist,
wird nicht gebaut (§11).

**Quelle:** 272 Vorkommen in `M15-Magic-Pack@3f7f8ce/data`, maschinell
ausgezählt; die Stelle, die den Zerteiler aufhielt, ist
`magic.mse-game`, Zeile 8448 des eingebundenen Initskripts.
**Erschlossen am:** 2026-08-16
**Sicherheit:** belegt
**Testfall:** TestModIstEinOperator

## Ein Schlüssel einer Abbildung darf auch eine Zahl sein

```
russian_number_map :=
[
0	: "ноль"
1	: "один"
```

Die Regel ist nicht „Zeichenketten sind erlaubt", sondern: **was fest in der
Datei steht, ist ein Schlüssel** — es lässt sich lesen, ohne etwas
auszurechnen. Ein **Ausdruck** ist weiterhin keiner.

**Quelle:** `magic.mse-game`, Zeile 8590 des eingebundenen Initskripts.
**Erschlossen am:** 2026-08-16
**Sicherheit:** belegt
**Testfall:** TestAuchEineZahlIstEinSchluessel

## Ein `if` ist ein Ausdruck und darf rechts von einem Operator stehen

```
"<sym>" + phy_matches[0] + "</sym> can be paid with … or 2 life. " +
if multiple then	"For each <sym>…"
else				"If life was paid, …"
```

**Es reicht bis ans Ende.** `"a" + if b then "c" else "d"` verkettet `"a"` mit
dem ganzen `if`; anders gälte der `else`-Zweig nur für sich, und die Vorlage
schriebe die Hälfte.

**Quelle:** `magic.mse-game`, Zeile 9726 des eingebundenen Initskripts.
**Erschlossen am:** 2026-08-16
**Sicherheit:** belegt
**Testfall:** TestEinIfDarfRechtsVonEinemOperatorStehen

## Ein Argumentname darf ein Schlüsselwort sein

`number_of_items(in: colors)` — `in` ist sonst der Kopf einer Schleife
(`for each x in …`).

**Zweideutig ist es nicht:** als Name steht ein Schlüsselwort nur dort, wo
unmittelbar ein `:` folgt, und diese Form gibt es in keiner Schleife.

**Quelle:** 25 Vorkommen von `(in: ` in `M15-Magic-Pack@3f7f8ce/data`;
`magic.mse-game`, Zeile 12650 des eingebundenen Initskripts.
**Erschlossen am:** 2026-08-16
**Sicherheit:** belegt
**Testfall:** TestAuchEinSchluesselwortIstEinArgumentname

## Hinter einer öffnenden Klammer darf der Ausdruck umbrechen

```
else	colors[
font_color_positions[
if		chosen(choice:"horizontal")	then	"horizontal"
```

Dieselbe Regel wie hinter `:=`, hinter einem Operator und hinter `name:` — ein
Umbruch mitten in einem Ausdruck ist nichts.

**Quelle:** `magic.mse-game`, Zeile 12908 des eingebundenen Initskripts.
**Erschlossen am:** 2026-08-16
**Sicherheit:** belegt
**Testfall:** TestHinterEinerKlammerDarfDerAusdruckUmbrechen

## `in`, `from`, `to` und `of` sind auch Namen

```
ar_position := {
pos := -1
for x from 0 to length(in)-1 do (
…
includes := {ar_position(of:input in:array) != -1}
```

Die vier tragen ihre Form nur in deren **Mitte** — `for x from a to b`,
`case x of` —, und der Zerteiler sucht sie dort ausdrücklich. Als Wert stehen
sie für sich.

**Es folgt notwendig** aus der Regel, dass ein Schlüsselwort ein Argumentname
sein darf: wer es nicht auch als Wert liest, hat eine Sprache, in der sich ein
Argument übergeben, aber nicht benutzen lässt.

**Nicht dazu gehören `if`, `for`, `not`, `nil`, `then`, `else`, `each`, `do`**
und die Verknüpfungen: sie **beginnen** eine Form, und als Namen machten sie
`if x then y` zweideutig.

**Quelle:** `magic.mse-game`, Zeilen 13295 und 13304; `to:` steht 111-mal,
`of:` 48-mal und `in:` 19-mal als Argumentname im Spielpaket und seinen Modulen.
**Erschlossen am:** 2026-08-16
**Sicherheit:** belegt
**Testfall:** TestInDarfEinNameSein

## `case … of` ist eine Fallunterscheidung

```
token_shape := {
case input of
"token": true,
"emblem": true,
else: false;
}
```

Paare aus Wert, Doppelpunkt und Ausdruck, durch Komma getrennt; `else` ist die
Vorgabe. Verglichen wird auf **Gleichheit**, der **erste Treffer** gewinnt, und
nur der genommene Zweig wird gerechnet.

**Ohne Treffer und ohne `else` ist das Ergebnis `nil`** — MSE kennt für diesen
Fall keinen Abbruch, und einen zu erfinden hieße, eine Regel zu bauen, die keine
Vorlage kennt.

**Quelle:** acht Vorkommen in vier Dateien von `M15-Magic-Pack@3f7f8ce/data`;
`magic.mse-game`, Zeile 13429 des eingebundenen Initskripts, und
`card_number_style` mit Ausdrücken statt Wahrheitswerten.
**Erschlossen am:** 2026-08-16
**Sicherheit:** belegt
**Testfall:** TestCaseIstEineForm, TestJederFehlerwegDerFallunterscheidung

## `xor` ist ein Operator und bindet wie `or`

```
if (set.CARDNAME_replacement == "card's type" xor reverse)
then (…)
```

Die Klammer um den Vergleich sagt die Vorrangstufe: `xor` bindet **loser** als
`==`.

**Und es rechnet beide Seiten** — anders als `and` und `or`. Ein „entweder
oder" lässt sich nicht abkürzen; wer es wie `or` baute, wertete die rechte Seite
nie aus, wenn die linke zutrifft.

**Quelle:** `magic.mse-game`, Zeile 16236 des eingebundenen Initskripts.
**Erschlossen am:** 2026-08-16
**Sicherheit:** belegt
**Testfall:** TestXorIstEinOperator, TestXorBrauchtWahrheitswerte

## `+` verkettet zwei Funktionen

`f + g` ergibt eine Funktion, die erst `f` anwendet und dann `g` auf deren
Ergebnis. Der Zwischenwert reist als **Stellungsargument und als `input`** —
eine eingebaute Funktion liest ihn an Stelle eins, ein Block `{ … }` als
`input`, und die Verkettung weiß nicht, welche Form die rechte Seite hat.

**Quelle:** `magic-mtgsalvation-forum.mse-export-template` Zeile 37 ff. reiht
zehn `replace@(…)` mit `+` aneinander; `magic-m15-godzilla-planeswalker`
verbindet `replace@(…)` mit `filter_text@(…)`. Gerufen wird die verkettete
Funktion als `paragraph_count(card.rule_text)` — mit einem Stellungsargument.
Sieben der neunzehn bezogenen Exportvorlagen hingen daran.
**Erschlossen am:** 2026-08-15
**Sicherheit:** belegt
**Testfall:** TestZweiFunktionenLassenSichVerketten, TestDieVerkettungLaeuftVonLinksNachRechts

## `sort_list(liste, order_by: {…})` ordnet nach einem Schlüssel

Der Schlüsselausdruck ist ein Block; das Glied steht darin als `input` **und**
als `card`.

**Quelle:** vier Exportvorlagen rufen sie. `magic-toexp` schreibt
`sort_list(cards, order_by: {card.card_number})`, `magic-forum-spoiler`
schreibt `{input.card_number}` — dasselbe Glied unter zwei Namen. Dass MSE
genau diese beiden setzt, sagt keine Datei; dass beide Formen laufen müssen,
sagen die Vorlagen.
**Erschlossen am:** 2026-08-15
**Sicherheit:** Vermutung — **unsere Referenzfassung setzt genau einen Namen**:
`card`, wenn die Liste ein Set ist, sonst `input` (`sort_script`,
`src/script/functions/basic.cpp`, Zeile 507, Fassung f9d9356 vom 2021-07-06).
Das Orakel bestätigt es am 2026-10-10: über `set.cards` geht `{input.name}`,
und `{card.name}` endet in „Variable not set: card"; über `set` geht
`{card.name}`. **Die Packs tragen `mse version: 2.5.8`** und benutzen beide
Namen über beiden Arten von Listen — `magic-toexp` und
`magic-skeleton-extractor` `{card.…}` über `cards`, `planechase-forum` und
`saga-forum` `{input.…}` über `set`. Unter f9d9356 scheiterten diese vier.
**Dass eine neuere Fassung beide Namen setzt, ist erschlossen**, und dieser
Auswerter folgt den Packs (F05, #2247).
**Testfall:** TestSortListeOrdnetNachDemSchluessel, TestSortListeKenntDasGliedAlsCard

## `sort_list` vergleicht die Schlüssel als Text, und klug

Der Schlüssel eines Glieds ist das **Textergebnis** des Ausdrucks, und zwei
Schlüssel stehen in der Reihenfolge von `smart_compare`: Zahlen im Text zählen
als Zahlen (`2/99` vor `10/99`), groß und klein sind gleich, ein Umlaut steht
bei seinem Grundbuchstaben. Gleiche Schlüssel bleiben in der Reihenfolge der
Liste (MSE nimmt `std::sort`, hier ist es stabil — eine Ausgabe, die sich ohne
Änderung unterscheidet, wäre schlimmer).

**Quelle:** MSE2, `src/script/functions/basic.cpp`, Zeilen 478–520
(`sort_script`, `smart_less_first`) und `src/util/string.cpp`, Zeilen 208–318
(`smart_compare`), Fassung f9d9356 vom 2021-07-06. Gemessen am
Leistungsvergleich (#2006): die Kartennummer von `magic.mse-game` ist Text wie
`0002` oder `2/99`, und Byte für Byte verglichen stand `10/99` vor `2/99`.
**Erschlossen am:** 2026-10-03
**Sicherheit:** belegt
**Testfall:** TestSortListeOrdnetKartennummernAlsText, TestSortListeOrdnetOhneAufGrossUndKleinZuSehen, TestKlugVergleicheWieMSE

## `to_text` nimmt die Auszeichnung heraus

`to_text(x)` ist `untag_hide_sep`: erst fällt jeder weiche Trenner
(`<sep-soft>`) samt Inhalt weg, dann jede Marke. Aus
`<word-list-type-en>Scheme</word-list-type-en><sep-soft> — </sep-soft>` wird
`Scheme`. Eine offene Klammer bleibt **nicht** stehen: ab einem `<` ohne `>`
fällt der Rest weg.

**Quelle:** MSE2, `src/script/functions/export.cpp`, Zeilen 361–364 (`to_text`)
und `src/util/tagged_string.cpp`, Zeilen 33–47 und 65–67 (`untag`,
`untag_hide_sep`), Fassung f9d9356 vom 2021-07-06. Das Orakel (`--cli`) bestätigt
es an `set.cards.2.type` von `Archenemy.mse-set`.
**Erschlossen am:** 2026-10-03
**Sicherheit:** belegt
**Testfall:** TestZuTextNimmtDieAuszeichnungHeraus

## Eine Mehrfachwahl, an der nichts gewählt ist, heißt wie ihr `empty choice:` sagt

Ein `card field:` der Art `multiple choice` kann `empty choice: <name>` tragen.
Ist nach Vorgabewert und Skript nichts gewählt, **ist der Wert dieser Name** —
nicht die leere Zeichenkette. `magic.mse-game` erklärt `colorless` an sieben
Kartenfeldern: `card color` bis `card color 4` und `indicator` bis `indicator 3`.
Ohne Angabe bleibt der Wert leer.

Das gilt für jede Mehrfachwahl und nicht nur für die Optionen eines Stils; hier
stand, ein Kartenfeld habe keinen Leernamen. Sichtbar wird es dort, wo ein
Skript den Wert liest: `color_of_card()` unterscheidet `colorless` von „leer",
und die Kartennummer hängt an dieser Unterscheidung.

**Und der Name bleibt eine Vorgabe.** MSE2 schreibt ihn über
`mutateDontChangeDefault`: ein Wert, der Vorgabe war, ist es danach noch, und
das Vorgabeskript rechnet ihn beim nächsten Durchgang neu. Wer den Namen vor
dem Vorgabeskript einsetzt und Vorgaben an ihrer Leere erkennt, rechnet sie
nie mehr aus — so stand am 2026-10-04 jede Karte des Kanons farblos da.

**Quelle:** MSE2, `MultipleChoiceValue::normalForm`,
`src/data/field/multiple_choice.cpp`, Zeilen 110–113, und
`MultipleChoiceField::empty_choice`, `.hpp` Zeile 27, Fassung f9d9356 vom
2021-07-06; `magic.mse-game/card_fields`, Zeilen 112, 157, 202, 247, 736, 782
und 828. Das Orakel (`--cli`) gibt `card_color` von `Slaughter Pact` aus
`Black.mse-set` als `"colorless"`, wo diese Fassung `""` rechnete — die eine
Karte von 3.788, die in der Ausfuhr von `magic-text` an anderer Stelle stand.
**Erschlossen am:** 2026-10-03
**Sicherheit:** belegt
**Testfall:** TestEinKartenfeldTraegtSeinenLeernamen, TestEineLeereMehrfachwahlEinesKartenfeldsSchreibtIhrenLeernamen, TestEineLeereMehrfachwahlNimmtErstIhreVorgabe

## `position(of: karte, in: set, order_by:, filter:)` zählt die Karte im geordneten Set

Mit `order_by:` oder `filter:` und einer Karte `of:` im Set `in:` ist das
Ergebnis die **Stelle der Karte** (von null an), wenn man die Karten des Sets
nach dem Schlüssel `order_by` ordnet und die weglässt, die `filter` nicht
durchlässt. Eine ausgenommene Karte hat die Stelle −1. Beide Ausdrücke laufen
je Karte des Sets mit `card` als dieser Karte; der Schlüssel ist Text und wird
wie bei `sort_list` klug verglichen. Ein Filter `nil` heißt: keiner. Ist `of`
keine Karte oder `in` nicht das Set, ist es ein Fehler — **aber ein Text auf
einer der beiden Seiten bleibt die Textsuche**, auch mit `order_by:`.

So rechnet `magic.mse-game` die Kartennummer: `card_number` ist `position(of:
card, in: set, order_by: {partition_index() + rarity_sort() + sort_index() +
sort_name(card.name) + sort_name(export_name())}, filter: set_filter()) + 1`.
`set_filter()` gibt je nach Karte einen von neun Blöcken — neun Ordnungen, jede
wird **einmal** gebaut und gemerkt (`Set::order_cache`, gekeyt nach beiden
Ausdrücken); MSE leert den Speicher am Anfang jeder Runde von Skripten
(`clearOrderCache`).

**Hier gilt:** das Set erkennt man an der `Kennung` „set", die Karte an der
Gestalt einer Abbildung; was in der Liste keine Kennung trägt, ist keine Karte.
Im **Zeichenpfad** führt `set` nur seine Angaben und die Karte keine Kennung —
`position` gibt dort −1 und `card_number` bleibt, was es war („0000"). Erst die
Ausfuhr rechnet die Nummer (nächster Abschnitt). **Gleiche Schlüssel** stehen in
der Reihenfolge der Liste; MSEs `std::sort` lässt sie, wie der Algorithmus will.

**Quelle:** MSE2, `src/script/functions/basic.cpp`, Zeilen 449–474
(`position_in_vector`) und 524–531 (`position_of`), `src/data/set.cpp`, Zeilen
277–299 und 315–318 (`positionOfCard`, `clearOrderCache`) und
`src/util/order_cache.hpp`, Zeilen 44–79 (`OrderCache`), Fassung f9d9356 vom
2021-07-06. Das Orakel (`--cli`) gibt `set.cards.2.card_number` als `"0002"`.
**Erschlossen am:** 2026-10-03
**Sicherheit:** belegt
**Testfall:** TestPositionOrdnetDieKartenDesSetsNachDemSchluessel, TestPositionZaehltNurDieKartenDesFilters, TestPositionBehandeltEinenFilterNilAlsKeinenFilter, TestPositionSuchtMitOrdnungInEinemTextWeiterImText, TestPositionInEinemSetOhneKartenGibtMinusEins, TestPositionBautDieOrdnungEinmalWennDerSpeicherDaIst

## Die Ausfuhr rechnet die Skriptfelder der Karten

Eine Karte in `cards` und `set.cards` trägt die Werte, die ihre Felder **nach
den Skripten** haben: erst der Vorgabewert (`default:`), dann das Wertskript
(`script:`), das den gespeicherten Wert überschreibt. `type`, `card number` und
`shape` von `magic.mse-game` tragen `save value: false` — in der Datei steht von
ihnen nichts, und eine Vorlage, die nach `card.card_number` ordnet (`magic-text`,
`magic-mws`), ordnete nach lauter leeren Schlüsseln.

MSE2 rechnet beim Laden: `SetScriptManager::updateAll` geht über jede Karte und
jedes ihrer Felder, danach `updateAllDependend` für alles, was von der
Kartenliste abhängt — `card_number` zählt die Karte **im Set**, nach Werten, die
erst die erste Runde berechnet hat. Die Vorlage selbst läuft im Kontext des Sets
und sieht damit die Funktionen des `init script`s.

**Hier gilt:** zwei Runden über denselben Weg wie die Auszählung. Die erste
rechnet ohne Kartenliste, die zweite mit den Ergebnissen der ersten als `set`
(Einträge: die Karten, Glieder: die Angaben und `cards`) und leerem Speicher der
Ordnungen. MSE rechnet in der zweiten Runde nur, was von der Liste abhängt; ein
Abhängigkeitsgraph ist hier nicht gebaut. **Der Stil der Karte fehlt**: MSE
rechnet jede Karte im Kontext ihres Stilbogens, die Exportvorlage nennt keinen —
`stylesheet` ist ein leerer Stil, und `shape` (`stylesheet; card_shape()`) gibt
den Wert des Spiels ohne Stil. Ein Wertskript, das scheitert, lässt den
gespeicherten Wert stehen und meldet **einen** Befund je Meldung.

**Quelle:** MSE2, `src/gui/html_export_window.cpp`, Zeilen 55–90 (`export_set`),
`src/script/script_manager.cpp`, Zeilen 265–300 (`updateAll`) und 30–75
(`getContext`), `src/data/field/text.cpp`, Zeilen 178–179 (`default_script.
invokeOnDefault … script.invokeOn`), Fassung f9d9356 vom 2021-07-06. Gemessen an
`Archenemy.mse-set` mit `magic-text` und `magic-mws`: beide Ausgaben sind
bytegleich mit MSE2.
**Erschlossen am:** 2026-10-03
**Sicherheit:** belegt
**Testfall:** TestDieAusfuhrGibtDenKartenDieWerteIhrerSkriptfelder, TestEinWertskriptSchlaegtDenGespeichertenWert, TestDieKartennummerZaehltDieKarteImSetUndOrdnetDieVorlage, TestDieKartennummerOrdnetNachGerechnetenWerten, TestSetCardsTraegtDieGerechnetenKarten, TestEinGescheitertesWertskriptMeldetEinenBefund

## Eine Option mit `type: boolean` ist ein Wahrheitswert

`if options.text_costs then …` liest keine Zeichenkette: die Option nennt ihren
Typ selbst. Zu deuten ist nur die Vorgabe, und `yes` ist wahr, alles andere
falsch.

**Quelle:** die neunzehn Exportvorlagen führen 36 Optionen mit `type: boolean`;
29 davon tragen ein `initial:`, ausnahmslos `yes` (11) oder `no` (18). Die
sieben ohne Angabe gelten als falsch — die andere Lesart machte aus einer
fehlenden Angabe ein eingeschaltetes Merkmal.
**Erschlossen am:** 2026-08-15
**Sicherheit:** belegt
**Testfall:** TestEineBoolescheOptionKommtAlsWahrheitswertAn, TestEineBoolescheOptionMitNo

---

## Auch eine Exportvorlage schreibt `include file:`

Das Einbinden von Modulen ist keine Eigenheit der Spiel- und Stilpakete. Eine
Exportvorlage darf ihr Skript mit `include file: <name>` beginnen und den Rest
aus einer zweiten Datei desselben Pakets holen — dieselbe Form, dieselbe
Auflösung ohne Schrägstrich.

Wer die Zeile stehen lässt, bekommt sie vom Zerteiler zurück, und die Meldung
führt in die Irre: `include` liest sich als Name, und an Spalte 9 heißt es
„hier hört der Ausdruck auf, aber der Text geht weiter". Der Fehler steht dann
in der ersten Zeile eines Skripts, das an dieser Stelle völlig in Ordnung ist.

**Quelle:** `magic-lackeybot.mse-export-template/export-template` Zeile 46
(`include file: script`) und `magic-namecheck.mse-export-template/export-template`
Zeile 36 (`include file: namelist`) aus `M15-Magic-Pack` — 2 von 19 Vorlagen.
Beide fingen ihr Skript genau so an.
**Erschlossen am:** 2026-08-16
**Sicherheit:** belegt — nach dem Einbinden wandert der Fehler in beiden
Vorlagen auf eine Bibliotheksfunktion (`match`), also über die eingebundene
Stelle hinaus.
**Testfall:** TestEinIncludeInEinerExportvorlageWirdAufgeloest

---

## Zehn Namen der Standardbibliothek, ausgezählt

Alle aus `M15-Magic-Pack`, jede mit ihrer Fundstelle bei ihrer Umsetzung in
`internal/msescript/`:

| Name | Vorkommen | Was sie tut |
| --- | ---: | --- |
| `match` | 616 in `magic.mse-game` | Muster trifft Text → Wahrheitswert |
| `split_text` | 176 | zerlegt an einem Muster → Liste |
| `chosen` | 300 | eine Wahl eines Mehrfachfeldes → Wahrheitswert |
| `break_text` | | die **Treffer** eines Musters → Liste |
| `position` | | Stelle in einer Liste, sonst `-1` |
| `to_int`, `to_number` | | Zahl, abgeschnitten |
| `remove_tag` | | eine **benannte** Auszeichnung heraus |
| `regex_escape` | | Text → Muster, das genau ihn trifft |
| `replace_text` | | `replace` unter einem zweiten Namen |

**Quelle:** die 329 Pakete von `M15-Magic-Pack`; die Zahlen sind ausgezählt,
die Fundstelle jeder einzelnen steht bei ihrer Umsetzung in
`internal/msescript/`.
**Erschlossen am:** 2026-08-16
**Sicherheit:** belegt
**Testfall:** TestMatchGibtEinenWahrheitswert, TestDieNachgezogenenBibliotheksfunktionen

---

## Die Regex-Sprache: RE2 trägt 99 Prozent

Ausgezählt über alle `match:`-Angaben von `M15-Magic-Pack`: **1645
verschiedene Muster**. Davon brauchen **17** etwas, das Gos `regexp` nicht
kann — dreizehn einen Vorausblick `(?=`/`(?!`, drei einen Rückblick `(?<`,
eines einen Rückverweis. 557 benutzen `(?i)`, das RE2 trägt.

Die siebzehn scheitern **laut**, mit dem Muster in der Meldung. Der stille Weg
— „nicht übersetzbar, also kein Treffer" — erzeugte eine Ausfuhr, die aussieht
wie eine richtige.

**Quelle:** `grep -rho 'match:[[:space:]]*"[^"]*"'` über alle 329 Pakete.
**Erschlossen am:** 2026-08-16
**Sicherheit:** belegt für die Zahlen. **Nicht** gesagt ist, dass ein von
`regexp` übersetztes Muster dasselbe trifft wie in MSE — beide sind Perl-nah,
und für die ausgezählten Formen decken sie sich.
**Testfall:** TestEinUnbrauchbaresMusterWirdBenannt

---

## Eine Zahl ist eine Bedingung, eine Zeichenkette nicht

`if to_number(substring(…)) then …` steht in `magic-field-test` (Zeile 291) und
`magic-lackeybot` (Zeile 577): eine Anzahl, auf „gibt es überhaupt" geprüft.
Null ist falsch, alles andere wahr.

**Nur als Bedingung eines `if`.** Für `1 and true` oder `not 1` ist keine
Fundstelle da, und dort bleibt es ein Fehler. Ein Fall macht keine Klasse.

**Für eine Zeichenkette gilt es nicht:** die Lesarten gehen auseinander („nicht
leer" oder „nicht `no`"?), und wer rät, macht aus `if styling.overlay then …`
etwas anderes, als die Vorlage meint.

**Quelle:** die beiden genannten Vorlagen.
**Erschlossen am:** 2026-08-16
**Sicherheit:** belegt für die Zahl; für die Zeichenkette ist es eine
Vermutung, und deshalb bleibt sie ein Fehler.
**Testfall:** TestDieNachgezogenenBibliotheksfunktionen, TestEineZeichenketteBleibtKeineBedingung

---

## Ein Block liest sein Stellungsargument als `input`

`({ input + "!" })("hallo")` ergibt `hallo!`. Vier der neunzehn Exportvorlagen
lesen `input` in einem Block, `magic.mse-game` ebenfalls.

**Ein benanntes `input:` schlägt die Stellung**, und ohne Stellungsargument
bleibt `input` aus der äußeren Umgebung sichtbar — wo es auch dort fehlt, ist
es ein unbekannter Name. Ein leerer Text machte aus einem Tippfehler ein
Ergebnis.

Hier stand bis dahin im Quelltext, ein Stellungsargument sei „nicht erreichbar
— in den gesichteten Vorlagen wird auch keines gelesen". Der zweite Halbsatz
war die Begründung, und er stimmte für die damals angesehenen Vorlagen, nicht
für die Menge.

**Quelle:** `magic-lackeyccg` Zeile 63, `magic-skeleton-extractor` Zeile 12,
`magic-field-test`, `magic-lackeybot`.
**Erschlossen am:** 2026-08-16
**Sicherheit:** belegt
**Testfall:** TestEinBlockLiestSeinStellungsargumentAlsInput

---

## Eine Abbildung und ein Text ergeben Text

`magic.mse-game` Zeile 8411 verbindet eine Zeichenkette mit einer Liste. Zwei
Abbildungen ergeben weiterhin eine **Vereinigung** (`map + [stylesheet: "…"]`);
eine Abbildung mit einer Zahl oder einem Wahrheitswert bleibt ein Fehler.

Das ist keine neue Entscheidung, sondern die vorhandene zu Ende geführt: die
Textform einer Abbildung ist längst die Folge ihrer Werte. Wo ein Text erwartet
wird, *ist* eine Liste ihr Text.

**Quelle:** `magic.mse-game/script`, Zeile 8411.
**Erschlossen am:** 2026-08-16
**Sicherheit:** belegt
**Testfall:** TestEineAbbildungUndEinTextErgebenText,
TestEineAbbildungUndEineZahlWerdenAneinandergesetzt,
TestPlusFaelltAufDieAneinandersetzungZurueck

---

## MSE-Script benutzt Perl-Muster, nicht RE2

MSE übersetzt jedes `match:` mit `boost::regex` in der **Perl**-Fassung. Damit
sind Vorausblick `(?=`/`(?!`, Rückblick `(?<` und Rückverweise `\1` gültig, und
der Zusatz `(?x)` lässt Leerraum im Muster weg.

Ausgezählt über alle `match:`-Angaben von `M15-Magic-Pack`: von **1645**
verschiedenen Mustern gehen **1630** durch Gos `regexp` (RE2), **15** nicht.
Diese fünfzehn laufen über eine zweite Maschine mit Rückverfolgung
(`internal/msescript/muster.go`).

**Ein Prozent der Muster machte ein Drittel der Befunde.** Das Muster
`(^|\|)[^ ]+(?=\|)` steht in `magic.mse-game` an zwölf aufeinanderfolgenden
Zeilen, die **jeder** der 190 Stile ausführt — gemessen 2280 Befunde.

**Hier stand bis zum 2026-08-16 die Notlösung**, ein solches Muster als
*Vorbehalt* mit Ersatzwert weiterlaufen zu lassen. Sie war unter der alten
MIT-Lizenz die einzige Möglichkeit, weil die Perl-Fassung nicht nachlesbar war.
Mit dem Wechsel auf GPL-2.0 ([0019](entscheidungen/0019-gpl-2-0-statt-mit.md))
ist die Ursache behoben, und die Umgehung ist gelöscht statt behalten.

**Ein kaputtes Muster bleibt ein Abbruch.** `"["` ist in *jeder* Regex-Sprache
falsch — ein Fehler der Vorlage, und dort hilft Weiterrechnen niemandem.

**Quelle:** `src/util/regex.hpp` Zeile 27 von twanvl/MagicSetEditor2;
`magic.mse-game/script` Zeilen 8426 bis 8437; die Auszählung über alle 329
Pakete.
**Erschlossen am:** 2026-08-16
**Sicherheit:** belegt
**Testfall:** TestEinVorausblickWirdUebersetzt, TestJedeMusterfunktionKenntDenVorausblick

---

## Ein Ersatztext in `replace` ist eine sed-Vorlage

`\1` bis `\9` sind die Gruppen, `&` ist der ganze Treffer, `\&` ein
kaufmännisches Und. Alles übrige steht wörtlich — insbesondere ist `$1` **kein**
Gruppenverweis, sondern ein Dollarzeichen mit einer Eins.

MSE gibt in `Regex::Results::format` ausdrücklich `boost::format_sed` mit; ohne
dieses Kennzeichen wäre es die Perl-Fassung mit `$1`.

Ausgezählt stehen in `M15-Magic-Pack` **33** verschiedene Ersatztexte mit
Rückverweis, darunter der Umbau der Erinnerungstexte
(`\2<blob>\4\3<moved>\1</moved></blob>`). Bis zum 2026-08-16 setzte dieses
Programm sie wörtlich ein — aus `replace: "\1"` wurde die Zeichenfolge `\1` im
Kartentext.

**Quelle:** `src/util/regex.hpp` Zeile 63 und
`src/script/functions/regex.cpp` von twanvl/MagicSetEditor2; die Auszählung über
alle 329 Pakete.
**Erschlossen am:** 2026-08-16
**Sicherheit:** belegt
**Testfall:** TestEinErsatztextKenntRueckverweise

---

## `tag_contents` ersetzt den Inhalt einer Auszeichnung

`tag_contents@(tag: "<soft>", contents: {""})`: die Auszeichnung bleibt stehen,
was zwischen ihr steht, wird durch das Ergebnis von `contents` ersetzt.
`contents` ist eine **Funktion** und bekommt den bisherigen Inhalt als `input`;
ein Wert, der keine Funktion ist, meint sich selbst.

**Nicht gierig**: zwei Vorkommen nacheinander sind zwei Ersetzungen und nicht
eine über alles.

**Quelle:** `magic.mse-game/script`, Zeile 10594.
**Erschlossen am:** 2026-08-16
**Sicherheit:** belegt
**Testfall:** TestTagContentsErsetztDenInhalt

## Ein Spiel erklärt `set field:` neben `card field:`

`magic.mse-game/set_fields` führt **89 `set field:`-Blöcke** — Titel,
Beschreibung, Künstler, Copyright, Setcode, Sprache, Symbol. Sie haben
**dieselbe Gestalt** wie ein `card field:`: `type:`, `name:`, `default:`,
`choice:`, `localized caption:`.

Die Skripte der Stile lesen daraus: über `M15-Magic-Pack` gezählt **1.400
Zugriffe auf `set.…` in 167 Dateien** — `set.symbol` 420-mal,
`set.border_color` 345-mal, `set.mainframe_rarity_name` 101-mal. Nach
`stylesheet.card_width` ist es der meistgelesene Name des Packs.

**Warum die Erklärung gebraucht wird und nicht nur das Ausgefüllte:** ein Set
füllt aus, was es ausfüllt. Ein leer gelassenes Feld muss trotzdem als
**leerer Text** in der Umgebung stehen und nicht fehlen — sonst bricht
`set.copyright + " x"` den ganzen Ausdruck ab, wo MSE weiterrechnet. Es ist
dieselbe Klasse wie bei den Kartenfeldern (#278).

**Quelle:** Datenpaket `magic.mse-game/set_fields`
**Erschlossen am:** 2026-08-16
**Sicherheit:** belegt
**Testfall:** TestEinSpielErklaertSeineSetfelder

## Auch ein `default:` eines Setfelds ist ein Skript

Von den 89 Setfeldern haben **zwei** ein `default:`, und beide stehen in
**Anführungszeichen**:

```
default: "Type custom types here; or replace this text with 'detect' to scan"
default: " *(\\.|\\n|;|,|(?=!)) *"
```

Eine Zeichenkette in Anführungszeichen ist ein gültiger MSE-Script-Ausdruck.
Roh durchgereicht stünden die Anführungszeichen im Wert.

**Die Gegenprobe zur Kartenseite:** dort sind alle 142 Vorgaben ungeklammerte
Skripte der Form `stylesheet; …` (#298). Beide Formen sind Skript, keine ist
geklammert — wer auf `{ … }` prüft, lässt **alle 144** fallen.

**Quelle:** Datenpaket `magic.mse-game/set_fields`, Zeilen 434 und 445
**Erschlossen am:** 2026-08-16
**Sicherheit:** belegt
**Testfall:** TestEinSetfeldOhneWertTraegtDieVorgabeDesSpiels

## `set.symbol` ist ein Symbolwert, kein Bildpfad

`magic-m15.mse-style` schreibt

```
no:  { symbol_variation(symbol: set.symbol, variation: "common") }
yes: { symbol_variation(symbol: set.symbol, variation: "invertedcommon") }
```

`symbol_variation` nimmt bei MSE2 einen **SymbolValue** entgegen — den Inhalt
einer `.mse-symbol`-Datei, also eine **Vektorzeichnung** —, dazu einen
`variation:`-Namen, und schlägt ihn in `style->variations` nach. Das Ergebnis
ist ein gerastertes Bild.

**Damit hängt das Seltenheitssymbol an einem Vektorrasterer**, den es hier
nicht gibt. `SymbolToImage` ist eine der fünf Bildquellen in
`generated_image.hpp`; die anderen vier sind gebaut.

**Das Symbol eines Sets kommt aus dem Set**, nicht aus der Vorlage — für die
Referenzkarten dieses Projekts liegt es deshalb nicht vor.

**Hier stand bis zum 2026-08-16: „im Pack liegt keine einzige
`.mse-symbol`-Datei".** Das war falsch, und der Fehler war die Reichweite der
Prüfung: `find` lief über das lokal installierte `M15-Magic-Pack`, und daraus
wurde eine Aussage über „das Pack". `MagicSetEditorPacks/Full-Magic-Pack` führt
**21** davon, in sechs Stilpaketen — `magic-comic`, `magic-vs`,
`magic-new-extra`, `magic-etherium`, `magic-nouveau`, `magic-nouveau_duels`.
Angestrichen hat das der Nutzende, nicht die Prüfung.

**Damit gibt es Prüfmaterial für einen Rasterer** — siehe die nächste
Feststellung.

**Quelle:** MSE2 `src/script/functions/image.cpp:142`, Fassung f9d9356 vom
2021-07-06; dazu Datenpaket `magic-m15.mse-style`, Zeilen 179–180, und der
Dateibaum von `MagicSetEditorPacks/Full-Magic-Pack`, Zweig `main`, abgefragt am
2026-08-16
**Erschlossen am:** 2026-08-16
**Sicherheit:** belegt
**Testfall:** TestDerSymbolrastererLiestJedeEchteZeichnung

Er läuft über die Naht `msepaket` → `service` → `render` und über alle **21**
Vektorzeichnungen des Packs.

**Was er nicht abnimmt:** ob die gerasterte **Form** stimmt. Dafür
bräuchte es ein Goldbild von MSE2, und MSE2 rastert ein Symbol nur im Rahmen
einer Karte, deren Set es führt — ein solches Set liegt nicht vor. Geprüft ist
damit, dass jede Zeichnung gelesen wird und eine nicht leere Maske ergibt
(#323).

## `.mse-symbol` ist eine Vektorzeichnung im MSE-Textformat

Dieselbe Syntax wie jedes andere MSE-Paket — Einrückung, `schlüssel: wert` —,
und darin ein Baum aus `part:`-Blöcken:

```
mse version: 0.3.5
part:
	type: symmetry
	name: Reflection (2)
	kind: reflection
	copies: 2
	center: (0.5000000000,0.5000000000)
	handle: (0.2000000000,0.0000000000)
	part:
		type: shape
		name: Rectangle
		combine: overlap
		point:
			position: (0.1880949567,0.3189438621)
			lock: free
			line after: curve
			handle after: (-0.0564330311,-0.0268493516)
```

**Was daraus abzulesen ist:**

- `type: shape` ist ein geschlossener Zug aus `point:`-Blöcken; `type: symmetry`
  vervielfältigt seine Kinder (`kind:`, `copies:`, `center:`, `handle:`).
- Ein Punkt trägt `position:` und, wenn `line after: curve`, die Bézier-Griffe
  `handle before:`/`handle after:` — **relativ zur Position**, nicht absolut.
- Alle Koordinaten sind auf **[0,1]** normiert. Die Zielgröße kommt vom
  Aufrufer, nicht aus der Datei.
- `combine:` sagt, wie die Form mit dem bisherigen Bild verrechnet wird
  (`overlap`, …).

**Damit ist das Format lesbar, ohne eine Zeile neuen Zerteilers**: `msepaket`
liest diese Syntax bereits. Was fehlt, ist das **Rastern** — Füllregel,
Bézier-Auswertung, Symmetrien, Varianten.

**Quelle:** Datenpaket `MagicSetEditorPacks/Full-Magic-Pack`,
`data/magic-nouveau.mse-style/flying.mse-symbol`, Zweig `main`, abgefragt am
2026-08-16
**Erschlossen am:** 2026-08-16, die Bedeutung am 2026-08-17
**Sicherheit:** belegt
**Testfall:** TestEineSymbolzeichnungWirdGelesen,
TestEinUnlesbarerVektorIstDerNullpunkt

## Und die Bedeutung steht in `src/data/symbol.cpp`

Am 2026-08-17 nachgelesen, Fassung f9d9356 vom 2021-07-06:

| Angabe | Werte | Bedeutung |
| --- | --- | --- |
| `type:` | `shape`, `symmetry`, `group` — **leer heißt `shape`** | Zeile 139 |
| `combine:` | `overlap`, `merge`, `subtract`, `intersection`, `difference`, `border` | Zeilen 153–160 |
| `kind:` | `rotation`, `reflection` | Zeilen 225–228 |
| `line after:` | `line`, `curve` | Zeilen 22–25 |
| `lock:` | `free`, `direction`, `size` | Zeilen 17–21 |

**`line after:` gehört der Strecke, nicht dem Punkt.** MSE leitet die Art der
Strecke *davor* ab: `p2->segment_before = p1->segment_after`
(`enforceConstraints`, Zeilen 205–212). Ein Zug ist **geschlossen** — die
letzte Strecke führt vom letzten Punkt zum ersten und trägt dessen Angabe.

**`lock:` ist eine Angabe für den Zeichner, nicht für den Rasterer.** Sie bindet
die beiden Bézier-Griffe eines Punktes aneinander, während jemand ihn zieht
(`onUpdateLock`).

**Die Varianten stehen nicht in der `.mse-symbol`.** `symbol_variation` schlägt
sie in `style->variations` nach — also im **Stilpaket**, nicht in der
Zeichnung. Die Datei trägt nur `parts:` (`IMPLEMENT_REFLECTION(Symbol)`, Zeile
330–332).

**Was damit gelesen werden kann und was nicht:** `msepaket` liest seit dem
2026-08-17 die Gestalt. Das **Rastern** — Bézier auswerten, Symmetrien
ausführen, `combine:` verrechnen, füllen — ist damit *nicht* gebaut, und es
lässt sich hier auch nicht abnehmen: die 21 Prüfdateien liegen in
`Full-Magic-Pack`, und das ist auf diesem Rechner nicht installiert
(#323).

**Quelle:** MSE2 `src/data/symbol.cpp`, Zeilen 17–25, 126–166, 205–212,
225–285 und 330–332, Fassung f9d9356 vom 2021-07-06
**Erschlossen am:** 2026-08-17
**Sicherheit:** belegt
**Testfall:** TestEineSymbolzeichnungWirdGelesen,
TestEinUnbekannterSymbolteilWirdBenannt, TestEinTeilOhneTypIstEineForm,
TestEineGruppeTraegtIhreKinder

## `italic name:` — kursiv wird benannt, nicht errechnet

MSE2 schrägt keine Glyphen. Bei gesetztem `FONT_ITALIC` nimmt es die Schrift,
deren **Namen der Stil angibt**:

```cpp
} else if (flags & FONT_ITALIC && !italic_name().empty()) {
  font = wxFont(size_i, wxFONTFAMILY_DEFAULT, wxFONTSTYLE_NORMAL, weight_i, underline(), italic_name());
}
```

`MPlantin` kursiv ist die Datei `MPlantin-Italic` — eine eigene Schrift mit
eigenen Formen; eine geneigte Fassung derselben Glyphen sähe anders aus.

**475 `italic name:` stehen in `M15-Magic-Pack`.** Fehlt die Angabe, bleibt es
bei der aufrechten Schrift — die Bedingung verlangt ausdrücklich
`!italic_name().empty()`.

**Fett geht einen anderen Weg**: es ist ein **Gewicht**, kein Name
(#326).

**Quelle:** MSE2 `src/data/font.cpp`, Zeilen 100–101, Fassung f9d9356 vom
2021-07-06; dazu das Datenpaket `M15-Magic-Pack`
**Erschlossen am:** 2026-08-16
**Sicherheit:** belegt
**Testfall:** TestEinFeldstilNenntSeineKursivschrift

## `<i…>` wird über ein Präfix erkannt, nicht über einen Namen

MSE prüft die Auszeichnungen der Reihe nach, und jede Prüfung ist ein
**Präfixvergleich**:

```cpp
if      (is_tag(text, tag_start, _( "<b")))   bold   += 1;
else if (is_tag(text, tag_start, _("</b")))   bold   -= 1;
else if (is_tag(text, tag_start, _( "<i")))   italic += 1;
else if (is_tag(text, tag_start, _("</i")))   italic -= 1;
else if (is_tag(text, tag_start, _( "<sym"))) symbol += 1;
```

Daraus folgt dreierlei:

- `<i>`, `<i-flavor>` und `<i-auto>` wirken **gleich**, und ein künftiges
  `<i-was-auch-immer>` ebenso. Eine Aufzählung der bekannten Namen ginge beim
  nächsten schief.
- **Gezählt wird, nicht geschaltet.** `<i>a<i>b</i>c</i>` ist durchgehend
  kursiv; wer beim ersten `</i>` aufhörte, ließe `c` aufrecht stehen.
- Der Präfixvergleich greift weiter, als sein Name sagt: `<bullet>` erfüllt
  `is_tag(…, "<b")` und setzt damit **fett**, bevor der eigene
  `</bullet`-Zweig weiter unten erreicht wird. Ob das Absicht ist, steht dort
  nicht — festgehalten, nicht nachgebaut.

**Quelle:** MSE2 `src/render/text/element.cpp`, Zeilen 78–83 und 169, Fassung
f9d9356 vom 2021-07-06
**Erschlossen am:** 2026-08-16
**Sicherheit:** belegt
**Testfall:** TestDerFlavortextWirdKursiv

## Mehrere Leerzeichen bleiben mehrere

MSE bricht **an** einem Leerzeichen um und versteckt allein das eine, das den
Umbruch verursacht hat — die Stelle heißt dort `hide_breaker`. Jedes andere
Leerzeichen ist ein Zeichen wie jedes andere und wird gesetzt.

**Hier stand bis zum 2026-08-16 das Gegenteil**, und zwar nur als Testfall:
„mehrere Leerzeichen zählen als eines". Das war unsere Annahme und keine
Beobachtung — der alte Umbruch benutzte `strings.Fields`, das jede Folge von
Leerraum zusammenzieht, und der Test schrieb das Werkzeug fest statt das Format.

**Quelle:** MSE2 `src/render/text/viewer.cpp`, Zeile 589, Fassung f9d9356 vom
2021-07-06
**Erschlossen am:** 2026-08-16
**Sicherheit:** belegt
**Testfall:** TestDerUmbruchFaelltAnDenErwartetenStellen

## Ein Wort, das allein nicht passt, wird zwischen zwei Zeichen gebrochen

MSE prüft den Umbruch **Zeichen für Zeichen**, nicht Wort für Wort. Reicht
ein Wort, das eine Zeile beginnt, über ihren rechten Rand hinaus, erzwingt es
dort einen Umbruch: das Wort bis zum vorigen Zeichen bleibt, das
überstehende Zeichen beginnt die nächste Zeile. **Ist es das einzige auf der
Zeile, bleibt es stehen** — sonst käme der Umbruch nie voran. Aufgegeben
wird nie: `stop_if_too_long` ist an jedem Aufruf `false`.

Ein **Symbolzug** bleibt hier ganz: er ist beim Umbruch ein Stück. Bei MSE
ist jedes Symbol ein eigenes Zeichen; ein Zug aus mehreren, der allein nicht
in eine Zeile passt, steht hier über — im Pack ist keiner so lang.

**Und der rechte Absatzrand zählt dabei nicht.** Gebrochen wird gegen
`lineRight` — den Rand der Maske abzüglich `padding_right`; `margin_right`
wirkt erst beim Abrücken einer Zeile (`fitLineWidth`) und beim Ausrichten
(`alignHorizontal`). Der linke Rand dagegen verschiebt den Zeilenanfang und
verkürzt so die Zeile.

**Hier stand bis zum Nachlauf über E/F das Gegenteil**, beides erschlossen:
ein zu langes Wort rage hinaus, weil eine Trennung eine Zeile ergäbe, die
niemand geschrieben hat, und der rechte Rand verkürze die Zeile wie der
linke.

**Quelle:** MSE2 `src/render/text/viewer.cpp`, Zeilen 453–541, 548, 614–633,
651–658, 739–741 und 834–835, Fassung f9d9356 vom 2021-07-06
**Erschlossen am:** 2026-10-11
**Sicherheit:** belegt
**Testfall:** TestEinZuLangesWortWirdZwischenZeichenGebrochen,
TestDerRechteAbsatzrandVerkuerztDenUmbruchNicht

## Ein Symbol ist nicht quadratisch — `image font size` und `horizontal space`

Die Größe, mit der MSE ein Symbol zeichnet, ist **nicht** die Schriftgröße im
Quadrat. Sie ist die **tatsächliche Bildgröße**, verhältnisgleich umgerechnet:

```cpp
RealSize SymbolInFont::size(Package& pkg, double size) {
  …
  return wxSize(actual_size * (int) (size) / (int) (img_size));
}
```

`img_size` steht in der `symbol-font`-Datei als **`image font size:`** — bei
`magic-mana-small` ist es **135**, und einzelne `symbol:`-Blöcke überschreiben
es (`image font size: 120`). Ein Bild von 135 × 135 bei Schriftgröße 14
erscheint also 14 × 14; ein Bild von 270 × 135 erschiene 28 × 14.

**Und zwar in ganzen Punkten**: `wxSize` rechnet ganzzahlig, die Schriftgröße
wird abgeschnitten (`(int) size`), und die Division ist eine ganzzahlige. Ein
Bild von 160 × 160 (`magic-mana-small`, `1.png`) bei 14 Punkten misst 16, nicht
16,59; bei 13,8 Punkten 15. Das Bild selbst legt `getImage` in
`(int) (actual * size / img_size)` an (Zeilen 115–127) — bei 13,8 also 16, und
damit größer als sein Kasten. Und die Schriftgröße ist die **verkleinerte**:
`SymbolTextElement::getCharInfo` gibt `font.size * scale` weiter
(`src/render/text/symbol.cpp`, Zeilen 22–26).

Dazu kommt je Symbol ein **Abstand**:

```cpp
return add_diagonal(sym.symbol->size(*this, font_size), spacing);
```

`spacing` sind `horizontal space:` und `vertical space:` aus derselben Datei —
bei `magic-mana-small` ist `horizontal space: 2`. **Ohne Angabe je 1**:
`spacing(1,1)` (`SymbolFont::SymbolFont`, Zeile 28). 75 der 76 Symbolschriften
des vollen Packs nennen keinen senkrechten Abstand — bei ihnen ist eine Zeile,
die ein Symbol hebt, einen Punkt höher als das Symbol.

**Gefragt über einen Orakelstil** am 2026-10-11: `<sym>1</sym>` aus
`magic-mana-small` bei 14 Punkten, Schrift 6 Punkte — MSE2 meldet
`content_height` 17 und `content_width` 18. **Bis dahin rechnete dieses
Programm mit Bruchteilen, ohne senkrechten Abstand und mit unverkleinerten
Symbolen** (16,59 hoch); bei „Icatian Infantry" stand jede Zeile ab der
zweiten zwei Punkte tiefer als bei MSE2
(#2300).

Und **innerhalb** seines Kastens wird das Bild ausgerichtet
(`align_in_rect(align, bmp_size, sym_rect)`), mit der `alignment:` der
Symbolschriftangabe des Feldes.

**Dieses Programm rechnete bis zum 2026-08-17 mit einem Quadrat der
Symbolgröße** — ohne Seitenverhältnis, ohne `image font size`, ohne Abstand,
ohne Ausrichtung im Kasten. Das gilt für die Manakosten (`symbolebenen`) wie
für die Symbole im Regeltext. Der Fehler ist gemessen: drei der vier
Referenzabstände stiegen, als die Symbole überhaupt anfingen zu zeichnen
(#330).

**Quelle:** MSE2 `src/data/symbol_font.cpp`, Zeilen 26–31, 50–52, 115–153,
279–294 und 384–401, und `src/render/text/symbol.cpp`, Zeilen 22–26, Fassung
f9d9356 vom 2021-07-06; dazu das Datenpaket
`magic-mana-small.mse-symbol-font/symbol-font`, Zeilen 41–42
**Erschlossen am:** 2026-08-16, die ganzen Punkte und die Vorgabe des Abstands
am 2026-10-11
**Sicherheit:** belegt, die Abstände und Maße auch gefragt
**Testfall:** TestEinSymbolkastenFolgtDemBildUndSeinemAbstand,
TestDieEigeneBildgroesseEinesSymbolsSchlaegtDieDerSchrift,
TestEinHohesSymbolMachtDieZeileHoeher, TestEinSymbolMisstInGanzenPunkten,
TestEinVerkleinerterTextVerkleinertSeineSymbole,
TestOhneAngabeBleibtDieBildgroesseNull

## Die Maske eines Textfelds bestimmt die Breite jeder Zeile

`mask:` an einem Textfeld beschneidet den Text nicht nur — **MSE bricht jede
Zeile an der Breite um, die die Maske an ihrer Oberkante lässt**:

```cpp
double TextViewer::lineLeft(RotatedDC& dc, const TextStyle& style, double y) const {
  return style.mask.getFromCache().rowLeft(y, dc.getInternalSize()) + style.padding_left;
}
```

`src/render/text/viewer.cpp`, Zeilen 736–741; gefragt wird mit `line.top`
beim Umbruch (Zeile 616), also vor Dehnung und Ausrichtung. Passt an einer
Stelle gar nichts, rückt die Zeile Punkt für Punkt hinunter (`fitLineWidth`,
Zeilen 546–554).

Die Maske wird dafür **auf die Größe des Felds gebracht** (`conform_image` in
`ScriptableImage::generate`, `src/script/image.cpp`, Zeilen 20–37, mit MSEs
Kastenfilter `resample`), und je Bildzeile zählen der erste und der letzte
Punkt, dessen **Rotkanal** mindestens 128 ist (`AlphaMask::loadRowSizes`,
`src/gfx/mask_image.cpp`, Zeilen 145–160). Unter dem Feld gilt wieder die
ganze Breite. **Auch eine ganz weiße Zeile endet einen Punkt vor dem Rand** —
`rights` ist der Index des letzten weißen Punkts.

**Wo es greift**: 52 Textfelder in 43 Stilen des vollen Packs tragen `mask:`,
fast alle Planeswalker. Bei `m15-mainframe-planeswalker` ist
`walker_text_mask.png` rechts unten neben dem Treueschild schwarz — der
letzte Absatz von „Saheeli, the Gifted" bricht dort um.

**Quelle:** MSE2 `src/render/text/viewer.cpp`, Zeilen 546–554, 557–735 und
736–741; `src/gfx/mask_image.cpp`, Zeilen 29–46 und 145–178;
`src/script/image.cpp`, Zeilen 20–37; `src/render/value/viewer.cpp`, Zeilen
94–100; `src/gfx/resample_image.cpp`, Zeilen 18–153 — Fassung f9d9356 vom
2021-07-06
**Erschlossen am:** 2026-10-11
**Sicherheit:** belegt
**Testfall:** TestDieMaskeBestimmtDieBreiteDerZeilen,
TestDerLinkeRandDerMaskeVerschiebtDieZeile,
TestEineLeereMaskenzeileSchiebtDieZeileHinunter,
TestDieMaskeWirdAufDasFeldGebracht, TestDieMaskeEinesTextfeldsReistZumSatz

## Ein Symbol sitzt mittig in seiner Zeile, nicht auf ihrer Oberkante

`SymbolFont::drawSymbol` richtet das Bild **in seinem Kasten** aus:

```cpp
RealSize  bmp_size = dc.trInvS(RealSize(bmp));
RealPoint bmp_pos  = align_in_rect(align, bmp_size, sym_rect);
```

Der Kasten kommt aus `split_left(rect, size)`, und `split_left` nimmt nur
Breite weg — die **Höhe bleibt die des Zeilenrechtecks**
(`RealRect rect(l.positions.front(), l.top, l.width(), l.line_height)`).
Senkrecht ausgerichtet wird also über die ganze Zeilenhöhe, und die Vorgabe
ist `ALIGN_MIDDLE_CENTER` (`SymbolFontRef::SymbolFontRef`).

Die Zeilenhöhe wiederum ist das Größte, was in der Zeile steht:
`line.line_height = line_size.height` über `add_horizontal`, das die Höhen
maximiert. Ein Symbol, das höher ist als die Schrift, **hebt die Zeile**.

**Quelle:** MSE2 `src/data/symbol_font.cpp`, Zeilen 279–294 und 553–557;
`src/render/text/viewer.cpp`, Zeilen 90–94, 603–608 und 663–684;
`src/util/real_point.hpp`, Zeilen 83–85 und 199–204 — Fassung f9d9356 vom
2021-07-06
**Erschlossen am:** 2026-08-17
**Sicherheit:** belegt
**Testfall:** TestEinSymbolImTextSitztMittigInDerZeile,
TestEinSymbolzugSitztDortWoDasFeldIhnHinstellt

---

## Fett ist ein Gewicht, kursiv ist ein Name

MSE behandelt die beiden Auszeichnungen **nicht** gleich. `Font::toWxFont`:

```cpp
static const String BOLD_STRING = _(" Bold");
…
} else if (flags & FONT_ITALIC && !italic_name().empty()) {
  font = wxFont(size_i, …, weight_i, underline(), italic_name());
} else {
  String familyName = name();
  if (familyName.EndsWith(BOLD_STRING)) {
    familyName = familyName.Left(familyName.length() - BOLD_STRING.length());
    weight_i = wxFONTWEIGHT_BOLD;
  }
  font = wxFont(size_i, …, style_i, weight_i, underline(), familyName);
}
```

- **Kursiv ist ein Name:** die Schrift aus `italic name:` des Stils.
- **Fett ist ein Gewicht:** Familienname plus `wxFONTWEIGHT_BOLD`, und das
  Betriebssystem sucht den Schnitt.
- **Beides zusammen** heißt: die Schrift aus `italic name:` **mit** dem fetten
  Gewicht — `style_i` bleibt in diesem Zweig ungenutzt.
- **Ein Name auf ` Bold` ist fett.** MSE schneidet den Zusatz ab und setzt das
  Gewicht. `magic-m15` schreibt `Beleren Bold`; ein `<b>` an einem solchen Feld
  ändert damit nichts mehr.

Im Kartentext wird `<b>` wie `<i>` **gezählt**, nicht geschaltet, und wie dort
über ein **Präfix** erkannt: `is_tag(text, pos, "<b")`. Damit setzt auch
`<bullet>` fett — es steht in derselben Kette vor dem eigenen
`</bullet`-Zweig.

Dieses Programm sucht Schriften über Namen (`internal/schriften`); wie aus dem
Gewicht wieder ein Name wird und was das kostet, steht in
[0023](entscheidungen/0023-fett-wird-ueber-den-namen-gesucht.md).

### Und `weight:` und `style:` im `font:`-Block

**Das Gewicht steht nicht nur im Text, sondern in der Schrift.** `Font::update`
macht aus `weight: bold` den Schalter `FONT_BOLD` und aus `style: italic`
`FONT_ITALIC` — für **jedes** Zeichen des Feldes —, und `<b>` und `<i>` legen
ihren Schalter dazu (`f->flags |= add_flags`):

```cpp
flags = (flags & ~FONT_BOLD & ~FONT_ITALIC)
      | (weight() == _("bold")   ? FONT_BOLD   : FONT_NORMAL)
      | (style()  == _("italic") ? FONT_ITALIC : FONT_NORMAL);
```

- **Verglichen wird genau**: `Bold` ist nicht fett. Im Pack steht überall
  `bold`.
- **Beide sind `Scriptable<String>`** — `{if loy_leveler() then "" else "bold"}`
  steht elfmal da.
- **768 `weight:` in 207 Stilen** des vollen Packs, 16 `style:` in 15 —
  darunter der Name jeder Archenemy-Karte (`name: Matrix`, `weight: bold`).
  Bis F02 (#2244) las dieses Programm keines von beiden.
- **`Matrix` fett ist `matrixb.ttf`**: Familie „Matrix", Unterfamilie „Bold",
  voller Name „Matrix-Bold". Die Frage „Matrix Bold" (0023) trifft das nur,
  weil die Suche jede Datei auch unter Familie und Schnitt zusammen führt.

### Und der Schatten

`shadow color:`, `shadow displacement x:`/`y:` und `shadow blur:` gehören in
denselben Block. **Nur die Farbe ist ein Skript**, und ihre Vorgabe ist
**durchsichtig** (`Color()`), nicht schwarz wie die der Schrift. MSE zeichnet
je Textelement erst den Schatten — um den Versatz verschoben, `shadow blur`-mal
mit `(2·Mitte + vier Nachbarn) / 6` an Ort und Stelle geweicht —, dann den
Text (`DrawTextWithShadow`). Ein durchsichtiger Schatten zeichnet nichts; im
Pack steht er meist als Skript, das für gewöhnliche Rahmen `rgba(0,0,0,0)`
gibt. **498 `shadow color:` in 107 Stilen**, `magic-old` schattet Name und
Stärke um (1, 1).

**Quelle:** MSE2 `src/data/font.cpp`, Zeilen 26–41, 54–56 und 118–134;
`src/util/rotation.cpp`, Zeilen 221–224; `src/gfx/resample_text.cpp`, Zeilen
144–201 —
Fassung f9d9356 vom 2021-07-06; die Namen der Matrix-Dateien gelesen mit
`fc-list` am 2026-10-11
**Erschlossen am:** 2026-10-11
**Sicherheit:** belegt
**Testfall:** TestEineSchriftNenntGewichtUndLage,
TestGewichtUndLageDerSchriftGeltenFuerJedenAbschnitt,
TestFettFindetDenSchnittBoldDerFamilie, TestEinSchattenSitztUmSeinenVersatz,
TestWeichzeichnenRechnetWieMSE

**Quelle:** MSE2 `src/data/font.cpp`, Zeilen 82–108, und
`src/render/text/element.cpp`, Zeilen 78–81 und 165–170 — Fassung f9d9356 vom
2021-07-06
**Erschlossen am:** 2026-08-17
**Sicherheit:** belegt
**Testfall:** TestFetteStellenWerdenEigeneAbschnitte,
TestEinFetterAbschnittWirdMitDemFettenSchnittGemessen,
TestKursivUndFettSindEinDritterSchnitt,
TestFettSuchtDenFettenSchnittDerFamilie

---

## `content_width`, `content_height` und `content_lines` kommen aus dem Satz

Sie sind **Rückmeldungen**, keine Angaben: erst wird der Text gesetzt, dann
steht fest, wie breit, wie hoch und wie viele Zeilen es sind.

```cpp
REFLECT_N("content_width",  layout->width);
REFLECT_N("content_height", layout->height);
REFLECT_N("content_lines",  layout->lines.size());
```

Der `layout` entsteht in `TextViewer::extractLayoutInfo` aus den gesetzten
Zeilen; `width` ist das Maximum der **Zeilenbreiten** nach dem Umbruch,
`height` die Unterkante der letzten Zeile:

```cpp
layout.width  = max(layout.width,  l.width());
layout.height = max(layout.height, l.bottom() - layout.top);
```

Gezählt über die 190 Stile von `M15-Magic-Pack`: **604 × `content_width`**
(199 davon für `casting_cost`), **221 × `content_lines`** (174 davon für
`text`), 5 × `content_height`. `magic-m15` entscheidet mit `content_lines` die
Ausrichtung des Regeltextes und mit `casting_cost.content_width` die rechte
Kante des Kartennamens.

**Für ein Symbolfeld ist die Breite die Summe der Symbolkästen** — und ein
Symbolkasten hat die Größe seines Bildes, nicht die der Schrift (siehe „Ein
Symbol ist nicht quadratisch"). Wer dort mit Quadraten rechnet, misst bei
`magic-mana-large` in Größe 15 gut ein Viertel zu wenig.

**Das `layout`-Objekt** gibt es seit dem 2026-08-17: `lines`, `paragraphs`,
`blocks`, `separators`, je Eintrag mit `width`, `top`, `height`, `bottom` und
`middle`. Es steht 103-mal im Pack, davon 30-mal `blocks`.

- Ein **Absatz** endet an einem harten Umbruch, ein **Block** an einem `<line>`.
- Die Maße einer Gruppe folgen `update_size`: die Breite ist die **größte**
  Zeilenbreite, die Höhe reicht von der Oberkante der ersten bis zur Unterkante
  der letzten Zeile.
- Ein **Trenner** liegt in der Mitte zwischen zwei Blöcken:
  `(blocks[i]->bottom() + blocks[i+1]->top)/2`.

**`<line>` schneidet den Block.** MSE zählt die Auszeichnung wie `<i>` und
`<b>` und gibt dem Element den Umbruch `LineBreak::LINE`
(`src/render/text/element.cpp`, Zeilen 84–85 und 235–236); im Satz endet daran
ein Block.

**Woher `<line>` kommt, ist der eigentliche Punkt.** Das Feld `text` von
`magic.mse-game` wird nicht gespeichert (`save value: false`), sondern
gerechnet:

```
script:
	combined_editor(field1: card.rule_text, separator: "<line>\n</line>", field2: card.flavor_text)
```

Regeltext, Blockgrenze, Flavortext — deshalb rechnen die Planeswalker-Stile
mit `layout.blocks[i]`, und deshalb steht `separators` überhaupt zur Verfügung.

**Seit dem 2026-08-17 rechnet dieses Programm das Feld aus.** Der Weg dorthin
steht in der nächsten Feststellung.

**Quelle:** MSE2 `src/data/field/text.cpp`, Zeilen 114–133,
`src/data/field/text.hpp`, Zeilen 50–67, und `src/render/text/viewer.cpp`,
Zeilen 350–388 — Fassung f9d9356 vom 2021-07-06
**Erschlossen am:** 2026-08-17
**Sicherheit:** belegt
**Testfall:** TestDieZeilenzahlKommtAusDemUmbruch,
TestDieInhaltsbreiteEinesSymbolfeldsMisstDieBilder,
TestDieInhaltsbreiteEinesSymbolfeldsWirdGemessen,
TestDasLayoutObjektStehtDerVorlageZurVerfuegung

---

## Der Vorgabewert eines Feldes kommt aus drei Stufen

Ein Kartenfeld, das die Karte nicht ausgefüllt hat, ist bei MSE nicht leer:

```cpp
ChoiceValue::ChoiceValue(const ChoiceFieldP& field, bool initial_first_choice)
  : Value(field)
  , value( !field->initial.empty() ? field->initial
         : initial_first_choice    ? field->choices->choiceName(0)
                                   : String())
```

Also: **`initial:`, sonst die erste `choice:`, sonst leer** — und über allem
steht `default:`, das ein Skript ist.

Gezählt über `M15-Magic-Pack`: 142 der 249 Kartenfelder von `magic.mse-game`
tragen ein `default:`, **17** ein `initial:`. Darunter `rarity`, dessen erste
Wahl `basic land` heißt und dessen `initial:` `common` lautet — ohne die zweite
Stufe trägt jede Karte ohne ausdrückliche Seltenheit ein Landsymbol.

**Die Wahlliste darf aus einem anderen Paket kommen.** `card symbol` schreibt
`include file: /magic-modules.mse-include/card-symbols/choices`; die erste Wahl
dort ist `none`, und `magic-m15` fragt genau danach:

```
name:
	left: { if card.card_symbol == "none" then 32 else 50 }
```

**Quelle:** MSE2 `src/data/field/choice.cpp`, Zeilen 279–283, Fassung f9d9356
vom 2021-07-06; dazu `magic.mse-game/card_fields`
**Erschlossen am:** 2026-08-17
**Sicherheit:** belegt
**Testfall:** TestEinAuswahlfeldNenntSeinenAnfangswert,
TestEinFehlendesFeldBekommtDenVorgabewert

---

## Ein Feldwert entsteht aus **zwei** Skripten

`TextValue::update` wendet beide an, und die Reihenfolge ist die Aussage:

```cpp
bool change = field().default_script.invokeOnDefault(ctx, value)
            | field().        script.invokeOn(ctx, value);
```

- **`default:` füllt** einen Wert, der leer ist (`invokeOnDefault` prüft
  `value.isDefault()`).
- **`script:` rechnet** ihn — immer, und der bisherige Wert steht ihm als
  `value` zur Verfügung (`ctx.setVariable(SCRIPT_VAR_value, …)`).

**Ein Feld kann damit gar nicht gespeichert sein.** `text` von
`magic.mse-game` trägt `save value: false`; in einem `.mse-set` steht es nicht,
und es entsteht erst hier:

```
script:
	if mainframe_walker() then mainframe_walker_text_script()
	else if alt_text() then alt_text_script()
	else combined_editor(field1: card.rule_text, separator: "<line>\n</line>", field2: card.flavor_text)
```

**Gezeichnet wird der gerechnete Wert, nicht der gespeicherte.** Wer die rohe
Karte liest, zeichnet für `text` ein leeres Feld.

**`combined_editor` setzt die Teile mit `<sep>`-Klammern zusammen:**

```cpp
new_value += _("<sep>") + separators[i - 1] + _("</sep>");
```

Die Nummerierung beginnt bei **null ohne Zusatz** — `field`, `field1`,
`field2`, ebenso `separator`, `separator1` —, und die Reihe endet an der ersten
Lücke. `hide_when_empty` lässt einen Trenner nur weg, solange **alles** davor
leer ist.

**Ein Wertskript, das scheitert, lässt den Wert stehen.** `invokeOn` weist nur
zu, was herauskommt. Das ist die Rettung für die **72** Kartenfelder von
`magic.mse-game`, deren Skript das noch nicht gebaute `text_filter` ruft — eine
Kette aus rund dreißig teilangewandten Funktionen, mit `+` zu einer Funktion
komponiert (`ScriptCompose`).

**Quelle:** MSE2 `src/data/field/text.cpp`, Zeilen 178–179;
`src/script/scriptable.hpp`, Zeilen 50–79; `src/script/functions/editor.cpp`,
Zeilen 28–140; `src/script/context.cpp`, Zeilen 383–400 und 466–472 — Fassung
f9d9356 vom 2021-07-06; dazu `magic.mse-game/card_fields`, Zeilen 1339–1360
**Erschlossen am:** 2026-08-17
**Sicherheit:** belegt
**Testfall:** TestEinFeldMitWertskriptWirdGerechnet,
TestCombinedEditorSetztFelderZusammen, TestEinWertskriptWirdAlsBlockGelesen

---

## Polster und Zeilenhöhen: der Kasten ist kleiner, der Abstand ungleich

**Das Polster schneidet die Fläche, bevor gesetzt wird.**

```cpp
line.top = style.padding_top;
lineLeft  = … + style.padding_left;
lineRight = … - style.padding_right;
RealSize s = add_diagonal(dc.getInternalSize(),
    -RealSize(padding_left + padding_right, padding_top + padding_bottom));
```

Umgebrochen wird also auf `breite − links − rechts`, die erste Zeile beginnt
bei `padding_top`, und die Ausrichtung rechnet mit der **verkleinerten**
Fläche.

**Und der Abstand zur nächsten Zeile hängt daran, wie die Zeile endet:**

```cpp
double line_height_multiplier = line.break_after == LineBreak::HARD ? style.line_height_hard
                              : line.break_after == LineBreak::LINE ? style.line_height_line
                              :                                       style.line_height_soft;
line.top += line.line_height * line_height_multiplier;
```

`magic-m15` schreibt **1,2** für den harten Umbruch, **1,5** für `<line>` und
**0,9** für den weichen: ein umgebrochener Satz steht dort enger als zwei
Absätze, und vor dem Flavortext ist die größte Lücke. Ohne Angabe ist jeder
Faktor **1,0** (`src/data/field/text.cpp`, Zeilen 62–64) — ein Nullwert wäre
keine Höhe und staucht den Zug auf eine Zeile.

**Der Faktor gilt dem Abstand zwischen zwei Zeilen**, nicht der Höhe der
letzten: `alignParagraph` misst die Höhe des Zugs bis zur **Unterkante** der
letzten Zeile.

Gezählt über die 190 Stile von `M15-Magic-Pack`: `padding top:` 279-mal,
`padding right:` 215-mal, `padding bottom:` 228-mal, `padding left:` 203-mal;
die drei Zeilenhöhen je rund 290-mal.

**`line height … max` dehnt einen zu kurzen Zug auf die Kastenhöhe.** Je
Umbruchart gibt es einen Spielraum `d = max(0, max − faktor)`; die drei werden
sortiert und in Stufen verteilt, begrenzt auf das, was der Kasten hergibt:

```cpp
double to_add = min(stop, (s.height - height) / sum);
l.top += add;  if (matching) add += to_add * l.line_height;
```

**Verglichen wird der Spielraum mit dem Zuwachs der Stufe, nicht mit dem
erreichten Stand** (`bool soft = d_soft >= stop;`) — eine Art kann damit über
ihren eigenen Höchstwert hinauswachsen, wenn die Stufen ungleich groß sind.
Nachgebildet und nicht behoben: die Parität ist das Ziel.

**Gedehnt wird vor dem Ausrichten**, und der senkrechte Versatz rechnet mit der
gedehnten Höhe. `hard max` und `line max` stehen je 271-mal im Pack,
`soft max` 22-mal; `magic-m15` gibt dem Regeltext 1,3 und 1,6 gegen 1,2 und
1,5.

**Quelle:** MSE2 `src/render/text/viewer.cpp`, Zeilen 567, 678–681, 736–748 und
768–800; `src/data/field/text.cpp`, Zeilen 62–67 und 155–162 — Fassung f9d9356
vom 2021-07-06
**Erschlossen am:** 2026-08-17
**Sicherheit:** belegt
**Testfall:** TestDasPolsterSchneidetDenKasten,
TestDieZeilenhoeheHaengtAmUmbruch, TestOhneZeilenfaktorenGiltEins,
TestEinFeldstilNenntPolsterUndZeilenhoehen, TestEinZuKurzerZugWirdGedehnt,
TestGedehntWirdNurSoWeitDerKastenReicht,
TestDerGedehnteZugWirdDanachAusgerichtet

---

## `alignment:` ist ein Skript, und der Wert des Knotens ist dann leer

MSE führt die Ausrichtung als `Scriptable<Alignment>` und rechnet sie mit jedem
Wechsel neu (`TextStyle::update`, `src/data/field/text.cpp`, Zeile 95).
`magic-m15` entscheidet damit, ob ein kurzer Regeltext mittig steht:

```
alignment:
	script:
		if (styling.center_text == "short text only" and
		    not contains(match:"\n", card.rule_text) and
		    card.flavor_text == "<i-flavor></i-flavor>" and
		    card_style.text.content_lines <= 2) or
		    styling.center_text == "always"
		then "middle center"
		else "middle left"
```

**Zwei Formen, und die zweite hat einen leeren Wert.** Steht die Angabe in
einer Zeile (`alignment: bottom shrink-overflow`), ist sie der Wert des
Knotens. Steht darunter ein `script:`-Block, ist der Wert **leer** und der Text
liegt im Kind. Wer nur den Wert läse, bekäme die Vorgabe — links oben.

Gezählt über die 190 Stile von `M15-Magic-Pack`: **1.597 Ausrichtungen, davon
216 als Skript.**

**Und `shrink-overflow` steckt dann im Ergebnis**, nicht im Quelltext:
`if styling.center_name then "center middle shrink-overflow" else "left middle
shrink-overflow"`. Wer den Quelltext nach dem Wort durchsucht, findet es
zufällig — und bei einem Skript ohne das Wort gar nicht.

**Quelle:** MSE2 `src/data/field/text.cpp`, Zeile 95, Fassung f9d9356 vom
2021-07-06; dazu `magic-m15.mse-style`, Zeilen 286–294
**Erschlossen am:** 2026-08-17
**Sicherheit:** belegt
**Testfall:** TestEineAusrichtungDarfEinSkriptSein,
TestDasStauchenStehtImErgebnisDesSkripts,
TestEineAusrichtungAlsSkriptblockWirdGelesen

---

## Ein Argument ohne Namen heißt `input` — und `+` verkettet Funktionen

**MSE kennt kein Stellungsargument.** `f(x)` setzt `x` als Kontextvariable
`input`, und jede eingebaute Funktion liest es dort
(`SCRIPT_PARAM_C(String, input)`). Die beiden Schreibweisen sind dieselbe
Sache.

**Sichtbar wird das an der Verkettung.** `+` auf zwei Funktionen ergibt eine
`ScriptCompose`, die das Zwischenergebnis als `input` weitergibt
(`src/script/context.cpp`, Zeilen 383–400 und 466–472):

```cpp
} else if (at == SCRIPT_FUNCTION && bt == SCRIPT_FUNCTION) {
  a = make_intrusive<ScriptCompose>(a, b);
```

`text_filter` in `magic.mse-game` reiht so rund **dreißig** teilangewandte
Funktionen — `remove_tag@(tag: …) + replace@(…) + expand_keywords@(…) + …` —
und ist das `script:` von **72 Kartenfeldern**.

**Die Kette entsteht beim Auswerten der Zuweisung.** Fehlt ein Name darin,
scheitert sie als Ganzes; das war der Grund, warum `expand_keywords` 190 Läufe
kostete und nicht nur einen.

**`expand_keywords` mit leerer Wortliste** räumt die alten Erinnerungstexte weg
und gibt den Rest zurück:

```cpp
String tagged = remove_keyword_tags(text);
if (!root) return tagged;
```

`remove_keyword_tags` nimmt den **Inhalt** von `<atom-reminder…>`,
`<atom-keyword…>` und `<atom-kwpph>` mit, lässt aber bei `<keyword-param…>`
und `<param-…>` den Inhalt stehen. Die Schlüsselwörter eines Spiels liest
dieses Programm noch nicht; solange das so ist, ist die Datenbank leer, und
dann ist genau das MSEs Verhalten.

**Quelle:** MSE2 `src/script/context.cpp`, Zeilen 383–400 und 466–472;
`src/data/keyword.cpp`, Zeilen 756–799; `src/script/functions/basic.cpp`,
Zeilen 661–683 — Fassung f9d9356 vom 2021-07-06; dazu
`magic.mse-game/script`, Zeilen 1679–1730
**Erschlossen am:** 2026-08-17
**Sicherheit:** belegt
**Testfall:** TestEinArgumentOhneNamenHeisstInput,
TestEinStellungsargumentSchlaegtDasBenannte,
TestExpandKeywordsRaeumtDieErinnerungstexteWeg,
TestExpandKeywordsReihtSichInEineKette

---

## Eine Wahlliste ist ein Text, und `filter_choices` siebt ihn

Ein `choice`-Feld mit Mehrfachauswahl trägt seine Werte als `"a, b, c"`.
`filter_choices` geht die Liste durch, **merkt** sich die genannten Wahlen und
nimmt sie aus dem Ergebnis; am Ende hängt es sie wieder an — ergänzt bis `min`,
gekürzt auf `max` (`src/script/functions/editor.cpp`, Zeilen 248–308, Fassung
f9d9356 vom 2021-07-06).

- `remove_choice` ruft es mit `min = max = 0`,
- `require_exclusive_choice` mit `min = max = 1` und `last_change:` als
  bevorzugter Wahl.

**Die Reihenfolge ändert sich dabei:** die unbeteiligten Wahlen stehen danach
vorn, die behaltenen hinten. Das ist kein Nebeneffekt, sondern die Bauart der
Funktion.

**Ergänzt wird die bevorzugte zuerst, gekürzt wird von hinten** — und die
bevorzugte fällt zuletzt (`if (max > 0 && choices[i] == prefered) continue;`).

**Genannt werden die Wahlen in drei Schreibweisen**, in dieser Reihenfolge
gelesen: `choices:` als Liste, `choice:` einzeln, und die nummerierten
`choice1:`, `choice2:` … Die Reihe der nummerierten endet an der ersten Lücke.

**Ohne genannte Wahlen bleibt die Eingabe** — MSE schreibt dazu selbst
„shouldn't happen, but better make sure".

**Quelle:** MSE2 `src/script/functions/editor.cpp`, Zeilen 216–224 und 248–363
— Fassung f9d9356 vom 2021-07-06
**Erschlossen am:** 2026-08-17
**Sicherheit:** belegt
**Testfall:** TestWahlenWerdenGesiebt,
TestDieAlteSchreibweiseDerWahlenGiltWeiter,
TestBeimKuerzenBleibtDieBevorzugte

---

## Eine Zeichenkette baut Ausdrücke ein — `"a{e}b"`

**Geschweifte Klammern in einer Zeichenkette sind ein Ausdruck**, kein Text.
Der Zerteiler macht daraus eine Verkettung, noch bevor irgendetwas gerechnet
wird (`src/script/parser.cpp`, Zeilen 219–252, Fassung f9d9356 vom
2021-07-06):

```cpp
case '{':
  parseOper(input, script, PREC_ALL, TOK_RBRACE, _("}"));
  script.addInstruction(I_BINARY, I_ADD);
```

`"<{tag}>{input}</{tag}>"` ist also `"<" + tag + ">" + input + "</" + tag +
">"` — und ohne diesen Schritt steht die geschweifte Klammer wörtlich auf der
Karte. Auf *Old-Money* war das die ganze Typzeile.

**`\{` schützt eine Klammer** und fällt dabei weg.

**Eine Zeichenkette darf in der Einbettung stehen** —
`"{if a then "x" else "y"}"`. MSE2 hält dafür einen **Klammerstapel**: `"`
legt `BRACE_STRING` auf (Zeilen 220–222), ein `{` in der Zeichenkette kehrt
zum gewöhnlichen Lesen zurück, **ohne den Stapel zu räumen** (Zeilen
302–307), ein `"` darin legt einen zweiten darauf (Zeilen 268–272), und das
`}` liest die äußere weiter (Zeilen 224–228). Verschachtelung ist damit der
Normalfall.

Hier stand bis zum 2026-08-18 die Gegenrichtung als Abweichung. Sie ist
gebaut (#468);
gemessen wurden über `M15-Magic-Pack`, nur streng als UTF-8 dekodierbare
Textdateien, **357 Stellen** in 23 Dateien aus 4 Paketen — davon 195 in
`magic.mse-game/script`.

**Was weiter abweicht:** MSE2 bricht ab, wenn eine Zeichenkette in einer
Einbettung nicht schließt (`Unexpected end of input in string constant`,
Zeile 260); dieses Programm liest weiter und behandelt die halbe Klammer als
Text (§8). `\x{20AC}` wird ebenso als Text behandelt: der Ersatztext von
`replace` folgt sed, und die Klammern gehören dort zur Zeichennummer
(#345).

**Quelle:** MSE2 `src/script/parser.cpp`, Zeilen 219–252, Fassung f9d9356 vom
2021-07-06
**Erschlossen am:** 2026-08-17
**Sicherheit:** belegt
**Testfall:** TestEineZeichenketteDarfAusdrueckeEnthalten,
TestEineZeichenketteImAusdruckWirdGelesen,
TestZerlegeZeichenketteInEinbettung,
TestWasSichNichtZerteilenLaesstIstText,
TestEineKlammerOhneEndeBleibtStehen

---

## Der Rumpf einer Funktion läuft in der Umgebung des Aufrufers

**Nicht in der Bibliothek.** MSE bindet dynamisch: eine gerufene Funktion sieht
die Variablen der Stelle, an der sie gerufen wird
(`ScriptClosure::operator()` und `Context::eval`, `src/script/context.cpp`).
Für die Ketten aus `+` heißt das, dass jedes Glied `set`, `card` und die
übrigen Kontextvariablen weiterhin vor sich hat.

Wer den Rumpf stattdessen über der Bibliothek auswertet, verliert sie alle —
messbar als **5.076** Befunde `inserts_count` und **1.330** `set` über die 190
Stile von `M15-Magic-Pack`.

**Quelle:** MSE2 `src/script/context.cpp`, Zeilen 466–472, Fassung f9d9356 vom
2021-07-06
**Erschlossen am:** 2026-08-17
**Sicherheit:** belegt
**Testfall:** TestEinVerketteterBlockSiehtSeineUmgebung,
TestAuchDerLinkeTeilSiehtSeineUmgebung,
TestEineLangeKetteTraegtDieUmgebungDurch

---

## Ein `script:` steht einzeilig oder als Block — der Kommentar darunter nicht

`magic.mse-game` schreibt Wertskripte in beiden Formen. Bei der einzeiligen
(`script: card_color_filter(value)`) steht das Skript im **Wert** des Knotens;
darunter folgen im Pack regelmäßig Kommentarzeilen (`##### Name line`), die zu
seinen Zeilen zählen, aber nicht zum Skript. Wer den Block bevorzugt, zerteilt
einen reinen Kommentar — **3.990** Befunde „hier fehlt ein Ausdruck".

**Quelle:** `magic.mse-game/card_fields` aus `M15-Magic-Pack`
**Erschlossen am:** 2026-08-17
**Sicherheit:** belegt
**Testfall:** TestEinWertskriptStehtEinzeiligOderAlsBlock

---

## `card_style` führt jedes Kartenfeld — und rechnet drei seiner sechs Maße

**Die Abbildung entsteht aus den Feldern des Spiels, nicht aus denen der
Vorlage:**

```cpp
REFLECT_IF_READING {
  card_style.init(game->card_fields);
  …
}
REFLECT(card_style);
```

`IndexMap::init` legt für **jeden** Schlüssel einen Eintrag an; was die Vorlage
nicht beschreibt, bekommt einen frischen `Style` mit Nullmaßen, und erst danach
überschreibt `REFLECT`, was dasteht. `card_style.<feld>` ist deshalb nie
`nichts` — auch für ein Feld, das diese Vorlage nicht zeichnet.

**Ein Stil hat sechs Maße je Richtung-Paar**, und drei davon rechnet MSE aus:

```cpp
if      (automatic_side & AUTO_LEFT)   left   = right - width;
else if (automatic_side & AUTO_WIDTH)  width  = right - left;
else if (automatic_side & AUTO_RIGHT)  right  = left + width;
else   { int lr = int(left + right); left = (lr - width) / 2; right = (lr + width) / 2; }
```

Welches fehlt, entscheidet der **Wert**: ein nicht beschriebenes Maß trägt
1.000.000, und `is_set` heißt `x < 100000`. Für die Länge gilt `is_setw`,
`|x| > 0.001` — ihre Vorgabe ist die Null, und eine Breite von null ist keine
Breite. Ein Skript zählt in beiden Fällen als gesetzt.

**Wozu die Karte das braucht:** die Saga-Vorlagen stapeln ihre Stufen über
`top: { card_style.level_1_text.bottom }`.

**Gemessen** über die 190 Stile von `M15-Magic-Pack`: 3 Befunde
„`content_width` gibt es nur an einer Abbildung" und 12 „ergab keine Zahl,
sondern nichts" — beide Klassen auf null, Befunde 3.211 → 3.192. Die vier
Referenzkarten stehen alle auf `magic-m15`, das jedes seiner Felder stylt und
keine Saga ist; ihr **Abstand ändert sich deshalb nicht**.

**Was hier abweicht:** eine Maßangabe, die ein Skript ist, wiegt bei der
Ableitung null. MSE rechnet sie vorher aus, weil seine Skripte in Wellen über
Abhängigkeiten laufen (#291).

**Quelle:** MSE2 `src/data/stylesheet.cpp`, Zeilen 106–113;
`src/util/index_map.hpp`, Zeilen 48–61; `src/data/field.cpp`, Zeilen 100–172 —
Fassung f9d9356 vom 2021-07-06
**Erschlossen am:** 2026-08-17
**Sicherheit:** belegt
**Testfall:** TestCardStyleFuehrtJedesKartenfeldDesSpiels,
TestEinFeldstilRechnetSeineFehlendenMasseAus

---

## Ein Symbol wird als Maske gerastert, nicht in Farbe

**MSE zeichnet ein Symbol in zwei Hilfsflächen und bringt sie am Ende auf den
Schirm:**

```cpp
if (borders)  dc.Blit(…, borders,  0, 0, wxOR);
if (interior) dc.Blit(…, interior, 0, 0, wxAND_INVERT);
```

Der Schirm beginnt **grün** (`Color(0,128,0)`), der Rand kommt weiß darauf, das
Innere nimmt weg. Was am Ende dasteht, liest der Filter als drei Bereiche:

```cpp
SymbolSet point = data[1] ? (data[0] ? SYMBOL_BORDER : SYMBOL_OUTSIDE)
                          : SYMBOL_INSIDE;
```

**Der Randstrich sitzt mittig auf der Kante**, und seine innere Hälfte wird vom
Inneren wieder überschrieben — sichtbar ist nur die äußere. `border_radius` ist
ein Anteil der Bildgröße, nicht ein Maß in Bildpunkten.

**Sechs `combine:`-Arten**, und `border` ist die, deren Name täuscht: sie malt
die **ganze Form** in den Randpuffer, nicht ihren Umriss.

**Die Teile werden von hinten nach vorn gezeichnet** (`FOR_EACH_CONST_REVERSE`)
— was in der Datei oben steht, liegt beim Zeichnen obenauf und verrechnet sich
mit dem darunter.

**Ohne Glättung**: jeder Punkt wird über `to_int` auf einen Bildpunkt gerundet,
und `to_int` rundet knapp unter der Hälfte (`d + 0.4999995`).

**Erst der Filter macht Farbe** — voll, linearer oder kreisförmiger Verlauf;
außen ist immer `Color(0,0,0,0)`.

**Was hier abweicht:** die Füllregel an Kanten, die genau auf einem Bildpunkt
liegen, ist bei wxWidgets plattformabhängig; gemessen wird deshalb in der Mitte
eines Bildpunkts. Und MSE zeichnet die erste Form unmittelbar auf den Schirm
und puffert erst ab der zweiten — hier wird immer gepuffert, mit demselben
Ergebnis (#323).

**Quelle:** MSE2 `src/render/symbol/viewer.cpp`, Zeilen 16–42 und 85–287;
`src/render/symbol/filter.cpp`, Zeilen 17–186; `src/gfx/bezier.cpp`, Zeilen
16–95; `src/util/vector2d.hpp`, Zeilen 20–22 und 148–164 — Fassung f9d9356 vom
2021-07-06
**Erschlossen am:** 2026-08-17
**Sicherheit:** belegt
**Testfall:** TestEinSymbolHatDreiBereiche, TestDieKombinationsartenEinerForm,
TestUmrandenMachtDieGanzeFormZumRand,
TestEineDrehsymmetrieVervielfaeltigtIhreKinder,
TestEineSpiegelsymmetrieSpiegeltJedeZweiteKopie,
TestDerVollfilterGibtZweiFarbenUndDurchsicht,
TestDerSymbolrasterlaufBleibtWieErIst

---

## MSE hat **eine** Variablentabelle — die Ebene wechselt nur beim Aufruf

```cpp
case I_CALL:
  new_scope.reset(new LocalScope(*this)); //new scope
```

```cpp
void Context::setVariable(Variable name, const ScriptValueP& value) {
  VariableValue& var = variables[name];
  if (var.level < level) {
    Binding bind = {name, var};   // keep shadow copy
    shadowed.push_back(bind);
  }
  var.level = level;
  var.value = value;
}
```

**Eine Schleife öffnet keine Ebene, eine geklammerte Gruppe auch nicht.** Eine
Zuweisung im Rumpf schreibt in denselben Platz und gilt damit im **nächsten
Durchgang** — und nach der Schleife weiter, bis der Aufruf endet, der sie
umschließt.

**Woran das sichtbar wurde:** `sub_type_filter` schält in einer Zählschleife
Wort für Wort von `input` ab (`input := replace(input, …)`). Mit einer eigenen
Ebene je Durchgang sah jeder wieder den ganzen Text — aus `Goblin Mage` wurde
`Goblin Goblin Goblin Goblin`.

**Das Gegenstück gilt auch:** der Rumpf einer gerufenen Funktion läuft in der
Umgebung des Aufrufers (dynamische Bindung), und was er zuweist, wird beim
Verlassen zurückgenommen.

**Quelle:** MSE2 `src/script/context.cpp`, Zeilen 43–44, 168–170 und 260–280 —
Fassung f9d9356 vom 2021-07-06
**Erschlossen am:** 2026-08-17
**Sicherheit:** belegt
**Testfall:** TestEineZuweisungInEinerSchleifeGiltImNaechstenDurchgang,
TestDerSchleifennameUeberlebtDenDurchgang

---

## `replace:` darf eine Funktion sein — und `x.0` ist ein Zeichen

**Zwei Regeln, die zusammen den Kartentext ausmachen.**

```cpp
if (replacement_function) {
  for (UInt sub = 0 ; sub < results.size() ; ++sub) {
    String name  = sub == 0 ? _("input") : String(_("_")) << sub;
    ctx.setVariable(name, to_script(results.str(sub)));
  }
  inside = replacement_function->eval(ctx)->toString();
}
```

**Je Treffer wird gerufen**, der ganze Treffer heißt `input`, die Gruppen `_1`
bis `_9`. Mit `recursive:` wird das Ergebnis erneut behandelt, höchstens zwanzig
Ebenen tief. Ist `replace:` **keine** Funktion, gilt die sed-Vorlage wie bisher.

```cpp
ScriptValueP getMember(const String& name) const override {
  // get member returns characters
  long index;
  if (name.ToLong(&index) && index >= 0 && (size_t)index < value.size()) {
    return to_script(String(1,value[index]));
  } else {
    return delay_error(…);
  }
}
```

**Eine Zeichenkette gibt ihre Zeichen her** — `split.0` ist das erste Wort einer
Zerlegung, `text.0` das erste Zeichen. MSE kennt Index und Namenszugriff nicht
getrennt.

**Was das gekostet hat, gemessen über die 190 Stile von `M15-Magic-Pack`:** 480
Befunde für den Zeichenzugriff; dazu die Typzeile und der Regeltext **jeder**
Karte — auf *Firebrand Incinerator* stand `Creature — Goblinfunktion<class-en>…`
und `deal Funktion damage`.

**Was hier abweicht:** MSE gibt bei einem Zeichen außerhalb der Zeichenkette
einen **verzögerten** Fehler — einen Wert, der erst beim Benutzen wirft. Hier
ist es `nichts`. Ebenso bei `to_int` auf einen nicht zahligen Text
(#345).

**Quelle:** MSE2 `src/script/functions/regex.cpp`, Zeilen 72–140;
`src/script/value.cpp`, Zeilen 140–167 und 376–384 — Fassung f9d9356 vom
2021-07-06
**Erschlossen am:** 2026-08-17
**Sicherheit:** belegt
**Testfall:** TestReplaceDarfEineFunktionSein,
TestRecursiveWendetDieErsetzungAufSichSelbstAn,
TestEineErsatzfunktionLaeuftAuchUeberDieRueckfalllinie

---

## `recolor_image` liest einen Bildpunkt als Mischung

```cpp
int lo = min(x.r,min(x.g,x.b));
int nr = x.r - lo; int ng = x.g - lo; int nb = x.b - lo; int nw = lo;
int total = max(255, nr+ng+nb+nw);
return RGB((nr*cr.r + ng*cg.r + nb*cb.r + nw*cw.r) / total, …);
```

**Der Abstand jedes Kanals zum dunkelsten ist der Anteil dieser Grundfarbe**,
der Rest ist Weiß. Ein Rahmen wird in MSE einmal in Rot, Grün, Blau und Weiß
gezeichnet und danach in vier beliebige Farben übersetzt.

**Mit nur einer Farbe entscheidet ihre Helligkeit:**

```cpp
Byte to_grayscale(RGB x) { return (6969*x.r + 23434*x.g + 2365*x.b) / 32768; }
bool dark = to_grayscale(cr) < 100;
recolor(img, cr, dark ? black : white, dark ? white : black, white);
```

Die Gewichte kommen aus **libpng**, nicht aus Rec. 601. **Geteilt wird durch die
Summe, nicht durch 255**, und **das Alpha bleibt** — MSE rechnet über drei
Kanäle.

**965 der Bildebenen** über `M15-Magic-Pack` hingen daran; sie wurden bis dahin
in ihrer Ausgangsfarbe gezeichnet
(#270).

**Quelle:** MSE2 `src/gfx/image_effects.cpp`, Zeilen 76–112;
`src/script/functions/image.cpp`, Zeilen 88–99 — Fassung f9d9356 vom 2021-07-06
**Erschlossen am:** 2026-08-17
**Sicherheit:** belegt
**Testfall:** TestRecolorImageLiestEinenPunktAlsMischung,
TestRecolorImageMitEinerFarbeEntscheidetDieHelligkeit,
TestRecolorImageLaesstDasAlphaStehen

---

## Der Vorgabewert läuft **nach** dem Initskript — und sieht `value`

```cpp
value = field().default_script.invokeOnDefault(ctx, value)
      | field().        script.invokeOn(ctx, value);
```

`TextValue::update`, `src/data/field/text.cpp`, Zeilen 178–179.

**Beides steht in `update`**, also lange nachdem der Skriptverwalter die
`init script:` von Spiel und Stil ausgeführt hat — und die legen die Funktionen
an, die ein Vorgabewert ruft: `separator_default`,
`additional_credit_default` und zwei Dutzend weitere.

**`value` ist der bisherige Wert**, und die Vorgabewerte des Packs rechnen
damit: `default: stylesheet; separator_default(value, margin_code: "sep1")`.

**Gemessen** über die 190 Stile von `M15-Magic-Pack`: ohne die Bindung von
`value` scheiterten **21.280** Vorgabewerte, ohne die Reihenfolge weitere
Hunderte. Beides war unsichtbar, weil ein gescheiterter Vorgabewert **lautlos**
zu leerem Text wurde — zu sehen war davon nur ein Dateiname wie
`…/separators/.png` im Rasterbefund
(#295).

**Quelle:** MSE2 `src/data/field/text.cpp`, Zeilen 178–179;
`src/script/scriptable.hpp`, Zeilen 57–69 — Fassung f9d9356 vom 2021-07-06
**Erschlossen am:** 2026-08-17
**Sicherheit:** belegt
**Testfall:** TestEinVorgabeskriptDarfDieFunktionenDesSpielsBenutzen,
TestEinGescheiterterVorgabewertWirdGemeldet

---

## Ein Maß eines Feldstils darf ein `script:`-Block sein

```
left:
	script:
		width := face_coordinates_map(1).width
		…
```

**Dieselbe Form wie bei `alignment:`**: der Wert des Knotens ist leer, und das
Skript steht in seinen Zeilen. `magic-modules.mse-include/stamps/card_fields`
schreibt alle vier Angaben so.

**Was das gekostet hat:** wer nur den Wert liest, bekommt nichts — und die
Ebene fällt auf die Maße der **ganzen Karte** zurück. Das 43 Punkte große
Prägesiegel wurde damit über die volle Karte gezeichnet, und mit ihm sechs
weitere Ebenen (#295).

**Quelle:** `magic-modules.mse-include/stamps/card_fields` aus `M15-Magic-Pack`
**Erschlossen am:** 2026-08-17
**Sicherheit:** belegt
**Testfall:** TestEinMassDarfEinSkriptblockSein, TestEinLeererBlockIstKeinMass

---

## Ein Stil bringt eigene Felder mit — `extra card field:`

```cpp
REFLECT(extra_card_fields);
REFLECT_IF_READING {
  if (extra_card_style.init(extra_card_fields)) { … }
}
REFLECT(extra_card_style);
```

**Daran hängt der Kasten für Stärke und Widerstandskraft.** `magic-m15`
zeichnet ihn nicht über ein Kartenfeld, sondern über ein Feld der Vorlage:

```
extra card field:
	type: choice
	name: pt box
	script: card.card_color
extra card style:
	pt box:
		left: 273 … visible: { card.pt != "" } … image: { module_ptbox() }
```

**Ihre Werte stehen als `extra_card`**, ihre Maße als `extra_card_style` —
`magic.mse-game` liest `extra_card.vorthos_box or else ""`, und `or else` fängt
ein fehlendes **Glied**, keinen fehlenden **Namen**: die Abbildung muss also
dastehen, auch wenn sie leer ist.

**Gezeichnet wird über beide Listen in einer Reihenfolge:**

```cpp
setStyles(stylesheet, stylesheet->card_style, &stylesheet->extra_card_style);
```

Es entscheidet allein `z index`. Wer sie hintereinander hängte, legte den
Kasten über den Rahmen statt darunter.

**Gemessen** über die 190 Stile von `M15-Magic-Pack`: Bildebenen 1.522 →
**2.281**, dazu 232 Befunde „der Name `extra_card` ist hier nicht bekannt" auf
null (#270).

**Quelle:** MSE2 `src/data/stylesheet.cpp`, Zeilen 116–125;
`src/render/card/viewer.cpp`, Zeile 138 — Fassung f9d9356 vom 2021-07-06; dazu
`magic-m15.mse-style/style`, Zeilen 328–360
**Erschlossen am:** 2026-08-17
**Sicherheit:** belegt
**Testfall:** TestEinStilDarfEigeneFelderMitbringen,
TestDieZusatzfelderEinesStilsZeichnenMit,
TestOhneZusatzfelderBleibtExtraCardLeer

---

## Die Werte der Zusatzfelder stehen je Stilvorlage unter `extra_data:`

```
card:
	stylesheet: nouveau
	extra_data:
		magic-nouveau:
			keyword_icons: flying, haste
```

**Zwei Ebenen, und die zweite ist der Punkt.** Unter `extra_data:` steht je
Stilvorlage ein Block, benannt nach dem **Paketnamen ohne Endung**
(`Package::name`), und darunter die Werte ihrer Zusatzfelder. MSE2 liest einen
Block erst, wenn jemand die Werte **dieser** Vorlage verlangt
(`DelayedIndexMaps::get`), und schreibt die übrigen ungelesen zurück. Die
Schlüssel liest `Reader` in kanonischer Form, Leerzeichen als Unterstrich.

**Ein mehrzeiliger Wert behält seine tiefere Einrückung.** `Reader::getValue`
schneidet nur die Einrückung des Blocks ab, nicht die Stufen darunter. Wer
jede Einrückung abschneidet — wie es für ein Skript richtig ist —, setzt
`keyword_icons` neben `magic-nouveau` statt darunter, und MSE2 liest den Wert
der Vorlage als leer (#2298).

**Gerechnet wird wie bei einem Kartenfeld:** der gespeicherte Wert vor der
Vorgabe, das `script:` darüber, eine Mehrfachwahl danach in Normalform
(`SetScriptManager::updateStyles`, `ChoiceValue::update`,
`MultipleChoiceValue::update`).

**Gezählt am 2026-10-11** über `Full-Magic-Pack`: 3.037 Zusatzfelder, davon
282 in 81 Stilen ohne `script:` und ohne `editable: false` — Werte, die nur
eine Karte selbst tragen kann (`vorthos box`, `promo label`,
`keyword icons`, `paintbrush`).

**Quelle:** MSE2 `src/data/card.cpp`, Zeilen 63–65 und 94;
`src/util/delayed_index_maps.hpp`, Zeilen 19–33 und 53–71;
`src/util/io/reader.cpp`, Zeilen 208–217 und 255–298;
`src/util/io/package.cpp`, Zeilen 43–53; `src/script/script_manager.cpp`,
Zeilen 208–222; `src/data/field/choice.cpp`, Zeilen 290–295 — Fassung f9d9356
vom 2021-07-06; dazu `magic-nouveau.mse-style/style`
**Erschlossen am:** 2026-10-11
**Sicherheit:** belegt — das Prüfset `Reihenprobe` zeichnet MSE2 aus der hier
geschriebenen Datei bildpunktgleich zum Original
**Testfall:** TestEinBlockBehaeltSeineTiefereEinrueckung,
TestZusatzdatenKommenMitIhrerTiefeInDenBestand,
TestEinZusatzfeldLiestDenGespeichertenWert,
TestEinFlacherAltwertLaesstDieZusatzdatenDerQuelleStehen

---

## Ein leeres Feld meldet einen leeren Satz

`card_style.<feld>.content_width`, `content_height` und `content_lines` sind
bei MSE2 Rückmeldungen aus dem gesetzten Text — **auch aus einem leeren**:

| Feld | `content_width` | `content_height` | `content_lines` |
| --- | --- | --- | --- |
| Text, leer | 0 | die Höhe eines Zeichens der Schrift | 1 |
| Symbolzug (`always symbol`), leer | 0 | die Höhe des Ersatzsymbols samt Abstand | 1 |
| Wahl ohne Bildanteil (`render style: text`, die Vorgabe) | 0 | 0 | — |
| Wahl mit Bild, leer | der Kasten (`image`) oder 1 (`both`) | ebenso | — |

**Eine Wahl meldet das Maß ihres vorbereiteten Bilds**, gefüllt wie leer
(`prepare_choice_viewer`, `src/render/value/choice.cpp`, Zeilen 35–58): bei
`render style: image` das Bild in den Kasten in ganzen Punkten eingepasst
(`ASPECT_FIT`, außer die Ausrichtung sagt `stretch`), bei `both` seine
Originalgröße. **Der leere Wert ist kein fehlendes Bild**: die
Nachschlagetafel der Wahlbilder gibt für ihn `nil`, und `nil` wie `""` werden
zu `BlankImage` in der gefragten Größe (`src/script/value.cpp`, Zeilen
369–375 und 450–452). Ein `image:` ohne Skript wird nie neu gerechnet und
meldet nichts; eine Mehrfachwahl als Liste ebenso. Ein Setsymbol wird auf
sein Seitenverhältnis beschnitten (`render_symbol`, `allow_smaller`), das
Standardsymbol eines Sets ohne eigenes ist ein Quadrat — `magic-m15` meldet
dafür 22 × 22. **Gefragt am 2026-10-11 über einen Orakelstil**: 30 × 10 im
Kasten 44 × 22 gibt 44 × 14, 10 × 30 gibt 7 × 22, ein breites Symbol 44 × 22,
ein quadratisches 22 × 22, ein Wert ohne Wahlbild und der leere 44 × 22; mit
`both` 30 × 10, 10 × 30 und 1 × 1. **Hier stand bis dahin** für eine leere
Wahl null und für eine gefüllte der Kasten oder — mit Schrift — die Breite
ihres Texts (#2303).

Das Ersatzsymbol ist das erste Symbol mit Muster, das „0" trifft; gibt es
keines, ein Bildpunkt samt Abstand (`SymbolFont::defaultSymbol`,
`defaultSymbolSize`).

**Gefragt am 2026-10-11 mit einem Orakelstil**, der die drei Maße des Felds
`illustrator` als Lage auf die Karte schreibt: leer gibt Breite 0, eine Zeile
und dieselbe Höhe wie „Probe Zeichner". Hier stand für ein leeres Feld die
deklarierte Breite; `magic-nouveau` setzt das Pinselzeichen auf
`150 + card_style.illustrator.content_width * -0.5` und zeichnete es so an
den Rand. **Gelesen werden die Maße oft**: `casting_cost.content_width`
412-mal im Pack, `text.content_lines` 346-mal — und Manakosten eines Lands
und Regeltext einer Vanillakreatur sind leer.

**Quelle:** MSE2 `src/render/text/viewer.cpp`, Zeilen 350–353 und 392–426;
`src/data/field/text.cpp`, Zeilen 114–121; `src/data/symbol_font.cpp`, Zeilen
263–268 und 403–410; `src/data/field/choice.cpp`, Zeile 168 — Fassung
f9d9356 vom 2021-07-06
**Erschlossen am:** 2026-10-11
**Sicherheit:** belegt (Orakel)
**Testfall:** TestEinLeeresFeldMeldetSeinenLeerenSatz,
TestEinLeererSymbolzugMisstSeinErsatzsymbol, TestEinGanzLeererBlockIstEineZeileHoch

---

## `draw text:` schreibt eine Mustergruppe auf das Symbolbild

```
code: ([XYZ0-9])/W
regex: yes
draw text: 1
text margin left: -.25
text font:
	name: MPlantin
	size: .6
	max stretch: 0.5
```

**Die Zahl ist die Gruppennummer**, nicht ein Schalter:

```cpp
int draw_text; ///< The index of the captured regex expression to draw,
               ///  or -1 to not draw text
```

Ohne Muster gibt es nur eine Gruppe, und das ist der Code selbst
(`sym->draw_text >= 0 ? sym->code : ""`). Eine Nummer außerhalb der Gruppen
zeichnet **nichts** — MSE prüft sie gegen `results.size()`.

**Gezeichnet wird in den Kasten des Bildes, nicht des Symbols:**

```cpp
sym_rect = RealRect(bmp_pos, bmp_size);
sym_rect.x      += font_size * sym.text_margin_left;
sym_rect.width  -= font_size * (sym.text_margin_left + sym.text_margin_right);
double size = font_size * sym.text_font->size;
```

**Die vier Ränder sind Anteile der Symbolgröße und dürfen negativ sein**;
`size` ist ebenfalls ein Anteil. Passt der Text nicht, wird der Grad in
Schritten kleiner — **`max stretch` ist dabei eine Schwelle**: erst wenn der
Text auch gestaucht nicht hineinpasst, sinkt der Grad. Die Stauchung selbst ist
dann `Kastenbreite / Textbreite`.

**Die Ausrichtung hat eine eigene Vorgabe**: `middle center`, während ein
**Feld** links oben ausrichtet. Wer beides gleichsetzt, schiebt jede Ziffer in
die linke obere Ecke ihres Kreises.

**Daran hängt jede Zahl in den Manakosten** — ohne sie steht auf der Karte der
leere Kreis (#336).

**Quelle:** MSE2 `src/data/symbol_font.cpp`, Zeilen 84, 100–113, 165–230 und
287–329 — Fassung f9d9356 vom 2021-07-06; dazu
`magic-mana-large.mse-symbol-font/symbol-font`, Zeilen 131–146
**Erschlossen am:** 2026-08-17
**Sicherheit:** belegt
**Testfall:** TestEinSymbolDarfTextAufSeinBildZeichnen,
TestEinSymbolSchreibtSeineGruppe, TestEinSymbolSchreibtTextAufSeinBild,
TestEinTextDerNichtPasstWirdKleiner

---

## Vier Namen auf dem Bildweg: `set_alpha`, `set_combine`, `type_name`, `reverse`

**`set_alpha` setzt die Durchsichtigkeit:**

```cpp
void set_alpha(Image& img, double alpha) {
  Byte b_alpha = Byte(alpha * 255);
  if (!img.HasAlpha()) { img.InitAlpha(); memset(img.GetAlpha(), b_alpha, …); }
  else { for (…) im[i] = (im[i] * b_alpha) / 255; }
}
```

**Ein Bild ohne Alpha bekommt den Wert gesetzt, eines mit Alpha ihn
multipliziert** — die vorhandene Form bleibt also erhalten. Der Wert wird
**abgeschnitten**, nicht gerundet.

**`set_combine` ändert am Bild nichts** (`return image->generate(opt)`) — es
setzt die **Mischart der Ebene**, und die trägt der Plan an der Ebene.

**`type_name` gibt MSEs englische Typnamen**: `integer`, `double`, `boolean`,
`nil`, `function`, `collection`, `color` — und für eine Zeichenkette
`string ("…")` mit ihrem Inhalt, nach dreißig Zeichen gekürzt. Die Vorlagen
vergleichen sie (`type_name(x) != type_name(nil)`).

**`reverse` dreht eine Zeichenkette um**, in Zeichen und nicht in Bytes.

**`exists_in_package` steht in keiner der beiden portierbaren Quellen** — weder
MSE2 f9d9356 noch `fenhl/magic-set-generator` kennen den Namen. Das Pack
benutzt ihn (`magic.mse-game/script`, Zeilen 2594, 3614, 3645), und was er tut,
ist an seiner Verwendung **erschlossen**: er fragt, ob eine Datei in einem Paket
liegt, und die Vorlage wählt danach zwischen einem besonderen Bild und einem
allgemeinen.

**Gemessen** über die 190 Stile von `M15-Magic-Pack`: Bildebenen 2.281 →
**2.900**, Befunde 3.219 → 2.836
(#270).

**Quelle:** MSE2 `src/gfx/blend_image.cpp`, Zeilen 107–119;
`src/gfx/generated_image.cpp`, Zeilen 162–186;
`src/script/functions/basic.cpp`, Zeilen 231–234 und 333–336;
`src/script/value.cpp`, Zeilen 223–324 und 402–444 — Fassung f9d9356 vom
2021-07-06. Für `exists_in_package`: `magic.mse-game/script` aus
`M15-Magic-Pack`
**Erschlossen am:** 2026-08-17
**Sicherheit:** belegt für die vier ersten, **erschlossen** für `exists_in_package`
**Testfall:** TestSetAlphaSetztDieDurchsichtigkeit,
TestSetCombineReichtDasBildDurch, TestTypeNameNenntDieArtEinesWerts,
TestReverseDrehtEineZeichenketteUm, TestExistsInPackageAntwortetUeberDiePakete

---

## `<sep>` kennt MSE nicht — `<soft>` zählt bei der Ausrichtung nicht mit

**Die Liste der gedeuteten Auszeichnungen** steht in `TextElements::fromString`:
`b`, `i`, `sym`, `line`, `soft-line`, `sep-soft`, `soft`, `li`, `atom-kwpph`,
`code…`, `color`, `font`, `size`, `margin`, `align`. **`sep` steht nicht
darin.** Eine unbekannte Auszeichnung wird übersprungen, ihr Inhalt bleibt
Text — `combined_editor` klammert damit den Trenner zwischen Regel- und
Flavortext: `<sep><line>\n</line></sep>`.

**Die Reihenfolge der Prüfungen zählt**, denn `is_tag` arbeitet auf einem
Präfix: `<soft-line` steht **vor** `<sep-soft`, und das vor `<soft`. Wer sie
umdrehte, machte aus jedem weichen Umbruch eine weiche Stelle.

**Weich heißt: bei der Ausrichtung übergangen.**

```cpp
bool soft : 1; ///< Is this a 'soft' character? soft characters are ignored for alignment
```

```cpp
size_t end_or_soft; ///< Index just beyond the last non-soft character
double width() const { return positions[end_or_soft-start] - positions.front(); }
```

**Nur der Schwanz fällt weg.** MSE merkt sich die Stelle hinter dem letzten
nicht weichen Zeichen (`if (!c.soft) word_end_or_soft = i + 1`); ein weiches
Stück **zwischen** zwei harten zählt mit.

**Wozu:** `sub_type_filter` hängt an die Typzeile ein
`"<soft>" + subtype_separator + "</soft>"` — einen Trenner für den Fall, dass
noch etwas folgt. Zählte er mit, stünde jede mittig gesetzte Typzeile um seine
halbe Breite zu weit links.

**Was hier abweicht:** MSE zeichnet weichen Text in der `separator color` und
ohne Schatten (`src/data/font.cpp`, Zeilen 66–69). Das ist hier **nicht**
gebaut; gezeichnet wird er in der Farbe des Feldes. Sichtbar wird der
Unterschied nur dort, wo eine Vorlage `separator color:` setzt und der weiche
Text nicht am Zeilenende steht.

**Gemessen** über die 190 Stile von `M15-Magic-Pack`: 367 Befunde
„die Auszeichnungen sep sind entfernt" auf null, Befunde 2.836 → **2.469**
(#345).

**Quelle:** MSE2 `src/render/text/element.cpp`, Zeilen 78–130 und 294–300;
`src/render/text/element.hpp`, Zeilen 36–47; `src/render/text/viewer.cpp`,
Zeilen 17–35 und 578–650 — Fassung f9d9356 vom 2021-07-06
**Erschlossen am:** 2026-08-17
**Sicherheit:** belegt
**Testfall:** TestSepIstKeinBefund, TestWeicherTextWirdAlsWeichGefuehrt,
TestSoftLineIstKeinWeicherText, TestSepSoftIstWeich,
TestEinWeicherTrennerZaehltNichtMit

---

## Ein Zugriff ins Leere ist ein **Wert**, und `or else` fängt ihn

```cpp
ScriptValueP ScriptValue::getMember(const String& name) const {
  return delay_error(ScriptErrorNoMember(typeName(), name));
}
```

```cpp
case I_OR_ELSE:
  if (at == SCRIPT_ERROR) a = b;
  break;
```

**Das ist die zentrale Regel dieser Sprache, und sie war falsch verstanden.**
Ein fehlender Zugriff bricht bei MSE **nicht** ab: er gibt einen
`ScriptDelayedError` zurück. Jede Umwandlung dieses Werts wirft, aber ein
weiterer Zugriff reicht ihn durch — und `or else` **ersetzt** ihn.

**`or else` fängt genau diesen Wert, nicht `nil`.** `nil or else 0` ergibt
`nil`; hier stand bis zum 2026-08-17 das Gegenteil, und damit fing `or else`
den falschen Fall.

**Wo überall ein Fehlerwert entsteht:**

| Ausdruck | vorher | jetzt |
| --- | --- | --- |
| `5.feld` — Glied an etwas ohne Glieder | Abbruch | Fehlerwert |
| `[a: 1].b` — Glied, das es nicht gibt | `nil` | Fehlerwert |
| `[a: 1]["b"]` — Schlüssel, den es nicht gibt | `nil` | Fehlerwert |
| `"wu"[9]` — über das Ende hinaus | `nil` | Fehlerwert |
| `to_int("kein Zahlwort")` | Abbruch | Fehlerwert |
| `to_int("")` | Abbruch | **null** (Zeile 121) |
| `to_int(true)` | Abbruch | **eins** (Zeile 111) |
| `[a: 1].b(1)` — einen Fehlerwert rufen | Abbruch | Fehlerwert (`value.cpp`, Zeile 163) |

**`to_int(true)` ist eins, `to_int("true")` nicht.** Die Umwandlung steht in
der Funktion und nicht am Typ: `ScriptBool` überschreibt `toInt` nicht, und der
Kommentar dort lautet „bools don't autoconvert to int". Eine Zeichenkette geht
in den Zeichenkettenzweig, und `str.ToLong` scheitert dort.

**Wozu die Vorlagen das brauchen:**

```
uninitialized := (style.layout.blocks[1].bottom or else 0) <= 0
map["stylesheet"] or else "m15-altered"
if value == "" then extra_card.vorthos_box or else "" else value
```

**Gemessen** über die 190 Stile von `M15-Magic-Pack`: Befunde 2.469 → **1.407**.
Allein die Klasse „`bottom` gibt es nur an einer Abbildung" fiel von 406 auf
null, `to_int` von 607 auf einen Rest
(#345).

**Quelle:** MSE2 `src/script/value.cpp`, Zeilen 67–74 und 140–167;
`src/script/context.cpp`, Zeilen 535–537; `src/script/functions/basic.cpp`,
Zeilen 106–132 — Fassung f9d9356 vom 2021-07-06
**Erschlossen am:** 2026-08-17
**Sicherheit:** belegt
**Testfall:** TestEinFehlgeschlagenerZugriffIstEinWert,
TestWerDenFehlerwertBenutztBekommtDenFehler,
TestToIntAufUnzahligemTextGibtEinenFehlerwert,
TestDerFehlerwertWirftAuchAlsZahl,
TestToIntMachtAusEinemWahrheitswertEinsOderNull,
TestEinGerufenerFehlerwertBleibtEinFehlerwert

---

## `visible:` entscheidet, ob eine Ebene gezeichnet wird

```cpp
bool ValueViewer::isVisible() const {
  return getStyle()->visible
    && bounding_box.width > 0 && bounding_box.height > 0
    && fabs(bounding_box.x) < 100000 && fabs(bounding_box.y) < 100000;
}
```

Geprüft **vor** dem Zeichnen (`CardViewer::draw`, Zeile 62), nicht danach. Der
Nullwert ist **sichtbar** (`visible(true)`, `src/data/field.cpp`, Zeile 106),
und die Angabe darf einzeilig oder als `script:`-Block dastehen.

**Was das kostet, wenn es fehlt:** `magic-m15` schreibt am Kasten für Stärke und
Widerstandskraft `visible: { card.pt != "" }`. Ohne die Angabe steht auf **jeder
Verzauberung** ein leerer weißer Kasten unten rechts.

**Gemessen** über die 190 Stile von `M15-Magic-Pack`: Bildebenen 2.892 →
**1.765**, Textebenen 2.541 → 2.352, Befunde 1.215 → **946**. Die Bildebenen
**fallen**, und das ist der Sinn der Angabe: 1.127 Ebenen, die eine Vorlage
ausdrücklich verbirgt, wurden bis dahin gezeichnet.

**Was hier abweicht:** MSE prüft zusätzlich die Maße des Kastens (Breite und
Höhe über null, Stelle unter 100.000). Das steht hier **nicht** — eine Ebene
ohne Größe zeichnet ohnehin nichts, und die Grenze von 100.000 ist dieselbe
Zahl, mit der `Style::update` „nicht gesetzt" meint
(#270).

**Quelle:** MSE2 `src/render/value/viewer.cpp`, Zeilen 36–42;
`src/render/card/viewer.cpp`, Zeile 62; `src/data/field.cpp`, Zeile 106 —
Fassung f9d9356 vom 2021-07-06
**Erschlossen am:** 2026-08-17
**Sicherheit:** belegt
**Testfall:** TestVisibleEntscheidetObEineEbeneGezeichnetWird,
TestVisibleWirdGelesen

---

## `combined_editor` kennt drei Fälle für seinen Trenner

```cpp
if (value_parts[i].second && new_value_empty && hide_when_empty) {
  // no separator
} else if (value_parts[i].second && soft_before_empty) {
  new_value += _("<sep-soft>") + separators[i - 1] + _("</sep-soft>");
  new_value_empty = false;
} else {
  new_value += _("<sep>") + separators[i - 1] + _("</sep>");
  new_value_empty = false;
}
```

**Ganz weg nur, wenn auch alles davor leer ist**; sonst **weich**; sonst
gewöhnlich. `magic.mse-game` baut die Typzeile mit beiden Angaben — bei einer
Zauberei ohne Untertyp steht der Gedankenstrich damit weich da, und weicher Text
zählt bei der Ausrichtung nicht mit.

**Quelle:** MSE2 `src/script/functions/editor.cpp`, Zeilen 96–116 — Fassung
f9d9356 vom 2021-07-06
**Erschlossen am:** 2026-08-17
**Sicherheit:** belegt
**Testfall:** TestSoftBeforeEmptyMachtDenTrennerWeich

---

## Drei kleine Namen: `exclusive_choice`, `forward_editor`, `[[.name.]]`

**`exclusive_choice` gegen `require_exclusive_choice`** — der Unterschied sind
die Grenzen:

```cpp
SCRIPT_RETURN(filter_choices(input, choices, 0, 1, last_change)); // exclusive
SCRIPT_RETURN(filter_choices(input, choices, 1, 1, last_change)); // require_exclusive
```

Es darf **keine** Wahl übrig bleiben, aber nicht zwei.

**`forward_editor` ist derselbe Name für `combined_editor`:**

```cpp
ctx.setVariable(_("forward_editor"), script_combined_editor); // compatability
```

**Ein POSIX-Kollationselement benennt ein Zeichen.** MSE übersetzt seine Muster
mit `wxRegEx` im erweiterten POSIX-Modus, und dort steht `[[.name.]]` für das
Zeichen mit diesem Namen. Im Pack kommt genau einer vor:

```
replace(match: "[[.quotation-mark.]]|”", replace: …)
```

**Was hier abweicht:** übersetzt wird nur, was im Pack ausgezählt ist. Ein
unbekannter Name bleibt stehen, wird von der Rückfalllinie angenommen und
trifft dann **nichts**. Das ist der harmlosere Fehlschlag — eine Ersetzung
unterbleibt, statt an der falschen Stelle zuzuschlagen —, aber er ist nicht
sichtbar, und deshalb steht er hier.

**Gemessen** über die 190 Stile von `M15-Magic-Pack`: Befunde 946 → **796**
(#270).

**Quelle:** MSE2 `src/script/functions/editor.cpp`, Zeilen 340–346 und 390 —
Fassung f9d9356 vom 2021-07-06; dazu `magic.mse-game/script` aus
`M15-Magic-Pack`
**Erschlossen am:** 2026-08-17
**Sicherheit:** belegt für die beiden Funktionen, **erschlossen** für die
Kollationstabelle
**Testfall:** TestExclusiveChoiceUndForwardEditor,
TestEinKollationselementBenenntEinZeichen,
TestEinUnbekanntesKollationselementTrifftNichts

---

## `split_text` behält leere Teile — `include_empty` ist wahr

```cpp
SCRIPT_PARAM_DEFAULT(bool, include_empty, true);
while (match->matches(results, start, input.end())) {
  if (include_empty || pos.first != start) {
    ret->value.push_back(to_script(String(start, pos.first)));
  }
  start = pos.second;
}
if (include_empty || start != input.end()) {
  ret->value.push_back(to_script(String(start, input.end())));
}
```

**Der leere Text ergibt damit ein Element, nicht keines.** Hier stand bis zum
2026-08-17 das Gegenteil, und die Begründung war **erschlossen**: „eine leere
Regelbox hätte danach eine Zeile". Sie war falsch — fürs Zeilenzählen schreibt
`magic.mse-game` ausdrücklich `include_empty: false`:

```
line_count := split_text@(match:"\n+", include_empty:false) + length
```

**Was das gekostet hat:** `pull_comma_array` liest `array[cell]` aus einer
Zerlegung. War die leer, wo MSE ein Element hat, gab der Zugriff einen
Fehlerwert — **280 Befunde** über die 190 Stile von `M15-Magic-Pack`.

**Quelle:** MSE2 `src/script/functions/regex.cpp`, Zeilen 212–232 — Fassung
f9d9356 vom 2021-07-06; dazu `magic.mse-game/script`, Zeile 47
**Erschlossen am:** 2026-08-17
**Sicherheit:** belegt
**Testfall:** TestSplitTextBehaeltLeereTeile

---

## Eine Karte trägt Notizen, ein Paket seinen Namen

```cpp
class Card … {
  IndexMap<FieldP, ValueP> data;
  /// Notes for this card
  String notes;
};
```

**`notes` ist kein Kartenfeld des Spiels**, sondern eine Eigenschaft jeder
Karte — es steht in keiner `card field:`-Liste. `magic.mse-game` liest es an
drei Stellen, darunter die Kartennummer.

```cpp
String short_name;      ///< Short name of this package
…
if (short_name.empty()) {
  if (!full_name.empty()) short_name = full_name;
  short_name = name();
}
```

**Die Vorlagen lesen `stylesheet.short_name`** — einmal je Stil. Die
Rückfallkette ist Teil der Regel: `short name:`, sonst `full name:`, sonst der
Dateiname.

**Was hier abweicht:** MSEs Kette hat kein `else` — der Dateiname überschreibt
den vollen Namen wieder. Das sieht wie ein Versehen aus; hier steht die Kette
**mit** `else`, sonst wäre `full name:` nie zu sehen.

**Und ein Fehlerwert ist kein Feldwert.** Ein Wertskript, das einen verzögerten
Fehler ergibt, wird gemeldet; der alte Wert bleibt. Ohne diese Prüfung landete
der Wortlaut `Fehler` — der Name der Wertart — als Kartenwert im Plan, 28-mal
über die 190 Stile
(#270).

**Quelle:** MSE2 `src/data/card.hpp`, Zeilen 36–40; `src/data/card.cpp`, Zeile
91; `src/util/io/package.hpp`, Zeile 266; `src/util/io/package.cpp`, Zeilen
539 und 623–625; `src/script/value.cpp`, Zeile 148 — Fassung f9d9356 vom
2021-07-06
**Erschlossen am:** 2026-08-17
**Sicherheit:** belegt
**Testfall:** TestEineKarteTraegtNotizen, TestEinPaketNenntSeinenNamen,
TestEinWertskriptMitFehlerwertWirdGemeldet

---

## `symbol_variation` rastert das Setsymbol in einer benannten Fassung

```cpp
FOR_EACH(v, style->variations) {
  if (v->name == variation) {
    return make_intrusive<SymbolToImage>(value, filename, value->last_update, v);
  }
}
throw ScriptError(_("Variation of symbol not found ('") + variation + _("')"));
```

**Die Fassungen stehen an zwei Orten**, und beide zählen: das Spiel erklärt die
Grundfassungen unter **`default set style:`** → `symbol:` → `variation:`
(`common`, `uncommon`, `rare`, `mythic rare`), die Vorlage ergänzt sie unter
`set info style:` (`magic-m15` legt `invertedcommon` dazu).

**Und bei gleichem Namen gewinnt die des Spiels, nicht die der Vorlage.** Hier
stand bis zum 2026-08-19 das Gegenteil. Der Grund steht im Altcode: ein
Stilpaket **erbt** die Fassungen des Spiels — `set_info_style.cloneFrom(game->default_set_style)`
steht **vor** `REFLECT(set_info_style)` (`src/data/stylesheet.cpp`, Zeilen
106–113) —, und `Reader::handle(name, vector<T>&)` hängt die eigenen
**dahinter** an, ohne zu leeren (`src/util/io/reader.hpp`, Zeilen 218–226). Die
Suche gibt beim **ersten** Treffer zurück (`image.cpp`, Zeilen 164–170).

**Am Artefakt belegt**: die elf `special`-Karten von `White` erscheinen in
MSE2s eigener Ausgabe weiß — die Randfarbe der **Spiel**-Fassung
`rgb(255,255,255)`; die der Vorlage hätte `rgb(0,0,0)` ergeben
(#548).

**Nicht am `set field:`.** Das Feld erklärt, *was* ein Setsymbol ist; wie es
aussieht, steht im Stil. Wer die Fassungen am Feld suchte, fände keine — und
`symbol_variation` kennte dann `common` nicht, die jede Karte braucht.

**Aber sie gehören trotzdem je einem Feld.** Der Stilblock beschreibt **jedes**
Setfeld, und `magic.mse-game` erklärt zwei mit Fassungen: `symbol:` (sieben)
und `masterpiece symbol:` (sechs). Sie unterscheiden sich an genau einer
Stelle — `watermark` ist unter `symbol:` `fill color: rgb(0,0,0)`, unter
`masterpiece symbol:` `rgb(255,255,255)`.

MSE sucht deshalb am Stil des **Feldes, aus dem der Wert stammt**:

```cpp
SymbolStyleP style = dynamic_pointer_cast<SymbolStyle>(
    set->stylesheetForP(card)->styleFor(value->fieldP));
```

`src/script/functions/image.cpp`, Zeile 161. **`value->fieldP`** — ein Feldwert
ist bei MSE ein Objekt, das sein Feld kennt. Wer die Fassungen aller Felder in
einen Topf wirft, lässt die Reihenfolge in der Datei entscheiden.

**Eine Fassung trägt `border radius:` (Vorgabe 0,05) und einen Filter:**
`solid`, `linear gradient` oder `radial gradient`, jeweils mit ihren Farben.

**Die Gestalt reist im Plan.** Wer die `.mse-symbol` liest, ist
`internal/service`; `internal/render` bekommt die Formen, wie der Rasterer sie
braucht (Vermerk 0022). Beide Pakete kennen einander nicht (§2), und die
Übersetzung steht an **einer** Stelle.

**Was hier abweicht:** gerastert wird in 128 × 128 Punkten. MSE rastert in der
Größe, die die Ebene braucht (`render_symbol(…, width, height)`); der Bildbaum
entsteht hier, **bevor** ein Rahmen feststeht. Auf einer Karte fällt das nicht
auf — `magic-m15` gibt dem Symbol 21 × 21 —, auf einem Druckbogen bei 600 dpi
schon.

**Gemessen** über die 190 Stile von `M15-Magic-Pack`: Bildebenen 1.771 →
**1.943**, Befunde 499 → **330**
(#323).

**Quelle:** MSE2 `src/script/functions/image.cpp`, Zeilen 141–195;
`src/data/field/symbol.cpp`, Zeilen 31–46 — Fassung f9d9356 vom 2021-07-06;
dazu `magic.mse-game/set_fields`, Zeilen 535–571
**Erschlossen am:** 2026-08-17
**Sicherheit:** belegt
**Testfall:** TestEinSetsymbolHatBenannteFassungen,
TestDasSpielErklaertDieGrundfassungen,
TestSymbolVariationRastertDasSetsymbol,
TestSymbolVariationRastertDieZeichnung,
TestDieGestaltDerSymboldateiReistMit, TestEineSymbolfassungGehoertZuIhremFeld

---

## Fünf Verknüpfungen, die aus einem Bild ein anderes machen

**`saturate` rechnet über den Mittelwert der drei Kanäle:**

```cpp
int factor = int(256 * amount);
if (factor == 0) return;
else if (factor == 256) { pix[0] = r+r > g+b ? 255 : 0; … }
else if (factor > 0) { int div = 768 - 3*factor;
                       pix[0] = col((768*r - factor*(r+g+b)) / div); … }
else { int f1 = -factor, f2 = 768 - 3*f1;
       pix[0] = (f2*r + f1*(r+g+b)) / 768; … }
```

**Drei Zweige und eine Division.** Bei genau eins teilte MSE durch null; dort
gilt „größer als die beiden anderen zusammen", und der Kommentar sagt warum:
„if we take infty to be 255, then it is a >avg test".

**`invert_image` dreht jeden Kanal um** — `n` ist dreimal die Zahl der
Bildpunkte, die Durchsichtigkeit bleibt.

**`crop` schneidet ein Stück heraus:**

```cpp
return image->generate(opt).Size(wxSize((int)width, (int)height),
                                 wxPoint(-(int)offset_x, -(int)offset_y));
```

**Der Versatz zählt negativ**: `offset_x: 2` zeigt ab Spalte zwei. Was über den
Rand ragt, bleibt leer.

**`flip_horizontal` und `flip_vertical`** spiegeln eine Achse (`Mirror(true)`
bzw. `false`), und **`to_image`** gibt sein Bild unverändert zurück — MSE
benutzt es, um einen Pfad in ein Bild zu verwandeln.

**Gemessen** über die 190 Stile von `M15-Magic-Pack`: Bildebenen 1.943 →
**1.967**, Befunde 330 → **297**
(#270).

**Quelle:** MSE2 `src/gfx/image_effects.cpp`, Zeilen 15–72;
`src/gfx/generated_image.cpp`, Zeilen 254–311;
`src/script/functions/image.cpp`, Zeilen 26–130 — Fassung f9d9356 vom
2021-07-06
**Erschlossen am:** 2026-08-17
**Sicherheit:** belegt
**Testfall:** TestSaturateRechnetUeberDenMittelwert,
TestInvertImageDrehtJedenKanalUm, TestCropSchneidetEinStueckHeraus,
TestDieSpiegelungenTauschenEineAchse

---

## `combine_blend` kennt 23 Mischarten — und `set_mask` liest den roten Kanal

```cpp
COMBINE_FUN(COMBINE_ADD,         top(a + b))
COMBINE_FUN(COMBINE_STAMP,       col(a - 2 * b + 256))
COMBINE_FUN(COMBINE_NEGATION,    255 - abs(255 - a - b))
COMBINE_FUN(COMBINE_MULTIPLY,    (a * b) / 255)
COMBINE_FUN(COMBINE_SCREEN,      255 - (((255 - a) * (255 - b)) / 255))
COMBINE_FUN(COMBINE_OVERLAY,     a < 128 ? (a * b) >> 7
                                         : 255 - (((255-a) * (255-b)) >> 7))
COMBINE_FUN(COMBINE_HARD_LIGHT,  b < 128 ? … )
COMBINE_FUN(COMBINE_SOFT_LIGHT,  b)
COMBINE_FUN(COMBINE_SHADOW,      (b * a * a) / (255 * 255))
COMBINE_FUN(COMBINE_SYMMETRIC_OVERLAY,
            (Combine<COMBINE_OVERLAY>::f(a,b) + Combine<COMBINE_OVERLAY>::f(b,a)) / 2)
```

MSEs eigene Quelle steht daneben: „Based on
http://www.pegtop.net/delphi/articles/blendmodes/".

**Drei Eigenheiten, die man nicht erraten würde:**

- **`soft light` gibt `b` zurück** — dieselbe Formel wie `normal`. Das ist MSEs
  Zeile, keine Vereinfachung.
- **`overlay` und `hard light` unterscheiden sich nur darin, wer entscheidet**:
  `overlay` fragt `a`, `hard light` fragt `b`.
- **Sechs Arten teilen und prüfen ihren Nenner einzeln** — `color dodge` gibt
  255, `color burn` null. Ein gemeinsamer Sonderfall wäre bei der Hälfte falsch.

**Das Alpha kommt vom zweiten Bild:**

```cpp
if (b.HasAlpha()) {
  if (!a.HasAlpha()) a.InitAlpha();
  memcpy(a.GetAlpha(), b.GetAlpha(), a.GetWidth() * a.GetHeight());
}
```

Die Form des Ergebnisses ist damit die des **oberen** Bildes.

**`set_mask` nimmt das Alpha aus dem roten Kanal:**

```cpp
im[i] = (im[i] * al[i*3]) / 255;
```

`al[i*3]` ist Rot, nicht die Helligkeit — derselbe Unterschied wie bei
`masked_blend`, und dort hat er schon einmal gekostet. **Die Maske wird
skaliert** (`resample`), `combine_blend` dagegen **behauptet** die gleiche Größe
(`assert`). Hier wird sie geprüft: eine Behauptung über fremde Daten ist keine
(§8).

**Was hier abweicht:** beim Skalieren der Maske wird der nächste Punkt genommen,
MSE mittelt. Der Unterschied ist eine harte gegen eine weiche Kante an der
Maskengrenze.

**Gemessen** über die 190 Stile von `M15-Magic-Pack`: Befunde 297 → **285**
(#270).

**Quelle:** MSE2 `src/gfx/combine_image.cpp`, Zeilen 20–140;
`src/gfx/generated_image.cpp`, Zeilen 134–160; `src/gfx/blend_image.cpp`,
Zeilen 79–87 — Fassung f9d9356 vom 2021-07-06
**Erschlossen am:** 2026-08-17
**Sicherheit:** belegt
**Testfall:** TestCombineBlendMischtInDreiundzwanzigArten,
TestDieMischartenMitDivisionPruefenIhrenNenner,
TestSetMaskNimmtDasAlphaAusDemRotenKanal,
TestZweiBilderVerschiedenerGroesseSindEinBefund

---

## Zwei Stellen, an denen MSE schweigt und dieses Programm gemeldet hat

**`choice images:` endet auf `or else nil`:**

```cpp
// generate the script:
//     [a: {uvw}, b: {xyz}][input]() or else nil
```

`ChoiceStyle::initImage`, `src/data/field/choice.cpp`, Zeilen 171–198. Ein Wert
ohne hinterlegte Wahl ergibt **nichts**, und die Ebene bleibt leer. Das ist kein
Mangel, sondern die Bauform: `card symbol: none` hat kein Bild, weil es keines
haben soll.

**`+` fällt auf die Aneinandersetzung zurück:**

```cpp
} else {
  a = to_script(a->toString() + b->toString());
}
```

`src/script/context.cpp`, Zeile 512. Der letzte Zweig fängt **jede**
Kombination — eine Abbildung mit einer Zahl, ein Wahrheitswert mit einem Text.

**Beide Male stand hier ein Befund, und beide Begründungen waren erschlossen:**
„ein fehlendes Seltenheitssymbol ist etwas, das jemand sehen will" und „für sie
ist keine Fundstelle da". Über die 190 Stile von `M15-Magic-Pack` waren das
zusammen **127 Meldungen**, und keine nannte einen Fehler.

**Was das über die Zahl sagt:** eine Befundzahl misst, was dieses Programm
**nicht kann** — nicht, was eine Vorlage anders meint. Wer sie liest, muss
wissen, dass beide Sorten darin gestanden haben
(#270).

**Quelle:** MSE2 `src/data/field/choice.cpp`, Zeilen 171–198;
`src/script/context.cpp`, Zeilen 496–513 — Fassung f9d9356 vom 2021-07-06
**Erschlossen am:** 2026-08-17
**Sicherheit:** belegt
**Testfall:** TestEineWahlOhneBildZeichnetNichts,
TestPlusFaelltAufDieAneinandersetzungZurueck

---

## Ein Farbfeld zeichnet eine Fläche — mit runden Ecken oder als Rahmen

Ein Feld vom Typ `color` trägt keinen Text und kein Bild, sondern eine **Farbe**,
und sein Zeichner malt sie:

```cpp
void ColorValueViewer::draw(RotatedDC& dc) {
  dc.SetPen(*wxTRANSPARENT_PEN);
  dc.SetBrush(value().value());
  … const AlphaMask& alpha_mask = getMask(dc);
  if (alpha_mask.isLoaded()) {
    dc.DrawImage(alpha_mask.colorImage(value().value()), RealPoint(0,0), style().combine);
  } else {
    bool clip = style().left_width < style().width  && style().right_width  < style().width &&
          style().top_width  < style().height && style().bottom_width < style().height;
    if (clip) { … r.Subtract(… inneres Rechteck …); dc.getDC().SetClippingRegion(r); }
    dc.DrawRoundedRectangle(dc.getInternalRect(), style().radius);
  }
  drawFieldBorder(dc);
}
```

**Der schwarze Kartenrand ist so ein Feld.** `magic-m15.mse-style` gibt
`border color` die Maße der ganzen Karte, `radius: 18`, die Breiten 17/17/17/18
und `mask: border_mask.png` — gemalt wird also die Maske in der Farbe des
Feldes, und ohne Maske ein Rahmen aus vier Streifen.

**Die Vorgabe der vier Breiten ist 100000**, nicht null:

```cpp
ColorStyle::ColorStyle(const ColorFieldP& field)
  : … left_width(100000), right_width (100000)
  , top_width (100000), bottom_width(100000)
```

MSEs `clip` verlangt, dass **alle vier** kleiner als das Feld sind; eine einzelne
Angabe macht daraus deshalb keinen Rahmen. Wer die fehlenden mit null vorbelegte,
ließe von jedem solchen Feld die Füllung verschwinden.

**Der Wert eines Farbfeldes hat eine Vorgabe:**

```cpp
value( !field->initial.isDefault() ? field->initial()
     : !field->choices.empty()     ? field->choices[0]->color
     :                               Color()
     , true)
```

Also `initial:`, sonst die Farbe der **ersten Wahl**, sonst Schwarz. Eine Wahl
steht in zwei Formen da — als Block mit `name:` und `color:` oder einzeilig, und
dann ist der Name **zugleich** die Farbangabe (`parse_color` wird auf ihn
angewandt).

**Gedeutet wird eine Farbangabe in fünf Formen:**

```cpp
if (wxSscanf(v.c_str(),_("rgb(%u,%u,%u)"),&r,&g,&b))             return Color(r,g,b);
else if (wxSscanf(v.c_str(),_("rgba(%u,%u,%u,%u)"),&r,&g,&b,&a)) return Color(r,g,b,a);
else if (v.size() > 0 && v[0] == '#') { … 17 * r … }
else if (v == _("transparent")) return Color(0,0,0,0);
else { wxColour c = wxTheColourDatabase->Find(v); … }
```

`#rgb` verdoppelt jede Ziffer (`17 * r`), und was sich nicht deuten lässt, wird
`Color()` — Schwarz mit einer Warnung.

**Was dieses Programm davon nicht hat:** die Farbnamen sind ein **Teil** von
wxWidgets' Datenbank. Gezählt über die Vorlagendateien von `M15-Magic-Pack`
stehen als Name genau vier — `black` (716-mal), `white` (273), `red` (263) und
`grey` (2) —, und ein fünfter würde schwarz gezeichnet. Ebenso fehlt `combine:`
an dieser Ebene und die Glättung der Rundung: MSE malt mit einem `wxRegion`, und
dessen Kanten sind so hart wie unsere.

**Quelle:** MSE2 `src/render/value/color.cpp`, Zeilen 18–66;
`src/data/field/color.cpp`, Zeilen 41–89; `src/gfx/color.cpp`, Zeilen 42–74 —
Fassung f9d9356 vom 2021-07-06; `magic-m15.mse-style/style`, Zeilen 190–201;
`magic.mse-game/set_fields`, Zeilen 145–163
**Erschlossen am:** 2026-08-17
**Sicherheit:** belegt
**Testfall:** TestEineFlaecheZeichnetRundeEckenUndEinenRand,
TestEinFarbfeldZeichnetEineFlaeche, TestEineFarbangabeWirdGedeutet,
TestEinFarbfeldTraegtEineFarbe, TestEineWahlEinesFarbfeldesTraegtEineFarbe

---

## Ein Wahlbild wird eingepasst, ein Bildfeld gestreckt

Beide zeichnen ein Bild in einen Kasten, und sie tun es verschieden.

**Ein Auswahlfeld mit `render style: image`** bekommt die Maße seines Kastens und
`ASPECT_FIT`:

```cpp
opts.width  = (int) rot.trX(style.width);
opts.height = (int) rot.trY(style.height);
opts.preserve_aspect = (style.alignment & ALIGN_STRETCH) ? ASPECT_STRETCH : ASPECT_FIT;
```

Danach steht das Bild in seiner **eigenen** Größe im Kasten, ausgerichtet nach
`alignment:`:

```cpp
RealRect rect(align_in_rect(style.alignment, size, dc.getInternalRect()), size);
```

Und `ASPECT_FIT` heißt:

```cpp
if (iw * options.height > ih * options.width) { // too much height requested
  w = options.width;  h = options.width * ih / iw;
} else {
  w = options.height * iw / ih; h = options.height;
}
```

**Die Maße werden nur in diesem Zweig gesetzt.** Zeichnet das Feld auch Text
(`render style: both`), gilt die Originalgröße — `get_options` lässt `opts.width`
dann ungesetzt.

**Ein Bildfeld — das Kunstfenster — wird gestreckt:**

```cpp
if (image_load_file(image, *image_file)) {
  image.Rescale(w, h);
}
```

**Daran hängt das Seltenheitssymbol.** Sein Feld hat `alignment: middle right`,
`render style: image` und einen Kasten, der breiter als hoch ist; gestreckt wurde
daraus ein Balken quer über die Typzeile.

**Ohne Datei gilt MSEs Standardsymbol** — ein Quadrat von (0,2/0,2) bis
(0,8/0,8), vier Ecken, gerade Linien:

```cpp
if (filename.empty()) {
  the_symbol = default_symbol();
} else {
  the_symbol = package->readFile<SymbolP>(filename);
}
```

Ein Set ohne eigenes Symbol ist der Normalfall, und MSE malt dort dieses
Quadrat.

**Und ein Verlauf ohne Angaben läuft von der Mitte in die Ecke:**

```cpp
GradientSymbolFilter::GradientSymbolFilter()
  : center_x(0.5), center_y(0.5)
  , end_x(1), end_y(1)
{}
```

**Keine** Symbolfassung von `magic.mse-game` nennt `center x:` oder `end x:`. Mit
null als Vorgabe wird MSEs `len` eins und `t` überall null — und **jeder Verlauf
war eine einfarbige Fläche** in `fill color 1`. Nachgemessen am
Seltenheitssymbol von *Firebrand Incinerator*: gleichmäßig `rgb(214,196,94)`
statt eines Verlaufs nach `rgb(95,84,40)`.

**Was hier offen bleibt:** MSEs eigene Ausgabe derselben Karte zeigt an dieser
Stelle den **grauen** Verlauf der Fassung `uncommon` (`rgb(224,224,224)` nach
`rgb(84,84,84)`), obwohl die Karte `rarity: rare` führt. Form, Größe und Stelle
stimmen nach dieser Änderung überein, die Fassungswahl nicht — und woran das
liegt, ist an den vorliegenden Dateien nicht zu sehen. Es steht hier als Frage
und nicht als Regel.

**Quelle:** MSE2 `src/render/value/choice.cpp`, Zeilen 62–124;
`src/render/value/image.cpp`, Zeilen 18–39; `src/gfx/generated_image.cpp`,
Zeilen 30–68 und 453–471; `src/data/symbol.cpp`, Zeilen 342–378;
`src/render/symbol/filter.cpp`, Zeilen 138–161 — Fassung f9d9356 vom
2021-07-06; `magic-modules.mse-include/rarities/card_fields`, Zeilen 2–11;
`magic.mse-game/set_fields`, Zeilen 535–600
**Erschlossen am:** 2026-08-17
**Sicherheit:** belegt
**Testfall:** TestEinBildWirdEingepasstUndAusgerichtet,
TestEinWahlbildWirdEingepasst, TestOhneDateiGiltDasStandardsymbol,
TestOhneSetsymbolGiltDasStandardsymbol, TestEinVerlaufOhneAngabeLaeuftVonDerMitte

---

## Ein Kasten ohne Fläche verbirgt ein Feld — er ist kein Fehler

`magic-m15-saga.mse-style` schreibt am Textfeld

```
text:
	width: {if is_saga() then 0 else 153}
```

Das ist die Art, wie eine Vorlage ein Feld **abschaltet**: MSE setzt den Text in
einen Kasten ohne Fläche, zeichnet nichts und meldet nichts.

**Eine negative Größe ist etwas anderes.** Sie kann keine Vorlage meinen; sie
entsteht, wenn `right` links von `left` liegt — und dann ist die Rechnung falsch,
nicht die Absicht. Deshalb wird sie weiterhin gemeldet.

**Ebenso zählt ein unsichtbares Zeichen nicht als fehlend.** Ein
Nullbreiten-Leerzeichen (U+200B), der weiche Trennstrich (U+00AD) und die
Reihenfolgemarke (U+FEFF) haben in keiner Schrift eine Glyphe — das ist ihre
Bauform. Eine Meldung darüber zeigt auf nichts und stünde in derselben Liste wie
die echten.

**Und `require_choice` fehlte:**

```cpp
SCRIPT_FUNCTION(require_choice) {
  …
  SCRIPT_RETURN(filter_choices(input, choices, 1, (int)choices.size(), last_change));
}
```

`min` ist eins, `max` die Zahl der Wahlen — alle dürfen bleiben, keine nicht.
`magic.mse-game` ruft es an der Kartenfarbe
(`require_choice(choices: "multicolor, hybrid, land, artifact")`); ohne die
Funktion fielen auf jeder mehrfarbigen Karte zwei Wertskripte aus.

**Quelle:** `magic-m15-saga.mse-style/style`, Zeile 204;
MSE2 `src/script/functions/editor.cpp`, Zeilen 331–337 — Fassung f9d9356 vom
2021-07-06
**Erschlossen am:** 2026-08-17
**Sicherheit:** belegt
**Testfall:** TestEinTextrahmenDerGroesseNullSchweigt,
TestEinUnsichtbaresZeichenFehltNicht, TestRequireChoiceVerlangtEineWahl

---

## `paragraph height:` legt jeden Absatz in ein Band fester Höhe

**Nicht ein Zeilenabstand.** Die Zeilenfaktoren (`line height soft/hard/line`)
verschieben eine Zeile gegen die vorige; `paragraph height:` gibt jedem
**Absatz** ein eigenes Rechteck und richtet ihn darin aus:

```cpp
} else {
  // per paragraph alignment
  size_t start = 0;
  int n = 0;
  for (size_t last = 0 ; last < lines.size() ; ++last) {
    if (lines[last].break_after != LineBreak::SOFT || last+1 == lines.size()) {
      alignParagraph(start, last + 1, chars, style,
        RealRect(0, style.padding_top + n*style.paragraph_height,
                 s.width, style.paragraph_height));
      start = last + 1;
      ++n;
    }
  }
}
```

Ein Wert **von null oder darunter** heißt: der ganze Kasten wird als eines
ausgerichtet (ebd. Zeile 749). Die Absatzgrenze ist `break_after != SOFT`.

**Das setzt die Fähigkeiten eines Planeswalkers neben ihre
Loyalitätssymbole.** Die drei Fähigkeiten stehen in **einem** Textfeld;
`magic-m15-planeswalker.mse-style` schreibt `paragraph height: 49`, und die
Symbole sitzen fest bei 333, 383 und 432. Ohne die Angabe stehen die drei
Zeilen im Abstand einer Zeilenhöhe zusammengedrängt — gemessen 16,09 statt 49.

**Wo es steht** (nur streng als UTF-8 dekodierbare Textdateien): 16 Angaben in
9 Paketen von `M15-Magic-Pack`, 46 in 36 des vollen Packs — sämtlich
Planeswalker-Vorlagen.

**Was hier abweicht:** MSE rechnet die Gesamthöhe eines Zuges mit Absatzband
über eine eigene Formel — `max_height * n` mit `n` aus der Kastenhöhe (ebd.
Zeilen 718–732). Hier steht die Unterkante der letzten Zeile. Der Unterschied
betrifft, ob ein Zug als „passt nicht" gilt, nicht seine Stellen.

**Quelle:** MSE2 `src/render/text/viewer.cpp`, Zeilen 718–732 und 749–764,
Fassung f9d9356 vom 2021-07-06
**Erschlossen am:** 2026-08-18
**Sicherheit:** belegt
**Testfall:** TestEinFesterAbsatzBekommtSeinBand,
TestOhneAbsatzhoeheAendertSichNichts

---

## Das Ausgangsbild einer Bildverknüpfung heißt `input`

**Nicht `image`.** MSE schreibt es als `SCRIPT_PARAM_C` — benannt **oder**
stellungsweise:

```cpp
SCRIPT_FUNCTION(set_combine) {
  SCRIPT_PARAM(String, combine);
  SCRIPT_PARAM_C(GeneratedImageP, input);
```

So auch bei `to_image`, `set_alpha`, `saturate`, `invert_image`,
`recolor_image`, `enlarge`, `crop`, `flip_horizontal`, `flip_vertical`,
`rotate`, `drop_shadow` und `built_in_image`.

**Die Ausnahmen**: `set_mask` nennt es wirklich `image` (dazu `mask`),
`masked_blend` nennt `light`/`dark`/`mask`, `linear_blend` und `combine_blend`
nennen `image1`/`image2`.

**Gemessen** (nur streng als UTF-8 dekodierbare Textdateien): über
`M15-Magic-Pack` schreiben **643** Aufrufe `input:` und **kein einziger**
`image:`; über das volle Pack 787 gegen null.

**Quelle:** MSE2 `src/script/functions/image.cpp`, Zeilen 33–210, Fassung
f9d9356 vom 2021-07-06
**Erschlossen am:** 2026-08-18
**Sicherheit:** belegt
**Testfall:** TestEineBildverknuepfungNimmtInput,
TestDasBildDarfBenanntKommen

---

## `set_combine` setzt die Mischart der **Ebene**, nicht des Bildes

Das Bild geht unverändert durch — `SetCombineImage::generate` reicht es weiter.
Was die Verknüpfung setzt, ist die Art, wie die Ebene auf das Bisherige gelegt
wird.

**MSE fragt sie am Baum, ohne ein Bild zu erzeugen:**

| Knoten | `combine()` |
| --- | --- |
| Blatt und Unbekanntes | `COMBINE_DEFAULT` (`generated_image.hpp:51`) |
| `SimpleFilterImage` — `set_mask`, `set_alpha`, `saturate`, `invert_image`, `recolor_image`, beide Spiegelungen, `rotate`, `enlarge`, `crop`, `drop_shadow` | das Kind (ebd. :79) |
| `linear_blend` | `image1` (`generated_image.cpp:104`) |
| `masked_blend` | **`light`** (ebd. :122) |
| `combine_blend` | `image1` (ebd. :139) |
| `set_combine` | seine eigene (ebd. :178) |

**Beim Auflegen** unterscheidet MSE genau bei `NORMAL`:

```cpp
if (combine <= COMBINE_NORMAL) {
  dc.DrawBitmap(img, x, y);
} else {
  …Blit(…, &dc, x, y);          // den Zielausschnitt aufnehmen
  combine_image(source, img, combine);
  dc.DrawBitmap(source, x, y);
}
```

`combine_image` kopiert **zuerst den Alphakanal der Quelle** in das Ziel und
rechnet dann je Farbkanal `f(Ziel, Quelle)`. Das Ergebnis wird gewöhnlich
aufgelegt — die Deckung stammt also aus dem neuen Bild.

**Gemessen**: über das volle Pack steht `set_combine` **698**-mal in 41
Textdateien aus 34 Paketen — 667-mal mit `shadow`, 17-mal `multiply`, 4-mal
`glow`, 3-mal aus einer Stiloption, je einmal `greater than` und `hard light`.

**Was hier nicht gebaut ist**: MSE führt eine Mischart auch am **Feldstil**
(`REFLECT(combine)` in `data/field/choice.cpp:269` und `color.cpp:73`). In
`M15-Magic-Pack` schreibt kein Feldstil sie — gesucht wurde nach einer Zeile
`combine: <wort>` in den 190 Stilen, gefunden wurden null.

**Quelle:** MSE2 `src/gfx/combine_image.cpp`, Zeilen 108–157, und
`src/gfx/generated_image.{hpp,cpp}`, Fassung f9d9356 vom 2021-07-06
**Erschlossen am:** 2026-08-18
**Sicherheit:** belegt
**Testfall:** TestDieMischartKommtAusDemBildbaum,
TestDieMischartWirktBeimAuflegen,
TestEineUnbekannteMischartWirdBenannt,
TestDieMischartAnDenRaendern

---

## Eine Maske wird auf die Bildgröße gebracht, und gelesen wird Rot

**Zwei Aussagen in vier Zeilen:**

```cpp
void set_alpha(Image& img, const Image& img_alpha) {
  Image img_alpha_resampled = resample(img_alpha, img.GetWidth(), img.GetHeight());
  if (!img.HasAlpha()) img.InitAlpha();
  Byte *im = img.GetAlpha(), *al = img_alpha_resampled.GetData();
  size_t size = img.GetWidth() * img.GetHeight();
  for (size_t i = 0 ; i < size ; ++i) {
    im[i] = (im[i] * al[i*3]) / 255;
  }
}
```

`src/gfx/blend_image.cpp`, Zeilen 79–87. Eine Maske **darf** eine andere Größe
haben als ihr Bild; `al[i*3]` liest den **roten** Kanal.

**Für die Ebenenmaske sagt MSE dasselbe noch deutlicher**, mit Kommentar:

```cpp
// Copy red chanel to alpha
Byte* from = img.GetData(), *to = alpha;
for (size_t i = 0 ; i < n ; ++i) {
  to[i] = from[3*i];
}
```

`src/gfx/mask_image.cpp`, Zeilen 41–44.

**Nicht die Helligkeit.** Bei einer echten Graustufenmaske ist beides dasselbe;
bei einer farbigen nicht — die übliche Gewichtung ergibt für Rot 54 statt 255
und für Grün 182 statt 0. Und fremde Daten sind der Normalfall (§8).

**Was hier abweicht:** MSE mittelt beim Umrechnen (`resample`), dieses Programm
nimmt den nächsten Punkt. Der Unterschied ist eine harte gegen eine weiche
Kante an der Maskengrenze.

**Quelle:** MSE2 `src/gfx/blend_image.cpp`, Zeilen 79–87, und
`src/gfx/mask_image.cpp`, Zeilen 41–44, Fassung f9d9356 vom 2021-07-06
**Erschlossen am:** 2026-08-18
**Sicherheit:** belegt
**Testfall:** TestEineKleinereMaskeWirdAufDieBildgroesseGebracht,
TestDieEbenenmaskeLiestDenRotenKanal

---

## `angle:` dreht ein ganzes Feld

**Jedes Feld bekommt einen eigenen gedrehten Zeichenkontext**, und Text,
Bilder und Symbole laufen darin, als wäre nichts gedreht:

```cpp
Rotation ValueViewer::getRotation() const {
  return Rotation(deg_to_rad(getStyle()->angle), bounding_box, 1.0, getStretch());
}
```

`src/render/value/viewer.cpp`, Zeilen 44–46. `bounding_box` ist
`RealRect(left, top, width, height)` des Feldstils (`data/field.hpp:124`).
Die Umrechnung steht in `src/util/rotation.cpp`, Zeilen 37–42:

```cpp
return RealPoint(c * x + s * y + origin.x,
                -s * x + c * y + origin.y);
```

**Für die geraden Winkel** — `origin` ist (`links`, `oben`), (`bx`, `by`) ein
Punkt im ungedrehten Feld:

| Winkel | Stelle auf der Karte |
| ---: | --- |
| 0° | (links + bx, oben + by) |
| 90° | (links + by, oben − bx) |
| 180° | (links − bx, oben − by) |
| 270° | (links − by, oben + bx) |

**Das Feld verlässt dabei seinen eigenen Kasten**, und das ist so gewollt.
`name 2` von `magic-m15-aftermath` steht bei links 342, oben 295, breit 181,
hoch 21 mit `angle: 270`; gezeichnet wird damit die Spalte x 321…342,
y 295…476 — senkrecht am rechten Rand der unteren Kartenhälfte.

**Wo es steht** (in `*.mse-style/style`):

| Wurzel | Angaben | Stilpakete | häufigste Werte |
| --- | ---: | ---: | --- |
| `M15-Magic-Pack` | 75 | 19 von 190 | `270` (23), gerechnet (19), `180` (16) |
| `full-magic-pack` | 213 | 59 von 376 | `180` (88), gerechnet (33), `270` (29), `90` (25) |

**Was hier nicht gebaut ist:** schräge Winkel. MSE unterscheidet sie selbst
(`is_straight(angle)`); im vollen Pack stehen sie viermal (`41`, `3.5`,
`0.65`) gegen 142 gerade. Ein schräger Winkel wird **benannt** und die Ebene
ungedreht gezeichnet (§8).

**Quelle:** MSE2 `src/render/value/viewer.cpp`, Zeilen 44–46, und
`src/util/rotation.cpp`, Zeilen 16–42, Fassung f9d9356 vom 2021-07-06
**Erschlossen am:** 2026-08-18
**Sicherheit:** belegt
**Testfall:** TestEineGedrehteEbeneWirdGedrehtGezeichnet,
TestEinSchraegerWinkelWirdBenannt

---

## Eine Saga speichert ihre Kapitel einzeln, nicht als `rule text`

**Eine Sagakarte führt `chapter text` und `level_1_text` bis `level_4_text`
als eigene Kartenfelder** — und **keinen** `rule text`. So steht es in
`Cleric Class` aus dem kanonischen Set `White`:

```
chapter text:
    <i-auto>(Gain the next level as a sorcery to add its ability.)</i-auto>
level_1_text:
    If you would gain life, you gain that much life plus 1 instead.
level_2_text:
    <margin:0:0:30>Whenever you gain life, put a +1/+1 counter …</margin:0:0:30>
```

**Wer einer Saga trotzdem einen `rule text` gibt, sieht ihn zweimal** — und
das ist kein Fehler des Zeichners, sondern das, was `magic.mse-game` sagt:

```
chapter_text_filter_default :=
{
    if margin_code == "chapter_text" then (if not a_saga() then ""
        else if card.rule_text != "" then card.rule_text
        else "<i-auto>(" + saga_reminder() + ")</i-auto>")
    …
}
text_filter_default :=
{
    if transfer_levels() then
    (
        if margin_code == "lv1" then card.rule_text
        …
    )
    else ""
}
```

`magic.mse-game/script`, Zeilen 3174–3195, dazu `transfer_levels := {true}` in
Zeile 369. Der Regeltext geht damit an den Kapitelkasten **und** an Stufe 1.

**Woran das aufgefallen ist:** die Demokarte „Zeitalter" speicherte einen
`rule text` und zeigte ihren Text doppelt
(#482). Eine
Vorrichtung, die etwas speichert, was das Original nie schreibt, ist die
vierte Ausprägung derselben Klasse (§3).

**Quelle:** Datenpaket `M15-Magic-Pack` — `magic.mse-game/script`, Zeilen 369
und 3174–3195; kanonisches Set `White`, Karte `Cleric Class`
**Erschlossen am:** 2026-08-18
**Sicherheit:** belegt
**Testfall:** TestDemokartenZeichnen

---

## `scale down to:` — ein Text, der nicht passt, wird kleiner gesetzt

```cpp
double FontTextElement::minScale() const {
  return min(font->size(), font->scale_down_to) / max(0.01, font->size());
}
double FontTextElement::scaleStep() const {
  return 1. / max(font->size() * 4, 1.);
}
```

`src/render/text/font.cpp`, Zeilen 51–56. MSE sucht dann zwischen dieser
Schranke und 1 einen Maßstab, bei dem der Text in den Kasten passt
(`prepareLinesTryScales`, `src/render/text/viewer.cpp`, Zeilen 444–500) —
**nicht den größten**, sondern bis zu einem Schritt darunter (unten). Der
Kommentar dort sagt „common case is: no".

**Es dehnt nicht**, es verkleinert nur: die obere Schranke ist 1.

**Fehlt die Angabe, wird nicht verkleinert.** MSEs Vorgabe ist 100000
(`src/data/font.cpp`, Zeile 18), also größer als jede Schrift; `min(size,
scale_down_to)` ist dann `size` und der Maßstab 1.

**Wo es steht:**

| Wurzel | Angaben | Stilpakete |
| --- | ---: | ---: |
| `M15-Magic-Pack` | **571** | 174 von 190 |
| `full-magic-pack` | 921 | 348 von 376 |

**Wie MSE sucht, und warum der Text kleiner steht als nötig:** es beginnt beim
Maßstab 1, schätzt aus der Unterkante der letzten Zeile eine untere und eine
obere Schranke (`bound_on_min_scale`, `bound_on_max_scale`, Zeilen 429–444)
und halbiert, bis beide weniger als einen Schritt `1/(4·size)` (mindestens
0,01) auseinanderliegen. **Der gefundene Maßstab liegt bis zu einem Schritt
unter dem größten passenden** — bei 14 Punkten bis zu 1,8 %. Passt der Text,
heißt das: die Unterkante der **letzten** Zeile liegt über dem unteren
Polster (Zeile 733), auch wenn sie leer ist; mit `paragraph height:` die
höchste Absatzhöhe mal die Zahl der Bänder, die der Kasten fasst (Zeilen
718–732). Der Maßstab des vorigen Laufs, mit dem MSE beginnt, ist nach jeder
Wertänderung 1 (`TextViewer::reset(false)`, Zeile 169). Symbole schrumpfen
mit demselben Faktor (`src/render/text/symbol.cpp`, Zeilen 22–26).

**Bis zum 2026-10-11 halbierte dieses Programm zehnmal** und fand den größten
passenden Maßstab — der Text stand größer als bei MSE2, ein Text mit leerer
letzter Zeile hinter `<line>` passte, wo er bei MSE2 nicht passt, und ein Feld
mit Absatzbändern wurde verkleinert, wo MSE2 es nicht tut (#2244).

**Quelle:** MSE2 `src/render/text/font.cpp`, Zeilen 51–56, und
`src/render/text/viewer.cpp`, Zeilen 444–500, Fassung f9d9356 vom 2021-07-06
**Erschlossen am:** 2026-08-18, die Suche nachgelesen am 2026-10-11
**Sicherheit:** belegt
**Testfall:** TestEinZuLangerTextWirdKleinerGesetzt, TestScaleDownToWirdGelesen, TestDerMassstabFolgtMSE2sSuche, TestEineLeereLetzteZeileZaehltBeimPassen, TestMitAbsatzbaendernZaehltDieHoechsteAbsatzhoehe

---

## `installer group:` und `position hint:` ordnen ein Paket in eine Liste

Zwei Kopfangaben, die dieses Projekt bis
#244
übersprungen hat:

```
installer group: magic/m15 style/normal cards
position hint: 095
```

- **`installer group:` ist ein Pfad mit `/` als Trenner.** MSE zerlegt ihn und
  legt je Stufe einen Baumknoten an; zwei Pakete derselben Gruppe teilen sich
  ihre Knoten.
- **Fehlt sie, setzt MSE `custom`** — und das ist wörtlich diese Zeichenkette:
  `_(S)` ist bei MSE ein Makro für Unicode-Literale (`src/util/string.hpp`,
  Zeile 58), kein Übersetzungsaufruf. Gezeichnet wird sie über
  `capitalize_sentence`, also als „Custom".
- **`position hint:` ordnet innerhalb einer Gruppe**, kleiner zuerst. Die Packs
  schreiben dreistellig mit führender Null (`095`, `071`), damit die Zahlen im
  Texteditor untereinander stehen — **Basis 10 ausdrücklich**: mit Basis 0 wäre
  `095` ein Fehler und `071` die 57.
- **Fehlt sie, gilt 1.000.000** (`src/gui/package_update_list.hpp`, Zeile 52).
  Groß heißt hinten: ein Paket ohne Hinweis steht *hinter* jedem mit.
- **Ein Gruppenknoten erbt die kleinste Einordnung seiner Kinder** — sonst
  stünde jede Gruppe hinten, weil ein frischer Knoten mit `TYPE_OTHER` und
  einer Million anfängt.
- **Vor der Gruppe kommt die Art.** MSE liest sie aus der **Endung** des
  Paketnamens und ordnet sie: Programm, Locale, Spiel, Stil, Exportvorlage,
  Symbolschrift, Include, Schrift, Übriges. Das ist eine Aussage darüber, was
  zuerst gebraucht wird — ohne Spiel kein Stil.
- **Führender Leerraum kommt vor.** Vier Stile im `full-magic-pack` schreiben
  ihre Gruppe mit einem Tabulator davor; roh zerlegt entstünde ein Knoten
  „⇥magic" neben „magic".

**Wo es steht:**

| Wurzel | mit `installer group` | mit `position hint` | Gruppen |
| --- | ---: | ---: | ---: |
| `full-magic-pack` | **416** von 526 | 495 von 526 | 332 |
| `M15-Magic-Pack` (installiert) | 225 von 324 | — | 156 |

Im installierten `M15-Magic-Pack` liegen 208 der 324 Pakete unter dem einen
Knoten `magic`, 13 unter `translations`, 99 unter keinem.

**Quelle:** MSE2 `src/gui/package_update_list.cpp`, Zeilen 18–95 und 200–212,
`src/gui/package_update_list.hpp`, Zeilen 45–75, und
`src/gui/packages_window.cpp`, Zeile 222, Fassung f9d9356 vom 2021-07-06; dazu
die Kopfangaben in `magic-m15-punk-show.mse-style/style` (Zeilen 6–7) und
`magic-m15-duplex.mse-style/style` (Zeilen 6–7)
**Erschlossen am:** 2026-08-19
**Sicherheit:** belegt — Regel aus dem Quelltext, Zahlen aus zwei Sammlungen
**Testfall:** TestEinPaketNenntSeineEinordnung, TestOhneEinordnungStehtDasPaketHinten,
TestEinUnlesbarerRangIstEinBefund, TestEineGruppeWirdGesaeubert,
TestDieArtenStehenInMSEsReihenfolge, TestDieEinordnungStehtImBestand,
TestDerVerwalterUebernimmtDieEinordnung

Den Baum selbst — Zerlegen, Zusammenfassen, Ordnen — hält
`frontend/src/seiten/paketbaum.test.ts` mit zwölf Fällen; der Wächter über
dieser Datei prüft nur Go-Testnamen, und ein Vitest-Name darin machte ihn rot.

---

## Ein Bildwert trägt einen **Dateinamen**, keinen Ausschnitt

```cpp
class ImageValue : public Value {
  ValueType filename;    ///< Filename of the image (in the current package), or ""
  Age       last_update; ///< When was the image last changed?
};
```

`src/data/field/image.hpp`, Zeilen 48–56. Mehr steht nicht darin — kein
Rechteck, kein Maßstab, kein Versatz.

**Der Zuschnitt geschieht beim Einsetzen und ist endgültig:**

```cpp
ImageSliceWindow s(…, image, style().getSize(), mask);
if (s.ShowModal() == wxID_OK) {
  LocalFileName new_image_file = getLocalPackage().newFileName(field().name, _(""));
  Image img = s.getImage();
  img.SaveFile(getLocalPackage().nameOut(new_image_file), wxBITMAP_TYPE_PNG);
  addAction(value_action(valueP(), new_image_file));
}
```

`src/gui/value/image.cpp`, Zeilen 44–52. Der Dialog gibt das **fertig
zugeschnittene** Bild; gespeichert wird es als neue PNG-Datei im Setpaket, und
im Set steht nur ihr Name.

**Wer den Ausschnitt später ändern will, setzt das Bild neu ein.** Das
Ursprungsbild ist zu diesem Zeitpunkt nicht mehr im Set — was dort liegt, ist
schon der Ausschnitt.

**Immer PNG**, auch wenn die Quelle ein JPEG war: `wxBITMAP_TYPE_PNG` steht
fest da, mit dem Kommentar „see #69. Disk space is cheap anyway."

**Warum diese Feststellung negativ ist und trotzdem hier steht:**
#246 verlangt,
vor dem Bau eines Zuschnittdialogs nachzusehen, „was davon belegt ist". Die
Antwort ist **nichts** — und die gehört aufgeschrieben, damit der nächste
Leser nicht dieselbe Stelle noch einmal sucht.

**Quelle:** MSE2 `src/data/field/image.hpp`, Zeilen 48–56, und
`src/gui/value/image.cpp`, Zeilen 36–53, Fassung f9d9356 vom 2021-07-06
**Erschlossen am:** 2026-08-19
**Sicherheit:** belegt — aus dem Quelltext, nicht aus einer Datei erschlossen
**Testfall:** TestDerZuschnittFolgtMSEsVorgabewahl, TestDerZuschnittHatDieGroesseDesFeldes

---

## Ein `.mse-locale` übersetzt die Oberfläche von MSE2 — und Feldnamen **nicht**

Ein Sprachpaket ist ein Verzeichnis `<kürzel>.mse-locale` mit einer Datei
`locale` und einem Flaggenbild. Es führt zwei Dinge, und nur eines davon
betrifft Pakete.

**Erstens die Oberflächentexte von MSE2 selbst**, nach Kategorie geordnet:

```
menu:
	file:				&Datei
	new set:			&Neu...
label:
	card:				Karte
error:
	…
```

Die zehn Kategorien sind ein Enum (`LOCALE_CAT_MENU` bis `LOCALE_CAT_ERROR`,
`src/util/locale.hpp`, Zeilen 27–39). **Für dieses Programm ohne Belang**: die
Oberfläche hat eigene Kataloge unter `frontend/src/sprache/` (§9).

**Zweitens ein `package:`-Block**, je Paket eine Unterliste:

```
package:
	magic.mse-game:
		name:		Name
		cc:			Manakosten
		type:		Typ
	magic-mana-*.mse-symbol-font:
		menu item W:	&Weißes Mana	 W
```

Der Schlüssel ist der **Dateiname des Pakets** (`pkg.relativeFilename()`), und
er darf `*` als Platzhalter tragen — verglichen wird als regulärer Ausdruck,
in dem `.` maskiert und `*` zu `.*` wird (`find_wildcard` und
`match_wildcard`, `src/data/locale.cpp`, Zeilen 57–66). Der erste passende
Eintrag gewinnt, und das Ergebnis wird unter dem gesuchten Namen **gemerkt**.

### Die Beschriftung eines Feldes kommt **nicht** von hier

Das ist die Feststellung, um die es geht — und sie widerlegt die Annahme, mit
der #561
eröffnet wurde.

`Field::caption` ist ein `LocalizedString` und wird aus dem **Paket** gelesen:

```cpp
class LocalizedString {
  String default_;
  unordered_map<String, String> translations;
};

#define REFLECT_LOCALIZED_N(name, var) \
  handler.handle(name, var.default_); \
  handler.handle(_("localized_") name, var.translations);
```

`src/data/localized_string.hpp`, Zeilen 19–30. In der Paketdatei steht damit

```
caption: Name
localized caption:
	de: Name
```

und `get()` wählt `translations[the_locale->name()]`, sonst `default_`
(`src/data/locale.cpp`, Zeilen 124–137). **Ist `caption:` leer, wird sie aus
dem Feldnamen gebildet:**

```cpp
if (caption.default_.empty()) caption.default_ = name_to_caption(name);
```

`src/data/field.cpp`, Zeile 69.

**Der `package:`-Block hat mit Feldern nichts zu tun.** Er wird über
`tr(const Package&, key, def)` gelesen, und diese Funktion trägt in 2.1
`[[deprecated]]` (`src/util/locale.hpp`, Zeilen 50–56). Sie hat **zwei**
Aufrufer im ganzen Quelltext, beide im Rasterer:

```cpp
String text = tr(viewer.getStylePackage(), value, capitalize_sentence);
```

`src/render/value/choice.cpp`, Zeile 99, und wortgleich
`src/render/value/multiple_choice.cpp`, Zeile 74. Beide übersetzen einen
**Wahlwert**, und beide schlagen unter dem **Stilpaket** nach.

### Gemessen: im ganzen Pack findet dieser Weg nichts

Über `full-magic-pack`, am 2026-08-20: **alle dreizehn** Sprachpakete führen
unter `package:` genau zwei Blöcke — `magic.mse-game` und
`magic-mana-*.mse-symbol-font`. Gesucht wird aber unter `magic-m15.mse-style`
und seinesgleichen; darauf passt keiner der beiden, auch nicht über die
Platzhalterregel.

**Was bleibt, ist der Rückfall** — `capitalize_sentence`, also der erste
Buchstabe groß (`src/util/string.cpp`, Zeilen 115–119). Seine Reichweite ist
ebenfalls gemessen: über 376 Stilpakete zeichnen von 1.359 Auswahlfeldstilen
**fünf** überhaupt Text; im kanonischen `magic-m15.mse-style` sind es **null**.

**Warum auch diese Feststellung negativ ist und trotzdem hier steht:** #561
verlangt, „nachgelesen und aufgeschrieben, **wie** MSE einen Locale-Eintrag
einem Feld zuordnet". Die Antwort ist: **gar nicht** — und die gehört
aufgeschrieben, damit der nächste Leser nicht dieselbe Stelle noch einmal
sucht.

**Quelle:** MSE2 `src/data/locale.cpp`, Zeilen 55–120, `src/data/localized_string.hpp`,
Zeilen 19–30, `src/data/field.cpp`, Zeile 69, `src/util/locale.hpp`, Zeilen 27–56,
`src/render/value/choice.cpp`, Zeile 99, Fassung f9d9356 vom 2021-07-06
**Erschlossen am:** 2026-08-20
**Sicherheit:** belegt — aus dem Quelltext, dazu die Zählung über `full-magic-pack`
**Testfall:** TestEinGezeichneterWahlwertBeginntGross, TestDieBeschriftungEinesFeldes


---

## `styling:` — die Stiloptionen des **Sets**, je Stilpaket

Neben `set info:` steht in der `set`-Datei ein Block `styling:`, eine Ebene je
Stilpaket:

```
styling:
	magic-m15-saga:
		text_box_mana_symbols: magic-mana-small.mse-symbol-font
	magic-m15-mainframe-dfc:
		text_box_mana_symbols: magic-mana-small.mse-symbol-font
		level_mana_symbols: magic-mana-large.mse-symbol-font
```

**Der Schlüssel ist der Paketname mit Präfix** — `magic-m15-saga`, nicht die
Kurzform `m15-saga`:

```cpp
IndexMap<FieldP, ValueP>& Set::stylingDataFor(const StyleSheet& stylesheet) {
  return styling_data.get(stylesheet.name(), stylesheet.styling_fields);
}
IndexMap<FieldP, ValueP>& Set::stylingDataFor(const CardP& card) {
  if (card && card->has_styling) return card->styling_data;
  else                           return stylingDataFor(stylesheetFor(card));
}
```

`src/data/set.cpp`, Zeilen 99–105. `Packaged::name()` ist der Dateiname ohne
Endung.

**Die Reihenfolge steht im `if`/`else`**: eine Karte mit `has_styling: true`
benutzt **ihre** Werte; jede andere die des Sets; und was auch dort fehlt,
steht auf dem `initial:` des Stils.

**Alle zehn kanonischen Sets führen den Block** (nachgesehen am 2026-08-20),
und `magic-set-generator` schreibt ihn ebenfalls — mit demselben präfixierten
Schlüssel (`src/mse.rs`, Zeilen 151–174, §11). Da alle 25.148 kanonischen
Karten `has_styling: true` tragen, ist der Zweig über den Kanon **nicht**
erreichbar; über ein von jenem Werkzeug geschriebenes Set schon.

**Hier wird er flach abgelegt** — ein Setwert `styling`, eine Zeile je Option
als `<paket>/<name>: <wert>`. Derselbe Weg wie bei der Karte, deren
`styling_data:` als gewöhnlicher Feldwert liegt; eine eigene Tabelle wäre eine
zweite Antwort auf dieselbe Frage.

**Quelle:** MSE2 `src/data/set.cpp`, Zeilen 99–105, Fassung f9d9356 vom
2021-07-06; dazu die zehn kanonischen Sets und `fenhl/magic-set-generator`,
`src/mse.rs`, Zeilen 151–174
**Erschlossen am:** 2026-08-20
**Sicherheit:** belegt — aus dem Quelltext und an echten Sets nachgesehen
**Testfall:** TestDieSetweitenStiloptionenGelten, TestDerSetweiteStilblockWirdEinSetwert


---

## `default:` an einer Stiloption ist ein **Skript** und schlägt `initial:`

**Eine Stiloption führt beide Angaben, nicht die eine oder die andere.** Bis
zum 2026-08-21 stand in `msepaket` der Satz, der Vorgabewert heiße „`initial:`
statt `default:`" — erschlossen und falsch.

**Die Reihenfolge ist die des Konstruktors und des Wertumlaufs:**

```cpp
ColorValue::ColorValue(const ColorFieldP& field)
  : Value(field)
  , value( !field->initial.isDefault() ? field->initial()
         : !field->choices.empty()     ? field->choices[0]->color
         :                               Color()
         , true)
{}

bool ColorValue::update(Context& ctx) {
  bool change = field().default_script.invokeOnDefault(ctx, value)
```

Das zweite Argument von `Defaultable` ist `true`: der Konstruktor setzt
`initial` **und markiert den Wert zugleich als Vorgabe**. `invokeOnDefault`
läuft genau dann — `default:` überschreibt `initial:` also, solange niemand die
Option angefasst hat.

**Und der Wert bleibt danach Vorgabe.** `change(Defaultable<T>&, …)` ruft
`assignDontChangeDefault`; die Markierung überlebt das Skript. Deshalb gibt
`ColorValue::toString()` weiterhin `default_name` zurück — ohne eigene Angabe
das übersetzte Wort „Default". **Im ganzen Pack steht `default name:` kein
einziges Mal**, die Beschriftung ist also immer diese.

**`default:` ist Quelltext, keine Zeichenkette** (`OptionalScript
default_script`). Am Orakel nachgefragt:

| Geschrieben | Ergibt |
| --- | --- |
| `"black"` | die Zeichenkette `black` |
| `rgb(0,0,0)` | die Farbe `rgb(0,0,0)` |
| `black` | **Fehler** — `Variable not set: black` |

**Gezählt über die 376 Stilpakete von `full-magic-pack`**: von 4.067
`styling field:`-Blöcken sind 105 vom Typ `color`, und **zehn** davon tragen
ein `default:` — achtmal `"black"`, einmal `rgb(0,0,0)`, einmal das nackte
`black`. Das letzte steht in `magic-m15-bigtext` neben einem `initial: black`,
das den Fehler auffängt.

**Und keiner der zehn bewegt einen Bildpunkt**: acht führen vier `choice:`
-Blöcke, deren erster schwarz ist, die beiden anderen fallen auf den leeren
Rückfall — der ebenfalls schwarz ist. Die Angabe war bei uns folgenlos, **weil
sie zufällig mit dem Rückfall zusammenfiel**, und genau deshalb ist sie zehn
Monate lang niemandem aufgefallen.

**Quelle:** MSE2 `src/data/field/color.cpp`, Zeilen 82–101, und
`src/script/scriptable.hpp`, Zeilen 36–78, Fassung f9d9356 vom 2021-07-06;
dazu `magicseteditor --cli … -q --raw` und die Stilpakete von
`full-magic-pack`
**Erschlossen am:** 2026-08-21
**Sicherheit:** belegt — aus dem Quelltext, am Orakel bestätigt und über das
Pack gezählt
**Testfall:** TestDieVorgabeEinerStiloption, TestDieVorgabeKommtAlsFarbeAn,
TestEineStiloptionFuehrtIhrenVorgabeausdruck


---

## Was MSE2 ausgibt — gezählt, nicht erschlossen

**Quelle:** MSE2, Fassung `f9d9356` vom 2021-07-06 — `src/data/format/formats.cpp`, `src/gui/set/window.cpp`, `src/data/format/image.cpp`, `src/gui/images_export_window.cpp`
**Erschlossen am:** 2026-08-22
**Sicherheit:** belegt
**Testfall:** — (eine Aussage über fremden Quelltext; die Zählungen stehen unten wörtlich, und `docs/PORTIERUNGSKARTE.md` führt die Kennzeichnungen)

Anlass: #824.

Phase P sagt „mindestens MSE2s Umfang", und §15 sagt, dass eine solche Menge am
Quelltext nachgelesen wird und nicht aus einem Menü. Nachgelesen ist sie hier.

### Ganze Sets

`init_file_formats()` führt drei und keine vierte (`formats.cpp:19-21`). Ein-
und Ausfuhr gehen über dieselbe Liste; **welche Richtung ein Format kann,
entscheiden `canImport()` und `canExport(game)`** — und die Antworten sind
nicht dieselben:

| Format | Name im Dialog | liest | schreibt | Fundstelle |
| --- | --- | --- | --- | --- |
| `mse2` | — | **ja** | **ja** | `mse2.cpp:22-23` |
| `mse1` | „Magic Set Editor version 1 files (\*.mse)" | **ja** | nein | `mse1.cpp:26-28` |
| `mtg_editor` | „Mtg Editor files (\*.set)" | **ja** | nein | `mtg_editor.cpp:27-29` |

**MSE2 schreibt ein ganzes Set also in genau einem Format.** Die anderen zwei
sind reine Lesewege für alte Dateien — 172 und 275 Zeilen C++, für Programme,
die seit über fünfzehn Jahren nicht mehr gepflegt werden.

### Bilder einer Karte

`src/gui/set/window.cpp:660` — der Filter des Speichern-Dialogs, wörtlich:

```
PNG images (*.png)|*.png|JPEG images (*.jpg)|*.jpg|Windows bitmaps (*.bmp)|*.bmp|TIFF images (*.tif)|*.tif
```

**Vier Formate: PNG, JPG, BMP, TIFF.**

Bestimmt wird das Format an der **Endung**, nicht am Filter:
`export_image` ruft `img.SaveFile(filename)` ohne Typ (`image.cpp:23`), und der
Kommentar daneben sagt warum — `wxImage` erkennt ihn selbst. Möglich ist damit
alles, wofür ein Handler registriert ist, und `main.cpp:89` ruft
`wxInitAllImageHandlers()`. **Der Filter ist die Zusage, nicht die Grenze.**

### Bilder vieler Karten

`src/gui/images_export_window.cpp` — ein Fenster mit einem
**Dateinamensmuster** und einer Regel für Namenskonflikte. Der im Dialog
gewählte Name wird **verworfen**; die Beschriftung sagt es selbst
(`filename is ignored`, Zeilen 63-64).

### HTML

`src/data/format/html.cpp` und `src/gui/html_export_window.cpp` — über
**Exportvorlagen**, also über MSE-Script, nicht über einen festen Schreiber.

### Was MSE2 **nicht** kann

| Format | Zählung |
| --- | --- |
| **PDF** | `grep -rin pdf src/` → **0 Treffer** |
| **SVG** | `grep -rin svg src/` → **0 Treffer** |

**Über den ganzen `src`-Baum gezählt**, nicht über ein Verzeichnis (§10). Beides
sind Zutaten dieses Projekts und in `docs/PORTIERUNGSKARTE.md` als **new**
gekennzeichnet — mit dem Satz, wo gesucht wurde.

### Wie viele Exportvorlagen laufen — und was die Zahl nicht enthält

**Quelle:** dieses Projekt gegen `M15-Magic-Pack` und `full-magic-pack` ·
**Erschlossen am:** 2026-08-22 ·
**Sicherheit:** belegt ·
**Testfall:** `TestExportvorlagenMitEchterKarte`, `TestExportvorlagenAusFremdpaketen`

Mit einer echten Karte aus `Vanguard.mse-set`:

| Wurzel | gelesen | liefen durch | außer Reichweite |
| --- | --- | --- | --- |
| `M15-Magic-Pack` | 19 | **18** | 1 |
| `full-magic-pack` | 22 | **21** | 1 |

**Beide Male dieselbe eine**: `magic-lackeyccg` liest `sts.card_regions`, das
es an der Fassung `f9d9356` nicht gibt
([0036](entscheidungen/0036-card-regions-ist-ausser-reichweite.md)).

#### Was die Zahl **nicht** sagt

- **Nicht, dass eine Vorlage das Richtige ausgibt.** Geprüft ist, dass sie
  **etwas** ausgibt und nicht scheitert. Ob der Text stimmt, weiß nur, wer ihn
  liest.
- **Nichts über andere Karten.** Eine Vorlage, die ein Feld liest, das diese
  eine Karte nicht führt, liefe hier durch und anderswo nicht. Der zweite Lauf
  deckt das von der anderen Seite ab: eine Probekarte mit fünf gefüllten
  Feldern und dem Rest aus dem Spiel.
- **Nichts über eine leere Ausgabe.** Sie wird **genannt**, nicht
  weggerechnet: `magic-namecheck` gibt für eine unauffällige Karte nichts aus,
  und das ist richtig so. Unterscheiden kann der Lauf die beiden nicht — ob
  leer richtig ist, weiß nur die Vorlage.
- **Nichts über eine andere Einstellung.** `card_regions` steht auch in
  `magic-cockatrice-v2` und `magic-lackeybot`; die laufen durch, weil der Zweig
  mit den Vorgabewerten nicht betreten wird. Wer sie anders einstellt, kann
  dort dasselbe treffen.

**Die Zusage im Wächter ist nicht die Zahl, sondern die Vollständigkeit**: jede
Vorlage, die in der gesetzten Wurzel liegt, wird gelesen — gezählt aus dem
Verzeichnis, nicht als feste Zahl. Bis
#779 stand dort
`!= 19`, und der Lauf wurde gegen das größere Pack rot, ohne dass etwas kaputt
war.

### Welche Einfuhr etwas bedeutet — und welche nicht

**Quelle:** dieses Projekt, gegen MSE2 `f9d9356` ·
**Erschlossen am:** 2026-08-22 ·
**Sicherheit:** belegt ·
**Testfall:** — (eine Aussage über den Zuschnitt; die Fundstellen stehen daneben)

Der Abnahmekasten von Phase P verlangt, dass für **jedes** Format entschieden
ist, ob eine Einfuhr überhaupt etwas bedeutet —
#841.

| Format | Einfuhr | Warum |
| --- | --- | --- |
| `.mse-set` | **gebaut** | `SetOeffnen`, `BeispielsetOeffnen` |
| MTG JSON | **gebaut**, Datei und Verzeichnis | `MTGJSONOeffnen` — **MSE2 kann das nicht** ([0009](entscheidungen/0009-phase-9-fremdformate-und-schnittstelle.md)) |
| Kartenbild (PNG, JPG, GIF) | **gebaut** | `BildSetzen`, `BildAusDatei` — das Bild wird ein **Feldwert**, keine Karte |
| PNG, JPG, BMP, TIFF **einer gerenderten Karte** | **bedeutungslos** | siehe unten |
| PDF (Druckbogen) | **bedeutungslos** | dito |
| SVG | **bedeutungslos** | dito |
| HTML aus einer Exportvorlage | **bedeutungslos** | dito |
| `mse1`, `mtg_editor` | **nicht gebaut**, und das ist eine Abwägung | siehe unten |

#### Warum eine gerenderte Karte nicht zurückkommt

**Sie trägt keine Feldstruktur mehr.** Ein PNG einer Karte ist ein Raster von
Farbwerten; welcher Bildpunkt zu `name` gehörte und welcher zu `rule text`,
steht nirgends darin. Dasselbe gilt für PDF, SVG und HTML: sie sind
**Darstellungen** eines Sets, nicht seine Daten.

Die Ausfuhr in diese Formate ist damit eine Einbahnstraße, und das ist keine
Lücke, sondern ihre Natur. Wer ein Set weitergeben und zurückbekommen will,
gibt `.mse-set` oder MTG JSON weiter.

**SVG ist der Grenzfall und ändert nichts daran.** Es trägt je Ebene eine Marke
mit `data-feld`, also mehr Struktur als ein PNG — aber die Ebenen sind
**gezeichnet** und nicht die Werte: aus `<text>Feuerball</text>` ließe sich der
Name zurückholen, aus einem umgebrochenen Regeltext mit einem `<tspan>` je
Zeile nicht mehr der Quelltext mit seinen `<i>`-Marken. Eine Einfuhr, die drei
von zwanzig Feldern trifft, ist schlechter als keine.

#### Warum `mse1` und `mtg_editor` fehlen

**Es ist eine Abwägung, keine Übersehung.** Dagegen sprechen: 172 und 275
Zeilen fremder Leser, Dateiformate zweier Programme, die seit über fünfzehn
Jahren nicht mehr gepflegt werden — und **hier liegt keine einzige Testdatei**
für eines von beiden, weder im Pack noch in den kanonischen Sets. Ein Leser
ohne echte Eingabe ist eine Vermutung mit Testabdeckung (§3, §15).

Dafür spricht: wer eine `.mse`-Datei von 2006 hat, hat sie nur dort.

**Für die Freigabe 0.1 überwiegt das Erste.** Wer es ändern will, braucht
zuerst eine echte Datei je Format — dann ist es ein Werkstück wie jedes andere,
und `mse1.cpp` beziehungsweise `mtg_editor.cpp` sagen, was zu tun ist.

### Und was hier fehlte, das MSE2 hat — seit #826 nicht mehr

**BMP und TIFF.** Für die *Einfuhr* eines Kartenbildes bleibt es dabei, und der
Grund steht in `internal/app/dialogtexte.go`: die Standardbibliothek liest beide
nicht, und ein Filter, der eine Datei zulässt, die danach abgewiesen wird, ist
schlechter als keiner.

**Für die Ausfuhr galt der Grund nie**: `golang.org/x/image` — damals v0.44.0,
bereits gepinnt und bereits Abhängigkeit — bringt `bmp.Encode` (`bmp/writer.go:162`)
und `tiff.Encode` (`tiff/writer.go:292`) mit. Beide sind seit
#826 gebaut;
damit ist der Umfang erreicht.

**Was dabei gemessen wurde**, statt es der Formatbeschreibung zu glauben: ein
Bild mit durchsichtiger Ecke, geschrieben und zurückgelesen.

| Format | Ecke danach |
| --- | --- |
| PNG | durchsichtig |
| TIFF | durchsichtig |
| JPG | weiß |
| BMP | weiß |

Und: von den vier Kodierern scheitert an einem Bild **ohne Fläche** nur PNG und
TIFF — JPEG und BMP schreiben es klaglos.

### Was eine SVG vom Zeichenplan trägt — und was nicht

**Sie entsteht aus dem Plan, nicht aus Bildpunkten.** Ein Rasterbild in ein
`<svg>` zu wickeln wäre der Name ohne die Sache; die Flächen sind Geometrie,
der Text ist Text.

| Was der Plan trägt | Im SVG |
| --- | --- |
| Fläche, Farbe, runde Ecken | `<rect>` mit `rx` |
| Rand aus vier Breiten | vier `<rect>`, MSEs `clip` entscheidet |
| Verlauf | `<linearGradient>` in `userSpaceOnUse` |
| Deckung | `opacity` an der Gruppe; `Deckung ≤ 0` fehlt ganz, wie im Bild |
| Winkel | `translate(links,oben) rotate(−Winkel)` |
| Maske | `<mask>` mit dem **roten** Kanal als Leuchtdichte |
| Bild, Symbol | eingebettetes PNG aus **demselben** Bildbaum wie beim Rastern |
| Passung | `preserveAspectRatio`, alle neun Ausrichtungen |
| Textumbruch, Ausrichtung, Polster, Verkleinerung | `<tspan>` je Stück, aus `render.SetzeAbschnitte` |
| Stauchung | `translate(kante,0) scale(f,1) translate(−kante,0)` |
| Laufweite | `letter-spacing` |
| **Mischart** | **nicht angewandt — je Ebene gemeldet** |

**Die letzte Zeile ist die einzige Auslassung, und sie ist begründet.** CSS
`mix-blend-mode` kennt `add`, `subtract`, `stamp`, `negation`, `reflect`,
`shadow` und `glow` nicht; `shadow` allein steht **667 der 698** Aufrufe im
vollen Pack. Fünf der zwanzig führt CSS gleichlautend — sie zu übernehmen und
die übrigen nicht ergäbe eine Ausgabe, die mal stimmt und mal nicht, ohne dass
jemand es sieht. Ein Wächter hält die Liste vollständig
(`internal/werkzeuge/svgebene_test.go`): eine neue Eigenschaft an
`render.Ebene`, die der Schreiber nicht liest, macht die Pipeline rot.

**Zwei Ebenenarten gehen durch den Rasterer**: der Symbolzug und ein Textzug
**mit** Symbolen. Wohin ein `<sym>` mitten im Regeltext kommt, weiß erst der
Umbruch; die Symbolmaschine ein zweites Mal zu bauen wäre die zweite Rechnung
über dieselbe Sache. Sie kommen als eingebettetes Bild an — dieselben
Bildpunkte wie im PNG, an derselben Stelle. Der Preis: für diese Ebene ist der
Text nicht markierbar.

**Zwei Abweichungen vom eigenen Rasterer, beide benannt:**

| | |
| --- | --- |
| **Schräge Winkel** | SVG dreht jeden exakt, `render.ebeneGedreht` nur die geraden. Das ist die Rückkehr zu MSE, das ebenfalls jeden dreht — **gemeldet** wird es trotzdem, denn dieselbe Karte sähe als PNG und als SVG verschieden aus. Im vollen Pack: viermal schräg gegen 142 gerade |
| **Die Grundlinie** | `dominant-baseline="text-before-edge"` statt einer gerechneten. Der Rasterer rastet die Grundlinie auf Viertelpunkte (MSEs `text_scaling = 4`), ein Betrachter mit derselben Schriftdatei tut es nicht — höchstens ein Viertelpunkt je Zeile |

**Bis zum 2026-08-22 stand `svg.AusPlan` da und niemand rief es.** Die Auswahl
in der Oberfläche bot vier Formate an, `service.Bildformat` kannte vier, und
das Paket hatte volle Testabdeckung. Ein Format, das niemand wählen kann, gibt
es für Anwendende nicht; der Zweitblick über Phase P hat es angestrichen
(#852). Ein
zweiter Wächter hält jetzt beide Listen gegeneinander
(`internal/werkzeuge/ausgabeformate_test.go`).

**Quelle:** MSE2 `src/util/rotation.cpp` 37–42, `src/render/value/viewer.cpp`
44–46, `src/gfx/mask_image.cpp` 41–44, `src/render/value/color.cpp` 47–50,
`src/gfx/generated_image.cpp` 49–59, `src/data/field/text.cpp` 73–81, Fassung
f9d9356 vom 2021-07-06; die Zählungen über `full-magic-pack`
**Erschlossen am:** 2026-08-22
**Sicherheit:** belegt — am Quelltext gelesen und am Pack gezählt
**Testfall:** TestJedeEbeneStehtImSVG, TestJedesFeldEinerEbeneStehtImSVGSchreiber,
TestJedesBildformatStehtInDerAuswahl

### Was die Ausfuhr kostet — je Format gemessen

Geschrieben und **zurückgelesen**, verglichen gegen den PNG-Lauf. Es sind
**zwei** Messungen, und der Unterschied zwischen ihnen ist die eigentliche
Auskunft:

| Vorlage | Format | Abstand | größte Abweichung | betroffene Fläche |
| --- | --- | --- | --- | --- |
| eine Halbfläche | TIFF, BMP | **0,000** | 0 | 0 % |
| eine Halbfläche | JPG (Güte 90) | 0,077 | 6 | 10 % |
| **eine Karte** | TIFF, BMP | **0,000** | 0 | 0 % |
| **eine Karte** | JPG (Güte 90) | **1,48** | **40** | **91 %** |

**Die untere Hälfte ist die, die zählt.** Bis zum 2026-08-22 stand hier nur die
obere, und ihre Vorlage war dreimal dieselbe einfarbige Halbfläche auf
durchsichtigem Grund — kein Verlauf, kein Rahmen, keine schräge Kante. Daran
kostet JPEG fast nichts, und eine an ihr geeichte Toleranz sagt über eine Karte
nichts (§10). Angestrichen vom Zweitblick über Phase P.

**Die Karte für die untere Hälfte** (`internal/service/bildgoldreich_test.go`)
trägt sechs Ebenen, jede für eine andere Art von Kante: eine volle Fläche, einen
weichen Verlauf, runde Ecken, einen Rahmen aus zwei Punkt breiten Bahnen, eine
um 30° gedrehte Fläche und eine halbdurchsichtige. **Ohne Text und ohne Bild**,
und das ist eine Grenze, die dazugehört: beides hinge an einer Systemschrift
oder einer Paketwurzel, und ein Goldbild, das den Rechner mitmisst, ist keines.

**Alle drei Zahlen sind zugesichert, nicht nur der Mittelwert.** Auch das kommt
aus dem Zweitblick: die Tabelle nannte drei, geprüft wurde eine. Ein kleiner,
örtlich stark beschädigter Bereich verschwindet im Mittelwert — genau davor
warnt `render.Vergleich.Innerhalb` an seiner eigenen Stelle.

**BMP und JPG werden dabei gegen ein auf Weiß gelegtes Goldbild gehalten** —
beide kennen keine Durchsichtigkeit, und ein Vergleich gegen ein durchsichtiges
Bild misst genau das statt des Formats.

**Die 40 bei JPG sind die Blockkante.** JPEG rechnet in Feldern von 8×8, und an
der Grenze zweier Farben weicht ein einzelner Bildpunkt weit ab. Die Gegenprobe
läuft **mit** (`TestGuete30BrichtJedeDerDreiSchwellen`): mit Güte 30 steigt der
Abstand auf **5,34**, die Spitze auf **112** und die Fläche auf **100 %** — jede
der drei Schwellen bricht, und ein Kommentar, der das behauptet, wäre kein
Beleg.

### Wie eine Bilddatei an ihren Platz kommt

**In eine Nachbardatei, dann umbenannt.** `os.Rename` ist innerhalb eines
Dateisystems atomar; bis dahin steht die alte Datei unverändert. Vorher öffnete
`os.WriteFile` mit `O_TRUNC` — eine vorhandene Datei war ab diesem Augenblick
leer, und scheiterte das Schreiben danach, blieb sie leer oder halb zurück. Ein
halbes PNG sieht im Dateibrowser aus wie ein ganzes
(#853).

**Was es kostet**: in einem nur lesbaren Verzeichnis mit vorhandener Zieldatei
scheitert die Ausfuhr jetzt, wo sie vorher überschrieben hätte — eine
Nachbardatei anzulegen braucht Schreibrecht am **Verzeichnis**, die vorhandene
Datei zu öffnen nur eines an der Datei.

**Was es nicht hält**: einen Absturz des Rechners zwischen Schreiben und
Umbenennen. Ohne `fsync` gibt diese Zusage niemand; liegen bleibt dann eine
Datei mit führendem Punkt neben dem Ziel — sichtbar, und nicht ein halbes Bild
unter dem richtigen Namen.

**Quelle:** eigene Entscheidung; dieselbe Bauform wie `storage.Sichern`
**Erschlossen am:** 2026-08-22
**Sicherheit:** belegt — die Gegenprobe mit `os.WriteFile` macht den Lauf rot
**Testfall:** TestInEinemNurLesbarenVerzeichnisBleibtDasZielUnberuehrt,
TestDieNachbardateiMeldetJedenIhrerFehler

### Und was eine **halbe** Deckung wird

| Format | Alpha 128 zurückgelesen |
| --- | --- |
| PNG, TIFF | Alpha bleibt 128 |
| BMP, JPG | gemischt auf Weiß: (200, 100, 50) wird (227, 177, 152) |

**Gemessen, nicht aus der Formatbeschreibung geschlossen** — und der erwartete
Wert kommt aus `aus = c · a/255 + 255 · (1 − a/255)`, nicht aus dem geprüften
Code. Bis zum 2026-08-22 belegte der Wächter nur Alpha 0 und Alpha 255; beide
Endpunkte sehen auch dann richtig aus, wenn dazwischen falsch gemischt wird, und
eine Ebene mit `Deckung: 0.5` ist auf einer Karte der Normalfall.

## Was von MSE-Script bewusst abweicht — und warum es folgenlos ist

**Vier Stellen.** Zwei haben dieselbe Gestalt — MSE ist strenger, dieses
Programm rechnet weiter —, die anderen beiden haben sie nicht: bei `format`
und bei der **Zahlklasse** ist es *dieses Programm*, das eine Zahl trägt, die
MSE gar nicht darstellen kann. Sie stehen hier an **einer** Stelle, weil eine
über den Quelltext verteilte Abweichung nicht zu überblicken ist
(#271).

### `to_int` nimmt eine breitere Zahlensyntax an als MSE

MSE benutzt `String::ToLong`, dieses Programm `strconv.ParseFloat` samt
`math.Trunc`. Am Orakel gefragt:

| Eingabe | MSE2 | hier |
| --- | --- | --- |
| `"1.5"` | Fehler | `1` |
| `"1e3"` | Fehler | `1000` |
| `" 12 "` | Fehler | `12` |
| `"1_000"` | Fehler | `1000` |
| **`"Inf"`** | Fehler | **`inf`** |
| `"0x10"`, `"12abc"` | Fehler | Fehler — gleich |
| `"+12"`, `"-3"` | `12`, `−3` | gleich |

**Und `5 / 2` unterscheidet sich nicht** — beide geben `2.5`. Die Abweichung
sitzt allein an der Umwandlung einer Zeichenkette, nicht an der Rechnung.

**Gemessen folgenlos**: über die 376 Stile und 8.915 Feldstile des Packs
bekommt `to_int` **kein einziges Mal** etwas, das dieses Programm annimmt und
MSE ablehnt — über die **ganze** Klasse gezählt, nicht nur über `"1.5"`.

**`"Inf"` bleibt trotzdem der unangenehmste Fall**: eine Unendlichkeit in einer
Maßangabe macht jede daraus gerechnete Stelle unbrauchbar. Sie tritt nicht ein;
wer die Stelle anfasst, weist sie MSE-treu ab.

**Quelle:** MSE2 `src/script/functions/basic.cpp`, Zeilen 116–124, Fassung
f9d9356 vom 2021-07-06; die Ergebnisse am Orakel (`magicseteditor --cli … -q
--raw`) gegen `White.mse-set`
**Erschlossen am:** 2026-08-20
**Sicherheit:** belegt — am laufenden MSE2 gefragt und am Pack gezählt
**Testfall:** TestToIntNimmtEineBreitereZahlensyntaxAn

### `or else` rechnet die rechte Seite nur bei Bedarf

MSE wertet **beide** Seiten aus, bevor `I_OR_ELSE` läuft
(`src/script/context.cpp`, Zeilen 535–537); dieses Programm kürzt ab, sobald
die linke Seite kein Fehlerwert ist. Sichtbar wird das nur, wenn die rechte
Seite **wirft**:

```
1 or else gibtsnicht    MSE2: Fehler — Variable not set    hier: 1
1 or else 2             beide: 1
nil or else 2           beide: nichts — `nil` ist kein Fehler
[a: 1].b or else 2      beide: 2
```

**Gemessen folgenlos**: über denselben Packlauf wird die rechte Seite
**370.807-mal** übersprungen. Wertet man sie trotzdem aus, ergibt sie
**0-mal** einen Abbruch und 905-mal einen **Fehlerwert** — und ein Fehlerwert
ist bei MSE ebenfalls ein Wert und kein Abbruch (`delay_error`,
`src/script/functions/basic.cpp`, Zeile 123). Die 905 sind damit auf beiden
Seiten harmlos.

**Hier stand bis zum 2026-08-20 „kein einziges Mal ein Befund"** — über 190
Stile und 219.806 Übersprünge, und ohne die Unterscheidung zwischen Abbruch
und Fehlerwert. Die Aussage bleibt richtig; ihre Grundlage ist größer und ihre
Trennung schärfer.

**Quelle:** MSE2 `src/script/context.cpp`, Zeilen 535–537, und
`src/script/functions/basic.cpp`, Zeilen 116–124, Fassung f9d9356 vom
2021-07-06; die Ergebnisse am Orakel
**Erschlossen am:** 2026-08-20
**Sicherheit:** belegt — am laufenden MSE2 gefragt und am Pack gezählt
**Testfall:** TestOrElseKuerztAbUndDasIstFolgenlos

### `format` rechnet in 32 Bit — bis 2⁵³, und darüber nicht mehr

MSE übergibt `input.toInt()` an `String::Format`, und `ScriptInt` hält einen
C++-`int`; ein geparster `long` wird beim Anlegen darauf verengt. **Dieses
Programm führt jede Zahl als `float64`** — genau ist sie damit bis 2⁵³.

Bis dorthin ist die Verengung nachgebaut und gleich; darüber weicht sie ab,
weil die unteren Bits schon vor dem Formatieren fehlen:

| Ausdruck | MSE2 | hier |
| --- | --- | --- |
| `%x` mit −42 | `ffffffd6` | gleich |
| `%X` mit −42 | `FFFFFFD6` | gleich |
| `%o` mit −42 | `37777777726` | gleich |
| `%x` mit −1 | `ffffffff` | gleich |
| `%d` mit 2147483648 | `-2147483648` | gleich |
| `%d` mit 4294967295 | `-1` | gleich |
| **`%d` mit 9223372036854775807** | **`-1`** | **`0`** |

**Die letzte Zeile ist die Abweichung**, und sie gehört dem **Werttyp**, nicht
der Formatierung: 9223372036854775807 kommt hier bereits als
9223372036854775808 an. Ein Sonderfall für diese eine Zahl behandelte den Fall
und nicht die Klasse (§10); die Klasse heißt „ganze Zahlen jenseits von 2⁵³",
und die trägt dieses Programm nicht.

**Und drei Buchstaben werden ausdrücklich nicht nachgebaut.** `D`, `I` und `O`
gehören bei MSE zur Ganzzahlklasse — `find_first_of("DdIiOoXx")` kennt sie —,
aber was die C-Bibliothek daraus macht, ist die ABI dieses Rechners:

| Ausdruck | MSE2 auf arm64 | hier |
| --- | --- | --- |
| `%D` mit −42 | `4294967254` | `-42` |
| `%I` mit −42 | `I` | `-42` |
| `%O` mit −42 | `37777777726` | gleich |

`%D` ist auf dieser Maschine `%ld` über einen nullerweiterten `int`, `%I` ein
Flag ohne Umwandlung dahinter. **Die Reichweite einer Messung ist die
Reichweite ihrer Aussage** (§10): auf x86-64 stünden im oberen Wort andere
Bits, und eine Messung auf einem Rechner belegt keine Sprachregel. Portiert ist
deshalb die **Klassenwahl**, nicht die Ausgabe.

**Gemessen folgenlos**: über das volle Pack gibt es **null** Fundstellen für
`format(` und sieben für `to_string(…, format:)`, alle mit einem Zeitformat.
Keine davon trägt eine Zahl jenseits von 2⁵³ oder einen der drei Buchstaben.

**Quelle:** MSE2 `src/script/functions/basic.cpp`, Zeilen 69–86 und 367–371,
`src/script/to_value.hpp`, Zeilen 390–392, Fassung f9d9356 vom 2021-07-06; jede
Zahl einzeln am Orakel (`magicseteditor --cli … -q --raw`) gegen
`Archenemy.mse-set` gefragt
**Erschlossen am:** 2026-08-22
**Sicherheit:** belegt — am Quelltext gelesen und am laufenden MSE2 gefragt
**Testfall:** TestFormatRechnetInZweiunddreissigBit,
TestFormatKenntDenCBuchstabenI,
TestJenseitsVonZweiHochDreiundfuenfzigWeichtEsAb,
TestDreiFormatbuchstabenSindPlattformzufall

### MSE-Script kennt zwei Zahltypen, dieses Programm einen

`ScriptInt` hält einen C++-`int`, `ScriptDouble` ein `double`; der Zerteiler
entscheidet beim Lesen eines Literals, und ein geparster `long` wird auf `int`
verengt (`src/script/parser.cpp:697`, `src/script/to_value.hpp:390–392`).
`internal/msescript` kennt **einen**: `ArtZahlwert` mit `Zahl float64`.

Am Orakel gemessen, jeder Ausdruck einzeln:

| Ausdruck | MSE2 | hier |
| --- | --- | --- |
| **`2147483647 + 1`** | **`-2147483648`** | `2147483648` |
| `to_number("2147483648")` | `-2147483648` | gleich |
| `to_number("9223372036854775807")` | `-1` | gleich |
| `to_number("1.5")` | `1.5` | gleich |
| `to_number("1e3")` | `1000` | gleich |
| `to_number("12abc")` | Fehler | gleich |

**Die erste Zeile ist die einzige, die bleibt** — und sie hängt nicht an einer
Funktion, sondern an der Arithmetik: MSE rechnet ganze Zahlen in 32 Bit und
läuft um.

**Bis zum 2026-08-22 stand `to_number("1.5")` auf 1.** Es reichte an `to_int`
durch und schnitt ab. Behoben in
#859; aus
einer Zeichenkette wird jetzt exakt verengt, und deshalb trifft auch die dritte
Zeile — `format` kann das nicht, dort kommt die Zahl schon als `float64` an.

**Gemessen folgenlos, und die Messung steht dabei**: über **406**
Skriptdateien des vollen Packs — jede `style`, jede `game`, jede
`export-template` — gibt es **ein** Zahlliteral ≥ 2 147 483 648, und das ist
eine **Steam-URL in einem Kommentar**. Die Vorlagen rechnen in Kartenpunkten;
zwischen 0 und 750 berührt keine Rechnung die Grenze.

**Deshalb wird der Werttyp nicht geteilt.** Das wäre eine Änderung quer durch
den Interpreter — Literale, `to_int`, `to_number`, Arithmetik, Vergleich,
Modulo, Indizierung —, deren Nutzen an einer Zeile hinge, die in keiner Vorlage
vorkommt. *Auch eine Ablehnung braucht Deckung* (§10), und das ist ihre.

**Was sie kostet, wenn sie falsch ist**: eine künftige Vorlage, die absichtlich
über 2³¹ rechnet, bekäme hier ein anderes Ergebnis — und zwar **stillschweigend**.

**Quelle:** MSE2 `src/script/functions/basic.cpp`, Zeilen 106–190,
`src/script/parser.cpp`, Zeile 697, `src/script/to_value.hpp`, Zeilen 390–392,
`src/script/context.cpp`, Zeilen 425, 474 und 495, Fassung f9d9356 vom
2021-07-06; die Ergebnisse am Orakel gegen `Archenemy.mse-set`
**Erschlossen am:** 2026-08-22
**Sicherheit:** belegt — am Quelltext gelesen, am laufenden MSE2 gefragt und am
vollen Pack gezählt
**Testfall:** TestToNumberAntwortetWieDasOrakel, TestToNumberWeistNichtzahlenAb

### Und vier, die keine mehr sind

| # | Was einmal abwich | Stand |
| --- | --- | --- |
| 3 | `+` auf zwei Abbildungen vereinigte, statt zu verketten | **behoben** am 2026-08-18 — `abbildungsplus.go` verkettet, am Orakel entschieden |
| 4 | Gleichheit zweier Abbildungen verlangte gleiche Namen | **tritt nicht ein** — im ganzen Pack **ein** Vergleich, und dort sind beide Seiten Listen |
| 5 | `f == f` gab immer falsch | **tritt nicht ein** — null Fundstellen im Pack |
| 6 | Eine verkettete Funktion sah eine frische Umgebung | **behoben** am 2026-08-17 — `ruferufen` reicht die Umgebung des Aufrufers durch (`ScriptCompose::eval`, `src/script/context.cpp`, Zeilen 383–400) |

**Quelle:** MSE2 `src/script/value.cpp`, Zeile 107, `src/script/context.cpp`,
Zeilen 383–400 und 465, Fassung f9d9356 vom 2021-07-06; dazu die Auszählungen
über `full-magic-pack`
**Erschlossen am:** 2026-08-20
**Sicherheit:** belegt — am Quelltext gelesen und am Pack gezählt
**Testfall:** TestZweiListenWerdenVerkettet

## MTG JSON: mehrseitige Karten und der Umschlag

**Quelle:** `https://mtgjson.com/data-models/card/card-set/` und
`https://mtgjson.com/data-models/meta/` — die Datenmodelle von MTGJSON selbst,
samt TypeScript-Modell und Feldbeschreibungen
**Erschlossen am:** 2026-08-24
**Sicherheit:** belegt
**Testfall:** TestEineKartenseiteBekommtIhrenEigenenNamen, TestDieFassungWirdGelesen

### `name` trägt bei mehrseitigen Karten beide Namen

> `name` — The name of the card. Cards with multiple faces are given a `//`
> delimiter. Example: „Wear // Tear"
>
> `faceName` — The name on the face of the card.

MTGJSON legt **je Seite ein eigenes Kartenobjekt** an; beide tragen in `name`
denselben zusammengesetzten Namen und unterscheiden sich in `faceName`, `side`
(`"a"`/`"b"`) und `otherFaceIds`.

### MSE führt eine mehrseitige Karte als **eine** Karte

**Zwei unabhängige Belege**, beide am 2026-08-24:

- `magic.mse-game/card_fields` definiert **249** Kartenfelder, davon **46 mit
  dem Zusatz ` 2`** — `name 2`, `casting cost 2`, `rule text 2`,
  `super type 2`, `power 2` und so fort. **Keines** trägt eine ` 3`.
- `magic-set-generator` (MIT, §11) schreibt es genauso: sein Makro `push_alt!`
  in `src/mse.rs` hängt bei der zweiten Seite ` 2` an jeden Schlüssel und legt
  **beide** Seiten in dieselbe Karte.

**Was dieses Programm daraus macht**: die Kartenobjekte einer Karte werden über
`otherFaceIds` zusammengeführt, die Reihenfolge kommt aus `side`. Die
Vorderseite füllt `name`, `casting cost`, …, die Rückseite `name 2`,
`casting cost 2`, … `faceName` gewinnt über `name`; eine gewöhnliche Karte hat
kein `faceName` und behält ihren Namen.

**Eine dritte Seite hat keinen Platz** und wird als Befund gemeldet — es gibt
kein Feld mit ` 3`.

**Der Rückweg trennt wieder**: aus einer MSE-Karte mit `name 2` werden zwei
Kartenobjekte, verbunden über `otherFaceIds`, mit `faceName`, `side` und dem
zusammengesetzten `name` in beiden. Die Kennungen werden aus dem Namen
**gebildet** und nicht gewürfelt — sonst schriebe jeder Lauf eine andere Datei.

**Hier stand bis zum 2026-08-24 das Gegenteil**: „Der Rückweg schreibt `name`,
denn eine MSE-Karte *ist* eine Seite." Das war falsch, und es hat den Import
mehrseitiger Karten zwei Tage lang zwei Karten erzeugen lassen, die keine
Vorlage als Vorder- und Rückseite zeichnen kann. Angestrichen vom Zweitblick
über Phase V.

### `meta` nennt Fassung und Baudatum

> `date` — The current release date in ISO 8601 format for the MTGJSON build.
>
> `version` — The current SemVer version for the MTGJSON build appended with
> the build date.

Beide dürfen fehlen: eine von Hand geschriebene Datei hat den Umschlag nicht.
Eine **unbekannte Hauptfassung** wird als Befund gemeldet und nicht abgewiesen —
was sich geändert hat, zeigt sich an den übrigen Befunden.
