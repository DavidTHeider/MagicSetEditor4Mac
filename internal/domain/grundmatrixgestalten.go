package domain

// **Welche Gestalt hat, was auf dem Tisch liegt** — Werkstücke M7 und M12 der
// Regelmaschine
// (#1609,
// #1614).
//
// # Eine eigene Datei, und der Schnitt folgt der Frage
//
// **Nebenan stehen die sieben Bereiche, die einen Ablauf beschreiben** — Zug,
// Stapel, Kampf, Priorität, Zustand, Typen, Mana. Hier steht der eine, der
// nach einem **Ding** fragt: ein Spielstein, ein Emblem, ein verdecktes
// Permanent, eine Karte mit zwei Seiten.

// gestalteneintrag gibt den Bereich [BereichSpielsteine].
func gestalteneintrag() Matrixeintrag {
	return Matrixeintrag{
		// **Handlungen mit Schlüsselwort** — Werkstück M7 der
		// Regelmaschine
		// (#1609).
		//
		// **`begleitet` und nicht `automatisch`**, und der Unterschied
		// ist der ganze Punkt: das Programm **führt** diese Handlungen
		// aus, wenn jemand sie stellt — es **setzt** sie nicht durch.
		// Niemand ruft sie von selbst, weil kein Kartentext gelesen
		// wird (CR 608.2c). Sie sind Bausteine, die dastehen, bis `M8`
		// sie ruft.
		//
		// **Und seit `M12`
		// (#1614)
		// die Gestalten, die keine Karte brauchen**: ein Spielstein
		// entsteht (CR 111.1, CR 701.7a), ein Emblem entsteht
		// (CR 114.1), die Kontrolle wechselt (CR 110.2, CR 613.1b) —
		// ohne das Permanent zu bewegen oder zu enttappen und mit dem
		// Kampfaustritt aus CR 506.4 —, ein verdecktes Permanent ist
		// eine 2/2 (CR 708.2a), und eine zweiseitige Karte wandelt sich
		// um (CR 701.27a).
		Bereich: BereichSpielsteine,
		Grad:    UnterstuetzungBegleitet,
		Quelle: quelle("110", "111", "114", "613", "701", "704",
			"708", "712"),
		Grenzen: []string{
			// **Zehn von neunundsechzig** — acht aus `M7`, dazu das
			// Erzeugen eines Spielsteins (CR 701.7a) und das Umwandeln
			// (CR 701.27a) aus `M12`.
			"zehn der neunundsechzig Handlungen gehen (CR 701.6, 701.7, " +
				"701.8, 701.13, 701.14, 701.17, 701.21, 701.22, 701.25, " +
				"701.27); die übrigen neunundfünfzig sind Bausteine " +
				"eines Kartentextes und brauchen M8",

			// **Und keine Handlung ruft sich selbst.**
			"keine Handlung geschieht von selbst — sie wird gestellt " +
				"(CR 608.2c)",

			// **Die Reihenfolge beim Hellsehen ist keine Wahl** —
			// CR 701.22a lässt beide frei.
			"beim Hellsehen bleibt die Reihenfolge der obenbleibenden " +
				"Karten, wie sie war (CR 701.22a)",

			// **Was ein Spielstein ist, sagt der, der ihn stellt** —
			// CR 701.7a lässt es den Effekt nennen, und ein Effekt ist
			// ein Kartentext (M8).
			"ein Spielstein bekommt seine Eigenschaften genannt — " +
				"niemand liest sie von einer Karte ab (CR 701.7a)",

			// **Ein Emblem hat nur Fähigkeiten** (CR 114.3), und eine
			// Fähigkeit ist ein Kartentext.
			"ein Emblem trägt keine Fähigkeit — es steht da und wirkt " +
				"nichts (CR 114.3)",

			// **Das Umwandeln geschieht auf Zuruf.** Kein Auslöser
			// feuert (CR 701.27e), und dass die Rückseite ein Spontan
			// oder eine Hexerei wäre, prüft niemand (CR 712.10).
			"beim Umwandeln geschieht nur der Seitenwechsel — kein " +
				"Auslöser (CR 701.27e) und keine Prüfung der Rückseite " +
				"(CR 712.10)",

			// **Zwei Gestalten des Verdeckten fehlen.** Morph
			// (CR 702.37) **nennt** beim Umdrehen Eigenschaften, und
			// das ist ein Schlüsselwort (M14).
			"ein verdecktes Permanent ist immer eine 2/2 — Morph und " +
				"Manifest nennen andere Werte (CR 702.37, CR 701.34)",

			// **Acht Kartengestalten sind Kartentexte.** Jede besteht
			// aus Fähigkeiten und Zählern, und beides braucht `M8`
			// (#1610)
			// und `M9` (#1611).
			"keine Saga, kein Verlies, keine Class, keine Attraction, " +
				"kein Prototype, kein Case, kein Omen, keine Station " +
				"(CR 309, 714, 716, 717, 718, 719, 720, 721)",

			// **`objekt.kopieren` nimmt die gedruckten Werte einer
			// Vorlage an** (CR 707.2) — seit `M18`
			// (#1619).
			"eine Kopie wird gestellt und geschieht nicht von selbst " +
				"(CR 707.2) — welcher Effekt kopiert, steht auf einer " +
				"Karte (M8)",
		},
	}
}
