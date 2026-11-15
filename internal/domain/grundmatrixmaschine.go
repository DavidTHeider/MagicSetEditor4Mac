package domain

// **Was die Regelmaschine durchsetzt** — Werkstücke M1 bis M20 von Phase 5.5
// (#1593).
//
// # Eine eigene Datei, und der Schnitt folgt der Frage
//
// **Nebenan stehen die drei Bereiche aus G4** — Zonen, Marken, Anhängsel:
// das, was der Tisch schon konnte, bevor es Regeln gab. Hier stehen die
// **acht**, die die Regelmaschine gebracht hat, und sie wachsen mit jedem
// Werkstück. `grundmatrix.go` stand mit beidem bei 433 Zeilen (§2).
//
// **Und der Bereich `spielsteine` steht seit `M12` wieder daneben**
// (`grundmatrixgestalten.go`): er beantwortet eine eigene Frage — welche
// Gestalt hat das, was auf dem Tisch liegt? — und diese Datei stand mit ihm
// bei 400 Zeilen.

// maschineneintraege gibt die Bereiche, die aus #1593 stammen.
func maschineneintraege() []Matrixeintrag {
	return append([]Matrixeintrag{
		{
			// **Die Zugstruktur** — Werkstück G9 von Phase 5.5
			// (#1592).
			//
			// **Zwölf Abschnitte, und jeder findet in jedem Zug statt**
			// (CR 500.1). Das Enttappen (CR 502.3), das Ziehen (CR 504.1
			// samt CR 103.8a), der Abwurf auf die Handgrenze (CR 514.1) und
			// das Entfernen des Schadens (CR 514.2) laufen von selbst — in
			// der Reihenfolge, die CR 703.3 vorgibt.
			Bereich: BereichZugstruktur,
			Grad:    UnterstuetzungAutomatisch,
			Quelle: quelle("500", "501", "502", "503", "504", "505", "506",
				"511", "512", "513", "514", "703", "103", "731"),
			Grenzen: []string{
				// **Der Enttappen-Schritt fängt mit Phasing an** (CR 502.1),
				// und Phasing braucht CR 702.26 — das ist Phase 5.8.
				"kein Phasing im Enttappen-Schritt (CR 502.1)",

				// **CR 703.4 führt fünfzehn zugbasierte Handlungen**
				// (703.4a–q, ohne `l` und `o`), und **fünf davon fehlen**.
				// Nachgezählt am 2026-09-18 gegen die Textfassung:
				// gebaut sind 703.4b/c (Tag und Nacht, `M20`), 703.4d
				// (Ziehen), 703.4i und 703.4j (Angreifer und Blocker,
				// `M6`), 703.4k und 703.4m (Kampfschaden, `M6`), 703.4n
				// (Abwerfen), 703.4p (Schaden entfernen) und 703.4q
				// (Manavorrat, `M3`).
				"fünf der fünfzehn zugbasierten Handlungen aus CR 703.4 " +
					"fehlen: Phasing (703.4a), Erzfeind (703.4e), " +
					"Saga-Marken (703.4f), Attraktionen (703.4g) und der " +
					"begrenzte Einflussbereich (703.4h)",

				// **Und CR 703.4k ist nur zur Hälfte da**: die Zuteilung
				// wird gerechnet und nicht angesagt, und bei mehreren
				// Blockern bekommt der erste alles (CR 510.1c) — das steht
				// bei `combat` ausführlich.
				"die Schadenszuteilung wird gerechnet und nicht angesagt " +
					"(CR 703.4k)",
			},
		},
		{
			// **Der Stapel und wie etwas darauf kommt** — Werkstück M4 der
			// Regelmaschine
			// (#1606),
			// auf `M1` (#1603)
			// aufbauend.
			//
			// # Warum die Sonderhandlung hier steht
			//
			// **Weil sie die andere Hälfte derselben Frage ist**: was tut
			// ein Spieler, der Priorität hat? Er wirkt etwas — das geht über
			// den Stapel (CR 601) — oder er nimmt eine Sonderhandlung vor,
			// die es gerade **nicht** tut (CR 116.1). Ein eigener Bereich
			// für eine einzige Sonderhandlung wäre eine Schublade mit einem
			// Stück darin.
			Bereich: BereichStapel,
			Grad:    UnterstuetzungAutomatisch,
			Quelle:  quelle("115", "116", "305", "405", "601", "608"),
			Grenzen: []string{
				// **Was ein Zauber tut, hängt seit `M8` davon ab, ob die
				// Karte ein Spielskript trägt** —
				// (#1610).
				// Die Grenze gilt den **Karten** und nicht dem
				// Mechanismus.
				//
				// **Kein bezogenes Paket trägt eines**: MSE kennt das Feld
				// nicht — MSE2 hat keinen Spielkern (§11). Wer eine eigene
				// Karte baut, schreibt es hinein.
				"was auf einer Karte ohne Spielskript steht, führt " +
					"niemand aus (CR 608.2c) — und kein bezogenes Paket " +
					"bringt eines mit",

				// **Vier der neun Wirkschritte fehlen** (CR 601.2b/d/e/g):
				// Modi, Wertverteilung, X und Kostenveränderungen — alle
				// stehen im Text der Karte.
				"keine Modi, kein X, keine zusätzlichen oder alternativen " +
					"Kosten (CR 601.2b, CR 601.2d, CR 601.2e, CR 601.2g)",

				// **Niemand kann etwas neutralisieren** — CR 701.5 ist eine
				// Handlung mit Schlüsselwort (M7).
				"nichts wird neutralisiert (CR 701.5)",

				// **Kein Auslöser aus einem Kartentext** — `M9`
				// (#1611)
				// baut CR 603, und die vier Fähigkeiten, die es dafür
				// gibt, stehen im **Regelwerk** (CR 725.2, CR 726.2,
				// CR 728.1). Ein Zauber, der gewirkt wird, löst nichts
				// aus, weil auf ihm nichts steht.
				"kein Auslöser, wenn etwas gewirkt wird — was auslösen " +
					"würde, steht auf der Karte (CR 603.2, M8)",

				// **Ob ein Zauber überhaupt ein Ziel verlangt**, steht in
				// seinem Text. Geprüft wird nur, ob die **genannten**
				// zulässig sind (CR 115.1) und ob sie beim Verrechnen noch
				// dort liegen (CR 608.2b).
				"welche und wie viele Ziele ein Zauber verlangt, prüft " +
					"niemand (CR 115.1)",

				// **Den Zeitpunkt prüft niemand** — PLAY-01.AK-04
				// (#1709).
				// Geprüft wird, **wer** wirkt (CR 117.1), nicht **wann**.
				// Wer die Zeitregel erzwänge, verböte ohne CR 702.8 jeden
				// Schnellzauber — und damit einen zulässigen Zug.
				"wann ein Zauber gewirkt werden darf, prüft niemand — ob " +
					"eine Hexerei in die eigene Hauptphase bei leerem " +
					"Stapel fällt (CR 307.1), halten die Spielenden",

				// **Elf der zwölf Sonderhandlungen fehlen** — jede braucht
				// einen Kartentext oder eine Variante, die #1593 ausgrenzt.
				"von den zwölf Sonderhandlungen geht nur das Landspielen " +
					"(CR 116.2a); die übrigen elf brauchen einen Kartentext",

				// **Aktivierte Fähigkeiten und Loyalität** stehen als Text
				// auf der Karte — CR 602, CR 606.
				"keine aktivierten Fähigkeiten und keine Loyalität " +
					"(CR 602, CR 606)",
			},
		},
		{
			// **Der Kampf** — Werkstück M6 der Regelmaschine
			// (#1608).
			//
			// **Die fünf Schritte standen seit G9 in der Zugfolge und waren
			// leer.** Jetzt haben drei davon Inhalt: Angreifer erklären
			// (CR 508.1), Blocker erklären (CR 509.1) und Kampfschaden
			// (CR 510) — und der geschieht **von selbst**, weil CR 510.1 und
			// CR 510.2 zugbasierte Handlungen sind.
			Bereich: BereichCombat,
			Grad:    UnterstuetzungAutomatisch,
			Quelle:  quelle("302", "506", "508", "509", "510", "511", "703"),
			Grenzen: []string{
				// **Was ein Effekt verbietet oder verlangt, weiß niemand** —
				// CR 508.1c und CR 509.1b fragen danach, und die Antwort
				// steht im Text der Karte (M8).
				"keine Einschränkung und keine Pflicht aus einer Fähigkeit " +
					"(CR 508.1c, CR 509.1b)",

				// **Zwölf von 194 Schlüsselwörtern werden gelesen** —
				// seit `M14`
				// (#1616)
				// waren es drei. **Welche zwölf, ist gezählt**: über die
				// zehn kanonischen Sets, 24.422 Karten mit Regeltext
				// (`docs/SCHLUESSELWOERTER.md`).
				"zwölf der 194 Schlüsselwortfähigkeiten aus CR 702 werden " +
					"gelesen — die übrigen stehen auf der Karte und wirken " +
					"nicht; fünfundvierzig davon meldet der Tisch als Befund",

				// **Und was sie brauchen, gibt es nicht**: `flash` (376
				// Karten) braucht die Zeitregeln aus CR 307, `devoid` (98)
				// und `protection` (127) die Farbe — und `protection` dazu
				// vier Wirkungen (CR 702.16s).
				"kein Schnellzauber, keine Farblosigkeit und kein Schutz " +
					"(CR 702.8, CR 702.114, CR 702.16) — sie brauchen die " +
					"Zeitregeln und die Farbe",

				// **Und gelesen wird der gedruckte englische Text.** Ein
				// Effekt, der Erstschlag **gibt**, steht in einem anderen
				// Kartentext (M8/M10).
				//
				// **Gelesen wird seit `M14` die Zeile und nicht das Wort**
				// (CR 702.1) — was das vorher kostete, ist gemessen: bei
				// `flying` 45 Prozent zu viel.
				// **Die halbe Grenze ist mit `M10` gefallen**
				// (#1612):
				// ein Effekt der sechsten Schicht verleiht seit da eine
				// Fähigkeit (CR 613.1f). Was bleibt, ist die Sprache.
				"gelesen wird der gedruckte englische Text — ein " +
					"deutsches Paket wird nicht verstanden",

				// **Die Schadensreihenfolge bei mehreren Blockern wählt
				// niemand** — CR 509.2 lässt den Angreifer wählen; hier
				// gilt die Reihenfolge der Erklärung.
				"bei mehreren Blockern bekommt der erste allen Schaden " +
					"(CR 509.2, CR 510.1c)",

				// **Kein Bandenkampf, keine zweite Kampfphase** — beides
				// braucht Effekte (CR 506.5, CR 702.22).
				"keine zweite Kampfphase und kein Bandenkampf " +
					"(CR 506.5, CR 702.22)",

				// **CR 506.4 nennt zwei Wege aus dem Kampf**, und seit
				// `M12` (#1614)
				// geht einer davon wirklich: wechselt die Kontrolle, ist
				// das Permanent danach weder Angreifer noch Blocker.
				//
				// **Der andere nicht.** Verlässt es das Schlachtfeld,
				// bleibt sein Eintrag im Kampfstand stehen; was zählt, ist
				// die Zuteilung, und die fragt nach der Zone (CR 510.1c/d)
				// — Schaden teilt es also keinen aus. Der Unterschied
				// zeigt sich erst, wenn jemand den Kampfstand liest.
				"ein Permanent, das das Schlachtfeld verlässt, bleibt im " +
					"Kampfstand stehen (CR 506.4) — Schaden teilt es " +
					"trotzdem keinen aus",
			},
		},
		{
			// **Die Priorität in der Gestalt, die ohne Stapel gilt.**
			//
			// **CR 500.2**: ein Abschnitt endet, wenn der Stapel leer ist
			// **und** alle nacheinander passen. Einen Stapel gibt es hier
			// nicht — er ist damit immer leer, und „alle passen
			// nacheinander" ist das einzige, was einen Abschnitt beendet.
			// Das ist keine Vereinfachung, sondern genau der Fall, den die
			// Regel beschreibt.
			// **Und sie gilt auch zu dritt und zu viert** — Werkstück M13
			// (#1615):
			// CR 800.4 lässt eine Mehrspielerpartie weiterlaufen, wenn
			// jemand geht, und die Zugreihenfolge wie die Priorität
			// **überspringen** ihn. Wer ihn mitzählte, wartete auf eine
			// Antwort, die niemand geben kann.
			Bereich: BereichPrioritaet,
			Grad:    UnterstuetzungAutomatisch,
			Quelle:  quelle("117", "500", "502", "514", "800"),
			Grenzen: []string{
				// **Der Stapel verrechnet seit `M1`** — aber was auf dem
				// Zauber **steht**, führt niemand aus (CR 608.2c).
				"beim Verrechnen geschieht nur der Zonenwechsel — was auf " +
					"dem Zauber steht, führt niemand aus (CR 608.2c)",

				// **CR 514.3a** greift, wenn zustandsbasierte Aktionen
				// anstünden oder Auslöser warteten. **Beides gibt es
				// inzwischen** — die ersten seit `M5`, die zweiten seit
				// `M9` (#1611)
				// —, und die Ausnahme ist trotzdem nicht gebaut.
				//
				// Was **wirklich** fehlt, ist der zweite Aufräumschritt.
				"im Aufräumschritt bekommt niemand Priorität, auch wenn " +
					"etwas ausgelöst hat — der zweite Aufräumschritt " +
					"fehlt (CR 514.3a)",

				// **Zehn Abschnitte vergeben Priorität** (zwölf minus
				// Enttappen und Aufräumen). `zug.abkuerzen` spart sie;
				// **verhandelt wird der Vorschlag nicht** (CR 732.2b).
				"eine Abkürzung wird ausgeführt und nicht verhandelt — das " +
					"Gegenüber kann sie nicht kürzen (CR 732.2b)",

				// **Die wahlweisen Mehrspielerregeln gibt es nicht** —
				// CR 801 bis 805 sind Optionen, die vor dem Spiel gewählt
				// werden (CR 800.2), und keine davon steht zur Wahl.
				"keine Mehrspieleroptionen: kein begrenzter Einflussbereich " +
					"(CR 801), kein Angriff nach links oder rechts " +
					"(CR 802, CR 803), keine Teams (CR 804, CR 805)",
			},
		},
		{
			// **Die Typzeile als Daten** — Werkstück M2 der Regelmaschine
			// (#1604).
			//
			// **`begleitet` und nicht `automatisch`**: das Programm
			// **liest** die Typzeile (CR 205.1) und sagt, welches Wort
			// CR 205.2a und CR 205.4a nicht kennen — es setzt keine
			// Typregel durch. Der Unterschied ist der zwischen einer
			// Warnung und einer Abweisung, und die Matrix aus G4 gibt es
			// genau dafür.
			Bereich: BereichTypen,
			Grad:    UnterstuetzungBegleitet,
			Quelle:  quelle("205", "208", "306", "310"),
			Grenzen: []string{
				// **CR 205.3 führt die Untertypen je Kartentyp in zwanzig
				// Unterabschnitten.** Gelesen wird, was dasteht; ob ein
				// „Bear" an einer Kreatur stehen darf, prüft niemand.
				"welcher Untertyp zu welchem Typ gehört, prüft niemand " +
					"(CR 205.3b–m)",

				// **CR 205.2b/205.2c** ordnen die Kartentypen; hier steht
				// die Reihenfolge, in der das Paket sie geschrieben hat.
				"die Reihenfolge der Typen wird nicht geordnet " +
					"(CR 205.2b, CR 205.2c)",

				// **Und was ein Typ bedeutet, weiß niemand.** Dass ein
				// Land keine Manakosten hat (CR 202.1b) oder ein Spontan
				// nicht auf dem Tisch bleibt (CR 304.4), setzt nichts
				// durch.
				"was ein Kartentyp bewirkt, setzt nichts durch " +
					"(CR 300–315)",

				// **Ein Effekt, der Typen ändert, fehlt** — CR 205.1b
				// verweist dafür auf die Schichten (CR 613.1d).
				"kein Effekt ändert einen Typ (CR 205.1b, CR 613.1d)",
			},
		},
		{
			// **Mana und was es kostet** — Werkstück M3 der Regelmaschine
			// (#1605).
			//
			// **`automatisch`, und die Zusage ist eng**: was hier steht,
			// setzt das Programm wirklich durch — der Vorrat leert sich am
			// Ende jedes Abschnitts, ein getapptes Land gibt nichts mehr,
			// eine Manaart, die es nicht gibt, kommt nicht hinein. **Was es
			// nicht durchsetzt, steht darunter als Grenze**, und die längste
			// davon ist, dass niemand bezahlt: dafür braucht es etwas zu
			// wirken, und das ist `M4`.
			Bereich: BereichMana,
			Grad:    UnterstuetzungAutomatisch,
			Quelle:  quelle("106", "118", "305", "500", "605"),
			Grenzen: []string{
				// **Bezahlt wird seit `M4`**
				// (#1606):
				// `Partie.manaAbziehen` zieht die Kosten aus dem Vorrat
				// (CR 601.2h).
				//
				// **Was offen bleibt, sind die Kosten neben dem Symbol**:
				// CR 601.2f–g kennt zusätzliche und alternative, und die
				// stehen auf der Karte.
				"bezahlt wird, was in der Manazeile steht (CR 601.2h) — " +
					"zusätzliche und alternative Kosten (CR 601.2f, " +
					"CR 601.2g) stehen auf der Karte und gelten nicht",

				// **Nur das Grundland hat eine Manafähigkeit.**
				"nur ein Grundland gibt Mana (CR 305.6) — jede andere " +
					"Manafähigkeit steht als Text auf der Karte und " +
					"braucht M8 (CR 605)",

				// **Drei Symbole lassen sich nicht ausgeben** — der Grund
				// steht je Symbol bei [Unbezahlbar].
				"kein X (CR 601.2b), kein Schneemana (CR 107.4h) und " +
					"kein phyrexianisches Bezahlen mit Leben (CR 107.4f)",

				// **Zwei Wahlen trifft das Programm.** CR 118 lässt sie dem
				// Spieler; einen Weg, ihn zu fragen, gibt es am Tisch nicht.
				"welche Art ein generisches Symbol zahlt und welche Hälfte " +
					"ein Hybridsymbol, wählt das Programm (CR 107.4e)",

				// **Die Ansage aus CR 106.4a/b ist eine Anzeige.** Der
				// Vorrat steht im Ereignis und im Tischbild; dass ihn jemand
				// **ansagt**, prüft niemand.
				"dass ein übriger Vorrat angesagt wird, prüft niemand " +
					"(CR 106.4a, CR 106.4b)",

				// **Kostenveränderungen sind Effekte.**
				"keine zusätzlichen oder verändernden Kosten (CR 601.2f)",
			},
		},
	}, gestalteneintrag(), faehigkeitseintrag(), zustandseintrag())
}
