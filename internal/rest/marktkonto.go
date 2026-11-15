package rest

// **Das eigene Konto im Dienst** — Werkstück `T11` von Phase 5.6
// (#1476).
//
// # Ein `GET` und keine zweite Methode
//
// **Was dieser Weg gibt, ist eine Übersicht** — und was er nicht gibt, ist
// eine Tür zum Schreiben. Der Anbieter kennt sie; gebaut ist keine
// ([0122](../../docs/entscheidungen/0122-ein-fremdes-konto-wird-gelesen.md)).
//
// **Und er hängt an zwei Schaltern**, die beide aus beginnen: dem Dienst und
// der Kontoschau. Wer über einen Port anfragt, schaltet keinen davon ein.

import (
	"context"
	"net/http"
	"strconv"
)

// Kontonaht ist, was der Dienst für die Kontoübersicht braucht.
//
// **Sie gibt `any` zurück**, wie die Listeneinfuhr daneben: die Gestalt kommt
// aus `internal/app` und wird hier nicht noch einmal geschrieben.
type Kontonaht interface {
	Marktkonto(ctx context.Context, quelle string, seite, menge int) (any, error)
}

// marktkontoZeigen gibt die Übersicht über das eigene Konto.
func (d *Dienst) marktkontoZeigen(
	schreiben http.ResponseWriter, anfrage *http.Request,
) {
	seite, gut := blaetterzahl(schreiben, anfrage, "seite")
	if !gut {
		return
	}

	menge, gutemenge := blaetterzahl(schreiben, anfrage, "menge")
	if !gutemenge {
		return
	}

	stand, fehler := d.naehte.Konto.Marktkonto(anfrage.Context(),
		anfrage.PathValue("quelle"), seite, menge)
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusOK, stand)
}

// blaetterzahl liest eine **wahlfreie** Zahl aus der Abfrage.
//
// **Neben [abfragezahl] und nicht statt ihrer**: jene verlangt den Wert und
// sagt es, wenn er fehlt — richtig, wo er die Frage trägt. Hier ist er eine
// Verfeinerung: wer nichts nennt, nimmt die Vorgabe des Anbieters, und ein
// Pflichtfeld daraus zu machen hieße, eine Blätterangabe zu erfinden.
//
// **Eine kaputte ist trotzdem ein Fehler.** „page=zwei" ist keine leere
// Angabe, sondern eine falsche, und sie still zu verwerfen gäbe eine andere
// Seite zurück als die gemeinte.
func blaetterzahl(
	schreiben http.ResponseWriter, anfrage *http.Request, name string,
) (int, bool) {
	roh := anfrage.URL.Query().Get(name)
	if roh == "" {
		return 0, true
	}

	zahl, fehler := strconv.Atoi(roh)
	if fehler != nil || zahl < 0 {
		nein(schreiben, http.StatusBadRequest, "validierung",
			name+" muss eine Zahl ab null sein, nicht \""+roh+"\"")

		return 0, false
	}

	return zahl, true
}
