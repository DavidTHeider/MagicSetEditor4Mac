package rest

// **Die bekannten Gegenüber im Dienst** — Werkstück N1 von Phase 5.4
// (#1454).
//
// Sieben Wege, und die Methode sagt, welcher was tut:
//
//	GET    /peer/kontakte                             wer bekannt ist
//	POST   /peer/kontakte                             ein Gegenüber eintragen
//	POST   /peer/kontakte/{kennung}/bestaetigung      den Abdruck vergleichen
//	PUT    /peer/kontakte/{kennung}/sperre            ausschließen
//	DELETE /peer/kontakte/{kennung}/sperre            auf unbestätigt zurück
//	PUT    /peer/kontakte/{kennung}/zwecke/{zweck}    freigeben
//	DELETE /peer/kontakte/{kennung}/zwecke/{zweck}    zurücknehmen
//	PUT    /peer/kontakte/{kennung}/adresse           wo es erreichbar ist
//
// **Die Adresse ist eine Ressource mit genau einem Wert**, und deshalb `PUT`
// und kein `POST`: zweimal dieselbe gesetzt ist dieselbe. Eine **leere**
// löscht die eingetragene — „ich kenne keine mehr" ist eine Aussage, und ein
// `DELETE` daneben wäre ein zweiter Weg zu demselben Zustand.
//
// **Die Sperre ist eine Ressource und kein Befehl.** `PUT` setzt sie, `DELETE`
// nimmt sie — beides idempotent, und zweimal gesperrt ist gesperrt. Die
// **Bestätigung** ist ein `POST`, weil sie etwas mitbringt, das geprüft wird:
// den verglichenen Abdruck.

import (
	"context"
	"net/http"
)

// Kontaktnaht ist die Naht zu `internal/app`.
type Kontaktnaht interface {
	Peerkontakte(ctx context.Context) (Peerkontakte, error)
	PeerkontaktAufnehmen(
		ctx context.Context, kennung, name string,
	) (Peerkontakte, error)
	PeerkontaktBestaetigen(
		ctx context.Context, kennung, verglichen string,
	) (Peerkontakte, error)
	PeerkontaktSperren(
		ctx context.Context, kennung string,
	) (Peerkontakte, error)
	PeerkontaktEntsperren(
		ctx context.Context, kennung string,
	) (Peerkontakte, error)
	PeerzweckErlauben(
		ctx context.Context, kennung, zweck string,
	) (Peerkontakte, error)
	PeerzweckEntziehen(
		ctx context.Context, kennung, zweck string,
	) (Peerkontakte, error)
	PeerkontaktAdresse(
		ctx context.Context, kennung, adresse string,
	) (Peerkontakte, error)
	PeerkontaktNotiz(
		ctx context.Context, kennung, notiz string,
	) (Peerkontakte, error)

	// **Einladen, prüfen, annehmen** (NET-01, #1705).
	KontakteinladungErstellen(
		ctx context.Context, name, zweck, adresse string,
	) (Kontakteinladung, error)
	KontakteinladungPruefen(
		ctx context.Context, code string,
	) (Einladungsvorschau, error)
	KontakteinladungAnnehmen(
		ctx context.Context, annahme Einladungsannahme,
	) (Peerkontakte, error)
}

// Peerkontakt ist ein bekanntes Gegenüber, wie der Dienst es nennt.
type Peerkontakt struct {
	Kennung              string   `json:"kennung"`
	Fingerabdruck        string   `json:"fingerabdruck"`
	SelbstgewaehlterName string   `json:"selbstgewaehlterName"`
	EigenerName          string   `json:"eigenerName"`
	Stand                string   `json:"stand"`
	Gesehen              string   `json:"gesehen"`
	Rechte               []string `json:"rechte"`
	Adresse              string   `json:"adresse"`

	// Erstmals ist die erste Begegnung, RFC 3339 — leer heißt: nicht bekannt
	// (T3, #1468). Notiz ist, was dieses Profil sich notiert hat.
	Erstmals string `json:"erstmals"`
	Notiz    string `json:"notiz"`

	// Lage ist der Zustand aus NET-01.AK-03; Vergleichscode steht nur an
	// einem unbestätigten Kontakt (NET-01, #1705).
	Lage           string `json:"lage"`
	Vergleichscode string `json:"vergleichscode"`
}

// Peerkontakte ist die Liste samt den möglichen Zwecken.
type Peerkontakte struct {
	Kontakte []Peerkontakt `json:"kontakte"`
	Zwecke   []string      `json:"zwecke"`

	// Gemeldet steht nur nach einer Annahme (NET-01b).
	Gemeldet *Einloesemeldung `json:"gemeldet,omitempty"`

	// Nachladefehler sagt: **die Änderung ist geschrieben**, nur die Liste
	// danach nicht zu lesen — dann ist `kontakte` leer und der Status
	// trotzdem 200 (§7, Regel 3; #1960). Ein 4xx oder 5xx hieße „nicht
	// geschehen", und wer daraufhin wiederholt, läuft gegen den geschriebenen
	// Stand.
	//
	// **Erweiternd im Sinne von 0033**, wie [Befund.Teilstand]: ein Feld
	// kommt hinzu, und wo nichts fehlt, steht es nicht da.
	Nachladefehler *Befund `json:"nachladefehler,omitempty"`
}

// kontakteZeigen beantwortet `GET /peer/kontakte`.
func (d *Dienst) kontakteZeigen(
	schreiben http.ResponseWriter, anfrage *http.Request,
) {
	d.kontakteAntworten(schreiben, func(ctx context.Context) (Peerkontakte, error) {
		return d.naehte.Kontakte.Peerkontakte(ctx)
	}, anfrage)
}

