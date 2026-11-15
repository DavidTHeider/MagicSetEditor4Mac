package rest

// **Die fünf Wege zu einer Partie** — was der Dienst tut, wenn jemand
// anfragt.
//
// # Eine eigene Datei, und der Schnitt folgt der Frage
//
// **Nebenan steht, welche Gestalt eine Antwort hat** — die Angaben, die über
// die Leitung gehen, samt der Zusage, was sich an ihnen ändern darf
// ([0033](../../docs/entscheidungen/0033-was-ueber-fassungen-hinweg-zugesagt-wird.md)).
// Hier steht, **wer sie beantwortet**. `partie.go` stand mit beidem bei 419
// Zeilen (§2) — die Legendenwahl aus `M12`
// (#1614)
// war die Zeile, die darüber hinausging.

import "net/http"

// partieAnlegen legt eine Partie an.
func (d *Dienst) partieAnlegen(schreiben http.ResponseWriter, anfrage *http.Request) {
	var rumpf struct {
		Spiel string       `json:"spiel"`
		Wirt  string       `json:"wirt"`
		Sitze []Sitzwunsch `json:"sitze"`

		// Abgewaehlt ist wahlfrei — P1a (#2335): die Bereiche, die die Runde
		// selbst übernimmt. Ob es sie gibt, entscheidet der Kern.
		Abgewaehlt []string `json:"abgewaehlt"`
	}

	if !gelesen(schreiben, anfrage, &rumpf) {
		return
	}

	spiel, gut := pflichttext(schreiben, "spiel", rumpf.Spiel)
	if !gut {
		return
	}

	// **Zwei Sitze und nicht einer.** Die Prüfung steht auch in `domain`;
	// hier steht sie, damit der Rufer einen Satz über **sein** Feld bekommt
	// statt einer Meldung über eine Partie, die es noch nicht gibt.
	//
	// **Und das Feld steht im Satz.** Die Fehlergestalt dieses Dienstes
	// trägt `code` und `meldung` und kein `feld` — anders als die des
	// Fensters. Ein `domain.NewFieldError` verlöre den Namen hier
	// unterwegs, und der Rufer bekäme eine Abweisung ohne die Auskunft,
	// **wo** er nachsehen soll. `pflichttext` macht es eine Zeile darüber
	// genauso.
	if len(rumpf.Sitze) < 2 {
		nein(schreiben, http.StatusBadRequest, "validierung",
			"sitze: eine Partie braucht mindestens zwei — mit einem ist es "+
				"kein Spiel, sondern ein Aufbau")

		return
	}

	angaben, fehler := d.naehte.Partien.PartieAnlegen(
		anfrage.Context(), spiel, rumpf.Wirt, rumpf.Sitze, rumpf.Abgewaehlt)
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusCreated, angaben)
}

// partieZeigen sagt, wie eine Partie für einen Sitz aussieht.
//
// # Warum die Partiekennung nicht geprüft wird
//
// **Weil sie den Handler nie leer erreicht.** Gemessen am 2026-09-15 gegen
// `net/http.ServeMux` in Go 1.27: `GET /partien//` bekommt einen **307** auf
// den bereinigten Pfad, `GET /partien/` einen **404** — kein Muster trifft.
// Eine Prüfung darauf wäre ein Zweig, den kein Lauf betreten kann (§3), und
// ein `//nolint` dafür verdeckte nur, dass er sinnlos ist.
//
// **Die Sitzkennung ist etwas anderes**: sie steht in der Abfrage und fehlt
// dort leicht.
func (d *Dienst) partieZeigen(schreiben http.ResponseWriter, anfrage *http.Request) {
	fuer, gut := pflichttext(
		schreiben, "fuer", anfrage.URL.Query().Get("fuer"))
	if !gut {
		return
	}

	angaben, fehler := d.naehte.Partien.Partie(
		anfrage.Context(), anfrage.PathValue("partieID"), fuer)
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusOK, angaben)
}

// partienZeigen nennt die Partien dieses Profils.
//
// **Ohne Abfrageparameter**: welche es gibt, hängt am Profil, und das steht
// am Zugang und nicht in der Adresse.
func (d *Dienst) partienZeigen(schreiben http.ResponseWriter, anfrage *http.Request) {
	zeilen, fehler := d.naehte.Partien.Partien(anfrage.Context())
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusOK, zeilen)
}

// partieBefehl stellt einen Spielbefehl.
func (d *Dienst) partieBefehl(schreiben http.ResponseWriter, anfrage *http.Request) {
	var befehl Spielbefehl

	if !gelesen(schreiben, anfrage, &befehl) {
		return
	}

	befehl.PartieID = anfrage.PathValue("partieID")

	// **Eine Liste und keine Abbildung.** Über eine `map` gelesen nennt
	// dieselbe unvollständige Anfrage bei zwei fehlenden Feldern mal das
	// eine und mal das andere — Gos Iterationsreihenfolge ist absichtlich
	// zufällig. Genau diese Gestalt hat der Nachlauf über Phase 4.2
	// angestrichen (#1289):
	// „zweimal dieselbe Eingabe, zwei Ergebnisse".
	// **Die Partiekennung steht nicht in der Liste** — der Grund steht über
	// [Dienst.partieZeigen]: sie erreicht den Handler nie leer.
	for _, pflicht := range []struct{ feld, wert string }{
		{"absender", befehl.Absender},
		{"kennung", befehl.Kennung},
		{"art", befehl.Art},
	} {
		if _, gut := pflichttext(schreiben, pflicht.feld, pflicht.wert); !gut {
			return
		}
	}

	angaben, fehler := d.naehte.Partien.PartieBefehl(anfrage.Context(), befehl)
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusOK, angaben)
}

// partieStand pausiert, setzt fort oder beendet.
func (d *Dienst) partieStand(schreiben http.ResponseWriter, anfrage *http.Request) {
	var rumpf struct {
		Fuer  string `json:"fuer"`
		Stand string `json:"stand"`
	}

	if !gelesen(schreiben, anfrage, &rumpf) {
		return
	}

	// **Die Partiekennung wird nicht geprüft** — der Grund steht über
	// [Dienst.partieZeigen].
	fuer, gut := pflichttext(schreiben, "fuer", rumpf.Fuer)
	if !gut {
		return
	}

	angaben, fehler := d.naehte.Partien.PartieStand(
		anfrage.Context(), anfrage.PathValue("partieID"), fuer, rumpf.Stand)
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusOK, angaben)
}
