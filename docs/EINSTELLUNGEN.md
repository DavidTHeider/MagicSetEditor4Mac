# Einstellungen — das Schema

**Welche Einstellungen es gibt, was sie annehmen, wem sie gehören und in
welcher Reihenfolge sie sich übersteuern** — OPS-01.AK-01
(#1718).

**Die Tafel unten ist gezählt, nicht abgeschrieben.**
[`internal/werkzeuge/einstellungsschema_test.go`](../internal/werkzeuge/einstellungsschema_test.go)
erzeugt sie aus `internal/domain` und vergleicht sie Zeile für Zeile; eine
neue Einstellung ohne Zeile hier macht ihn rot, und seine Meldung nennt die
Zeile, die fehlt. **Was er nicht hält**: ob die Prosa darüber noch stimmt —
die Vorrangfolge steht im Code an drei Stellen, die hier benannt sind.

## Was eine Zeile sagt

| Spalte | Frage | Quelle |
| --- | --- | --- |
| **Schlüssel** | unter welchem Namen die Einstellung gespeichert, gesetzt und ausgetauscht wird | `internal/domain/einstellung.go` |
| **Bereich** | wo im Fenster sie steht | `EinstellungsBereich` |
| **Ebene** | wem sie gehört — und damit, **wie weit sie reist** | `EbeneVon`, [0079](entscheidungen/0079-einstellungsebenen-und-portabler-import.md) |
| **Werte** | was sie annimmt: eine **Liste** in Anzeigereihenfolge, *geprüft* für eine Regel, die sich nicht aufzählen lässt (eine Adresse, eine Breite, eine Zone), oder *frei* — dann wird nur das Nullbyte abgewiesen | `ErlaubteEinstellungswerte`, `EinstellungsArt` |
| **Vorgabe** | was gilt, solange niemand gewählt hat — leer heißt „wie das System" | `EinstellungsVorgabe` |
| **Wirkung** | wann ein geänderter Wert gilt: `sofort` oder `neustart` | `EinstellungsWirkung` (UX-04.AK-02) |

**Die Ebenen**, von außen nach innen:

| Ebene | Wem sie gehört | Reist in einem Austausch |
| --- | --- | --- |
| `geraet` | diesem Rechner und dieser Installation | **nie** — eine Datei mit diesem Abschnitt bekommt je Zeile den Befund „beschreibt einen Rechner" |
| `nutzer` | dem Menschen, in jedes seiner Profile | ja |
| `profil` | der Arbeitsumgebung | ja |
| `arbeitsplatz` | dem Fenster: welche Spalte steht und wie breit | **nur auf Wunsch** |

## Der Vorrang

**Von unten nach oben; was weiter oben gesetzt ist, gilt.**

1. **Die Vorgabe** aus der Tafel.
2. **Der Wert des Menschen** (`nutzereinstellungen`).
3. **Der Wert des Profils** (`settings`). Für die vier übersteuerbaren
   Einstellungen — Sprache, Erscheinungsbild, Schriftgröße, Schriftart —
   steht dort `erben` als Vorgabe, und `erben` reicht zum Menschen durch
   (`internal/domain/uebersteuerung.go`). **Was gilt und wessen Entscheidung
   es war**, sagen `wirksam` und `herkunft` an jeder Einstellung — im Fenster,
   in `mse einstellungen`, über REST und über MCP dieselbe Antwort
   (OPS-01.AK-04).

**Eine Schranke sticht nicht, sie faltet** — für die Dienstschalter,
`richtlinie.*` und `sicherheit.faktorbeimwechsel` gilt der strengste
gesetzte Wert beider Ebenen
([0079](entscheidungen/0079-einstellungsebenen-und-portabler-import.md),
`internal/domain/richtlinie.go`). **Und „nur lokal" legt genau die
Dienstschalter auf `aus`** — `dienst.<name>` für jeden fremden Dienst, mit
der Herkunft der Schranke. **Was unter einem Dienst steht, schaltet nichts
und behält seinen Wert**: die Adresse eines Anbieters
(`dienst.<name>.adresse`) und die vier Pfade des Dienstzeugnisses
(`dienst.tls.*`). Bis zum 2026-10-02 traf die Schranke jeden Schlüssel auf
`dienst.`, und ein Dienst jenseits der Schleife suchte ein Zertifikat namens
„aus" (#2004;
`mitNurLokal` in `internal/domain/geltung.go`).

