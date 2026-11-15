# 0190 — Das Repositorium wird öffentlich, und jede Freigabe trägt ihren Quelltext

**Stand:** angenommen am 2026-10-13 · #2331 (D01).
**Entschieden vom Auftraggeber** am 2026-10-13: „Dieses Repository wird später
öffentlich gestellt." **Umgesetzt** ist damit nur der Vertrag; die
Sichtbarkeit ändert ein Mensch, und zwar erst mit einem gesonderten Auftrag.

## Lage

Befund A-08 aus dem Audit zur Produktionsreife 0.9.0
(Zweitblick (Entwicklungsarchiv)): das Repositorium ist
privat, `GET /repos/DavidTHeider/MagicSetEditor4Mac/releases/latest`
antwortet **ohne Anmeldung mit 404**. Damit erreicht die eingebaute
Aktualisierungsprüfung (`internal/aktualisierung/klient.go`) für Anwendende
keine Fassung, und ein Quelltextangebot zur GPL gibt es nur für den, der ohnehin
Zugang hat.

Zur Wahl standen nach dem Plan (D01): das **bestehende Repositorium
öffentlich**, ein **getrenntes öffentliches Freigaberepositorium** mit
vollständigem Quelltext, oder ein **eigener öffentlicher Kanal**.

## Entscheidung

**Das bestehende Repositorium wird öffentlich**, und seine Freigaben sind der
Kanal. Daraus folgt der Vertrag dieses Vermerks:

| Was | Festlegung |
| --- | --- |
| **Kanal** | `https://github.com/DavidTHeider/MagicSetEditor4Mac/releases` — derselbe, den `aktualisierung.Repositorium` heute schon fragt; **keine Änderung am Klienten** |
| **Herausgeber und Verantwortlicher** | David T. Heider, Inhaber des Repositoriums |
| **Signatur** | jede Freigabe trägt `pruefsummen.txt` und `pruefsummen.txt.asc` mit dem Schlüssel `BF81923EA5D38635` (`internal/aktualisierung/signatur.go`); `pruefsummen.txt` nennt **auch das Quelltextpaket** |
| **Quelltext je Binärstand** | eine Beigabe derselben Freigabe: `MagicSetEditor4Mac-<fassung>-quelltext.tar.gz` — der Baum am signierten Tag `v<fassung>`, dazu die Go-Module (`vendor/`) und die npm-Pakete, die ins Bündel gehen, samt Bauanleitung (`README.md`, `docs/RELEASE.md`) und Lizenzhinweisen (`LICENSE`, `DRITTANBIETER.md`, `docs/ABHAENGIGKEITEN.md`); dazu `QUELLEN.md` mit Fassung, Lizenzdatei beziehungsweise Lizenzangabe und Ursprung jeder Abhängigkeit, eine Tafel der **erzeugt eingebetteten Formen** — SQLite als WebAssembly, `runtime.js` von Wails, Tailscales vorgebaute Weboberfläche samt der Schrift Inter — und deren Bauquellen unter `quellen/` |
| **Adresse des Quelltexts** | `https://github.com/DavidTHeider/MagicSetEditor4Mac/releases/tag/v<fassung>` — dieselbe Seite wie das Abbild |

**Warum eine eigene Beigabe und nicht nur der Tag.** GPL-2.0 §3 nennt
vollständigen Quelltext „all the source code for all modules it contains" —
und ein Go-Programm enthält die Module, die es statisch einbindet, ebenso die
Oberfläche die npm-Pakete in ihrem Bündel. Das Archiv, das GitHub zu einem Tag
bildet, enthält nur den eigenen Baum. **Und §3 letzter Absatz** sagt, dass
gleichwertiger Zugang zum Quelltext **an derselben Stelle** als Weitergabe
zählt: deshalb liegt das Paket neben dem Abbild und nicht auf einem zweiten
Weg.

**Was noch offen ist** (Befund 1 des Zweitblicks und des Nachlaufs über
Phase 0): die Abhängigkeiten liegen in der Form bei, in der ihre Pakete
erscheinen, und für einige ist das nicht die Form zum Bearbeiten — die
gebauten npm-Pakete, die Quelle von SQLite selbst, die Schrift Inter und der
Stand von Tailscales Weboberfläche, den das vorgebaute Modul nicht nennt.
`QUELLEN.md` führt jede dieser Formen mit dem, was beiliegt, und dem, was
fehlt. **Offen ist nicht, ob sie beiliegen, sondern wie sie beschafft
werden** — #2393,
und es ist eine **Voraussetzung der ersten Freigabe** an Anwendende, nicht
eine Wahl. **Seit dem 2026-10-16 beantwortet** — Nachtrag unten.

**Was ausdrücklich nicht hineingehört**: die mitgelieferten Vorlagen aus dem
bezogenen Pack (`scripts/mitliefern.sh`). Sie sind Daten und kein Modul des
Programms — „mere aggregation" nach §2 letzter Absatz — und sie stehen unter
den Bedingungen ihrer Rechteinhaber (`DRITTANBIETER.md`), nicht unter der GPL.

## Verworfen

- **Ein getrenntes Freigaberepositorium.** Es verdoppelte die Pflege — Tags,
  Notizen, Signatur —, und der Quelltext darin müsste trotzdem vollständig
  sein. Der Auftraggeber hat das bestehende gewählt.
- **Ein eigener Kanal** (eigene Domain, eigener Server). Er brächte Betrieb,
  Zertifikate und Kosten, und der Klient bräuchte eine zweite Gegenstelle.
