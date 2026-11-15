# Beleg: R02 — der Umzug der Kennung auf macOS 27.0.1

**Stand:** Commit `f79e0b2f9` · **Datum:** 2026-10-15 10:32 UTC · **Befehl:** `node scripts/kennunggast.mjs <alt> <neu> mse090-r02` · **Rückgabe:** 0

**Maschine:** in der Lume-VM `mse090-r02` — macOS 27.0.1 (26A434), VirtualMac2,1.

**Bündel:** alt `com.wails.MagicSetEditor4Mac`, Programm `bcbf7c5d5a24`; neu `io.github.davidtheider.magicseteditor4mac`, Programm `98bf1dab7962` — beide Freigabebauten, ohne Quarantäne eingesetzt (die Freigabe unter „Dennoch öffnen" gibt ein Mensch, §12).

**Ergebnis:** alle 27 Schritte gelungen.

| Teil | Schritt | Ergebnis | ms | Wert oder Fehler |
| --- | --- | --- | ---: | --- |
| Fall A, Schreibweise des Bündels | das alte Bündel einsetzen und einmal starten | gelungen | 12783 | 1 |
| Fall A, Schreibweise des Bündels | die alte Probevorgabe unter com.wails.MagicSetEditor4Mac | gelungen | 35 | alt |
| Fall A, Schreibweise des Bündels | eine Marke im alten Webspeicher | gelungen | 5 | 1 |
| Fall A, Schreibweise des Bündels | der Dienst unter dem alten Etikett, geladen | gelungen | 45 | 1 geladen |
| Fall A, Schreibweise des Bündels | die neue Fassung an derselben Stelle starten | gelungen | 4425 | 1 |
| Fall A, Schreibweise des Bündels | die Vorgabe steht unter der neuen Domäne | gelungen | 6 | alt |
| Fall A, Schreibweise des Bündels | der Webspeicher ist kopiert | gelungen | 0 | 1 |
| Fall A, Schreibweise des Bündels | der alte Webspeicher bleibt für den Rückweg | gelungen | 0 | 1 |
| Fall A, Schreibweise des Bündels | genau ein Dienst, unter dem neuen Etikett | gelungen | 17 | 1 |
| Fall A, Schreibweise des Bündels | keine alte plist, die neue startet dasselbe Programm | gelungen | 0 | 1 |
| Fall A, Schreibweise des Bündels | ein zweiter Start ändert nichts | gelungen | 8222 | 1 |
| Fall A, Schreibweise des Bündels | ein abgebrochener Umzug wird nachgeholt | gelungen | 4101 | 1 |
| Fall A, Schreibweise des Bündels | Rückweg: das alte Bündel findet seine Vorgaben | gelungen | 8384 | Vorgabe »alt«; der Dienst steht unter dem neuen Etikett und zeigt auf dasselbe Programm |
| Fall B, klein geschrieben | das alte Bündel einsetzen | gelungen | 63 | 1 |
| Fall B, klein geschrieben | die alte Probevorgabe unter com.wails.magicseteditor4mac | gelungen | 11 | alt |
| Fall B, klein geschrieben | eine Marke im alten Webspeicher | gelungen | 1 | 1 |
| Fall B, klein geschrieben | der Dienst unter dem alten Etikett, geladen | gelungen | 18 | 1 geladen |
| Fall B, klein geschrieben | die neue Fassung an derselben Stelle starten | gelungen | 4294 | 1 |
| Fall B, klein geschrieben | die Vorgabe steht unter der neuen Domäne | gelungen | 6 | alt |
| Fall B, klein geschrieben | der Webspeicher ist kopiert | gelungen | 0 | 1 |
| Fall B, klein geschrieben | der alte Webspeicher bleibt für den Rückweg | gelungen | 0 | 1 |
| Fall B, klein geschrieben | genau ein Dienst, unter dem neuen Etikett | gelungen | 16 | 1 |
| Fall B, klein geschrieben | keine alte plist, die neue startet dasselbe Programm | gelungen | 0 | 1 |
| Fall B, klein geschrieben | ein zweiter Start ändert nichts | gelungen | 8284 | 1 |
| Fall B, klein geschrieben | ein abgebrochener Umzug wird nachgeholt | gelungen | 4139 | 1 |
| Fall B, klein geschrieben | Rückweg: das alte Bündel findet seine Vorgaben | gelungen | 8391 | Vorgabe »alt«; der Dienst steht unter dem neuen Etikett und zeigt auf dasselbe Programm |
| Dokumenttyp | der Typ für .msepackage ist unter der neuen Kennung angemeldet | gelungen | 1689 | 4 Nennungen |

**Künstlich**: die Probevorgabe `R02Probe`, die Marke im Webspeicher und die plist des Dienstes unter dem alten Etikett — sie stehen da, wie das alte Fenster sie hinterließe; kein echter Bestand und kein Konto.

**Was der Lauf nicht zeigt**: die neue Rückfrage von macOS für das lokale Netz und die Kontakte — sie erscheint erst, wenn die Anwendung danach fragt, und ein Fenster des Freigabebaus lässt sich im Gast nicht fernsteuern. Das bleibt eine Handprüfung am Fenster.
