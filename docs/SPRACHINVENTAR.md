# Sprachinventar

**Welche Sprachen es gibt, woher das weiß, wer es sagt, und wie weit jede
ist.** LOC-01 (#1687).

Zwei Listen, die nicht dieselbe sind: die **Sprachpakete**, die MSE mitbringt,
und die **Fenstersprachen**, in denen diese Anwendung spricht. Die erste ist
gezählt, die zweite entschieden.

## 1 Die Sprachpakete des Packs

**Gezählt, nicht erinnert** — am 2026-09-29 über
`MagicSetEditorPacks/Full-Magic-Pack`, Stand `622160f` vom 2026-08-07, dem
Pack aus [0025](entscheidungen/0025-die-vorlagen-duerfen-mitgeliefert-werden.md).
Jedes Paket trägt `version: 2025-06-18`; die Prüfsumme gilt der Datei
`locale` im Paket.

| Paket (MSE-Name) | `full name:` | BCP 47 | Fenstersprache | SHA-256 von `locale` | Bytes |
| --- | --- | --- | --- | --- | ---: |
| `ch-s.mse-locale` | 简体中文 (Simplified Chinese) | `zh-Hans` | **zh** | `3137969f46fd0045d979958e2b777732fed3fa5aac167277c55f0289f6a100d2` | 41578 |
| `ch-t.mse-locale` | 繁體中文 (Traditional Chinese) | `zh-Hant` | — | `7ea5b9346fb078881280d2c636231344823840704182d526a671b5f28ad29da0` | 41729 |
| `da.mse-locale` | Dansk (Danish) | `da` | — | `44005735c6906f45eb3fb586f52b1f468372cf663598fec50eecf55303ac166c` | 48743 |
| `de.mse-locale` | Deutsch (German) | `de` | **de** | `d07049698354a183ecbc561106e17ce812ba16b07145af21a359c1d5f809c627` | 43122 |
| `en.mse-locale` | English | `en` | **en** | `d8735b291a49713b6a33f3e2f678688584410547e20601b7066572b18b8de68c` | 46895 |
| `es.mse-locale` | Español (Spanish) | `es` | **es** | `e1ed97202e91e801b2fea120255e654ff33f9c5b1aa3304c22e3dff0ade1962b` | 44083 |
| `fr.mse-locale` | Français (French) | `fr` | **fr** | `a5a996386e880e8f7a656a73d3adf634a64b718b8cf8a8767156584c3a425d08` | 52253 |
| `it.mse-locale` | Italiano (Italian) | `it` | **it** | `8dd151fb32a1986cf726913ee8c373435561ecf72a1689c44206020eae4698b8` | 43797 |
| `jp.mse-locale` | 日本語 (Japanese) | `ja` | **ja** | `a86a17688ac31fb47474b3fd150957808b57359f409f14d5a05e447ed8b94c65` | 49677 |
| `ko.mse-locale` | 한국어 (Korean) | `ko` | — | `16ccf8e3f53306a11db5cd4b66d4adda9798c48a2fe1f108da58ede6f0f7d577` | 51594 |
| `pl.mse-locale` | Polski (Polish) | `pl` | — | `4f83e032763fc88b7c77b1a7af9fc1055ed8391366a22d3e66ae16cf18b90804` | 51023 |
| `pt-br.mse-locale` | Português do Brasil (Brazilian Portuguese) | `pt-BR` | — | `9d9052e34d6e6cbc7b3d23bb5d39de584814f8fc4a1b0a40ab46b113a90cc88f` | 44574 |
| `ru.mse-locale` | Русский (Russian) | `ru` | **ru** | `4bbc2eccac113d8a7c1d5fc6a3478d680921c4a6d0a9dd9ebc658ff65fc9428c` | 54783 |

**Der MSE-Name ist kein Sprachcode.** `jp` ist das Länderkürzel Japans,
`ch-s` keines von beidem. MSE fragt Übersetzungen in einem Spiel- oder
Stilpaket unter genau diesem Namen ab (`the_locale->name()`,
`src/data/locale.cpp`, Zeilen 138–141, Fassung f9d9356 vom 2021-07-06) — in
`magic.mse-game/set_fields` etwa `ch-s: 标题`. Die Abbildung steht deshalb an
**einer** Stelle, `domain.Paketsprache`, und gilt für jede Aufschrift aus
einem Paket. **Bis LOC-01 fehlte sie**, und niemand sah es: mit zwei Sprachen
hießen Fenster und Paket gleich.

**Nachzählen**:

```bash
for d in /Users/david/full-magic-pack/data/*.mse-locale; do
  shasum -a 256 "$d/locale"
done
```

## 2 Was „unterstützt" für ein Paket heißt

**Jedes der dreizehn Pakete ist lesbar** — `msepaket` liest `.mse-locale`
(`msepaket.ArtSprache`, Operationenmatrix in [`FAEHIGKEITEN.md`](FAEHIGKEITEN.md)),
und jede Aufschrift eines Spiel- oder Stilpakets, die unter einem dieser Namen
steht, wird gefunden, sobald das Fenster die Sprache spricht.

**Fünf davon haben keinen Fensterkatalog**: `ch-t`, `da`, `ko`, `pl`,
`pt-br`. Sie sind inventarisiert und lesbar, aber keine Wahl in den
Einstellungen — LOC-01 nennt acht Fenstersprachen, und die Pakete dieser fünf
übersetzen **MSEs** Oberfläche, nicht diese. Was ihre Kataloge daraus machen
könnten, ist der Weg aus LOC-02 (#1688):
eine Übersetzung, die jemand selbst mitbringt.

**Was MSE2 mit ihnen tut und hier fehlt**, steht als Kompatibilitätsrest in
[`PARITAET.md`](PARITAET.md): MSE2 nimmt die Beschriftungen seiner eigenen
Oberfläche aus dem Paket; dieses Programm hat eine andere Oberfläche und
eigene Kataloge.

## 3 Die Fenstersprachen

| Kürzel | BCP 47 (`lang`, `Intl`) | Stand | Herkunft | übersetzt | redaktionell geprüft | Paket |
| --- | --- | --- | --- | --- | --- | --- |
| `de` | `de` | **Ausgangssprache** | Ausgangssprache | — | — | `de` |
| `en` | `en` | maschinell übersetzt | mit dem Code | — | — | `en` |
| `es` | `es` | maschinell übersetzt | Sprachmodell | 2026-09-29 | — | `es` |
| `fr` | `fr` | maschinell übersetzt | Sprachmodell | 2026-09-29 | — | `fr` |
| `it` | `it` | maschinell übersetzt | Sprachmodell | 2026-09-29 | — | `it` |
| `ja` | `ja` | maschinell übersetzt | Sprachmodell | 2026-09-29 | — | `jp` |
| `ru` | `ru` | maschinell übersetzt | Sprachmodell | 2026-09-29 | — | `ru` |
| `zh` | `zh-Hans` | maschinell übersetzt | Sprachmodell | 2026-09-29 | — | `ch-s` |

**Herkunft, Tag und Redaktion stehen seit H04 im Fenster**
(#2258) —
unter der Sprachwahl, aufklappbar, je Sprache in der Sprache des Fensters.
Das Register dazu ist `sprachherkunft` in `frontend/src/sprache/kataloge.ts`,
und `sprachinventar.test.ts` hält diese Tafel dagegen, Zeile für Zeile. **Eine
Redaktion, die stattfand, bekommt ihren Tag hier und dort** — und keine
andere Stelle sagt, dass es eine gab. „Mit dem Code" heißt: Englisch entstand
zusammen mit dem Programm, ohne eigenen Übersetzungslauf.

Das Register steht in `frontend/src/sprache/kataloge.ts` (`sprachstaende`),
die Wahl in der Einstellung `sprache` (`internal/domain/einstellungsliste.go`),
die Menüleiste und die Dialoge in `internal/app/sprachkatalog*.go`.

### Vier Stände, und sie sind nicht dasselbe (AK-03)

| Stand | Heißt | Belegt durch |
| --- | --- | --- |
| **ladbar** | jeder Schlüssel ist da, jeder Platzhalter auch | `frontend/src/sprache/zusatzsprachen.test.ts`, Menü- und Dialogkataloge in `internal/app` |
| **maschinell übersetzt** | ein Modell hat übersetzt, kein Mensch hat gelesen | das Register — **und nicht** eine Zählung |
| **redaktionell geprüft** | ein Mensch, der die Sprache spricht, hat jeden Text gelesen | ein Eintrag hier mit Name und Datum |
| **abgenommen** | jemand hat die Anwendung in dieser Sprache benutzt, samt Fenster | ein Beleg unter `belege/` |

**Englisch steht als maschinell übersetzt da**, und das ist keine
Bescheidenheit: es entstand zusammen mit dem Code, und niemand hat es
redigiert. **Gleiche Schlüssel belegen keine Qualität** — die sechs neuen
Sprachen wurden am 2026-09-29 von Sprachmodellen übersetzt, Datei für Datei
gegen die englische Gestalt geprüft, und von keinem Menschen gelesen. Die
Einstellung sagt es unter der Sprachwahl.

### Chinesisch heißt vereinfacht (AK-02)

**Das Fenster führt eine chinesische Schrift, und es ist die vereinfachte**
(`zh-Hans`, Paket `ch-s`). Beide Pakete liegen im Pack; eine zweite Schrift
verdoppelte den Katalog, ohne dass eine Übersetzung dafür vorliegt.

- **Das Dokument trägt `lang="zh-Hans"`**, nicht `zh`. Ein Han-Zeichen hat
  auf dem Festland, in Taiwan und in Japan verschiedene Formen; ohne die
  Schrift darf WebKit die traditionelle ziehen. Japanisch trägt `ja`, und
  dieselbe Regel gibt ihm die japanischen Formen.
- **Ein System in `zh-TW` oder `zh-HK` bekommt das vereinfachte**, nicht das
  deutsche: näher liegt es allemal, und die Wahl nennt die Schrift
  (简体中文).

### Zählformen

Deutsch, Englisch, Spanisch, Französisch und Italienisch kennen **eine** und
**viele**; Japanisch und Chinesisch zählen nicht (beide Formen gleich).
**Russisch braucht eine dritte** für 2 bis 4 — `zaehlwort` fragt
`Intl.PluralRules` der Sprache, die gerade spricht, und nimmt `wenige`, wo
die Regel `few` sagt. **Die Regel reist mit der Zählform**: eine Form, deren
Wörter eine eigene Übersetzung setzt, zählt nach deren Kennung, die übrigen
nach der Basis, deren Wörter sie tragen (0147). Zwei Sprachräume
nebeneinander zählen damit jeder nach seiner — bis zum Nachlauf über Phase
6.5 stand die Regel in einem Zustand des Moduls, und der zuletzt gezeichnete
Sprachraum galt für alle.

## 4 Was nicht übersetzt wird

| Was | Warum | Beleg |
| --- | --- | --- |
| **Feldnamen, Spielnamen, Stilbezeichnungen** aus einem MSE-Paket | Sie stehen so in der Datei; eine Übersetzung machte jedes Paket unlesbar (§9) | — |
| **Kommandozeile, REST und MCP** | Sie sprechen Deutsch, und ihre Fehler tragen einen **Schlüssel**, der in keiner Sprache etwas anderes heißt | [0146](entscheidungen/0146-die-flaechen-ohne-fenster-sprechen-deutsch.md) |
| **Fehlercodes und Feldnamen der Schnittstellen** | kanonische Form ([`LOKALISIERUNG.md`](LOKALISIERUNG.md)) | Vertragstest zwischen Go und den Katalogen |
