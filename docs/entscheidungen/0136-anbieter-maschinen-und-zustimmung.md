# 0136 — Anbieter, Maschinen und Zustimmung

**Stand:** entschieden am 2026-09-27 · ID-05
(#1704)

## Was dieser Vermerk tut

**Er bindet Zusagen an Belege und schließt eine Lücke.** Was ID-05 verlangt,
galt zum größten Teil schon — verteilt über
[0078](0078-der-wirksame-akteur.md) (der wirksame Akteur und der Nachweis),
[0080](0080-maschinenzugaenge-und-widerrufbare-sitzungen.md)
(Maschinenzugänge) und [0082](0082-der-dienstkatalog.md) (der
Dienstkatalog). **Die Lücke**: für die vier Anmeldeanbieter stand nirgends,
wogegen ihr Weg geprüft ist. Jetzt steht es je Anbieter an der Nutzerseite —
und die ehrliche Antwort ist bei allen vieren *nur gegen eine Nachbildung*.
Dieselbe Form wie [0128](0128-identitaet-bleibt-lokal.md) für ID-01.

## AK-01 — jeder Anbieter sagt, wie weit er geprüft ist

| Anbieter | Was dasteht | Beleg |
| --- | --- | --- |
| **Anmeldeanbieter** — GitHub, Google, Microsoft, X | Nutzerseite › *Konto*: je Anbieter *eingerichtet* — der Anmeldeweg **kann beginnen**: eine `client_id` im Bestand und, wo der Anbieter es verlangt, ein Geheimnis im Bund (GitHub), wo er es verbietet, keines (Microsoft) — oder *nicht eingerichtet*, dazu *nur gegen eine Nachbildung geprüft*. **Neu mit diesem Vermerk** (`domain.AnbieterPruefstand`, `AnbieterAntwort.Staende`). *Nur gegen eine Nachbildung* ist für **jeden** der vier ein durchlaufener Weg — beginnen, tauschen, Identität lesen — gegen einen Server im Test, der wie sein Anbieter antwortet; ein neuer Anbieter ohne Nachbildung macht den Lauf rot. Ein Lauf gegen den echten Anbieter braucht ein eigenes Klientenkonto, und dieses Projekt liefert keines mit ([0066](0066-nutzer-profile-und-die-vier-entscheidungen.md)); wer einen macht, trägt ihn in `anbieterlaeufe` ein | `TestJederAnbieterLaeuftGegenEineNachbildung`, `TestStartbereitIstDiePruefungDesBeginnens`, `TestJederAnmeldeanbieterNenntWogegenErGeprueftIst`, `TestJederAnbieterSagtObErEingerichtetUndWieErGeprueftIst`, `anmeldung.test.tsx` |
| **Daten- und KI-Anbieter** — Scryfall, magicthegathering.io, MTGJSON, Scrydex, CardTrader, Anthropic, OpenAI, xAI | Einstellungen › *Dienste*: je Fähigkeit der Status aus `domain.Dienstlage` (aus, Zugang fehlt, bereit …) und der **Nachweis** — *ungeprüft* oder *geprüft am … mit/nicht mit dieser Fassung* (0082). Das ist dieselbe Dreiheit unter anderen Wörtern: eingerichtet, nur gegen Beispiele, gegen den echten Dienst | `TestDerKatalogNenntJedeFaehigkeitMitStatusUndVertrag`, `TestEinNachweisNenntLaufDatumUndFassung` |
| **Ollama** | Kein fremder Dienst: er rechnet auf diesem Rechner ([0098](0098-was-ollama-belegt-und-was-nicht.md)); was ein Modell kann, wird gemessen und nicht erraten | 0098 |

**Kein zentrales Konto, kein Profil über eine gleiche Adresse.** Ein
Anbieterkonto hängt an **Anbieter und Subjekt**; `domain.Fremdkennung` trägt
gar keine Adresse. Wer sich über GitHub mit einem Anzeigenamen anmeldet, der
die Adresse eines Nutzers dieses Rechners ist, bekommt einen **neuen**
Nutzer — nicht den mit der Adresse, nicht dessen Profil, und jener trägt
danach kein fremdes Konto (`TestEineGleicheAdresseUebernimmtKeinKonto`). Ein
zentrales Konto dieses Programms gibt es nicht (0128).

## AK-02 — Maschinenzugänge: was entschieden ist

**Die Entscheidung steht in 0080, und sie gilt weiter.** Hier steht, welche
Frage aus ID-05 sie beantwortet:

| Frage aus ID-05 | Antwort | Sitz |
| --- | --- | --- |
| Serviceidentität | ein Token mit **Principal** (wer dahintersteht, Mensch oder Maschine) und **Bezeichnung** (wofür), an **ein** Profil und wahlweise einen Sitz gebunden; das Profil wechselt es nie | 0080; `service.Tokenverwaltung.Aendern` |
| Rotation | am selben Datensatz, das alte Geheimnis gilt eine Stunde, eine zweite Rotation wirft es weg, ein Widerruf sperrt beide | 0080, `domain.Tokenuebergang`, `internal/flaechen/tokenrotation_test.go` |
| Ablauf | ein Tag, dreißig Tage, ein Jahr, nie — »nie« steht offen da | 0080 |
| Minimale Befugnis | `lesen` reicht bis `lokalLesen`, `schreiben` bis `lokalSchreiben` — **nie nach draußen**, nie über die Rolle hinaus, und kein Token stellt einen Nachweis aus | `domain.Akteur.Darf`, `internal/domain/akteur_test.go` |
| Erstfreigabe | **Eine interaktive Erstfreigabe am Fenster gibt es nicht, mit Grund**: angelegt wird ein Token von jemandem mit der Rolle `mitarbeiter` am **Zielprofil** (#1817), und kein Token reicht über `lokalSchreiben` hinaus. Was darüber geht — nach draußen, verbindlich —, verlangt einen Nachweis, und den stellt nur das Fenster aus (AK-03). Eine Freigabe davor wäre eine zweite Tür vor einem Raum, in den ein Token ohnehin nicht kommt | [`AUTORISIERUNG.md`](../AUTORISIERUNG.md), Zeile *Zugriffstoken … führen* |
| Keine stille Erweiterung | den Umfang eines Tokens ändert nur `TokenAendern` — ein ausdrücklicher Schritt eines Mitarbeiters am Profil, an jeder Fläche sichtbar, und nie über `lokalSchreiben` hinaus. **Profil und Sitz wechseln dabei nie** | `service.Tokenverwaltung.Aendern` |
| Widerruf wirkt sofort | der Akteur entsteht je Ruf; ein Entzug ist beim nächsten wirksam — am Fenster, am Werkzeugserver, am Dienst | `TestEinEntzugWirktBeimNaechstenAufruf` (`internal/app/entzug_test.go`), `TestDerServerFragtJedenSchreibendenRufNeu` (`cmd/mse-mcp/entzug_test.go`), `TestDerDienstFragtJedeSchreibendeAnfrageNeu` (`cmd/mse-rest/entzug_test.go`) — die beiden letzten wechseln zwischen zwei Rufen die Rolle im Bestand und prüfen **an der Wirkung**, dass im fremden Profil nichts entstand |

## AK-03 — Zustimmung ist ein Nachweis, kein Wahrheitswert

**Das war schon 0078, und es war ausdrücklich gegen ein `bestaetigt`-Feld
entschieden** — „das beweist, dass er das Feld kennt". Ein `domain.Nachweis`
bindet **Nutzer, Profil, Zweck, Ziel und Revision** (und wo Geld im Spiel ist
eine Grenze), gilt höchstens 24 Stunden, wird **nur am Fenster** von einem
Menschen ohne Token ausgestellt und **genau einmal** eingelöst. Ein
wiederverwendeter, abgelaufener oder auf ein anderes Ziel, einen anderen
Zweck, eine andere Revision oder einen anderen Menschen gebundener Nachweis
erlaubt nichts — `internal/domain/nachweis_test.go`,
`internal/service/nachweise_test.go`. **`externSchreiben` und `verbindlich`**
verlangen ihn zusätzlich zur Rolle `verwalter`.

## AK-04 — was nach einem Entzug, einer Trennung, einem Fehler bleibt

| Fall | Was gilt | Beleg |
| --- | --- | --- |
| Token widerrufen | beim nächsten Ruf keine Wirkung mehr, auch mitten in einer Sitzung | siehe AK-02 |
| Anbieter getrennt | `KontoLoesen` nimmt die Verbindung; der Nutzer bleibt und mit ihm sein Profil | `KontoLoesen` an allen Flächen (`docs/FAEHIGKEITEN.md`) |
| Anbieter antwortet falsch | jeder Fehler des Anmeldewegs reicht durch und wird gemeldet, statt still einen Nutzer anzunehmen | `TestJederFehlerwegDerAnmeldungReichtDurch` |
| Lokale Wiederherstellung | nur über den Zettel, sie setzt nur das Kennwort (ID-02) | `TestDieWiederherstellungSetztNurDasKennwort` |
| Geheimnisse | Ein Token liegt nur **gehütet** im Bestand, sein Geheimnis erscheint einmal beim Anlegen oder Rotieren; Anbietergeheimnisse und Zugänge liegen im Schlüsselbund und gehen in keine Sicherung; das Protokoll kennt nur angemeldete Merkmale | [`ADMINISTRATION.md`](../ADMINISTRATION.md) § *Sicherung*, `internal/werkzeuge/protokollmerkmale_test.go` |

## Was die Belege nicht halten

**Ob ein Anmeldeanbieter heute noch so antwortet wie seine Nachbildung.** Das
sagt nur ein Lauf gegen ihn — und genau das steht jetzt an der Oberfläche,
statt dass der Knopf es verschweigt.

## Nachlauf am 2026-09-27

**Drei Befunde galten, einer nicht.** *Simuliert* war für Microsoft nicht
belegt — es gab nur einen Test, der ein Geheimnis abweist; jetzt läuft jeder
Anbieter den ganzen Weg. *Eingerichtet* hieß GitHub schon mit einer
`client_id`, obwohl der Weg ohne Geheimnis nicht beginnt; jetzt fragt die
Liste dieselbe Prüfung wie der Beginn (`anmeldung.Anbieter.Startbereit`).
**Die zwei Serverbelege gab es** — in `cmd/mse-mcp` und `cmd/mse-rest`, nicht
unter `internal/`, wo der Nachlauf suchte; die Tabelle nennt jetzt ihre
Dateien.

## Verworfen

| Verworfen | Weil |
| --- | --- |
| Eine Zustimmung per `bestaetigt: true` an REST oder MCP | 0078: es beweist nur, dass der Rufer das Feld kennt |
| Einen Anmeldeanbieter „geprüft" nennen, weil sein Protokoll der Spezifikation folgt | Die Nachbildung prüft **unseren** Lauf, nicht den des Anbieters — dieselbe Trennung wie der Nachweis im Dienstkatalog |
