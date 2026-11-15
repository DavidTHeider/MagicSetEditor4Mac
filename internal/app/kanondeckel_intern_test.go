//go:build fremdpakete

// **Der Deckel des Kanonlaufs — und wie er dahin gekommen ist.**
//
// Eigene Datei, weil `kanonlauf_intern_test.go` mit ihr über 400 Zeilen ginge
// (§2). Der Schnitt folgt der Frage: dort steht, **was** gemessen wird, hier,
// **wogegen** die Zahl gehalten wird und warum sie so hoch ist und nicht
// anders.
//
// **Die Geschichte steht mit**, und das ist Absicht: eine Zahl ohne ihre
// Herkunft ist eine, die beim nächsten roten Lauf jemand hochsetzt.
package app

// **Der heutige Stand, damit ein Rückschritt rot wird.**
//
// Gemessen am 2026-08-17 über 250 Karten aus zehn Sets: erst **2.794**,
// nach #405
// noch 294, nach #409
// 244 und nach #415
// noch **178** Befunde — 2.500 davon waren zehn Meldungen derselben
// Symbolschrift auf jeder Karte. Wer die Zahl senkt, senkt sie hier mit;
// sie ist ein Deckel, kein Ziel.
//
// **Kein Vergleich mit MSE2s Bild.** Dafür bräuchte es dessen Ausgabe zu
// diesen Sets, und die liegt nicht vor; dieser Lauf sagt, was **wir**
// nicht deuten, nicht was anders aussieht (§10).
// **Seit #420 sind es 131**: die Messstufe rechnet ihre Schriftangaben aus,
// statt sie roh zu lesen, und damit misst sie überhaupt erst für die vier
// Fünftel der Feldstile, deren `font: size:` ein Skript ist.
//
// **Seit #422
// sind es 105**: ein Symbolbild darf ein Skript sein, und eine Farbe darf
// als Zeichenkette ankommen. Die 26 Befunde davor waren **ein** Symbol —
// das Chaos-Zeichen auf jeder Planechase-Karte.
//
// **Seit #424
// sind es 81**: `set.set_info` gibt es, und damit entscheidet
// `magic.mse-game` wieder, ob eine mehrfarbige Karte einen Verlauf bekommt.
//
// **Seit #429
// sind es 28**: eine Auszeichnung, die MSE selbst übergeht, ist kein
// Mangel dieser Anwendung. Von den 70 Auszeichnungsbefunden bleiben 17 —
// `soft-line`, `margin:`, `li` und `bullet`, wo MSE etwas tut und wir
// noch nicht.
//
// **Seit #431
// sind es 16**: `<soft-line>` bricht weich um, statt gemeldet und
// weggeworfen zu werden.
//
// **Seit #433
// sind es 12**: `<margin:…>` rückt einen Absatz ein — samt MSEs Fehlgriff
// beim Abgreifen der Zahlen ([0027](../../docs/entscheidungen/0027-mses-margin-fehlgriff-wird-nachgebaut.md)).
// Es bleibt **eine** Auszeichnung: `li`.
//
// **Seit dem Aufzählungspunkt sind es 11** — keine Auszeichnung mehr in
// der Liste (#431).
//
// **Seit #438
// sind es 9**: ein zweites bloßes Argument ist bei MSE kein Fehler,
// sondern eine zweite Zuweisung an `input` — und die erste gewinnt.
//
// **Seit #440
// sind es 4**: `card.<bildfeld>` ist ein Bild aus dem Set-Paket und kein
// Dateiname im Stilpaket. Es bleiben zwei Masken, die es im Pack wirklich
// nicht gibt, und zwei fehlende Glyphen.
//
// **Seit #442
// sind es 2**: eine Ebene ohne Fläche wird gar nicht erst betreten und
// sucht damit auch keine Maske. Was bleibt, sind zwei fehlende Glyphen.
// **Seit #499
// sind es 7 — bei einer 2,7-mal größeren Stichprobe.** Die Zahlen sind
// deshalb **nicht** vergleichbar: 2 von 250 sind 0,8 %, 7 von 680 sind
// 1,0 %. Was die größere Stichprobe zutage gefördert hat:
//
//	2 × in der Schrift Beleren Small Caps Bold fehlen Zeichen   (#446)
//	1 × in der Schrift Matrix fehlen Zeichen                    (#446)
//	2 × das Bild ist in keinem Paket zu finden                  (#497)
//
// **Seit #502
// sind es 5 statt 7**: eine Schrift steht jetzt auch unter ihrem
// **Familiennamen** im Index. „die Schrift Beleren Small Caps ist nicht zu
// finden" stand über den ganzen Kanon **48-mal** — für eine Datei, die
// dalag.
// **Der Deckel steht auf der Art, nicht auf der Summe**
// (#513).
//
// Eine Summe aus „dieses Programm kann etwas nicht", „die Vorlage rechnet
// ins Leere" und „auf diesem Rechner fehlt eine Schrift" sagt bei keinem
// der drei, was zu tun ist.
//
// **Portierung steht auf null**, und das ist die Zusage: was dieses
// Programm nicht kann, ist gezählt und soll keines sein.
// **Die anderen beiden bekommen einen Deckel, keine Null.** Sie hängen an
// den Vorlagen und an diesem Rechner; auf null zu deckeln hieße, eine
// fremde Datei zur eigenen Zusage zu machen.
// **Die Deckel sind Raten, keine Zahlen** — je 1.000 gezeichnete Karten.
// Ein Deckel als Zahl gälte nur für die Stichprobe, für die er gesetzt
// wurde; wer `MSE_STICHPROBE=alle` fährt, bekäme ihn rot, ohne dass sich
// etwas verschlechtert hätte
// (#517).
//
// **`Umgebung` steht seit dem Ausweich auf null**
// (#515):
// ein Zeichen, das der genannten Schrift fehlt, wird aus einer anderen
// geholt — so wie wxWidgets es für MSE tut. Über 1.663 gezeichnete Karten
// bleibt keiner.
//
// **Auf einem Rechner ohne die Paketschriften kann er wieder steigen**,
// und dann ist die Zahl richtig und der Deckel falsch — das ist die Art,
// die an der Maschine hängt.
// **Über alle 23.168 Karten waren es 15 der Art `Vorlage`** — die
// Fusible-Split-Karten, deren Vorlage einen Pfad auf `true` ausrechnet
// (#497).
// Das war eine Rate von 0,65 je 1.000, und der Deckel lag mit 1,0 knapp
// darüber.
//
// **Am 2026-08-20 neu gemessen: 43 über 25.148 Karten, also 1,71** — und
// damit war der Deckel rot, während die Stichprobe von 680 Karten **null**
// zeigte (#610).
//
// Die 43 sind 40 × „das Bild ist in keinem Paket zu finden" und 3 × ein
// Symbol für ein Leerzeichen. Beides ist Art `Vorlage` **mit Grund**: die
// Vorlage rechnet einen Pfad aus, den es nicht gibt, und MSE tut dort
// dasselbe — es schweigt nur dazu (`befundart.go`, am Orakel belegt).
//
// **Woher der Anstieg kommt, ist eine Vermutung und keine Messung**:
// 25.148 − 23.168 sind genau die **1.980** Karten, die
// #539
// zurückgeholt hat — sie fielen vorher an einer unpaarigen Klammer aus dem
// Import und konnten deshalb keinen Befund tragen. Nachgezählt je Satz ist
// das nicht; der volle Lauf dauert 44 Minuten.
//
// **Seit #610
// wird geschichtet gezogen**, und die Zahl darunter meint deshalb eine
// andere Menge als vorher.
//
// Gemessen auf dem Stand **vor**
// #664,
// gleiche Größe, 680 Karten aus zehn Sätzen:
//
//	Ziehung           | fand von den 40
//	------------------|----------------
//	jede n-te         |  0
//	geschichtet       | 27
//
// Eine Rate hilft gegen unterschiedliche **Größen** der Stichprobe; sie
// hilft nicht gegen eine, die den Fall gar nicht **enthält**. Nach
// Stilangaben geschichtet enthält sie ihn.
//
// **Was auch die geschichtete Ziehung nicht fängt**: einen Befund, der an
// einer **einzelnen Karte** hängt und an keiner Stiloption. Die drei
// verbliebenen Befunde des vollen Laufs sind von dieser Art — ein
// `<sym>`-Block, der über einen Zeilenumbruch reicht, an drei Karten von
// 25.148. Keine Schichtung nach Stilangaben trifft sie; dafür gibt es nur
// den vollen Lauf, und der dauert knapp drei Stunden.
//
// **Das steht hier, statt den Deckel danach zu wählen**: ein Deckel, der
// über der bekannten Zahl liegt, meldet nichts mehr.

