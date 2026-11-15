# 0174 — Wie weit eine Datei festgeschrieben wird, sagt ihre Rolle

**Stand:** angenommen am 2026-10-07 · BK-06a, BK-06b, BK-06c, SEC-20
(#2130,
#2131,
#2132,
#2133).

## Die Frage

**Atomar ist nicht haltbar.** `schreibeAtomar` schrieb in eine Nachbardatei
und benannte um — ein abgebrochener **Prozess** ließ die alte Datei stehen.
Ein Absturz des **Systems** zwischen Umbenennen und Rückschreiben des
Puffers konnte dagegen eine leere Datei unter dem neuen Namen hinterlassen:
weder Datei noch Verzeichnis wurden je festgeschrieben, und der Kommentar
sagte das auch. Eine Sicherung vor der Migration entstand direkt unter ihrem
endgültigen Namen (BK-06a), ein Schließfehler beim Auspacken galt als Erfolg
(BK-06c), und eine vertrauliche Ausgabe übernahm den Modus der Datei, die sie
ersetzte (SEC-20).

## Die Messung

Am 2026-10-07, APFS auf der eingebauten SSD, 31 Läufe je Zelle, Nachbardatei
schreiben, festschreiben, umbenennen, Verzeichnis festschreiben, nacheinander
(Last um 3,7 durch die Speicherverwaltung des Systems):

| Größe | ohne | `fsync` | `F_FULLFSYNC` |
| --- | --- | --- | --- |
| 4 KiB | 0,23 ms (p90 0,36) | 0,26 ms (p90 0,29) | 8,0 ms (p90 8,3) |
| 1 MiB | 0,39 ms (p90 0,54) | 0,56 ms (p90 2,8) | 8,1 ms (p90 8,4) |
| 16 MiB | 6,1 ms (p90 8,2) | 5,0 ms (p90 7,9) | 16,9 ms (p90 18,8) |

`fsync` ist unter macOS fast umsonst, weil es den Puffer **des Datenträgers**
nicht leert; `F_FULLFSYNC` tut das und kostet je Datei samt Verzeichnis rund
8 ms — bei einer Bildausfuhr von 300 Karten wären das 2,4 Sekunden.

## Die Entscheidung

| Rolle | Wer | Festgeschrieben | Modus |
| --- | --- | --- | --- |
| **Ausgabe** | alles, was jemand speichert: Sets, Bilder, PDF, Beiwerk, Symbole, Pakete | `fsync`, Datei und Verzeichnis — **dieselbe Stufe wie der Bestand** (`synchronous=FULL`) | neu 0600, sonst der der Datei davor |
| **vertrauliche Ausgabe** | Personendatenauskunft, Diagnose, Einstellungsausfuhr | wie Ausgabe | **immer 0600** |
| **Wiederanlauf** | Epochen- und Rückspielmarke, Vermerk eines Paketersatzes, eingesetzter Bestand, Sicherung vor der Migration | `F_FULLFSYNC`, Datei und Verzeichnis | 0600 |
| **Zwischenspeicher** | der Bildvorrat wartender KI-Vorschläge | gar nicht | neu 0600 |

**Fehler vor dem Veröffentlichen** — Schreiben, Festschreiben, Schließen,
Umbenennen — lassen die alte Datei Byte für Byte stehen und keine
Nachbardatei zurück. **Ein Verzeichnis, das sich nicht festschreiben lässt,
ist keine Erfolgsmeldung**: die Datei steht dann da, und der Satz sagt es.
Beim Einsetzen eines Bestands legt der Rückweg in diesem Fall den alten
zurück.

**Eine Sicherung erscheint erst vollständig** (BK-06a): `VACUUM INTO` in
eine versteckte Nachbardatei, festschreiben, mit `link(2)` veröffentlichen —
atomar und ohne zu überschreiben —, Verzeichnis festschreiben. Ein vergebener
Name ist ein Konflikt, und die vorhandene Datei bleibt, wie sie ist.

**Ein Schließfehler ist ein Fehler** (BK-06c): beim Auspacken eines Pakets
die einzige schreibende Schließung des Pakets `bezug`; scheitern Schreiben
und Schließen, bleiben beide Gründe stehen. Die übrigen drei Schließungen dort
schließen Leser.

## Was es nicht ist

- **Keine Zusage über einen Stromausfall.** Ob der Datenträger einen
  Puffer verliert, entscheidet seine Bauart; geprüft ist hier der
  Prozessabbruch (zweiter Prozess, `sicherungsabbruch_test.go`,
  `atomarabbruch_test.go`) und jeder eingespielte Fehler. Eine Ausgabe hat
  dieselbe Stufe wie der Bestand, nicht mehr.
- **Kein Festschreiben in ein Verzeichnis ohne Leserecht**: es lässt sich
  beschreiben, aber nicht öffnen. Dort zu schreiben ging vorher, und ein
  Fehler wäre ein Rückschritt; der neue Name ist dort nicht festgeschrieben.
- **Kein Aufräumen liegengebliebener Nachbardateien**: ob eine zu einem Lauf
  gehört, der gerade noch schreibt, weiß niemand. Sie tragen einen Punkt vorn.
- **Kein `chmod` fremder Dateien**: nur eine vertrauliche Ausgabe setzt 0600,
  und zwar an der Datei, die sie selbst schreibt.

## Verworfen

- **`F_FULLFSYNC` für alles** — 8 ms je Datei; die Stufe des Bestands ist
  `fsync`, und eine Ausgabe, die haltbarer ist als der Bestand, von dem sie
  stammt, gewinnt nichts.
- **`renameatx_np(RENAME_EXCL)` statt `link(2)`** — dasselbe unter macOS,
  aber nur dort; `link` geht auf jedem Unix.

## Nachtrag vom 2026-10-19: der neue Bestand

**Ein neu angelegter Bestand entsteht mit `0600`**, wie ein eingesetzter
(#2508,
gefunden beim Erststart aus dem Abbild, HAND6). Bis dahin legte SQLite ihn mit
der `umask` an, `0644`, und dieselbe Datei hatte je nach Weg zwei Rechte. Seine
Begleiter folgen ihm: WAL und Journal über `modeof`, die `-shm`, indem sie vor
dem Öffnen mit den Rechten des Bestands angelegt wird. Der Treiber legte beide
sonst mit `0666` und der `umask` an (`internal/storage/bestandsrechte.go`).
**Ein vorhandener Bestand behält seine Rechte**: kein `chmod` fremder Dateien,
wie oben.
