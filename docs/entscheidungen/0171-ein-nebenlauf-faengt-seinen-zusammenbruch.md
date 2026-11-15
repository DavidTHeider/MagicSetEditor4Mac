# 0171 — Ein Nebenlauf fängt seinen Zusammenbruch selbst

**Stand:** angenommen am 2026-10-06 · BK-05
(#2111,
#2112,
#2113,
#2114).

## Die Frage

**Eine Panik in einer eigenen Goroutine fängt niemand.** Wails fängt sie im
Ruf einer Bindung, `net/http` in seinem Behandler — in einer Goroutine, die
ein Ruf selbst startet, fängt sie keiner, und der Prozess endet: mit ihm jeder
Tisch, jeder Dienst und jede offene Bearbeitung. Gemessen am 2026-10-06, je
mit einer künstlichen Panik und vor der Änderung:

| Nebenlauf | Wirkung vorher |
| --- | --- |
| Werkzeugruf über stdio (`mcp.Server.Bediene`) | der Prozess endet |
| Werkzeugruf über HTTP (`mcp.UeberHTTP`) | `net/http` fängt sie — **keine** Antwort, der Ruf bleibt als laufend stehen |
| Suche je Schnittstelle (`entdeckung.ueberAlle`) | der Prozess endet |
| Ziel eines KI-Stapels (`app.batchFahren`) | der Prozess endet |

## Die Entscheidung

**Wer einen Nebenlauf startet, fängt seinen Zusammenbruch** — mit
`defer domain.Auffangen(…)`, das daraus `domain.ErrZusammengebrochen` macht.
Ein geworfener Fehler bleibt dabei erkennbar (`errors.Is` geht durch beide).
Was danach geschieht, entscheidet der Starter, und es ist dasselbe wie bei
einem gewöhnlichen Fehler an dieser Stelle:

| Nebenlauf | Nach dem Zusammenbruch | Frei wird |
| --- | --- | --- |
| MCP-Werkzeugruf | **genau eine** Fehlerantwort `-32603` mit „zusammengebrochen", wenn der Ruf nicht abgebrochen war; der Grund steht im Protokoll | der Eintrag in der Liste laufender Rufe |
| mDNS je Schnittstelle | ein Fehler **dieser** Schnittstelle; gescheitert ist die Suche erst, wenn jede scheitert | `WaitGroup` |
| KI-Stapelziel | ein gescheitertes Ziel ohne Vorschlag, `unbekannt` mit Vorgangskennung — der Grund steht unter ihr im Protokoll; die anderen Ziele fahren weiter, der Stapel bleibt **ein** Posten | Schranke und `WaitGroup` |
| REST-Auftrag | schon vorher gefangen (`rest.auftraege.go`) | — |

**Der Grund geht nicht an das Gegenüber**: eine Panik kann sagen, was im
Bestand steht. Er geht ins Protokoll, und die Server schreiben über den
schwärzenden Schreiber (`app.Serverprotokoll`).

## Und die Nebenläufe von Partie und Peer (BK-05a)

**Dort genügt das Fangen nicht** — frei werden muss, was der Lauf hielt, und
ein Stand, der danach „läuft" sagte, wäre ein stilles Weiterlaufen. Gemessen
je mit einer künstlichen Panik, vorher jedes Mal das Ende des Prozesses:

| Nebenlauf | Ausgelöst über | Nach dem Zusammenbruch | Frei wird |
| --- | --- | --- | --- |
| Handschlag am Horchposten, **vor** der Übergabe (`service.Horchposten.bedieneEine`) | eine Leitung, die beim Lesen zusammenbricht | gemeldet; 33 Zusammenbrüche hintereinander, danach entsteht eine gewöhnliche Sitzung | der Platz unter `HoechstensHandschlaege` (Gegenprobe ohne Freigabe: die nächste Sitzung wird abgewiesen), die Verbindung |
| dieselbe Stelle, **nach** der Übergabe (die Aufnahme) | eine Aufnahme, die zusammenbricht | gemeldet; die Sitzung geht zu, statt als Leiche in der Übersicht zu stehen | Platz und Sitzung |
| Tisch des Gastes (`amTischSitzen`, `partieErfragen`, `rueckkehrMelden`) | das erste Tischbild am Meldeweg des Fensters | wie eine abgerissene Leitung: Sitzung zu, Beitritt **pausiert**, der Platz gehört dem Gast weiter | die Sitzung — ein Fang für alle drei (`amFremdenTisch`) |
| Sitzung eines Gastes beim Wirt (`partieBedienen`) | der Zug des Gastes am Meldeweg des Wirts | nur diese Sitzung endet; der Tisch des Wirts bleibt offen | Platz in der Runde, Besetzung, Sitzung |
| Horchen und Fernhorchen (die Goroutine um `Bediene`) | ein Lauscher, der in `Nimm` zusammenbricht | der Stand sagt „läuft nicht", die Ankündigung im Netz bzw. die Einladung geht mit; ein neuer Start gelingt | Lauscher, Kontext, Stand |
| Ansage eines Dienstes (`leseAnsage`) | ein Leser, der zusammenbricht | ein Fehler des Starts | — |

**Was keinen Fang bekam, und warum**: das Warten auf das Kind eines
Dienstes (`befehl.Wait` in `dienstwerk.go`) — kein fremder Code, der
zusammenbrechen könnte, und ein Fang, den kein Lauf je erreicht, wäre eine
ungeprüfte Zusage.

**Der Grund steht im Protokoll** — über `zusammenbruchProtokollieren` mit der
Vorgangskennung des Laufs, am Horchposten auf der Stufe `Error`; die
Prozesse schreiben über den schwärzenden Schreiber.

## Verworfen

- **Eine Panik weiterwerfen, nachdem aufgeräumt ist.** Dann endete der
  Prozess trotzdem — sauberer, aber mit demselben Ergebnis für den, der
  gerade etwas anderes tat.
- **Ein eigener Fehlerschlüssel „zusammengebrochen".** Der Rückfall
  `unbekanntMitVorgang` sagt schon, was zu tun ist, und führt mit der
  Kennung zum Grund im Protokoll (#1601); ein zweiter Satz für denselben
  Rat wäre eine Übersetzung mehr in acht Sprachen ohne neue Auskunft.
