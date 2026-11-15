// **Das Regelregister, erste Hälfte** — Spielaufbau, Zonenwechsel,
// Priorität und Stapel, Ziele und Kosten (PLAY-01.AK-03,
// #1709).
//
// Was ein Eintrag ist und wann er gilt, steht in `regelregister.go`.

package domain

// regelnDesAblaufs gibt die ersten vier Gruppen des Katalogs.
func regelnDesAblaufs() []Regeleintrag {
	return []Regeleintrag{
		{
			Nummer: "103.4", Fassung: Grundprofilquelle,
			Katalog: KatalogSpielaufbau, Bereich: BereichZonen,
			Beschreibung: "Jeder Sitz beginnt mit dem Leben, das sein Regelwerk nennt — " +
				"im Grundfall zwanzig",
			Stand:    RegelDurchgesetzt,
			Stelle:   "domain.Partie.MitStartleben",
			Aufrufer: "domain.Partie.MitStartleben",
			Positiv:  "TestEinSitzFaengtMitDemLebenDesProfilsAn",
			// **Der Gegenfall ist Commander**: dort gilt CR 903.7, und
			// wer 103.4 anwendete, gäbe zwanzig statt vierzig.
			Negativ: "TestEineCommanderpartieFaengtMitVierzigLebenAn",
		},
		{
			Nummer: "103.8a", Fassung: Grundprofilquelle,
			Katalog: KatalogSpielaufbau, Bereich: BereichZugstruktur,
			Beschreibung: "Wer die Partie beginnt, überspringt das Ziehen seines ersten Zuges",
			Stand:        RegelDurchgesetzt,
			Stelle:       "domain.Partie.zugZiehen",
			Aufrufer:     "domain.Partie.Weiter",
			Positiv:      "TestWerAnfaengtZiehtImErstenZugNicht",
			Negativ:      "TestImZweitenZugZiehtAuchDerStarter",
		},
		{
			Nummer: "400.7", Fassung: Grundprofilquelle,
			Katalog: KatalogZonenwechsel, Bereich: BereichZonen,
			Beschreibung: "Ein Objekt, das die Zone wechselt, ist danach ein neues Objekt; " +
				"was am alten hing, bleibt zurück",
			Stand:    RegelDurchgesetzt,
			Stelle:   "domain.Partie.zonenwechsel",
			Aufrufer: "domain.Partie.Exilieren",
			Positiv:  "TestEinEffektAmZonenwechslerFaelltWeg",
			Negativ:  "TestEinZielDasDieZoneHaeltBleibtZulaessig",
		},
		{
			Nummer: "608.2n", Fassung: Grundprofilquelle,
			Katalog: KatalogZonenwechsel, Bereich: BereichStapel,
			Beschreibung: "Ein verrechneter Spontan- oder Hexereizauber geht in den " +
				"Friedhof seines Eigners",
			Stand:           RegelDurchgesetzt,
			Voraussetzungen: []string{"400.7"},
			Stelle:          "domain.Partie.gehtInDenFriedhof",
			Aufrufer:        "domain.Partie.Verrechnen",
			Positiv:         "TestEinSpontanzauberGehtInDenFriedhofSeinesEigners",
			// **Der Gegenfall ist CR 608.3a**: ein bleibendes Objekt kommt
			// auf den Tisch und nicht in den Friedhof.
			Negativ: "TestEinBleibendesObjektKommtAufDenTisch",
		},
		{
			Nummer: "117.1", Fassung: Grundprofilquelle,
			Katalog: KatalogPrioritaetStapel, Bereich: BereichPrioritaet,
			Beschreibung: "Wirken darf nur, wer Priorität hat",
			Stand:        RegelDurchgesetzt,
			Stelle:       "domain.Partie.Wirken",
			Aufrufer:     "domain.Partie.Wirken",
			Positiv:      "TestEinZauberWandertAufDenStapelUndKostet",
			Negativ:      "TestOhnePrioritaetWirdNichtGewirkt",
		},
		{
			Nummer: "117.4", Fassung: Grundprofilquelle,
			Katalog: KatalogPrioritaetStapel, Bereich: BereichPrioritaet,
			Beschreibung: "Ein Abschnitt endet erst, wenn alle nacheinander passen; " +
				"ein Passen gibt die Priorität nur weiter",
			Stand:    RegelDurchgesetzt,
			Stelle:   "domain.Partie.Passen",
			Aufrufer: "domain.Partie.Passen",
			Positiv:  "TestAlleNacheinanderBeendenDenAbschnitt",
			Negativ:  "TestEinPassenGibtSieWeiter",
		},
		{
			Nummer: "608.2b", Fassung: Grundprofilquelle,
			Katalog: KatalogPrioritaetStapel, Bereich: BereichStapel,
			Beschreibung: "Beim Verrechnen werden die Ziele nachgesehen; ist keines mehr " +
				"zulässig, verpufft der Zauber",
			Stand:           RegelDurchgesetzt,
			Voraussetzungen: []string{"115.1"},
			Stelle:          "domain.Partie.Verrechnen",
			Aufrufer:        "domain.Partie.Verrechnen",
			Positiv:         "TestEinZauberOhneZulaessigesZielVerpufft",
			Negativ:         "TestEinSitzBleibtBeimVerrechnenZulaessig",
		},
		{
			Nummer: "307.1", Fassung: Grundprofilquelle,
			Katalog: KatalogPrioritaetStapel, Bereich: BereichStapel,
			Beschreibung: "Ein Hexereizauber wird nur in der eigenen Hauptphase bei " +
				"leerem Stapel gewirkt",
			Stand:    RegelManuell,
			Aufrufer: "domain.Partie.Wirken",
			// **Der Lauf belegt, dass nichts abgewiesen wird**: ein zweiter
			// Zauber geht auf einen vollen Stapel. Wer hier die Zeitregel
			// erzwänge, verböte ohne CR 702.8 jeden Schnellzauber (AK-04).
			Positiv: "TestDerZweiteZauberLiegtUeberDemErsten",
			Grenze: "Den Zeitpunkt prüft das Programm nicht — die Spielenden halten " +
				"ihn; es weist deshalb auch keinen zulässigen Schnellzauber ab",
		},
		{
			Nummer: "115.1", Fassung: Grundprofilquelle,
			Katalog: KatalogZieleKosten, Bereich: BereichStapel,
			Beschreibung: "Die Ziele werden beim Ankündigen genannt, und nur, was es gibt, " +
				"kann eines sein",
			Stand:    RegelDurchgesetzt,
			Stelle:   "domain.Partie.Wirken",
			Aufrufer: "domain.Partie.Wirken",
			Positiv:  "TestEinZielWirdBeimAnkuendigenGenannt",
			Negativ:  "TestEinZielDasEsNichtGibtKommtNichtDurch",
		},
		{
			Nummer: "305.6", Fassung: Grundprofilquelle,
			Katalog: KatalogZieleKosten, Bereich: BereichMana,
			Beschreibung: "Ein Grundland gibt beim Tappen das Mana seines Grundtyps",
			Stand:        RegelDurchgesetzt,
			Stelle:       "domain.Partie.grundlandart",
			Aufrufer:     "domain.Partie.LandFuerMana",
			Positiv:      "TestJedesGrundlandGibtSeinMana",
			Negativ:      "TestEinLandOhneGrundtypGibtKeinMana",
		},
		{
			Nummer: "500.5", Fassung: Grundprofilquelle,
			Katalog: KatalogZieleKosten, Bereich: BereichMana,
			Beschreibung: "Am Ende jedes Schrittes und jeder Phase leert sich der Manavorrat",
			Stand:        RegelDurchgesetzt,
			Stelle:       "domain.Partie.Manaleeren",
			Aufrufer:     "domain.Partie.Weiter",
			Positiv:      "TestDerVorratIstAmEndeDesAbschnittsLeer",
			Negativ:      "TestOhneAbschnittswechselBleibtDerVorratStehen",
		},
		{
			Nummer: "903.8", Fassung: Grundprofilquelle,
			Katalog: KatalogZieleKosten, Bereich: BereichCommander,
			Beschreibung: "Ein Commander aus der Kommandozone kostet für jedes frühere " +
				"Wirken von dort zwei mehr",
			Stand:    RegelDurchgesetzt,
			Stelle:   "domain.Partie.mitCommandersteuer",
			Aufrufer: "domain.Partie.Wirken",
			Positiv:  "TestEinCommanderAusDerKommandozoneKostetMehr",
			Negativ:  "TestAusDerHandKostetDerCommanderNichtsDazu",
		},
		{
			Nummer: "601.2", Fassung: Grundprofilquelle,
			Katalog: KatalogZieleKosten, Bereich: BereichMana,
			Beschreibung: "Ein Zauber wird angekündigt, bezahlt und dann auf den Stapel " +
				"gelegt; scheitert die Zahlung, bleibt alles, wie es war",
			Stand:    RegelTeilweise,
			Stelle:   "domain.Partie.Wirken",
			Aufrufer: "domain.Partie.Wirken",
			Positiv:  "TestEinZauberWandertAufDenStapelUndKostet",
			Negativ:  "TestWasNichtZuZahlenIstWirdBenannt",
			Grenze: "Von den neun Schritten a bis i geht das Programm die, die ohne " +
				"Kartentext gehen; Modi, Wahlen, X und Kostenveränderungen fehlen",
		},
	}
}
