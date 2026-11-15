package rest

// **Die Recherche am Dienst** — Werkstück A16 von Phase 5.3
// (#1515).
//
// # Drei Wege, und der dritte ist kein Schreiben
//
// **`POST /kontextpaket` legt nichts an.** Er ist ein `POST`, weil sein
// Auftrag nicht in eine Abfragezeile passt — Aufgabe, Frage, Anbieter,
// Modell, zwei Budgetzahlen —, und nicht, weil etwas entstünde. Gesendet wird
// nichts; das Paket ist die **Antwort**, nicht die Wirkung.
//
// **Das ist genau die Einordnung nach
// [0033](../../docs/entscheidungen/0033-was-ueber-fassungen-hinweg-zugesagt-wird.md)**:
// alle drei Wege sind **erweiternd** — sie kommen hinzu und ändern nichts an
// einem bestehenden.

import (
	"context"
	"net/http"
	"net/url"
)

// Rechercheauskunft ist die Naht zur Anwendung.
type Rechercheauskunft interface {
	// **`freigegeben` nennt die privaten Korpora, die gesucht werden
	// dürfen** — A13.4 (#1452). Leer heißt: keine.
	Recherche(ctx context.Context, fall, frage string, grenze int,
		sprache, korpusstand, fortsetzung string,
		freigegeben []string) (any, error)
	AehnlicheKarten(ctx context.Context, karteID int64, stand string,
		grenze int, sprache, fortsetzung string,
		freigegeben []string) (any, error)
	Kontextpaket(ctx context.Context, wunsch Kontextwunsch) (any, error)
}

// Kontextwunsch ist der Auftrag im Rumpf von `POST /kontextpaket`.
type Kontextwunsch struct {
	Aufgabe      string `json:"aufgabe"`
	Zielrevision string `json:"zielrevision"`

	Fall   string `json:"fall"`
	Frage  string `json:"frage"`
	Grenze int    `json:"grenze"`

	Anbieter string `json:"anbieter"`
	Modell   string `json:"modell"`

	// Kontextgrenze und Ausgabereserve sind das Budget — **beide Pflicht**.
	//
	// **Sie gehören zum Zielmodell, und dieser Dienst kennt es nicht**: eine
	// Vorgabe hier wäre eine Zahl, die für kein Modell gilt, und ein Paket,
	// das das Fenster sprengt, schneidet das Modell still ab.
	Kontextgrenze  int `json:"kontextgrenze"`
	Ausgabereserve int `json:"ausgabereserve"`

	HoechstensBelege int `json:"hoechstensBelege"`

	// Freigegeben nennt die **privaten** Korpora, die gesucht werden dürfen —
	// A13.4 (#1452). Leer heißt: keine.
	Freigegeben []string `json:"freigegeben"`
}

// freigabeAusAbfrage liest die Freigabe aus der Abfragezeile.
//
// **Ein Parameter je Korpus und keine Liste mit Komma** — `?freigegeben=a&
// freigegeben=b` ist die Form, die `net/url` selbst schreibt und liest; ein
// Komma darin wäre eine zweite Zerlegungsregel, die niemand erwartet.
func freigabeAusAbfrage(abfrage url.Values) []string {
	return abfrage["freigegeben"]
}

// recherche beantwortet `GET /recherche`.
func (d *Dienst) recherche(schreiben http.ResponseWriter, anfrage *http.Request) {
	abfrage := anfrage.URL.Query()

	ergebnis, fehler := d.naehte.Recherche.Recherche(anfrage.Context(),
		abfrage.Get("fall"), abfrage.Get("frage"),
		grenzeAusAbfrage(abfrage.Get("grenze")), abfrage.Get("sprache"),
		abfrage.Get("korpusstand"), abfrage.Get("fortsetzung"),
		freigabeAusAbfrage(abfrage))
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusOK, ergebnis)
}

// aehnlichekarten beantwortet `GET /recherche/aehnlich/{karte}`.
//
// **Der Stand steht in der Abfrage und nicht im Weg**: er ist eine Bedingung
// an die Antwort und kein Teil der Adresse — dieselbe Karte bleibt dieselbe
// Karte, auch wenn sie sich geändert hat.
func (d *Dienst) aehnlichekarten(
	schreiben http.ResponseWriter, anfrage *http.Request,
) {
	karteID, gut := kennung(schreiben, anfrage, "karte")
	if !gut {
		return
	}

	abfrage := anfrage.URL.Query()

	ergebnis, fehler := d.naehte.Recherche.AehnlicheKarten(anfrage.Context(),
		karteID, abfrage.Get("stand"),
		grenzeAusAbfrage(abfrage.Get("grenze")), abfrage.Get("sprache"),
		abfrage.Get("fortsetzung"), freigabeAusAbfrage(abfrage))
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusOK, ergebnis)
}

// kontextpaket beantwortet `POST /kontextpaket`.
//
// **Und sendet nichts.** Was zurückkommt, ist, was hinausginge — Belege,
// Tokenschätzung, Datenkategorien. Der Ruf an ein Modell ist ein anderer.
func (d *Dienst) kontextpaket(
	schreiben http.ResponseWriter, anfrage *http.Request,
) {
	var wunsch Kontextwunsch

	if !gelesen(schreiben, anfrage, &wunsch) {
		return
	}

	paket, fehler := d.naehte.Recherche.Kontextpaket(anfrage.Context(), wunsch)
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusOK, paket)
}
