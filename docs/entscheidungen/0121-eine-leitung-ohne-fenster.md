# 0121 — Eine Leitung darf auch ohne Fenster gefahren werden

**Stand:** angenommen
**Datum:** 2026-09-19

## Lage

**Jede Fähigkeit, die eine Verbindung aufbaut oder annimmt, gibt es heute nur
am Fenster** — Horchen, Fernhorchen, Sichtbarkeit, Partiebeitritt,
Partiefreigabe. `docs/FAEHIGKEITEN.md` begründet das an fünf Stellen mit
demselben Satz:

> „ein Prozess, der bis ⌃C daran sitzt, ist kein Befehl"

**Dieses Programm liefert zwei solche Prozesse aus.** `mse-rest` hält einen
Webdienst, `mse-mcp` spricht über die Standardeingabe; beide laufen, bis
jemand sie beendet, beide behandeln ⌃C und SIGTERM über
`signal.NotifyContext`, und beide sind gebaute, geprüfte Befehle. Die
Begründung sagt also, dass es nicht geht, während es zweimal danebensteht —
die Gestalt aus §10: eine Ablehnung, die wie eine Begründung aussieht.

**Und ohne Fläche gibt es keine Abnahme.** `T12`
(#1477)
verlangt einen Lauf über zwei Rechner; der zweite ist eine Debian-VM, und die
fährt kein Wails-Fenster. Jede Prüfung eines Transports endet heute bei einem
Doppel oder bei einer Schleife im selben Prozess.

## Entscheidung

**Es gibt eine Fläche ohne Fenster für dauerhafte Verbindungen, und sie ist
die Kommandozeile.** Sie hat die Gestalt der beiden vorhandenen
Dauerprozesse: Rückgabewert, Protokoll auf die Standardfehlerausgabe, Abbruch
über SIGINT und SIGTERM.

**Nichts horcht von selbst.** Die Zusage aus N2 bleibt Zeichen für Zeichen:
ohne bewusste Auswahl geschieht nichts, auch nach einem Neustart nicht. **Ein
Befehl, den jemand tippt, ist eine bewusste Auswahl** — ein Autostart wäre
keine, und es gibt keinen.

## Verworfen

- **Ein MCP-Werkzeug.** Die Begründung dort ist eine andere und bleibt: „ein
  Port, den ein Modell öffnet, überlebt seinen Ruf"
  ([0106](0106-drei-vertrauensgrenzen-und-nur-eine-wird-gebaut.md)). Sie ist
  ein Satz über **Werkzeuge** und nicht über die Kommandozeile.
- **Ein REST-Endpunkt fürs Horchen.**
  [0115](0115-die-schleife-ist-eine-vorgabe.md) knüpft jede Öffnung an eine
  sichtbare Zustimmung **an diesem Rechner**, und wer über einen Port
  anfragt, kann sie nicht geben.
- **Ein REST-Endpunkt fürs Tragen**, und dieser Grund ist ein anderer und
  wurde erst beim Bauen gemessen: das Tragen wirkt `externSenden`, und **ein
  Token reicht nie so weit** ([0068](0068-ein-token-ist-kein-kennwort.md),
  [0078](0078-der-wirksame-akteur.md), `domain.tokenreichweite`). Der Dienst
  prüft jede Anfrage gegen ein Token — der Weg wäre dort **immer** eine
  Absage. **Er stand schon gebaut da**, samt Route, Naht und Wächterzeile;
  herausgenommen wurde er, als die Wirkungsklasse feststand. Ein Endpunkt,
  der nie glückt, ist keine Fläche, sondern eine Zusage, die niemand einlöst
  (§10).
- **Ein Dienst, der beim Anmelden startet.** Das wäre die „unbemerkte
  Dauerwerbung aller Profile", die N2 ausschließt.

## Folgen

- Die fünf Zeilen in `docs/FAEHIGKEITEN.md` bekommen ihre CLI-Spalte; die
  alte Begründung fällt, **auch wo keine Spalte dazukommt**, weil sie nicht
  stimmt. **Eine Spalte kam dazu** (`mse horchen`); die vier übrigen stehen
  seither als **Lücke** da und nicht als Ablehnung.
- **Das Fenster bekommt das Tragen mit**, obwohl dieses Issue von der Fläche
  **ohne** Fenster handelt. Der Grund ist §16 und ein Loch, das beim Bauen
  sichtbar wurde: `HandelSchicken` legt in den Ausgang, und **getragen hat
  das niemand** — ein Fenster konnte einen Handel annehmen und keinen
  hinausbringen. Ein Angebot lag damit für immer im Ausgang.
- `docs/HAERTUNG.md` bekommt eine Ebene zwischen der Pipeline und dem Lauf
  gegen einen fremden Dienst: **der Lauf zwischen zwei Prozessen über einen
  echten Draht**.
- Was einen zweiten **Rechner** braucht, bleibt eine Handprüfung
  ([0007](0007-kein-entfernter-lauf.md)) — die Fläche macht sie möglich, sie
  macht sie nicht automatisch.
