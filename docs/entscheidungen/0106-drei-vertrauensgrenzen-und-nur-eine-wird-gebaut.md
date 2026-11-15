# 0106 — Drei Vertrauensgrenzen, und nur eine wird gebaut

**Datum:** 2026-09-13 · **Stand:** entschieden ·
**Anlass:** Werkstück A14 von Phase 5.3
(#1453),
Punkt „Architektur-Gate für Remote-MCP"

## Die Frage

Das Issue nennt sie als Entscheidungspunkt und verlangt, sie **vor** dem Bauen
zu klären:

> Lokaler MSE-Server für Agenten, MSE als Client fremder Tools und
> providerseitiger Remote-MCP-Zugriff sind drei unterschiedliche
> Vertrauensgrenzen. Aus Anhang entsteht keine Pflicht, einen öffentlichen
> MSE-MCP-Server/Cloudtunnel zu betreiben.

Sie sehen sich ähnlich — dreimal „MCP" —, und genau daran erkennt man die
Falle: wer sie als **eine** Frage behandelt, baut die dritte, weil die erste
schon da ist.

## Die drei, ausgeschrieben

| | Wer ruft | Wer antwortet | Wem muss man glauben |
| --- | --- | --- | --- |
| **A — lokaler Server** | ein Agent auf **diesem** Rechner | `mse-mcp` über stdio | dem Agenten, den ein Mensch selbst gestartet hat |
| **B — MSE als Klient** | diese Anwendung | ein **fremder** MCP-Server | einem Dienst, den jemand anders betreibt |
| **C — Remote-Server** | irgendwer aus dem Netz | `mse-mcp` über HTTP | **jedem**, der die Adresse kennt |

## Was gilt

**Gebaut ist A, und nur A.** `cmd/mse-mcp` liest `os.Stdin` und schreibt
`os.Stdout`; im Paket steht nichts, was lauscht — kein `net.Listen`, kein
`http.Server`, nachgeprüft am 2026-09-13. Der Grund steht seit
[0010](0010-mcp-als-eigener-prozess.md) da: „Über stdio gibt es keinen Port,
keine Adresse und nichts, was auf etwas wartet."

**B ist abgegrenzt und nicht gebaut.** Diese Anwendung ruft heute keinen
fremden MCP-Server. Käme er, verlangte das Issue eine ganze Liste, die hier
nicht halb entstehen soll: Positivliste der Werkzeuge, endpunktgebundene
Zugangsdaten, Nutzlastvorschau mit Freigabe, ein Vertrag über SSRF, Herkunft,
Transport und Anmeldung, **Werkzeugergebnisse als unvertrauenswürdig**,
begrenzte Orchestrierung und begrenzte Werkzeugschleifen.

**C wird nicht gebaut, und zwar nicht „noch nicht".** Ein öffentlich
erreichbarer MSE-MCP-Server hieße: ein Dienst auf dem Schreibtisch eines
Menschen, der Kartenbestände, Sammlungswerte und Zugangsdaten führt, wartet auf
Verbindungen aus dem Netz. Dieselbe Überlegung hat in
[0096](0096-welche-optionalen-datenquellen-zugaenglich-sind.md) schon die
Scrydex-Webhooks abgelehnt — sie bräuchten „einen von außen erreichbaren Server
auf dem Schreibtisch".

**Und aus dem Anhang entsteht keine Pflicht.** Dass ein Anbieter Remote-MCP
kann, ist eine Auskunft über ihn und kein Umfang für dieses Programm.

## Was das kostet

**Ein Agent auf einem anderen Rechner erreicht diese Anwendung nicht.** Wer das
will, braucht einen Tunnel, den er selbst betreibt und verantwortet — und dann
steht zwischen seinem Bestand und dem Netz nur, was er dort einrichtet. Das ist
eine bewusste Zumutung und kein Versehen.

**Und `mse-rest` ist kein Gegenbeweis.** Es lauscht auf der Schleifenadresse,
prüft Herkunft und Marke und ist damit dieselbe Vertrauensgrenze wie A: ein
Programm auf **diesem** Rechner. Wer es ins Netz stellt, tut etwas, das dieses
Programm nirgends vorschlägt.

## Wie man das Gegenteil erkennt

**An einem Listener.** Ein `net.Listen` oder `http.Server` in `internal/mcp`
oder `cmd/mse-mcp` wäre C, auch wenn es „nur localhost" hieße: aus `localhost`
wird `0.0.0.0` durch eine Zeile. Der Wächter über die Netzgrenze
(`internal/schichten/netz_test.go`) führt, wer **hinaus** darf; wer **herein**
lässt, ist diese Entscheidung — und ein Lauf hält sie
(`internal/mcp/keinlauscher_test.go`).

## Verworfen

**„Erst einmal auf localhost, das ist ja sicher."** Es ist die Gestalt, in der
C entsteht: der Port ist da, die Bindung ist eine Zeile, und die Prüfung, die
dann fehlt, fehlt an jeder einzelnen Stelle.

**„Nur lesend."** Ein lesender Zugriff auf einen Kartenbestand ist ein Zugriff
auf Namen, Sammlungswerte und Lagerorte. Die Unterscheidung hilft beim
Schreiben und nicht beim Preisgeben.
