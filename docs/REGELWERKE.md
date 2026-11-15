# Die offiziellen Regelwerke — wie sie benutzt werden

**Verlegt aus `AGENTS.md` §11 am 2026-09-22** — ENG-01.AK-04
(#1726).
**Die Regeln stehen weiter dort**: nachgelesen statt erschlossen, zitiert
wird Nummer und Fassung, kein Wortlaut, kein Dokument im Baum, keine
Nummer ohne Nachlesen. Hier steht, woran das schon gescheitert ist.

---

## Eine neue Fassung wird abgeglichen, nicht übernommen

**Am 2026-09-25 erschien eine neue Fassung der CR** — gesehen am Tag zuvor
beim Nachlesen für #1839. Die Fassung in `AGENTS.md` §11 stand danach noch
einen Tag auf der alten, und zwar mit Absicht: **eine Fassung im Kommentar,
die niemand gelesen hat, sähe aus wie ein Beleg**
(#1842).
Fortgeschrieben wurde sie erst, als beide Fassungen **Regel für Regel**
nebeneinander gelesen waren; was sich geändert hat, steht je Kapitel in
`REGELABDECKUNG.md` §1. **Die Zitate im Kern wurden dabei nicht umgeschrieben**:
sie nennen, was gelesen wurde — und wo eine Nummer in der neuen Fassung
etwas anderes meint (CR 506.6 bis 506.8), steht das in derselben Tafel.

## Ältere Verweise auf die MTR tragen keine Fassung

**Gezählt am 2026-09-27: 90 Zeilen in 26 Quelltextdateien** nennen die
Magic Tournament Rules — meist §3.1, Appendix C oder Appendix E — ohne
Fassung; **9 Zeilen** tragen eine. Gezählt wird **je Zeile**, mit diesem
Kommando aus der Wurzel des Repositoriums:

```bash
grep -rnE 'MTR|Magic Tournament Rules|Appendix [A-Z]' internal cmd frontend/src scripts --include='*.go' --include='*.ts' --include='*.tsx' --include='*.mjs' | grep -v 'Fassung'
```

**Was die Zählregel nicht sieht**: eine Fassung in der **nächsten** Zeile —
die Nennung zählt dann als eine ohne. Die Zahl ist deshalb eine Obergrenze,
und bis zum Nachlauf über Phase 6.6 stand hier eine andere (85 in 35), mit
einem anderen Muster und ohne Kommando — der Nachlauf hat es angestrichen. Sie stammen überwiegend aus Phase 4.7, bevor §11 die Fassung
verbindlich machte. Angestrichen hat es der Zweitblick über Phase 6.6 an den
**neuen** Stellen; die tragen seitdem *Fassung 2026-02-27*, denn für sie ist
§3.1 und Appendix E gegen dieses Dokument nachgelesen.

**Den älteren wird keine Fassung untergeschoben** — aus dem Grund, der im
Abschnitt darüber steht: eine Fassung, die niemand gelesen hat, sähe aus wie
ein Beleg. Nachgetragen wird sie, wenn jemand die Stelle gegen das Dokument
liest. Einen Wächter dafür gibt es nicht; **diese Zahl ist die einzige
Zusage**, und sie gilt für den genannten Tag.

## Eine falsche Nummer sieht aus wie ein Beleg

**Hier stand `CR 712.4`, und das war die falsche Nummer** — sie gehört zu
den **Meld-Karten**; die Vorderseite außerhalb von Schlachtfeld und Stapel
ist CR 712.8a. Gefunden am 2026-09-17 beim Bau von `M12`
(#1614),
als dessen erster Entwurf das Umwandeln unter `CR 711` ablegte — das
Kapitel heißt **Leveler Cards**, umgewandelt wird nach CR 701.27a.

**Und `CR 103.2` für den Startspieler, an sechs Stellen** — in Kern,
Dienst, Fassade und Regelabdeckung. Wer anfängt, bestimmt CR 103.1; 103.2
sind die zusätzlichen Schritte davor (Sideboard, Companion, Commander),
und zwar in **beiden** Fassungen, 2026-08-19 wie 2026-09-25. Gefunden am
2026-09-25 beim Nachlesen für den Zweitblick über Phase 6.9, als der
Beginn einer Partie gegen CR 103.8 geprüft wurde — die Nummer stand seit
G9 da und sah aus wie ein Beleg.

**Die ersten beiden Male stand die falsche Nummer im Beispiel für
„nachgelesen statt erfunden".** Genau davon spricht §10: eine Nummer sieht aus wie ein Beleg,
und niemand schlägt sie nach. Der Wächter
[`internal/werkzeuge/regelwortlaut_test.go`](../internal/werkzeuge/regelwortlaut_test.go)
hält den **Wortlaut** fern; dass eine Nummer das Richtige trifft, hält er
nicht — das steht in seinem Kopf, und es bleibt Handarbeit.

## Vierundvierzig Sätze

**Die Regel „kein Wortlaut" stand seit dem 2026-09-15 da und hat nicht
getragen.** Gemessen
am 2026-09-16: **vierundvierzig** abgeschriebene Sätze aus CR und MTR in
**sechzehn** Dateien, alle älter als die Regel und keiner je wieder
angesehen — darunter einer in `internal/domain/standardland.go` unter der
Überschrift „Die Regel, wörtlich nachgelesen“.
