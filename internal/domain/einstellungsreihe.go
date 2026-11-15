package domain

// **Woraus die Einstellungsreihe wächst** — herausgelöst mit Werkstück A18
// von Phase 5.3
// (#1533).
//
// # Warum eine eigene Datei
//
// **Weil `einstellung.go` mit der Adresse je Dienst über die Zeilengrenze aus
// §2 ging.** Der Schnitt folgt der Frage: dort steht, **was eine Einstellung
// ist** — ihre Gestalt, ihre Bereiche, ihre Prüfung —, hier, **welche es
// gibt, die nicht von Hand aufgeschrieben sind**: je Dienst, je Anbieter, je
// Frage.

// init hängt die Dienstschalter an die Reihe.
//
// # Warum in `init` und nicht als Literale darüber
//
// **Damit es eine Aufzählung bleibt.** Die Dienste stehen in `dienste.go` mit
// ihrer Vorgabe und ihren Zugangsangaben; wer sie hier ein zweites Mal
// aufschriebe, hätte die Falle, die bei den Einstellungen schon zweimal
// zugeschnappt ist — eine Liste ändert sich, die andere nicht, und niemand
// sieht es (siehe den Kopf dieser Datei).
//
// **Sie stehen am Ende der Reihe**, hinter allem, was die Anwendung selbst
// betrifft: ein fremder Dienst ist die Ausnahme, nicht der Normalfall.
func init() {
	// **Zuerst der Ordner der Bildausgabe**, gleich hinter ihrem
	// Namenskonflikt, mit dem die Liste endet (AP-10, #2367).
	einstellungen = append(einstellungen, bildordnereinstellung())

	// **Und davor, wie das Fenster steht** — Werkstück O3 von Phase 5.2
	// (#1427).
	// Der Arbeitsplatz betrifft **jede Fläche des Fensters** und steht
	// deshalb vor allem, was nach draußen geht — dieselbe Ordnung nach
	// Tragweite wie in der Reihe darüber.
	haengeArbeitsplatzAn()

	// **Und gleich dahinter, wie die Listen stehen** — UX-11.AK-04
	// (#1683).
	haengeAnsichtenAn()

	// **Und welches Ziel die Ausgabe je Set zuletzt zeigte** — dieselbe Frage
	// an das Fenster (AP-10, #2367).
	einstellungen = append(einstellungen, ausgabezieleinstellung())

	// **Und was es überhaupt anbietet** — OPS-10 (#1908). Hinter dem
	// Arbeitsplatz, weil es dieselbe Frage weiter fasst: nicht wie eine
	// Fläche steht, sondern ob es sie gibt.
	einstellungen = append(einstellungen, schaltereinstellungen()...)

	// **Erst die Schalter, dann die Wahlen.** „Darf er gefragt werden" steht
	// vor „wer beantwortet was" — die zweite Frage stellt sich erst, wenn die
	// erste beantwortet ist.
	for _, zeile := range dienstzeilen {
		einstellungen = append(einstellungen, einstellung{
			Schluessel: DienstEinstellung(zeile.Dienst),
			Wirkung:    WirkungSofort,
			Bereich:    BereichDienste,
			Werte:      []string{DienstAus, DienstAn},
			Vorgabe:    zeile.Vorgabe,
		})

		// **Und die Adresse, wo es eine geben darf** — Werkstück 3 von A18
		// (#1533).
		//
		// **Gleich hinter dem Schalter**, weil es dieselbe Frage vertieft:
		// „darf er gefragt werden" und „wen fragt er dann eigentlich".
		//
		// **Ohne Werteliste, mit Regel** — wie [KIAdresse]. Leer heißt: die
		// Vorgabe des Anbieters.
		if zeile.Adresse {
			einstellungen = append(einstellungen, einstellung{
				Schluessel: DienstAdresse(zeile.Dienst),
				Wirkung:    WirkungSofort,
				Bereich:    BereichDienste,
				Pruefung:   PruefeKIAdresse,
			})
		}
	}

	// **Und die Kennung je Anbieter** — Nacharbeit zu Werkstück 4 (#1351).
	//
	// **Über [Anbieter] und nicht abgeschrieben**, aus demselben Grund wie
	// oben: eine zweite Aufzählung liefe beim fünften Anbieter auseinander.
	//
	// **Ohne Werteliste**, weil jeder Anbieter eine andere Gestalt vergibt —
	// GitHub `Ov23…`, Google eine Adresse, Microsoft eine UUID. Eine Prüfung
	// darauf wäre eine Regel, die dieses Programm sich ausdenkt, und die erste
	// Änderung beim Anbieter machte jede gültige Kennung unbrauchbar.
	//
	// **Das Geheimnis steht NICHT hier.** Es gehört in den Schlüsselbund und
	// nicht in eine Datei, die man weitergibt — dieselbe Trennung wie bei den
	// Zugangsdaten der Kartendienste
	// ([0056](../../docs/entscheidungen/0056-drei-weitere-kartendienste.md)).
	for _, anbieter := range Anbieter() {
		einstellungen = append(einstellungen, einstellung{
			Schluessel: AnbieterKennung(anbieter),
			Wirkung:    WirkungSofort,
			Bereich:    BereichAnmeldung,
		})
	}

	for _, frage := range Fragen() {
		dienste := DiensteFuer(frage)
		werte := make([]string, 0, len(dienste))

		for _, dienst := range dienste {
			werte = append(werte, string(dienst))
		}

		einstellungen = append(einstellungen, einstellung{
			Schluessel: FrageEinstellung(frage),
			Wirkung:    WirkungSofort,
			Bereich:    BereichDienste,
			Werte:      werte,

			// **Der erste ist die Vorgabe** — siehe [DiensteFuer].
			Vorgabe: werte[0],
		})
	}

	// **Und ganz am Ende die Schranken** — Werkstück O3 von Phase 5.2
	// (#1427). Sie stehen **hinter** den Dienstschaltern, weil sie über sie
	// urteilen: wer die Liste von oben nach unten liest, sieht erst, was
	// einzeln geschaltet wird, und dann, was das alles auf einmal abstellt.
	// **Das Protokoll** — OPS-04 (#1721).
	einstellungen = append(einstellungen, protokolleinstellungen()...)
	einstellungen = append(einstellungen, messeinstellungen()...)
	einstellungen = append(einstellungen, zwischenspeichereinstellungen()...)
	einstellungen = append(einstellungen, sicherungsplaneinstellungen()...)

	haengeRichtlinienAn()

	// **Und die fünf des Handels daneben** — Werkstück `T1` von Phase 5.6
	// (#1466).
	// Sie stehen in einer eigenen Datei, weil sie eine eigene Frage
	// beantworten; dass sie **hier** angehängt werden, hält sie in derselben
	// Reihe und damit in derselben Faltung.
	haengeHandelsschrankenAn()
}
