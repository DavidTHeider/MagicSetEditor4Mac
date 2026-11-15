# 0050 — Die Anwendung ersetzt sich nicht selbst

**Datum:** 2026-08-26 · **Status:** angenommen ·
**Anlass:** #1100,
Werkstück 5

---

## Die Frage

Stufe 2 der Aktualisierung lädt das `.dmg` ins Downloadverzeichnis. Danach
könnte ein Helfer das Bündel austauschen und die Anwendung neu starten — so
machen es Sparkle und die meisten macOS-Programme.

Das Issue hat die Frage ausdrücklich offen gelassen und verlangt, sie zu
**messen**, nicht zu behaupten.

## Was gemessen wurde

**Am 2026-08-26, an einer Datei, die diese Anwendung selbst geladen hat:**

```
$ xattr -l ~/Downloads/MagicSetEditor4Mac-99.9.9.dmg
com.apple.macl:
com.apple.provenance:
```

**Kein `com.apple.quarantine`.** Ein Go-Programm, das über HTTP lädt und mit
`os.Rename` ablegt, setzt das Attribut nicht — anders als Safari oder der
Finder, die es über LaunchServices bekommen. Das war die erste Frage des
Issues, und die Antwort ist damit belegt und nicht angenommen.

**Was das *nicht* belegt:** dass eine so geladene Anwendung startet. Ein
Bündel ohne Quarantäne umgeht die Erststartabfrage; ad hoc signiert bleibt es
trotzdem (§12, [0003](0003-keine-notarisierung.md)). Der Beleg dafür ist der
Start aus dem **echten** `.dmg`, und der steht im Freigabelauf
([`RELEASE.md`](../RELEASE.md)) — nicht hier.

## Die Entscheidung

**Der Weg bleibt „geladen, bitte installieren".** Die Anwendung nennt den Ort
der geprüften Datei und sagt, was zu tun ist. Sie tauscht sich nicht selbst
aus.

## Warum

**Ein Selbsttausch braucht Schreibrecht auf das eigene Bündel.** Wer die
Anwendung nach `/Applications` gezogen hat, hat sie dort oft mit Rechten
liegen, die dem Benutzerkonto das Schreiben verwehren. Der übliche Ausweg ist
ein privilegierter Helfer — also eine Rechteabfrage.

**Und genau diese Abfrage wäre die falsche Gewohnheit.** Ein Programm, das
nicht notarisiert ist, dessen Herkunft Gatekeeper nicht bestätigen kann und
das beim ersten Start eine ausdrückliche Freigabe braucht, sollte nicht ein
halbes Jahr später nach dem Administratorkennwort fragen, um sich selbst zu
überschreiben. Wer diese Frage bejaht, hat gelernt, sie zu bejahen.

**Der Nutzen ist klein.** Der Unterschied ist ein Doppelklick und ein Ziehen —
gegen einen Helferprozess, eine Rechteabfrage, einen Neustartweg und eine
Fehlerklasse („der Tausch ist auf halbem Weg gescheitert"), die man ohne
Notarisierung nicht sauber auffangen kann.

**Was stattdessen gebaut ist, deckt das Risiko ab, das wirklich zählt:** die
Prüfsumme wird verglichen, **bevor** die Datei ihren Namen bekommt
(#1100,
Stufe 2), und die Prüfsummenliste selbst trägt seit dem 2026-08-25 eine
GPG-Signatur (#892).
Wer will, prüft beides von Hand nach; der Weg steht im `README.md`.

## Was das kostet

**Zwei Handgriffe je Aktualisierung** — öffnen und ziehen. Und die
Erststartfreigabe, falls das Bündel doch mit Quarantäne ankommt, weil jemand
es über den Browser statt über diese Anwendung geladen hat.

**Und eine Fassung, die niemand installiert, bleibt liegen.** Das ist der
ehrliche Preis: eine Meldung ist keine Aktualisierung. Er ist bewusst bezahlt
— bei einem Programm ohne Notarisierung ist der stille Selbsttausch nicht die
freundlichere, sondern die undurchsichtigere Variante.

## Verworfen

| Weg | Warum nicht |
| --- | --- |
| **Sparkle einbinden** | Es setzt eine Developer-ID-Signatur voraus, um sinnvoll zu prüfen — und die gibt es hier nicht ([0003](0003-keine-notarisierung.md)). Ohne sie prüfte es eine Signatur, die niemandem gehört |
| **Eigener Helfer mit `SMJobBless`** | Verlangt ein Developer-ID-Zertifikat für Anwendung **und** Helfer. Fällt aus demselben Grund weg |
| **Tausch ohne Rechteabfrage, solange das Bündel schreibbar ist** | Zwei Wege je nach Ort der Anwendung — und der eine wird nie geprüft, weil dieser Rechner sie woanders liegen hat (§10: die Reichweite einer Messung) |
| **Das `.dmg` nach dem Laden von selbst öffnen** | Ein Programm, das ungefragt ein Fenster des Finders aufmacht, tut mehr als es angekündigt hat (§16, Regel 6) |

## Was daran später zu hinterfragen ist

**Wenn es je einen Apple-Developer-Account gibt**, ändert sich die Rechnung
vollständig: dann ist Sparkle der gebahnte Weg, und diese Entscheidung fällt
mit [0003](0003-keine-notarisierung.md) zusammen. Wer sie ändern will, ändert
zuerst jene.
