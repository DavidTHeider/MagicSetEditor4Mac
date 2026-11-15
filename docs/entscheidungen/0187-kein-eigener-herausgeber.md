# 0187 — Kein eigener Herausgeber: der Schlüsselbund bleibt über `security`

**Stand:** angenommen am 2026-10-11 · #2295.
**Ergänzt** [0185](0185-ein-eigener-herausgeber-truege-die-bindung.md) und
[0181](0181-der-schluesselbund-bleibt-ungebunden-solange-ad-hoc-signiert-wird.md).
**Entschieden in diesem Auftrag**, nicht von einem Menschen: der
Auftraggeber hat die Frage am 2026-10-11 ausdrücklich abgegeben („beim
Schlüsselbund selbst entscheiden").

## Die Frage

0185 hat gemessen, dass ein **selbst ausgestellter Herausgeber** die Bindung
eines Schlüsselbundeintrags über eine Fassung trägt, und den Bau nach
#2295
verschoben — beginnend mit der Frage, **wo der Signierschlüssel liegt**. Die
lag bei einem Menschen. Jetzt liegt sie hier, und mit ihr die zweite: **hält
eine Bindung überhaupt auf, wogegen sie gedacht ist** — ein fremdes Programm
desselben Nutzers?

## Was gemessen ist — am 2026-10-11

Im Lume-Gast mit macOS 15.2, an einem künstlichen Bund
([`belege/i01-bindung.md`](../belege/i01-bindung.md), `node
scripts/bindunggast.mjs`), zu E1–E4 aus 0185 ein Fall **E5**: eine kleine
Bibliothek (`scripts/huellenprobe/gast/fremd.c`) wird über
`DYLD_INSERT_LIBRARIES` in den gebundenen Leser geladen und fragt vor dessen
`main` nach demselben Eintrag.

| Fall | Ergebnis |
| --- | --- |
| **E5** fremder Code in der gebundenen Fassung `leser-a1` — signiert vom Herausgeber, **nicht gehärtet** | **liest still** |
| **E5** die gehärtete Fassung `leser-a3` (`--options runtime`, derselbe Herausgeber) | liest still — die Bindung trägt auch gehärtet |
| **E5** fremder Code in der gehärteten Fassung | **läuft gar nicht** |

Dazu drei Feststellungen ohne eigenen Lauf, je mit ihrer Stelle:

- **Die Bündel dieser Anwendung sind nicht gehärtet.** `scripts/release.sh`
  und `scripts/bauen.sh` signieren mit `codesign --force --deep --sign -`,
  ohne `--options runtime`.
- **Wer `codesign` starten kann, signiert mit dem Herausgeber, ohne gefragt
  zu werden**, sobald der Schlüssel `codesign` vertraut. Genau so hat die
  Probe ihre Leser signiert (`security import … -T /usr/bin/codesign`, dann
  `codesign --sign` aus einem Perl-Treiber — im Protokoll desselben Belegs,
  ohne Rückfrage).
- **Das Repositorium ist privat** (`gh repo view`, am 2026-10-11): die
  Abbilder erreichen, wer Zugang hat. **Das sagt nicht**, dass Bau- und
  Laufrechner derselbe sind oder dass jeder Laufrechner den Schlüssel trüge
  — nur, dass heute niemand außer dem Herausgeber Abbilder bekommt.

**Was davon gemessen und was geschlossen ist** (Befund 6 des Nachlaufs über
Phase I): gemessen ist der **Mechanismus** am Leser — E5. Dass er an `mse`,
`mse-mcp`, `mse-rest` und dem Fenster ebenso greift, ist ein **Schluss**:
sie sind signiert wie der Leser, ad hoc und ohne `--options runtime`, und
die Probe hat keines von ihnen gebunden.

## Die Entscheidung

**Kein eigener Herausgeber und keine Bindung.** #2295 wird nicht gebaut; die
Einträge bleiben, wie [0181](0181-der-schluesselbund-bleibt-ungebunden-solange-ad-hoc-signiert-wird.md)
sie beschreibt.

- **Ohne Härtung hält die Bindung nichts auf** — gemessen am Leser (E5),
  geschlossen für die vier Programme: ein fremdes Programm startet ein
  gebundenes mit seiner eigenen Bibliothek und liest in dessen Namen, still.
  Die Bindung allein verschiebt die Lücke von `security` auf
  `DYLD_INSERT_LIBRARIES`.
- **Mit Härtung hängt es an der Verwahrung, und die hier angenommene trägt
  nicht.** Angenommen ist eine **Verwahrung auf dem Mac, der signiert, ohne
  einen Menschen beim Signieren** — der Auftraggeber will nicht gebraucht
  werden, und Kennwörter gibt dieser Auftrag nicht ein (§17). Dann darf
  `codesign` den Schlüssel still benutzen, und damit jedes Programm desselben
  Nutzers: es signierte sich einen Leser mit dem Herausgeber, und der läse
  still. **Auf dem Mac, der signiert, schützt die Bindung damit nichts.** Auf
  einem anderen Mac, der den Schlüssel nicht trägt, schützte sie mit Härtung
  — **das ist der Fall, unter dem diese Entscheidung wieder aufzunehmen ist**,
  und er setzt eine andere Verwahrung voraus als die angenommene.
- **Der Preis wäre trotzdem ganz zu zahlen**: die native API in allen vier
  Programmen statt `/usr/bin/security`, die Härtung samt ihrer eigenen
  Messung (WKWebView, `purego` in `internal/biometrie`, das cgo-Programm),
  ein Umzug jedes Eintrags — und bei verlorenem Schlüssel genau die
  Rückfragen an Programmen ohne Fenster, die 0181 verworfen hat.

## Was gilt

- **Das Risiko bleibt benannt** (`docs/HAERTUNG.md`, SEC-05): ein Programm
  desselben Nutzers liest die Einträge über `security`, ohne dass macOS
  fragt. **E5 ergänzt**: eine Bindung ohne Härtung änderte daran nichts.
- **Ein gesperrter Bund schützt** (E4).
- **Wieder aufzunehmen, wenn beides zutrifft**: die Anwendung erreicht Macs,
  auf denen der Schlüssel nicht liegt (ein öffentliches Abbild, andere
  Nutzer), **und** der Schlüssel liegt dort, wo ein Programm desselben
  Nutzers ihn ohne einen Menschen nicht erreicht (ein Hardwareschlüssel, ein
  zweiter Rechner, der nur signiert). Dann zuerst die Härtung als eigenes
  Werkstück, danach die Schritte aus #2295.

## Verworfen

- **Der Schlüssel in einem eigenen Bund, dessen Kennwort im Anmeldebund
  steht.** Das Kennwort läse jedes Programm über `security` — dieselbe
  Lücke wie heute, eine Ebene tiefer.
- **Ein Hardwareschlüssel.** Er hielte fremde Programme ab, verlangte aber
  bei jeder Freigabe einen Menschen — gegen die Weisung, die diese
  Entscheidung erst möglich gemacht hat.
- **Nur härten.** Solange die Einträge `/usr/bin/security` vertrauen (E1),
  ist die erste Lücke `security` und nicht eine eingeschleuste Bibliothek;
  die Härtung allein schützte keinen Eintrag.
- **Binden ohne Härtung, „besser als nichts".** E5 zeigt, dass es nichts ist
  — und es kostete die native API, den Umzug und die Verwahrung.
