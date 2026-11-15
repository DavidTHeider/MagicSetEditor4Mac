// **Zugriffstoken über REST** — Werkstück 10 von Phase 4.12
// (#1351),
// Punkt 13.
//
// **Vier Wege, alle erweiternd** (§16): sie kommen hinzu, keiner ändert einen
// bestehenden.
//
// # Und der Zugang, der sie verwaltet, ist nicht der, den sie öffnen
//
// **Wer hier ein Token anlegt, weist sich mit der Prozessmarke aus** — oder
// mit einem Token, dessen Erlaubnis `schreiben` lautet. Das zweite ist
// gewollt: ein Automat, der schreiben darf, darf sich auch selbst ein neues
// Token einrichten. Wer das nicht will, gibt ihm eines mit `lesen`.

package rest

import "net/http"

// Token ist ein Zugriffstoken, wie der Dienst es ausgibt — **ohne sein
// Geheimnis**.
type Token struct {
	ID       string `json:"id"`
	ProfilID string `json:"profilID"`

	Bezeichnung string `json:"bezeichnung"`
	Principal   string `json:"principal"`
	Scope       string `json:"scope"`
	Ablauf      string `json:"ablauf"`
	Angelegt    string `json:"angelegt"`
	Abgelaufen  bool   `json:"abgelaufen"`

	// **Werkstück O2 von Phase 5.2** (#1426).
	LetzteNutzung string `json:"letzteNutzung"`
	VorgaengerBis string `json:"vorgaengerBis"`

	// Partie und Sitz nennen die Bindung an einen Sitz (#1819).
	Partie string `json:"partie"`
	Sitz   string `json:"sitz"`

	Fassung string `json:"fassung"`
}

// Tokenangaben ist, was ein Aufruf an einem Token setzen darf.
//
// **Die Kennung steht im Weg und nicht im Rumpf** (`json:"-"`) — dieselbe
// Bauform wie bei [Profilangaben].
type Tokenangaben struct {
	TokenID  string `json:"-"`
	ProfilID string `json:"-"`

	Bezeichnung string `json:"bezeichnung"`
	Principal   string `json:"principal"`
	Scope       string `json:"scope"`
	Ablauf      string `json:"ablauf"`

	// Partie und Sitz binden ein neues Token an einen Sitz (#1819).
	Partie string `json:"partie"`
	Sitz   string `json:"sitz"`

	Fassung string `json:"fassung"`
}

// token gibt die Zugriffstoken eines Profils.
func (d *Dienst) token(schreiben http.ResponseWriter, anfrage *http.Request) {
	zeilen, fehler := d.naehte.Token.Token(
		anfrage.Context(), anfrage.PathValue("profilID"))
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusOK, map[string]any{"token": zeilen})
}

// tokenAnlegen legt eines an und gibt seinen Klartext **einmal**.
//
// **`201` und nicht `200`** — es entsteht etwas, und der Weg dorthin steht in
// der Antwort. Dieselbe Bauform wie beim Anlegen eines Profils.
func (d *Dienst) tokenAnlegen(schreiben http.ResponseWriter, anfrage *http.Request) {
	var rumpf Tokenangaben

	if !gelesen(schreiben, anfrage, &rumpf) {
		return
	}

	rumpf.ProfilID = anfrage.PathValue("profilID")

	angelegt, klartext, fehler := d.naehte.Token.TokenAnlegen(anfrage.Context(), rumpf)
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	// **`klartext` steht neben `token` und nicht darin.** In der Zeile hätte
	// er ein Feld, das bei jeder anderen Antwort leer wäre — und ein leeres
	// Feld liest sich wie „es gibt keins".
	ja(schreiben, http.StatusCreated,
		map[string]any{"token": angelegt, "klartext": klartext})
}

// tokenAendern schreibt Bezeichnung, Erlaubnis und Ablauf fort.
func (d *Dienst) tokenAendern(schreiben http.ResponseWriter, anfrage *http.Request) {
	var rumpf Tokenangaben

	if !gelesen(schreiben, anfrage, &rumpf) {
		return
	}

	rumpf.TokenID = anfrage.PathValue("tokenID")

	geaendert, fehler := d.naehte.Token.TokenAendern(anfrage.Context(), rumpf)
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusOK, map[string]any{"token": geaendert})
}

// tokenLoeschen zieht eines zurück.
func (d *Dienst) tokenLoeschen(schreiben http.ResponseWriter, anfrage *http.Request) {
	kennung := anfrage.PathValue("tokenID")

	if fehler := d.naehte.Token.TokenLoeschen(anfrage.Context(), kennung); fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusOK, map[string]any{"geloescht": kennung})
}

// tokenRotieren gibt einem Token ein neues Geheimnis — Werkstück O2 von
// Phase 5.2 (#1426). **`201` wie beim Anlegen**: es entsteht ein Geheimnis,
// und es steht nur in dieser Antwort.
func (d *Dienst) tokenRotieren(schreiben http.ResponseWriter, anfrage *http.Request) {
	rotiert, klartext, fehler := d.naehte.Token.TokenRotieren(
		anfrage.Context(), anfrage.PathValue("tokenID"))
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusCreated, map[string]any{"token": rotiert, "klartext": klartext})
}

// tokenAlleWiderrufen zieht alle Token eines Profils zurück.
func (d *Dienst) tokenAlleWiderrufen(schreiben http.ResponseWriter, anfrage *http.Request) {
	gezogen, fehler := d.naehte.Token.TokenAlleWiderrufen(
		anfrage.Context(), anfrage.PathValue("profilID"))
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusOK, map[string]any{"gezogen": gezogen})
}

// tokenAblaufwahlen nennt die Vorschläge für den Ablauf.
func (d *Dienst) tokenAblaufwahlen(schreiben http.ResponseWriter, anfrage *http.Request) {
	wahlen, fehler := d.naehte.Token.Tokenablaufwahlen(anfrage.Context())
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusOK, map[string]any{"wahlen": wahlen})
}
