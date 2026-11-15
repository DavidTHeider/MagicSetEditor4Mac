# 0161 — Die Kontingente lassen sich abschalten, und nichts sonst

**Stand:** angenommen am 2026-10-02 · QA-08.AK-03
(#1739),
Marke `new`: MSE2 hat weder einen Dienst noch einen Werkzeugserver.

## Die Frage

Das Kriterium verlangt, dass Nutzende **geeignete Ratenbegrenzungen** für
kontrollierte Umgebungen abschalten dürfen. Authentifizierung, TLS-Pflichten,
Parser-, Rumpf- und Speichergrenzen und grundlegende Fristen sollen dabei
bleiben. Zu entscheiden war, **welche** Grenzen geeignet sind, wo der
Schalter sitzt und wie weit er reicht.

## Was es gibt — nachgelesen am Stand `52a58825`

| Grenze | Sitz | Wogegen |
| --- | --- | --- |
| **Guthaben je Principal und global**, mit Reserve | `domain.Kontingent`, `service.Kontingente` ([0083](0083-ressourcenbudgets-ohne-aussperrung.md)) | wie viel ein Maschinenzugang **in der Zeit** darf |
| Plätze je Ressourcenklasse | `service.Haushalt` | wie viel schwere Arbeit **zugleich** läuft |
| Sperre je Nachweis, Drossel für alle Ratenden | `service.Zugangssperre`, `rest.Drossel` | das Raten einer Marke oder eines Tokens |
| Rumpf und Antwort, 32 MiB; eine MCP-Nachricht, 8 MiB | `internal/rest/dienst.go`, `internal/mcp/ueberhttp.go` | Speicher, den ein Fremder bestimmt |
| Auftragsbuch: vier laufend, sechzehn behalten | `internal/rest/auftraege.go` | Speicher für die Antworten langer Läufe |
| Verbindungen, sechzehn | `internal/rest/gedraenge.go` | Dateigriffe |
| Fristen: Kopf 10 s, Antwort 5 min, Leerlauf 2 min | `cmd/mse-rest/dienst.go`, `cmd/mse-mcp/ueberhttp.go` | eine Verbindung, die ewig hängt |
| Zugang und TLS jenseits der Schleife | `internal/rest`, [0115](0115-die-schleife-ist-eine-vorgabe.md) | wer überhaupt etwas darf |

## Die Entscheidung

1. **Abschaltbar sind genau die Guthaben** — je Principal, global und mit
   ihnen die Reserve. Sie sind die einzige Grenze in der Tafel, die eine
   **Rate** ist: wie viel legitime Arbeit in der Minute durchgeht. Alles
   andere schützt etwas, das eine kontrollierte Umgebung genauso hat — die
   Kerne, den Speicher, die Dateigriffe, den Zugang.
2. **Was bleibt, und warum**:
   - **Die Plätze** sind keine Rate, sondern eine Ressource. Ohne sie
     liefen zehn Rasterungen zugleich, und der Speicher bestimmte, wann die
     Maschine aufgibt. Der Lastlauf zeigt, dass sie auch ohne Guthaben die
     schwere Arbeit in eine Schlange stellen (`docs/LASTMESSUNG.md`).
   - **Die Sperre und die Drossel für Ratende** gehören zur
     Authentifizierung; das Kriterium nimmt sie ausdrücklich aus.
   - **Rumpf, Antwort, Nachricht und Auftragsbuch** sind Speichergrenzen,
     **die Verbindungsgrenze** schützt die Dateigriffe, **die Fristen** sind
     die grundlegenden.
3. **Eine Einstellung des Geräts**: `grenzen.kontingente`, `an` oder `aus`,
   Vorgabe `an`, Bereich `netz`, Wirkung `neustart`. Dienst und
   Werkzeugserver lesen sie **beim Start** aus dem gezeigten Profil — wie die
   Spur und die Grenze des Zwischenspeichers. **Nur ein ausdrückliches `aus`
   schaltet ab**; ein unlesbarer Wert und ein Lesefehler lassen die Guthaben
   gelten.
4. **Nicht `dienst.kontingente`.** `richtlinie.nurlokal` schaltet jeden
   Schlüssel auf `dienst.` ab (`internal/domain/geltung.go`, `mitNurLokal`).
   Wer „nur lokal" wählte, hätte damit die Kontingente abgeschaltet — eine
   Schranke hätte eine andere geöffnet. Ein Test hält es fest.
5. **Sichtbar an jeder Fläche**: die Auskunft über die Grenzen trägt
   `kontingente: "an"` oder `"aus"` — `GET /grenzen`, `lesen.grenzen`,
   `mse grenzen` und das Fenster unter *Hilfe › Was hinausgeht und
   hereinkommt*. Dienst und Werkzeugserver nennen, **was sie beim Start
   gelesen haben**; Fenster und Kommandozeile nennen, was ein jetzt
   gestarteter Dienst läse. Zwischen einem Umschalten und dem Neustart
   unterscheiden sich beide — das ist die Wirkung `neustart`, und die
   Einstellungsseite sagt sie an der Zeile.

## Einordnung (§16)

**Erweiternd.** Eine neue Einstellung, ein neues Feld in einer Auskunft; die
Vorgabe und damit das Verhalten jedes bestehenden Bestands bleiben.

## Verworfen

- **Ein Schalter je Grenze.** Das Kriterium zählt auf, was bleiben muss; ein
  Schalter für die Plätze oder die Sperre wäre einer, den es verbietet.
- **Eine Fahne am Dienst** (`--ohne-kontingente`). Ein Schalter, der nur für
  einen Lauf gilt, ließe Fenster und Dienst zur selben Zeit verschiedene
  Stände zeigen — dieselbe Absage wie in `docs/EINSTELLUNGEN.md` für die
  Kommandozeile.
- **Eine Umgebungsvariable.** Umgebungsvariablen wählen hier Orte, keine
  Einstellungen (`docs/EINSTELLUNGEN.md`).
- **Größere Zahlen statt „aus".** Wer eine kontrollierte Umgebung hat, will
  keine Grenze, die er später erreicht; und die Zahlen selbst sind eine
  andere Frage ([0083](0083-ressourcenbudgets-ohne-aussperrung.md)).

## Folgen

- **Mit `aus` fällt die Reserve.** Sie hielt einen Teil des globalen
  Guthabens frei; ohne Guthaben gibt es nichts freizuhalten. Was das Fenster
  dann noch schützt, sind die Plätze **des Dienstes**: Dienst und
  Werkzeugserver sind eigene Prozesse mit eigenem Haushalt
  ([0083](0083-ressourcenbudgets-ohne-aussperrung.md), Folgen), und ihre
  schwere Arbeit bleibt auf zwei Rasterungen und eine Einfuhr zugleich
  begrenzt. Die gemeinsame Grenze aller Prozesse ist der Bestand
  ([0017](0017-die-prozesse-bemerken-einander.md)).
- **Ein abgeschaltetes Guthaben merkt sich niemanden**: die Tabelle der
  Principals bleibt leer.
