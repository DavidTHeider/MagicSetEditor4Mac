# 0193 — Der freie Wert einer Panik bleibt am Rechner

**Stand:** angenommen am 2026-10-16 · #2399.

## Lage

Eine gefangene Panik schrieb ihren Wert ins Protokoll — unter `grund`, im
Satz von Wails' `RuntimeError` (`panic: …`), in der Kette von
`domain.ErrZusammengebrochen` und in `panikFehler`. Das Protokoll wird beim
Schreiben **nach Gestalt** geschwärzt (`redigiere`): ein langes Token fällt,
ein kurzes Kennwort aus Buchstaben nicht. Und das Protokoll geht in die
Diagnose, die jemand kopiert, und ins Diagnosepaket. **Die Absturzspur
dagegen** nimmt einen Panikwert nur in der Gestalt, die das Laufzeitsystem
schreibt (`absturzspurgestalt.go`) — die enge Zusage galt für die eine Spur
und nicht für die andere.

## Zur Wahl standen

1. **Den Wert schon beim Schreiben ins Protokoll schwärzen** — wie die Spur.
   Dann stünde auch am Rechner nur noch `«geschwärzt»`, und F02 und F03
   sagen dem Verwalter zu, dass unter der Kennung eines Fehlers der Rohtext
   steht; die Hüllenproben `panik09` und `fehlerform09` lesen ihn dort.
2. **Den Wert am Rechner lassen und auf dem Weg hinaus schwärzen.**

## Entscheidung

**Die zweite.** Im Protokoll dieses Rechners steht der Wert einer gefangenen
Panik ganz — unter dem eigenen Merkmal `panik` an den vier Stellen, die ihn
selbst fangen, und in den Sätzen der übrigen. **Was den Rechner verlassen
kann** — die Diagnose und das Diagnosepaket, beide über `protokollende` —,
trägt ihn nur in der Gestalt des Laufzeitsystems, mit derselben Liste wie die
Absturzspur (`paniktexteSchwaerzen`).

**Warum**: wer am eigenen Rechner einen Fehler sucht, braucht, was geworfen
wurde; was hinausgeht, geht nach einer Vorschau an jemand anderen, und dort
gilt die engere Zusage. Beides an derselben Stelle zu entscheiden, hieße eine
der beiden Zusagen aufzugeben.

## Was bleibt

- **Das Unified Log und die Standardfehlerausgabe** tragen den Wert wie die
  Datei — beide bleiben am Rechner, bis jemand sie selbst weitergibt.
- **Der Filter erkennt Gestalten** — das Merkmal `panik` in beiden
  Formaten des Protokolls (`panik=…` und, seit dem Zweitblick über Phase 5,
  `"panik":"…"` für JSON) und die Sätze `panic: `, `zusammengebrochen: `,
  `eine Panik: `. Ein neuer Weg, der einen
  Panikwert in einem anderen Merkmal und in einem anderen Satz schreibt,
  fiele durch; deshalb schreiben die Stellen, die selbst fangen, unter
  `panik`. **Auch in einer Gruppe** (Nachlauf über Phase 5, Befund 1): bis
  dahin machte die Schwärzung der Weiche eine Gruppe zu ihrer Zeichenkette,
  `[panik=Kennwort hunter]`, und der Filter nahm nur `Kennwort`. Seither
  bleibt sie eine Gruppe, die Weiche schreibt `grund.panik="…"` oder
  `"grund":{"panik":"…"}`, und `paniktexteweiche_intern_test.go` fährt beide
  Formate durch die Weiche selbst — mit Gruppen, gebundenen Gruppen,
  mehrzeiligen und angeführten Werten.
