//go:build fremdpakete

// **Die Geschichte der Zahl** — woher der Abstand zu MSE2 kommt, worüber er
// erhoben ist und was ihn bewegt hat.
//
// Eigene Datei, weil `mse2kanon_intern_test.go` mit ihr über 400 Zeilen ginge
// (§2). Sie trägt keinen Code: **eine Zahl ohne ihre Geschichte lädt dazu ein,
// sie mit einer über eine andere Menge zu vergleichen** — der Fehler, den §10
// am häufigsten anstreicht.
//
// # Der Stand am 2026-08-19, und worüber er erhoben ist
//
// **573 Karten aus vier Sets**, je 168 über das Set verteilt (`Archenemy` hat
// nur 70), gegen MSE2s eigene Ausgabe:
//
//	Archenemy    70 von  70 — Mittel 16,391, schlechtester 25,821
//	White       167 von 168 — Mittel 24,224, schlechtester 38,948
//	Green       168 von 168 — Mittel 22,639, schlechtester 35,424
//	Red         168 von 168 — Mittel 22,942, schlechtester 41,407
//	zusammen    573         — Mittel 22,426, Mitte 22,491
//
// # Der Sprung vom 2026-08-19
//
// **Von 26,714 auf 23,043 — 13,7 %** durch ein einziges Werkstück: eine leere
// Schlusszeile zählte zur Blockhöhe, und der Regeltext saß dadurch rund
// dreizehn Bildpunkte zu hoch
// (#528).
// Betroffen waren 9.380 der 25.148 kanonischen Karten.
//
// **Und von 23,043 auf 22,646** durch den Skalierfilter: MSE rechnet beim
// Verkleinern bilinear und sonst das Flächenmittel, wir taten es genau
// vertauscht (#532).
//
// **Und von 22,646 auf 22,542** durch die boolean-Stiloption
// (#531).
//
// **Die kleine Zahl täuscht.** Die Stichprobe ist proportional und zu 491 von
// 573 `m15-altered`; die Vorlage benutzt das betroffene Feld nicht. Wo es
// wirkt, wirkt es stark — der Lauf **je Vorlage** zeigt es:
//
//	m15-mainframe-planeswalker   47,344 -> 29,529   (−37,6 %)
//	Ajani Goldmane, ganze Karte  55,103 -> 20,220
//	Ajani Goldmane, Kunstfenster 71,941 ->  7,156
//
// Das ist der Grund, warum es zwei Läufe gibt (§10).
//
// **Und von 22,542 auf 22,478** durch die kartenweisen Stiloptionen
// (#547).
// Auch hier täuscht die kleine Zahl — je Vorlage:
//
//	m15-mainframe-dfc   36,858 -> 34,518, schlechteste 73,768 -> 42,588
//	m15-saga            30,986 -> 26,611
//
// **Und von 22,478 auf 22,426** durch die Befunde, die #547 erst sichtbar
// gemacht hat (#555)
// und die Normalform der Mehrfachwahl
// (#556).
//
// **Hier stand einmal, die schlechteste Karte falle „von 66,578 auf 41,407".**
// Das war eine Zahl aus einem früheren Lauf gegen eine aus einem späteren, und
// die Tabelle oben widersprach ihr: sie führte `Green` weiter mit 66,578.
// Angestrichen vom Zweitblick über Phase J.
//
// **Nachgemessen** steht es so: `Green` fällt auf **35,424**, und die
// schlechteste Karte des ganzen Laufs ist die aus `Red` mit **41,407**.
//
// # Diese Zahlen sind **nicht** mit den vorigen vergleichbar
//
// Bis zum Nachlauf über Phase I standen hier 510 Karten und 26,573. Der
// Unterschied ist **keine Verschlechterung am Bild** — es hat sich am Renderer
// nichts geändert. Er kommt aus zwei berichtigten Fehlern der Messung:
//
//  1. **Das Referenzbild wurde unter dem rohen Kartennamen gesucht.** MSE2
//     legt es unter dem **bereinigten** ab; Komma und Apostroph verschwinden
//     spurlos. 64 von 574 Karten fielen damit still aus, und bis auf eine
//     fehlte keine davon wirklich
//     (#535).
//  2. **Die Stichprobe erreichte das Ende ihrer Gruppe nie**
//     (#536).
//
// **Die neue schlechteste Karte ist `Azusa’s Many Journeys`** mit 66,584 — sie
// trägt einen Apostroph und war deshalb bis heute unsichtbar. Sie zeichnet
// über `m15-mainframe-dfc`, also die Klasse aus
// #529: die
// Sagaseite bleibt leer.
//
// **Der Deckel auf der schlechtesten Karte steigt deshalb von 49,000 auf
// 67,000.** Das ist keine Lockerung, sondern die erste ehrliche Zahl — der
// alte Deckel galt für eine Menge, aus der die schlimmsten Fälle
// herausgefiltert waren (§10).
//
// **Die Stichprobe ist seit
// #499
// gerechnet und nicht geraten**: 688 Karten sind die Schranke, unter der mit
// 99,9 % Vertrauen weniger als 1 % der Karten einen Befund liefern; über vier
// Sätze verteilt sind das 168 je Satz.
//
// # Was der Lauf nicht ist
//
// **Kein Beweis für Parität**, sondern ihre Messgröße. Er sagt den Abstand in
// Bildpunkten, nicht, ob ein Mensch die beiden unterscheiden könnte.
//
// **Die gewöhnlichen Karten liegen weiter weg als die Schemen**: 27,7 gegen
// 17,9. Das ist der Grund, warum nicht nur `Archenemy` in diesem Lauf steht.
package app
