# 0039 — Kein Dienst auf der Rückschleife

> **Umgekehrt am 2026-08-23 durch
> [0046](0046-doch-ein-dienst-auf-der-rueckschleife.md).** Der Fall, den der
> Abschnitt „Was das nicht heißt" vorgesehen hat, ist eingetreten: es gibt
> einen benannten Abnehmer. Dieser Vermerk bleibt vollständig stehen — seine
> Rechnung ist die Grundlage, auf der 0046 rechnet, und drei seiner vier
> Punkte gelten unverändert.

**Stand:** überholt, siehe 0046
**Datum:** 2026-08-22
**Schreibt fort:** [0010](0010-mcp-als-eigener-prozess.md)
**Berührt:** [0031](0031-sieben-phasen-bis-zur-freigabe.md) — Phase M

## Lage

Phase M (#591)
heißt „MCP und REST vollständig: Eingabevalidierung, Ratenbegrenzung, Bindung
ausschließlich an `127.0.0.1`, Authentifizierung über JWKS" — und widerspricht
damit zwei Zusagen, die heute dastehen.

**`docs/ROADMAP.md`, „Was ausdrücklich nicht kommt":** kein Netzwerkdienst,
keine Anmeldung, keine Synchronisierung.

**`internal/schichten/netz_test.go`**, der Wächter über die dritte Zusage aus
#54: die
Anwendung lädt von sich aus nichts herunter. Er prüft strukturell über den
Importgraphen; `net/http` erreicht genau ein ausgeliefertes Paket,
`internal/bezug`, und das ist sein Zweck.

**Ein REST-Dienst auf `127.0.0.1` ist ein Netzwerkdienst**, auch wenn ihn
niemand von außen erreicht.

Deshalb verlangt die Phase diesen Vermerk **vor** der ersten Zeile, und
[0031](0031-sieben-phasen-bis-zur-freigabe.md) hält ausdrücklich fest: fällt er
gegen den Dienst, ist die Phase fertig — „ein gültiger Ausgang", der „dieselbe
Deckung wie ein Ja" braucht (§10).

## Entscheidung

**Es kommt kein Dienst auf die Rückschleife.** Das MCP-Gegenüber startet
weiterhin `mse-mcp` als Unterprozess und spricht über stdio; es gibt keinen
Port, keine Adresse und nichts, das wartet.

**Was von Phase M bleibt**, bleibt: die MCP-Werkzeuge werden vollständig, ihre
Eingaben werden validiert, und `netz_test.go` sagt danach, was heute gilt. Die
REST-Kästen entfallen.

## Warum, in vier Punkten

### 1. Es gibt keinen benannten Abnehmer

**Gesucht am 2026-08-22 über `docs/`, die Vermerke und die Roadmap:** außer der
Phasenzeile selbst nennt keine Stelle einen, der REST braucht. Kein Issue, kein
Anwendungsfall, keine Rückmeldung.

Ein Dienst ohne benannten Abnehmer ist eine Angriffsfläche ohne Zweck. Das ist
kein Formalismus: die Fläche muss verteidigt werden, unabhängig davon, ob sie
jemand benutzt.

### 2. [0010](0010-mcp-als-eigener-prozess.md) hat die Rechnung schon aufgemacht

Dort steht, wörtlich, was Phase M bauen würde:

> Ein lauschender Port ist eine Angriffsfläche, die es zu verteidigen gilt —
> Herkunftsprüfung, Zugangsmarken, Bindung an die Rückschleife —, und all das
> für einen Nutzen, den stdio ohne eine Zeile davon liefert.

**Herkunftsprüfung, Zugangsmarken, Rückschleife** sind genau die drei der vier
Testklassen, die #591 je Endpunkt verlangt. Der Vermerk von 2026-08-15 hat sie
benannt und abgelehnt; seither hat sich an der Lage nichts geändert, was diese
Rechnung umdrehte.

### 3. Die Lücke liegt im Umfang, nicht im Protokoll

**Gezählt am 2026-08-22:**

| | Werkzeuge / Bindungen | davon schreibend |
| --- | --- | --- |
| MCP über stdio | **5** (`lesen.sets`, `lesen.karten`, `lesen.felder`, `lesen.suche`, `schreiben.feld`) | **1** |
| Das Fenster | 61 | 29 |

**Ein Sprachmodell kann heute ein Feld setzen und sonst nichts ändern** — keine
Karte anlegen, kopieren, löschen, umordnen; keine Stiloption, keinen Setwert,
keine Ausfuhr.

> **Nachtrag vom 2026-08-23.** Die Zahlen oben sind eine Aufzeichnung ihres
> Tages und bleiben stehen; der Satz davor ist es nicht — er steht im Präsens
> und stimmt nicht mehr. Der Lebenszyklus einer Karte steht seit
> [0040](0040-der-lebenszyklus-einer-karte-gehoert-dazu.md) über MCP:
> **9 Werkzeuge gegen 64 Bindungen**, davon 5 schreibend (am 2026-08-23
> erhoben; hier stand 63, gezählt bevor `AutoErsetzen` dazukam). Der **Schluss** von
> Punkt 3 gilt unverändert — die Lücke liegt im Umfang, nicht im Protokoll —,
> und genau deshalb baut [0043](0043-neun-phasen-bis-0-2-0.md) den Fachkern
> (Phase T) **vor** dem Dienst (Phase U).

Diese Lücke schließt kein zweites Protokoll. Sie schließt sich, indem die
Werkzeuge dazukommen — und die kämen über stdio genauso wie über HTTP, nur ohne
Port, ohne JWKS und ohne Ratenbegrenzung.

**Ein REST-Dienst neben einem unvollständigen MCP wäre eine zweite
unvollständige Fläche.**

### 4. Was ein Ja kostete, ist nicht der Port

Sondern das, was danach dauerhaft gilt: zwei Zusagen der Roadmap fielen, ein
Wächter müsste gelockert werden, und je Endpunkt kämen vier Testklassen dazu —
gültig, Grenze überschritten, fremde Herkunft, abgelaufenes Token. Das ist die
Arbeit **jeder künftigen Änderung** an dieser Fläche, nicht nur der ersten.

## Was das nicht heißt

**Nicht, dass die Entscheidung für immer steht.** Sie steht, solange die vier
Punkte gelten. Käme ein benannter Abnehmer, dem stdio nachweislich nicht reicht
— etwa ein Werkzeug, das den Unterprozess nicht starten kann —, wäre die
Rechnung neu aufzumachen. **Dieser Vermerk ist dann fortzuschreiben, nicht zu
umgehen.**

**Nicht, dass die Schnittstelle unwichtig ist.** Das Gegenteil: Punkt 3 sagt,
dass sie zu schmal ist. Was Phase M an Arbeit behält, ist genau das.

## Was der Wächter danach prüft

`internal/schichten/netz_test.go` prüft weiter, dass `net/http` nur
`internal/bezug` erreicht — **und diese Entscheidung ist der Grund, warum er das
darf.** Ohne sie stünde er als Zusage da, die eine spätere Phase still
gelockert hätte.

Was er **nicht** prüft, steht neben ihm und gehört auch hierher: er sieht den
Importgraphen, nicht das Verhalten. Ein Paket, das `net/http` einbindet und
nichts damit tut, fiele ihm auf; eines, das über ein fremdes Paket ins Netz
ginge, nicht.
