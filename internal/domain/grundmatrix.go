package domain

// **Was dieses Programm von den Spielregeln durchsetzt** — Werkstück G4 von
// Phase 5.5
// (#1461).
//
// # Eine eigene Datei, und der Schnitt folgt der Frage
//
// **Nebenan steht, welches Profil gilt** — Fassung, Quelle, Startleben,
// Spielerzahl. Hier steht, **was davon durchgesetzt wird**: je Bereich ein
// Grad, eine Quelle und die Grenzen, an denen die Zusage endet.
//
// **Sie wächst mit jedem Werkstück der Regelmaschine**, und `grundprofil.go`
// stand mit ihr bei 460 Zeilen (§2).

// quelle gibt eine Regelquelle auf die Comprehensive Rules.
//
// **Auf Paketebene und nicht als Abschluss in `grundmatrix`**, seit die
// Einträge der Regelmaschine daneben stehen (`grundmatrixmaschine.go`): zwei
// gleichlautende Abschlüsse wären zwei Stellen, an denen die Fassung
// auseinanderlaufen kann.
func quelle(nummern ...string) Regelquelle {
	return Regelquelle{
		Dokument: DokumentCR,
		Fassung:  Grundprofilquelle,
		Nummern:  nummern,
	}
}

// grundmatrix gibt die Unterstützungsmatrix des Grundprofils.
func grundmatrix() Regelmatrix {
	eintraege := []Matrixeintrag{
		{
			Bereich: BereichZonen,
			Grad:    UnterstuetzungAutomatisch,
			Quelle:  quelle("400"),
			Grenzen: []string{
				"kein Auslöser, wenn eine Karte die Zone wechselt",
				"keine Zusatzzonen über die sieben hinaus",

				// **Ein Zonenwechsel macht ein neues Objekt** (CR 400.7) —
				// seit `M18` (#1619)
				// an jeder der zwölf Stellen und auch beim Schieben von
				// Hand.
				//
				// **Was bleibt, ist die Kennung** — sie ist die Identität
				// in diesem Programm und nicht die des Regelwerks. Ohne
				// sie spräche der Ereignisstrom von etwas, das niemand
				// findet.
				"die Kennung eines Objekts überlebt den Zonenwechsel — " +
					"sie ist die Identität in diesem Programm (CR 400.7)",

				// **Die zwölf Ausnahmen** aus CR 400.7a–m sprechen alle
				// von Effekten, Fähigkeiten und Auslösern **eines
				// Quellobjekts**. `M9`
				// (#1611)
				// bringt ausgelöste Fähigkeiten, und keine davon hilft
				// hier: die vier eingebauten haben ausdrücklich **keine
				// Quelle** (CR 113.8), und ohne Quelle gibt es nichts,
				// das ein neues Objekt wiederfände.
				"keine der zwölf Ausnahmen aus CR 400.7a–m — sie brauchen " +
					"alle ein Quellobjekt mit Text (M8)",
			},
		},
		{
			Bereich: BereichMarken,
			Grad:    UnterstuetzungAutomatisch,
			Quelle:  quelle("122"),
			Grenzen: []string{
				"eine Marke wird gezählt, ihre Wirkung nicht gekannt",

				// **Marken überleben den Zonenwechsel nicht** — seit `M18`
				// (#1619,
				// CR 400.7). Und die zustandsbasierte Prüfung auf Marken
				// gibt es seit `M5` (CR 704.5q).
				"welche Marke was bewirkt, sagt ein Kartentext (M8)",
			},
		},
		{
			Bereich: BereichAnhaengsel,
			Grad:    UnterstuetzungAutomatisch,
			Quelle:  quelle("301", "303"),
			Grenzen: []string{
				"ob etwas dort hängen darf, prüft niemand",

				// **`M5` löst ein Anhängsel durch die Regel** (CR 704.5m),
				// und `M18` (#1619)
				// löst es beim Zonenwechsel (CR 400.7) — in beide
				// Richtungen.
				"was ein Anhängsel bewirkt, sagt ein Kartentext (M8)",
			},
		},
	}

	eintraege = append(eintraege, maschineneintraege()...)

	// **Der Rest ist `manuell` und nicht `unbekannt`** — es ist eine
	// Entscheidung und kein Versäumnis.
	for _, bereich := range Regelbereiche() {
		gefunden := false

		for _, eintrag := range eintraege {
			if eintrag.Bereich == bereich {
				gefunden = true
			}
		}

		if !gefunden {
			eintraege = append(eintraege, Matrixeintrag{
				Bereich: bereich, Grad: UnterstuetzungManuell,

				// **Ein Bereich einer Variante ist hier keine Lücke**
				// ([0123](../../docs/entscheidungen/0123-streng-zaehlt-nur-was-zur-variante-gehoert.md),
				// #1626).
				// Eine Sandkastenpartie spielt kein Commander; sie führte
				// ihn sonst als offenen Rückstand, und es fehlt keine
				// Arbeit.
				NurVariante: nurVariante(bereich),
			})
		}
	}

	return Regelmatrix{Eintraege: eintraege}
}

// nurVariante sagt, welche Bereiche einer **Spielvariante** gehören.
//
// **Einer, und er steht namentlich da** — eine Aufzählung ist ehrlicher als
// eine Regel über Namen: wer den nächsten hinzufügt, entscheidet dabei, ob
// er einer ist ([0123](../../docs/entscheidungen/0123-streng-zaehlt-nur-was-zur-variante-gehoert.md)).
func nurVariante(bereich Regelbereich) bool {
	return bereich == BereichCommander
}
