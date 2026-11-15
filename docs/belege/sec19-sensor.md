# Beleg: die echte Bindung am Touch-ID-Sensor — SEC-19

**Gefahren am 2026-10-07** an Commit `f5b08e4f2`, sauberer Baum, auf diesem
Rechner: macOS 27.0.1 (26A434), go1.27.1 darwin/arm64, Touch ID eingebaut.
Die Läufe stehen hinter dem Tag `sensorlauf`
(`internal/biometrie/sensorlauf_test.go`) und rufen die **echte** Bindung —
`tmc/apple` über `purego`, kein Doppel.

```text
go test -tags sensorlauf -count=1 -v \
  -run '^(TestDieGeladenenFassungen|TestDerSensorIstDa|TestEineFristEndetOhneSpaetenErfolg|TestEinAbbruchEndetOhneSpaetenErfolg)$' \
  ./internal/biometrie/

=== RUN   TestDieGeladenenFassungen
--- PASS: TestDieGeladenenFassungen (0.00s)
=== RUN   TestDerSensorIstDa
--- PASS: TestDerSensorIstDa (0.02s)
=== RUN   TestEineFristEndetOhneSpaetenErfolg
    sensorlauf_test.go:90: nach 4.004s: Validierung fehlgeschlagen: context deadline exceeded
--- PASS: TestEineFristEndetOhneSpaetenErfolg (4.00s)
=== RUN   TestEinAbbruchEndetOhneSpaetenErfolg
    sensorlauf_test.go:151: nach 2.004s: Validierung fehlgeschlagen: context canceled
--- PASS: TestEinAbbruchEndetOhneSpaetenErfolg (2.00s)
PASS
```

## Die Kriterien

| AK | Stand |
| --- | --- |
| Genau `tmc/apple` v0.6.18 und `purego` v0.11.0-alpha.6 geladen | **erfüllt** — gelesen aus der Bauauskunft des Testprogramms; die Gegenprobe mit v0.6.17 als Erwartung wurde rot |
| Frist → kein später Erfolg | **erfüllt** — der Dialog stand, niemand antwortete; nach 4,00 s ein Fehler der Art `ErrValidation`, danach nichts |
| Abbruch → kein später Erfolg | **erfüllt** — Abbruch des Kontexts nach 2 s mitten im Dialog; `Invalidate` schloss ihn |
| Ablehnung → kein Erfolg, benannter Fehler | **nicht gefahren** — siehe unten |
| Zustimmung → Erfolg | **nicht gefahren** — dazu muss ein Finger auf den Sensor |

## Warum die Ablehnung nicht gefahren ist

Die Freigabe des Auftraggebers erlaubte, den Dialog **ohne Menschen** zu
bedienen. Gemessen ist, dass kein Weg dafür offensteht, ohne eine
Sicherheitseinstellung zu ändern:

- Den Dialog zeigt `coreautha`, kein Programm, das sich als App freigeben
  ließe; Computer-Use führt ihn nicht, und für sein Vordergrundprogramm war
  die Freigabe abgelehnt.
- Der Prozess, der die Läufe fährt, hat weder Bedienungshilfen noch das
  Recht, Ereignisse zu senden: `AXIsProcessTrusted()` und
  `CGPreflightPostEventAccess()` geben beide `false` — abgefragt, ohne
  eine Rückfrage auszulösen.

Beides einzuschalten ist eine Einstellung unter *Datenschutz & Sicherheit*,
und die setzt nur der Eigentümer des Rechners. **Die Ablehnung braucht einen
Klick**, und der Lauf dafür ist eine Zeile:

```text
go test -tags sensorlauf -count=1 -v -run TestEineAblehnungIstKeinErfolg ./internal/biometrie/
```

Er wartet 90 s auf „Abbrechen". Endet er erst an der Frist, wird er **rot** —
eine Frist ist keine Ablehnung, und er gibt sie nicht als solche aus.

## Was dieser Beleg nicht sagt

- **Den Weg durch die Anwendung**: der Dialog kommt hier aus einem
  Testprogramm, nicht aus dem Fenster beim Entsperren eines Profils.
- **Face ID** und einen Rechner ohne Sensor — beides gibt es hier nicht.
