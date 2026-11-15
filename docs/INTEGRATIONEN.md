# Integrationen — Bestand, Lücken und was als Nächstes kommt

**Stand:** 2026-10-04 · AI-04
(#1745).

Diese Datei beantwortet zwei Fragen, und sie hält sie auseinander:
**was heute angebunden ist und wie weit es geprüft ist** (AK-01), und **was
als Nächstes angebunden werden sollte** (AK-02 bis AK-04). Die erste ist
nachgelesen und belegt; die zweite ist eine Abwägung, und **eine Recherche
erhält hier nicht den Status einer geprüften Integration** (AK-04).

## Drei Stände, und nur einer davon heißt „geprüft"

| Stand | Heißt |
| --- | --- |
| **real geprüft** | ein Lauf gegen das **Original** ist gelaufen — mit Datum, Datei und Ergebnis in [`HAERTUNG.md` §4](HAERTUNG.md#4--lauf-gegen-einen-echten-fremden-dienst) oder einem Beleg unter `docs/belege/` |
| **Sandbox** | ein Lauf gegen eine Testumgebung des Anbieters — **gibt es heute für keine Integration** |
| **Mock** | geprüft gegen **selbst geschriebene** Antworten in der gemessenen oder dokumentierten Gestalt; das Original hat diese Anwendung nie gesehen |

## Der Bestand (AK-01)

Die Zugänge liegen im **Schlüsselbund** und nie im Bestand
([0056](entscheidungen/0056-drei-weitere-kartendienste.md)); Fristen und
Proxy stehen je Transport in [`NETZEIGENSCHAFTEN.md`](NETZEIGENSCHAFTEN.md).
Die Spalte **Grenze** nennt, was **dieses Programm** einhält — unter der
Zahl des Anbieters (§16, Satz 4).

| Integration | Paket | Zugang | Umfang | Bedingungen | Grenze hier | Ohne Netz | Stand |
| --- | --- | --- | --- | --- | --- | --- | --- |
| **Scryfall** | `internal/scryfall` | keiner; `User-Agent` und `Accept` sind Pflicht | lesen: Karten, Bilder, Sammeldaten | [`DRITTANBIETER.md` §3](../DRITTANBIETER.md#scryfall) | 600 ms je Kartenanfrage, 125 ms sonst | Zwischenspeicher auf der Platte (Antworten, Bilder, Sammeldatei) | **real geprüft** |
| **magicthegathering.io** | `internal/mtgapi` | keiner | lesen: Karten | [0056](entscheidungen/0056-drei-weitere-kartendienste.md) | 5 s | — | **real geprüft** |
| **MTG JSON** (Dateien) | `internal/mtgjson` | keiner | lesen und schreiben (Dateiformat), lesen `api/v5` | [`DRITTANBIETER.md`](../DRITTANBIETER.md#mtg-json) | — | die Datei liegt lokal | **real geprüft** (2026-09-14) |
| **MTGJSON Preise** | `internal/mtgjsonpreise` | keiner | lesen: `AllPricesToday` strömend | [0093](entscheidungen/0093-eine-preisreihe-hat-sieben-teile.md) | — | Momentaufnahmen im Bestand | **real geprüft** (2026-09-14) |
| **MTGJSON GraphQL** | `internal/mtggraphql` | Marke (Patreon) | lesen: Karten | [0056](entscheidungen/0056-drei-weitere-kartendienste.md) | 10 s | — | **Mock** — das Schema am 2026-09-02 von Hand geprüft — **der Lauf gegen das Original steht bereit** (`netzlauf_test.go`, Zugang aus `MSE_NETZLAUF_…`) und meldet ohne Zugang **BLOCKED** (#2049) |
| **Scrydex** | `internal/scrydex` | Schlüssel **und** Team | lesen: Karten, nur Magic | [0056](entscheidungen/0056-drei-weitere-kartendienste.md), [0096](entscheidungen/0096-welche-optionalen-datenquellen-zugaenglich-sind.md) | 1 s (Anbieter: 100/s) | — | **Mock** — **der Lauf gegen das Original steht bereit** (`netzlauf_test.go`, Zugang aus `MSE_NETZLAUF_…`) und meldet ohne Zugang **BLOCKED** (#2049) |
| **CardTrader** | `internal/cardtrader` | Marke | lesen: vier Adressen mit `GET`; das eigene Konto nur mit `konto.schau` | [0095](entscheidungen/0095-was-cardtrader-sagt-und-was-davon-gilt.md) | 1 s; dazu die globale Grenze 200/10 s | — | **Mock** — ohne Marke 401 — **der Lauf gegen das Original steht bereit** (`netzlauf_test.go`, Zugang aus `MSE_NETZLAUF_…`) und meldet ohne Zugang **BLOCKED** (#2049) |
| **Anmeldeanbieter** (GitHub, Google, Microsoft, X) | `internal/anmeldung` | `client_id` der eigenen Anwendung beim Anbieter | Identität, nichts sonst | Vermerk zu Werkstück 4 von Phase 4.12 | eine Gesamtfrist von 30 s | — | **Mock** — Fluss und Rückruf gegen einen eigenen Server |
| **Ollama** | `internal/ollama` | keiner; lokal | Text, feste Form, Einbettung | [0098](entscheidungen/0098-was-ollama-belegt-und-was-nicht.md) | Fristen 5 s / 5 min / 60 s | ja — läuft auf diesem Rechner | **real geprüft** (2026-09-12; 2026-10-04 mit vier kleinen Modellen, `docs/belege/modellmessung.md`, Hüllenprobe `ai02`; Bildmodelle führt 0.35.1 nicht aus) |
| **Apples Gerätemodell** | `internal/applefm` | Lizenz von `fm`, angenommen von einem Menschen | Text, feste Form | [0166](entscheidungen/0166-was-fm-serve-belegt.md) | die üblichen Fristen; Start bis `/health` 10 s | ja — läuft auf diesem Rechner | **real geprüft** (2026-10-04, `docs/belege/fmlauf.md`, Hüllenprobe `ai05`) |
| **Anthropic, OpenAI, xAI** | `internal/anthropic`, `openai`, `xai` | API-Schlüssel | Text, feste Form; xAI auch Bild | [0099](entscheidungen/0099-was-anthropic-belegt.md), [0100](entscheidungen/0100-was-openai-nicht-sagt.md), [0101](entscheidungen/0101-was-xai-belegt.md) | Journal, Freigabe und Budget vor jedem Ruf | — | **Mock** — ein Lauf bräuchte ein Konto **mit Guthaben** — **der Lauf gegen das Original steht bereit** (`netzlauf_test.go`, Zugang aus `MSE_NETZLAUF_…`) und meldet ohne Zugang **BLOCKED** (#2049) |
| **Fernleitung** (tailcat) | `internal/fernleitung` | keiner; kein Konto | Verbindung zu einem Gegenüber | [0114](entscheidungen/0114-tailcat-gepinnt.md) | — | — | **Mock** über eine Naht (`net.Pipe`); **tailcat selbst ist nicht belegt** (kein DERP, kein Direktweg) |
| **Entdeckung** (mDNS) | `internal/entdeckung` | keiner | Ankündigen und Suchen im eigenen Netz | [0113](entscheidungen/0113-kein-libp2p-im-lan.md) | Frist als Wanduhr | — | **real geprüft** im Lauf „Zwei Macs auf einem" (2026-10-01) |
| **Paketbezug** | `internal/bezug` | keiner | lesen: Vorlagenpakete an einem Stand | [0025](entscheidungen/0025-die-vorlagen-duerfen-mitgeliefert-werden.md) | — | das Paketverzeichnis | **real geprüft** (2026-10-04, `internal/service/paketbezug_netzlauf_test.go`): Basic-M15-Magic-Pack am gepinnten Stand, 5.858 Dateien, Prüfsumme und Übergangenes mit Pfad; ein Zweig ist auch am Original kein Stand |
| **Aktualisierung** | `internal/aktualisierung` | ohne Marke, **seit dem 2026-11-15 öffentlich** ([0190](entscheidungen/0190-das-repositorium-wird-oeffentlich.md), Nachtrag), mit der Grenze von 60 Anfragen je Stunde, deren Erschöpfung (403 oder 429) als Drossel mit Uhrzeit gemeldet wird (R01, #2333); bis zum Öffnen brauchte es eine Marke (`GH_TOKEN`), sonst kam ein fachliches Nein | lesen: Fassung, Signatur | [0050](entscheidungen/0050-die-anwendung-ersetzt-sich-nicht-selbst.md), [0051](entscheidungen/0051-die-signatur-wird-geprueft.md) | — | — | **real geprüft** (2026-10-04, `internal/aktualisierung/netzlauf_test.go`): ohne Marke das fachliche Nein des privaten Repositoriums, mit Marke die Freigabe 0.5.0 samt **geprüfter** Signatur der Prüfsummenliste |

**Seit #2049 (2026-10-04)**: Paketbezug und Aktualisierung sind gegen das
Original gelaufen; für MTGJSON GraphQL, Scrydex, CardTrader und die drei
Cloudanbieter steht je ein Lauf bereit, der mit dem Zugang des Auftraggebers
fährt und ohne ihn BLOCKED meldet. Anmeldeanbieter und Fernleitung brauchen
eigene Anwendungen beim Anbieter beziehungsweise zwei Rechner und bleiben
**BLOCKED** (`HAERTUNG.md` §4).

**Was die Tabelle bis dahin zeigte**: von siebzehn Anbindungen waren **sieben** gegen das
Original gelaufen, **acht** nur gegen selbst geschriebene Antworten, und
**zwei** stehen in keiner Prüfliste. Bei sieben der acht liegt es am
**Zugang** (Marke, Schlüssel, Guthaben) und nicht am Willen; bei der
Fernleitung am zweiten Rechner im Internet.

## Die Lücken, aus denen eine Entscheidung folgt (AK-02)

1. **Geprüft ist weniger als angebunden.** Acht Anbindungen haben das
   Original nie gesehen; seit #2049 steht für sechs davon ein Lauf bereit,
   der mit dem Zugang des Auftraggebers fährt. Eine neue Anbindung
   vergrößerte diese Zahl, bevor sie sinkt.
2. **Bezug und Aktualisierung standen in keiner Prüfliste** — erledigt mit
   #2049: beide sind am 2026-10-04 gegen ihr Original gelaufen
   (`HAERTUNG.md` §4). Was der Lauf der Aktualisierung zeigte, bleibt eine
   Entscheidung: das Repositorium ist privat, und ohne Lesezugriff erfährt
   niemand von einer neuen Fassung — **bis es öffentlich wird**, entschieden
   am 2026-10-13 ([0190](entscheidungen/0190-das-repositorium-wird-oeffentlich.md));
   R01 (#2333)
   bindet den Kanal dann an.
3. **Preise in Euro kommen über drei Umwege** — MTGJSON (mit der Reihe
   `cardmarket`), CardTrader und Scryfall (`eur`, `eur_foil`). Einen
   direkten Dienst gibt es nicht, und **er ist heute nicht zu haben**:
   Cardmarket und TCGplayer nehmen keine neuen API-Zugänge an, nachgelesen
   am 2026-10-04 ([0168](entscheidungen/0168-kein-direkter-preisdienst.md)).
4. **Kleine Textaufgaben sind gedeckt** — durch das Gerätemodell und Ollama,
   beide auf diesem Rechner. Ein weiterer Cloudanbieter füllte keine Lücke.
5. **Teilen und Kontakte sind gebaut** — über den Teilen-Dialog von macOS
   (NET-04) und vCard (ID-04). Hier fehlt keine Integration.

## Der Vergleich (AK-03)

| Kandidat | Nutzen | Machbarkeit | Kosten | Daten nach außen | Wartung | Ausstieg |
| --- | --- | --- | --- | --- | --- | --- |
| **Echtläufe für die acht Mock-Anbindungen** | hoch: aus „gebaut" wird „geprüft" | hoch — die Läufe hinter `netzlauf` haben eine Form | ein Zugang je Dienst; bei Cloud-KI Guthaben | nur, was ein Lauf fragt | ein Lauf je Fassung des Anbieters | keiner nötig — es ist Prüfung, nicht Bindung |
| **Läufe für Bezug und Aktualisierung** | mittel: zwei Wege ohne jeden Beleg | hoch — keine Zugänge nötig | keine | keine | gering | — |
| **Ein direkter Preisdienst** (Cardmarket/TCGplayer) | mittel: Euro-Preise ohne Umweg | **nein** — beide nehmen keine neuen Zugänge an, nachgelesen am 2026-10-04 ([0168](entscheidungen/0168-kein-direkter-preisdienst.md)) | — | — | — | — |
| **Ein weiterer Cloud-KI-Anbieter** | gering: die kleinen Aufgaben sind lokal gedeckt | hoch — drei Muster liegen da | Guthaben | Kartentexte | ein Blatt mehr, ohne Echtlauf | hoch |
| **Scraping eines Preisportals** | — | **abgelehnt**: keine Nutzungsgrundlage, und eine solche wird nicht als API-Anbindung behandelt (AK-03) | — | — | — | — |

## Die Entscheidungsliste (AK-04)

1. **Erst prüfen, dann anbinden.** Die nächste Arbeit an Integrationen sind
   **Echtläufe** für die Anbindungen, die heute nur Mocks kennen — mit dem
   Zugang des Auftraggebers, wo es einen gibt. → #2049.
2. **Bezug und Aktualisierung bekommen einen Lauf hinter `netzlauf`** und
   ihre Zeile in `HAERTUNG.md` §4. → Teil von #2049.
3. **Ein direkter Preisdienst ist eine Recherche**, kein Bauauftrag: erst
   die Bedingungen nachlesen (Zugang, Nutzung, Grenzen), dann entscheiden.
   → #2050,
   **entschieden**: abgelehnt, weil keiner der beiden neue Zugänge vergibt
   ([0168](entscheidungen/0168-kein-direkter-preisdienst.md)); kippt mit
   einer wieder geöffneten Anmeldung.
4. **Kein weiterer Cloud-KI-Anbieter** in dieser Iteration und der
   nächsten — die Lücke, die er füllte, gibt es nicht.
5. **Kein Scraping.**

**Was diese Liste nicht ist**: eine Zusage, dass irgendetwas davon gebaut
wird. Sie sagt, in welcher Reihenfolge es sich lohnt, darüber zu reden.