// kontaktAufnehmen beantwortet `POST /peer/kontakte`.
func (d *Dienst) kontaktAufnehmen(
	schreiben http.ResponseWriter, anfrage *http.Request,
) {
	var wunsch struct {
		Kennung string `json:"kennung"`
		Name    string `json:"name"`
	}

	if !gelesen(schreiben, anfrage, &wunsch) {
		return
	}

	d.kontakteAntworten(schreiben, func(ctx context.Context) (Peerkontakte, error) {
		return d.naehte.Kontakte.PeerkontaktAufnehmen(
			ctx, wunsch.Kennung, wunsch.Name)
	}, anfrage)
}

// kontaktBestaetigen beantwortet `POST /peer/kontakte/{kennung}/bestaetigung`.
func (d *Dienst) kontaktBestaetigen(
	schreiben http.ResponseWriter, anfrage *http.Request,
) {
	var wunsch struct {
		Fingerabdruck string `json:"fingerabdruck"`
	}

	if !gelesen(schreiben, anfrage, &wunsch) {
		return
	}

	d.kontakteAntworten(schreiben, func(ctx context.Context) (Peerkontakte, error) {
		return d.naehte.Kontakte.PeerkontaktBestaetigen(
			ctx, anfrage.PathValue("kennung"), wunsch.Fingerabdruck)
	}, anfrage)
}

// kontaktSperren beantwortet `PUT /peer/kontakte/{kennung}/sperre`.
func (d *Dienst) kontaktSperren(
	schreiben http.ResponseWriter, anfrage *http.Request,
) {
	d.kontakteAntworten(schreiben, func(ctx context.Context) (Peerkontakte, error) {
		return d.naehte.Kontakte.PeerkontaktSperren(
			ctx, anfrage.PathValue("kennung"))
	}, anfrage)
}

// kontaktEntsperren beantwortet `DELETE /peer/kontakte/{kennung}/sperre`.
func (d *Dienst) kontaktEntsperren(
	schreiben http.ResponseWriter, anfrage *http.Request,
) {
	d.kontakteAntworten(schreiben, func(ctx context.Context) (Peerkontakte, error) {
		return d.naehte.Kontakte.PeerkontaktEntsperren(
			ctx, anfrage.PathValue("kennung"))
	}, anfrage)
}

// zweckErlauben beantwortet `PUT /peer/kontakte/{kennung}/zwecke/{zweck}`.
func (d *Dienst) zweckErlauben(
	schreiben http.ResponseWriter, anfrage *http.Request,
) {
	d.kontakteAntworten(schreiben, func(ctx context.Context) (Peerkontakte, error) {
		return d.naehte.Kontakte.PeerzweckErlauben(
			ctx, anfrage.PathValue("kennung"), anfrage.PathValue("zweck"))
	}, anfrage)
}

// zweckEntziehen beantwortet `DELETE /peer/kontakte/{kennung}/zwecke/{zweck}`.
func (d *Dienst) zweckEntziehen(
	schreiben http.ResponseWriter, anfrage *http.Request,
) {
	d.kontakteAntworten(schreiben, func(ctx context.Context) (Peerkontakte, error) {
		return d.naehte.Kontakte.PeerzweckEntziehen(
			ctx, anfrage.PathValue("kennung"), anfrage.PathValue("zweck"))
	}, anfrage)
}

// kontaktAdresse beantwortet `PUT /peer/kontakte/{kennung}/adresse`.
//
// **Eine leere Adresse löscht die eingetragene** — sie steht als Feld da und
// nicht als fehlendes, damit das Löschen eine Angabe bleibt.
func (d *Dienst) kontaktAdresse(
	schreiben http.ResponseWriter, anfrage *http.Request,
) {
	var wunsch struct {
		Adresse string `json:"adresse"`
	}

	if !gelesen(schreiben, anfrage, &wunsch) {
		return
	}

	d.kontakteAntworten(schreiben, func(ctx context.Context) (Peerkontakte, error) {
		return d.naehte.Kontakte.PeerkontaktAdresse(
			ctx, anfrage.PathValue("kennung"), wunsch.Adresse)
	}, anfrage)
}

// kontaktNotiz beantwortet `PUT /peer/kontakte/{kennung}/notiz`.
func (d *Dienst) kontaktNotiz(
	schreiben http.ResponseWriter, anfrage *http.Request,
) {
	var wunsch struct {
		Notiz string `json:"notiz"`
	}

	if !gelesen(schreiben, anfrage, &wunsch) {
		return
	}

	d.kontakteAntworten(schreiben, func(ctx context.Context) (Peerkontakte, error) {
		return d.naehte.Kontakte.PeerkontaktNotiz(
			ctx, anfrage.PathValue("kennung"), wunsch.Notiz)
	}, anfrage)
}

// kontakteAntworten fährt einen der acht Wege und schreibt die Liste.
func (d *Dienst) kontakteAntworten(
	schreiben http.ResponseWriter,
	fahre func(context.Context) (Peerkontakte, error), anfrage *http.Request,
) {
	kontakte, fehler := fahre(anfrage.Context())
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	// **Leere Listen und kein `null`** — sonst verschwänden sie aus der
	// Antwort, und wer sie liest, hielte „keine Zwecke" für „keine Angabe"
	// (§3, „zwischen den Prüfebenen").
	if kontakte.Kontakte == nil {
		kontakte.Kontakte = []Peerkontakt{}
	}

	if kontakte.Zwecke == nil {
		kontakte.Zwecke = []string{}
	}

	ja(schreiben, http.StatusOK, kontakte)
}
