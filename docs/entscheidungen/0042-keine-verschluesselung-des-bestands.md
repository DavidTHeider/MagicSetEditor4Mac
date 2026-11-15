# 0042 — Keine Verschlüsselung des Bestands

> **Abgelöst am 2026-09-06 durch
> [0072](0072-ein-treiber-der-verschluesseln-kann.md).** Der Bestand **wird**
> verschlüsselt.
>
> **Beide Hälften der Abwägung unten sind hinfällig, und keine davon war
> falsch, als sie geschrieben wurde:**
>
> 1. **Was zu schützen wäre.** Hier stand „Kartensets: Namen, Regeltexte,
>    Bildpfade, Stilangaben. **Keine** Zugangsdaten, keine Daten Dritter." Das
>    galt am 2026-08-22 und gilt seit Phase 4.12 nicht mehr: im Bestand liegen
>    Namen, Geburtsdaten, Anschriften, Fotos, Anbieterkonten und
>    TOTP-Geheimnisse. Der Satz „wer eine Sammlung selbstgebauter Magic-Karten
>    in die Hände bekommt, hat eine Sammlung selbstgebauter Magic-Karten" ist
>    heute unwahr.
> 2. **Was es kostet.** Hier stand „ein zweiter Treiber, cgo und der Wegfall
>    der Kreuzübersetzung aus §12". Das war eine Annahme über **einen** Weg —
>    SQLCipher —, und der Vermerk sagt selbst, dass kein Vergleich der Wege
>    erhoben wurde. `ncruces/go-sqlite3` ist **cgo-frei**: es führt SQLite als
>    WebAssembly aus. **Gemessen am 2026-09-06**, nach dem Tausch:
>    `GOOS=windows go build ./...` und `GOOS=linux go build ./...` glücken
>    beide. Die Prüfstufe fällt nicht weg.
>
> Was von diesem Vermerk **stehen bleibt**, ist der Abschnitt „Was das nicht
> abdeckt": eine Sicherung auf einem USB-Stick ist unverschlüsselt, gleich wie
> der Rechner eingestellt ist. Das gilt weiter, und 0072 entscheidet es
> ausdrücklich noch einmal — eine Sicherung, die nur auf ihrem
> Entstehungsrechner aufgeht, wäre keine.

**Stand:** abgelöst · **Datum:** 2026-08-22 ·
**Betrifft:** #592,
#803 ·
**Ergänzt** [0002](0002-sqlite-als-arbeitsformat.md), [0041](0041-die-setgrenze-liegt-im-dienst.md)

## Die Frage

Phase N verlangt eine Aussage: **wird der Bestand verschlüsselt?** Entweder
SQLCipher samt Pinnung, Baufolgen und zweitem Treiber — oder eine begründete
Ablehnung mit dem, was stattdessen gilt.

## Was gemessen ist

| Frage | Gemessen am 2026-08-22 |
| --- | --- |
| Kennt `modernc.org/sqlite@v1.56.0` einen Schlüssel? | **Nein.** Kein `sqlite3_key`, kein `SQLITE_HAS_CODEC` im ganzen Modul |
| Was kostet cgo bei der Kreuzübersetzung? | `CGO_ENABLED=1 GOOS=linux go build ./...` endet mit **1**: `runtime/cgo`, `gcc_amd64.S:27: unknown token in expression` |

**Die Reichweite der zweiten Zahl gehört dazu** (§10): sie sagt, was auf
**diesem** Rechner mit **dieser** Werkzeugkette geschieht — nicht, dass es
unmöglich wäre. Mit einer eingerichteten Linux-Werkzeugkette ginge es. Die
Aussage ist deshalb nicht „cgo geht nicht", sondern: die Stufe `types` aus §12,
die für Windows und Linux übersetzt, damit plattformabhängiger Code nicht
unbemerkt bricht, **fiele heute sofort aus** — und ihr Ersatz wäre eine
Werkzeugkette je Ziel, die jemand pflegen muss.

## Was hier überhaupt zu schützen wäre

Der Bestand hält Kartensets: Namen, Regeltexte, Bildpfade, Stilangaben. **Keine
Zugangsdaten, keine Zahlungsdaten, keine Daten Dritter.** Wer eine Sammlung
selbstgebauter Magic-Karten in die Hände bekommt, hat eine Sammlung
selbstgebauter Magic-Karten.

Das ist kein Freibrief — es ist die Antwort auf die Frage, wie viel eine
Maßnahme kosten darf. Eine, die einen zweiten Treiber, cgo und den Wegfall
einer Prüfstufe kostet, ist hier zu teuer.

