# 0079 — Einstellungsebenen, Schranken und ein Austausch, der nichts still übernimmt

**Stand:** angenommen
**Datum:** 2026-09-08
**Anlass:** Werkstück O3 von Phase 5.2
(#1427) —
Einstellungen sollen mitreisen können, ohne dass dabei etwas mitreist, das
auf diesem Rechner bleibt; und eine Sicherheitsschranke soll keine gewöhnliche
Vorliebe sein.

## Lage

**Was am 2026-09-08 galt, ist nachgelesen und nicht erinnert** — am Stand
`4a27570c`:

| Frage | Antwort |
| --- | --- |
| Wo Einstellungen liegen | zwei Tafeln: `settings` je Profil (19 Schlüssel), `nutzereinstellungen` je Mensch (7) |
| Wie sie aufgelöst werden | Profilwert → Nutzerwert → Vorgabe ([domain.Verschmolzen], Werkstück 9 von #1351) |
| Was eine Antwort trug | den gewählten und den wirksamen Wert — **nicht**, wessen Entscheidung der wirksame war |
| Was mitreisen konnte | **nichts.** Es gab keinen Weg hinaus und keinen hinein — an keiner der vier Flächen |
| Was eine Schranke war | nichts. `dienst.scryfall` ist eine Einstellung wie `schriftgroesse`: das Profil sticht, und ein „nein" des Menschen gäbe es nicht |

**Und was das kostet, ist die Frage aus dem Issue**: ein Mensch mit zwei
Profilen sagt dieselbe Antwort zweimal, ein Mensch mit zwei Rechnern gar
nicht — und wer „dieses Programm geht nicht ins Netz" meint, sagt es je Profil
und je Rechner neu.

## Entscheidung

### Vier Ebenen, und sie sagen, wem eine Frage gehört

Jeder Schlüssel beider Reihen steht in **genau einer** Ebene
([domain.EbeneVon], `internal/domain/einstellungsebene.go`):

| Ebene | Heißt | Reist |
| --- | --- | --- |
| `geraet` | dieser Rechner, diese Installation | **nie** |
| `nutzer` | der Mensch — er nimmt es in jedes Profil mit | ja |
| `profil` | die Arbeitsumgebung | ja |
| `arbeitsplatz` | wie das Fenster steht | **nur auf Wunsch** |

**Die Ebene sagt nicht, in welcher Tafel der Wert liegt.** Wo er liegt, bleibt,
wie es war: `settings` je Profil, `nutzereinstellungen` je Mensch. Ein Umzug
wäre eine Verhaltensänderung ohne Anlass (§15, Satz 4) und brächte nichts, was
diese Zuordnung nicht schon bringt.

**Wofür sie gebraucht wird, sind zwei Fragen, die es vorher nicht gab**: darf
dieser Wert mitreisen, und in welchen Abschnitt eines Austauschs gehört er.

**Die vier Erscheinungseinstellungen sind `nutzer`**, obwohl das Profil sie
übersticht: die Frage gehört dem Menschen, und das Profil **übersteuert** sie
nur — so steht es seit Werkstück 9 in `uebersteuerung.go`. Ihre Übersteuerung
reist trotzdem im Abschnitt `profil`, denn dort ist sie **gesetzt**.

### Die Auflösung bleibt — die Herkunft kommt dazu

`Verschmolzen` gilt unverändert. Neu ist, dass jede Antwort sagt, **wessen
Entscheidung** der wirksame Wert war: `profil`, `nutzer` oder `vorgabe`
([domain.Geltung]). Ohne sie sagt „dunkel" nicht, was ein Zurücksetzen
bewirkte — und ein Import weiß nicht, ob er etwas überschreibt oder bestätigt.
Sie steht an **allen vier** Flächen, zusammen mit der Ebene.

### Eine Richtlinie wird gefaltet, nicht gestochen

Vier Schlüssel sind **Schranken** und keine Vorlieben: jeder `dienst.*`,
`richtlinie.nurlokal`, `richtlinie.kostengrenze` und
`sicherheit.faktorbeimwechsel`. Für sie gilt die **Schnittmenge aller
geltenden Ebenen** ([domain.Verschaerft]): der strengste gesetzte Wert
gewinnt, und die Reihenfolge der Ebenen spielt keine Rolle.

**Eine Ebene, auf der niemand etwas eingetragen hat, sagt nichts.** Deshalb
geht die **rohe** Abbildung in die Faltung und nicht die aufgefüllte: sonst
erstickte die Vorgabe eines Profils die Wahl eines Menschen, sobald sie
strenger ist. Und deshalb ist eine Vorgabe auch keine Schranke gegen einen
Import — `dienst.mtgapi` steht ohne Wahl auf „aus", und wer das für eine
Schranke hielte, ließe keinen Dienst je durch eine Datei anschalten.

**„Nur lokal" ist die einzige Regel über mehrere Schlüssel**: steht sie an,
gilt jeder `dienst.*` als `aus`, und die Herkunft ist die der Schranke. Sie
steht in `geltung.go` und nicht in der Faltung, weil sie nichts über sich
selbst sagt, sondern über jeden Schalter.

**Die Kostengrenze ist heute Form ohne Wirkung.** Sie wird geprüft (eine ganze
Zahl in Cent, höchstens 10.000.000) und gefaltet; wer sie liest, kommt mit
Phase 5.3. Das steht in ihrem Kommentar und nicht zwischen den Zeilen (§10).

### Der Arbeitsplatz bekommt Schlüssel, keine Bindung

Vier neue Schlüssel — `arbeitsplatz.navigator`, `…inspektor`,
`…navigatorbreite`, `…inspektorbreite` —, je Profil mit `erben` und beim
Menschen mit dem Rückfall (sichtbar, 240 und 320 Punkte wie in [0076](0076-die-ordnung-der-arbeitsplaetze.md), Grenzen 120 bis 800).

**Und keine eigene Bindung**, obwohl das Issue beides zur Wahl stellt: zwei
Bindungen hießen zwei neue Wege an **jeder** der vier Flächen (§16), und beide
täten genau das, was `EinstellungSetzen` schon tut. Was hier dazukommt, sind
Schlüssel, kein Mechanismus — Werkstück U4
(#1415)
liest sie über die Einstellungen, die das Fenster ohnehin lädt.

### Ein Austausch in drei Schritten, und der mittlere ändert nichts

**Ausfuhr → Vorschau → Übernahme.** Das Format trägt eine `fassung` (heute 1)
und `bereiche` je Ebene; darin stehen **nur gesetzte** Werte.

| Abschnitt | Woher | Welche Schlüssel |
| --- | --- | --- |
| `nutzer` | `nutzereinstellungen` | Ebene `nutzer` |
| `profil` | `settings` | Ebene `nutzer` oder `profil` |
| `arbeitsplatz` | `settings` | Ebene `arbeitsplatz` — nur auf Wunsch |
| — | — | Ebene `geraet` reist nicht |

**Keine Vorgaben.** Eine Datei voller Vorgabewerte sähe drüben wie eine Wahl
aus und fröre den heutigen Stand ein: wer später eine Vorgabe ändert,
erreichte damit niemanden mehr, der einmal importiert hat.

**Zwei Zahlen, und beide sind gerechnet.** Die größte Datei ist **64 KiB** —
der größte Austausch, den dieses Programm schreiben kann, sind rund dreißig
Zeilen von zusammen etwa 1,5 KB, also das Vierzigfache Luft; wer mehr schickt,
schickt etwas anderes. Die größte Tiefe ist **3**, und es sind genau drei:
`{ fassung, bereiche: { ebene: { schlüssel: wert } } }`. Beide greifen **vor**
dem Entziffern — eine Datei, die nicht gelesen werden soll, wird nicht gelesen.

**Was auffällt, wird genannt und nicht verschluckt** (§8): ein unbekanntes
Feld, ein fremder Abschnitt, ein Schlüssel der falschen Ebene und ein
unerlaubter Wert sind **Befunde**; die übrigen Zeilen bleiben. Eine **neuere
Fassung** ist dagegen ein Fehler, und die Meldung nennt beide Zahlen.

### Die Revision ist eine Marke und kein Zähler

Die Vorschau gibt eine **Revision** ([domain.Standmarke]) — sechzehn Hexzeichen
über Profil, Mensch und beide Wertabbildungen. Die Übernahme verlangt sie;
weicht der Stand ab, bleibt alles stehen und die Antwort sagt es (§7, Regel 2).

**Warum keine Zählertafel wie in [0078](0078-der-wirksame-akteur.md)**: dort
steigt die Revision in **einem** Auslöser. Hier wären es **zwei** Tafeln, und
jeder Schreibweg müsste ihn treffen — wer einen vergäße, hätte einen Stand,
der sich ändert, ohne dass es jemand sieht. Eine Marke über die Werte selbst
kann das nicht: sie **ist** der Stand. Und sie fasst mit, wer wo arbeitet — ein
Profilwechsel zwischen Vorschau und Übernahme wäre sonst bei zwei frischen
Profilen gar nicht zu sehen.

### Eine Schranke lässt sich nicht durch eine Datei aufweichen

Geprüft wird gegen das, was **ohne diese Zeile** gälte. Wer als Mensch „nur
lokal" gesagt hat, hat es für jedes Profil gesagt; eine Datei, die in einem
Profil einen Dienst anschaltet, hebt diese Entscheidung auf — das ist ein
**Befund** und wird nicht angewendet.

**Die eigene Wahl ist dabei keine Schranke gegen sich selbst**: ein Profil,
das seinen Dienst abgeschaltet hat, darf ihn per Datei wieder anschalten.
Sonst wäre jede einmal getroffene Wahl endgültig.

### Die Übernahme ist atomar und fragt vorher

Beide Tafeln in **einer** Transaktion (`storage.Einstellungen.EinstellungenUebernehmen`).
Ein halber Import ist schlimmer als keiner: wer die Sprache übernimmt und die
Barrierefreiheit nicht, hinterlässt einen Stand, den keine Vorschau je gezeigt
hat.

**Und der Riegel steht davor** (0078): die Übernahme ist `lokalSchreiben` und
fragt die Autorisierung, bevor sie liest — das Nein kommt vor jeder anderen
Antwort. Ausfuhr und Vorschau sind `lokalLesen`. Die Wache ist hier **Pflicht**
und kein `MitPruefer`: was Pflicht sein kann, wird Pflicht.

## Verworfen

- **Eine eigene Tafel je Ebene.** Vier Tafeln für dieselbe Frage, drei
  Wanderungen und vier neue Schreibwege — und der Gewinn wäre eine Zuordnung,
  die eine Abbildung im Code genauso gut hergibt. Die Ebene ist eine Aussage
  über die **Frage**, nicht über den Speicherort.
- **Eine Gerätetafel für die Schranken.** Sie wäre die dritte Ebene der
  Faltung. Was heute fehlt, ist nicht die Tafel, sondern ein Anlass: `nutzer`
  und `profil` reichen für jede Schranke, die es gibt, und `Verschaerft` nimmt
  beliebig viele Werte — die dritte Ebene kostet dort keine Zeile.
- **Den Diff des Aufrufers übernehmen.** Er ist eine Behauptung. Die Übernahme
  liest die Datei noch einmal und rechnet den Diff selbst; was der Aufrufer
  beiträgt, ist die **Auswahl** und die Revision.
- **Eine Profilkennung in der Datei.** „Aus welchem Profil kam das" wäre eine
  lokale Kennung, und der Zweck des Formats ist gerade, dass eine Datei von A
  nach B geht. Ein fremdes Profil fällt trotzdem auf: die Marke fasst es mit.
- **Die Übernahme als `PUT`.** Sie ist nicht idempotent — derselbe Ruf ein
  zweites Mal trifft eine andere Revision und bekommt einen Konflikt.

## Folgen

- **Drei neue Fähigkeiten an allen vier Flächen** (`FAEHIGKEITEN.md`), drei
  Zeilen in `AUTORISIERUNG.md`, zwei neue schreibende Bindungen in
  `SCHREIBWEGE.md`.
- **Zwölf neue Schlüssel** — vier für den Arbeitsplatz und zwei Schranken, je
  auf beiden Ebenen. Sie ändern **nichts** an einem bestehenden Bestand: jede
  Vorgabe ist der Zustand von gestern.
- **Eine Verhaltensänderung, und sie steht im Änderungsprotokoll**: ein
  `dienst.*` wird jetzt gefaltet statt gestochen. Solange niemand eine
  Nutzerschranke setzt, kommt dasselbe heraus wie vorher.
- **Keine Wanderung.** Beide Tafeln bleiben, wie sie sind.

## Nachtrag 2026-09-09 — die Seitenleiste

Werkstück U14 von Phase 5.1
(#1490)
legt einen Schlüssel auf die **Geräteebene**: `seitenleiste` (sichtbar,
verborgen; Vorgabe sichtbar). Er sagt, ob die Seitenleiste des Fensters
steht — eine Fensterfrage, die nicht mit dem Profil reist und in der
Austausch-Ausfuhr fehlt wie jeder Gerätewert. Der Arbeitsplatz-Schlüssel
`arbeitsplatz.navigator` liegt bewusst **nicht** dort: er beschreibt eine
Arbeitsumgebung, die Seitenleiste das Gerät.

## Nachtrag 2026-10-02 — „nur lokal" trifft Schalter, keine Angaben

**Der Satz „steht sie an, gilt jeder `dienst.*` als `aus`" galt, solange
jeder Schlüssel auf `dienst.` ein Schalter war.** Seit #1540 steht unter
einem Dienst seine Adresse (`dienst.<name>.adresse`), seit NET-03 (#1707)
stehen dort die vier Pfade des Dienstzeugnisses (`dienst.tls.*`). Die
Schranke traf sie mit, und ein Dienst jenseits der Schleife suchte ein
Zertifikat namens „aus" — Folgebefund F2 des Lastlaufs QA-08
(#2004).

**Die Regel heißt jetzt: „nur lokal" legt genau die Dienstschalter auf
`aus`** — `dienst.<name>` für jeden Dienst aus `domain.Dienste`, erkannt wie
in der Faltung (`istDienstschalter`). Was darunter steht, schaltet nichts und
behält Wert und Herkunft. Das ist, was die Richtlinie von Anfang an sagte:
„schaltet jeden fremden Dienst ab" (`richtlinie.go`) — eine Adresse oder ein
Zeugnispfad ist keiner. Gehalten von `TestNurLokalTrifftGenauDieDienstschalter`
über die ganze Reihe, damit ein neuer Schlüssel unter `dienst.` die Frage
stellt, statt still mitgeschaltet zu werden.
