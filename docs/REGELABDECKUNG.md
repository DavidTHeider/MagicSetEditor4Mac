# Regelabdeckung — was von den Comprehensive Rules heute wirklich gilt

**Stand: 2026-09-16.** Diese Datei geht **jedes Kapitel** der *Magic: The
Gathering Comprehensive Rules* durch und sagt je Kapitel, was dieses Programm
davon durchsetzt, woran das zu sehen ist und welches Werkstück den Rest baut.

Sie beantwortet die Frage, die `docs/FAEHIGKEITEN.md` für die Flächen
beantwortet und `docs/SCHREIBWEGE.md` für die Schreibwege: **worüber reden wir
eigentlich, wenn wir sagen, das Programm setze Regeln durch?**

**Sie ist nicht dasselbe wie die Regelmatrix.** Die Matrix
(`internal/domain/regelmatrix.go`) hat **dreizehn** Bereiche und sagt je Bereich
einen von vier Graden. Diese Datei hat **147** Kapitel und sagt je Kapitel,
**welche Zeile** den Grad trägt. Die Matrix ist die Zusage an die Spielenden;
diese Datei ist ihre Deckung.

---

## 1 Die Quelle, und was von ihr hier steht

**Comprehensive Rules, Fassung 2026-09-25** (`MagicCompRules 20260925.txt`,
in Kraft seit dem 2026-09-25), abgerufen von `magic.wizards.com/en/rules`.
**Bis zum 2026-09-25 stand hier die Fassung 2026-08-19**
(`MagicCompRules 20260819.txt`, in Kraft seit dem 2026-08-07); der Abgleich
steht gleich darunter (#1842).

### Der Abgleich mit der Fassung 2026-09-25

**Beide Fassungen nebeneinander gelesen, Regel für Regel**, am 2026-09-25:
jede Nummer der alten gegen dieselbe Nummer der neuen, und was nur in einer
steht. **Sechs Regeln haben sich geändert, elf sind neu, sieben fort** — in
**fünf** von 147 Kapiteln. **Keine davon trifft eine umgesetzte Regel**;
Befunde gibt es deshalb keine, nur Zeilen:

| Kapitel | Was sich geändert hat | Trifft den Kern? |
| --- | --- | --- |
| **CR 111** (Tokens) | ein **neuer vordefinierter Spielstein** (111.10x) | nein — 111.10 ist offen |
| **CR 205** (Type Line) | ein **Artefakttyp mehr** in 205.3g; 205.4c und 205.3g sonst nur im Umbruch | nein — 205.3b–m sind offen, und der Teil von 205.4c, den `standardland.go` trägt, ist gleich geblieben |
| **CR 506** (Combat Phase) | eine **neue Begriffsbestimmung** als 506.6 (allein angreifen); die bisherigen 506.6 und 506.7 samt 506.7a–g **heißen jetzt 506.7 und 506.8 samt 506.8a–g** | nein — zitiert werden 506.1, 506.2, 506.4 und 506.5, und die sind gleich; **wer künftig 506.6 bis 506.8 zitiert, zitiert die neue Zählung** |
| **CR 509** (Declare Blockers Step) | 509.1b nur im Umbruch | nein — 509.1b ist offen |
| **CR 701** (Keyword Actions) | eine **siebzigste Handlung** (701.71); 701.67b bekommt ein Beispiel | nein — umgesetzt sind zwölf andere; aus „zwölf von neunundsechzig" werden **zwölf von siebzig** |

**Was dieser Abgleich nicht ist**: eine Prüfung des Glossars und der
Einleitung — dort steht keine Regel, auf die sich der Kern beruft.

**Das Grundprofil nennt weiter die Fassung 2026-08-19**
(`domain.Grundprofilquelle`): gegen sie ist seine Matrix geschrieben, und
nach diesem Abgleich stimmt jede seiner Grenzen auch in der neuen — mit einer
Zahl, die dort um eins größer wäre („zehn der neunundsechzig Handlungen" aus
CR 701). **Die Quelle wechselt mit der nächsten Fassung des Grundprofils**,
denn eine Partie friert die Nummer ein und nicht den Text; eine neue Quelle
unter derselben Nummer wäre eine Zusage, die niemand auseinanderhalten kann.
**Die Zitate im Kern bleiben**, wo sie stehen: sie nennen die Fassung, die
gelesen wurde, und für jedes davon ist die Regel in beiden gleich.

**Kein Wortlaut, keine Kopie** (§11). Die CR ist urheberrechtlich geschützt und
hat keine Weitergabelizenz. Was hier steht, ist die **Nummer** samt Fassung und
eine **eigene** kurze Beschreibung. Das Dokument selbst kommt nicht in den Baum
— es wächst mit jedem Set, und eine eingecheckte Kopie wäre am Tag der nächsten
Freigabe falsch.

**Jede Nummer in dieser Datei ist in der Datei nachgesehen worden.** Das ist
kein Höflichkeitssatz: beim Anlegen von
#1592 standen
drei erfundene Nummern im Text (118 statt 119 für Leben, 119.3 statt 120.3 für
Schaden, 122.1d statt 122.1f für Vergiftungsmarken), und alle drei sahen aus
wie ein Beleg.

---

## 2 Wie gezählt wurde, und wie weit die Zahl reicht

**Die Reichweite einer Messung ist die Reichweite ihrer Aussage** (§10).
Deshalb steht hier das Verfahren und nicht nur das Ergebnis.

**Die Einheit ist das Kapitel**, also `100`, `101`, … `905`. Die CR führt
**147** davon. Ein Kapitel ist eine **grobe** Einheit: CR 702 enthält 771
nummerierte Regeln, CR 203 genau eine. Wer aus „ein Kapitel offen" auf „ein
gleich großes Stück Arbeit" schlösse, hätte die Zahl überdehnt — deshalb steht
in Abschnitt 4 neben jedem Abschnitt, wie viele Regeln darin liegen.

**Gezählt wurde über den Rumpf der Textfassung**, also ab der Überschrift
`1. Game Concepts` bis vor das Glossar. Eine Zeile gilt als
**Hauptregel**, wenn sie mit `NNN.M.` beginnt, und als **Unterregel**, wenn sie
mit `NNN.Ma` beginnt. Ergebnis: **1.175 Hauptregeln, 1.991 Unterregeln, 3.166
nummerierte Regeln** in der Fassung 2026-09-25 (6.925 Zeilen); in der Fassung
2026-08-19 waren es 1.173, 1.989 und 3.162.

> **Zwei Tippfehler stehen im Quelldokument selbst**, und sie sind der Grund,
> warum hier 3.162 steht und nicht 3.160: `119.1d.` trägt einen Punkt zu viel
> (eine Unterregel in der Form einer Hauptregel), und `606.5` trägt keinen. Ein
> strenges Muster übersieht beide still. Gefunden wurden sie, indem jede Zeile,
> die wie eine Regel aussieht und in keiner der beiden Gruppen landete, noch
> einmal angesehen wurde — die Gegenprobe zur Zählung und nicht die Zählung
> selbst.

**Was diese Datei nicht prüft:** ob eine Regel, die als `umgesetzt` oder
`teilweise` dasteht, **richtig** umgesetzt ist. Der Beleg sagt, **wo** es
geschieht; ob es dort stimmt, sagen die Tests an dieser Stelle. Diese Datei
findet die Stellen, an denen gar nichts steht — und die Stellen, an denen die
Zusage weiter reicht als der Code.

**Jeder Beleg ist gelesen worden.** Ein Feld, das existiert, ist keine Regel,
die greift: `domain.Regelprofil.Startleben` steht seit G4 da und **kein
produktiver Weg liest es** (Befund 3 unten). Solche Fälle stehen als
`teilweise` mit ausdrücklicher Grenze, nicht als `umgesetzt`.

---

## 3 Das Ergebnis in vier Zahlen

| Stand | Kapitel | Anteil |
| --- | ---: | ---: |
| **umgesetzt** | **0** | 0 % |
| **teilweise** | **83** | 56 % |
| **offen** | **46** | 31 % |
| **trifft nicht zu** | **18** | 12 % |
| **Summe** | **147** | 100 % |

**Die erste Zeile ist die Aussage dieser Datei.** Es gibt **kein einziges
Kapitel der Comprehensive Rules, das dieses Programm vollständig trägt** — auch
keines der drei, die das Grundprofil als `automatisch` führt. CR 122 (Marken)
scheitert an 122.2, CR 400 (Zonen) an 400.7, CR 301/303 an der
Frage, ob etwas dort hängen darf.

**Die zweite Zeile hat sich am 2026-09-17 um fünfzehn Kapitel bewegt** — G9
(#1592) hat
die Zugstruktur samt Priorität gebaut, `M17`
(#1618) das
Spielende samt den drei zustandsbasierten Aktionen, die einen Sitz kosten. Neun
Kapitel aus Abschnitt 5, vier aus Abschnitt 1, je eines aus 7 und 8.

**Und am selben Tag um vier weitere** — `M2`
(#1604)
hat aus Typzeile, Manakosten, Farbe und Werten **Daten** gemacht: CR 105 und
CR 208 kamen dazu, dazu CR 306 und CR 310 mit ihren Zahlen, und CR 107, 109,
122, 202 und 205 tragen deutlich mehr als vorher.

**Und um zwei weitere** — `M1`
(#1603)
hat den **Stapel** gebaut: CR 608 (das Verrechnen), CR 724 (den Zug vorzeitig
beenden) und CR 732 (Abkürzungen) kamen dazu, CR 405, 117 und 500 tragen mehr.

**Und die Null ist zum Teil ein Artefakt der Körnung.** Ein Kapitel ist erst
`umgesetzt`, wenn **jede** seiner Regeln trägt, und CR 400 hat 29. Wer feiner
zählte, käme auf ein freundlicheres Bild. Das ist kein Grund, gröber zu
schreiben: die Zusage, die das Programm heute an der Oberfläche macht, lautet
„**zehn von dreizehn** Bereichen automatisch" — und **kein Kapitel
vollständig** ist die ehrlichere Übersetzung davon.

### Je Abschnitt

| Abschnitt | Kapitel | Regeln | umgesetzt | teilweise | offen | trifft nicht zu |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| 1 Game Concepts (100–123) | 24 | 505 | 0 | 23 | 1 | 0 |
| 2 Parts of a Card (200–213) | 14 | 120 | 0 | 6 | 3 | 5 |
| 3 Card Types (300–315) | 16 | 164 | 0 | 6 | 5 | 5 |
| 4 Zones (400–408) | 9 | 78 | 0 | 8 | 0 | 1 |
| 5 Turn Structure (500–514) | 15 | 148 | 0 | 14 | 1 | 0 |
| 6 Spells, Abilities, Effects (600–616) | 17 | 320 | 0 | 9 | 7 | 1 |
| 7 Additional Rules (700–733) | 34 | 1.525 | 0 | 14 | 19 | 1 |
| 8 Multiplayer (800–811) | 12 | 171 | 0 | 2 | 10 | 0 |
| 9 Casual Variants (900–905) | 6 | 131 | 0 | 1 | 0 | 5 |
| **Summe** | **147** | **3.162** | **0** | **83** | **46** | **18** |

> **Diese Zeilen sind gezählt und nicht fortgeschrieben.** Sie standen bis zum
> 2026-09-18 auf dem Stand vor Phase 5.5 und summierten sich auf 52 und 77,
> während die Bilanz oben 83 und 46 nannte — **zwei Statistiken über dasselbe
> Dokument, und keine sagte, welche gilt**. Angestrichen vom Zweitblick über
> Phase 5.5 (§10).
>
> Gezählt wird über die Kapitelzeilen dieser Datei selbst; wer sie ändert,
> zählt wieder.

**Abschnitt 7 trägt fast die Hälfte aller Regeln** (1.525 von 3.162) und
besteht zu **siebzig Prozent aus zwei Kapiteln**: CR 701 (294 Regeln) und
CR 702 (771) machen zusammen 1.065 davon aus. Wer den Aufwand dieser Phase
schätzt, schätzt im Wesentlichen diese beiden.

---

## 4 Die Kapitel

Lesehilfe für die Spalte **Werkstück**: `M1`…`M15` und `M17`…`M20` sind die
Arbeitspakete aus
#1593, `G9`
ist das Werkstück aus
#1592.
**„— fehlt"** heißt: kein Werkstück nennt dieses Kapitel, weder mit Nummer noch
dem Gegenstand nach. `(M3)` in Klammern heißt: das Werkstück meint den
Gegenstand, nennt die Nummer aber nicht — das ist keine Lücke, aber es lässt
sich nicht nachschlagen.

**`M16` fehlt in der Reihe, und das ist Absicht.** Es trug CR 733 — zwei
Regeln — und ist in `M4` aufgegangen; die Kennung wird nicht weitervergeben.

**Diese Spalte ist am 2026-09-16 nachgezogen worden**, nachdem der Zuschnitt
die Regelmaschine (damals Phase 5.8, seit dem 2026-09-17 Teil von 5.5) auf diese Datei hin neu geschnitten wurde: vierundvierzig Zeilen
tragen seither ein anderes Werkstück als beim Erheben. **Der Rest der Datei ist
der Stand von vorher** — Stände, Belege und Grenzen sind gegen den Quelltext
erhoben und ändern sich nicht dadurch, dass jemand einen Plan umschreibt.

### 4.1 Game Concepts — CR 100–123

| Regel | Stand | Beleg | Grenze | Werkstück |
| --- | --- | --- | --- | --- |
| **CR 100 — General** | teilweise | `internal/domain/format.go:175`, `internal/domain/deckurteil.go:165` | Nur die **Deckbauzahlen** (100.2a/b, 100.4a/b) stehen. **100.1 weicht ab und zwar bewusst**: die CR gilt ab zwei Spielern, `internal/domain/partie.go:210` lässt einen Sitz zu (Goldfish, G3) — begründet an Ort und Stelle. 100.3 (Münzen und Würfel) offen | M15 · M13 |
| **CR 101 — The Magic Golden Rules** | teilweise | `internal/domain/ausloeserstapel.go` (101.4) | **101.4 trägt seit `M9`**: warten mehrere ausgelöste Fähigkeiten, legt der aktive Spieler zuerst, dann die anderen in Zugreihenfolge — dieselbe Reihenfolge, die CR 603.3b verlangt. **101.1 (die Karte schlägt die Regel), 101.2 (Unmögliches wird ignoriert) und 101.3 („kann nicht" schlägt „kann") nennt niemand** — und sie sind die Grundlage jeder Auslegung | M9 · M8 (101.1–101.3) |
| **CR 102 — Players** | teilweise | `internal/domain/prioritaet.go:20` (der aktive Spieler, 102.1), `:193` (die Reihenfolge am Tisch, 102.1) | **Ein Spieler ist ein Sitz mit Leben, Marken und einer Hand.** 102.2 (Mitspieler und Gegner) und 102.3 (Teams) stehen nicht — Zweiergruppen gibt es nicht | G9 · M13 |
| **CR 103 — Starting the Game** | teilweise | `internal/service/partieaufbau.go` (das Startleben wird gesetzt, 103.4), `internal/domain/prioritaet.go:67` (wer anfängt, 103.1), `internal/domain/eroeffnung.go` (Mischen, Starthand und Mulligan, 103.3, 103.5, 103.5c), `internal/service/eroeffnungsbefehle.go` (die vier Befehle), `internal/domain/zughandlungen.go:84` (103.8a) | **103.3 und 103.5 tragen seit #1839** (nachgelesen in der Fassung 2026-09-25): die Bibliothek wird gemischt und die Starthand **gezogen**, Karte für Karte (121.2) — reicht die Bibliothek nicht, ist der Rest ein Griff ins Leere (104.3c); ein Mulligan mischt die Hand zurück, zieht neu und verlangt so viele Karten unter die Bibliothek, wie Mulligans genommen sind, zu mehreren den ersten nicht (103.5c), und nie mehr, als auf der Hand liegt (101.3); nach der leeren Starthand ist Schluss. **103.8 trägt seit #1851** (Fassung 2026-09-25, am 2026-09-26 nachgelesen: der erste Zug folgt auf die Schritte aus 103.1 bis 103.7; `internal/domain/beginn.go`): **vom Aufbau an** nimmt der Sandkasten bis zum letzten Behalten nur den Beginn an, dazu Aufgeben (104.3a) und die Würfe aus 103.1 — auch wenn noch niemand eine Starthand gezogen hat. Bis dahin galt das erst ab der ersten (Befund 2 des Zweitblicks über Phase 6.9), und eine Partie, in der niemand anfing, spielte ohne Beginn. **Offen**: 103.1 als Wahl (der Wirt fängt an), die **Reihenfolge der Erklärungen** aus 103.5 (jeder Sitz nimmt seinen Mulligan für sich), 103.5c für **Brawl** zu zweit (die Partie kennt ihr Format nicht), 103.6–103.7 | G9 · M4 · #1839 · #1851 |
| **CR 104 — Ending the Game** | teilweise | `internal/domain/spielende.go:58` (104.1, 104.2a, 104.4a), `:27` (die vier Gründe, 104.3a–d), `internal/domain/zustandsaktionen.go:47` (704.5a/b/c als Erzeuger) | **Vier Gründe von zehn**: Aufgabe (104.3a), Leben (104.3b), leere Bibliothek (104.3c), Gift (104.3d). **104.2b–e** (Gewinnen durch Karteneffekt), **104.3e–j** und **104.4b–j** brauchen Effekte (`M10`, `M11`). **104.6 (Neustart) fehlt** — CR 727 hängt an *Karn Liberated* und damit an Fähigkeiten (`M8`) | M17 · M4 |
| **CR 105 — Colors** | teilweise | `internal/domain/kartenfarbe.go:22` (die fünf, 105.1), `internal/domain/manakostenwert.go:97` (105.2a–c über CR 202.2) | **105.3 fehlt**: Effekte, die eine Farbe ändern oder geben, brauchen Schichten (CR 613, `M10`). 105.4 (eine Farbe wählen) hat keinen Aufrufer, weil niemand wählt | M2 · M10 |
| **CR 106 — Mana** | teilweise | `internal/domain/manavorrat.go` (106.1b, 106.4), `internal/domain/manaabschnitt.go` (das Leeren), `internal/domain/manazahlung.go` (Ausgeben) | **Sechs Manaarten** (106.1b), **ein Vorrat je Spieler** (106.4), und er **leert sich am Ende jedes Abschnitts** (mit CR 500.5). **Offen**: 106.2/106.3 (wer Mana erzeugt — braucht `M8`), 106.5 (unbezahlbares Mana), 106.6–106.12. **106.4a/b (den übrigen Vorrat ansagen) ist eine Anzeige und keine Prüfung** — er steht an jeder Fläche, dass ihn jemand ansagt, prüft niemand | M3 · M8 |
| **CR 107 — Numbers and Symbols** | teilweise | `internal/domain/manasymbol.go:88` (107.4a–h), `:143` (107.3b) | **Gedeutet werden jetzt alle Manasymbole**: farbig, generisch, `{X}`, farblos, hybrid — auch der einfarbige `{2/B}` —, phyrexianisch und Schnee. Hier stand bis zum 2026-09-17 „gedeutet wird keines". **Offen**: 107.1/107.2 (Zahlen im Regeltext), 107.5–107.12 (Tap, Untap, Loyalität als Symbol), 107.13 (Farbanzeiger — MSE hat kein Feld dafür), 107.14 (Energie), 107.16 (Klassenstufen) | M2 · M8 |
| **CR 108 — Cards** | teilweise | `internal/domain/spielobjekt.go:142`, `internal/service/partieaufbau.go:244`, `internal/domain/objekteigenschaften.go:33` | **108.3 (der Eigner) steht am Objekt** und wird beim Aufbau gesetzt; über ihn findet ein Objekt seine Karte im **eingefrorenen** Abzug. 108.1 (Oracle), 108.2, 108.4–108.6 (außerhalb des Spiels) offen | M2 · M12 |
| **CR 109 — Objects** | teilweise | `internal/domain/spielobjekt.go:131`, `internal/domain/karteneigenschaften.go:58` (109.3), `internal/domain/objekteigenschaften.go:33` | **Zehn der zwölf Eigenschaften aus 109.3 stehen**: Name, Manakosten, Farbe, Kartentyp, Untertyp, Obertyp, Stärke, Widerstandskraft, Loyalität, Verteidigung. **Regeltext und Fähigkeiten fehlen** — ein Text ist eine Zeichenkette, solange ihn niemand ausführt (`M8`). Hier stand bis zum 2026-09-17 „keine davon". 109.4 (nur Stapel und Tisch haben einen Beherrscher) und 109.5 offen | M2 · M8 |
| **CR 110 — Permanents** | teilweise | `internal/domain/spielobjekt.go:156`, `internal/service/sandkasten.go:176`, `internal/domain/kartengestalten.go` (110.2 über `Beherrschen`) | **Zwei der vier Statusarten** aus 110.5 sind da: getappt und verdeckt. Kein „geflippt", kein Phasing. **Der Kontrollwechsel trägt seit `M12`**: `objekt.beherrschen` wechselt den Beherrscher (110.2, mit CR 613.1b als Schicht 2), **ohne** zu bewegen und **ohne** zu enttappen, nimmt das Permanent aus dem Kampf (506.4) und stempelt den Zug neu (302.6). *Hier stand `110.2a` — die spricht davon, unter wessen Kontrolle etwas ins Spiel **kommt**.* „Bleibende Karte" ist kein Begriff des Kerns: ein Objekt auf dem Tisch ist ein Objekt in einer Zone | M12 |
| **CR 111 — Tokens** | teilweise | `internal/domain/kartengestalten.go` (111.1, 111.2 über `Spielstein`), `internal/domain/zustandsobjekte.go:76` (111.7 über CR 704.5d) | **Ein Spielstein entsteht seit `M12`** (111.1, 701.7a): `spieler.spielstein` legt ihn mit **genannten** Eigenschaften auf den Tisch, unter die Kontrolle dessen, der ihn erzeugt (111.2), und mit dem laufenden Zug (302.6). Verlässt er das Schlachtfeld, hört er auf zu existieren (111.7 über 704.5d, seit `M5`). **Was fehlt**: 111.3–111.6 (was ein Effekt über einen Spielstein sagen darf) und die Spielsteinkopie (`M18`) | M12 · M5 · M18 |
| **CR 112 — Spells** | teilweise | `internal/domain/wirken.go` (112.1, 112.2) | **Eine Karte auf dem Stapel ist ein Zauber** (112.1), und wer ihn ankündigt, beherrscht ihn (112.2a). **Offen**: 112.3 (die vier Arten von Fähigkeiten), 112.4–112.8 — sie setzen einen Kartentext voraus | M4 · M8 |
| **CR 113 — Abilities** | teilweise | `internal/service/kartenskript.go` (113.3a, 113.7, 113.8) | **Eine der vier Arten aus 113.3 trägt**: die **Zauberfähigkeit** (113.3a) — sie geschieht beim Verrechnen und existiert nicht als eigenes Objekt. Das Skript sieht, wem sie gehört (113.8, `you`), woher sie kommt (113.7, `this`) und worauf sie zielt (115.1, `target`). **Offen: aktivierte (CR 602), ausgelöste aus einem Kartentext (CR 603.1) und statische (CR 604)** — die drei brauchen einen Ort, an dem eine Fähigkeit **steht**, statt einmalig zu geschehen | M8 |
| **CR 114 — Emblems** | teilweise | `internal/domain/kartengestalten.go` (114.1, 114.2 über `Emblem`) | **Ein Emblem entsteht seit `M12`** (114.1) und liegt in der Kommandozone, im Besitz und unter der Kontrolle dessen, der es bekommt (114.2). **Ohne Fähigkeit** — 114.3 sagt, dass es nichts anderes hat, und eine Fähigkeit ist ein Kartentext (`M8`); 114.4 (sie wirken aus der Kommandozone) hat damit nichts zu wirken | M12 · M8 |
| **CR 115 — Targets** | teilweise | `internal/domain/ziele.go` (115.1, 115.3), `internal/domain/stapel.go` (608.2b) | **Ziele werden beim Ankündigen genannt** (115.1) und **beim Verrechnen noch einmal geprüft** (608.2b); **dasselbe Ziel nicht zweimal** (115.3). **Was offen bleibt, ist die Frage davor**: **ob** ein Zauber ein Ziel verlangt und wie viele, steht in seinem Text (`M8`) — geprüft wird, ob die **genannten** zulässig sind. 115.2, 115.4–115.10 offen | M4 · M8 |
| **CR 116 — Special Actions** | teilweise | `internal/domain/landspielen.go` (116.1, 116.2a) | **Eine von zwölf**: ein Land spielen (116.2a) — und sie geht **nicht über den Stapel** (116.1). Es ist die einzige, die ohne Kartentext geht; die übrigen elf brauchen ein Schlüsselwort (aussetzen, vorhersagen, planen, Gefährte — `M14`), eine verdeckte Karte (`M12`) oder eine Variante, die #1593 ausgrenzt | M4 · M12 · M14 |
| **CR 117 — Timing and Priority** | teilweise | `internal/domain/prioritaet.go:20` (117.1), `:144` (Passen und 117.4), `internal/service/sandkasten.go:176` (wer ohne Priorität nichts darf) | **117.3b trägt seit M1** (`internal/domain/prioritaet.go`): nach dem Verrechnen bekommt der **aktive** Spieler Priorität. **117.3c** (nach dem Wirken) braucht `M4`, **117.5** ist mit 405.5 gebaut. **CR 732 trägt teilweise** — `zug.abkuerzen` spart die zwanzig Passen je Zug, die Verhandlung nach 732.2b fehlt | G9 · M1 · M4 |
| **CR 118 — Costs** | teilweise | `internal/domain/manazahlung.go` (118.3, 601.2h) | **Was ein Vorrat deckt, ist gerechnet** — farbig, farblos, generisch, hybrid und einfarbig-hybrid, und **ganz oder gar nicht** (118.3, 601.2h). **Gebraucht wird es noch von niemandem**: bezahlt wird beim Wirken, und das ist `M4`. **Offen**: 118.1/118.2, 118.4 (`{X}`), 118.5 (`{0}` ausdrücklich zahlen), 118.6–118.12 | M3 · M4 |
| **CR 119 — Life** | teilweise | `internal/domain/partie.go` (der Lebensstand am Sitz, 119.1), `internal/domain/spielerstand.go:57` (gesetzt aus dem Profil), `internal/service/zugbefehle.go:129` (der Befehl `zug.leben`) | **Ein Sitz hat Leben, und null kostet die Partie** (CR 704.5a, `M17`). Offen: 119.3 (Verlust und Gewinn als Ereignis) und 119.4–119.8 — sie brauchen Effekte; **niemand fügt Schaden zu**, der Stand wird gesagt | G9 · M17 · M6 |
| **CR 120 — Damage** | teilweise | `internal/domain/spielobjekt.go:186` (markierter Schaden, 120.3), `internal/domain/zughandlungen.go:133` (er geht im Aufräumschritt weg, 514.2) | **Niemand fügt ihn zu.** `zug.schaden` trägt ihn ein, weil ein Mensch es sagt; 120.1–120.2 (wer wem wie), 120.3c–e (tödlicher Schaden, Lebensverlust) und 120.6 (Verhinderung) brauchen `M6` und `M11` | G9 · M6 · M11 |
| **CR 121 — Drawing a Card** | teilweise | `internal/domain/zughandlungen.go:95` (121.1), `:83` (die zugbasierte Ziehung, 504.1), `internal/domain/partie.go` (der Vermerk über die leere Bibliothek, 704.5b), `internal/domain/eroeffnung.go` (die Starthand, 121.2) | **Die oberste Karte wird gezogen** (die Eins liegt oben), und eine **leere** Bibliothek wird vermerkt statt verschluckt. **121.2 trägt für die Starthand** (#1839): sieben Einzelziehungen über denselben Weg. Offen: 121.2 für Effekte, die mehrere ziehen lassen, und 121.3 (Ersetzungen) — beides braucht `M11` | G9 · M17 · M11 · #1839 |
| **CR 122 — Counters** | teilweise | `internal/domain/spielobjekt.go:173`, `internal/service/sandkasten.go:209` | **Das Grundprofil führt diesen Bereich als `automatisch`** (`internal/domain/grundprofil.go:115`) und nennt zwei Grenzen. **Eine dritte fehlt und sie ist die härteste: CR 122.2** — Zähler überleben den Zonenwechsel nicht, und `schieben` nimmt sie nicht weg (Befund 1). **122.1a trägt seit dem 2026-09-17** (`internal/domain/objekteigenschaften.go:79`): +1/+1 und −1/−1 ändern Stärke und Widerstandskraft, und **122.3 hebt sie gegeneinander auf** (`internal/domain/objekteigenschaften.go:110`, CR 704.5q). 122.1b–j (Schlüsselwortmarken, Schild, Betäubung, Loyalität, Gift), 122.4, 122.5–122.8 offen | M18 (122.2) · M5 · M14 |
| **CR 123 — Stickers** | offen | — | — | M2 |

### 4.2 Parts of a Card — CR 200–213

**Die ganze Zeile ist hier zu lesen, bevor ein Stand missverstanden wird.**
Dieses Programm ist ein **Karteneditor**: es zeichnet jeden dieser Kartenteile,
und es tut das seit Phase 2. Was es nicht tut, ist sie als **Merkmale eines
Spielobjekts** führen. `domain.Card` ist eine geordnete Liste von
`Value{Field, Text}` (`internal/domain/set.go:11` und `:40`) — Text, immer, und
die Typisierung steht im Spielpaket. Genau deshalb ist CR 208 („Power/
Toughness") hier `offen`, obwohl jede gezeichnete Karte ihre Stärke trägt.

| Regel | Stand | Beleg | Grenze | Werkstück |
| --- | --- | --- | --- | --- |
| **CR 200 — General** | teilweise | `internal/domain/set.go:40`, `internal/domain/karte.go:64` | Eine Karte ist eine Liste benannter **Textfelder**, deren Definition aus dem Spielpaket kommt. Keiner der vierzehn Kartenteile aus 200.1 ist ein eigener Begriff | M2 |
| **CR 201 — Name** | teilweise | `internal/domain/deckurteil.go:344`, `internal/domain/standardland.go:130` | **201.3b (austauschbare Namen)** trägt: „Universes Beyond" und „Universes Within" zählen über die `oracle_id` als **eine** Karte. Sonst ist der Name ein Feld | M2 |
| **CR 202 — Mana Cost and Color** | teilweise | `internal/domain/manakosten.go:28` | **202.2 und 202.3 tragen seit dem 2026-09-17**: die Farbe folgt den Symbolen (`internal/domain/manakostenwert.go:97`, samt 202.2b–d für farblos, mehrfarbig und hybrid), der Manawert wird gerechnet (`:127`). **202.1b** (ein Land hat keine Kosten) fällt als leerer Kosten heraus. **Offen**: 202.3b–d — zweiseitige, verschmolzene und geteilte Karten brauchen mehrere Seiten zugleich (`M12`) | M2 · M12 |
| **CR 203 — Illustration** | trifft nicht zu | — | **CR 203.1 sagt es selbst**: die Illustration hat keine Wirkung auf das Spiel. Für die **Regelmaschine** gibt es hier nichts durchzusetzen — der Editor zeichnet sie (`internal/render`), und das ist eine andere Frage | — |
| **CR 204 — Color Indicator** | offen | — | — | M2 |
| **CR 205 — Type Line** | teilweise | `internal/domain/standardland.go:51`, `frontend/src/bausteine/deckstatistik.ts:79` | **Die Typzeile ist seit dem 2026-09-17 kein Text mehr** (`internal/domain/kartentyp.go:120`): 205.1, 205.2a (die **fünfzehn** Kartentypen), 205.3a (Untertypen hinter dem Strich) und 205.4a (die **fünf** Supertypen) sind Daten, und ein Wort, das CR 205.2a nicht kennt, ist ein **Befund** (§8). Daneben tragen 205.4c (`basic` macht ein Standardland — samt der Falle, dass „Dryad Arbor" den Untertyp `Forest` trägt und keines ist) und die Deckstatistik. **Offen**: 205.2b/205.2c (die Reihenfolge der Typen), 205.3b–m (welcher Untertyp zu welchem Typ gehört), 205.4b/205.4d–e. **`BereichTypen` steht in der Matrix seither auf `begleitet`** und nicht mehr auf `manuell`: das Programm liest und **warnt**, es setzt keine Typregel durch (Grundprofilfassung 4) | M2 · M12 |
| **CR 206 — Expansion Symbol** | trifft nicht zu | — | **CR 206.1**: keine Wirkung auf das Spiel. Der Editor zeichnet es | — |
| **CR 207 — Text Box** | teilweise | `internal/domain/set.go:11` | Der Regeltext ist ein Feld, und **das ist der Kern der Wand**: was dort steht, hat ein Mensch getippt. Niemand deutet ihn | M8 |
| **CR 208 — Power/Toughness** | teilweise | `internal/domain/kartenwert.go:27` (208.1–208.3), `internal/domain/objekteigenschaften.go:79` (mit Marken) | **Drei Antworten statt zweier**: eine Nichtkreatur hat **keine** Stärke und nicht null (208.3), und ein `*` bleibt ein Stern (208.2). **Was der Stern bedeutet, sagt eine Fähigkeit** (CR 604.3) — die gibt es nicht (`M8`). 208.2a–c und 208.4 offen | M2 · M8 |
| **CR 209 — Loyalty** | offen | — | — | M2 |
| **CR 210 — Defense** | offen | — | Verteidigung gehört zu den Battles (CR 310), und die nennt kein Werkstück | M2 |
| **CR 211 — Hand Modifier** | trifft nicht zu | — | Nur auf Vanguard-Karten (CR 902), und #1593 grenzt Vanguard ausdrücklich aus | — |
| **CR 212 — Life Modifier** | trifft nicht zu | — | Dito | — |
| **CR 213 — Information Below the Text Box** | trifft nicht zu | — | **CR 213.1**: keine Wirkung auf das Spiel | — |

### 4.3 Card Types — CR 300–315

| Regel | Stand | Beleg | Grenze | Werkstück |
| --- | --- | --- | --- | --- |
| **CR 300 — General** | offen | — | Die fünfzehn Kartentypen aus 300.1 sind kein Begriff dieses Programms | M2 |
| **CR 301 — Artifacts** | teilweise | `internal/domain/spielobjekt.go:164`, `internal/service/sandkasten.go:245` | **Das Grundprofil führt `anhaengsel` als `automatisch`** (`internal/domain/grundprofil.go:124`). Was trägt, ist die **Buchführung**: `HaengtAn` weiß, woran etwas hängt. Was nicht trägt, steht als Grenze ausdrücklich da — „ob etwas dort hängen darf, prüft niemand" (301.5c) — und **CR 400.7**: eine Ausrüstung behält ihr Ziel auch dann, wenn sie den Tisch verlässt (Befund 1) | M12 · M7 |
| **CR 302 — Creatures** | teilweise | `internal/domain/kampf.go` (302.6), `internal/domain/spielobjekt.go` (`Seit`) | **Die Einsatzverzögerung trägt** (302.6): was in diesem Zug gekommen ist, greift nicht an — es sei denn, auf der Karte steht **Eile** (CR 702.10b, den CR 508.1a namentlich nennt). Der Zug, in dem ein Permanent gekommen ist, steht am Objekt. **Offen**: 302.1–302.5 (was eine Kreatur ist), 302.7/302.8 | M6 · M12 |
| **CR 303 — Enchantments** | teilweise | `internal/service/sandkasten.go:245`, `internal/domain/zustandsanhaengsel.go:58` (704.5m) | Wie CR 301: die Aura hängt, **ob sie darf, prüft niemand**. **Was seit `M5` trägt**: eine Aura an nichts oder an etwas, das nicht mehr auf dem Tisch liegt, geht in den Friedhof ihres **Eigners** (CR 704.5m) | M12 · M7 · M5 |
| **CR 304 — Instants** | offen | — | — | M4 |
| **CR 305 — Lands** | teilweise | `internal/domain/landmana.go` (305.6), `frontend/src/bausteine/Deckkennzahlen.tsx:111` | **305.6 trägt zweimal**: seit `M3` als **innewohnende Fähigkeit** — ein Grundland tappt für sein Mana, und das ist die **einzige** Manafähigkeit, die im Regelwerk steht statt im Kartentext — und weiterhin für die **Deckstatistik**, also für die Frage „welche Farben gibt dieses Deck". **305.1 (ein Land spielen ist eine Sonderhandlung)** und 305.2–305.5 offen | M3 · M4 |
| **CR 306 — Planeswalkers** | teilweise | `internal/domain/kartentyp.go:41` (der Kartentyp), `internal/domain/karteneigenschaften.go:74` (die gedruckte Zahl, 306.5b), `internal/domain/zustandsobjekte.go` (704.5i) | **Der Typ, die Zahl und das Sterben stehen.** Ein Planeswalker mit Loyalität null geht seit `M5` in den Friedhof (CR 704.5i) — **sobald Loyalitätsmarken dastehen**. Dass er **mit so vielen ins Spiel kommt** (306.5b), ist ein Ersetzungseffekt (CR 614.1c) und braucht `M11`; ohne ihn legt sie niemand ab. Loyalitätsfähigkeiten sind CR 606 (`M4`) | M2 · M5 · M11 · M4 |
| **CR 307 — Sorceries** | offen | — | — | M4 |
| **CR 308 — Kindreds** | offen | — | — | M2 |
| **CR 309 — Dungeons** | offen | — | Nichttraditionelle Karten mit eigener Zone (408) | M12 |
| **CR 310 — Battles** | teilweise | `internal/domain/kartentyp.go:31` (der Kartentyp, 205.2a), `internal/domain/karteneigenschaften.go:82` (die Zahl, 310.1/210.1), `internal/domain/zustandsobjekte.go` (704.5v/w) | **Typ, Zahl und das Sterben stehen**: eine Kampfkarte mit Verteidigung null geht seit `M5` in den Friedhof. 310.2–310.11 (Verteidiger, Beschützer, Schaden) brauchen Kampf (`M6`) — und mit ihnen CR 704.5x/y. **Der Feldname der Verteidigung ist eine Annahme**, und das steht am Code (§15) | M2 · M5 · M6 |
| **CR 311 — Planes** | trifft nicht zu | — | Nur Planechase (CR 901), von #1593 ausgegrenzt | — |
| **CR 312 — Phenomena** | trifft nicht zu | — | Dito | — |
| **CR 313 — Vanguards** | trifft nicht zu | — | Nur Vanguard (CR 902), ausgegrenzt | — |
| **CR 314 — Schemes** | trifft nicht zu | — | Nur Archenemy (CR 904), ausgegrenzt | — |
| **CR 315 — Conspiracies** | trifft nicht zu | — | Nur Conspiracy Draft (CR 905), ausgegrenzt | — |

### 4.4 Zones — CR 400–408

**Das ist der Abschnitt, den das Grundprofil als `automatisch` führt**, und
deshalb steht hier Zeile für Zeile, was das trägt und was nicht.

| Regel | Stand | Beleg | Grenze | Werkstück |
| --- | --- | --- | --- | --- |
| **CR 400 — General** | teilweise | `internal/domain/spielobjekt.go:79` (die sieben Zonen, 400.1), `:113` (öffentlich und verborgen, 400.2), `internal/domain/zonenordnung.go:26` (die Reihenfolge, 400.5), `internal/domain/zonenwechsel.go` (400.7 über `NeuesObjekt`) | **400.7 trägt seit `M18`**: was die Zone wechselt, ist ein **neues** Objekt — ohne Marken, ohne Zustand, ohne markierten Schaden, ohne Anhängsel, und was daran hing, hängt an nichts mehr. **An einer Stelle und nicht an zwölf**: jeder Zonenwechsel in der Domäne geht über `Partie.zonenwechsel`, und der Sandkasten nimmt dieselbe Funktion. **Die Kennung überlebt** — sie ist die Identität in diesem Programm und nicht die des Regelwerks; das steht als Grenze in der Matrix. **Die zwölf Ausnahmen aus 400.7a–m fehlen**, und zwar alle: jede einzelne spricht von Effekten, Fähigkeiten oder Auslösern (`M8`, `M9`). **400.3 steht nur in der Anzeige** (`frontend/src/tisch/tischplan.ts:110`) und nicht im Zustand: wer eine fremde Kreatur beherrscht, legt sie beim Sterben in **seinen** Friedhof (Befund 5). 400.6, 400.9–400.11 offen | M18 (400.6–400.11) · M8 · M9 |
| **CR 401 — Library** | teilweise | `internal/service/partieaufbau.go:227` (aus dem Deck wird eine Bibliothek, 401.1), `internal/domain/spielobjekt.go:121` (verdeckt, 401.2), `internal/service/sandkastenzonen.go:150` (mischen) | **Der Blick in die eigene Bibliothek ist erlaubt** (`internal/service/sandkastenzonen.go:92`) — die benannte Sandkastenausnahme, und sie zahlt mit einer **Mischpflicht** (`internal/domain/mischpflicht.go:55`). 401.3 (jeder darf zählen), 401.4 (eine leere Bibliothek ist noch keine Niederlage), 401.5–401.7 offen | G9 |
| **CR 402 — Hand** | teilweise | `internal/domain/spielobjekt.go:118` | Die Hand gehört ihrem Beherrscher, und niemand sonst sieht hinein (402.3 und die Sichtbarkeitsinvariante aus G1). **402.2 (die Höchstzahl, normalerweise sieben) fehlt** | G9 |
| **CR 403 — Battlefield** | teilweise | `internal/service/partieaufbau.go:245` | Das Schlachtfeld fängt leer an (403.1) — alles liegt in der Bibliothek. **403.3 („jedes Objekt auf dem Schlachtfeld ist eine bleibende Karte") prüft niemand**; eine Instant-Karte lässt sich dorthin schieben | M12 |
| **CR 404 — Graveyard** | teilweise | `internal/domain/zonenordnung.go:45` (400.5), `internal/domain/zonenwechsel.go` (404.1 über `obenauf`) | **Befund 2 ist behoben**: G9 führt den Friedhof seit dem 2026-09-17 unter den **geordneten** Zonen (CR 400.5), und `M18` vergibt seitdem die Ordnung — **was hineinkommt, kommt obenauf** (404.1), je Eigner gezählt, weil ein Friedhof einem Spieler gehört. Bis dahin bekam jede Karte die Null, und alle lagen an derselben Stelle. **404.3 fehlt** (der Eigner ordnet, was gleichzeitig hineinkommt): gleichzeitig kommt hier nichts hinein — jede Handlung legt eine Karte, und die zustandsbasierte Prüfung legt sie nacheinander | M18 · M9 |
| **CR 405 — Stack** | teilweise | `internal/domain/spielobjekt.go:72`, `internal/domain/zonenordnung.go:28`, `internal/domain/stapel.go:44` (405.2), `internal/domain/prioritaet.go` (405.5) | **Was oben liegt, verrechnet zuerst** (405.2), und passen alle nacheinander, verrechnet es (405.5). **Nichts legt von selbst etwas darauf** — das Wirken ist `M4`, Fähigkeiten sind `M8`; bis dahin kommt ein Objekt über `objekt.schieben` dorthin. 405.3 (APNAP beim gleichzeitigen Legen), 405.4 (ein Zauber trägt die Eigenschaften seiner Karte) und 405.6 offen | M1 · M4 · M8 |
| **CR 406 — Exile** | teilweise | `internal/domain/spielobjekt.go:64` | Zone vorhanden und öffentlich. 406.3 (offen oder verdeckt im Exil) nur, soweit `Verdeckt` es hergibt; 406.2, 406.4–406.8 offen | M12 |
| **CR 407 — Ante** | trifft nicht zu | — | **CR 407.1**: in keinem sanktionierten Spiel erlaubt, und nur dort, wo Gesetz und Regelwerk es nicht verbieten. `domain.Zonen()` führt sie nicht — das ist eine Entscheidung und keine Lücke | — |
| **CR 408 — Command** | teilweise | `internal/domain/spielobjekt.go:75`, `internal/domain/kartengestalten.go` (114.1), `internal/domain/commander.go` (903.9a) | Die Zone ist da, **und seit `M12` und `M19` liegt etwas darin**: ein **Emblem** (CR 114.1) entsteht dort, und ein **Commander** kommt aus Friedhof oder Exil dorthin zurück (CR 903.9a). **Ein Verlies fehlt** (CR 309) — es ist ein Kartentext (`M8`). **Und 408.2 fehlt**: dass eine Fähigkeit aus der Kommandozone wirkt, setzt eine voraus | M19 · M8 |

### 4.5 Turn Structure — CR 500–514

**Hier stand bis zum 2026-09-17 „alle fünfzehn Kapitel sind offen".** G9
(#1592) hat
die Zugfolge gebaut: zwölf Abschnitte, zehn davon vergeben Priorität, und
`zugstruktur` wie `prioritaet` stehen in der Regelmatrix seither auf
`automatisch` statt auf `manuell`. **Neun der fünfzehn Kapitel sind damit
teilweise**; offen bleiben die sechs des Kampfes (CR 506–511), und die gehören
`M6`.

**Und die Spalte „Beleg" ist neu.** Solange alles offen war, gab es nichts zu
belegen, und die Tabelle hatte vier Spalten statt fünf — eine Zeile mit Stand
`teilweise` und ohne Fundstelle wäre genau die Behauptung, die diese Datei
verhindern soll.

| Regel | Stand | Beleg | Grenze | Werkstück |
| --- | --- | --- | --- | --- |
| **CR 500 — General** | teilweise | `internal/domain/zugstruktur.go:88` (die zwölf Abschnitte, 500.1), `internal/domain/prioritaet.go:144` (ein Abschnitt endet, wenn **der Stapel leer ist und** alle nacheinander passen, 500.2 — die erste Hälfte kam mit `M1`) | **500.5 trägt seit `M3` zur Hälfte**: der Manavorrat leert sich am Ende jedes Abschnitts (`internal/domain/prioritaet.go`, `abschnittVerlassen`) — **die andere Hälfte fehlt**, Effekte enden mit dem Abschnitt nicht (CR 611, `M10`). 500.3, 500.4 (übersprungene Abschnitte), 500.6–500.11 offen | G9 · M3 · M10 |
| **CR 501 — Beginning Phase** | teilweise | `internal/domain/zugstruktur.go:92` | Die drei Schritte stehen in der Zugfolge und werden betreten | G9 |
| **CR 502 — Untap Step** | teilweise | `internal/domain/zughandlungen.go:58` (502.3), `internal/domain/zugstruktur.go:120` (keine Priorität, 502.4) | **502.1 (Phasing) fehlt** — es braucht CR 702.26 (`M14`). Enttappt wird **alles** des aktiven Spielers auf dem Tisch; Effekte, die etwas festhalten, gibt es nicht | G9 · M14 (502.1) |
| **CR 503 — Upkeep Step** | teilweise | `internal/domain/zugstruktur.go:93` | Der Schritt wird betreten und vergibt Priorität. **Ausgelöste Fähigkeiten gibt es nicht** (`M9`) — und sie sind der Grund, warum es diesen Schritt gibt | G9 · M9 |
| **CR 504 — Draw Step** | teilweise | `internal/domain/zughandlungen.go:83` (504.1), `:84` (103.8a) | Gezogen wird die oberste Karte. **Eine leere Bibliothek wird vermerkt** (CR 704.5b) und kostet bei der nächsten Prüfung die Partie | G9 · M17 |
| **CR 505 — Main Phase** | teilweise | `internal/domain/zugstruktur.go:95`, `:101` | Beide Hauptphasen stehen in der Zugfolge und vergeben Priorität. **Was man darin tut, gibt es nicht**: kein Land spielen (505.5b), kein Zauber (`M4`) | G9 · M4 |
| **CR 506 — Combat Phase** | teilweise | `internal/domain/zugstruktur.go` (506.1), `internal/domain/kampf.go` (506.2) | **Die fünf Schritte haben Inhalt** (506.1), und **wer angreift und wer verteidigt** steht fest (506.2). **Offen**: 506.3 (übersprungene Schritte), **506.4 (aus dem Kampf entfernen)** — ein Permanent, das den Kampf verlässt, bleibt im Kampfstand stehen und teilt trotzdem keinen Schaden aus —, 506.5 (Bandenkampf) | M6 · M12 |
| **CR 507 — Beginning of Combat Step** | offen | — | Wer verteidigt (im Mehrspielerfall) | M6 |
| **CR 508 — Declare Attackers Step** | teilweise | `internal/domain/kampf.go` (508.1a, 508.1f) | **Die Erklärung trägt**: ungetappt (508.1a), Kreatur, dem aktiven Spieler gehörend, ein zulässiges Ziel — ein gegnerischer Sitz, ein Planeswalker oder eine Kampfkarte —, und die Gewählten werden dabei **getappt** (508.1f). **Auch keine Angreifer ist eine Erklärung.** **Offen**: 508.1c/d (Einschränkungen und Pflichten aus Effekten — `M8`), 508.1g (optionale Angriffskosten), 508.2–508.4 **Und seit `M14` liest die Erklärung Schlüsselwörter**: ein Verteidiger greift nicht an (702.3b), und Wachsamkeit hebt das Tappen auf (702.20b). | M6 · M14 · M8 |
| **CR 509 — Declare Blockers Step** | teilweise | `internal/domain/kampf.go` (509.1a, 509.1h) | **Die Erklärung trägt**: ungetappt, Kreatur des Verteidigers, und sie blockt einen **erklärten** Angreifer (509.1a). **Blocken tappt nicht** — CR 508.1f gilt nur fürs Angreifen. **Offen**: 509.1b (Einschränkungen — `M8`), **509.2 (die Schadensreihenfolge bei mehreren Blockern)** — hier gilt die Reihenfolge der Erklärung —, 509.3–509.7 **Und seit `M14`**: Flugfähiges blockt nur, wer fliegt oder Reichweite hat (702.9b, 702.17b), und Bedrohliches braucht zwei (702.111b). | M6 · M14 · M8 |
| **CR 510 — Combat Damage Step** | teilweise | `internal/domain/kampfschaden.go` (510.1a–d, 510.2), `internal/domain/erstschlag.go` (510.4) | **Zuteilen und austeilen sind zwei Schritte in einem** (510.1, 510.2): erst rechnet jede Kreatur ihre Stärke gegen den Stand **vor** dem Austeilen, dann wird **gleichzeitig** ausgeteilt — eine 2/2, die von einer 3/3 erschlagen wird, schlägt trotzdem zurück. **Erst- und Doppelschlag teilen den Schritt** (510.4): dieses Programm liest die zwei Schlüsselwörter, weil CR 510.4 sie **namentlich nennt**. **Offen**: 510.3 (Schaden an einem Planeswalker umleiten), 510.5 **Und seit `M14`**: Trampelschaden gibt den Rest weiter (702.19b), Lebensverknüpfung gibt Leben (702.15b), Todesberührung macht jeden Schaden tödlich (702.2b). | M6 · M14 |
| **CR 511 — End of Combat Step** | teilweise | `internal/domain/prioritaet.go` (511.3) | **511.1 trägt von selbst** — der Schritt hat keine zugbasierte Handlung, und Priorität gibt es dort wie überall. **511.3**: mit dem Zugende ist niemand mehr Angreifer oder Blocker; der Kampfstand geht mit dem Zug auf null. **511.2 offen** — es gibt keine Effekte „bis zum Ende des Kampfes" (`M10`) | M6 · M10 |
| **CR 512 — Ending Phase** | teilweise | `internal/domain/zugstruktur.go:102` | Beide Schritte stehen in der Zugfolge | G9 |
| **CR 513 — End Step** | teilweise | `internal/domain/zugstruktur.go:102` | Der Schritt wird betreten und vergibt Priorität; **was hier auslöst, gibt es nicht** (`M9`) | G9 · M9 |
| **CR 514 — Cleanup Step** | teilweise | `internal/domain/zughandlungen.go:133` (514.1, 514.2), `internal/domain/prioritaet.go:121` (514.3) | **Die zweite Hälfte von 514.2 trägt seit `M10`**: was „bis zum Ende des Zuges" gilt, endet im Aufräumschritt. Die Abwurfpflicht **wartet** auf die Wahl des Spielers statt selbst zu wählen | G9 · M10 |

### 4.6 Spells, Abilities, and Effects — CR 600–616

**Das ist der Abschnitt mit den meisten Lücken im Zuschnitt.** Sieben von
siebzehn Kapiteln nennt kein Werkstück, und zwei davon — CR 604 (statische
Fähigkeiten) und CR 609 (Effekte) — sind Voraussetzung für Kapitel, die sehr
wohl eines haben.

| Regel | Stand | Grenze | Werkstück |
| --- | --- | --- | --- |
| **CR 600 — General** | trifft nicht zu | **Das Kapitel trägt keine nummerierte Regel** — es ist eine Überschrift. Nachgezählt: 0 Haupt- und 0 Unterregeln | — |
| **CR 601 — Casting Spells** | teilweise | `internal/domain/wirken.go` (601.2a, 601.2c, 601.2f, 601.2h, 601.2i) | **Fünf der neun Schritte**: die Karte wandert auf den Stapel und bekommt ihren Beherrscher (601.2a), die Ziele werden genannt (601.2c), die Kosten stehen fest (601.2f), sie werden **ganz** bezahlt (601.2h), danach ist er gewirkt (601.2i). **Die vier offenen brauchen einen Kartentext**: Modi (601.2b), Wertverteilung (601.2d), `{X}` (601.2e) und Manafähigkeiten mitten im Bezahlen (601.2g). **Und der ganze Vorgang ist atomar** (601.2h, CR 733): was scheitert, lässt Vorrat, Hand und Stapel unberührt | M4 · M8 |
| **CR 602 — Activating Activated Abilities** | offen | — | M4 · M8 |
| **CR 603 — Handling Triggered Abilities** | teilweise | `internal/domain/ausloeser.go` (603.2, 603.4), `internal/domain/ausloeserstapel.go` (603.3, 603.3b) | **Der Mechanismus steht seit `M9`, der Vorrat nicht**: eine Fähigkeit löst aus und **tut nichts** (603.2), kommt auf den Stapel, wenn das nächste Mal jemand Priorität bekäme (603.3), und zwar in APNAP-Reihenfolge — **innerhalb eines Sitzes ordnet er selbst** (603.3b, Befehl `spieler.reihung`). Die Bedingung einer „intervening if"-Fähigkeit wird beim Auslösen **und** beim Verrechnen geprüft (603.4). **Vier Fähigkeiten gibt es, und alle stehen im Regelwerk** (725.2, 726.2, 728.1) — 603.1 verlangt einen **Kartentext**, und den bringt `M8`. Offen: Zustandsauslöser (603.8), verzögerte (603.7), reflexive (603.12) und der zweite Teil von 603.3b | M9 · M8 |
| **CR 604 — Static Abilities** | offen | **Statische Fähigkeiten**. M10 baut CR 613, also die Ordnung, in der Dauereffekte wirken — was sie **erzeugt**, steht hier | M8 |
| **CR 605 — Mana Abilities** | teilweise | `internal/domain/landmana.go` (605.3b) | **605.3b trägt**: die eine Manafähigkeit, die es gibt — das Grundland —, wird **sofort** verrechnet und geht nicht über den Stapel; niemand kann darauf antworten. **605.1a (was überhaupt eine Manafähigkeit ist) bleibt offen**, und das ist kein Versäumnis: die Frage stellt sich erst, wenn ein Kartentext eine Fähigkeit trägt, und das ist `M8` | M3 · M8 |
| **CR 606 — Loyalty Abilities** | offen | Ohne sie gibt es keine Planeswalker (CR 306). **`M4` hat sie nicht gebaut, und das ist begründet**: eine Loyalitätsfähigkeit steht als **Text** auf der Karte, und einen Text auszuführen ist `M8`. Sie steht als Grenze in der Regelmatrix | M4 · M8 |
| **CR 607 — Linked Abilities** | offen | 25 Regeln. Ein Kartentext, der eine Karte exiliert und sie später spielbar macht, ist ohne dieses Kapitel nicht auszulegen | M8 |
| **CR 608 — Resolving Spells and Abilities** | teilweise | `internal/domain/stapel.go:71` (608.1), `:106` (608.2n, 608.3a) | **Der Weg steht, der Inhalt nicht**: das oberste Objekt verlässt den Stapel und landet, wo die Regel es hinlegt — ein Spontan- oder Hexereizauber im Friedhof seines **Eigners** (608.2n), ein bleibendes Objekt auf dem Schlachtfeld (608.3a). **„Tu, was der Zauber sagt" (608.2c) geschieht seit `M8`** — wenn die Karte ein `game script` trägt, und **seit dem 2026-09-19 zum richtigen Zeitpunkt**: gelesen wird beim **Verrechnen** und nicht mehr beim Ankündigen (#1625); zwischen beiden liegt mindestens eine Prioritätsrunde. Kein bezogenes Paket bringt eines mit; wer eine eigene Karte baut, schreibt es hinein. 608.2a–b (Modi und Ziele) brauchen `M4`, 608.3b–e mehrseitige und angehängte Objekte (`M12`). **Und seit `M9` geht eine Fähigkeit ihren eigenen Weg**: sie ist keine Karte, wird ausgeführt und hört danach auf zu existieren (608.2n, zweiter Satz) | M1 · M4 · M8 · M9 |
| **CR 609 — Effects** | offen | Was ein Effekt ist, wer ihn beherrscht, wie lange er gilt | M8 |
| **CR 610 — One-Shot Effects** | offen | — | M8 |
| **CR 611 — Continuous Effects** | teilweise | `internal/domain/dauereffekt.go` (611.1) | **Ein Dauereffekt gilt für einen festen oder unbestimmten Zeitraum** (611.1) — seit `M10`. **Zwei Fristen von vielen**: bis zum Zugende (CR 514.2, zweite Hälfte) und unbestimmt. **„Solange …" fehlt** (611.2b): es braucht eine Bedingung, die laufend geprüft wird. 611.3 (was ein Effekt betrifft) bleibt offen, solange ein Effekt sein Objekt beim Namen nennt | M10 · M8 |
| **CR 612 — Text-Changing Effects** | offen | **Schicht 3 von CR 613 ist genau dieses Kapitel** (613.1c verweist darauf) | M10 |
| **CR 613 — Interaction of Continuous Effects** | teilweise | `internal/domain/dauereffekt.go` (613.1, 613.4b/c, 613.7) | **Sieben Schichten** (613.1a–g), 41 Regeln — die Zahl „30" in #1593 ist die der **Unterregeln** und nicht die der Schichten. **Zwei davon wirken seit `M10`**: die sechste (Fähigkeiten geben, 613.1f) und die siebte (Werte, 613.4b setzen und 613.4c ändern). **Innerhalb einer Schicht entscheidet der Zeitstempel** (613.7). **613.8 (Abhängigkeit) hat keinen Fall** — eine Abhängigkeit entsteht nur **innerhalb** einer Schicht (613.8a), und dort nennt jeder Effekt sein Objekt bei der Kennung; ein Abhängigkeitssystem ohne Abhängigkeit wäre ein Zweig, den kein Weg betritt (§10). Die acht übrigen Schichten haben keinen Effekt, der sie erzeugte | M10 · M8 |
| **CR 614 — Replacement Effects** | teilweise | `internal/domain/ersetzungseffekt.go` (614.1, 614.1a, 614.5, 614.6) | **37 Regeln**, und seit `M11` trägt der Kern davon: ein Effekt wartet auf ein Ereignis und setzt ein anderes an seine Stelle (614.1), das ursprüngliche **geschieht nie** (614.6), und er bekommt **eine** Gelegenheit (614.5). **Ersetzbar ist ein Ereignis**: der Weg in den Friedhof, ersetzt durch das Exil (614.1a). Alles andere führt dieses Programm nicht als Ereignis — Ziehen, Wirken, das Hineinkommen eines Permanents. **614.15 (Selbstersetzung) fehlt**: sie braucht einen Kartentext, der sich selbst meint (`M8`) | M11 · M8 |
| **CR 615 — Prevention Effects** | teilweise | `internal/domain/ersetzungseffekt.go` (615.1, 615.10), `internal/domain/kampfschaden.go` | **Verhinderter Schaden wird nicht markiert** (615.1) — im Kampf (510.2) wie bei einer Fähigkeit (120.3). **Und nur so viel, wie der Effekt nennt** (615.10): ein Effekt über eins verhindert eins und nicht drei; ohne Zahl verhindert er alles. **Was fehlt, ist die Quelle**: ein Effekt, der nur Schaden **einer bestimmten Quelle** verhindert, braucht einen Kartentext (`M8`) | M11 · M8 |
| **CR 616 — Interaction of Replacement and/or Prevention Effects** | teilweise | `internal/domain/ersetzungseffekt.go` (616.1) | **Treffen mehrere zusammen, greift die älteste** — und das ist eine **benannte Abweichung** (§15, Punkt 4): 616.1 ließe den **Betroffenen** wählen, und gefragt wird er nicht. Der Weg dazu gäbe es (die Reihung aus `M9` hat dieselbe Gestalt); er ist nicht gebaut. **Welcher Effekt gegriffen hat, ist am Ergebnis ablesbar** | M11 |

### 4.7 Additional Rules — CR 700–733

**Vierunddreißig Kapitel, 1.525 Regeln — knapp die Hälfte des Regelwerks.**
Zwölf davon nennt ein Werkstück, **einundzwanzig nicht**, und eines (CR 703)
wird gebaut, ohne genannt zu werden.

| Regel | Stand | Beleg | Grenze | Werkstück |
| --- | --- | --- | --- | --- |
| **CR 700 — General** | offen | — | 35 Regeln: was ein Ereignis ist, „als ob", „kann nicht", „solange", Zahlenbereiche, „bis zu", Zufallsauswahl | M8 |
| **CR 701 — Keyword Actions** | teilweise | `internal/domain/schluesselhandlungen.go` (701.6, 701.8, 701.13, 701.14, 701.17, 701.21), `internal/domain/schluesselwahlen.go` (701.22, 701.25), `internal/service/sandkastenzonen.go:150` (701.24), `internal/service/sandkasten.go:176` (701.26), `internal/domain/kartengestalten.go` (701.7, 701.27) | **Zwölf von siebzig** (701.2 bis 701.71; 701.1 ist die Einleitung) — in der Fassung 2026-08-19 waren es neunundsechzig bis 701.70. **Acht kamen mit `M7`**: neutralisieren (701.6), zerstören (701.8), exilieren (701.13), kämpfen (701.14), mahlen (701.17), opfern (701.21), hellsehen (701.22), überwachen (701.25) — es sind die, die das **Regelwerk allein** beschreibt. Zwei standen schon als Buchführung da: mischen (701.24) und tappen/enttappen (701.26). **Zwei kamen mit `M12`**: erzeugen (701.7) und umwandeln (701.27) — beide sind Gestalten und keine Bewegung, und beide brauchten erst die Felder am Objekt. **Die übrigen siebenundfünfzig sind Bausteine eines Kartentextes** und brauchen `M8`. **Und keine geschieht von selbst** — sie wird gestellt | M7 · M8 · M12 |
| **CR 702 — Keyword Abilities** | teilweise | `internal/domain/schluesselfaehigkeit.go` (702.1), `internal/domain/kampf.go` (702.3b, 702.9b, 702.17b, 702.20b, 702.111b), `internal/domain/kampfschaden.go` (702.19b, 702.15b, 702.2b), `internal/domain/erstschlag.go` (702.7b, 702.4b), `internal/domain/zustandsobjekte.go` (702.12b) | **Zwölf von 194 Fähigkeiten** (702.2 „Deathtouch" bis 702.195 „Storied"; 702.1 ist die Einleitung), 771 Regeln — das größte Kapitel der CR. **Welche zwölf, ist gezählt** und nicht gefühlt: über die zehn kanonischen Sets, 24.422 Karten mit Regeltext (`docs/SCHLUESSELWOERTER.md`). **Und gelesen wird die Zeile** (702.1) und nicht das Wort — was die Teilzeichenkette aus `M6` kostete, ist gemessen: bei `flying` 45 Prozent zu viel. **Fünfundvierzig weitere meldet der Tisch als Befund**; die übrigen 137 kamen in der Stichprobe nicht vor und stehen in keiner Liste | M14 · M8 |
| **CR 703 — Turn-Based Actions** | offen | — | **Fünfzehn zugbasierte Handlungen** (703.4a–q). G9 baut vier davon und nennt das Kapitel nicht | G9 · M6 |
| **CR 704 — State-Based Actions** | teilweise | `internal/domain/zustandsaktionen.go:47` (die Runde, 704.3/704.4), `internal/domain/zustandsobjekte.go` (704.5d/f/g/i/v/w), `internal/domain/zustandsanhaengsel.go` (704.5m/n/p), `internal/domain/objekteigenschaften.go` (704.5q) | **Fünfzehn von fünfundzwanzig** — 704.5a bis **704.5aa** (die Buchstaben `l` und `o` fallen aus, danach geht es doppelt weiter). Gebaut: a, b, c, d, f, g, **h**, i, **j**, m, n, p, q, v, w. **704.5h kam mit `M14`**: Schaden aus einer Quelle mit Todesberührung zerstört, gleich wie viel — und Unzerstörbarkeit hält gegen 704.5g **und** 704.5h (CR 702.12b). **Sie laufen, wann immer jemand Priorität bekäme** (704.3), **gleichzeitig** (704.4) und **so oft, bis keine mehr anliegt** — eine sterbende Kreatur lässt in derselben Prüfung ihre Aura fallen. **704.5j ist die einzige mit einer Wahl, und sie trägt seit `M12`**: die Prüfung **stellt** die Frage (`Partie.Legendenwahlen`), und `spieler.legende` beantwortet sie — den Weg, jemanden zu fragen, gab es vorher nicht. **Die zehn offenen hängen je an einem Werkstück**: 704.5e (`M18`), 704.5k (ein **Zeitstempel**, CR 613.7, bei `M10`), 704.5r (`M8`), 704.5s/t/z (`M8`), 704.5u/aa (`M14`), 704.5x/y (`M6`). #1593 schreibt „704.5a–y" — die Zahl stimmt, die Spanne nicht | M5 · M6 · M8 · M10 · M12 · M14 · M18 |
| **CR 705 — Flipping a Coin** | teilweise | `internal/service/zufallsbefehle.go` (705.1, 705.2) | **Zwei Seiten** (705.1) und **wer ruft, gewinnt oder verliert** (705.2) — beides steht im Ereignisstrom, und der hängt an der Hashkette (G1): ein nachträglich geänderter Wurf bricht sie. **Die Quelle ist dieselbe wie beim Mischen.** **Offen**: 705.3 (ein Effekt setzt das Ergebnis) — das ist ein Kartentext (`M8`) | M15 · M8 |
| **CR 706 — Rolling a Die** | teilweise | `internal/service/zufallsbefehle.go` (706.1, 706.2) | **Die Anweisung nennt Seitenzahl und Zahl der Würfel** (706.1), und die Augen stehen **einzeln** da (706.2, 706.4) — eine Summe wäre schon eine Auswertung. **Offen**: Veränderer (706.2), Ergebnistabellen (706.3), der Planarwürfel (706.5ff.) und alles, was ein Kartentext mit den Augen tut | M15 · M8 |
| **CR 707 — Copying Objects** | teilweise | `internal/domain/kopieren.go` (707.2, 707.3), `internal/domain/zonenwechsel.go` (707.4 mit 400.7) | **Ein Objekt wird zur Kopie eines anderen seit `M18`** (707.2): es nimmt die **kopierbaren Werte** an, und die sind hier genau die **gedruckten** — ohne Effekte gibt es keine anderen. **Eingefroren und nicht verwiesen**, wie 707.3 es beschreibt: der Eigner ändert sich nicht (CR 108.3), und ein Verweis auf die Karte der Vorlage ginge im Deckabzug des Falschen ins Leere. **Der Zustand bleibt der Kopie** (707.3 mit 110.5) — getappt, Marken, Beherrscher. **Und eine Kopie endet mit der Zone** (707.4 mit 400.7). **Was fehlt**: die Wahlen eines Zaubers auf dem Stapel (Modus, X, Ziele — 707.2 nennt sie, und sie brauchen `M8`), Kopien mit Ausnahmen (707.9ff.) und der Spielsteinkopiereffekt (707.8). 33 Regeln insgesamt | M18 · M8 |
| **CR 708 — Face-Down Spells and Permanents** | teilweise | `internal/service/sandkasten.go:182`, `internal/domain/objekteigenschaften.go` (708.2a über `VerdeckteEigenschaften`) | Ein Objekt lässt sich umdrehen, und verdeckt sieht es niemand. **Was eine verdeckte bleibende Karte *ist*, weiß der Kern seit `M12`**: eine 2/2 ohne Namen, ohne Typ, ohne Kosten (708.2a) — **auf dem Schlachtfeld**, denn eine verdeckte Handkarte bleibt die Karte, die sie ist. **Was fehlt**: der andere Fall aus 708.2 — eine Fähigkeit, die beim Umdrehen Eigenschaften **nennt** (Morph 702.37, Manifest 701.34) — und 708.2b | M12 · M14 |
| **CR 709 — Split Cards** | offen | — | `frontend/src/bausteine/deckstatistik.ts:60` sagt es ausdrücklich: eine geteilte Karte trägt nach 709.4 die Merkmale **beider** Hälften, gezählt wird die vordere | M12 |
| **CR 710 — Flip Cards** | offen | — | — | M12 |
| **CR 711 — Leveler Cards** | offen | — | Die Levelstufen sind ein Kartentext mit Marken und brauchen `M8` und `M14`. **Nicht zu verwechseln mit dem Umwandeln** — das ist 701.27 und 712; der erste Entwurf von `M12` legte es hier ab, weil 712 danebensteht | M8 · M14 |
| **CR 712 — Double-Faced Cards** | teilweise | `internal/domain/kartengestalten.go` (712.9 über `Umwandeln`), `internal/domain/objekteigenschaften.go` (712.8a), `internal/domain/kartenseiten.go:39`, `frontend/src/bausteine/deckstatistik.ts:65` | **Umgewandelt wird seit `M12`**: `objekt.umwandeln` dreht ein Permanent auf seine andere Seite (701.27a), und nur, wo es eine zweite gibt (712.9). **712.8a trägt für die Eigenschaften und für die Statistik**: außerhalb von Schlachtfeld und Stapel gelten die der Vorderseite. *Hier stand `712.4` — die gehört zu den Meld-Karten.* `HatZweiteSeite` erkennt die zweite Seite an einem nichtleeren `name 2` — eine Aussage über die **Vorlage**, nicht über den Spielzustand. **Was fehlt**: der modale Fall (712.3, 712.11b), Meld (712.4, 712.21) und die Prüfung der Rückseite (712.10). 58 Regeln insgesamt | M12 · M8 |
| **CR 713 — Substitute Cards** | trifft nicht zu | — | **Eine Ersatzkarte ist ein physisches Hilfsmittel** für eine doppelseitige Karte am Tisch (713.1). In einem gezeichneten Tisch gibt es nichts umzudrehen. **M12 nennt „708–713" und nimmt sie damit ohne Not mit** | — entfällt (kein Tischbegriff) |
| **CR 714 — Saga Cards** | offen | — | Kapitelmarken, eine zugbasierte Handlung (703.4f) | M12 |
| **CR 715 — Adventurer Cards** | offen | — | — | M12 |
| **CR 716 — Class Cards** | offen | — | — | M12 |
| **CR 717 — Attraction Cards** | offen | — | Nichttraditionelle Karten mit eigenem Deck | M12 |
| **CR 718 — Prototype Cards** | offen | — | — | M12 |
| **CR 719 — Case Cards** | offen | — | — | M12 |
| **CR 720 — Omen Cards** | offen | — | — | M12 |
| **CR 721 — Station Cards** | offen | — | — | M12 |
| **CR 722 — Preparation Cards** | offen | — | — | M12 |
| **CR 723 — Controlling Another Player** | offen | — | **Ein Effekt steuert einen anderen Spieler** — er braucht `M8`, und dieses Programm hat keinen Weg, für jemanden zu entscheiden | M20 · M8 |
| **CR 724 — Ending Turns and Phases** | teilweise | `internal/domain/abkuerzung.go:98` (724.1b–d) | **Der Zug endet vorzeitig**: alles auf dem Stapel wird exiliert (724.1b), die zustandsbasierten Aktionen laufen ohne Priorität (724.1c), und das Spiel springt in den **Aufräumschritt** (724.1d). **Ausgelöst wird es von einem Menschen** (`zug.beenden`) und nicht von einer Karte — Karten, die etwas tun, gibt es nicht (`M8`). 724.1a (wartende Auslöser) braucht `M9`, der zweite Teil von 724.1d einen Effekt im Aufräumschritt (`M10`), 724.2 (die Kampfphase beenden) den Kampf (`M6`) | M1 · M6 · M9 · M10 |
| **CR 725 — The Monarch** | teilweise | `internal/domain/bezeichnungen.go` (725.1, 725.3, 725.4), `internal/domain/ausloeser.go` (725.2) | **Die Bezeichnung gibt es seit `M20`**: ein Sitz wird Monarch (725.1), **nur einer zugleich** (725.3), und geht er, bekommt sie der aktive Spieler — oder, geht der selbst, der Nächste in Zugreihenfolge (725.4). **Und seit `M9` geschieht sie**: der Monarch zieht zu Beginn seines Endschritts, und wer ihm mit einer Kreatur Kampfschaden zufügt, wird selbst Monarch (725.2, beide). 725.5 (ein Dauereffekt ohne Monarch) braucht `M10` | M20 · M9 · M10 |
| **CR 726 — The Initiative** | teilweise | `internal/domain/bezeichnungen.go` (726.1, 726.3, 726.4), `internal/domain/ausloeser.go` (726.2, zweite) | **Dieselbe Gestalt wie der Monarch** und dieselben drei Regeln: die Bezeichnung, nur einer zugleich, und der Weg, wenn ihr Träger geht. **Eine der drei Fähigkeiten aus 726.2 trägt seit `M9`**: wer dem Träger mit Kreaturen Kampfschaden zufügt, nimmt die Initiative — **je Beherrscher einmal** und nicht je Kreatur, so steht es dort. **Die anderen beiden fehlen**, und beide aus demselben Grund: sie stoßen in den Undercity vor (701.49), und ein Verlies ist eine Karte (CR 714, `M8`) | M20 · M9 · M8 |
| **CR 727 — Restarting the Game** | offen | — | Eine Karte (Karn Liberated) | M20 |
| **CR 728 — Rad Counters** | teilweise | `internal/domain/bezeichnungen.go` (728.1), `internal/domain/sitz.go`, `internal/domain/ausloeser.go` (728.1, die Fähigkeit) | **Die Marken liegen seit `M20` am Spieler** (728.1 mit CR 122) — dieselbe Gestalt wie die Vergiftungsmarken (CR 122.1f), und unter null gibt es keine. **Und seit `M9` mahlen sie**: zu Beginn der ersten Hauptphase, so viele Karten wie Marken, ein Leben und eine Marke je gemahlenem **Nichtland**. **Was 728.1 offen lässt, ist nichts** — das Kapitel hat zwei Regeln, und 728.2 ist die Auskunft, dass Marken sich nicht selbst entfernen | M20 · M9 |
| **CR 729 — Subgames** | offen | — | Eine Karte (Shahrazad), 16 Regeln | M20 |
| **CR 730 — Merging with Permanents** | offen | — | Mutate (702.140) | M20 |
| **CR 731 — Day and Night** | teilweise | `internal/domain/bezeichnungen.go` (731.1, 731.2a–c), `internal/domain/prioritaet.go` (502.2) | **Der Wechsel läuft seit `M20`**: eine Partie fängt mit **keinem** von beidem an (731.1), und im Enttappschritt (CR 502.2) wird der vorige Zug geprüft — Tag wird Nacht ohne Zauber (731.2a), Nacht wird Tag ab **zwei** (731.2b), und ohne Tageszeit findet die Prüfung gar nicht statt (731.2c). **Gezählt werden die Zauber des aktiven Spielers** (`Zugstand.Zauber`), und die Zahl geht mit dem Zug auf null. **Was fehlt**: wie es überhaupt Tag wird (731.3ff. — das sagt eine Karte, `M8`) und die Mehrspielervariante aus 731.2a/b mit gemeinsamen Zügen | M20 · M8 |
| **CR 732 — Taking Shortcuts** | teilweise | `internal/domain/abkuerzung.go:39` (732.1, 732.2a) | **Der Befehl `zug.abkuerzen` reicht durch bis zum Ende des Zuges** und hält an, sobald etwas geschieht: der Stapel ist nicht leer (405.5), jemand muss abwerfen (514.1), die Partie endet. Ohne ihn waren es **zwanzig Passen je Zug** bei zwei Sitzen. **732.2b fehlt**: der Vorschlag wird ausgeführt und nicht verhandelt — am gemeinsamen Tisch verabreden sich zwei Menschen mit einem Satz, über das Netz kann das Gegenüber nicht dazwischengehen. 732.3 (zerfallende Schleifen) und 732.4 (eine Pflichtschleife ist unentschieden) offen | M1 · M13 |
| **CR 733 — Handling Illegal Actions** | offen | `internal/service/zuruecknahme.go` (nennt 733.1, **um sie nicht zu benutzen**) | Zwei Regeln — das kleinste Kapitel mit eigenem Werkstück. **Die Rücknahme aus P8 stützt sich nicht darauf** (#2376, Vermerk 0197): 733.1 gilt für regelwidrige Handlungen, die Rücknahme nimmt regelgerechte zurück und ist eine Absprache des Tisches | M4 · P8 |

### 4.8 Multiplayer Rules — CR 800–811

**M13 nennt „CR 104, 800–809".** Das lässt zwei Kapitel offen, und die
Abgrenzung von #1593 sagt dazu „Zwei Spieler und der Mehrspielerfall aus
CR 800–809 sind der Umfang" — die Ausgrenzung ist also gewollt. Sie steht
trotzdem hier, weil **CR 122.1f für Vergiftungsmarken ausdrücklich auf CR 810
verweist**.

| Regel | Stand | Beleg | Grenze | Werkstück |
| --- | --- | --- | --- | --- |
| **CR 800 — General** | teilweise | `internal/domain/spielende.go:209` (800.4a, die Objekte), `internal/domain/gehen.go:23` (800.4a, die Priorität; 800.4j), `internal/domain/prioritaet.go` (800.4k in `naechsterSitz`, 800.4j in `prioritaetFuer`, CR 117.4 in `Passen`), `internal/domain/kampfschaden.go` (800.4e) | **Fünf Regeln von 21**, gelesen in der Fassung 2026-09-25: eine Mehrspielerpartie läuft weiter, wenn jemand geht (800.4); wer geht, nimmt seine Objekte mit (800.4a, **nach Eigner und nicht nach Beherrscher**) und **gibt die Priorität an den Nächsten in Zugreihenfolge, der noch im Spiel ist** (800.4a, seit #1971); **Kampfschaden an ihn fällt aus** (800.4e); geht der aktive Spieler in seinem Zug, **läuft der Zug ohne ihn zu Ende** — wo er Priorität bekäme, bekommt sie der Nächste, und eine Abwurfpflicht geht mit ihm (800.4j, seit #1971); und sein nächster Zug beginnt nicht, die Zugreihenfolge überspringt ihn (800.4k). „Alle passen nacheinander" zählt nur die, die noch dabei sind (CR 117.4) — seit #1971 auch ein Passen von vor dem Gehen. **Das ist die härteste Stelle**: eine Reihenfolge, die einen Gegangenen mitzählt, gibt ihm Priorität, und niemand kann sie ihm nehmen — die Partie hängt. **Bis #1971 hing sie trotzdem**, wenn der Gehende die Priorität **im Augenblick des Gehens** hielt; `FuzzZugfolge` fährt seitdem drei Sitze. **800.4b–d und 800.4f–i fehlen**: sie sprechen von Effekten, Auslösern und Kontrollwechseln durch Effekte (`M8`, `M9`, `M10`) | M13 · M8 · M9 · M10 |
| **CR 801 — Limited Range of Influence Option** | offen | — | 28 Regeln. **Eine Option, die vor dem Spiel gewählt wird** (CR 800.2), und dieses Programm bietet keine an — das steht seit `M13` als Grenze in der Regelmatrix. Sie gilt immer im Emperor (CR 809) und wird bei fünf und mehr Spielern verwendet; das Profil lässt vier zu | **— fehlt (ausgegrenzt)** |
| **CR 802 — Attack Multiple Players Option** | offen | — | **Eine Option** (CR 800.2), und zwar die Vorgabe des Free-for-All. Der Kern lässt jeden Angreifer ein eigenes Ziel nennen (CR 508.1a) und **kennt keine Einschränkung** — was CR 802 zusätzlich regelt, sind die Fälle mit Optionen daneben | **— fehlt (ausgegrenzt)** |
| **CR 803 — Attack Left and Attack Right Options** | offen | — | **Eine Option** (CR 800.2); dieses Programm bietet keine an, und das steht seit `M13` als Grenze in der Regelmatrix | **— fehlt (ausgegrenzt)** |
| **CR 804 — Deploy Creatures Option** | offen | — | **Eine Teamoption** (CR 800.2), und Teams gibt es nicht — CR 102.3 steht seit G9 als offen | **— fehlt (ausgegrenzt)** |
| **CR 805 — Shared Team Turns Option** | offen | — | 25 Regeln. **Eine Teamoption** (CR 800.2), und Teams gibt es nicht | **— fehlt (ausgegrenzt)** |
| **CR 806 — Free-for-All Variant** | teilweise | `internal/domain/grundprofil.go`, `internal/domain/prioritaet.go` | **Das ist die Variante, die dieses Programm spielt** — sie ist der Grundfall einer Mehrspielerpartie, und seit `M13` trägt sie: bis zu vier Sitze, reihum, und wer geht, wird übersprungen (CR 800.4). **Was fehlt**: 806.3 (die Plätze werden zufällig vergeben — hier sagt sie der, der anlegt) und die Optionen aus 806.2, die es alle nicht gibt | M13 |
| **CR 807 — Grand Melee Variant** | offen | — | Ab zehn Spielern; das Profil lässt vier zu (`internal/domain/grundprofil.go:73`) | M13 |
| **CR 808 — Team vs. Team Variant** | offen | — | — | M13 |
| **CR 809 — Emperor Variant** | offen | — | — | M13 |
| **CR 810 — Two-Headed Giant Variant** | offen | — | 27 Regeln. **CR 122.1f verweist darauf** | **— fehlt (ausgegrenzt)** |
| **CR 811 — Alternating Teams Variant** | offen | — | — | **— fehlt (ausgegrenzt)** |

### 4.9 Casual Variants — CR 900–905

| Regel | Stand | Beleg | Grenze | Werkstück |
| --- | --- | --- | --- | --- |
| **CR 900 — General** | trifft nicht zu | — | Die Einleitung zu den Varianten; #1593 grenzt alle bis auf Commander aus | — |
| **CR 901 — Planechase** | trifft nicht zu | — | Ausgegrenzt | — |
| **CR 902 — Vanguard** | trifft nicht zu | — | Ausgegrenzt | — |
| **CR 903 — Commander** | teilweise | `internal/domain/commander.go`, `internal/domain/deckurteil.go` (903.5a/b/e) | **Vier Regeln tragen seit `M19`**: vierzig Leben (903.7), die **Steuer** von zwei generischem Mana je vorherigem Wirken aus der Kommandozone (903.8), **einundzwanzig Kampfschaden desselben Commanders** als zustandsbasierte Aktion (903.10a) und der **Weg zurück** in die Kommandozone (903.9a — ein Dürfen, und deshalb ein Befehl). Dazu die Deckbauzahlen aus Phase 4 (903.5a/b/e). **Was fehlt**: die **Farbidentität** (903.4), ob die genannte Karte legendär ist (903.3), die Partner- und Hintergrundvarianten (903.3a–d) und die Fähigkeit mancher Commander (903.6, `M8`). 30 Regeln insgesamt | M19 · M8 |
| **CR 904 — Archenemy** | trifft nicht zu | — | Ausgegrenzt | — |
| **CR 905 — Conspiracy Draft** | trifft nicht zu | — | Ausgegrenzt | — |

---

## 5 Sechs Befunde, und jeder hat eine Zeile

**Das sind die Stellen, an denen eine Zusage weiter reicht als der Code.** Sie
stehen hier zusammen, weil sie das Eigentliche sind, was diese Datei gefunden
hat — die Tabelle oben ist ihr Rahmen.

### Befund 1 — Drei Bereiche stehen auf `automatisch`, und CR 400.7 hält keiner

`internal/service/sandkastenschieben.go:38` setzt beim Zonenwechsel `Zone` und
`Ordnung` und ruft `zustaendeSetzen` für `Getappt` und `Verdeckt`. **`Zaehler`
und `HaengtAn` bleiben unberührt.**

Damit überlebt eine +1/+1-Marke die Reise vom Tisch in den Friedhof, und eine
Ausrüstung bleibt an einer Kreatur hängen, die es nicht mehr gibt. **CR 400.7**
(Fassung 2026-08-19) sagt: ein Objekt, das die Zone wechselt, wird ein neues
ohne Erinnerung an sein voriges Dasein. **CR 122.2** sagt es für Marken noch
einmal ausdrücklich: sie werden nicht entfernt, sie hören auf zu bestehen.

`internal/domain/grundprofil.go` führt **zonen**, **marken** und **anhaengsel**
als `automatisch` und nennt je zwei Grenzen. Keine davon ist diese. Das ist
genau die Gestalt aus §10: ein Kopf, der mehr verspricht, als sein Rumpf hält.

**Der Kommentar an `schieben` kennt die Frage** — „dass ein Objekt beim
Zonenwechsel nach den Regeln ein **neues** wird, ist eine Regelfrage und gehört
in die Matrix von G4". Sie ist dort nur nicht angekommen.

### Befund 2 — Der Friedhof hat zwei Antworten auf dieselbe Regel

`internal/domain/zonenordnung.go:31` führt `ZoneFriedhof` unter den
**ungeordneten** Zonen, mit der Begründung, in einem Friedhof liege alles in
irgendeiner Lage.

`frontend/src/tisch/zonenfeld.ts:69` führt ihn unter den **geordneten**, und
zwar unter ausdrücklicher Berufung auf **CR 400.5**.

**Die Regel gibt der Anzeige recht.** CR 400.5 (Fassung 2026-08-19) nennt drei
Zonen, deren Reihenfolge sich nur ändern lässt, wenn eine Regel oder ein Effekt
es erlaubt: Bibliothek, **Friedhof** und Stapel.

Was daran hängt, ist nicht theoretisch:

- `objekt.schieben` setzt für jede Karte im Friedhof `Ordnung = 0` und ruft
  `einsortieren` gar nicht erst auf (`internal/service/sandkastenschieben.go:66`).
  **Alle Karten im Friedhof liegen an derselben Stelle.**
- `gemeinteStelle` weist „oben" für den Friedhof ab
  (`internal/service/sandkastenschieben.go:105`) — dabei ist „oben auf den
  Friedhof" genau das, was CR 404.1 beschreibt.
- **CR 404.3** (der Eigner ordnet, was gleichzeitig hineinkommt) hat nichts,
  worauf es zeigen könnte.

### Befund 3 — `Startleben` steht da und niemand liest es

`internal/domain/grundprofil.go:83` setzt `Startleben: 20`, belegt mit CR 103.4
und mit einer sorgfältigen Notiz darüber, dass zwei frühere Nummern erfunden
waren.

**Gesucht am 2026-09-16 über `internal/`, `frontend/src/` und `cmd/`, ohne
Testdateien:** `Startleben` kommt in `internal/domain/regelprofil.go` (die
Definition) und `internal/domain/grundprofil.go` (die Zuweisung) vor, und sonst
nirgends. Dasselbe für `HausregelStartleben`: geprüft wird, dass eine Zahl
dabeisteht (`internal/domain/hausregel.go:86`) — **angewandt wird sie nie.**

`domain.Sitz` (`internal/domain/partie.go:102`) führt Kennung, Name und Deck.
Kein Leben.

**Das ist bekannt und in G9 vermerkt**; es steht hier, weil es die Bauform ist,
nach der diese Datei gesucht hat, und weil es nicht der einzige Fall ist —
siehe Befund 4.

### Befund 4 — Spielsteine und Embleme gibt es als Wort

`internal/domain/spielobjekt.go:45` und `:48` erklären `ObjektSpielstein` und
`ObjektEmblem`. **Gesucht über `internal/` und `cmd/` ohne Testdateien**: beide
Konstanten kommen außerhalb ihrer eigenen Datei **nirgends** vor. Im Fenster
tauchen sie nur in `frontend/src/test/tischszenen.ts` auf, also in einer
Vorrichtung.

Für `spielsteine` ist das in der Matrix richtig abgebildet — der Bereich steht
auf `manuell`. Für CR 114 (Embleme) gibt es gar keinen Bereich. Beide Kapitel
stehen deshalb als `teilweise` mit ausdrücklicher Grenze und nicht als `offen`:
wer nur nach der Objektart suchte, fände etwas und zöge den falschen Schluss.

### Befund 5 — CR 400.3 steht in der Anzeige und nicht im Zustand

`frontend/src/tisch/tischplan.ts:110` rechnet für Bibliothek, Friedhof und Hand
mit dem **Eigner** und beruft sich auf CR 400.3. Das ist richtig und es ist
**eine Aussage darüber, wo etwas gezeichnet wird**.

Im Zustand behält ein Objekt beim Zonenwechsel seinen `Beherrscher`
(`internal/service/sandkastenschieben.go:38`). Wer eine fremde Kreatur
beherrscht und sie sterben lässt, hat sie danach in **seinem** Friedhof stehen
— das Bild zeigt sie beim Eigner, der Zustand führt sie beim Beherrscher.
Solange nichts die Kontrolle wechselt (CR 110.2a fehlt, siehe CR 110), fallen
die beiden nie auseinander. **Mit M12 fallen sie auseinander.**

### Befund 6 — Vier Zahlen in #1593 sind um eins daneben oder falsch geschrieben

Nachgezählt über den Rumpf der Textfassung (Verfahren in Abschnitt 2):

| #1593 sagt | Nachgezählt | Warum |
| --- | --- | --- |
| „195 Schlüsselwortfähigkeiten (CR 702)" | **194** | 702.1 ist die Einleitung ohne eigene Fähigkeit, das erste Schlüsselwort ist 702.2 „Deathtouch", das letzte 702.195 „Storied". 195 nummerierte Regeln, 194 Fähigkeiten. Keine Lücke in der Nummerierung |
| „70 Schlüsselworthandlungen (CR 701)" | **69** | Dasselbe: 701.1 ist die Einleitung, 701.2 „Activate" bis 701.70 „Recruit" — **in der Fassung 2026-08-19**; seit der vom 2026-09-25 sind es siebzig |
| „25 zustandsbasierte Aktionen (CR 704.5a–y)" | **25, aber a–aa** | Die Zahl stimmt. Die **Spanne** nicht: die CR lässt `l` und `o` aus und geht nach `z` mit `aa` weiter. „a–y" wären 23 |
| „30 Schichtregeln (CR 613)" | **30 Unterregeln, 7 Schichten, 41 Regeln** | Die Zahl ist die der Unterregeln von 613 und damit richtig gemessen — „Schichtregeln" liest sich aber wie „Schichten", und derer sind es sieben (613.1a–g) |

**Keine davon ist ein Fehler, der etwas kaputtmacht.** Sie stehen hier, weil
eine Zahl in einem Issue unter keinem Wächter steht (§16 erzählt, wie eine
Bindungszahl in den Leitlinien jahrelang falsch stand) und weil der Zuschnitt
von M14 an der ersten hängt.

---

## 6 Was das Grundprofil heute zusagt, und was daraus wird

### Heute

`internal/domain/grundmatrix.go` führt **zehn von dreizehn** Bereichen als
`automatisch`: Zonen (CR 400), Marken (CR 122), Anhängsel (CR 301/303),
Zugstruktur (CR 500ff.), Priorität (CR 117), Stapel (CR 405/601/608), Kampf
(CR 506–511), zustandsbasierte Aktionen (CR 704), Mana (CR 106/305.6/500.5)
und Wechselwirkungen (CR 603/608/707). **Zwei stehen auf `begleitet`** — die
Typen (CR 205ff.) und die Spielsteine samt den Handlungen aus CR 701 —, und
**einer auf `manuell`**: `commander`. Der Kopf der Datei sagt dazu den
richtigen Satz: „Buchführung, nicht Regelkenntnis."

**`commander` ist dort mit Absicht `manuell`.** Die Variante hat ein
**eigenes** Profil (`internal/domain/commander.go`, `Commanderfassung` 1), und
in dem steht derselbe Bereich auf `automatisch` mit vier Grenzen. Eine
Sandkastenpartie ist keine Commanderpartie; ein Profil, das es trotzdem
zusagte, wäre die Gestalt aus §10.

`Regelmatrix.TraegtStrenge` rechnet daraus aus, dass `begleitet` zur Wahl steht
und `streng` nicht (`internal/domain/regelmatrix.go`). Das ist nachprüfbar
statt behauptet, und es ist die gute Hälfte von G4.

**Und seit dem 2026-09-19 zählt es nur, was zur Variante gehört**
([0123](entscheidungen/0123-streng-zaehlt-nur-was-zur-variante-gehoert.md),
#1626).
**Hier lag ein Befund**: `Luecken()` ging über **jeden** Bereich, also auch
über `commander` — und damit konnte `streng` für das Grundprofil **nie** wahr
werden, gleich wie viel gebaut wird. Eine Stufe, die niemand erreichen kann,
ist keine Stufe.

**Gemessen am 2026-09-18** waren es drei Lücken: `spielsteine`, `typen` und
`commander`. **Heute sind es zwei** — und beide sind Arbeit:

| Bereich | Warum er `begleitet` bleibt |
| --- | --- |
| `spielsteine` | neun Grenzen; die tragende ist „keine Handlung geschieht von selbst — sie wird gestellt" (CR 608.2c). Zehn der neunundsechzig Handlungen aus CR 701 gehen, wenn jemand sie stellt. Dazu fehlen die Kartengestalten aus CR 309, 714, 716–721 und die genannten Werte bei Morph und Manifest (CR 702.37, CR 701.34) |
| `typen` | vier Grenzen; die tragenden sind CR 205.3b–m (welcher Untertyp zu welchem Typ gehört), CR 205.2b/c (die Reihenfolge) und **CR 310 (Battles)** — vierundzwanzig Regeln, die kein Werkstück nennt. Dazu Schicht 4 aus CR 613: kein Effekt ändert heute einen Typ |

**`commander` zählt hier nicht mehr mit**, und das ist keine Nachsicht: er
steht auf `manuell`, weil eine Sandkastenpartie die Variante nicht spielt.
Ein Bereich, den niemand schließen soll, ist keine Lücke. **Im
Commanderprofil zählt er**, denn dort steht er auf `automatisch`.

### Was diese Datei daraus gemacht hat

**Hier stand eine Forderung, und sie ist erledigt.** Sie lautete: die drei
`automatisch` von G4 halten nicht, was ihr Grad verspricht — nicht, weil zu
wenig gebaut wäre, sondern weil ihre **Grenzen** unvollständig sind. Drei
fehlten:

| Bereich | Fehlende Grenze | Was daraus wurde |
| --- | --- | --- |
| `zonen` | CR 400.7 (ein Objekt wird beim Zonenwechsel ein neues), CR 400.3 (fremde Zonen), CR 404.3/400.5 (die Reihenfolge im Friedhof) | **`M18` hat 400.7 gebaut** statt eingegrenzt, und `M18` mit G9 die Ordnung im Friedhof (404.1). Was **bleibt**, steht jetzt als Grenze da: die Kennung überlebt den Wechsel, und die zwölf Ausnahmen aus 400.7a–m fehlen. **400.3 ist offen** und in Abschnitt 5 benannt |
| `marken` | CR 122.2 (Marken hören beim Zonenwechsel auf zu bestehen) | **`M18` hat sie gebaut** — 122.2 fällt mit 400.7 |
| `anhaengsel` | CR 400.7 (ein Anhängsel löst sich, wenn sein Ziel die Zone wechselt) | **`M18` hat es gebaut** — dieselbe Stelle, `Partie.zonenwechsel` |

**Die Forderung war „drei zusätzliche Zeilen in `Grenzen`", und die Antwort war
besser als die Forderung.** `automatisch` heißt „das Programm setzt es durch";
eingegrenzt hätte die Zusage ehrlich gemacht, gebaut hat sie eingelöst. Was
heute nicht geht, steht daneben — sonst wäre es die Gestalt aus §10, und die
ist hier schon einmal teuer gewesen.

**Und `Grundprofilfassung` ist dabei gestiegen**
(`internal/domain/grundprofil.go:41`), weil sich an der Matrix etwas geändert
hat — inzwischen auf **20**. Eine laufende Partie bleibt bei der Fassung, unter
der sie angefangen hat.

### Ab wann `Regelmatrix.TraegtStrenge(StrengeStreng)` wahr werden kann

`TraegtStrenge` verlangt für `streng`, dass **jeder** der dreizehn Bereiche
`automatisch` ist (`internal/domain/regelmatrix.go:235`). Was jeder Bereich
dafür braucht, steht hier — und es ist **mehr als die sechzehn Werkstücke von
5.8**:

| Bereich | Braucht | Steht in |
| --- | --- | --- |
| `zonen` | CR 400 vollständig — **400.7 trägt seit `M18`**, offen bleiben 400.3 (eine fremde Kreatur stirbt in den Friedhof ihres Eigners), 400.6 und 400.9–400.11 | M18 · **400.3: kein Werkstück** |
| `zugstruktur` | CR 500–505, 511–514 | G9 |
| `stapel` | CR 405, 601, 608 — **seit `M4` automatisch**; die längste offene Zusage war, dass niemand ausführt, was auf einem Zauber **steht** (CR 608.2c). **Seit `M8` führt er es aus** — wenn die Karte ein `game script` trägt | M1 · M4 · M8 |
| `prioritaet` | CR 117, 500.2 | G9 · M1 |
| `combat` | CR 302.6, 506–511, 703.4g–m — **seit `M6` automatisch**; die längste offene Zusage ist, dass kein Effekt das Angreifen oder Blocken einschränkt (CR 508.1c, CR 509.1b) | M6 · M8 · M14 |
| `zustandsbasiert` | CR 704 — **fünfzehn der 25**: dreizehn seit `M5`, dazu 704.5j (die **Wahl**) seit `M12` und 704.5h (Todesberührung) seit `M14`; **jede der zehn offenen hängt an einem benannten Werkstück**. Hier stand „vierzehn" und „elf" — seit `M14` falsch, gefunden bei PLAY-01 | M5 · M6 · M8 · M10 · M14 · M18 |
| `spielsteine` | CR 111, 701 — **seit `M7` begleitet**: zehn der neunundsechzig Handlungen gehen, wenn jemand sie stellt — acht seit `M7`, dazu das Erzeugen eines Spielsteins (CR 701.7a) und das Umwandeln (CR 701.27a) seit `M12` —, und von selbst geschieht keine. Hier stand „acht" und „erzeugt wird kein Spielstein" — seit `M12` falsch, gefunden bei PLAY-01 | M7 · M8 · M12 |
| `marken` | CR 122 vollständig — **122.1a trägt seit `M5`, 122.2 seit `M18`**; 122.1b–j und 122.4–122.8 nennen Marken, deren Wirkung auf einer Karte steht | M5 · M18 · M14 · M8 |
| `anhaengsel` | CR 301.5, 303.4, 701.3 („Attach"), 704.5m/n | M7 · M5 |
| `typen` | CR 205, 300–310, und Schicht 4 aus 613 | M2 · M10 · M6 (CR 310.2–310.11) |
| `mana` | CR 106, 118, 305.6, 500.5, 605 — **der zwölfte Bereich, seit `M3`**; die längste offene Zusage ist, dass niemand **bezahlt** (CR 601.2f–h) | M3 · M4 · M8 |
| `wechselwirkungen` | CR 603, 608 **und CR 707** — **seit `M9` automatisch**: eine Fähigkeit löst aus, wartet, wird geordnet, kommt auf den Stapel und hört nach dem Verrechnen auf zu existieren. Die längste offene Zusage ist dieselbe wie bei `stapel`: niemand führt aus, was auf einem Zauber **steht** (CR 608.2c) | M1 · M9 · M8 |
| `commander` | CR 903 — **im Grundprofil `manuell`, und zwar mit Absicht**: eine Sandkastenpartie ist keine Commanderpartie. Im **eigenen** Profil (`Commanderfassung` 1) steht der Bereich auf `automatisch` mit vier Grenzen | M19 |

**Und `streng` kann für dieses Profil nie wahr werden**, gleich wie viel noch
gebaut wird: `TraegtStrenge` verlangt **jeden** Bereich, und `commander`
gehört einer Variante, die eine Sandkastenpartie nicht spielt. Das ist keine
fehlende Arbeit, sondern eine Entscheidung —
#1626 trägt
sie, samt der beiden Bereiche, die wirklich Arbeit sind (`spielsteine` und
`typen`).

**Zwei der dreizehn Bereiche nennen heute eine Lücke, die kein Werkstück
schließt**: `zonen` (CR 400.3) und `typen` (CR 310 als Kartentyp, Abschnitt
7.1 — `M6` deckt davon nur 310.2–310.11).
**Drei waren es bis zum 2026-09-17**: `M2`
(#1604) hat
den Kartentyp „battle" gebaut und seine Zahl gelesen, und was CR 310 darüber
hinaus verlangt — Verteidiger, Schaden, Verteidigungsmarken — gehört zum
Kampf und damit zu `M6`.
Solange das so bleibt, kann `streng` nach Abschluss von 5.8 nicht wahr werden —
und die erste Abnahme von #1593 („alle Bereiche") ließe sich nicht einlösen.

---

## 7 Die Kapitel ohne Werkstück

**Vierzig Kapitel nennt kein Werkstück.** Zwei davon sind ausdrücklich
ausgegrenzt (CR 810, 811) — bleiben **achtunddreißig**, die gebaut werden
müssten und in keinem Werkstück stehen. Dazu kommen **drei Kapitel und drei
Regeln**, die ein Werkstück dem Gegenstand nach meint, ohne die Nummer zu
nennen (CR 605 in M3, CR 611 in M10, CR 703 in G9; CR 101.1–101.3, CR 122.2,
CR 400.7). Sie stehen hier nach Dringlichkeit gruppiert, nicht nach Nummer.

### 7.1 Trägt einen Bereich der Regelmatrix — 61 Regeln

| Kapitel | Regeln | Warum |
| --- | ---: | --- |
| **CR 707 — Copying Objects** | 33 | `internal/domain/kopieren.go` baut 707.2 und 707.3 seit `M18`; **was kopiert, steht auf einer Karte** (`M8`) |
| **CR 310 — Battles** | 24 | Ein Kartentyp aus CR 300.1; ohne ihn ist `typen` unvollständig |
| **CR 404 — Graveyard** | 3 | Die Reihenfolge, die CR 400.5 verlangt — `Zone.Geordnet` bejaht sie seit G9, **vergeben wird sie von niemandem** (Befund 2, halb behoben) |
| **CR 210 — Defense** | 1 | Die Zahl, mit der ein Battle das Schlachtfeld betritt |
| **CR 400.7** (Regel in CR 400) | — | Trägt `zonen`, `marken` und `anhaengsel` zugleich |
| **CR 122.2** (Regel in CR 122) | — | Trägt `marken` |

### 7.2 Voraussetzung für ein Werkstück, das es gibt — 116 Regeln

| Kapitel | Regeln | Wofür |
| --- | ---: | --- |
| **CR 604 — Static Abilities** | 8 | M10 ordnet Dauereffekte; erzeugt werden sie hier |
| **CR 609 — Effects** | 12 | M1, M10, M11 handeln alle von Effekten |
| **CR 610 — One-Shot Effects** | 13 | M7 („zerstöre", „opfere") erzeugt genau diese |
| **CR 612 — Text-Changing Effects** | 11 | Schicht 3 von CR 613 (M10) |
| **CR 616 — Interaction of Replacement/Prevention** | 9 | `M11` baut 616.1 über den Zeitstempel statt über die Wahl des Betroffenen |
| **CR 700 — General** | 35 | „als ob", „kann nicht", „bis zu", Zufallsauswahl — die Auslegungsregeln für M7 und M8 |
| **CR 101.1–101.3** | 3 | Die Karte schlägt die Regel — die Grundlage von M8 |
| **CR 607 — Linked Abilities** | 25 | Ein Kartentext, der exiliert und das Exilierte später spielbar macht |

**Drei Kapitel werden gebaut, ohne genannt zu werden**, und stehen deshalb
nicht in der Zählung der vierzig: **CR 605** (Manafähigkeiten, 13 — M3 meint
sie), **CR 611** (Dauereffekte, 13 — M10 meint sie) und **CR 703**
(zugbasierte Handlungen, 20 — G9 baut vier der fünfzehn). Eine Nummer, die
niemand nennt, lässt sich nicht nachschlagen; das ist keine Lücke, aber es
kostet beim nächsten Leser.

### 7.3 Ausdrücklich eingeschlossen und trotzdem ohne Werkstück — 52 Regeln

| Kapitel | Regeln | Warum es auffällt |
| --- | ---: | --- |
| **CR 903 — Commander** | 49 | #1593 nennt Commander als die **benannte Ausnahme** von der Variantenausgrenzung. Gebaut ist heute der Deckbau (903.5a/b/e); 903.7 (40 Leben), 903.8 (Steuer), 903.9, 903.10a (21 Commanderschaden) hat niemand |
| **CR 408 — Command** | 3 | Die Zone gibt es; Embleme und heimgeholte Commander liegen darin (M12, M19) |

### 7.4 Kartentypen und Kartengestalten, die niemand nennt — 123 Regeln

CR 309 (Dungeons, 16), CR 306 (Planeswalkers, 13), CR 714 (Saga, 13), CR 715
(Adventurer, 12), CR 720 (Omen, 12), CR 722 (Preparation, 12), CR 718
(Prototype, 10), CR 717 (Attraction, 9), CR 716 (Class, 8), CR 721 (Station,
7), CR 719 (Case, 6) — **zusammen 118 Regeln**. Dazu **CR 606
(Loyalitätsfähigkeiten, 5)**, ohne die es keinen Planeswalker gibt.

**M12 hört bei 713 auf**, und die Reihe geht bis 722 weiter. Dabei ist das
dieselbe Art von Sache: Karten mit besonderer Gestalt, deren Regeln an dieser
Gestalt hängen.

### 7.5 Spielzustände neben dem Spiel — 101 Regeln

CR 723 (Controlling Another Player, 13), CR 724 (Ending Turns and Phases, 15),
CR 725 (The Monarch, 5), CR 726 (The Initiative, 5), CR 727 (Restarting, 9),
CR 728 (Rad Counters, 2), CR 729 (Subgames, 16), CR 730 (Merging, 18), CR 731
(Day and Night, 6), CR 732 (Taking Shortcuts, 12). **Zusammen 101 Regeln.**

**CR 724 und CR 732 sind die beiden, die nicht warten können.** 724 ist eine
Ausnahme vom Verrechnen und gehört zu M1; 732 entscheidet, ob man mit dem
Ergebnis überhaupt spielen mag — ohne Abkürzungen fragt das Programm bei jedem
Prioritätsdurchlauf zweimal nach.

### 7.6 Ohne Bezug zu einem heutigen Werkstück — 71 Regeln

CR 111 (Tokens, 34 — die Objektart gibt es, niemand erzeugt einen), CR 123
(Stickers, 26 — nirgends genannt), CR 108 (Cards, 11 — nur 108.3 trägt).

### 7.7 Zwei ausdrücklich ausgegrenzt — 35 Regeln

CR 810 (Two-Headed Giant, 27) und CR 811 (Alternating Teams, 8) liegen
außerhalb des Umfangs, den #1593 nennt („CR 800–809"). **CR 122.1f verweist für
Vergiftungsmarken ausdrücklich auf CR 810** — wer Gift baut, stößt darauf. Das
ist kein Grund, den Umfang zu ändern; es ist ein Grund, die Grenze dort zu
nennen, wo Gift gebaut wird.

---

## 8 Was hier nicht steht

**Ob eine als `umgesetzt` oder `teilweise` geführte Stelle richtig ist.** Diese
Datei liest, **ob** an einer Stelle etwas geschieht, nicht **ob es stimmt**.
Für das Zweite gibt es die Tests an dieser Stelle und die Sichtprüfung.

**Der Aufwand.** Ein Kapitel mit 771 Regeln und eines mit einer stehen hier in
derselben Zeile. Die Regelzahl steht je Abschnitt und bei jedem Befund daneben,
damit niemand die Zeilen zählt und daraus einen Plan macht.

**Die Tournament Rules.** Die MTR regelt Decklisten, Sideboarding, Zeitregeln
und Rundenablauf; was davon gilt, steht bei den Turnierwegen und nicht hier.

**Die Kartentexte.** Diese Datei sagt, welche **Regeln** greifen. Ob eine
bestimmte Karte spielbar ist, hängt zusätzlich daran, ob ihre Fähigkeit in
MSE-Script steht — und **eine Bibliothek handgeschriebener Umsetzungen für
30.000 gedruckte Karten ist ausdrücklich nicht Teil dieser Phase** (#1593,
Abgrenzung).

---

## 8 Das Register je Regel

**Stand: 2026-09-22** — PLAY-01
(#1709).
Diese Datei fragt je **Kapitel**; das Register fragt je **Regel**, und es
steht nicht hier, sondern im Code: `domain.Regelregister()`
(`internal/domain/regelregister*.go`). **29 Regeln** über die sieben Gruppen
aus AK-03 — Spielaufbau, Zonenwechsel, Priorität und Stapel, Ziele und
Kosten, Kampf, Zustandsaktionen, das Zusammenspiel von Effekten —, davon
**25 durchgesetzt, 3 teilweise und 1 manuell**.

**Je Regel stehen dort** die Nummer, die Fassung, eine eigene Beschreibung,
der Stand, die Voraussetzungen, die **Stelle** (die Funktion, die die Regel
nennt), der **Aufrufer** (der produktive Einstieg, der dorthin führt) und
zwei Läufe: der zulässige Fall und der Gegenfall.

**Vier Fragen, getrennt geprüft** (AK-02,
`internal/werkzeuge/regelregister_test.go`): gibt es Stelle und Aufrufer,
nennt die Stelle die Regel, erreicht der Aufrufer die Stelle und
`internal/service` oder `internal/app` den Aufrufer, und gibt es die zwei
Läufe. **Beim ersten Lauf hat er zweimal recht gehabt**: für den
Kampfschaden stand `Partie.Kaempfen` im Register — die Schlüsselhandlung
„kämpfen", nicht der Schaden des Kampfschrittes —, und für die Ersetzung
614.1, wo die Stelle 614.1a nennt.

**Und das Fenster sagt „automatisch" nicht mehr ohne Einschränkung.** Jeder
automatische Bereich muss Grenzen nennen (`Matrixeintrag.pruefe`), und das
Wort sagt jetzt, dass sie darunter stehen — die GUI darf nach AK-02 einen
nur teilweise kontrollierten Bereich nicht als ganzen zusagen.

**Eine fehlende Regel verbietet keinen zulässigen Zug** (AK-04). Das
Beispiel im Register ist CR 307.1: den Zeitpunkt eines Hexereizaubers prüft
das Programm nicht, und deshalb weist es auch keinen zulässigen
Schnellzauber ab — der Stand ist `manuell`, und der Lauf belegt, dass ein
zweiter Zauber auf einen vollen Stapel geht.

**Und „sichtbar" heißt: in einer Grenze der Matrix.** Das Register liest am
Tisch niemand; die Grenzen zeigen alle vier Flächen über `Regelfreigabe`.
Deshalb muss jede Regel, die nicht `durchgesetzt` ist, in einer Grenze ihres
Bereichs genannt sein — eine Unterregel deckt ihre Regel, umgekehrt nicht
(`TestJedeNichtDurchgesetzteRegelStehtInDenGrenzen`). **Beim ersten Lauf
fehlte genau 307.1**: dass niemand den Zeitpunkt prüft, stand nur im
Register und in einem Kommentar. Beim Stapel ist es seitdem eine Grenze.

### Eine neue Fassung einspielen

1. **`Grundprofilquelle` heben.** Danach ist **jeder** Eintrag des Registers
   ungültig (`Regeleintrag.Pruefe`), und `TestDasRegisterIstGueltig` nennt
   jeden einzeln.
2. **Je Eintrag nachsehen** — Nummer und Sinn in der neuen Fassung, die
   Beschreibung in eigenen Worten anpassen, dann erst die Fassung des
   Eintrags heben. Ein Ersetzen über alle Einträge umginge genau diesen
   Schritt, und er ist der Grund für das Feld.
3. **Die Kapitel dieser Datei ebenso**, samt der Zahlen in Abschnitt 3.
4. **Kein Wortlaut**, auch nicht vorübergehend:
   `internal/werkzeuge/regelwortlaut_test.go` weist ein englisches Zitat ab,
   und das Dokument selbst kommt nicht in den Baum (AGENTS.md §11).
