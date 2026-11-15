# Beleg: purego v0.11.1 an der echten Bindung — #2305

**Gefahren am 2026-10-12** an Commit `1ed34917b` (Zweig
`chore/2305-purego`, sauberer Baum), auf diesem Rechner: Mac17,3, macOS
27.0.1, go1.27.2 darwin/arm64. Geladen nach der Bauauskunft des
Abnahmebündels:

```text
dep	github.com/ebitengine/purego	v0.11.1
dep	github.com/tmc/apple	v0.6.18
```

`tmc/apple` verlangt weiter v0.11.0-alpha.6; gebaut wird nach MVS gegen die
höhere Fassung. Gemessen ist deshalb jeder Weg, der über purego ruft.

## Der Sensorlauf — die Bindung an `LocalAuthentication`

```text
go test -tags sensorlauf -count=1 -v \
  -run '^(TestDieGeladenenFassungen|TestDerSensorIstDa|TestEineFristEndetOhneSpaetenErfolg|TestEinAbbruchEndetOhneSpaetenErfolg)$' \
  ./internal/biometrie/

=== RUN   TestDieGeladenenFassungen
--- PASS: TestDieGeladenenFassungen (0.00s)
=== RUN   TestDerSensorIstDa
--- PASS: TestDerSensorIstDa (0.01s)
=== RUN   TestEineFristEndetOhneSpaetenErfolg
    sensorlauf_test.go:92: nach 4.005s: Validierung fehlgeschlagen: context deadline exceeded
--- PASS: TestEineFristEndetOhneSpaetenErfolg (4.01s)
=== RUN   TestEinAbbruchEndetOhneSpaetenErfolg
    sensorlauf_test.go:153: nach 2.004s: Validierung fehlgeschlagen: context canceled
--- PASS: TestEinAbbruchEndetOhneSpaetenErfolg (2.00s)
PASS
ok  	github.com/DavidTHeider/MagicSetEditor4Mac/internal/biometrie	6.244s
```

Dieselben vier Fälle wie in [`sec19-sensor.md`](sec19-sensor.md) mit
alpha.6; Frist und Abbruch enden wie dort ohne späten Erfolg. **Nicht
gefahren** wie dort: Ablehnung und Zustimmung, die einen Menschen am Sensor
brauchen.

## Am nativen Fenster — Teilen, Vollbild, Start

`node scripts/huellenprobe.mjs net04 vollbild start`, Lauf
`2026-10-12T00-05-24Z-14672`, Bauabdruck `66ad4bcbc707`:

| Drehbuch | ruft über purego | Ergebnis |
| --- | --- | --- |
| [`net04`](huellenprobe/net04.md) | `NSSharingServicePicker` (`teilen_darwin.go`) | bestanden, 7 von 7 |
| [`vollbild`](huellenprobe/vollbild.md) | `toggleFullScreen:` und das Fenstermenü (`vollbild_darwin.go`) | bestanden, 11 von 11 |
| [`start`](huellenprobe/start.md) | die Abnahmeleitung, `class_replaceMethod` für den ersten Klick (`abnahme_hand_darwin.go`) | bestanden, 15 von 15 |

**Was er nicht zeigt**: ein Ziel des Teilen-Dialogs, das wirklich sendet —
das Drehbuch öffnet den Dialog und schließt ihn wieder.

**Rückweg**, falls eine Probe später rot wird: die Pinnung auf
v0.11.0-alpha.6 in `go.mod`, eine Zeile, mit dem Befund in #2305.
