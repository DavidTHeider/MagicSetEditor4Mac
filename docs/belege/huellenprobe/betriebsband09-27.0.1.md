# Beleg: OPS02 — ein Dienst, der endet, meldet sich im Betriebsband, macOS 27.0.1

**Stand:** Commit `00426d84f` · **Datum:** 2026-10-15 19:25 UTC · **Befehl:** `node scripts/betriebsband09gast.mjs mse090-r02` · **Rückgabe:** 0

**Maschine:** in der Lume-VM `mse090-r02` — macOS 27.0.1 (26A434), VirtualMac2,1, 4 Kerne.

**Bündel:** Abnahmebau, Baustand `dd819dc6f398`, Programm `487825c2de59bc1d…`, gebaut mit `minos 15.2`.

**Ergebnis:** alle 4 Schritte gelungen.

| Teil | Schritt | Ergebnis | ms | Wert oder Fehler |
| --- | --- | --- | ---: | --- |
| 1 | ein Wegwerfort, das Fenster steht, kein Hinweis auf den Dienst | gelungen | 10549 | kein Hinweis, 0 Hinweise sonst |
| 2 | der Dienst geladen, über einem Bestand ohne Rechte | gelungen | 1586 | Ausgang 1 |
| 3 | das Fenster nennt den Dienst, und der Hinweis bleibt über vier Abrufe | gelungen | 120360 | 4 verschiedene Läufe gesehen |
| 4 | die Rechte zurück: der Dienst läuft durch, der Hinweis geht | gelungen | 15397 | fort |

**Künstlich**: zwei Wegwerforte mit eigenem Zuhause, Bestand und Bund — einer für das Fenster, einer für den Dienst, dessen Ordner zuerst keine Rechte hat. Die plist folgt der der Anwendung (`dienstplist` in `internal/app/schnittstellenziele.go`), dazu nur die Umgebung des Wegwerforts. Kein echter Bestand, kein Konto.

```
launchctl print nach dem Laden: ausgang 1, code 0, laeufe 1, zustand spawn scheduled
Runde 1: ausgang 1, code 0, laeufe 4, zustand spawn scheduled
Runde 2: ausgang 1, code 0, laeufe 7, zustand spawn scheduled
Runde 3: ausgang 1, code 0, laeufe 10, zustand spawn scheduled
Runde 4: ausgang 1, code 0, laeufe 13, zustand spawn scheduled
danach: ausgang 1, code 0, laeufe 14, zustand running
```

**Gefragt wurde die Bindung `Betriebsband` über die Naht des Fensters**, nicht die gezeichnete Seite: hinter dem Sperrbildschirm des Gastes ist sie verborgen, und das Fenster fragt dann nicht — so ist es gebaut. Wie der Hinweis aussieht und wohin er führt, belegt `betriebsband09` am Wirt.

**Was der Lauf nicht zeigt**: Platz, Sicherung und Webprozess — die belegt die Hüllenprobe `betriebsband09` am Wirt; und das Einrichten des Dienstes über das Fenster, das `ops03gast` belegt.