// RateVorlage und RateUmgebung sind die Deckel — **Raten je 1.000 gezeichnete
// Karten**, nicht Zahlen.
//
// Ein Deckel als Zahl gälte nur für die Stichprobe, für die er gesetzt wurde;
// wer `MSE_STICHPROBE=alle` fährt, bekäme ihn rot, ohne dass sich etwas
// verschlechtert hätte
// (#517).
const (
	RateVorlage  = 2.0
	RateUmgebung = 0.0
)

// **Ein Deckel auf dem Mittel** — er fängt eine allgemeine
// Verschlechterung. Über **573** Karten erhoben; die Zahl ist deshalb
// **nicht** mit der über 510 vergleichbar (#535, #536).
//
// **Seit #528, #532 und #531 steht er auf 22,700 statt 26,800.**
//
// **Am 2026-08-20 auf 21,800 nachgezogen** (#529): gemessen waren 21,581.
// Phase G und die Arbeiten danach haben ihn gedrückt, und ein Deckel, der
// einen Punkt über dem Stand liegt, fängt keine Verschlechterung mehr.
//
// **Und am selben Tag auf 20,900** (#546): das Viertelraster von Zeilenhöhe
// und Grundlinie hat 21,553 auf **20,692** gebracht.
//
// **Am 2026-08-21 auf 20,500** (#598): das dritte Maß derselben Messkette,
// die **Vorsilbenbreite**, hat 20,662 auf **20,333** gebracht. Die
// Textebene allein von 81,203 auf 79,724.
//
// **Und am selben Tag auf 19,550** (#752): **Kerning**. `beleren-bold` —
// die Schrift jedes Kartennamens — führt 1.983 Paare, und wx wendet sie
// an; dieses Programm summierte einzelne Vorschübe. 20,365 → **19,371**,
// die Textebene von 79,775 auf 75,730.
//
// **Der erste Anlauf machte es schlechter** (20,365 → 20,377), weil das
// Kerning nur in die gemessene Breite ging und nicht in den Vorschub beim
// Zeichnen: die Zeile wurde schmaler gemessen und weiter ungekernt
// gesetzt. Es sind **vier** Summenstellen, und sie gehören zusammen.
//
// # Am 2026-08-22 neu erhoben — und warum das kein Hochsetzen ist (#862)
//
// Der Lauf war rot, und die Ursache lag **nicht** im Code. Sie ist gemessen:
//
//	M15-Magic-Pack, 3 Sets, 503 Karten   20,047   (auch am Stand 536b7c2)
//	volles Pack,    5 Sets, 610 Karten   19,672
//
// **`Archenemy` fiel still aus.** Es verlangt `magic-archenemy.mse-style`,
// und `M15-Magic-Pack` führt es nicht; dieses Programm zeichnete die
// Schemakarten dann im Kartenmaß des **Sets** — 375×523 gegen 435×621 —, und
// der Vergleich wies sie ab. Archenemy ist mit **14,517** das beste Set;
// ohne es steigt das Mittel um rund 0,4.
//
// **Und `Multicolor` ist dazugekommen** — 3.008 Karten, die nie gemessen
// wurden, weil MSE2 das Set nie exportiert hatte. Es liegt mit **21,202**
// über den einfarbigen und hebt das Mittel um die restlichen 0,12.
//
// **Deshalb 19,750 und nicht 19,550**: dieselbe Messung über eine **größere**
// Menge. Wer die alte Zahl vergleichen will, vergleicht sie mit 573 Karten
// aus vier Sets ohne Mehrfarbige — das ist eine andere Grundgesamtheit und
// keine bessere Anwendung.
//
// # Und am selben Abend auf 20,400 (#880)
//
// **Weil 19,750 aus einer Stichprobe stammte.** Erhoben über **610 gezogene**
// Karten, gehalten gegen **25.141** — und eine Stichprobe zieht die schlimmen
// Fälle selten:
//
//	610 Karten, 5 Sätze     19,672     Spanne 10,7 … 50,7
//	25.141 Karten, 10 Sätze 20,213     Spanne  8,3 … 73,5
//
// Das ist **keine Verschlechterung**. `Green` gibt in beiden Läufen aufs
// Tausendstel dieselbe Zahl (19,702); der Lauf ist deterministisch, und was
// sich ändert, ändert sich an der Menge.
//
// **Der Deckel gilt ab jetzt für `MSE_STICHPROBE=alle`.** Ein Lauf mit der
// Vorgabestichprobe kommt darunter heraus und sagt damit weniger — er ist der
// schnelle Blick, nicht der Beleg.
const deckelMittel = 20.400

