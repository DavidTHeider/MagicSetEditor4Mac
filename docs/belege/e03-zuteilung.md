# E03 — Zuteilung von Ausfuhr und frischem Zeichenwerk, A/B/A

**Rohbeleg zur Tafel in [`LEISTUNGSSCHNITTE.md`](../LEISTUNGSSCHNITTE.md)**
(Abschnitt E03, #2240),
abgelegt am 2026-10-11 nach Befund 7 des nachgeholten Zweitblicks in J05
(#2269):
bis dahin standen dort nur die zusammengefassten Werte und gekürzte Hashes.

## Stände und Programme

| | Commit | Testprogramm `internal/app` (SHA-256) |
| --- | --- | --- |
| A | `a0569df45a2a855799801ccdb742c840a7d1ee67` | `3e282e8e66e2bddcf8438ced9c393ca7386bc6fe28938ecfcf0bb43ff7dfa617` |
| B | `6b5404029b8491ef42bd57d0580d8905ca60f3e9` | `980e09d8ddcacca98b4f112a63032100e70f5aef6e3512ea22d59b6e57310c3c` |

Beide Programme sind mit `go test -c` aus ihrem Stand übersetzt; A lief im
Hauptbaum, B in einem eigenen Arbeitsbaum (`messung.sh`, `baum`). Maschine:
Apple M5, 10 Kerne, macOS 27.0.1, go1.27.2, am 2026-10-10 zwischen 12:05 und
12:22 UTC. **Die Last lag bei 13 bis 15** — andere Arbeit lief daneben. Die
Zuteilungen zählt das Profil bei `memprofilerate=1` vollständig und
unabhängig von der Last; **RSS und Spitze hängen von ihr ab** und sind
deshalb nur als Richtung zu lesen.

## Befehl

Das Skript, wie es lief (`$S` ist das Arbeitsverzeichnis der Sitzung):

```bash
#!/usr/bin/env bash
# A/B/A der Zuteilung je Lauf (E03): Differenz aus 4 und 1 Durchlauf bei
# memprofilerate=1, dazu RSS/Spitze bei Vorgabe-Rate über 5 Durchläufe.
set -u
S="$1"
export MSE_FREMDPAKETE=/Users/david/full-magic-pack/data
export MSE_KANON="$HOME/Library/Application Support/MagicSetEditor4Mac/kanonische-sets"
source /Users/david/MagicSetEditor4Mac-e/scripts/activate.sh >/dev/null 2>&1
summe() { go tool pprof -sample_index="$2" -top -nodecount=1 "$3" "$1" 2>/dev/null | sed -n 's/.* of \([0-9.]*[kMG]B\) total.*/\1/p; s/.* of \([0-9]*\) total.*/\1/p' | head -1; }
for stand in A1 B A2; do
  bin="$S/e03/app-${stand%[12]}.test"
  baum=/Users/david/MagicSetEditor4Mac; [ "$stand" = B ] && baum=/Users/david/MagicSetEditor4Mac-e
  for schnitt in ausfuhr werk; do
    bench='^BenchmarkWegAusfuhr$'; [ $schnitt = werk ] && bench='^BenchmarkWegKopflos$/^Karte,_frisches_Werk$'
    for n in 1 4; do
      ( cd "$baum/internal/app" && "$bin" -test.run '^$' -test.bench "$bench" -test.benchtime ${n}x -test.memprofilerate 1 -test.memprofile "$S/e03/m-$stand-$schnitt-$n.mem" >/dev/null 2>&1 )
    done
    ( cd "$baum/internal/app" && /usr/bin/time -l "$bin" -test.run '^$' -test.bench "$bench" -test.benchtime 5x > "$S/e03/t-$stand-$schnitt.txt" 2>&1 )
    echo "$stand $schnitt $(date -u +%TZ) last=$(sysctl -n vm.loadavg | tr -d '{}')" \
      "raum1=$(summe "$S/e03/m-$stand-$schnitt-1.mem" alloc_space "$bin")" \
      "raum4=$(summe "$S/e03/m-$stand-$schnitt-4.mem" alloc_space "$bin")" \
      "obj1=$(summe "$S/e03/m-$stand-$schnitt-1.mem" alloc_objects "$bin")" \
      "obj4=$(summe "$S/e03/m-$stand-$schnitt-4.mem" alloc_objects "$bin")" \
      "heap4=$(summe "$S/e03/m-$stand-$schnitt-4.mem" inuse_space "$bin")" \
      "rss=$(awk '/maximum resident set size/{print $1}' "$S/e03/t-$stand-$schnitt.txt")" \
      "spitze=$(awk '/peak memory footprint/{print $1}' "$S/e03/t-$stand-$schnitt.txt")"
  done
done
echo ENDE
```

**Ausgang**: der Rückgabewert jedes Laufs ist nicht festgehalten. Jeder
Lauf hinterließ sein Profil, und `pprof` las jedes (unten); ein
abgebrochener Lauf schriebe keines.

## Rohwerte

`raum` ist `alloc_space`, `obj` `alloc_objects`, `heap4` `inuse_space` nach
vier Durchläufen — je die Summe über das ganze Profil (`pprof -top`). **Je
Lauf** ist die Differenz aus vier und einem Durchlauf, geteilt durch drei:
so fallen Aufbau und erster Durchlauf heraus. RSS und Spitze stammen aus
`/usr/bin/time -l` über fünf Durchläufe bei gewöhnlicher Profilrate.

| Stand | Schnitt | Zeit (UTC) | raum1 | raum4 | Zuteilung je Lauf | obj1 | obj4 | Zuteilungen je Lauf | heap4 | RSS | Spitze |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| A1 | `ausfuhr` | 12:06:04Z | 3222.96MB | 7.78GB | **1581.3 MB** | 13792384 | 33814707 | **6.674 Mio.** | 7.61MB | 122.2 MB | 88.1 MB |
| A1 | `werk` | 12:07:59Z | 1094.84MB | 4.09GB | **1031.1 MB** | 1890649 | 6230210 | **1.447 Mio.** | 7.58MB | 523.8 MB | 489.4 MB |
| B | `ausfuhr` | 12:12:02Z | 1619.66MB | 3.86GB | **777.7 MB** | 6747239 | 16200985 | **3.151 Mio.** | 7.64MB | 99.9 MB | 66.0 MB |
| B | `werk` | 12:13:48Z | 258.16MB | 848.39MB | **196.7 MB** | 1884441 | 6205460 | **1.440 Mio.** | 7.57MB | 102.6 MB | 67.9 MB |
| A2 | `ausfuhr` | 12:19:51Z | 3221.34MB | 7.78GB | **1581.8 MB** | 13791690 | 33813403 | **6.674 Mio.** | 7.59MB | 126.9 MB | 93.1 MB |
| A2 | `werk` | 12:21:42Z | 1094.50MB | 4.10GB | **1034.6 MB** | 1890601 | 6230138 | **1.447 Mio.** | 7.59MB | 523.1 MB | 488.8 MB |

**Die Tafel in `LEISTUNGSSCHNITTE.md` weicht hier um bis zu einem MB ab**
(1.581,0 statt 1.581,3 MB bei A1): `pprof` gibt die Summe über vier
Durchläufe auf drei Stellen gerundet aus (`7.78GB`), und die Tafel rechnete
mit den Zahlen aus den Profilen selbst. An der Aussage — B teilt bei der
Ausfuhr rund die Hälfte zu, beim frischen Werk ein Fünftel — ändert das
nichts.

## Die Profile

Sie liegen nicht im Repositorium — zwölf Dateien zu je rund 0,6 MB —, sondern
nur ihre Prüfsummen; wer sie neu erzeugt, vergleicht gegen die Summen oben.

| Datei | SHA-256 |
| --- | --- |
| `m-A1-ausfuhr-1.mem` | `0b654630107cf6172f5f5e593174f2e368b2351cd46e5a7d5ea13752f4dfb745` |
| `m-A1-ausfuhr-4.mem` | `3dd691b53c982ba6705b6e89efa2021efd2a055fef403dbd6a1b4e308c819bf3` |
| `m-A1-werk-1.mem` | `1b1cab47dd1d9e31a41fd599fff79859e8c753c71628610606ea5c7c41945fdf` |
| `m-A1-werk-4.mem` | `00c24b1d0b97e01639dc95cbb2b7362c2e0507375b34bde6035f2682de6448bd` |
| `m-A2-ausfuhr-1.mem` | `da3ffa19d08aa81d1586b9fed6435add05c2e6dbd7a0470b3d9444edb1fe746e` |
| `m-A2-ausfuhr-4.mem` | `b037ad4fbbbba115bc23bd47c548d4c029b733875e320c634cd60d9f720879bf` |
| `m-A2-werk-1.mem` | `a56fd399981a06123fbbb3e70d36e241f72b7d7f908e72c76e2195f8288a99cb` |
| `m-A2-werk-4.mem` | `2763c8a9a1b8dd894e7219dd646431cdc96b156de237cd5bb4051028a61fd073` |
| `m-B-ausfuhr-1.mem` | `ce507efca654aa77c87eb3c7f0d39e33569ea80d74b70cee42efbc012c7f72fa` |
| `m-B-ausfuhr-4.mem` | `f79050dafcf31bcc5ba804bb25670e723aff58499fdad463ce64f71a1a068cbc` |
| `m-B-werk-1.mem` | `95eb7f6fbdd3b5e42cb5a1437cec1965bbe967450b4170e8d1839626211219b4` |
| `m-B-werk-4.mem` | `888d0fadebf866fa2ed13170bd49f7782c66ba957675bf9b93a9f364df9870d3` |
| `t-A1-ausfuhr.txt` | `3c1661df2d435274d1a36e4d2800a891bee9128e2c3a75a643677f5313006a68` |
| `t-A1-werk.txt` | `58a91df7d412b0d1dd7deed715911155f082075d0300adb5dde518d8dd4e9725` |
| `t-A2-ausfuhr.txt` | `f8a9b16b74361f66829cf02cc26e57eb03cb42bc686867c97b067b4c5b78a029` |
| `t-A2-werk.txt` | `d2845b98821211ec109ba672bd5b5b163e12c3734e60199f1aa62193beb91ca9` |
| `t-B-ausfuhr.txt` | `897a365aaeee857b06c9a90b1621d07d656838fa5e4a0a68d622df8089dbac51` |
| `t-B-werk.txt` | `b337d8dab166290feb389dfc2df1d4ceea0fd753140b14b1b8aeb918c2d57284` |

## Was dieser Beleg nicht sagt

- **Ob die Halbierung im Fenster ankommt.** Gemessen sind zwei kopflose
  Schnitte über das Testprogramm, nicht die laufende Anwendung.
- **Ob sie über andere Sets trägt.** Gemessen ist Archenemy über dem vollen
  Pack (`MSE_FREMDPAKETE`), eine Ausgabekohorte.
- **Eine ruhige Maschine.** Die Last oben gilt für RSS und Spitze.
