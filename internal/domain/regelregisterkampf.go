// **Das Regelregister, zweite Hälfte** — Kampf, Zustandsaktionen und das
// Zusammenspiel von Effekten (PLAY-01.AK-03,
// #1709).

package domain

// regelnDesKampfs gibt die letzten drei Gruppen des Katalogs.
func regelnDesKampfs() []Regeleintrag {
	return append(kampfregeln(), append(zustandsregeln(), effektregeln()...)...)
}

// kampfregeln sind die Regeln des Kampfes.
func kampfregeln() []Regeleintrag {
	return []Regeleintrag{
		{
			Nummer: "508.1a", Fassung: Grundprofilquelle,
			Katalog: KatalogKampf, Bereich: BereichCombat,
			Beschreibung: "Angreifen darf nur eine ungetappte Kreatur des aktiven Spielers",
			Stand:        RegelDurchgesetzt,
			Stelle:       "domain.Partie.AngreiferErklaeren",
			Aufrufer:     "domain.Partie.AngreiferErklaeren",
			Positiv:      "TestEinAngreiferWirdErklaertUndGetappt",
			Negativ:      "TestWerGetapptIstGreiftNichtAn",
		},
		{
			Nummer: "302.6", Fassung: Grundprofilquelle,
			Katalog: KatalogKampf, Bereich: BereichCombat,
			Beschreibung: "Eine Kreatur, die erst in diesem Zug unter die Kontrolle ihres " +
				"Spielers kam, greift nicht an",
			Stand:           RegelDurchgesetzt,
			Voraussetzungen: []string{"508.1a"},
			Stelle:          "domain.Partie.angreiferstelle",
			Aufrufer:        "domain.Partie.AngreiferErklaeren",
			Positiv:         "TestEinAngreiferWirdErklaertUndGetappt",
			Negativ:         "TestWasInDiesemZugKamGreiftNichtAn",
		},
		{
			Nummer: "509.1a", Fassung: Grundprofilquelle,
			Katalog: KatalogKampf, Bereich: BereichCombat,
			Beschreibung: "Blocken darf eine ungetappte Kreatur des Verteidigers, und nur " +
				"einen erklärten Angreifer",
			Stand:    RegelDurchgesetzt,
			Stelle:   "domain.Partie.blockPruefen",
			Aufrufer: "domain.Partie.BlockerErklaeren",
			Positiv:  "TestEinBlockerWirdErklaertUndBleibtUngetappt",
			Negativ:  "TestJederFehlerwegDerBlockererklaerungSagtWarum",
		},
		{
			Nummer: "510.2", Fassung: Grundprofilquelle,
			Katalog: KatalogKampf, Bereich: BereichCombat,
			Beschreibung: "Der zugeteilte Kampfschaden wird gleichzeitig ausgeteilt",
			Stand:        RegelDurchgesetzt,
			Stelle:       "domain.Partie.kampfschaden",
			Aufrufer:     "domain.Partie.Weiter",
			Positiv:      "TestEineDreiDreiGegenEineZweiZwei",
			Negativ:      "TestVerhinderterKampfschadenWirdNichtMarkiert",
		},
		{
			Nummer: "510.1b", Fassung: Grundprofilquelle,
			Katalog: KatalogKampf, Bereich: BereichCombat,
			Beschreibung: "Ein ungeblockter Angreifer teilt seinen Schaden dem zu, den er " +
				"angreift",
			Stand:    RegelDurchgesetzt,
			Stelle:   "domain.Partie.angreiferzuteilung",
			Aufrufer: "domain.Partie.Weiter",
			Positiv:  "TestEinUngeblockterAngreiferTrifftDenSpieler",
			// **Der Gegenfall ist der geblockte Angreifer**: sein Schaden
			// geht an den Blocker und nicht an den Spieler.
			Negativ: "TestEineDreiDreiGegenEineZweiZwei",
		},
		{
			Nummer: "510.1c", Fassung: Grundprofilquelle,
			Katalog: KatalogKampf, Bereich: BereichCombat,
			Beschreibung: "Ein geblockter Angreifer teilt seinen Schaden unter den " +
				"Blockern auf",
			Stand:    RegelTeilweise,
			Stelle:   "domain.Partie.angreiferzuteilung",
			Aufrufer: "domain.Partie.Weiter",
			Positiv:  "TestEineDreiDreiGegenEineZweiZwei",
			Grenze: "Bei mehreren Blockern bekommt der erste allen Schaden — die " +
				"Reihenfolge wählt niemand (CR 509.2)",
		},
	}
}

