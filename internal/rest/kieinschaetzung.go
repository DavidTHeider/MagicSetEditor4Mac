package rest

// **Der belegte Regelhinweis über den Dienst** — Werkstück A13.2a von Phase
// 5.3 (#1452).
//
// **`POST` und nicht `GET`**, obwohl es keine Karte ändert: es fragt ein
// Modell, bucht einen Kostenposten und kostet Zeit — dieselbe Einordnung wie
// beim Vorschlag. Ein `GET` gilt als folgenlos, und das wäre es nicht.

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// KIEinschaetzungsbezug ist, was für eine Einschätzung gebraucht wird.
type KIEinschaetzungsbezug struct {
	KarteID int64 `json:"karteId"`

	// Art ist `regel` oder `balance`.
	Art string `json:"art"`

	Anbieter string `json:"anbieter"`
	Alias    string `json:"alias"`

	// Frage geht **auch** an den Suchindex — je genauer sie ist, desto
	// einschlägiger sind die Belege.
	Frage string `json:"frage"`

	// Format ist, worauf sich ein Vergleich bezieht (A13.2b, #1452).
	Format string `json:"format"`

	// Freigegeben nennt die **privaten** Korpora, die mitgesucht werden
	// dürfen — A13.4 (#1452). Leer heißt: keine, und ein Vergleich sieht
	// dann keine einzige eigene Karte.
	Freigegeben []string `json:"freigegeben"`
}

// **Die Antwort ist `any`** — dieselbe Wahl wie bei `POST /kontextpaket`
// nebenan: die Gestalt steht in `internal/app` und ist dort einmal
// beschrieben; sie hier ein zweites Mal zu tippen hieße, zwei Wahrheiten über
// dieselbe Antwort zu führen, von denen eine beim nächsten Feld veraltet.

func (d *Dienst) kieinschaetzung(schreiben http.ResponseWriter, anfrage *http.Request) {
	kennung, fehler := strconv.ParseInt(anfrage.PathValue("karteID"), 10, 64)
	if fehler != nil {
		fehlerhaft(schreiben, fmt.Errorf("%w: die Kartenkennung %q ist keine Zahl",
			domain.ErrValidation, anfrage.PathValue("karteID")))

		return
	}

	var bezug KIEinschaetzungsbezug

	if !gelesen(schreiben, anfrage, &bezug) {
		return
	}

	bezug.KarteID = kennung

	sicht, rufefehler := d.naehte.Preise.KIEinschaetzung(anfrage.Context(), bezug)
	if rufefehler != nil {
		fehlerhaft(schreiben, rufefehler)

		return
	}

	ja(schreiben, http.StatusOK, sicht)
}