// **Und einer auf der schlechtesten Karte**, denn ein Mittel deckt eine
// einzelne zu: eine Verschlechterung auf einer Karte verschwindet darin,
// und eine Verbesserung anderswo hebt sie sogar auf
// (#474).
//
// **Er stand nach dem Nachlauf über Phase I auf 67,000**: `Azusa’s Many
// Journeys` lag bei 66,584 und war bis dahin unsichtbar, weil ihr
// Apostroph aus dem Dateinamen der Referenz fällt (#535).
//
// **Seit den kartenweisen Stiloptionen steht er auf 42,000** (#547): die
// Karte ist eine doppelseitige Saga, und sie zeichnet jetzt eine.
//
// **Am 2026-08-20 auf 41,500 nachgezogen** (#529). Die beiden Karten aus
// dem Befund sind seither halbiert — `Kumano Faces Kakkazan` von 78,054
// auf 37,568, `Jugan Defends the Temple` von 73,607 auf 39,875 —, und der
// Kanon führte dann 41,407 als schlechteste.
//
// **Und auf 40,500 mit dem Viertelraster** (#546): die schlechteste liegt
// jetzt bei 40,297.
//
// **Am 2026-08-22 auf 51,000** (#862), und der Grund ist die **Stichprobe**
// und nicht der Code: mit fünf Sets zieht sie 135 Karten je Set statt 167,
// und damit andere. Die neue schlechteste heißt `Tempest Efreet` und liegt
// bei **50,662** — sie war in keiner der bisherigen Ziehungen.
//
// **Ein Deckel auf einer einzelnen Karte ist gegen die Ziehung spröde**,
// und das gehört ausgesprochen: er fängt eine Verschlechterung nur, wenn
// sie eine Karte trifft, die die Stichprobe zieht. Was er dafür fängt und
// das Mittel nicht, steht darüber.
//
// **Am selben Abend auf 74,000** (#880), und die 51,000 belegen die
// Sprödigkeit von eben: sie stammten aus 610 gezogenen Karten. Über alle
// 25.141 liegt die schlechteste bei **73,538** — eine blaue Karte aus
// `m15-altered`, die keine Ziehung je erwischt hatte. Die Spanne wuchs von
// 10,7 … 50,7 auf 8,3 … 73,5.
const deckelSchlechtester = 74.000