## Was die Messungen tragen — und was nicht

**Sie tragen die Kosten, nicht die Entscheidung.** Der Zweitblick über Phase N
hat das angestrichen, und er hat recht: Messung 1 reicht bis „der heutige
Treiber hat keinen Codec", Messung 2 bis „diese Werkzeugkette scheitert an
diesem Kreuzbau". Keine von beiden sagt etwas über den **Schutzbedarf**, über
andere Wege zur Verschlüsselung oder über die Kosten einer eingerichteten
Zielwerkzeugkette.

Hier stand die Ablehnung so, als folge sie aus den Zahlen. Sie folgt nicht
daraus. **Sie ist eine Abwägung**, und die Zahlen sind eine Seite davon.

## Die Entscheidung

**Der Bestand wird nicht verschlüsselt** — als Abwägung, nicht als Herleitung.

Auf der einen Seite steht, was zu schützen wäre: Kartensets, keine
Zugangsdaten, keine Daten Dritter. Auf der anderen ein zweiter Treiber, cgo und
der Wegfall der Kreuzübersetzung aus §12, die heute dafür sorgt, dass
plattformabhängiger Code nicht unbemerkt bricht.

**Für die Freigabe 0.1 ist das entschieden.** Nicht „nie": wer den Schutzbedarf
anders beurteilt — etwa weil jemand Sets mit unveröffentlichtem Material baut —,
hat einen Grund, den diese Abwägung nicht enthält. Er ändert dann diesen
Vermerk, und der Abschnitt unten sagt, was dazugehört.

Und was stattdessen gilt:

1. **Die Platte ist Sache der Plattform.** macOS bringt FileVault mit; auf
   Apple Silicon liegt das Laufwerk ohnehin verschlüsselt, und FileVault bindet
   den Schlüssel an die Anmeldung. Das deckt den Fall ab, um den es bei
   „Verschlüsselung des Bestands" in aller Regel wirklich geht — den verlorenen
   oder gestohlenen Rechner. **Nicht hier nachgemessen**, sondern
   Plattformverhalten; wer es genau wissen will, sieht in *Systemeinstellungen →
   Datenschutz & Sicherheit → FileVault* nach, ob es an ist.
2. **Die Datei gehört den Nutzenden.** Sie liegt offen unter
   `~/Library/Application Support/MagicSetEditor4Mac/`, und das ist Absicht:
   sie lässt sich kopieren, sichern und mit `sqlite3` ansehen. Dieselbe
   Offenheit macht [0041](0041-die-setgrenze-liegt-im-dienst.md) ausdrücklich —
   die Setgrenze schützt vor einer Verwechslung, nicht vor einem Willen.
3. **Wer weitergehen will, hat den Weg.** Ein verschlüsseltes Disk-Image oder
   ein verschlüsselter Ordner nimmt die Datei auf, ohne dass diese Anwendung
   etwas davon wissen muss.

## Was das nicht abdeckt

**Den Fall, in dem jemand die Datei hat und die Anmeldung nicht.** FileVault
schützt das Laufwerk dieses Rechners — nicht eine Kopie, die anderswo liegt.

Und genau diese Kopie ist seit
#798 ein
Handgriff: `storage.Sichern` schreibt eine vollständige Datei, und wo sie
landet, entscheiden die Nutzenden. Ein USB-Stick mit einer Sicherung ist
unverschlüsselt, gleich wie der Rechner eingestellt ist.

**Deshalb steht es dort, wo gesichert wird**, und nicht nur hier:
`docs/RELEASE.md` sagt es beim Sichern.

## Wer das ändern will

Ändert zuerst diesen Vermerk und [0002](0002-sqlite-als-arbeitsformat.md) — die
Wahl des Treibers ohne cgo ist die Voraussetzung, an der alles hängt. Und er
trägt nach, was §12 dann noch prüft: eine Kreuzübersetzung, die es nicht mehr
gibt, ist keine Prüfung, sondern eine Zeile in einer Datei.

**Was dazugehört, und was hier bewusst nicht steht:** ein Vergleich der Wege.
SQLCipher ist einer; ein verschlüsseltes Disk-Image, das die Anwendung gar
nichts angeht, ein zweiter; eine Verschlüsselung einzelner Felder ein dritter.
Erhoben ist keiner davon — diese Entscheidung sagt, dass sie für 0.1 nicht
erhoben werden, und nicht, dass sie nichts taugen.
