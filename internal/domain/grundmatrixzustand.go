package domain

// **Was von selbst geschieht** — Werkstück M5 der Regelmaschine
// (#1607).
//
// # Eine eigene Datei, und der Schnitt folgt der Frage
//
// **Nebenan stehen die Bereiche, die einen Ablauf beschreiben** — Zug,
// Stapel, Kampf, Priorität, Typen, Mana. Hier steht der eine, der **keinen**
// hat: die zustandsbasierten Aktionen geschehen, wann immer jemand Priorität
// bekäme (CR 704.3), und sie warten auf niemanden.
//
// **`grundmatrixmaschine.go` stand mit ihm bei über 400 Zeilen** (§2) — die
// vierzehnte und die fünfzehnte Aktion aus `M12` und `M14` waren die Zeilen,
// die darüber hinausgingen.

// zustandseintrag gibt den Bereich [BereichZustandsbasiert].
func zustandseintrag() Matrixeintrag {
	return Matrixeintrag{
		// **Was von selbst geschieht** — Werkstück M5 der Regelmaschine
		// (#1607).
		//
		// **Fünfzehn der fünfundzwanzig Aktionen aus CR 704.5** laufen,
		// wann immer jemand Priorität bekäme (CR 704.3), **gleichzeitig**
		// (CR 704.4) und **so oft, bis keine mehr anliegt**: 704.5a–d,
		// f, g, h, i, j, m, n, p, q, v und w.
		//
		// **Die vierzehnte ist CR 704.5j** — die Legendenregel, seit
		// `M12` (#1614).
		// **Die fünfzehnte ist CR 704.5h** — Todesberührung, seit `M14`
		// (#1616).
		Bereich: BereichZustandsbasiert,
		Grad:    UnterstuetzungAutomatisch,
		Quelle:  quelle("704", "104", "110", "122"),
		Grenzen: []string{
			// **Zehn fehlen, und jede hängt an einem Werkstück** — 704.5e,
			// k, r, s, t, u, x, y, z und aa.
			//
			// **CR 704.5h gehört nicht dazu**: `M14`
			// (#1616)
			// liest Todesberührung, und jeder Schaden aus einer
			// solchen Quelle zerstört.
			"keine Kopien (704.5e) und keine Zählergrenze aus einer " +
				"Fähigkeit (704.5r)",

			// **Der Weltenzauber ist ein Zeitstempel** — CR 704.5k
			// fragt, was am kürzesten dasteht (CR 613.7), und einen
			// Zeitstempel gibt es erst mit `M10`
			// (#1612).
			//
			// **Die Legendenregel gehört nicht dazu**: `M12`
			// (#1614)
			// hat den Weg gebaut, jemanden zu fragen — CR 704.5j
			// hinterlässt eine offene Wahl in der Partie, und ein
			// Befehl beantwortet sie.
			"kein Weltenzauber (704.5k): er braucht einen Zeitstempel",

			// **Todesberührung ist ein Schlüsselwort** (CR 702.2), und
			// Schlüsselwörter sind M14.

			// **Saga, Dungeon, Rollen, Space Sculptor und „start your
			// engines!"** brauchen Kartengestalten, die es nicht gibt.
			"keine Saga, kein Dungeon, keine Rollen (704.5s, 704.5t, " +
				"704.5z, 704.5u, 704.5aa)",

			// **Der Beschützer einer Kampfkarte** ist eine Kampffrage.
			"kein Beschützer einer Kampfkarte (704.5x, 704.5y)",

			// **Ob eine Aura das verzaubern darf, woran sie hängt**,
			// steht in ihrem Untertyp und damit in einer Fähigkeit.
			// Geprüft wird nur, ob es das Ding gibt und ob es liegt.
			"ob ein Anhängsel dort hängen darf, prüft niemand (704.5m)",
		},
	}
}