// zustandsregeln sind die zustandsbasierten Aktionen.
func zustandsregeln() []Regeleintrag {
	eintrag := func(nummer, stelle, beschreibung, positiv, negativ string) Regeleintrag {
		return Regeleintrag{
			Nummer: nummer, Fassung: Grundprofilquelle,
			Katalog: KatalogZustandsaktionen, Bereich: BereichZustandsbasiert,
			Beschreibung: beschreibung,
			Stand:        RegelDurchgesetzt,
			Stelle:       stelle,
			Aufrufer:     "domain.Partie.Zustandspruefung",
			Positiv:      positiv,
			Negativ:      negativ,
		}
	}

	return []Regeleintrag{
		eintrag("704.5a", "domain.Partie.Zustandspruefung",
			"Wer null oder weniger Leben hat, scheidet aus",
			"TestBeiNullLebenScheidetEinSitzAus", "TestEinLebenMehrBeendetNichts"),
		eintrag("704.5b", "domain.grundFuer",
			"Wer aus einer leeren Bibliothek ziehen wollte, scheidet aus — "+
				"eine leere Bibliothek allein reicht nicht",
			"TestWerAusLeererBibliothekZiehenWollteScheidetAus",
			"TestEineLeereBibliothekHaeltNichtAn"),
		eintrag("704.5c", "domain.Partie.Zustandspruefung",
			"Wer zehn oder mehr Giftmarken hat, scheidet aus",
			"TestBeiZehnVergiftungsmarkenScheidetEinSitzAus",
			"TestBeiNeunVergiftungsmarkenGeschiehtNichts"),
		eintrag("704.5g", "domain.Partie.stirbtAnWiderstandskraft",
			"Eine Kreatur mit tödlichem Schaden wird zerstört",
			"TestEineKreaturMitToedlichemSchadenWirdZerstoert",
			"TestEinSchadenUnterDerWiderstandskraftTutNichts"),
		eintrag("704.5j", "domain.Partie.legendenpruefen",
			"Beherrscht jemand zwei Legenden gleichen Namens, wählt er "+
				"eine, und die übrigen gehen in den Friedhof",
			"TestZweiLegendenVerlangenEineWahl", "TestZweiBaerenVerlangenKeineWahl"),
		{
			Nummer: "704.5m", Fassung: Grundprofilquelle,
			Katalog: KatalogZustandsaktionen, Bereich: BereichAnhaengsel,
			Beschreibung: "Eine Aura, die an nichts Zulässigem hängt, geht in den Friedhof",
			Stand:        RegelDurchgesetzt,
			Stelle:       "domain.Partie.anhaengselaktionen",
			Aufrufer:     "domain.Partie.Zustandspruefung",
			Positiv:      "TestEineAuraAnNichtsGehtInDenFriedhof",
			Negativ:      "TestEinAnhaengselAnEinemPermanentBleibt",
		},
		{
			Nummer: "122.3", Fassung: Grundprofilquelle,
			Katalog: KatalogZustandsaktionen, Bereich: BereichMarken,
			Beschreibung: "+1/+1- und -1/-1-Marken auf demselben Permanent heben sich " +
				"paarweise auf",
			Stand:    RegelDurchgesetzt,
			Stelle:   "domain.Partie.markenVerrechnen",
			Aufrufer: "domain.Partie.Zustandspruefung",
			Positiv:  "TestPlusUndMinusMarkenHebenSichAuf",
			Negativ:  "TestMarkenAusserhalbDesTischesWerdenNichtVerrechnet",
		},
	}
}

// effektregeln sind die Regeln über das Zusammenspiel von Effekten.
func effektregeln() []Regeleintrag {
	return []Regeleintrag{
		{
			Nummer: "613.1", Fassung: Grundprofilquelle,
			Katalog: KatalogEffekte, Bereich: BereichWechselwirkungen,
			Beschreibung: "Fortdauernde Effekte wirken in Schichten; die Schicht geht vor " +
				"dem Zeitstempel",
			Stand:    RegelDurchgesetzt,
			Stelle:   "domain.Schichten",
			Aufrufer: "domain.Partie.Werte",
			Positiv:  "TestDieSchichtSchlaegtDenZeitstempel",
			Negativ:  "TestEineUnbekannteSchichtWirdAbgewiesen",
		},
		{
			Nummer: "614.1a", Fassung: Grundprofilquelle,
			Katalog: KatalogEffekte, Bereich: BereichWechselwirkungen,
			Beschreibung: "Ein Effekt, der etwas statt eines Ereignisses tun lässt, " +
				"ersetzt es — hier ins Exil statt in den Friedhof, und nur für das " +
				"Objekt, dem er gilt",
			Stand:    RegelDurchgesetzt,
			Stelle:   "domain.Partie.friedhofErsetzen",
			Aufrufer: "domain.Partie.Zerstoeren",
			Positiv:  "TestStattInDenFriedhofInsExil",
			Negativ:  "TestEinErsatzFuerEinAnderesObjektGreiftNicht",
		},
		{
			Nummer: "616.1", Fassung: Grundprofilquelle,
			Katalog: KatalogEffekte, Bereich: BereichWechselwirkungen,
			Beschreibung: "Treffen mehrere Ersetzungen dasselbe Ereignis, wird eine nach " +
				"der anderen angewandt",
			Stand:           RegelTeilweise,
			Voraussetzungen: []string{"614.1a"},
			Stelle:          "domain.Partie.ersetzungFuer",
			Aufrufer:        "domain.Partie.Anwenden",
			Positiv:         "TestNachDerErstenErsetzungGehtEsWeiter",
			Grenze: "Wer die Reihenfolge wählt, fragt das Programm nicht — bei den " +
				"heutigen Wirkungen gibt jede Reihenfolge dasselbe",
		},
	}
}
