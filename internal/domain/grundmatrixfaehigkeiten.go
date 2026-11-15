package domain

// **Was geschieht, wenn etwas ausgelöst hat** — Werkstück M9 der
// Regelmaschine
// (#1611).
//
// # Eine eigene Datei, und der Schnitt folgt der Frage
//
// **Nebenan stehen die Bereiche, die einen Ablauf beschreiben** — Zug,
// Stapel, Kampf, Priorität, Zustand, Typen, Mana —, und daneben der eine,
// der nach einem **Ding** fragt (`grundmatrixgestalten.go`). Hier steht der,
// der nach einer **Wirkung** fragt: was tut ein Objekt, während es verrechnet
// wird, und wie kommt es überhaupt dorthin.
//
// **`grundmatrixmaschine.go` stand mit ihm bei über 400 Zeilen** (§2).

// faehigkeitseintrag gibt den Bereich [BereichWechselwirkungen].
func faehigkeitseintrag() Matrixeintrag {
	return Matrixeintrag{
		Bereich: BereichWechselwirkungen,

		// **`automatisch` und nicht `begleitet`**, und der Unterschied ist
		// belegbar: die vier eingebauten Fähigkeiten lösen **von selbst**
		// aus, warten von selbst (CR 603.2), kommen von selbst auf den
		// Stapel (CR 603.3), werden von selbst verrechnet und hören von
		// selbst auf zu existieren (CR 608.2n). Niemand stellt sie.
		//
		// **Die eine Wahl darin ist die Reihung** (CR 603.3b) — und sie
		// gehört dem Spieler, weil das Regelwerk sie ihm gibt. Ein Programm,
		// das sie selbst träfe, setzte nicht mehr durch, sondern entschiede.
		Grad: UnterstuetzungAutomatisch,
		Quelle: quelle("101", "113", "603", "608", "611", "613", "707",
			"725", "726", "728"),

		Grenzen: []string{
			// **Die härteste, und sie ist dieselbe wie bei `stapel`** —
			// seit `M8` (#1610)
			// eine Grenze über die Karten und nicht über den Mechanismus.
			"was auf einer Karte ohne Spielskript steht, führt " +
				"niemand aus (CR 608.2c) — und kein bezogenes Paket " +
				"bringt eines mit",

			// **Von den vier Arten aus CR 113.3 geht eine** — die
			// Zauberfähigkeit, und zwar beim Verrechnen (CR 608.2c).
			"nur Zauberfähigkeiten (CR 113.3a) — keine aktivierten " +
				"(CR 602), keine ausgelösten aus einem Kartentext " +
				"(CR 603.1) und keine statischen (CR 604)",

			// **Ein Skript rechnet beim Verrechnen**, wie CR 608.2c es
			// verlangt
			// (#1625)
			// — dafür steht hier keine Grenze.

			// **Dreizehn benannte Wege, und keiner mehr.** Was ein
			// Skript nicht rufen kann, kann es nicht tun — die Liste
			// steht bei [Wirkungsarten]; `M10` und `M11` haben vier
			// dazugelegt. **Die Zahl steht unter einem Wächter**
			// (`zugmatrix_test.go`), weil sie hier schon einmal auf
			// „neun" stehengeblieben ist.
			"ein Spielskript kennt dreizehn Wirkungen — alles andere aus " +
				"CR 609 und CR 610 fehlt, und verknüpfte Fähigkeiten " +
				"(CR 607) gibt es nicht",

			// **CR 101.1 und CR 101.3 haben keinen Gegenstand** — der
			// Grund steht bei [Partie.Anwenden].
			"die goldenen Regeln greifen zur Hälfte: CR 101.2 lässt aus, " +
				"was nicht geht; CR 101.1 und CR 101.3 haben nichts, " +
				"worüber sie entscheiden könnten",

			// **Vier Fähigkeiten, und alle vier stehen im Regelwerk.**
			// CR 113.8 nennt ausdrücklich eine Ausnahme: sie haben keine
			// Quelle. Alles andere ist Kartentext.
			"vier ausgelöste Fähigkeiten gibt es, und alle stehen im " +
				"Regelwerk (CR 725.2, CR 726.2, CR 728.1) — jede weitere " +
				"ist ein Kartentext (CR 603.1, M8)",

			// **Zwei der drei aus CR 726.2 fehlen**, und beide aus
			// demselben Grund: sie lassen in den Undercity vorstoßen, und
			// ein Verlies ist eine Karte (CR 714).
			"zwei der drei Fähigkeiten der Initiative fehlen — sie stoßen " +
				"in den Undercity vor, und ein Verlies ist eine Karte " +
				"(CR 726.2 mit CR 701.49, CR 714)",

			// **Drei Gestalten von Auslösern fehlen**, und jede entsteht
			// aus einem Kartentext oder aus dem Verrechnen eines Zaubers.
			//
			// **CR 603.12 und nicht CR 603.11**: das Issue nennt für die
			// reflexiven die 603.11, und die handelt von verknüpften
			// Fähigkeiten (CR 607). Nachgelesen am 2026-09-18 (§11).
			"keine Zustandsauslöser (CR 603.8), keine verzögerten " +
				"(CR 603.7) und keine reflexiven (CR 603.12)",

			// **CR 603.3b ist ein Zweischritt**, und gebaut ist der erste.
			// Der zweite betrifft Fähigkeiten, die auslösen, **weil** eine
			// andere Fähigkeit ausgelöst hat — die gibt es hier nicht, und
			// deshalb ist der Zweischritt heute nicht unterscheidbar.
			"die Reihung kennt nur den ersten Teil von CR 603.3b — " +
				"Fähigkeiten, die auf das Auslösen anderer auslösen, gibt " +
				"es nicht",

			// **Kopiert wird auf Zuruf** — `M18` baut CR 707, aber was
			// kopiert, steht auf einer Karte.
			"eine Kopie wird gestellt und geschieht nicht von selbst — " +
				"was kopiert, steht auf einer Karte (CR 707.2, M8)",

			// **Zwei der zehn Schichten wirken** — CR 613.1 (M10,
			// #1612):
			// die sechste (Fähigkeiten, CR 613.1f) und die siebte (Werte,
			// CR 613.1g mit CR 613.4b/c). Die übrigen acht haben keinen
			// Effekt, der sie erzeugte — Kopien und Kontrolle ändern das
			// Objekt unmittelbar, Text, Typ und Farbe gibt es nicht.
			"acht der zehn Schichten aus CR 613.1 wirken nicht — es gibt " +
				"keinen Effekt, der sie erzeugte",

			// **CR 613.8 hat keinen Fall** — der Grund steht ausführlich
			// im Kopf von `dauereffekt.go`: eine Abhängigkeit entsteht nur
			// **innerhalb** einer Schicht (CR 613.8a), und dort nennt jeder
			// Effekt sein Objekt bei der Kennung.
			"keine Abhängigkeit zwischen Effekten (CR 613.8) — innerhalb " +
				"einer Schicht nennt jeder Effekt sein Objekt, und keiner " +
				"nimmt einem anderen die Wirkung",

			// **Und ein Effekt nennt sein Objekt.**
			"ein Dauereffekt gilt einem Objekt und keiner Menge — " +
				"„alle Kreaturen, die du beherrschst" + `"` + " braucht einen " +
				"Kartentext, der eine Menge beschreibt (M8)",

			// **Zwei Fristen von vielen** — CR 611.2b nennt „solange …",
			// und das braucht eine Bedingung, die laufend geprüft wird.
			"ein Dauereffekt gilt bis zum Zugende oder unbestimmt — " +
				"„solange …" + `"` + " gibt es nicht (CR 611.2b)",

			// **Zwei Ereignisse sind ersetzbar** — CR 614.1 (M11,
			// #1613):
			// der Weg in den Friedhof und der Schaden. Was dieses
			// Programm nicht als Ereignis führt, lässt sich nicht
			// ersetzen.
			"nur zwei Ereignisse lassen sich ersetzen — der Friedhof " +
				"(CR 614.1a) und der Schaden (CR 615.1); Ziehen, Wirken " +
				"und das Hineinkommen eines Permanents nicht",

			// **CR 616.1 entscheidet der Zeitstempel und nicht der
			// Spieler** — eine benannte Abweichung (§15, Punkt 4).
			"treffen mehrere Ersetzungen zusammen, greift die älteste — " +
				"CR 616.1 ließe den Betroffenen wählen, und gefragt wird " +
				"er nicht",

			// **Und die Selbstersetzung fehlt** — CR 614.15 gibt ihr den
			// Vorrang, und sie braucht einen Kartentext, der sich selbst
			// meint (M8).
			"keine Selbstersetzung (CR 614.15)",
		},
	}
}