**Und was ausdrücklich keine Ebene ist:**

- **Eine Austauschdatei.** Sie gilt nicht, solange sie daliegt: sie wird
  gelesen, als Unterschied gezeigt und erst mit der Übernahme zu
  **gespeicherten** Werten der Stufen 2 und 3 — atomar, gegen die Revision
  der Vorschau (`internal/service/einstellungseinfuhr.go`).
- **Die Umgebung** übersteuert **eine** Familie, und dort liegt sie
  **unter** der Einstellung: `netz.proxy.*` leer heißt „`HTTP_PROXY`,
  `HTTPS_PROXY` und `NO_PROXY` gelten", gesetzt schlägt die Einstellung sie
  (`internal/domain/netzeinstellung.go`). Die übrigen Umgebungsvariablen
  wählen **Orte**, keine Einstellungen — `MSE_BESTAND` den Bestand,
  `MSE_SCHLUESSELBUND` den Schlüsselbund, `MSE_DIENSTORDNER` weitere Ordner
  für einen offenen Server, die
  Variablen auf `_ADRESSE` die Adresse eines fremden Dienstes für Prüfläufe;
  sie stehen in [`ADMINISTRATION.md`](ADMINISTRATION.md). **Eine gehört
  daneben und wählt keinen Ort**: `MSE_NETZ_SCHLEIFE=an` öffnet für
  Prüfläufe die Schleife als Paketquelle (SEC-P4, #2000) — eine Einstellung
  könnte setzen, wer schreiben darf, und gerade gegen den steht die Sperre.
- **Die Kommandozeile** übersteuert nichts. `mse einstellungen setzen`
  schreibt einen **gespeicherten** Wert auf derselben Stufe wie das Fenster.
  **Mit Absicht**: ein Schalter, der nur für einen Lauf gilt, ließe
  Kommandozeile und Fenster zur selben Zeit verschiedene wirksame Werte
  zeigen — genau das, was AK-04 ausschließt.

## Wo ein Wert liegt — und warum nicht nach seiner Ebene

**Die Ebene sagt, wem eine Frage gehört und wie weit sie reist, nicht in
welcher Tafel ihr Wert liegt.** `settings` hält alles aus der ersten Tafel
unten **je Profil** — auch, was `geraet` oder `nutzer` ist —,
`nutzereinstellungen` die zweite **je Mensch**. So steht es seit
[0079](entscheidungen/0079-einstellungsebenen-und-portabler-import.md), und die
Einstellungsseite sagt es an den betroffenen Zeilen („Gespeichert wird sie je
Profil").

**Nachgeprüft am 2026-09-26 für Befund M13 des Zweitblicks über Phase 6.3**
(#1665),
und die Entscheidung bleibt: ein Umzug der Geräte- und Nutzerwerte aus
`settings` wäre eine Wanderung über jeden Bestand, und was sie brächte —
derselbe Wert in einem zweiten Profil —, bringt heute schon der Austausch
oder ein zweites Setzen. **Was es kostet, gehört ausgesprochen**: wer zwei
Profile führt, stellt die Fenstergröße oder den Proxy in jedem einzeln ein.

## Die Fassung

**Eine Austauschdatei trägt `fassung`** — heute `1`, geführt in
[`FASSUNGEN.md`](FASSUNGEN.md) als `AustauschFassung` mit Aufstieg und
Abstieg. Eine Datei aus 0.5.0 liegt als Vorrichtung unter
`testdaten/fassung-0.5.0/einstellungen.json` und wird bei jedem Lauf
übernommen.

## Geheimnisse und was an einem Rechner hängt

**Kein Geheimnis steht in einer Einstellung.** Zugangsdaten fremder Dienste,
das Geheimnis eines Anmeldeanbieters und der Schlüssel des Bestands liegen im
Schlüsselbund von macOS ([`ADMINISTRATION.md`](ADMINISTRATION.md)); eine
Einstellung nennt höchstens, **dass** es eines gibt. Eine Ausfuhr kann deshalb
keines enthalten (OPS-01.AK-03).

**Was an diesem Rechner hängt, ist `geraet`** — die Kennung, unter der sich
diese Installation bei einem Anbieter ausweist, ein Pfad auf dieser Platte.
Es reist nicht mit, und drüben wird es dort gesetzt, wo es hingehört. **Die
Kontoschau reist ebenfalls nicht**, obwohl sie `profil` ist: wer eine Datei
weitergäbe, gäbe eine Freigabe weiter, die niemand erteilt hat.

## Die Tafel des Profils und des Geräts

**Alles, was `mse einstellungen` zeigt**, in der Reihenfolge der
Einstellungsseite. Die Ebene sagt, ob eine Zeile reist; eine leere Vorgabe
heißt „wie das System" oder „ohne".

| Schlüssel | Bereich | Ebene | Werte | Vorgabe | Wirkung |
| --- | --- | --- | --- | --- | --- |
| `sprache` | `anwendung` | `nutzer` | `erben`, `de`, `en`, `es`, `fr`, `it`, `ja`, `ru`, `zh` | `erben` | sofort |
| `sprache.eigene` | `anwendung` | `nutzer` | *geprüft* | — | sofort |
| `darstellung` | `anwendung` | `nutzer` | *geprüft* | — | sofort |
| `zeitzone` | `anwendung` | `geraet` | *geprüft* | — | sofort |
| `filter.gespeichert` | `anwendung` | `profil` | *geprüft* | — | sofort |
| `kuerzel.ziffern` | `anwendung` | `nutzer` | *geprüft* | — | sofort |
| `erscheinungsbild` | `anwendung` | `nutzer` | `erben`, `system`, `hell`, `dunkel` | `erben` | sofort |
| `schriftgroesse` | `anwendung` | `nutzer` | `erben`, `klein`, `normal`, `gross`, `sehrgross` | `erben` | sofort |
| `schriftart` | `anwendung` | `nutzer` | `erben`, `pergament`, `buch`, `rund`, `serifenlos`, `fantasy`, `antike`, `kalligrafie` | `erben` | sofort |
| `farbwelt` | `anwendung` | `profil` | `pergament`, `klar`, `schiefer`, `moos` | `pergament` | sofort |
| `schriftbereich` | `anwendung` | `profil` | `ueberschriften`, `ganz` | `ueberschriften` | sofort |
| `hintergrund` | `anwendung` | `profil` | `aus`, `fuellen`, `einpassen`, `kacheln` | `aus` | sofort |
| `gestaltung` | `anwendung` | `profil` | *geprüft* | — | sofort |
| `aktualisierung` | `anwendung` | `geraet` | `aus`, `melden`, `holen` | `aus` | neustart |
| `ersteschritte` | `anwendung` | `geraet` | `an`, `aus` | `an` | sofort |
| `seitenleiste` | `anwendung` | `geraet` | `sichtbar`, `verborgen` | `sichtbar` | sofort |
| `leistenlage` | `anwendung` | `geraet` | `senkrecht`, `waagerecht` | `senkrecht` | sofort |
| `fensterstart` | `anwendung` | `geraet` | `maximiert`, `normal`, `zuletzt` | `maximiert` | neustart |
| `fensterbreite` | `anwendung` | `geraet` | *geprüft* | — | neustart |
| `fensterhoehe` | `anwendung` | `geraet` | *geprüft* | — | neustart |
| `vorgabe.zustand` | `vorgaben` | `profil` | `ohne`, `mint`, `nearMint`, `excellent`, `good`, `lightPlayed`, `played`, `poor` | `ohne` | sofort |
| `vorgabe.waehrung` | `vorgaben` | `profil` | *geprüft* | — | sofort |
| `vorgabe.einzelspiel` | `vorgaben` | `profil` | `aus`, `an` | `aus` | sofort |
| `vorgabe.hochladepfad` | `vorgaben` | `geraet` | *geprüft* | — | sofort |
| `vorgabe.ausfuhrpfad` | `vorgaben` | `geraet` | *geprüft* | — | sofort |
| `vorgabe.sicherungspfad` | `vorgaben` | `geraet` | *geprüft* | — | sofort |
| `vorgabe.diagnosepfad` | `vorgaben` | `geraet` | *geprüft* | — | sofort |
| `ki.laufzeit` | `dienste` | `profil` | `aus`, `an` | `aus` | sofort |
| `ki.geraetemodell` | `dienste` | `profil` | `automatisch`, `aus` | `automatisch` | sofort |
| `ki.vorrangmodell` | `dienste` | `profil` | *geprüft* | — | sofort |
| `ki.adresse` | `dienste` | `geraet` | *geprüft* | — | sofort |
| `konto.schau` | `dienste` | `profil` | `aus`, `an` | `aus` | sofort |
| `netz.proxy.http` | `netz` | `geraet` | *geprüft* | — | sofort |
| `netz.proxy.https` | `netz` | `geraet` | *geprüft* | — | sofort |
| `netz.proxy.aus` | `netz` | `geraet` | *frei* | — | sofort |
| `netz.frist.verbindung` | `netz` | `geraet` | *geprüft* | — | sofort |
| `netz.frist.kopf` | `netz` | `geraet` | *geprüft* | — | sofort |
| `netz.frist.stillstand` | `netz` | `geraet` | *geprüft* | — | sofort |
| `netz.schnittstelle` | `netz` | `geraet` | *geprüft* | — | sofort |
| `netz.lauschen` | `netz` | `geraet` | *geprüft* | — | neustart |
| `dienst.tls.zertifikat` | `netz` | `geraet` | *geprüft* | — | neustart |
| `dienst.tls.schluessel` | `netz` | `geraet` | *geprüft* | — | neustart |
| `dienst.tls.kette` | `netz` | `geraet` | *geprüft* | — | neustart |
| `dienst.tls.anker` | `netz` | `geraet` | *geprüft* | — | neustart |
| `grenzen.kontingente` | `netz` | `geraet` | `an`, `aus` | `an` | neustart |
| `bildkonflikt` | `bildausfuhr` | `profil` | `altbehalten`, `ueberschreiben`, `nummerieren`, `nummerierenneu` | `nummerierenneu` | sofort |
| `tisch.autoanordnung` | `anwendung` | `nutzer` | `an`, `aus` | `an` | sofort |
| `tisch.beschriftung` | `anwendung` | `nutzer` | `aus`, `klein`, `normal`, `gross` | `normal` | sofort |
| `bildausfuhr.ordner` | `bildausfuhr` | `geraet` | *geprüft* | — | sofort |
| `arbeitsplatz.navigator` | `arbeitsplatz` | `arbeitsplatz` | `erben`, `sichtbar`, `verborgen` | `erben` | sofort |
| `arbeitsplatz.inspektor` | `arbeitsplatz` | `arbeitsplatz` | `erben`, `sichtbar`, `verborgen` | `erben` | sofort |
| `arbeitsplatz.navigatorbreite` | `arbeitsplatz` | `arbeitsplatz` | *geprüft* | `erben` | sofort |
| `arbeitsplatz.inspektorbreite` | `arbeitsplatz` | `arbeitsplatz` | *geprüft* | `erben` | sofort |
| `arbeitsplatz.ansicht.sets` | `arbeitsplatz` | `arbeitsplatz` | `erben`, `kacheln`, `liste` | `erben` | sofort |
| `arbeitsplatz.ansicht.decks` | `arbeitsplatz` | `arbeitsplatz` | `erben`, `liste`, `kacheln` | `erben` | sofort |
| `arbeitsplatz.ansicht.sammlung` | `arbeitsplatz` | `arbeitsplatz` | `erben`, `tabelle`, `galerie`, `gruppen` | `erben` | sofort |
| `ausgabe.ziele` | `arbeitsplatz` | `arbeitsplatz` | *geprüft* | — | sofort |
| `schalter.sammeln` | `schalter` | `geraet` | `an`, `aus` | `an` | sofort |
| `schalter.spielen` | `schalter` | `geraet` | `an`, `aus` | `an` | sofort |
| `schalter.handeln` | `schalter` | `geraet` | `an`, `aus` | `an` | sofort |
| `schalter.ressourcen` | `schalter` | `geraet` | `an`, `aus` | `an` | sofort |
| `schalter.ki` | `schalter` | `geraet` | `an`, `aus` | `an` | sofort |
| `schalter.netz` | `schalter` | `geraet` | `an`, `aus` | `an` | sofort |
| `schalter.markt` | `schalter` | `geraet` | `an`, `aus` | `an` | sofort |
| `schalter.turniere` | `schalter` | `geraet` | `an`, `aus` | `an` | sofort |
| `schalter.dienst` | `schalter` | `geraet` | `an`, `aus` | `an` | sofort |
| `schalter.werkzeugserver` | `schalter` | `geraet` | `an`, `aus` | `an` | sofort |
| `dienst.scryfall` | `dienste` | `profil` | `aus`, `an` | `an` | sofort |
| `dienst.mtgapi` | `dienste` | `profil` | `aus`, `an` | `aus` | sofort |
| `dienst.mtgjson` | `dienste` | `profil` | `aus`, `an` | `aus` | sofort |
| `dienst.scrydex` | `dienste` | `profil` | `aus`, `an` | `aus` | sofort |
| `dienst.cardtrader` | `dienste` | `profil` | `aus`, `an` | `aus` | sofort |
| `dienst.anthropic` | `dienste` | `profil` | `aus`, `an` | `aus` | sofort |
| `dienst.anthropic.adresse` | `dienste` | `profil` | *geprüft* | — | sofort |
| `dienst.openai` | `dienste` | `profil` | `aus`, `an` | `aus` | sofort |
| `dienst.openai.adresse` | `dienste` | `profil` | *geprüft* | — | sofort |
| `dienst.xai` | `dienste` | `profil` | `aus`, `an` | `aus` | sofort |
| `dienst.xai.adresse` | `dienste` | `profil` | *geprüft* | — | sofort |
| `anbieter.github.kennung` | `anmeldung` | `geraet` | *frei* | — | sofort |
| `anbieter.google.kennung` | `anmeldung` | `geraet` | *frei* | — | sofort |
| `anbieter.microsoft.kennung` | `anmeldung` | `geraet` | *frei* | — | sofort |
| `anbieter.x.kennung` | `anmeldung` | `geraet` | *frei* | — | sofort |
| `frage.suche` | `dienste` | `profil` | `scryfall`, `mtgapi`, `mtgjson`, `scrydex` | `scryfall` | sofort |
| `protokoll.stufe` | `anwendung` | `geraet` | `fehler`, `warnung`, `info`, `ausfuehrlich` | `info` | neustart |
| `protokoll.datei` | `anwendung` | `geraet` | `an`, `aus` | `an` | neustart |
| `protokoll.system` | `anwendung` | `geraet` | `an`, `aus` | `an` | neustart |
| `protokoll.ausgabe` | `anwendung` | `geraet` | `aus`, `an` | `aus` | neustart |
| `protokoll.format` | `anwendung` | `geraet` | `text`, `json` | `text` | neustart |
| `protokoll.grenze` | `anwendung` | `geraet` | *geprüft* | `4` | neustart |
| `protokoll.aufbewahrung` | `anwendung` | `geraet` | *geprüft* | `1` | neustart |
| `protokoll.komponenten` | `anwendung` | `geraet` | *geprüft* | — | neustart |
| `protokoll.ort` | `anwendung` | `geraet` | *geprüft* | — | neustart |
| `messung.spur` | `anwendung` | `geraet` | `aus`, `an` | `aus` | neustart |
| `messung.abtastung` | `anwendung` | `geraet` | *geprüft* | `100` | neustart |
| `messung.spurgrenze` | `anwendung` | `geraet` | *geprüft* | `1000` | neustart |
| `messung.aufbewahrung` | `anwendung` | `geraet` | *geprüft* | `60` | neustart |
| `messung.profile` | `anwendung` | `geraet` | *geprüft* | `3` | sofort |
| `zwischenspeicher.grenze` | `anwendung` | `geraet` | *geprüft* | `512` | neustart |
| `sicherung.planordner` | `sicherung` | `geraet` | *geprüft* | — | sofort |
| `richtlinie.nurlokal` | `richtlinien` | `nutzer` | `aus`, `an` | `aus` | sofort |
| `richtlinie.kostengrenze` | `richtlinien` | `nutzer` | *geprüft* | — | sofort |
| `richtlinie.handel` | `richtlinien` | `nutzer` | `an`, `aus` | `an` | sofort |
| `richtlinie.handel.markt` | `richtlinien` | `nutzer` | `aus`, `an` | `aus` | sofort |
| `richtlinie.handel.menge` | `richtlinien` | `nutzer` | *geprüft* | — | sofort |
| `richtlinie.handel.wert` | `richtlinien` | `nutzer` | *geprüft* | — | sofort |
| `richtlinie.handel.frist` | `richtlinien` | `nutzer` | *geprüft* | — | sofort |
| `feldtiefe` | `anwendung` | `profil` | `normal`, `erweitert`, `experte` | `normal` | sofort |

## Die Tafel des Menschen

**Was ein Mensch in jedes seiner Profile mitnimmt.** Vier Schlüssel stehen
in beiden Tafeln — Sprache, Erscheinungsbild, Schriftgröße, Schriftart —,
weil dieselbe Frage auf zwei Ebenen gestellt wird: hier der Boden, dort die
Übersteuerung mit `erben` als Vorgabe.

| Schlüssel | Werte | Vorgabe |
| --- | --- | --- |
| `barrierefreiheit.bewegung` | `system`, `reduziert` | `system` |
| `barrierefreiheit.kontrast` | `system`, `hoch` | `system` |
| `sicherheit.faktorbeimwechsel` | `verlangen`, `nurAnmelden` | `verlangen` |
| `sprache` | `de`, `en`, `es`, `fr`, `it`, `ja`, `ru`, `zh` | — |
| `erscheinungsbild` | `system`, `hell`, `dunkel` | `system` |
| `schriftgroesse` | `klein`, `normal`, `gross`, `sehrgross` | `normal` |
| `schriftart` | `pergament`, `buch`, `rund`, `serifenlos`, `fantasy`, `antike`, `kalligrafie` | `pergament` |
| `arbeitsplatz.navigator` | `sichtbar`, `verborgen` | `sichtbar` |
| `arbeitsplatz.inspektor` | `sichtbar`, `verborgen` | `sichtbar` |
| `arbeitsplatz.navigatorbreite` | *geprüft* | `240` |
| `arbeitsplatz.inspektorbreite` | *geprüft* | `320` |
| `arbeitsplatz.ansicht.sets` | `kacheln`, `liste` | `kacheln` |
| `arbeitsplatz.ansicht.decks` | `liste`, `kacheln` | `liste` |
| `arbeitsplatz.ansicht.sammlung` | `tabelle`, `galerie`, `gruppen` | `tabelle` |
| `richtlinie.nurlokal` | `aus`, `an` | `aus` |
| `richtlinie.kostengrenze` | *geprüft* | — |
| `richtlinie.handel` | `an`, `aus` | `an` |
| `richtlinie.handel.markt` | `aus`, `an` | `aus` |
| `richtlinie.handel.menge` | *geprüft* | — |
| `richtlinie.handel.wert` | *geprüft* | — |
| `richtlinie.handel.frist` | *geprüft* | — |
