# Netzeigenschaften — wie dieses Programm ins Netz geht und wie es erreichbar ist

**Jede Eigenschaft hier ist gemessen, nicht erinnert** (AGENTS.md §10). Wo
eine Zahl steht, steht daneben, worüber sie erhoben wurde.

Entstanden mit Werkstück A18 von Phase 5.3
(#1533).

---

## Als Klient

**Es gibt genau drei Stellen, die einen HTTP-Klienten bauen** — die dritte
seit AI-05 (#2042),
und sie geht nicht ins Netz (unten). Ein Wächter hält die Aufzählung
(`internal/werkzeuge/httpklienten_test.go`) — eine vierte erzwingt die
Entscheidung, statt still einen Transport mit anderen Eigenschaften
aufzumachen.

> **Seit N5 (#1571)
> haben beide denselben Proxyweg.** Der Wächter hielt die Aufzählung und
> **nicht**, dass beide dasselbe tun — das steht so in seinem eigenen Kopf,
> und das Issue hat es angestrichen. Was gleich geworden ist, steht in der
> Tabelle; was verschieden bleibt, steht darunter mit Grund.

| | `internal/netz` | `internal/anmeldung` |
| --- | --- | --- |
| **Wofür** | jeder fremde Dienst — Karten, Preise, KI, Pakete | der Anmeldefluss über einen fremden Anbieter |
| **Proxy** | die **Einstellung** `netz.proxy.*`, und wo sie leer ist `HTTP_PROXY`, `HTTPS_PROXY`, `NO_PROXY` über `httpproxy.FromEnvironment()` | **dieselbe Auswahl** seit N5 (`netz.Proxyauswahl`) |
| **Protokoll** | HTTP/2, wo die Gegenstelle es kann; sonst HTTP/1.1 | dito, ausdrücklich gesetzt seit N5 |
| **Fristen** | drei: Verbindung, Antwortkopf, **Stillstand des Rumpfes**; einstellbar über `netz.frist.*` | eine Gesamtfrist von 30 s, **nicht** einstellbar |
| **Drossel, Erlaubnis, 429** | ja | nein — der Fluss ist einmalig und hat keinen Anbieter, der drosselt |
| **Wo zu sehen** | *Netz und Erreichbarkeit* im Fenster, `mse netzweg`, `lesen.netzweg`, `GET /netz` | dito — es ist derselbe Weg |
| **Wo zu ändern** | *Einstellungen → Wie dieses Programm hinausgeht*, an jeder Fläche über `EinstellungSetzen` | dito |

### Der dritte: ein Unix-Socket zu Apples Gerätemodell

**`netz.NeuerKlientUeberSocket`** spricht mit `fm serve`, das dieses Programm
selbst startet — **über einen Unix-Socket und nie über einen Port**
([0166](entscheidungen/0166-was-fm-serve-belegt.md)). Gemessen am 2026-10-03:
mit einem Socketpfad über 104 Bytes legt `fm` **keinen** Socket an und
lauscht still auf **allen** Schnittstellen. Der Pfad wird deshalb geprüft,
und gefragt wird erst, wenn der Socket dasteht.

| | `netz.NeuerKlientUeberSocket` |
| --- | --- |
| **Wofür** | Apples Gerätemodell (`internal/applefm`) |
| **Proxy** | **keiner**, auch wenn die Umgebung oder `netz.proxy.*` einen nennt — ein Proxy schickte eine Anfrage an diesen Rechner hinaus |
| **Protokoll** | HTTP/1.1 über den Socket |
| **Fristen** | die üblichen (unten); dazu **10 s bis `/health`** beim Start, gemessen waren 0,15 s |
| **Adresse und Port** | `$TMPDIR/mse-fm/fm-<pid>.sock`, der Ordner mit `0700` — ein vorhandener mit anderen Rechten wird abgewiesen; **kein Port** |
| **Weiterleitung, Drossel, Zielsperre** | nein — ein Dienst am Socket hat niemanden, an den er weiterleiten dürfte, und keine Adresse, die eine Sperre prüfen könnte |
| **Wo zu sehen** | *Einstellungen → KI*, `mse ki-modelle`, `lesen.ki.modelle`, `GET /ki/modelle` |

### Und was das Fenster selbst holt

**Die Vorschaubilder der Kartensuche** lädt die WebView selbst, an
`internal/netz` vorbei: die Adresse nennt der Dienst, und die
Schutzrichtlinie des Fensters lässt genau drei Ursprünge zu —
`https://cards.scryfall.io`, `https://gatherer.wizards.com`,
`https://gatherer-static.wizards.com` (Vermerk
[0179](entscheidungen/0179-die-webview-laeuft-unter-einer-gemessenen-schutzrichtlinie.md),
gemessen am 2026-10-07). **Für diese Ladungen gilt keine Zeile oben**:
weder `netz.proxy.*` noch die Fristen, die Drossel oder die Zielsperre.
**Gemessen mit D06** (#2237,
2026-10-10, Drehbuch `d06` am nativen Fenster): mit `netz.proxy.http` auf
einem eigenen Proxy ging die Suche über ihn, **die Bilder nicht** — beide
Bildwirte wurden direkt erreicht, am Proxy kam kein `CONNECT` an. Die WebView
nimmt den Proxy dieser Anwendung also nicht; welchen des Systems sie nähme,
ist mit diesem Lauf nicht gemessen. Wer hinter einem Proxy ohne direkten
Ausgang sitzt, findet Karten und sieht ihre Vorschaubilder nicht.

**Und die Bilder der Handelszeilen** (P10,
#2378)
ebenso — **ohne Klick**: der Vorrat der Drucke (`druck/vorrat:`) hält Name,
Set, Nummer und die **Adresse** des Bilds, nicht das Bild. Sobald eine Zeile
sichtbar wird, lädt die WebView es von `https://cards.scryfall.io`; ohne Netz
bleibt der Platz leer, Name und Druck stehen trotzdem da. Nach **Daten** fragt
das Öffnen eines Handels nicht (§16, Regel 5). **Entschieden am 2026-10-20**
(#2531,
Gesamtreview über 0.9.0, Befund 5): so bleibt es, wie bei der Kartensuche —
weder lokal gehalten noch erst auf Wunsch geladen.

### Wohin ein Klient wählen darf

**Seit SEC-P4 (#2000)
prüft `internal/netz`, wohin es verbindet** — vorher wählte der Klient jede
Adresse, die man ihm nannte, und folgte jeder Weiterleitung nach Gos
Vorgabe. Befund P4 in [`PENTEST.md`](PENTEST.md).

| Klient | Bereich | Benannter Schalter | Wer das Ziel setzt |
| --- | --- | --- | --- |
| Dienste mit fester Adresse — Scryfall, MTGJSON, CardTrader, die drei Kartendienste, die Aktualisierung | **Grundsperre** | — | das Programm; `MSE_*_ADRESSE` lenkt für Läufe um |
| die KI — lokale Laufzeit, Einbettung und die drei Cloudanbieter | **Grundsperre**: Schleife und eigenes Netz bleiben erreichbar | — | ein **Mensch**: `ki.adresse`, `dienst.*.adresse` und `netz.proxy.*` setzt nur, wer `externSenden` darf, **ein Token nie** |
| Paketbezug (`internal/bezug`) | **nur öffentlich** — dazu gesperrt: Schleife und eigenes Netz | `MSE_NETZ_SCHLEIFE=an` öffnet die **Schleife** und nur sie | der Rufer von `POST /pakete` — auch ein Token `schreiben`; **seit G04** derselbe Klient für die Auskunft über Stände (`api.github.com/repos/…/commits/HEAD` und `…/compare/…`; nennt der Vergleich 300 Dateien, dazu `…/git/trees/{commit}?recursive=1` für beide Stände — gemessen am 2026-10-11 rund 10 MB je Baum bei `Full-Magic-Pack`), gerufen von `GET /paketbezuege?fragen=1` und der Vorschau |
| die Leitung **zum Proxy** | **Grundsperre** — ein Firmenproxy steht im eigenen Netz | — | Umgebung oder Einstellung (s. o.) |
| ein Gegenüber (`internal/leitung`) | **keine Bereichssperre** — das eigene Netz ist dort der Zweck, und die Verbindung ist auf die Kennung festgenagelt | — | die Einladung; **ihre Annahme meldet sich erst nach `externSenden`** |
| der Anmeldefluss (`internal/anmeldung`) | keine Zielprüfung — drei feste Adressen der Anbieter | — | das Programm |
| Apples Gerätemodell (`internal/applefm`) | **kein Netz** — ein Unix-Socket dieses Prozesses | — | das Programm |

**Die Bereiche**, in `internal/netz/ziel.go` und dort mit einer Tafel geprüft:

| | gesperrt |
| --- | --- |
| **Grundsperre** — kein Dienst braucht sie | `0.0.0.0/8` und `::` · `169.254.0.0/16` und `fe80::/10`, dort stehen Metadatendienste · Gruppenadressen · `240.0.0.0/4` samt Rundruf |
| **eigenes Netz** | `10/8`, `172.16/12`, `192.168/16`, `fc00::/7` · `100.64.0.0/10` (geteilt, auch Tailscale) · `192.0.0.0/24` · `198.18.0.0/15` · `fec0::/10` |
| **Schleife** | `127.0.0.0/8`, `::1` |

Eine IPv4-abgebildete IPv6-Adresse (`::ffff:127.0.0.1`) und eine NAT64-Adresse
(`64:ff9b::/96`) werden **als die eingebettete IPv4-Adresse** geprüft.

**Geprüft wird an drei Stellen**: vor der Anfrage das Schema (`http`,
`https`) und eine Adresse in Ziffern — die einzige Prüfung, die auch hinter
einem Proxy greift —; **nach der Auflösung** in `net.Dialer.Control`, bevor
ein Paket hinausgeht; und **bei jeder Weiterleitung** über `CheckRedirect`,
nach höchstens zehn wie unter Gos Vorgabe.

**Gemessen am 2026-10-02** (Go 1.27.1, darwin/arm64): eine Verbindung an
`0.0.0.0:<port>` und an `[::ffff:127.0.0.1]:<port>` erreicht einen Lauscher
auf `127.0.0.1:<port>`. Darum steht `0.0.0.0` in der Grundsperre und nicht
nur `127.0.0.0/8` in der Sperre der Schleife — und darum zielt der Lauf
`internal/flaechen/zielsperreki_test.go` gerade dorthin.

**Was die Prüfung nicht deckt**: einen **Namen**, den ein Proxy auflöst —
wohin er weiterreicht, entscheidet er —, und eine IPv4-Adresse in 6to4 oder
Teredo; sie erreicht ein fremdes Relais und nicht diesen Rechner.

**Was dafür nicht mehr geht**: ein Paket von einem Git-Server im eigenen Netz
zu beziehen. Ein Ollama auf einem zweiten Rechner (`http://gpu.intern:11434`)
geht weiter — dort ist die Sperre der **Riegel am Schreiben** und nicht eine
am Ziel.

### Warum die Frist des Anmeldeflusses eine andere bleibt

**Weil sie eine andere Frage beantwortet.** Die drei Fristen von
`internal/netz` trennen „die Gegenstelle antwortet gar nicht" von „der Bezug
ist groß und die Leitung langsam"; ein Anmeldefluss sind **zwei** Rufe, und
zwischen ihnen wartet ein Mensch vor einem Fenster. Drei Regler dafür wären
drei Entscheidungen über einen Fall, der eine braucht.

**Der Proxy ist es nicht** und ist deshalb weggezogen: dass eine Anfrage über
`proxy.firma:3128` geht, ist eine Eigenschaft der **Leitung** und keine des
Dienstes. Bis N5 war es anders, und der Preis war messbar: wer hinter einem
Firmenproxy saß, konnte Karten laden und sich nicht anmelden.

### Was die Einstellung kann und was nicht

| | wirkt ab |
| --- | --- |
| `netz.proxy.http`, `netz.proxy.https`, `netz.proxy.aus` | der **nächsten Anfrage** — `http.Transport` ruft seine Auswahl je Anfrage, **mit dem Kontext dieser Anfrage**: er trägt das Profil, dessen Einstellung gilt |
| `netz.frist.*` | dem **nächsten Klienten** — sie sitzen im Wähler und im Transport |

**Bis D06 galt die Einstellung nicht für jeden Klienten** (2026-10-10,
#2237):
Scryfall, die anderen Kartendienste und der Preisbezug fragten ihre
Proxyquelle ohne Kontext, der Speicher fand darin kein Profil, und ein
Lesefehler heißt hier „wie ohne Einstellung" — sie gingen direkt hinaus.
Gefunden hat es das Drehbuch `d06` am nativen Fenster; der Test dazu liest
über den echten Speicher (`internal/app/proxyprofil_intern_test.go`), weil die
Proben davor nach keinem Profil fragten. **Was weiter gilt**: eine Anfrage
ohne Profil im Kontext geht wie ohne Einstellung — die Umgebung entscheidet.

**Der Unterschied steht im Fenster** und wird nicht verschwiegen: eine Fläche,
die beides gleich aussehen ließe, verspräche mehr, als ihr Rumpf hält (§10).

### Warum nicht `http.ProxyFromEnvironment`

**Weil es die Umgebung beim ersten Ruf zwischenspeichert.** Gemessen am
2026-09-12:

```
nach dem ersten Setzen:  http://erster:3128
nach dem zweiten Setzen: http://erster:3128   ← die zweite Zuweisung wirkt nicht
```

Das ist zweimal teuer: ein Lauf, der die Variable setzt, prüfte einen anderen
Wert als den gesetzten und wäre **grün**, und eine Einstellung könnte die
Umgebung nie überschreiben. `httpproxy.FromEnvironment()` liest **je Klient**
neu.

### Was `NO_PROXY` leistet — und was ohnehin gilt

**Die Schleifenadresse bleibt von sich aus außen vor**, ohne jeden Eintrag
(gemessen). Die Sorge, ein Ollama auf `127.0.0.1` liefe ohne `NO_PROXY` über
den Firmenproxy, war unbegründet.

**Gebraucht wird es für ein Ollama auf einem *anderen* Rechner im eigenen
Netz** (`http://gpu.intern:11434`). Dort greift die Schleifenregel nicht, und
ohne `NO_PROXY` ginge die lokale Einbettung durch den Proxy.

### Die Fristen

| | Verbindung | Antwortkopf | Stillstand |
| --- | ---: | ---: | ---: |
| übliche (`netz.UeblicheFristen`) | 15 s | 30 s | 60 s |
| Ollama | 5 s | **5 min** | 60 s |
| Anthropic | 10 s | 2 min | 60 s |
| OpenAI, xAI | 10 s | **3 min** | 2 min |
| Apples Gerätemodell — die üblichen, warm gemessen 0,7 bis 1,3 s je Antwort | 15 s | 30 s | 60 s |

**Die Kopffrist ist die, auf die es bei langen Aufträgen ankommt**: sie misst
bis zur Statuszeile, und ein Bildmodell antwortet erst, wenn es fertig ist —
es schickt keinen Kopf und dann langsam den Rumpf. Wer sie zu knapp setzt,
bricht den Auftrag ab, **während der Anbieter ihn noch abrechnet**.

**Gemessen mit A17** (#1532):
eine Gegenstelle, die schweigt, wird von der Frist beendet, und eine, die
langsam **innerhalb** der Frist antwortet, kommt durch. Die zweite Hälfte ist
die, ohne die die erste nichts sagte — eine Frist, die jede Antwort
abschnitte, ließe keinen Bildauftrag je zurückkommen.

**Und ein Bildauftrag ist daran gemessen** (A17,
`internal/openai/bilderwarten_test.go`): langsam innerhalb der Frist kommt
durch, jenseits scheitert er — **mit dieser Frist**. Bis D02 hieß das der
Wortlaut „timeout awaiting response headers"; seitdem trägt der Fehler die
Klasse `ErrKeineAntwort` und die Frist, die griff (unten). Auf bloß „zu
spät" zu prüfen wäre zu wenig: die Stillstandsfrist ist eine andere Aussage.
Die Gegenprobe (Kopffrist länger als der Stillstand) nennt dann die
Stillstandsfrist und wird rot.

**Ein Abbruch endet auch draußen** — gezählt an der Gegenstelle und nicht am
Aufrufer. **Zwei Zeilen gehören dafür in jede solche Vorrichtung**: sie muss
den Anfragerumpf **lesen**, bevor sie wartet. `net/http` startet den
Hintergrundleser, der einen Verbindungsabbruch bemerkt, erst wenn der Rumpf
verbraucht ist — und ein Bildauftrag ist ein POST. Ohne sie schlief der
Handler seine vollen 30 Sekunden weiter, während der Aufrufer längst zurück
war; gemessen am 2026-09-12, und es war eine Eigenschaft der **Vorrichtung**
und nicht der Naht.

**Diese vier Zeilen hält ein Wächter gegen den Quelltext**
(`internal/app/kifristendoku_intern_test.go`): er liest die Zellen als
**Dauern** und vergleicht sie mit den Konstanten der Nähte — „60 s" und
„1 min" sind dieselbe Frist, und eine Schreibweise zu verlangen wäre keine
Zusage über eine Zahl. **Was er nicht hält**, ist die Wirkung; die misst
der Absatz darüber.

**Was damit weiterhin unbelegt ist**: ob die Zahlen der drei KI-Nähte (2 bis
5 Minuten) für die Modelle reichen, die es wirklich gibt. Das sagt erst ein
Lauf gegen einen echten Dienst, und der kostet Geld.

**Die Stillstandsfrist greift auch unter HTTP/2** (gemessen): Kopf kommt,
Rumpf schweigt, die Frist schlägt zu. Das war nicht selbstverständlich — h2
teilt eine Verbindung auf mehrere Ströme, und die Frist hängt an der
Verbindung (Befund E-37, #916).

### Was ein Fehlschlag sagt — sechs Fälle und ein 429

**D02 (#2233).**
Bis dahin ging jeder Transportfehler roh hinaus — „… fragen: dial tcp: lookup
…: no such host" —, im Fenster als „Es ist ein unerwarteter Fehler
aufgetreten", und **mit der ganzen Adresse samt Abfrage** im Satz und im
Protokoll. Eingeordnet wird jetzt in `internal/netz/netzklassen.go`, **nach
Fehlertyp und Phase**, nicht nach Meldungstext (§8):

| Fall | erkannt an | Klasse (`domain/gegenstelle.go`) | Fenster: Schlüssel | REST |
| --- | --- | --- | --- | ---: |
| der Name löst sich nicht auf | `*net.DNSError` | `ErrNameUnaufloesbar` | `netzName` | 502 |
| niemand nimmt an | `ECONNREFUSED` oder eine Zeitüberschreitung **beim Wählen** | `ErrNichtErreichbar` | `netzNichtErreichbar` | 502 |
| keine Antwort | eine Zeitüberschreitung, **bevor der Kopf da ist** | `ErrKeineAntwort` | `netzKeineAntwort`, `{sekunden}` | 504 |
| die Antwort stockt | eine Zeitüberschreitung **beim Lesen des Rumpfes** | `ErrAntwortStockt` | `netzAntwortStockt`, `{sekunden}` | 504 |
| die Verbindung reißt ab | die Gegenstelle nahm an und schloss (`io.EOF`) oder setzte zurück (`ECONNRESET`), **bevor der Kopf da ist** — oder schloss **mitten im Rumpf** (`io.ErrUnexpectedEOF`); gemessen in `internal/netz/abriss_test.go` (#2487) | `ErrVerbindungAbgerissen` | `netzAbgerissen` | 502 |
| der Proxy verlangt eine Anmeldung | `407` auf `CONNECT` (`OnProxyConnectResponse`) **oder** als Antwort über einen Proxy für `http://` | `ErrProxyAnmeldung`, Wirt ist **der Proxy** | `netzProxyAnmeldung` | 502 |
| zu viele Anfragen | `429` | `ErrGedrosselt` (wie bisher) | `gedrosselt` | 502 |

**Jeder Satz nennt den nächsten Schritt** — im Fenster aus dem Katalog, an
der Kommandozeile als zweite Zeile (`cmd/mse/netzrat.go`); MCP gibt den Satz
durch. Der Code an allen Flächen ist `netz` (§16: erweiternd — bis dahin
`unbekannt` und 500; für den Abriss seit #2487 ebenso). **Keine Wiederholung**
durch diese Einordnung: ein Abriss wird einmal gefragt und gemeldet (§16,
Satz 4). **Was sie nicht erreicht**: `netz.Strom` gibt den Rumpf ungelesen
heraus, und was beim Lesen dort reißt, ordnet der Aufrufer ein — oder
niemand.

**Welche Frist der Satz nennt, ist die, die griff.** Bevor der Kopf da ist,
laufen zwei: `ResponseHeaderTimeout` und die Lesefrist, die der Wähler als
Stillstandsfrist auf die Verbindung legt. Mit den üblichen Werten (30 s und
60 s) greift die erste; ist der Stillstand kürzer, die zweite — und dann
steht sie im Satz.

**Ein Abbruch bleibt ein Abbruch — gefragt wird der Kontext, nicht der
Fehler.** Gemessen: die Kopffrist von `net/http` antwortet auf
`errors.Is(…, context.DeadlineExceeded)` mit **Ja** (`timeoutError.Is`,
Go 1.27.2). Hätte der Grund seine Kette behalten, wäre sie in `internal/app`
und `internal/rest` als Abbruch übersetzt worden — und das Fenster blendet
einen Abbruch aus. Der Grund reist deshalb **nur als Satz** mit, fürs
Protokoll. Gefunden hat das erst ein Lauf mit **getrennten** Fristen; mit
gleichen gewann die Lesefrist, und der Fall kam nie vor.

**Ohne Geheimnis.** Im Satz und im Protokoll steht die Adresse **ohne
Zugangsdaten und ohne Abfrage** — Schema, Wirt und Pfad; ein Proxy steht
ohne Kennwort darin. Das gilt auch für den Abbruch und für das 429, das die
Adresse bis dahin ganz nannte. **Der Vorgang führt ins Protokoll**: der
Satz im Fenster nennt die Kennung, unter der „Einstellungen › Diagnose"
zeigt, was der Transport sagte — im Satz und nicht nur als Knopf, denn die
Seiten machen aus einer Ursache sofort Text, und „Diagnose öffnen" steht nur
an drei Stellen.

**Gemessen am 2026-10-09** (`internal/netz/netzklassen_test.go`, n = 20,
Mac17,3, macOS 27.0.1, **neben** einer laufenden Hüllenprobe), mit einer
Kopffrist von 300 ms und einer Stillstandsfrist von 600 ms gegen
Testserver auf der Schleife:

| Fall | min | Median | max |
| --- | ---: | ---: | ---: |
| ein Name unter `.invalid` (RFC 6761) | 1 ms | 1 ms | 2 ms |
| ein geschlossener Port | 0 ms | 0 ms | 0 ms |
| keine Antwort — die Kopffrist greift | 301 ms | 302 ms | 303 ms |
| die Antwort stockt — die Stillstandsfrist greift | 601 ms | 602 ms | 603 ms |

**Über eine Bindung des Fensters** (`internal/app/netzdiagnose_test.go`):
`ScryfallSammeldatenHolen` mit umgelenkter Adresse, durch Dienst,
Scryfall-Klienten und `netz` mit dem Transport, den das Programm baut —
ein Name ohne Auflösung und ein Port ohne Annahme kommen als `netz` mit
Wirt und Vorgang an.

**Was das nicht belegt**: einen Namensdienst, der **hängt** statt zu
antworten (`.invalid` beantwortet der Resolver sofort); einen echten
Firmenproxy; die Einordnung über HTTP/2 jenseits der Stillstandsfrist; und
wie das Fenster den Satz zeigt — das gehört an das native Fenster.

---

### Wie alt eine Antwort ist — Stand der Quelle, Abrufzeit, Frist

**F04** (#2246):
drei Zeitangaben, die nicht dasselbe sind. Der **Stand der Quelle** ist, was
der Dienst über seinen Inhalt sagt — ein Beobachtungstag, ein `updated_at`.
Die **Abrufzeit** ist, wann dieses Programm gefragt hat. Die **Frist** ist,
wie lange es eine Antwort ohne neue Frage gelten lässt. Ein heute geholter
Preis kann von gestern sein, und eine Datei von vorgestern kann den Stand von
heute tragen. Stand am 2026-10-10, nachgelesen im Code:

| Dienst | Stand der Quelle | Abrufzeit | Frist | Ohne Netz |
| --- | --- | --- | --- | --- |
| **Scryfall** (Karten, Suche, Preise) | nennt keinen | Eintrag des Zwischenspeichers (`internal/netz/speicher.go`); Preise tragen `Erhoben` und zeigen ihr Alter | **24 h** (`scryfall.Speicherfrist`) | der letzte Eintrag mit seinem Alter; ein **unlesbares** Preisfeld widerruft den letzten gültigen Preis nicht mehr (`PreisUngeprueft`) |
| **Scryfall** (Sammeldaten) | `updated_at` des Angebots — gezeigt, solange Scryfall antwortet | **Änderungszeit der Datei** — gezeigt als „Geholt …“, auch ohne Netz | keine: neu geholt wird von Hand | die Datei mit ihrer Abrufzeit |
| **MTGJSON** (Preise) | Beobachtungstag je Preis — gezeigt als „Stand der Quelle …“ | `Erhoben` je Abruf | keine: geholt wird von Hand, eine Tagesdatei | die abgelegten Stände mit ihrem Alter; ein Abruf gilt ganz oder gar nicht |
| **MTGJSON** (Kartendatei) | `meta.date` wird gelesen und **nicht** gezeigt | keine — die Datei bringt jemand mit | — | — |
| **magicthegathering.io**, **CardTrader**, **MTGJSON GraphQL**, **Scrydex** (Suche, Konto) | nennen keinen | keine | keine | **nichts** — jede Antwort ist eine Frage im Augenblick und wird nicht abgelegt |
| **Scrydex** (Preisverlauf) | Tag je Punkt — in der Preiszeile als „Stand der Quelle …“; die Verlaufsliste darunter nennt die Abrufzeit | `Erhoben` je Abruf | keine | die abgelegten Punkte |
| **Paketbezug** | der gepinnte Stand und seine Prüfsumme; **seit G04** auf Wunsch die Spitze der Quelle als Commit samt Datum (`bezug.Spitze`, GitHubs Auskunft) | keine — die Spitze wird bei jeder Frage neu geholt und nicht abgelegt | keine — ein Stand ändert sich nicht | das eingebundene Paket; die Frage nach der Spitze ist ein eigener Zustand „unbekannt" mit Grund, nicht „aktuell" |

**Ob ein Stand noch stimmt, prüft ein Driftlauf**: Scryfall,
magicthegathering.io und MTGJSON stehen hinter `dienstdrift`, und `make belege`
wird rot, wenn ihr letzter Lauf älter als 30 Tage ist (`docs/HAERTUNG.md`,
Abschnitt 4).

**Was offen ist** — und hier steht, damit es nicht als erledigt gilt:

- **„Veraltet“ ist kein eigener Zustand.** Der Zwischenspeicher gibt ein Alter
  und kein „überfällig“; ein gescheiterter Abruf, hinter dem ein alter Eintrag
  steht, kommt als dieser Eintrag zurück, ohne dass jemand von dem Fehlschlag
  erfährt — nur das Alter sagt es. Die Preiszeile nennt einen Preis ab
  **sieben Tagen** veraltet; mit der Frist von 24 h hat das nichts zu tun.
- **Das Alter einer Preisliste ist das ihres ältesten Standes** und gilt für
  jede Zeile darin.
- **`meta.date` der MTGJSON-Kartendatei** wird gelesen und nirgends gezeigt,
  und die Verlaufsliste eines Preises nennt je Punkt nur die Abrufzeit.
- **Die Dienste ohne Ablage** haben ohne Netz nichts zu zeigen — das ist die
  Eigenschaft eines Live-Dienstes und keine Lücke, solange die Seite es sagt.

## Als Dienst

**Seit N6 (#1572)
sind es zwei**: `mse-rest` und `mse-mcp` binden beide an eine **einstellbare**
Adresse, und beide lassen sich aus dem Fenster schalten. Bis dahin stand hier
„nur `127.0.0.1`" und „`mse-mcp` bedient nur stdio"; beides galt bis zum
2026-09-15.

| | Wert | Wo |
| --- | --- | --- |
| Adresse | `--adresse`, **wiederholbar seit G08** (#2272) — jede bekommt ihren Lauscher. **Ohne sie** gilt für `mse-rest` die Einstellung `netz.lauschen` (so startet auch der LaunchAgent dort, wo gewählt wurde) und ohne diese **`127.0.0.1`**; `mse-mcp` bleibt ohne `--adresse` bei stdio, das Fenster gibt ihm die Wahl mit | `internal/domain/bindung.go`, `lauschwahl.go`, `internal/rest/vereint.go`, `internal/app/lauschwahlnaht.go` |
| Port | `--port`, `0` sucht einen freien | `cmd/mse-rest/dienst.go`, `cmd/mse-mcp/ueberhttp.go` |
| TLS | **keines auf der Schleife, Pflicht jenseits davon** | [0115](entscheidungen/0115-die-schleife-ist-eine-vorgabe.md) |
| Zeugnis | **Vorgabe**: selbst unterschrieben, auf der **Peeridentität**. **Wahlweise eine eigene PKI** — Zertifikat, Schlüssel, Kette, Anker aus `dienst.tls.*` oder `--zeugnis`/`--schluessel`/`--kette`/`--anker`, vor dem Start geprüft und am Handschlag getauscht ([0154](entscheidungen/0154-das-dienstzeugnis-ist-nicht-die-identitaet.md)). **TLS 1.3** als Untergrenze | `internal/leitung/dienstzeugnis.go`, `eigeneszeugnis.go`, `zeugniswerk.go`; die Wahl in `internal/app/dienstzeugnis.go` |
| Zugang | eine Marke je Öffnung; jenseits der Schleife ohne Ausnahme | `cmd/mse-rest/bedienen.go` |
| Herkunft | `Origin` wird geprüft — Schleife plus die **eigenen** Namen und Adressen, **und jede Adresse, auf der gelauscht wird**: ein Dienst auf `::1` wies seit N6 jede eigene Anfrage ab, behoben mit G08 | `internal/rest/herkunft.go`, `app.Rechnerteile` |
| Protokoll | **HTTP/1.1**, kein h2c | gemessen, `internal/rest/dienstfristen_test.go` |
| Lesefrist | 10 s | `cmd/mse-rest/dienst.go` |
| **Schreibfrist** | **5 min** | dito — sie zählt **ab dem Anfragekopf**, nicht ab der Antwort |
| Leerlauf | 2 min | dito |
| **Übertragung einer Datei** | **30 min** zum Lesen des Rumpfs, danach 30 min für die Antwort; zum Abholen 30 min; höchstens **1 GiB** | `internal/rest/dateiuebertragung.go` — **nur hinter der Marke** und nur für die Wege, die eine Datei nehmen oder abgeben (FMT-03.AK-03). Gemessen: mit der Lesefrist von 10 s brach ein Set über eine langsame Leitung ab (`internal/flaechen/uebertragung_test.go`) |
| gleichzeitige Verbindungen | 16 — **`mse-rest` und `mse-mcp` über HTTP**, durch denselben Lauscher; eine darüber wartet. Bis #2003 nur `mse-rest`, und die Zahl stand in keiner Auskunft — jetzt steht sie in `domain.Vorgabegrenzen` und als `verbindungen` in `grenzen` an jeder Fläche. **Ist jeder Platz belegt, wird eine ruhende Verbindung geschlossen**, die älteste zuerst — ein Pool mit `keepAlive` sperrte sonst alle anderen bis zum Leerlauf aus (QA-08, gemessen in [`LASTMESSUNG.md`](LASTMESSUNG.md)) | `internal/rest/gedraenge.go`, gesetzt in `cmd/mse-rest/bedienen.go` und `cmd/mse-mcp/ueberhttp.go`; gehalten am gebauten Werkzeugserver von `internal/flaechen/werkzeugserververbindungen_test.go` |
| **Aufträge** (nur `mse-rest`) | vier laufend, sechzehn beendete behalten, eine Stunde; ein fünfter bekommt `429` mit `Retry-After: 5`. **Ein Abbruch** (`DELETE /auftraege/{id}`) **kostet kein Guthaben und wird nie gedrosselt** (#2003) | `domain.Vorgabegrenzen`, `internal/rest/auftraege.go`; genannt als `auftraege` in `grenzen` |
| **Guthaben je Zugang und für alle** | 120 Einheiten je Minute und Zugang, 600 für alle — **abschaltbar** über `grenzen.kontingente` ([0161](entscheidungen/0161-die-kontingente-lassen-sich-abschalten.md)); Plätze, Rumpf- und Antwortgrenze, Fristen, Zugang und TLS bleiben. **Was ein Ruf kostet, sagt eine Tafel für beide Dienste** (`domain.Kostentafel`, #2003) — dieselbe Fähigkeit kostet über REST und MCP dasselbe | `internal/service/kontingente.go`; gemessen in [`LASTMESSUNG.md`](LASTMESSUNG.md) |

**Mehrere Adressen** (G08): jede gewählte bekommt einen Lauscher, und der
Dienst sieht sie vereint — Verbindungsgrenze, Fristen und Ende gelten für
alle zusammen. **Eine Adresse, die es beim Start nicht gibt** (eine
verschwundene Schnittstelle, ein belegter Port), wird angesagt (`Ohne …`) und
protokolliert, und die übrigen laufen; **ohne eine einzige** startet nichts.
**Verlässt eine den Rechner, gelten TLS und Token für den ganzen Dienst**, und
ein eigenes Zeugnis muss jede dieser Adressen tragen. Gehalten von
`internal/rest/lauschenaufallen_test.go`, `cmd/mse-rest/lauschwahl_test.go`,
`cmd/mse-mcp/lauschwahl_test.go` und `internal/app/dienstzeugnismehr_intern_test.go`.

**Die Adressbindung bleibt eine Zusage mit wirksamer Gegenprobe**
(`internal/rest/lauscher_test.go`, `internal/rest/bindungjenseits_test.go`):
ein Name wie `localhost` wird abgewiesen statt aufgelöst, und **ohne Zeugnis
geht jenseits der Schleife der Lauscher gar nicht erst auf** — die Prüfung
steht **vor** `net.Listen`, damit kein Port für die Dauer eines Fehlerwegs
offen steht.

**Kein h2c, und das ist ein Nein mit Grund**: auf der Schleife ohne TLS
brächte Multiplexing nichts — es gibt einen Aufrufer —, und der Upgrade-Weg
wäre eine Fläche mehr, die jeder Aufrufer kennen müsste. Ein Lauf hält das
Nein fest, damit es eine Messung bleibt.

### Was jenseits der Schleife wirklich geht — gemessen

**Am 2026-09-15 auf diesem Rechner**, mit gebautem `mse-rest` an `0.0.0.0` und
Anfragen über die **eigene** LAN-Adresse (`internal/flaechen/dienstbindung_test.go`):

| Anfrage | Antwort |
| --- | --- |
| `https://<LAN>:<port>/sets` mit Marke | `200` |
| dieselbe ohne Marke | `401` |
| `http://<LAN>:<port>/sets` mit Marke | **keine** — der TLS-Lauscher liest den Klartext als Handschlag |
| über den **Namen** dieses Rechners statt über die Adresse | `200` |
| auf `127.0.0.1` gebunden, über die LAN-Adresse gefragt | **keine** — Verbindung abgelehnt |

**Und der angesagte Fingerabdruck ist der des vorgezeigten Zeugnisses** —
gemessen am echten Handschlag (`vorgezeigterAbdruck`), nicht aus derselben
Identität zweimal gerechnet. Das ist die Zusage aus 0115: wer sich verbindet,
hat nichts anderes zu vergleichen.

### Womit man den Dienst **nicht** erreicht

**Das mitgelieferte `curl` von macOS kann das Zeugnis nicht.** Gemessen am
2026-09-15:

| Werkzeug | Ed25519-Zeugnis | ECDSA-Zeugnis (Gegenprobe) |
| --- | --- | --- |
| `/usr/bin/curl` 8.7.1, LibreSSL 3.3.6 | **Rückgabe 35** — der Handschlag scheitert | `200`, Rumpf kommt an |
| `/usr/bin/openssl` = LibreSSL 3.3.6 | `Cipher is (NONE)` | — |
| OpenSSL 3.6.4 (Homebrew) | TLS 1.3, `Server public key is 256 bit` | — |
| Go 1.26 `net/http` | `200` | — |

**Die Gegenprobe ist der Punkt.** Derselbe Server, dasselbe TLS 1.3, dasselbe
`-k` — nur der Schlüsseltyp unterscheidet sich, und damit steht fest, dass es
an Ed25519 liegt und nicht am selbst unterschriebenen Zeugnis, am Port oder an
der Fassung des Protokolls. LibreSSL 3.3.6 kennt Ed25519 in einem
Serverzeugnis nicht.

**Was daraus folgt**: wer von der Kommandozeile zugreift, nimmt `curl` aus
Homebrew, ein Programm mit OpenSSL, oder eine Sprache mit eigener
TLS-Bibliothek — Go, Python, Node. **Was nicht daraus folgt**: ein Wechsel des
Schlüsseltyps. Das Zeugnis trägt die **Peeridentität**, und die ist Ed25519
(N1); ein zweites Schlüsselpaar allein für den Dienst wäre eine zweite
Identität desselben Profils.

**Seit NET-03 gibt es den zweiten Weg**
([0154](entscheidungen/0154-das-dienstzeugnis-ist-nicht-die-identitaet.md)):
mit einem ECDSA- oder RSA-Zeugnis aus einer eigenen Stelle erreicht auch
`/usr/bin/curl` den Dienst, **mit `--cacert` und ohne `-k`** — gemessen am
2026-09-30 am gebauten `mse-rest` (`internal/flaechen/dienstpki_test.go`).
Der Schlüsseltyp der Peeridentität bleibt dabei, wie er ist.

## Als Gegenstelle: die Peerleitung

**Die dritte Art, und sie ist weder Klient noch Dienst** — Werkstück N2 von
Phase 5.4
(#1455).
`internal/leitung` wählt **und** horcht, spricht kein HTTP und wird von keinem
der beiden Abschnitte darüber erfasst.

| | Wert | Wo |
| --- | --- | --- |
| Protokoll | **TCP mit TLS 1.3**, Mindestfassung erzwungen | `internal/leitung/tls.go` |
| Nachrichtengrenzen | Längenpräfix über vier Byte, höchstens ein Mebibyte je Nachricht | `internal/leitung/leitung.go` |
| Zeugnis | **selbst unterschrieben**, Schlüssel ist die Peeridentität, je Verbindung frisch | `zeugnisFuer` |
| Prüfung | **Festnageln auf die erwartete Kennung**, keine Zertifizierungsstelle, kein Rechnername | `pruefeGegenueber` |
| Wo geprüft wird | `VerifyConnection` und **nicht** `VerifyPeerCertificate` | dito |
| Kanalbindung | `ExportKeyingMaterial`, Marke `mse4mac/kanal/1`, 32 Byte | `schliesseHandschlagAb` |
| Adresse beim Horchen | **die, die angegeben wird** — keine Einengung auf die Schleife | `Horche` |
| Port | der angegebene; `0` sucht einen freien | dito |
| Wahlfrist | 20 s, Verbindung **und** Handschlag | `Wahlfrist` |
| Handschlagfrist beim Horchen | 20 s je hereinkommender Verbindung | `Handschlagfrist` |
| **Rahmenfrist** | **60 s** für den Rumpf eines angekündigten Rahmens, auch in einer Sitzung ohne Frist; eine frühere Frist des Kontexts gilt vor ihr (D05, #2236) | `rahmenfrist`, `rumpffristSetzen` |
| **Bildfrist am Gast** | **10 s** auf die Antwort des Wirts zu einem Bildwunsch; danach ein Konflikt, und eine späte Antwort fällt still weg. Ein Bild darf mit Umschlag höchstens die Nachrichtengrenze füllen — der Wirt prüft es vor dem Senden, der Gast prüft, dass es ein PNG mit den genannten Maßen ist (P9, #2377, [0198](entscheidungen/0198-kartenbilder-am-fremden-tisch.md)) | `app.Bildfrist`, `service.bildreserve`, `app.pruefePNG` |
| **Weiterreichfrist am Wirt** | **8 s** auf die Antwort eines Gastes, dessen mitgebrachtes Deck die erfragte Karte stellt — **kürzer als die Bildfrist** des Fragenden, damit der einen Grund liest statt selbst aufzugeben; gefragt wird nie der Fragende selbst und nur ein Gast, dessen eigener Bestand die Zeile hält. **Gewartet wird nebenher**, nicht in der Empfangsschleife — zwei Gäste, die zugleich nach der Karte des anderen fragten, warteten sonst beide die Frist ab (P9b, #2377, [0199](entscheidungen/0199-ein-gast-bringt-sein-deck-mit.md); Zweitblick über Phase 7, Befunde 5 und 6) | `service.weiterreichfrist` |
| Proxy | **keiner** — gewählt wird über einen `net.Dialer`; `HTTP_PROXY` und die Einstellung gelten den HTTP-Klienten darüber, nicht dieser Leitung | `tls.go` |
| Proxy | **keiner** — das ist kein HTTP | — |

### Warum die Adresse hier nicht auf die Schleife eingeengt ist

**Weil zwei Rechner sich sonst nicht erreichen.** `internal/rest` bindet in
der Vorgabe an `127.0.0.1` ([0046](entscheidungen/0046-doch-ein-dienst-auf-der-rueckschleife.md));
seit N6 und G08 ist die Adresse wählbar, und jenseits der Schleife gelten TLS
und ein Zugang ohne Ausnahme (siehe „Als Dienst“ oben). Auf der Schleife ist
die Einengung dort also die Vorgabe und keine Grenze des Pakets. Hier wäre
dieselbe Einengung das Ende der Sache.

**Was an ihre Stelle tritt, sind drei andere Zusagen**: es wird **nur auf
Aufruf** gehorcht, jede hereinkommende Verbindung muss ein Ed25519-Zeugnis
vorzeigen, und **wer eintreten darf, entscheidet die Kontaktliste** nach dem
Handschlag — `Peerkontakt.Darf` fragt Stand und Zweck.

### Was ein Gegenüber eintippen soll — und was ein Lauscher stattdessen meldet

**Ein Lauscher sagt, wo er gebunden hat, nicht, wo jemand ankommt.** Gemessen
am 2026-09-14 auf diesem Rechner:

```
:41234          -> [::]:41234
127.0.0.1:41234 -> 127.0.0.1:41234
:0              -> [::]:55260
```

Das Fenster zeigte genau diese Zeichenketten unter „Erreichbar unter …". Das
erste ist keine Adresse, zu der sich wählen ließe; das zweite zeigt auf den
eigenen Rechner. Brauchbar war `192.168.178.51:41234`, und die stand an keiner
Fläche. Behoben mit `leitung.ErreichbarUnter`.

**Ein Name gehört dazu, weil eine Adresse geliehen ist.** Sie kommt in den
meisten Netzen über DHCP, und eine Lease endet; wer sie weitergibt, gibt eine
Momentaufnahme weiter. Ein Name, den das Netz auflöst, zeigt nach dem Wechsel
auf die neue Adresse.

**Und er wird geprüft statt geraten.** Am selben Tag gemessen (der
Rechnername ist für die Veröffentlichung ersetzt, die Gestalt nicht):

```
os.Hostname()                               = "MacBook-Air-von-Ada.fritz.box"
LookupHost("MacBook-Air-von-Ada.fritz.box") = [192.168.178.51 2003:… fdc0:…]
LookupHost("…fritz.box.local")              = []  (i/o timeout nach 3 s)
```

Der Router vergibt hier einen auflösbaren Namen; ein blind angehängtes
`.local` ergibt einen Namen, den niemand kennt, und kostet drei Sekunden. Ein
Name zählt deshalb nur, wenn er auf eine Adresse **dieses** Rechners auflöst —
sonst schickte er jemanden zu einem Fremden. `.local` kommt nur an einen
Namen **ohne** Punkt.

| | gilt | Grenze |
| --- | --- | --- |
| **Bindeadresse** | immer | sagt nichts darüber, wo jemand ankommt |
| **IP-Adresse** | in diesem Netz | eine Lease endet |
| **Geprüfter Name** | in diesem Netz, über den Wechsel hinweg | er muss aufgelöst werden können |
| **Peerkennung** | überall | trägt erst mit N3 ([0112](entscheidungen/0112-tailcat-statt-eines-eigenen-dienstes.md)) |

**Ausgelassen werden** die Rückschleife (sie zeigt auf den Fragenden selbst),
Link-Local (`fe80::…` ist ohne Zonenangabe nicht wählbar, und die Zone des
einen Rechners sagt über die des anderen nichts) und alles, was keine
IP-Adresse ist. **Wer auf der Rückschleife horcht, bekommt eine leere Liste**
— und keine erfundene Adresse.

**Die Namensauflösung hat eine Frist von zwei Sekunden für alle Anwärter
zusammen** (`leitung.Namensfrist`), nicht je einer: ein Name, der nicht
auflöst, läuft in eine Zeitüberschreitung, und wer je Anwärter wartete, hielte
den Knopfdruck doppelt so lange auf.

### Warum `VerifyConnection`

**Weil `VerifyPeerCertificate` bei einer wiederaufgenommenen Sitzung nicht
gerufen wird.** TLS 1.3 nimmt über ein Ticket wieder auf, und dann prüfte
niemand mehr das Zeugnis — das Festnageln wäre still übergangen.
Angestrichen von gosec (G123) am 2026-09-14, und der Befund war richtig.

### Was gemessen ist

| Aussage | Wie |
| --- | --- |
| Ein Mitleser auf der Strecke sieht die Nachricht nicht | ein weiterleitender Lauscher zwischen beiden schreibt jedes Byte mit; „Mausbraten" steht nicht darin. Gegenprobe: über den rohen Strom steht es darin |
| Ein anderes Gegenüber als erwartet wird abgewiesen | zwei Identitäten, gewählt wird mit der Kennung einer dritten |
| Ein Zeugnis mit ECDSA kommt nicht durch | ein Klient mit P-256-Zeugnis; die horchende Seite weist ab. Gegenprobe: mit `if erwartet == "" { return nil }` meldet der Lauf `<nil>` |
| Beide Enden nennen denselben Kanal, und er steht nicht in den Adressen | verglichen, und gegen `127.0.0.1`, die Adresse und `tcp` geprüft |
| Eine unerreichbare Adresse endet mit einem Fehler und nicht mit Warten | ein Lauscher, der öffnet und sofort wieder zumacht |
| Ein schweigendes Gegenüber hält den Horcher nicht fest | eine Verbindung ohne ein einziges Byte; der Handschlag hängt am Kontext. Gegenprobe: ohne ihn wartet der Lauf fünf Sekunden vergeblich |
| **Eine Einladung wird über dieselbe Leitung eingelöst** (NET-01b, [0156](entscheidungen/0156-eine-einladung-genuegt.md)) | eine Paarungssitzung, dann zwei Nachrichten; **Wählen, Handschlag und Antwort zusammen höchstens 15 s** (`service.Einloesefrist`), auf beiden Seiten. Zwei Anwendungen über TCP auf der Rückschleife (`TestEineEinladungGenuegt`) und drei Prozesse über jede Fläche (`TestEineEinladungGenuegtUeberJedeFlaeche`). Gegenprobe: ein Horcher, der die Paarung nicht beantwortet — **so verhält sich die Fassung vor NET-01b** —, liefert sofort „unbeantwortet" und keine Frist |

---

## Als Gegenstelle über fremde Netze: die Fernleitung

**Neu mit N3** (#1456),
entschieden in [0114](entscheidungen/0114-tailcat-gepinnt.md).

| | `internal/fernleitung` |
| --- | --- |
| **Unterbau** | `tailscale/tailcat v0.7.0` — WireGuard, magicsock, DERP |
| **Verschlüsselung** | zweifach: WireGuard im Tunnel, **darüber** TLS 1.3 mit dem festgenagelten Zeugnis aus N2 |
| **Rendezvous** | ein **DERP-Relais von Tailscale**, Karte von `https://tailcat.dev/derpmap.json` — so nachgelesen am 2026-09-14 ([0112](entscheidungen/0112-tailcat-statt-eines-eigenen-dienstes.md)); seither nicht erneut abgerufen, **nicht nachgeprüft am 2026-10-11** |
| **Port im Tunnel** | fest, `4123` — der Tunnel gehört diesen beiden Rechnern allein |
| **Frist zum Wählen** | 60 s, gegen 20 s im eigenen Netz: erst Rendezvous, dann der Versuch des direkten Weges |
| **Von selbst** | **nichts.** `Horche` und `Waehle` entstehen nur auf Aufruf |
| **Proxy** | **Zum Relais derselbe wie für jeden Klienten**, seit NET-02 (#1706): Einstellung vor Umgebung, je Wählen gerechnet, Anmeldung als Basic aus der Adresse. Bis dahin **keiner** — tailcat fragt vor dem Wählen eines DERP-Knotens `feature.HookProxyFromEnvironment` (`derphttp.dialNode`), und gesetzt hätte ihn erst `tailscale.com/feature/useproxy`, das dieser Bau nicht bindet (`go list -deps .`, 2026-09-23 und 2026-09-30). Jetzt setzt `internal/fernleitung/proxy.go` den Haken; **gemessen am echten `derphttp`-Klienten**: sein `CONNECT` samt `Proxy-Authorization` kommt beim Lauscher an (`proxyrelais_intern_test.go`). **Die Gegenstelle selbst geht nie über den Proxy** — WireGuard spricht UDP; hinter einem Proxy, der nur sich selbst durchlässt, gibt es deshalb nur den Weg über das Relais |
| **Weg je Gegenüber** | **direkt**, **über das Relais** oder **im Aufbau** — aus tailcats Stand der Gegenstellen (`CurAddr`, `Relay`), nur an der Seite, die eingeladen hat, und auf Knopfdruck gelesen (NET-02.AK-03). **Bis v0.6.0 war dieser Stand immer leer** (`locoBackend.Status` ohne `WantPeers`, SEC-21, #2162); **seit v0.7.0 ist er gefüllt** (K01, #2187) — gemessen am echten DERP-Weg auf einem Rechner: gewöhnlich `direkt` samt Relaisregion, mit erzwungenem DERP nur das Relais (`docs/belege/sec21-fernweg.md`). Gezählt werden daneben weiter die **offenen Ströme**: ein Eintrag im Stand ist eine Gegenstelle, kein Strom, und zwischen Anbinden und erstem Weg heißt der Weg „unbestimmt" statt „niemand" |
| **Rückfall aufs Relais** | **magicsocks, nicht unserer** ([0182](entscheidungen/0182-der-rueckfall-aufs-relais-gehoert-magicsock.md), D03): ein direkter Weg gilt 6,5 s nach seiner letzten Bestätigung, danach geht jedes Paket an beide Wege. Gemessen am 2026-10-10 auf einem Mac: direkt < 1 ms; nur Relais (DERP Frankfurt) Median 26–41 ms; **UDP einer Seite mitten in der Verbindung gesperrt: Lücke 6,46 s**, danach Relais, kein Verlust; UDP wieder offen: direkt nach 0,3 s |
| **Spiel und Handel über den Fernweg** | **Gemessen am 2026-10-10 auf einem Mac** (D04, #2235), zwei Anwendungen mit eigenem Bestand: **Partie** — Beitritt über TCP 27 ms, über tailcat direkt 1,1 s, nur über das Relais 4,2 s; ein verdeckter Zug des Wirts und ein Zug des Gastes danach je 23 ms direkt und rund 110 ms über das Relais, die verdeckte Karte beim Gast ohne Namen (`internal/app/fernpartie_netzlauf_intern_test.go`). **Handel** zwischen zwei Prozessen über `mse fernhorchen` bis zur offenen Übergabe: 7,9 s wie tailcat wählt, 12,9 s nur über das Relais, beide Stände bis auf `ich` gleich (`internal/flaechen/fernhandel_netzlauf_test.go`). **Was es nicht ist**: zwei NAT-Grenzen — der direkte Weg ging über die globale IPv6-Adresse dieses Macs |
| **Protokoll** | tailcats Sätze gehen bereinigt hinaus: Zugangsdaten in Adressen und — seit SEC-21 — **Knoten- und Disco-Schlüssel** als `nodekey:…`/`discokey:…`; beide sind Teil einer Einladung |

### Was an das Relais geht — und was nicht

**Verschlüsselter WireGuard-Verkehr, und sonst nichts.** Keine Karten, keine
Decklisten, keine Namen. Das ist keine Regel an einen fremden Dienst, sondern
eine Eigenschaft des Aufbaus.

**Was das Relais trotzdem sieht**: dass zwei Punkte miteinander reden, wie viel
und wann. Verbindungsdaten sind Daten.

**Und es ist ein Dienst, den wir nicht betreiben.** Tailscale sagt für seine
öffentlichen Relais nichts zu — „no uptime SLAs or throughput targets" —,
drosselt im Relaisfall und komprimiert nicht. Für N4 ist der Durchsatz damit
eine reale Größe. **Die Quelle** ist laut [0112](entscheidungen/0112-tailcat-statt-eines-eigenen-dienstes.md)
das README von tailcat, nachgelesen am 2026-09-14 (die Stelle nennt der Vermerk
nicht einzeln); der Wortlaut ist seither nicht erneut abgerufen — **nicht
nachgeprüft am 2026-10-11**.

### Warum die Einladung eine Fassung trägt

**Weil tailcats Drahtformat sich ändern darf.** Bricht es, kommt gar keine
Verbindung zustande: der signierte Handschlag aus N1 prüft seine eigene Fassung
und weist ab — aber er läuft nie. Ein Mensch läse „nicht erreichbar", und das
wäre ein Verbindungsfehler, der in Wahrheit ein Fassungsfehler ist.

Eine Einladung heißt deshalb `mse4mac:1:<tailcat-Addr>` und wird **vor** dem
Wählen geprüft. **Was das nicht leistet**: eine Einladung derselben Zahl, deren
tailcat sich trotzdem anders verhält, endet weiter in einem
Verbindungsfehler. Die Zahl schützt vor dem **bekannten** Bruch.

---

## Als Rufer und Gerufener im selben Netz: die Entdeckung

**mDNS, und nur auf Aufruf** — Werkstück N2 von Phase 5.4
(#1455).
`internal/entdeckung` kündigt an und fragt; von selbst tut es keines von beidem.

| | Wert | Wo |
| --- | --- | --- |
| Protokoll | **mDNS** über `hashicorp/mdns` v1.0.7, Multicast UDP 5353 | `internal/entdeckung` |
| Dienstart | `_mse4mac._tcp` in `local` | `Dienstart`, `Bereich` |
| Was hinausgeht | **Peerkennung** und selbstgewählter Name, dazu der Port des Horchpostens | `Melde` |
| Was **nicht** hinausgeht | der **Rechnername** dieses Macs — die Bibliothek nähme ihn ohne Zutun aus `os.Hostname()` | dito |
| Name der Instanz | der **Fingerabdruck**, nicht die Kennung — ein DNS-Etikett trägt höchstens 63 Byte | dito |
| Adressen | die der gewählten Schnittstelle, sonst die aller — **keine Namensauflösung** | `adressenVon` |
| Was **hereinkommt** | Adresse **und** `.local`-Name, beides aus der Antwort — zum **Wählen** die Adresse, zum **Merken** der Name (#1633). **Weiterhin keine eigene Auflösung**: was mDNS mitschickt, wird benutzt, mehr nicht | `Fund.Merkadresse` |
| Schnittstelle | **jede taugliche** — aktiv, multicastfähig, keine Rückschleife, mit IPv4-Adresse; `MSE_ENTDECKUNGSSCHNITTSTELLE` schlägt die Wahl, und ohne taugliche bleibt es bei der des Betriebssystems | `Taugliche`, `internal/app` |
| Protokollwahl einer Suche | **erst beide, bei „unerreichbar" noch einmal nur IPv4** — der Grund steht unten | `Suche`, `frageEinmal` |
| Frist einer Suche | höchstens 5 s, dazu **1,5 s Nachlauf** | `Hoechstfrist`, `Nachlauf` |
| Warteschlange | 64 Antworten | `Schlangenlaenge` |
| Proxy | **keiner** — das ist kein HTTP | — |

### Warum über jede Schnittstelle gefragt wird

**Weil „die vom System zugewiesene" ein Glücksspiel ist.** `hashicorp/mdns`
ruft `setInterface` nur, wenn eine genannt ist; sonst wählt das
Betriebssystem. Beim Ankündigen dasselbe: `net.ListenMulticastUDP("udp4",
nil, …)` horcht nur auf der vom System zugewiesenen — und **wer eine Frage
vom WLAN nicht hört, beantwortet sie auch nicht**.

**Gemessen am 2026-09-21**: dieser Rechner führt acht Schnittstellen, und
**genau eine** taugt — `en0` mit `192.168.178.74`, das WLAN. Genau die, die
vorher von Hand eingetragen werden musste.

**Gefragt wird nebenläufig**, weil eine Suche bis zu 6,5 Sekunden dauert;
vier Schnittstellen nacheinander wären eine halbe Minute. **Und ein Fehler
nimmt die anderen nicht mit**: ein VPN ohne Route darf nicht verhindern, dass
das WLAN antwortet. Gescheitert ist die Suche erst, wenn **jede**
Schnittstelle scheitert.

### Warum eine Suche zweimal fragen kann

**Weil die Bibliothek beim ersten Schreibfehler abbricht.**
`hashicorp/mdns@v1.0.7`, `client.go`, `sendQuery`:

```go
if c.ipv4UnicastConn != nil {
    _, err = c.ipv4UnicastConn.WriteToUDP(buf, ipv4Addr)
    if err != nil { return err }
}
if c.ipv6UnicastConn != nil {
    _, err = c.ipv6UnicastConn.WriteToUDP(buf, ipv6Addr)
    if err != nil { return err }
}
```

Geschickt wird **erst über IPv4, dann über IPv6** — und jeder Schreibfehler
kehrt zurück. Gibt macOS für `ff02::fb` ein `EHOSTUNREACH`, weil für
IPv6-Multicast keine Route steht, scheitert die ganze Abfrage, **obwohl der
IPv4-Ruf schon hinausgegangen war**. Zwei Anwendungen im selben WLAN fanden
einander deshalb nicht; der Ausweg war, die IPv4-Schnittstelle von Hand zu
nennen.

**Beim Ankündigen gibt es das nicht**: `server.go` verwirft den Fehler des
IPv6-Listeners (`ipv6List, _ := net.ListenMulticastUDP(…)`) und scheitert nur,
wenn **beide** fehlen. Die Ankündigung war tolerant, die Suche nicht.

**Und tolerant hieß bis K02 auch: blind.** Gemessen am 2026-10-07 in einem
macOS-15.2-Gast mit der gebauten Anwendung (`belege/k02-lokalesnetz.md`):
verwehrt macOS die Erlaubnis, binden die Listener trotzdem, und der Server
sendet erst als Antwort — deren Fehler verschluckt `hashicorp/mdns`. Seit K02
(#2188)
schickt `Melde` nach dem Aufstellen **eine Frage nach der eigenen Dienstart an
`224.0.0.251:5353`** (`internal/entdeckung/hinaus.go`), über die Schnittstellen
der Posten, bis eine hinausgeht; geht keine, werden die Posten abgebaut, und
die Ankündigung endet wie die Suche mit `ErrOhneNetzerlaubnis` oder
`ErrOhneNetzroute`. **Beim ersten Start** löst genau diese Probe die
Rückfrage von macOS aus und scheitert, solange sie offen ist — der Satz am
Feld sagt deshalb, dass nach dem Erlauben ein zweiter Versuch nötig ist.

**Deshalb fragt eine gescheiterte Suche ein zweites Mal, mit `DisableIPv6`** —
und erst wenn auch das mit „unerreichbar" endet, meldet sie die fehlende
Erlaubnis *Lokales Netzwerk*. Das trennt zwei Fälle, die vorher beide als
derselbe endeten.

**Was der Rückfall kostet**: ein Gegenüber, das **nur** über IPv6 erreichbar
ist, wird im zweiten Ruf nicht gehört. Deshalb steht IPv6 im **ersten**.

**Was er nicht belegt**: dass zwei Rechner einander im WLAN finden. Der Lauf
dazu speist einen Schreibfehler ein; `docs/HAERTUNG.md` sagt für den ganzen
mDNS-Weg, dass er über `lo0` geht — „Multicast über die Rückschleife ist nicht
Multicast über ein WLAN".

### Warum die Frist zweimal genannt wird

**Weil eine davon nicht die Wanduhr ist.** Gemessen am 2026-09-14 über `lo0`:

```
ctx 200 ms, Timeout 5 s   → 5,00 s     Timeout 1 s → 1,00 s
ctx 5 s,    Timeout 200 ms → 0,20 s     Timeout 2 s → 3,01 s
                                        Timeout 5 s → 6,00 s
```

**Der Kontext bricht das Warten nicht ab** — die Bibliothek sagt in ihrer
Dokumentation, sie achte „auf den Kontext **und** die Frist"; sie tut es nicht.
**Und ab zwei Sekunden Frist kommt eine hinzu**: sie fragt ein zweites Mal. Was
ein Aufrufer erwarten darf, heißt deshalb `Frist + Nachlauf`.

### Was der Lauf zeigt und was nicht

**Er geht über `lo0`** — nachgemessen trägt die Rückschleife Multicast
(`flags=up|loopback|multicast|running`), und was darüber geht, verlässt diesen
Rechner nicht. **Nicht gezeigt** sind damit ein zweiter Rechner, ein WLAN mit
Client-Isolation und die lokale Netzwerkberechtigung von macOS; sie stehen in
[`HAERTUNG.md`](HAERTUNG.md).

**Und `&net.Interface{}` ist nicht „keine"**, sondern „alle": ein Index von
null gab am 2026-09-14 neunzehn Adressen. Deshalb entscheidet in
`internal/entdeckung` `nil` und nicht ein leerer Verbund.
---

## Die Matrix — welcher Weg wo trägt

**NET-02.AK-01** (#1706):
„Eine getestete Matrix umfasst gleicher Rechner, gleiches LAN und
verschiedene Netze jeweils mit/ohne Proxy; VPN, mehrere Interfaces, IPv4/IPv6
und Netzwechsel ergänzen die Matrix. Nicht unterstützte Wege werden
ausdrücklich ausgewiesen."

**Vier Marken, und nur eine davon heißt „geht"**:

- **gemessen** — mit dem Lauf, der es zeigt;
- **von Hand** — eine Abnahme mit zwei Rechnern oder echtem Netz, geführt
  in [`HANDPRUEFUNGEN.md`](HANDPRUEFUNGEN.md) (die Sammelabnahme
  #1871 ist
  seit dem 2026-10-01 geschlossen; was sie offen ließ, steht dort);
- **nicht unterstützt** — mit dem Grund und dem Ersatzweg;
- **ungeprüft** — weder gemessen noch ausgeschlossen.

**Vier Wege**, denn sie gehen verschieden hinaus:

- die **Peerleitung** (`internal/leitung`, TCP mit TLS 1.3) trägt Handel und
  Tisch im eigenen Netz;
- die **Entdeckung** (`internal/entdeckung`, mDNS) findet Nachbarn;
- die **Fernleitung** (`internal/fernleitung`, tailcat) trägt über fremde
  Netze;
- die **Dienste** (`internal/netz` und seine Verbraucher) holen Karten,
  Preise und Antworten von KI-Anbietern über HTTP.

| Fall | Peerleitung | Entdeckung | Fernleitung | Dienste |
| --- | --- | --- | --- | --- |
| **Gleicher Rechner, ohne Proxy** | **gemessen**: zwei Prozesse über einen echten Port (`flaechen/handelsdraht_test.go`); Ausfall, Verlust, Latenz, Neustart und Schlüsselwechsel über einen Vermittler (`flaechen/netzfaelle_test.go`) | **gemessen** über `lo0` (`flaechen/nachbarn_test.go`) und über echtes Multicast hinter `netzlauf` (`entdeckung/zweiinstanzen_netzlauf_test.go`) | **gemessen** auf einem Rechner über echtes DERP (SEC-21, 2026-10-07, `belege/sec21-fernweg.md`; D04, 2026-10-10, oben); die Entscheidungen des Horchers (`fernleitung_intern_test.go`) | **gemessen** (`flaechen/nullnetz_test.go` ohne Netz, die `netzlauf`-Läufe gegen die echten Dienste) |
| **Gleicher Rechner, mit Proxy** | **nicht betroffen**: `leitung` wählt mit `net.Dialer` und kennt `internal/netz` nicht (Schichtenregel) — ein Proxy gilt für HTTP, die Peerleitung ist keines | **nicht betroffen**: mDNS ist UDP-Multicast | **gemessen** bis zum Proxy: der Weg zum Relais geht durch ihn (`proxyrelais_intern_test.go`) | **gemessen**: dieselbe Einstellung an jeder Fläche, Einstellung vor Umgebung (`flaechen/netzweg_test.go`) |
| **Gleiches LAN, ohne Proxy** | **über eine Systemgrenze gemessen**: TCP vom Wirt in einen Horcher der Abnahme-VM (2026-10-01) und TLS 1.3 zwischen zwei VM-Gästen in beide Richtungen (2026-10-07, `belege/sec22-vmlauf.md`); **zwei echte Geräte im WLAN: offen** ([`HANDPRUEFUNGEN.md`](HANDPRUEFUNGEN.md)) | **gemessen** zwischen Wirt und VM (2026-10-01) und zwischen zwei Gästen (2026-10-07); die gebaute Anwendung hat unter *Lokales Netzwerk* gefunden und ist gefunden worden (2026-10-07, `belege/k02-lokalesnetz.md`); zwei echte Geräte offen (`HANDPRUEFUNGEN.md`) | **zwischen zwei VM-Gästen im selben Netz gemessen**, gewöhnlich und mit erzwungenem Relais (2026-10-07, `belege/sec22-vmlauf.md`) — im selben Netz ist die Peerleitung der kürzere Weg | wie oben — das LAN ändert an HTTP nichts |
| **Gleiches LAN, mit Proxy** | **nicht betroffen**, wie oben — und **im selben Netz der Normalfall**: ein Firmenproxy hindert zwei Rechner im selben Segment nicht | **nicht betroffen** | wie „gleicher Rechner, mit Proxy" | wie oben |
| **Verschiedene Netze, ohne Proxy** | **ungeprüft** über eine Portfreigabe mit „Adresse von Hand"; **zugesagt ist es nicht** — der Weg dafür ist die Fernleitung | **nicht unterstützt**: Multicast überquert keinen Router. Ersatz: Adresse von Hand oder Einladung | **zwei Netze hinter verschiedenen Routern: offen** ([`HANDPRUEFUNGEN.md`](HANDPRUEFUNGEN.md); #1871 ist geschlossen) — der Zweck dieses Weges; bis zum Umschalten auf das Relais auf einem Mac gemessen (D03, 2026-10-10, oben) | wie oben |
| **Verschiedene Netze, mit Proxy** | **nicht unterstützt**: die Peerleitung geht nicht durch einen HTTP-Proxy. Ersatz: die Fernleitung | **nicht unterstützt** | **nur über das Relais**: der Weg zum Relais geht durch den Proxy (gemessen bis dorthin), ein direkter Weg entsteht nicht; **Ende zu Ende offen** ([`HANDPRUEFUNGEN.md`](HANDPRUEFUNGEN.md); #1871 ist geschlossen) | wie oben |
| **VPN** | **Routing entscheidet** (AK-02): die Adresse wird gewählt, das Betriebssystem routet; **das Programm erkennt kein VPN** und schaltet deshalb auch nichts frei. Über ein VPN **offen** ([`HANDPRUEFUNGEN.md`](HANDPRUEFUNGEN.md); #1871 ist geschlossen) | **nicht zugesagt**: ob ein VPN Multicast weiterreicht, entscheidet das VPN. Ersatz: Adresse von Hand | **ungeprüft** | wie oben — ein VPN ist für HTTP eine Route |
| **Mehrere Schnittstellen** | das Betriebssystem wählt nach der Route | **gemessen**: gefragt wird über **jede** taugliche Schnittstelle, eine klagende fällt allein aus (`entdeckung/schnittstellen_intern_test.go`); **übersteuerbar** mit `netz.schnittstelle`, leer heißt Automatik | tailcat wählt selbst | das Betriebssystem wählt nach der Route |
| **IPv4 und IPv6** | **gemessen** über beide Rückschleifen (`leitung/tls_test.go`, `leitung/ipv6_test.go`); die Adresse nimmt `[2001:db8::1]:41234` an | **IPv4 nötig**: der Rückfall aus NAT-01 fragt ausdrücklich ohne IPv6, **ein Netz nur mit IPv6 ist nicht unterstützt** (`entdeckung/schnittstellen.go`) | tailcat entscheidet | wie das Betriebssystem |
| **Netzwechsel** | eine offene Leitung reißt ab. **Handel**: die Nachricht bleibt im Ausgang (`flaechen/netzfaelle_test.go`, `TestBeiAusfallBleibtDerHandelImAusgang`), eine abgerissene Reihe nennt, was durchkam (`TestEinAbrissMittendrinNenntDieZahl`). **Tisch**: der Gast kehrt über Ereignisse oder Abzug zurück, eine krumme Rückkehr hält an (`service/partiegastrueckkehr_test.go`). Der Wechsel selbst **offen** ([`HANDPRUEFUNGEN.md`](HANDPRUEFUNGEN.md); #1871 ist geschlossen) | eine Suche ist ein Aufruf und fragt beim nächsten neu; eine laufende **Ankündigung** über eine verschwundene Schnittstelle ist **ungeprüft** | **ungeprüft** | die nächste Anfrage geht über die neue Route |

### Was nach einem Abriss gilt — AK-04

„Sessions überleben transportseitige Unterbrechungen entsprechend ihrem
Fachvertrag; Timeout, Wiederverbindung, Doppelnachrichten und geordneter
Abbau hinterlassen keinen still abweichenden Spiel- oder Handelsstand."

**Die Verträge sind zwei, und keiner verlässt sich auf die Leitung:**

- **Der Handel legt ab, bevor er schickt**, und räumt erst mit der
  Bestätigung von drüben. Eine Doppelnachricht wirkt einmal
  (`service/handelsstrom_test.go`, `TestVerlustDuplikateUndUmordnung`;
  `service/handelsleitung_test.go`, `TestDerselbeVorschlagZweimalIstEinmal`;
  `service/handelsverbuchung_test.go`, `TestZweimalVerbuchenBuchtEinmal`),
  eine überholte schreibt nichts ([0152](entscheidungen/0152-was-geschrieben-ist-bleibt.md)),
  und ein Widerspruch führt in die Klärung.
- **Der Tisch zählt nach der Partie und nicht nach der Leitung**: ein
  doppelt zugestelltes Ereignis wirkt einmal
  (`service/partieempfang_test.go`), eine Lücke führt zur Wiederaufnahme
  (`service/partiestoerung_test.go`), und falsche Hashes, alte Epochen oder
  krumme Folgen halten an, statt weiterzuspielen.
- **Fristen** enden, statt zu hängen (`leitung/leitung_test.go`,
  `TestEineAbgelaufeneFristLaesstNichtsWarten`), und ein Ende schließt beide
  Richtungen (`leitung/schleifenpuffer_test.go`).

**Eine Nachlieferung ohne Zutun gibt es nicht, und das ist entschieden**
(Befund N4 des Nachlaufs über Phase 6.8 hatte sie hierher verwiesen). Zwei
Gründe:

1. **Sie hätte kein Ende.** Weggeräumt wird eine Nachricht erst, wenn ein
   Mensch ihre Ankunft bestätigt (`HandelPostbestaetigung`); eine fachliche
   Quittung auf der Leitung gibt es nicht. Eine Wiederholung im Takt trüge
   dieselben Nachrichten also so lange zu einem fremden Rechner, bis jemand
   klickt — harmlos für den Stand, denn eine Doppelnachricht wirkt einmal,
   aber Verkehr ohne Ertrag.
2. **Von selbst geht in diesem Programm nichts hinaus** — nicht die Suche,
   nicht die Ankündigung, nicht das Horchen. Ein Takt, der einen fremden
   Rechner anwählt, wäre die erste Ausnahme.

**Nachgeliefert wird deshalb auf Aufruf** — „Jetzt übertragen" im Fenster
und `mse handelszustellung` —, und der Ausgang zeigt, was noch aussteht.
**Seit P13 überträgt auch ein Zug einmal**, an seinem Ende
(#2381,
[0200](entscheidungen/0200-ein-zug-je-entscheidung.md)): der Klick, der ihn
auslöst, ist der Aufruf. Scheitert es, bleibt die Nachricht im Ausgang, und
„Jetzt übertragen" steht am Zug — ein zweiter Versuch ohne Zutun kommt
nicht.
**Was das kostet**: wer offline war, bekommt ein Angebot erst, wenn der
Absender überträgt. Eine fachliche Quittung, die das änderte, wäre eine
Änderung des Drahtformats (`mse4mac/trade/1`) und damit eine eigene
Entscheidung.


### Gemessen: Wiederverbindung und Abgleich — D01

**D01 (#2232).**
Die Verträge oben sagen, **dass** nichts still abweicht; hier steht, **wie
schnell** beide Enden nach einem Ausfall wieder dasselbe sehen. Zwischen Gast
und Wirt liegt ein Relais, das ein Netz spielt, das stockt (anhalten) oder
abreißt (kappen) — **unterhalb** von TLS und Rahmung.

**Startereignis** ist die Freigabe des Relais, beim Abriss samt Rückkehr des
Gastes (`PartieBeitreten`) beziehungsweise Zustellung (`mse
handelszustellung`); **Endereignis** ist am Tisch das erste Tischbild beim
Gast, das der Sicht **gleicht**, die der Wirt für dessen Sitz rechnet — die
ganze Partie samt Fassung —, beim Handel der Vorgang bei Anna. **Frist: 10 s.**

| Fall | Weg | gemessen (n = 10) | dazu gehalten |
| --- | --- | ---: | --- |
| kurzer Ausfall — 2 s angehalten, beide ziehen | Tisch, TCP über das Relais | 22–24 ms | das Relais hielt wirklich zurück; der Verlauf wuchs um genau die zwei Befehle |
| langer Ausfall — gekappt, der Wirt zieht weiter | Tisch, Rückkehr | 23–28 ms | der Gast pausiert; ein Wunsch ins Leere **scheitert laut** und wirkt nach der Rückkehr nicht |
| veralteter Gaststand — eine offen gesehene Karte geht auf die Hand, dazu mehr Züge, als eine Sitzung nachreicht | Tisch, Rückkehr | 22–24 ms | beim Gast ist sie danach **geheim**, ohne Karte, Name und Bild |
| Handel nach einem Ausfall | zwei `mse`-Prozesse über den Vermittler | 127 ms (n = 1) | beide `handel stand` gleich bis auf die eigene Sicht, mit demselben **Abdruck**; eine zweite Zustellung ändert drüben nichts |

Gemessen am 2026-10-09 auf Mac17,3 (macOS 27.0.1), auf der Rückschleife,
**neben** einer laufenden Hüllenprobe; die Zahlen enthalten den Abfragetakt
der Messung (20 ms am Tisch, 100 ms beim Handel). Die Läufe:
`internal/app/wiederverbindung_intern_test.go`,
`internal/flaechen/wiederverbindunghandel_test.go`.

**Die negative Gegenprobe**: ohne Rückkehr des Gastes gleichen sich die
Stände nicht, und der Lauf scheitert an der Frist von 10 s.

**Die Rückkehr geschieht auf Aufruf, nicht von selbst** — aus demselben
Grund wie die Zustellung oben: von selbst geht in diesem Programm nichts
hinaus. Gemessen ist deshalb ab dem Aufruf; wie lange ein Mensch braucht, um
nach dem Hinweis „Die Verbindung zum Wirt ist abgerissen" wieder „Mitspielen"
zu drücken, gehört nicht dazu.

**Was das nicht belegt**: zwei Rechner, Verlust und Umordnung **unterhalb**
von TCP, Schlafen und Aufwachen, ein Relais über DERP — das sind D03 und D04.
---

## Zwei Macs auf einem — gemessen am 2026-10-01

**Tragen Verbindung und Leistung für Spielen und Handeln, wenn zwei
Anwendungen einander über das Netz erreichen?** Gemessen zum Abschluss von
Phase 6.7 (#1668)
auf **einem** Mac (Apple M5, 32 GB, macOS 27.0), zwei Läufe hinter dem Build-Tag
`netzlauf`:

```bash
go test -tags netzlauf ./internal/app/ -run ZweiMacs -v -count=1
go test -tags netzlauf ./internal/flaechen/ -run ZweiMacs -v -count=1
```

**Was dabei wirklich läuft**: zwei Anwendungen mit je eigenem Bestand,
eigener Identität und eigenen Kontakten; gefunden über **echtes mDNS**,
verbunden über Horchposten, TCP, TLS 1.3 festgenagelt auf die Kennung und den
Handschlag — keine eingesetzte Leitung, kein eingesetzter Wähler. Der Handel
läuft über **getrennte Prozesse**: Anna horcht in `mse horchen`, Bert stellt
mit `mse handelszustellung` zu. Die Laufzeit kommt von einem Vermittler, der
jedes Stück je Richtung um 10, 40 und 100 ms verzögert, ohne den Fluss
anzuhalten — ein Netz im Haus, eines im selben Land, eines über einen Ozean.

### Spielen (Median, n = 20 je Zeile; Aufbau n = 10)

| Weg | Aufbau (Wählen, TLS, Handschlag) | Beitritt | Satz am Tisch (eine Richtung) | Zug (Wunsch hin, Tischbild zurück) |
| --- | ---: | ---: | ---: | ---: |
| Schleife | 1,8 ms | 1,4 ms | 0,2 ms | 1,2 ms |
| `en0` | 2,2 ms | 3,6 ms | 0,1 ms | 1,0 ms |
| `en0` + 10 ms je Richtung | 47,6 ms | 50,3 ms | 11,6 ms | 26,7 ms |
| `en0` + 40 ms je Richtung | 171,8 ms | 172,0 ms | 43,2 ms | 88,7 ms |
| `en0` + 100 ms je Richtung | 412,0 ms | 413,8 ms | 102,6 ms | 208,5 ms |

**Gelesen heißt das**: ein Satz kostet eine Laufzeit, ein Zug eine
Umlaufzeit und wenige Millisekunden, der Aufbau gut zwei Umlaufzeiten. Das
p95 liegt in jeder Zeile höchstens 3 ms über dem Median. **Gefunden wurde
über mDNS in jedem der drei Versuche** — und zwar nach 5,0 s, denn eine
Suche dauert ihre Frist (`entdeckung.Hoechstfrist`) und meldet danach alle.

### Handeln (Median, n = 5; netto = Zustellung − Programmstart ohne Weg)

| Weg | Zustellung gesamt | Programmstart | Leitung netto |
| --- | ---: | ---: | ---: |
| Schleife | 102,1 ms | 18,9 ms | 82,9 ms |
| + 10 ms je Stück | 152,8 ms | 19,0 ms | 134,2 ms |
| + 40 ms je Stück | 286,9 ms | 18,6 ms | 267,7 ms |
| + 100 ms je Stück | 591,7 ms | 18,5 ms | 572,9 ms |

**Ein Angebot über einen Ozean ist in gut einer halben Sekunde drüben**;
die Leitung kostet rund 2,5 Umlaufzeiten und einen festen Anteil von gut
80 ms auf beiden Seiten, der nicht mit der Laufzeit wächst. Jede Runde trug
**genau eine** Nachricht, und bei Anna stand danach der Vorgang.

### Was der Lauf gefunden hat — zwei Fehler, die kein Test sah

1. **Wer aufstand, kam nicht wieder an den Tisch.** Der Wirt schloss eine
   beendete Partiesitzung nie, der Gast nicht nach einem Abriss; eine tote
   Sitzung sperrte jede neue zum selben Gegenüber bis zum Profilwechsel —
   „läuft schon eine Sitzung für partie". Damit ging auch die Rückkehr nach
   einem Abriss nie, für die es `Pausiert` gibt.
2. **An einen laufenden Tisch konnte sich niemand setzen.** Der Wirt reichte
   Antwort, jedes fehlende Ereignis und das Bild in einem Zug nach und leerte
   erst danach; die Schlange fasst `app.Sitzungstiefe` = 8, und ab gut sechs
   Ereignissen endete die Sitzung mit „die Warteschlange dieser Sitzung ist
   voll".

**Beide waren bei grüner Pipeline und voller Abdeckung da**: die Läufe über
beide Enden setzten sich an einen frischen Tisch und standen nie wieder auf.
Behoben mit roten Läufen an Dienst und Anwendung
(`internal/app/partiewiederkehr_intern_test.go`,
`internal/service/partiewirtschub_test.go`).

### Was er **nicht** belegt

**Stand dieser Liste ist der 2026-10-01**, der Tag des Laufs. Was seither
gelaufen ist, steht bei dem Punkt, den es betrifft.

- **Ein Netz.** macOS stellt ein Paket an die eigene Adresse über die
  Rückschleife zu: der Weg `en0` misst dasselbe wie die Schleife, nur mit der
  Adresse, die ein anderer Mac wählen würde. **Verlust, Engpass, Router, NAT,
  WLAN und Firewall** misst er nicht; die Laufzeit kommt vom Vermittler.
- **Zwei Bildschirme, zwei Menschen, VoiceOver** — die Abnahme am nativen
  Fenster mit zwei Geräten stand in
  #1871
  (geschlossen am 2026-10-01); was sie offen ließ, führt
  [`HANDPRUEFUNGEN.md`](HANDPRUEFUNGEN.md).
- **Die Fernleitung.** tailcat geht über DERP und damit über einen fremden
  Dienst; ein Lauf dorthin ist ein Lauf gegen das Internet und steht nicht
  hier. **Seither gelaufen**: auf einem Rechner über echtes DERP (SEC-21,
  2026-10-07, [`belege/sec21-fernweg.md`](belege/sec21-fernweg.md)), zwischen
  zwei VM-Gästen im selben Netz, gewöhnlich und mit erzwungenem Relais
  (2026-10-07, [`belege/sec22-vmlauf.md`](belege/sec22-vmlauf.md)), und auf
  einem Mac Spiel und Handel über den Fernweg (D04, 2026-10-10, oben).
- **Eine zweite VM.** Eine macOS-VM (Lume/CUA) hätte einen zweiten
  Netzstapel gebracht, aber ein fremdes Installationsskript und ein Abbild
  von vielen Gigabyte verlangt — das ist eine Entscheidung des Menschen vor
  dem Rechner und nicht dieser Nacht. **Seither gefallen**: am 2026-10-07 liefen
  zwei Lume-Gäste im selben Netz ([`belege/sec22-vmlauf.md`](belege/sec22-vmlauf.md),
  [`belege/k02-lokalesnetz.md`](belege/k02-lokalesnetz.md)).

---

## Was hier nicht steht

- **SOCKS-Proxys.** Die Einstellung weist sie ab (`domain.PruefeProxyadresse`);
  aus der Umgebung nimmt `httpproxy` sie, geprüft ist es nicht.
- **Proxy-Authentifizierung für die Dienste.** Ein Kennwort in der
  Einstellung wird abgewiesen, weil es unverschlüsselt im Bestand läge
  (S13.NAT-02); aus der Umgebung schickt Go es als `Proxy-Authorization`
  mit — **geprüft ist das nur für den Weg zum Relais**, nicht für die
  HTTP-Klienten. Gezeigt wird es nirgends (`domain.ProxyadresseOhneKennwort`).
- **QUIC und HTTP/3.** Kein Verbraucher, kein Lauf — und für die Peerleitung
  ausdrücklich abgelehnt ([0113](entscheidungen/0113-kein-libp2p-im-lan.md)).
- **Wiederaufnahme einer TLS-Sitzung.** Der Klient führt keinen
  Sitzungsspeicher, also kommt sie nicht vor; dass `VerifyConnection` sie
  abdeckte, ist damit **nicht gemessen**, sondern aus der Fläche abgeleitet.
- **Eine Peerleitung über ein echtes Netz.** Alles hier läuft über
  `127.0.0.1` — bis auf die Läufe zwischen zwei VM-Gästen im selben Netz vom
  2026-10-07 ([`belege/sec22-vmlauf.md`](belege/sec22-vmlauf.md)) und die
  Probe zwischen Wirt und Abnahme-VM vom 2026-10-01
  ([`HANDPRUEFUNGEN.md`](HANDPRUEFUNGEN.md)). Die Firewall, isolierte
  Gäste-WLANs, Schlafen und Aufwachen und der Netzwechsel bleiben
  Handprüfungen (#1455).
- **Eine zweite Maschine.** Alles hier ist auf einem Rechner gemessen
  ([0007](entscheidungen/0007-kein-entfernter-lauf.md)) — die zwei VM-Gäste auf
  demselben Mac vom 2026-10-07 ausgenommen; zwei physische Rechner fehlen. **Für die Bindung
  jenseits der Schleife ist das keine Lücke**, und der Grund steht im Kopf von
  `internal/flaechen/dienstbindung_test.go`: eine Anfrage über die **eigene**
  LAN-Adresse geht durch denselben Schnittstellenstapel wie die eines fremden
  Rechners. Was ungemessen bleibt, ist das **Netz dazwischen** — Firewall,
  Gäste-WLAN, Routen.
- **Ob ein Gateway tut, was es verspricht.** Dass die Adresse einstellbar ist
  und der Schlüssel daran hängt, ist geprüft; dass hinter
  `https://gateway.firma.intern` wirklich ein OpenAI-verträglicher Dienst
  steht, sagt niemand hier.
