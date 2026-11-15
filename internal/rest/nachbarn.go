package rest

// **Wer im selben Netz da ist, im Dienst** — Werkstück N2 von Phase 5.4
// (#1455).
//
//	GET /peer/nachbarn    wer sich im selben Netz meldet
//
// **Nur das Lesen.** Eine Ankündigung öffnet einen Port auf **jeder**
// Schnittstelle dieses Rechners, und `mse-rest` bindet ausdrücklich nur an
// `127.0.0.1` ([0046](../../docs/entscheidungen/0046-doch-ein-dienst-auf-der-rueckschleife.md)).
// Ein Weg, über den sich darüber ein Port aufmachen ließe, höbe genau diese
// Entscheidung auf.

import (
	"context"
	"net/http"
)

// Nachbarnaht ist die Naht zu `internal/app`.
type Nachbarnaht interface {
	Nachbarn(ctx context.Context) (Nachbarn, error)
}

// Nachbar ist ein Gegenüber, das sich gemeldet hat.
type Nachbar struct {
	Kennung       string `json:"kennung"`
	Fingerabdruck string `json:"fingerabdruck"`
	Name          string `json:"name"`
	Adresse       string `json:"adresse"`

	// Merkadresse ist, was in die Kontaktadresse gehört — der `.local`-Name,
	// wo es einen gibt (#1633).
	Merkadresse string `json:"merkadresse"`

	Bekannt bool   `json:"bekannt"`
	Stand   string `json:"stand"`

	// GemerkteAdresse und AdresseGeaendert: ein bekannter Nachbar, der
	// woanders steht (NET-02.NAT-02, #1706).
	GemerkteAdresse  string `json:"gemerkteAdresse"`
	AdresseGeaendert bool   `json:"adresseGeaendert"`
}

// Nachbarn ist die Liste samt dem, was nicht zu deuten war.
type Nachbarn struct {
	Nachbarn  []Nachbar `json:"nachbarn"`
	Verworfen int       `json:"verworfen"`
}

// nachbarnZeigen beantwortet `GET /peer/nachbarn`.
func (d *Dienst) nachbarnZeigen(
	schreiben http.ResponseWriter, anfrage *http.Request,
) {
	nachbarn, fehler := d.naehte.Nachbarn.Nachbarn(anfrage.Context())
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	// **Eine leere Liste und kein `null`** — sonst verschwände sie aus der
	// Antwort, und wer sie liest, hielte „niemand da" für „keine Angabe".
	if nachbarn.Nachbarn == nil {
		nachbarn.Nachbarn = []Nachbar{}
	}

	ja(schreiben, http.StatusOK, nachbarn)
}