- **Ein schriftliches Angebot nach §3 b.** Drei Jahre Bereitschaft, auf
  Anfrage Quelltext zu schicken, ist mehr Pflicht als ein Archiv neben dem
  Abbild — und prüfbar ist nur das Archiv.
- **Nur der Tag als Quelltext.** Siehe oben: er enthält die eingebundenen
  Module nicht.

## Folgen

- **R01** (#2333)
  bindet den Kanal an: der Klient bleibt, die Fehlerwege 404, 429, Frist und
  falsche Signatur werden getrennt geprüft, und `scripts/release.sh` erzeugt,
  prüft und signiert das Quelltextpaket.
- **Bis zum Umstellen der Sichtbarkeit bleibt A-08 offen**: ein Gast ohne
  Zugang sieht weiter 404. Das Umstellen ist eine Handlung eines Menschen in
  den Einstellungen des Repositoriums und gehört zum Freigabegate (REL,
  #2385);
  HAND6 (#2384)
  belegt danach den Bezug ohne Zugang.
- **Vor dem Öffnen** ist das Repositorium auf Inhalte zu prüfen, die nicht
  öffentlich gehören: Geheimnisse, fremde Vorlagen (§11), personenbezogene
  Daten in Belegen. Das ist ein eigener Schritt im Freigabegate und kein
  Nebensatz.
- Ein Wächter hält den Vertrag: `internal/werkzeuge/quelltextangebot_test.go`.

## Nachtrag vom 2026-10-16: wie die Bearbeitungsformen beiliegen

**Entschieden vom Auftraggeber am 2026-10-16** in
#2393: Weg 1
— beziehen und beilegen, je Freigabe. Zur Wahl standen außerdem, nur SQLite
beizulegen und die veröffentlichten npm-Pakete als Bearbeitungsform
festzuhalten, oder die Frage bis zur Freigabe offen zu lassen.

**Was jetzt gilt**

- **Tailscales Weboberfläche samt Inter und die Diagnoseseite des
  Ereignisbusses mit htmx stecken in keinem Programm.** Die Bauschalter
  `ts_omit_webclient` und `ts_omit_debugeventbus` der gebundenen Fassung von
  tailscale.com lassen sie weg; diese Anwendung zeigt keine der beiden. Die
  zweite stand in der Inventur vom 2026-10-13 nicht — gefunden beim Zählen
  der Einbettungen. Je Programm rund 1,5 MB weniger (`mse` 53,5 → 52,0 MB).
  Die Liste steht in `scripts/releaseschalter.sh`; `release.sh` prüft sie am
  fertigen Programm und an jedem Werkzeug.
- **Je npm-Paket im Bündel liegt das Repositorium an dem Commit bei, aus
  dem es veröffentlicht ist** — `gitHead` aus der Registry, sonst das Tag —,
  und die **SQLite-Quelle**, genau die Dateien, die `build/download.sh` von
  go-sqlite3-wasm lädt. **32 Archive, 77,0 MB je Freigabe**, unter
  `quellen/bearbeitungsformen/`; `QUELLEN.md` nennt je Paket Herkunft und Ziel.
  Erhoben werden sie mit `node scripts/bearbeitungsformen.mjs` bei jeder
  Hebung, beim Freigabebau geladen und gegen ihre SHA-256-Summe geprüft — ein
  Archiv, das sich auf dem Server geändert hat, hält die Freigabe an.
- **Zwei Pakete liegen nicht an ihrem eigenen Commit**, und beide sind
  benannt: `react-remove-scroll-bar` 2.3.8 und `react-style-singleton` 2.2.3
  sind aus Ständen veröffentlicht, die nie ins Repositorium kamen. Gegen das
  letzte Tag ändern sie nur `package.json` — Fassung und Bereiche der
  Peer-Abhängigkeiten —, der Code ist bytegleich (verglichen am 2026-10-16);
  beigelegt ist deshalb dieses Tag. **`@wailsio/runtime`** liegt schon über
  die Go-Fassung von Wails unter `quellen/` bei. **Sechs Pakete tragen nur
  Typen** (`@types/*`, `csstype` …) und gehen nicht ins Bündel.
- **Drei Wächter**: `einbettungen.test.sh` (jede Einbettung eines fremden
  Pakets ist erzeugt oder Quelle, mit Gegenprobe ohne die Schalter),
  `bearbeitungsformen.test.mjs` (die Liste gegen Bündel, `go.mod` und
  `download.sh`, ohne Netz), `releasequelltext.test.sh` (Beilegen, falsche
  Summe, verdorbener Speicher, keine offene Zelle in der Tafel).

**Was das nicht beantwortet**: ob ein Archiv in Jahren noch unter seiner
Adresse liegt. Deshalb reist es im Quelltextpaket jeder Freigabe mit — das
Paket ist die Zusage, nicht die Adresse.


## Nachtrag vom 2026-11-15: „das bestehende" ist ein neues unter demselben Namen

**Mit 1.0.0 ist das Repositorium neu angelegt**, unter demselben Namen; das
bisherige heißt `MagicSetEditor4Mac-archiv` und bleibt privat
([0205](0205-1-0-0-ist-ein-neuanfang.md)). Der Vertrag dieses Vermerks gilt
unverändert für das neue — Kanal, Herausgeber, Quelltext je Binärstand, Adresse
des Quelltexts —, mit einer Ausnahme: **signiert wird mit dem
Freigabeschlüssel `44C59428DE1F766E`**, nicht mehr mit `BF81923EA5D38635`.
Der Grund für das neue Repositorium ist der Schritt „Vor dem Öffnen“ oben: die
alte Historie trug eine private Adresse, und über die PRs wäre sie erreichbar
geblieben.
