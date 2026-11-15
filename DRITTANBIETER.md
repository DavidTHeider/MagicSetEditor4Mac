# Fremde Rechte, fremder Code, fremde Inhalte

Diese Datei sagt, was in diesem Projekt **nicht** uns gehört und unter welchen
Bedingungen es benutzt wird. Sie wird mit jeder Abhängigkeit fortgeschrieben,
nicht danach (`AGENTS.md` §9).

---

## 1 Magic Set Editor 2 — GPL-2.0, und was daraus folgt

[Magic Set Editor 2](https://github.com/twanvl/MagicSetEditor2) von Twan van
Laarhoven steht unter der **GNU General Public License, Version 2**.

**Aus MSE2 darf übernommen werden, seit dem 2026-08-16.** Dieses Projekt
steht selbst unter GPL-2.0-or-later
([0019](docs/entscheidungen/0019-gpl-2-0-statt-mit.md)); damit ist die
Bearbeitung erlaubt, die bis dahin ausgeschlossen war.

**Was übernommen wird, wird benannt.** Die GPL verlangt, dass eine Bearbeitung
als solche kenntlich ist (§2 a). Wo eine Regel aus MSE2 stammt, steht das im
Quelltext an der Stelle, mit Datei und Fassung — und ein späterer Leser kann
unterscheiden, was aus einem Datenpaket erschlossen und was aus dem fremden
Quelltext übernommen wurde.

**Womit gearbeitet wird, ist erlaubt:**

| Quelle | Warum unbedenklich |
| --- | --- |
| Die Datenpakete (`.mse-game`, `.mse-style`, …) | Struktur und Syntax einer Datei sind kein schutzfähiges Werk. Sie werden gelesen wie jede fremde Datei, die ein Programm öffnen können soll |
| Das Verhalten des laufenden Programms | Beobachtung ist keine Vervielfältigung |
| Öffentliche Dokumentation, Forenbeiträge, das MSE-Wiki | Als Beschreibung zitierfähig; übernommen wird die Erkenntnis, nicht der Wortlaut |

Jede Formaterkenntnis in `docs/FORMATE.md` nennt ihre **Quelle** — welche
Paketdatei, welche Stelle. Eine Erkenntnis ohne Quelle ist von einer
abgeschriebenen nicht zu unterscheiden, und genau das soll sie sein.

---

## 1a magic-set-generator — MIT, und was daraus folgt

[`fenhl/magic-set-generator`](https://github.com/fenhl/magic-set-generator) von
Max Dominik Weber („Fenhl") steht unter der **MIT-Lizenz**,
Copyright (c) 2017.

**Aufgenommen am 2026-08-16.** Es ist ein Rust-Werkzeug, das aus Kartennamen
`.mse-set`-Dateien **schreibt** — und damit die einzige belegte Auskunft
darüber, was in einem echten Set steht. Ein echtes `.mse-set` liegt diesem
Projekt nicht vor; alle Belegsets sind selbst geschrieben.

**Was daraus übernommen wird, wird benannt** — dieselbe Regel wie bei MSE2,
auch wenn die MIT-Lizenz von magic-set-generator sie nicht erzwingt. Sie steht
hier für den Leser: wer eine Regel im Quelltext findet, muss unterscheiden
können, ob sie erschlossen oder abgelesen ist.

**Der erste Ertrag**, und er war sofort ein behobener Fehler: `stylesheet:` in
einem `.mse-set` trägt die **Kurzform**, das zugehörige Paket heißt
`<spiel>-<stylesheet>.mse-style` (`src/mse.rs`). Bis dahin suchte dieses
Programm den rohen Wert, und jedes echte Set hätte ohne seinen Stil geöffnet.

**Was die MIT-Lizenz von magic-set-generator verlangt:** Lizenztext und
Urhebervermerk gehen bei jeder Weitergabe mit. **Bis zum 2026-08-16 stand hier
nur die Behauptung, sie stünden hier** — der Zweitblick über Phase A hat
angestrichen, dass der Text fehlte. Er steht jetzt vollständig da:

> MIT License
>
> Copyright (c) 2017 Max Dominik Weber ("Fenhl")
>
> Permission is hereby granted, free of charge, to any person obtaining a copy
> of this software and associated documentation files (the "Software"), to deal
> in the Software without restriction, including without limitation the rights
> to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
> copies of the Software, and to permit persons to whom the Software is
> furnished to do so, subject to the following conditions:
>
> The above copyright notice and this permission notice shall be included in all
> copies or substantial portions of the Software.
>
> THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
> IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
> FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
> AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
> LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
> OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
> SOFTWARE.

**Gelesen wurde die Fassung `1b8fb7f` vom 2021-08-09.** Eine Herkunftsangabe
ohne Fassung zeigt bei der nächsten Änderung am fremden Repositorium auf etwas
anderes.

**Und was sie in eine Richtung erlaubt:** MIT-Code darf in dieses
GPL-2.0-Projekt übernommen werden. **Umgekehrt nicht** — was hier entsteht,
lässt sich nicht ohne Weiteres dorthin zurückgeben.

**Nicht mitgeliefert wird davon nichts:** es wird gelesen, nicht gebündelt. Der
Quelltext wird bei Bedarf geklont und liegt außerhalb dieses Repositoriums.

---

## 2 Wizards of the Coast

*Magic: The Gathering* ist eine Marke der Wizards of the Coast LLC.

**Dieses Projekt ist mit Wizards of the Coast weder verbunden noch von ihnen
unterstützt, gebilligt oder beauftragt.**

Nicht Bestandteil dieses Repositoriums und nicht von der GPL-2.0 gedeckt:

- Kartenrahmen, Hintergründe, Rahmenverzierungen
- Mana-Symbole, Set-Symbole, Seltenheitsmarken
- Schriftarten der gedruckten Karten (Beleren, Matrix, Mplantin und andere)
- Kartennamen, Regeltexte, Schlüsselwörter, Illustrationen
- Aufmachung und Erscheinungsbild der gedruckten Karten

Wer solche Inhalte in dieser Anwendung verwendet, tut das in eigener
Verantwortung. Die Anwendung stellt sie nicht bereit; sie liest, was auf dem
Rechner der Nutzenden liegt.

Die [Fan Content Policy](https://company.wizards.com/en/legal/fancontentpolicy)
von Wizards of the Coast beschreibt, was für nichtkommerzielle Fanprojekte
zulässig ist. Sie ist keine Lizenz und ersetzt keine Prüfung im Einzelfall.

---

## 3 Vorlagensammlungen

**Keine davon wird eingecheckt und keine mitgeliefert.** Bezogen wird auf den
Rechner der Nutzenden, von der jeweiligen Quelle, nicht über uns — der
Paketverwalter aus Phase 7 tut das seit #139 über einen **Stand** (Freigabe
oder Commit), zeigt vorher, wessen Inhalte das sind und dass deren Lizenz nicht
unsere ist, und **packt keine ausführbaren Dateien aus**. Was er übergeht,
nennt er beim Namen. Die Prüfsumme des Archivs steht am Ergebnis, damit
nachweisbar bleibt, was bezogen wurde.

### Die Lizenzlage aller vier, nachgemessen

Seit #263
bietet die Paketseite vier Quellen zum Anklicken an. Am **2026-08-19** über
`gh api users/MagicSetEditorPacks/repos` nachgesehen:

| Repositorium | Lizenz | Größe (GitHub) |
| --- | --- | ---: |
| `Basic-M15-Magic-Pack` | **keine** | 316 MB |
| `Full-Non-Magic-Pack` | **keine** | 575 MB |
| `M15-Magic-Pack` | **keine** | 855 MB |
| `Full-Magic-Pack` | **keine** | 1.734 MB |

**Zehn Repositorien führt das Konto, keines trägt eine Lizenz.** Ohne Lizenz
gilt das Urheberrecht in seiner Vollform: keine Weiterverbreitung, keine
Bearbeitung. Deshalb wird hier nichts übernommen, nichts bearbeitet und nichts
weitergegeben — es wird **verwiesen**.

`internal/werkzeuge/bezugsquellen_test.go` hält die Liste in
`internal/bezug/quellen.go` gegen diese Datei: eine Quelle, die die Anwendung
anbietet und hier fehlt, macht die Pipeline rot. **Was er nicht prüft:** ob die
Angabe „keine Lizenz" heute noch stimmt. Das weiß nur, wer nachsieht.

#### Gefragt wird trotzdem

**Eine fehlende Lizenz lässt sich erfragen, statt sie hinzunehmen**
(#264).

| | |
| --- | --- |
| **Stand am 2026-08-20** | **noch nicht gestellt** |
| Der Text | liegt fertig an #264 |
| Wer ihn stellt | **ein Mensch, im eigenen Namen** — das ist keine Aufgabe für ein Werkzeug |
| Frist | **acht Wochen** ab dem Tag der Anfrage |

**Bleibt die Antwort nach der Frist aus, wird das hier vermerkt** — mit dem
Datum der Anfrage und dem des Ablaufs. Die Lage heißt dann nicht mehr
„ungeklärt", sondern **geklärt und ungünstig**, und an den Regeln oben ändert
sich nichts: es wird verwiesen und nicht übernommen.

**Warum eine Frist überhaupt dasteht:** ohne sie bliebe „wir haben gefragt" ein
Zustand ohne Ende, und niemand könnte sagen, ob die Lage geprüft oder nur
angestoßen wurde.

### Basic-M15-Magic-Pack

[MagicSetEditorPacks/Basic-M15-Magic-Pack](https://github.com/MagicSetEditorPacks/Basic-M15-Magic-Pack)
— „normal M15 frames, recommended for slim builds". **Keine Lizenz**, siehe
oben. Die kleinste der vier Quellen und die erste in der Liste: wer zum ersten
Mal ein Paket bezieht, kommt damit am schnellsten zu einer gezeichneten Karte.

### M15-Magic-Pack

[MagicSetEditorPacks/M15-Magic-Pack](https://github.com/MagicSetEditorPacks/M15-Magic-Pack)
— „public M15 templates, recommended for new users", rund 324 Vorlagenpakete.
**Keine Lizenz**, siehe oben.

Das README nennt CardConjurer als Quelle einiger Vorlagenbilder; die Rahmen
zeigen geschützte Gestaltung von Wizards of the Coast (Abschnitt 2). Die
Schriften unter `Magic - Fonts` haben je eigene Lizenzen, die eine Weitergabe
meist ausschließen. **Seit 0.6.0 gehen sie trotzdem mit** — im Bündel unter
`Contents/Resources/vorlagen/Magic - Fonts/`, auf Anweisung und Verantwortung
des Auftraggebers ([0169](docs/entscheidungen/0169-die-schriften-gehen-mit.md)):
ohne sie zeichnete eine frisch installierte Anwendung Karten ohne Text. Wer das
Pack bezieht, bekommt sie außerdem von der Quelle selbst.

### Full-Magic-Pack

[MagicSetEditorPacks/Full-Magic-Pack](https://github.com/MagicSetEditorPacks/Full-Magic-Pack)
sammelt rund 526 Vorlagenpakete. **Keine Lizenz**, siehe oben. Dazu:

1. **Geschützte Inhalte.** Kartenrahmen, Symbole und Schriften nach
   Abschnitt 2.
2. **Fremde Binärdateien.** Es liefert `mse.exe`, `magicseteditor.exe` und
   einen Deinstallierer mit — ausführbarer Code unbekannter Herkunft. Der
   Paketverwalter packt sie **nicht** aus und nennt sie beim Namen.

Es ist die Quelle, gegen die der Kanonlauf dieses Projekts misst.

### Full-Non-Magic-Pack

[MagicSetEditorPacks/Full-Non-Magic-Pack](https://github.com/MagicSetEditorPacks/Full-Non-Magic-Pack)
— Vorlagen für andere Spiele als Magic. **Keine Lizenz**, siehe oben. Die
Rechtslage der einzelnen Vorlagen ist damit ebenso ungeklärt; wer sie benutzt,
tut das in eigener Verantwortung.

### Sechs weitere, die **nicht** angeboten werden

Und warum nicht — denn eine Ablehnung braucht Deckung (§10):

| Repositorium | Warum nicht |
| --- | --- |
| `Beta-Update-Pack` | „may not be ready for full packs yet", sagt es selbst |
| `Installer-Pack` | Installationsdateien; `bezug` packt Ausführbares nicht aus — der Bezug ergäbe ein leeres Verzeichnis |
| `Outdated-Templates` | „kept on hand to help transition very old files" |
| `Magic-Mini-Packs` | Einzelstücke für DownGit, keine Sammlung |
| `TypesLocalizer` | ein Skript, kein Paket |
| `Font-Pack` | wird nicht gebraucht: die vier Packs bringen ihre Schriften unter `Magic - Fonts` selbst mit, und genau die holt der Bezug mit. Ein zweiter Bezug ergäbe eine zweite Fassung derselben Schrift im selben Baum |

### MTG JSON

[mtgjson.com](https://mtgjson.com) gibt Kartendaten als JSON aus — Namen,
Regeltexte, Typzeilen, Seltenheiten. Das sind Inhalte nach Abschnitt 2: sie
gehören Wizards of the Coast, nicht MTGJSON und nicht uns.

**Nicht eingecheckt und nicht mitgeliefert.** Die Anwendung *liest* solche
Dateien, wenn jemand sie ihr gibt. Im Repository liegen unter
`testdaten/mtgjson/` ausschließlich **selbst geschriebene** Beispieldateien mit
erfundenen Karten — dieselbe Syntax, eigener Inhalt, wie bei den MSE-Testpaketen
(`testdaten/LIESMICH.md`).

Der Gegenlauf gegen eine **bezogene** Datei steht hinter dem Build-Tag
`fremdpakete` und wird von Hand angestoßen (`internal/mtgjson/fremdpakete_test.go`).

### magicthegathering.io

[api.magicthegathering.io](https://api.magicthegathering.io) gibt Kartendaten
aus. Wie bei MTG JSON und Scryfall gilt: das sind Inhalte nach Abschnitt 2 —
sie gehören Wizards of the Coast.

**Nachgelesen am 2026-09-02** unter
[docs.magicthegathering.io](https://docs.magicthegathering.io/):

| Was | Wortlaut, verkürzt |
| --- | --- |
| Zugang | **keiner** — „No API key is required" |
| Grenze | die Seite sagt „5000 requests per hour". **Gemessen am selben Tag sagt der Dienst `ratelimit-limit: 1000`** — fünffach weniger. Dieses Programm drosselt auf **5 s** (720/h, 28 % Abstand zur gemessenen Zahl) |
| Über der Grenze | **403** mit „Rate Limit Exceeded" — **nicht** 429. Derselbe Code heißt sonst „du darfst nicht"; unterschieden wird am Rumpf |
| Kopfzeilen der Antwort | `Ratelimit-Limit`, `Ratelimit-Remaining` |
| Blätterung | `page`, `pageSize`; „default and maximum: 100" |
| Nennung | **Die Seite sagt dazu nichts.** Keine Lizenz, keine Nutzungsbedingung, keine Nennungspflicht — nachgelesen und nicht gefunden. Genannt wird der Dienst hier trotzdem, weil eine Herkunft an dem hängt, was sie geliefert hat (§16, Regel 6) |

**Der eine Dienst, der sich hier gegen sein Original prüfen lässt.** Ein Lauf
hinter dem Build-Tag `netzlauf` fragt ihn wirklich — und er hat beim ersten
Mal die Grenze angestrichen. Für die beiden folgenden gibt es das nicht.

### MTGJSON GraphQL

[graphql.mtgjson.com](https://graphql.mtgjson.com/) gibt Kartendaten,
Setdaten und Preise aus — dieselbe Quelle wie MTG JSON, über eine andere Tür.

**Nachgelesen am 2026-09-02** unter
[mtgjson.com/mtggraphql](https://mtgjson.com/mtggraphql/) und
[mtgjson.com/license](https://mtgjson.com/license/):

| Was | Wortlaut, verkürzt |
| --- | --- |
| Lizenz der Daten | **MIT**, „Copyright © 2018 – Present, Zach Halpern". Sie deckt die **Sammlung**, nicht den Inhalt: wem die Karten gehören, sagt Abschnitt 2 |
| Nennung | Die Lizenz von MTGJSON verlangt, dass Lizenztext und Urhebervermerk bei jeder Weitergabe mitgehen — deshalb steht der Vermerk hier |
| Zugang | `authorization: Bearer <Access Token>`, **nur für Patreon-Unterstützende**. „If you do not have an access token and have subscribed to our Patreon, contact us on Discord" |
| Stand | **Beta** — „The beta rollout of the service will be available to all Patreon subscribers" |
| Grenzen | „capped at `1,000` requests per IP Address per hour and `500` requests per access token per hour". Dieses Programm drosselt auf **10 s** — 360/h, 28 % unter der **engeren** der beiden |
| Ohne Marke | **HTTP 200** mit `data: null` und `errors[0]` „Access denied!" — gemessen am 2026-09-02. Wer nur den Statuscode liest, sieht einen Erfolg |
| Der Code dazu | `INTERNAL_SERVER_ERROR` — irreführend. Dieses Programm deutet ihn als Zugangsfehler und reicht ihn nicht durch |

**Nie gegen das Original gelaufen, aber nicht blind.** Das **Schema** ist ohne
Marke lesbar (gemessen am 2026-09-02): die Feldnamen sind daran geprüft. Die
**Antwort** ist es nicht.

### Scrydex

[api.scrydex.com](https://api.scrydex.com) gibt Kartendaten und Preise aus —
für sechs Spiele; hier zählt eines.

**Nachgelesen am 2026-09-02** unter
[scrydex.com/docs](https://scrydex.com/docs) und
[scrydex.com/terms](https://scrydex.com/terms):

| Was | Wortlaut, verkürzt |
| --- | --- |
| Eigentum an den Daten | „Any third-party card data, metadata, trademarks, or related content accessible through the Services remains the property of its respective owners or licensors. Scrydex does not claim ownership of third-party intellectual property." Damit gilt Abschnitt 2 unverändert |
| Nennung | **Die Bedingungen verlangen keine.** Was sie verlangen, ist, den **Dienst** nicht weiterzuverkaufen — das tut dieses Programm nicht |
| Zugang | **zwei** Kopfzeilen: `X-Api-Key` **und** `X-Team-ID`. Konto, Abonnement und ein Team sind nötig — „even solo users need a team" |
| Grenze | „a rate limit of 100 requests per second"; darüber **429**. Daneben **Kontingente**, die Geld kosten: „Each request consumes a specific number of credits" |
| Ohne Kopfzeilen | die Doku sagt „Requests made without authentication headers will proceed, but with heavily reduced rate limits". **Gemessen am selben Tag: HTTP 401 mit `INVALID_CREDENTIALS`.** Es geht nicht weiter — deshalb fragt dieses Programm ohne Zugang gar nicht erst |
| Blätterung | `page`, `page_size`; „The highest allowable value is `100`" |

**Warum hier auf eine Sekunde gedrosselt wird und nicht auf ein
Hundertstel.** Die technische Grenze erlaubt hundertmal mehr; was sie nicht
erlaubt, ist die Rechnung am Monatsende. Ein Programm, das im Namen der
Nutzenden Geld ausgibt, drosselt an der Grenze, die weh tut.

### Zwei von drei sagten etwas anderes als ihre Dokumentation

Am 2026-09-02 nachgemessen, und beide Male wich der Dienst von seiner Seite ab:

| Dienst | Die Seite sagt | Der Dienst sagt |
| --- | --- | --- |
| magicthegathering.io | 5.000 Anfragen je Stunde | `ratelimit-limit: 1000` |
| Scrydex | ohne Kopfzeilen geht es weiter | `HTTP 401`, `INVALID_CREDENTIALS` |

**Das ist kein Vorwurf an die Anbieter.** Es ist der Grund, warum §16 Regel 1
„nachgelesen, nicht erinnert" sagt — und warum das **Nachmessen** danach noch
einmal etwas anderes ist als das Nachlesen.

### Scryfall

[scryfall.com](https://scryfall.com) gibt Kartendaten, Regeltexte, Rulings und
Bilder aus. Wie bei MTG JSON gilt: das sind Inhalte nach Abschnitt 2 — sie
gehören Wizards of the Coast, nicht Scryfall und nicht uns.

**Scryfall arbeitet unter der Fan-Content-Policy von Wizards of the Coast.**
Wer die Daten benutzt, arbeitet unter denselben Bedingungen. Nachgelesen am
2026-08-23 unter [scryfall.com/docs/api](https://scryfall.com/docs/api):

| Was | Wortlaut, verkürzt |
| --- | --- |
| Kopfzeilen | „All HTTP requests to api.scryfall.com **must** include a User-Agent header"; „The Accept header **must** be present" |
| Grenzen | `/cards/search`, `/cards/named`, `/cards/random`, `/cards/collection`: **2/s**; `/cards/manifest`: 10/min; alles Übrige: 10/s |
| Zwei Drosseln | Dieses Programm hält **600 ms** für die Kartenwege (erlaubt: 500) und **125 ms** für alles Übrige (erlaubt: 100). Beide unter der Grenze, §16 Regel 4 — und **zwei**, weil eine gemeinsame die Rulings, Setangaben und die Symbologie auf ein Fünftel bremste, ohne dass jemand etwas davon hätte |
| Über Prozesse hinweg geteilt | **Alle** Zugänge zu Scryfall nehmen ihre Drossel aus **einem** Werk (`internal/app/scryfalldrossel.go`). Das Budget gilt für Host und IP, nicht je Instanz; vier Klienten mit eigener Drossel hielten es nicht ein. **Und seit Phase 3.2 auch über Prozessgrenzen**: Fenster, `mse-mcp` und `mse-rest` sprechen sich über eine Nachbardatei beim Zwischenspeicher ab (#1163). Bis dahin stand hier, wer alle drei gleichzeitig fahre, habe dreimal das Budget |
| Ein 429 sperrt wirklich | `Retry-After` wird in einen **Sperrzustand** übersetzt, nicht nur in eine Meldung. Ohne die Kopfzeile gelten 30 Sekunden. Bis dahin geht **aus keinem Prozess dieser Anwendung** eine Anfrage an Scryfall hinaus — nicht aus einem anderen Zugang, nicht aus der anderen Endpunktklasse und seit Phase 3.2 auch nicht aus einem anderen Prozess (#1163). Die Sperre steht in derselben Nachbardatei wie der Abstand. Die beiden Abstände bleiben getrennt, die **Sperre** gilt für das ganze Werk: ein 429 ist eine Aussage der Gegenstelle über sich, nicht über einen Endpunkt. **Bis zum 2026-08-24 stand hier dieselbe Zusage, und der Code hielt sie nicht** — gesperrt wurde nur die Drossel des gescheiterten Aufrufs (#1037) |
| 429 | sperrt **30 Sekunden**. „It is not acceptable to ignore HTTP 429 responses" — wer weitermacht, riskiert eine dauerhafte Sperre |
| Zwischenspeichern | „at least for 24 hours" |
| Sammeldaten | „Bulk data is only collected once every 12-24 hours." Sie sind damit **kein Ersatz** für den Abruf, sondern der Rückfall ohne Netz — und ein Befund sagt es, wenn eine Antwort von dort kommt. **Geladen wird sie einmal**, auch wenn mehrere Prozesse sie zugleich wollen: einer holt, die anderen warten und finden sie danach vor (#1163) |
| Bilder | nicht beschneiden, verzerren, filtern, nicht mit eigenen Marken versehen. Beim `art_crop` **müssen** Künstlername und Copyright in derselben Oberfläche stehen |

**Was die Absprache über Prozessgrenzen nicht kann**, gehört dazu:

- **Ein anderes Programm auf demselben Rechner hält sich nicht daran.** Sie ist
  beratend; dieselbe Grenze gilt trotzdem für die IP.
- **Sie zählt keine Anfragen.** Sie hält einen **Abstand** und eine **Sperre**;
  wie viele in einer Stunde hinausgehen, sagt sie nicht.
- **Nach einer Minute gibt sie auf** und lässt die Anfrage durch. Ein Halter
  gibt sie nach höchstens einem Abstand zurück — bei 100 ms stehen damit
  sechshundert Wartende in der Schlange, bevor jemand aufgibt. **Eine Sperre
  aus einem 429 wird ohne die Nachbardatei abgewartet**: sie kann Minuten
  dauern, und wer sie mit der Absprache in der Hand abwartete, sperrte alle
  anderen für diese Zeit aus.

**Der Wortlaut der Bildbedingung**, weil an ihr eine Wache hängt:

> When using the art_crop, list the artist name and copyright elsewhere in the
> same interface presenting the art crop, or use the full card image elsewhere
> in the same interface. Users should be able to identify the artist and source
> of the image somehow.

**Was dieses Programm daraus macht**: `internal/netz` hält Fristen, Kopfzeilen
und die Drossel, `internal/scryfall` das fremde DTO und die Zuordnung.

Für Bilder heißt es fünferlei, und keines davon ist ein Hinweis:

1. **Ein `art_crop` ohne Künstlernamen wird abgewiesen** — ohne ihn lässt sich
   die Bedingung nicht einhalten, und ein Bild auszuliefern, dessen Bedingung
   man nicht erfüllen kann, ist keine Auslegungsfrage
   (`internal/scryfall/bild.go`).
2. **Die Nennung reist mit dem Bild** und wird nicht daneben gebaut. Wer das
   Bild hat, hat damit auch, was neben ihm stehen muss — im Fenster, in der
   MCP-Antwort, in der REST-Antwort und auf der Kommandozeile.
3. **Sie bleibt an der Karte, wenn das Bild längst abgelegt ist.** Die Herkunft
   des **aktuellen** Werts steht in `wert_herkunft` (Migration 0014), gepflegt
   in **derselben Transaktion** wie der Wert; die Kartenvorschau baut daraus
   Künstler, Copyright und Quelle und zeigt sie unter dem Bild.

   **Hier stand bis zum 2026-08-24 etwas anderes**, und es war zu kurz
   gesprungen: die Nennung hing am **offenen Zuschnittkasten**. Wurde er
   geschlossen, stand das Bild auf der Karte und die Nennung nirgends. Eine
   Bedingung, die das Schließen eines Kastens nicht überlebt, ist keine
   (#1025).
4. **Der Künstler geht zusätzlich als `illustrator` in die Karte**, damit ihn
   auch die `.mse-set`-Ausfuhr trägt.

   **Das ist die Kür, nicht die Pflicht**, und der Unterschied gehört
   ausgesprochen: das Schreiben geschieht **nach** dem Bild und in einer
   eigenen Transaktion (§7, Regel 3 — „geschrieben ist geschrieben"). Scheitert
   es, oder führt die Vorlage gar kein Feld `illustrator`, dann steht der Name
   nicht auf der Karte — **die Nennung aus Punkt 3 steht trotzdem**, weil sie
   an der Herkunft hängt und nicht an der Vorlage. Ohne bekannten Künstler
   nennt sie den Rechteinhaber allein.
5. **Beide Zeilen tragen ihre Herkunft im Verlauf** — die Kartenseite bei
   Scryfall. Wer den Verlauf liest, unterscheidet eine Übernahme von einer
   Eingabe, ohne den Wert zu erraten.

### Die zweite Auslegung, und warum sie hier nicht genommen wird

Die Bedingung nennt **zwei** Wege: Künstler und Copyright in derselben
Oberfläche — **oder** „use the full card image elsewhere in the same
interface".

**Der zweite wird nicht in Anspruch genommen.** Es läge nahe: die
Kartenvorschau zeigt eine ganze Karte. Aber sie zeigt **unsere** Karte — eine
Rasterung aus Vorlage und Feldwerten, mit dem Ausschnitt darin. Sie ist nicht
„the full card image", das Scryfall meint; auf ihr steht kein fremder
Copyright-Vermerk, und wer sie ansieht, erfährt über Künstler und Quelle
nichts.

**Damit bliebe der zweite Weg eine Behauptung**, und eine Ausrede sieht wie
eine Begründung aus (AGENTS.md §10). Genommen wird deshalb der erste, und zwar
so, dass er nicht am Zustand einer Oberfläche hängt.

**Die Bytes gehen unverändert durch.** Nicht verzerrt, nicht geschärft, nicht
entsättigt, ohne eigene Marke — das verlangen die vier Zeilen über der
`art_crop`-Bedingung. Zugeschnitten wird auf die Feldfläche, wie bei jedem
anderen Bild auch; das ist der Zuschnitt, den MSE selbst rechnet.

**Nicht eingecheckt und nicht mitgeliefert.** Unter
`internal/scryfall/testdaten/` liegen ausschließlich **selbst geschriebene**
Beispielantworten mit erfundenen Karten — dieselbe Gestalt, eigener Inhalt, wie
bei MTG JSON und den MSE-Testpaketen.

### Saat

Unter `saat/` liegen eigene, minimale Vorlagen — genug, um die Anwendung ohne
fremdes Material zu benutzen und zu prüfen. Sie enthalten keine geschützten
Grafiken. Sie sind **Quelltext dieses Repositoriums** — Textdateien im
MSE-Paketformat und selbst gezeichnete SVG — und damit von der GPL-2.0-or-later
gedeckt wie alles andere hier.

Das ist kein Widerspruch zum Geltungsbereich in `LICENSE`: der nimmt Inhalte
**Dritter** aus, und an der Saat hat niemand sonst Rechte. Der Zweitblick über
die Portierung hat die frühere Formulierung („wie der übrige Inhalt") zu Recht
angestrichen — sie las sich, als stünde hier auch fremdes Material.

---

## 4 Bibliotheken

Die vollständige, maschinell erzeugte Stückliste entsteht beim Release
(`make sbom`) und liegt dem Artefakt bei. Diese Tabelle nennt, was die Form des
Projekts bestimmt:

| Bibliothek | Lizenz | Wofür |
| --- | --- | --- |
| Go-Standardbibliothek | BSD-3-Clause | Unterbau |
| [Wails 3](https://github.com/wailsapp/wails) (`v3.0.0-beta.30`) und [`@wailsio/runtime`](https://www.npmjs.com/package/@wailsio/runtime) in derselben Fassung | MIT | Fenster, Menüleiste, Dialoge, Zwischenablage und die Brücke zur Oberfläche — seit ENG-02 ([0124](docs/entscheidungen/0124-wails-3-als-gepinnte-vorabfassung.md), [0132](docs/entscheidungen/0132-wails-3-als-huelle.md)); das Bündel legt `scripts/bauen.sh`. **Gemessen am 2026-09-24** über `go list -deps`: das Fenster bindet **53** Module statt 56 — fort sind Wails 2 und fünf seiner Helfer, neu sind `adrg/xdg` und `mattn/go-isatty` (beide MIT), und `ProtonMail/go-crypto` steigt von 1.1.6 auf 1.3.0, weil Wails 3 es so verlangt. Die Kommandozeilenwerkzeuge binden Wails seitdem **gar nicht** mehr. `@wailsio/runtime` hat keine eigenen Abhängigkeiten (277 KB) |
| [modernc.org/sqlite](https://gitlab.com/cznic/sqlite) | BSD-3-Clause | SQLite ohne cgo |
| [golang.org/x/image](https://pkg.go.dev/golang.org/x/image) | BSD-3-Clause | Glyphen rastern ([0008](docs/entscheidungen/0008-schriftrasterung.md)) — **und seit #826 die Kodierer für BMP und TIFF**, die Go selbst nicht mitbringt |
| [dlclark/regexp2](https://github.com/dlclark/regexp2) | MIT | Muster mit Vorausblick, Rückblick und Rückverweis — 15 von 1645 in `M15-Magic-Pack` ([`msescript.Muster`](internal/msescript/muster.go)) |
| [ProtonMail/go-crypto](https://github.com/ProtonMail/go-crypto) | BSD-3-Clause | OpenPGP: die **Signatur über den Prüfsummen einer Freigabe** prüfen, bevor ein Abbild abgelegt wird ([0051](docs/entscheidungen/0051-die-signatur-wird-geprueft.md)). **Lag schon im Baum** — Wails bringt sie mit; sie steht seit dem 2026-08-26 nur ausdrücklich in `go.mod`, weil dieses Programm sie jetzt selbst benutzt |
| [golang.org/x/crypto](https://pkg.go.dev/golang.org/x/crypto) | BSD-3-Clause | **Argon2id** für Kennwörter und Wiederherstellungscodes ([`internal/kennwort`](internal/kennwort/), Werkstück 6 von Phase 4.12). **Lag schon im Baum** — sie steht seit dem 2026-09-05 nur ausdrücklich in `go.mod`, weil dieses Programm sie jetzt selbst benutzt |
| [tmc/apple](https://github.com/tmc/apple) | MIT | **Touch ID und Face ID** über `LocalAuthentication`, cgo-frei ([`internal/biometrie`](internal/biometrie/), Werkstück 7) und, **nur im Abnahmebau**, die Hüllenprobe in `internal/fenster` (#1727). **v0.6.18 und keine 1.0**, und darunter `purego v0.11.0-**alpha.6**` — gemessen am 2026-09-05; seit dem 2026-10-11 `purego` v0.11.1, eine Endfassung vor 1.0 (#2305). Zwei unfertige Schichten unter einem Sicherheitsmerkmal; deshalb steht jeder Ruf hinter einem Riegel, der einen `panic` in eine Antwort verwandelt, und deshalb ist **TOTP** der Weg, der immer geht |
| [ebitengine/purego](https://github.com/ebitengine/purego) | **Apache-2.0** | Der Unterbau von `tmc/apple`: Rufe in ein ObjC-Framework ohne cgo. **Seit dem 2026-10-07 auch direkt**, nur im Abnahmebau (`internal/fenster/abnahme_hand_darwin.go`) — dieselbe Fassung. **Seit dem 2026-10-15 direkt auch im ausgelieferten Bau**: [`internal/prozessspeicher`](internal/prozessspeicher/) misst den Speicher der WebKit-Prozesse über libproc (OPS02, #2348). **Die erste Apache-2.0-Abhängigkeit dieses Projekts** — und damit die erste, die den Absatz darunter wirklich braucht: mit GPL-**2.0** allein wäre sie nicht vereinbar, mit GPL-3.0 ist sie es, und `LICENSE` sagt „or any later version". Gelesen am 2026-09-05 in `LICENSE` des Moduls; **nicht** doppelt lizenziert |
| [tailscale/tailcat](https://github.com/tailscale/tailcat) | BSD-3-Clause | **Die Verbindung über fremde Netze** — Tailscales Datenebene **ohne** deren Kontrollebene: WireGuard zur Verschlüsselung, magicsock zur NAT-Durchquerung, DERP als Rendezvous und Rückfall, **ohne Konto** ([`internal/fernleitung`](internal/fernleitung/), N3 von Phase 5.4). **v0.7.0, gepinnt** (gehoben am 2026-10-07, K01 #2187: erst v0.7.0 nennt die Gegenstellen am Server) — es sagt ausdrücklich nichts zu: „the Go API, the CLI flags and output, and the wire format may all change" ([0114](docs/entscheidungen/0114-tailcat-gepinnt.md)). **Was es kostet, ist gemessen**: 525 neue Module im Graphen, davon **301 wirklich übersetzte** Fremdpakete, und rund 21 MB Binärzuwachs. Lizenzzeile „Copyright (c) 2020 Tailscale Inc & contributors" — gelesen am 2026-09-15 |
| [tailscale.com](https://github.com/tailscale/tailscale) | BSD-3-Clause | Der Unterbau von `tailcat`: WireGuard, magicsock, DERP. **Kommt an einer Stelle direkt vor**: `internal/fernleitung/proxy.go` setzt zwei Haken aus `tailscale.com/feature`, damit der Weg zum Relais den eingestellten Proxy nimmt (NET-02, #1706). Sonst ruft dieses Programm ihn nicht selbst |
| [gvisor.dev/gvisor](https://github.com/google/gvisor) | Apache-2.0 | Der Netzwerkstapel, den `tailscale.com` im Benutzerraum betreibt. **Kommt nicht direkt vor** |
| [FiloSottile/age](https://github.com/FiloSottile/age) | BSD-3-Clause | **Sicherungen mit Kennwort** ([`internal/sicherung`](internal/sicherung/), OPS-06, #1723). **v1.3.2 vom 2026-08-29**. Ein offenes Format: wer das Kennwort hat, öffnet die Sicherung mit `age -d` auch ohne diese Anwendung. Die Lizenzzeilen lauten „Copyright 2019 The age Authors", „Copyright 2019 Google LLC" und „Copyright 2022 Filippo Valsorda" — gelesen am 2026-09-26 in `LICENSE` des Moduls |
| [FiloSottile/hpke](https://github.com/FiloSottile/hpke) | BSD-3-Clause | Der Unterbau der Empfänger mit Schlüssel in `age`. **Kommt nicht direkt vor** — Sicherungen mit Kennwort benutzen scrypt und nicht HPKE |
| [hashicorp/mdns](https://github.com/hashicorp/mdns) | MIT | **mDNS** — wer im selben Netz da ist ([`internal/entdeckung`](internal/entdeckung/), N2 von Phase 5.4). **v1.0.7 vom 2026-06-11**, und die Wahl ist eine Frage der Pflege: `grandcat/zeroconf` ist bei v1.0.0 vom Januar 2020 stehengeblieben ([0113](docs/entscheidungen/0113-kein-libp2p-im-lan.md)). Die Lizenzzeile lautet „Copyright IBM Corp. 2014, 2026“ — gelesen am 2026-09-14 in `LICENSE` des Moduls |
| [miekg/dns](https://github.com/miekg/dns) | BSD-3-Clause | Der Unterbau von `hashicorp/mdns`: DNS-Nachrichten lesen und schreiben. **Kommt nicht direkt vor** — dieses Programm ruft ihn nicht selbst |
| React | MIT | Oberfläche |
| Vite | MIT | Bau des Frontends |
| Tailwind CSS | MIT | Gestaltung |
| PixiJS | MIT | der gezeichnete Spieltisch (#1460) |
| Mantine (`@mantine/core`, `@mantine/hooks`) | MIT | die Bauteilbibliothek (#1679, [0125](docs/entscheidungen/0125-mantine-als-bauteilbibliothek.md)). **Gemessen am 2026-09-20**: sie bringt fünf eigene Abhängigkeiten mit — `@floating-ui/react`, `clsx`, `react-number-format`, `react-remove-scroll`, `type-fest` —, zusammen 16 neue Pakete. Alle MIT |
| Vitest, Testing Library | MIT | Frontend-Tests |
| axe-core | MPL-2.0 | Prüfung auf Barrierefreiheit |

Alle mit GPL-2.0-or-later vereinbar. **Die Richtung der Prüfung hat sich am
2026-08-16 gedreht**: bis dahin war GPL die verbotene Lizenz, jetzt ist sie die
eigene. Aufzupassen ist seither auf Apache 2.0 — mit GPL-**2.0** ist sie nicht
vereinbar, wohl aber mit GPL-3.0, und weil `LICENSE` „or any later version"
sagt, löst es sich dort. Wer eine Abhängigkeit aufnimmt, prüft ihre Lizenz
vorher und trägt sie hier ein.

> **Seit dem 2026-09-05 ist das kein Vorbehalt mehr, sondern der Fall.**
> `purego` steht unter Apache-2.0, und damit ist die **ausgelieferte**
> Zusammenstellung GPL-**3.0**-or-later — nicht mehr wahlweise 2.0. Der Code
> dieses Projekts bleibt, was er ist; was sich ändert, ist die Fassung, unter
> der das Ganze weitergegeben werden darf.
>
> **Wer das nicht will, baut ohne Biometrie**: sie liegt hinter einem Bautag,
> und `TOTP` ist der Weg, der immer geht
> ([0066](docs/entscheidungen/0066-nutzer-profile-und-die-vier-entscheidungen.md)).
> Ohne `darwin` kommt weder `tmc/apple` noch `purego` in den Bau.

MPL-2.0 (axe-core) ist dateibezogen ansteckend und damit unbedenklich, solange
die Bibliothek unverändert benutzt wird. Sie ist ohnehin nur
Entwicklungsabhängigkeit und wird nicht ausgeliefert.

---

### Ein Werkzeug des Systems: Apples Gerätemodell

**`/usr/bin/fm` gehört zu macOS 27 und wird nicht mitgeliefert.** Dieses
Programm startet `fm serve` über einen Unix-Socket, wenn eine kleine
Textaufgabe es braucht (AI-05,
#2042).
Die Bedingungen zeigt `sudo fm license`; sie binden an die Softwarelizenz von
macOS. Der Auftraggeber hat sie am 2026-10-03 gelesen: **nichtkommerzielle
Nutzung ist erlaubt, die Schnittstelle darf eingebunden werden** — dieses
Programm steht unter GPL und wird nicht verkauft. Die Fassung meldet
`fm license --status`: *FM1 version 1.0*. **Neu zu entscheiden** ist das, wenn
das Programm verkauft werden soll oder eine andere Fassung gemeldet wird
([0166](docs/entscheidungen/0166-was-fm-serve-belegt.md)). **Die Lizenz nimmt
dieses Programm nie an** — das tut ein Mensch, für alle Nutzer der Maschine.

## 5 Schriften

**Mit 0.6.0 gehen die Schriften des `M15-Magic-Pack` mit** — `Magic - Fonts/`,
26 MB, unter `Contents/Resources/vorlagen/`. Bis dahin stand hier: „Ausgeliefert
wird **keine** Schrift, die nicht frei lizenziert ist." Das gilt nicht mehr, und
der Grund dafür ist nicht, dass sich die Lizenzen geändert hätten: Beleren
gehört Wizards of the Coast, MPlantin geht auf Monotypes Plantin zurück, und die
meisten Lizenzen schließen eine Weitergabe aus. **Der Auftraggeber hat sie
trotzdem mitgeben lassen, auf seine Verantwortung**
([0169](docs/entscheidungen/0169-die-schriften-gehen-mit.md)); ohne sie zeichnete
eine frisch installierte Anwendung Rahmen und Bild, aber keinen Text. **Diese
Lizenz** (GPL-2.0-or-later) gilt für sie nicht.

Die Anwendung benutzt die Systemschriften von macOS für ihre eigene Oberfläche
und liest für die Kartenausgabe, was das jeweilige Paket vorgibt — zuerst von
dort, wo die Nutzenden Schriften installiert haben, dann aus den Paketen.

Fehlt eine Schrift, wird das **benannt** und nicht stillschweigend ersetzt: eine
Karte, die mit einer falschen Schrift gerendert wird, sieht richtig aus und ist
es nicht.
