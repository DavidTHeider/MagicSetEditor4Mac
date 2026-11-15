package rest

// **Die Naht zu Nutzern, Profilen und Konten** — abgetrennt von `naehte.go`,
// als die Datei mit den Anbieterkonten (#1351) über die Zeilengrenze aus §2
// kam.
//
// **Nach der Frage geschnitten**: nebenan stehen die Nähte zum **Bestand** —
// Sets, Karten, Decks, Sammlung. Hier steht, wer sie benutzt.

import (
	"context"
	"encoding/json"
)

// Fremdnutzer führt die Nutzer dieses Bestands und ihre Profile
// (#1351).
//
// **Ein Nutzer ist die Person, ein Profil eine Arbeitsumgebung dieser Person.**
type Fremdnutzer interface {
	Nutzer(ctx context.Context) ([]Nutzer, error)
	NutzerAnlegen(ctx context.Context, name, bild string) (Nutzer, error)
	ProfilAnlegen(ctx context.Context, nutzerID, name string) (Profil, error)

	// **Der Wechsel** — Werkstück 3 (#1351).
	GezeigtesProfil(ctx context.Context) (Profil, error)
	ProfilWechseln(ctx context.Context, profilID, nachweis string) (Profil, error)

	// **Wer arbeitet** — Werkstück 17 von Phase 4.12
	// (#1400).
	WerArbeitet(ctx context.Context) (string, error)
	NutzerWechseln(ctx context.Context, nutzerID, nachweis string) (Profil, error)

	// **Der Arbeitskontext** — Werkstück O1 von Phase 5.2
	// (#1425).
	Arbeitskontext(ctx context.Context) (Arbeitskontext, error)

	// **Die Anmeldung** — Werkstück 4 desselben Issues.
	//
	// **`Anmelden` fehlt hier mit Absicht.** Der Fluss öffnet einen Browser
	// auf dem Rechner, auf dem der Dienst läuft, und wartet auf einen
	// Menschen. Wer den Dienst von einem Skript aus anspricht, hat weder das
	// eine noch das andere — ein Endpunkt dafür wäre eine Zusage ohne
	// Einlösung (§10).
	Anbieterkonten(ctx context.Context, nutzerID string) ([]Konto, error)
	KontoLoesen(ctx context.Context, nutzerID, anbieter string) error

	// **Die vier Kontoarten** — Werkstück 6 desselben Issues.
	//
	// **Alle drei, anders als bei MCP.** Ein Kennwort im Rumpf einer
	// Anfrage an `127.0.0.1` geht von einem Prozess zum anderen und landet
	// in keinem Verlauf; über MCP stünde es im Mitschnitt eines
	// Sprachmodells.
	KontoEintragen(ctx context.Context, kennung, kennwort, name string) (Nutzer, error)
	KontoAnmelden(ctx context.Context, kennung, kennwort, nachweis string) (Nutzer, error)
	KontoAnonym(ctx context.Context) (Nutzer, error)

	// **Der zweite Faktor** — Werkstück 7 desselben Issues.
	//
	// **Alle vier, anders als bei MCP.** Das Einrichten gibt ein Geheimnis
	// zurück; über MCP stünde es im Mitschnitt eines Sprachmodells, hier geht
	// es von einem Prozess zum anderen auf `127.0.0.1`.
	Faktoren(ctx context.Context, nutzerID string) (Faktorstand, error)
	FaktorEinrichten(ctx context.Context, nutzerID string) (Faktoreinrichtung, error)
	FaktorBestaetigen(ctx context.Context, nutzerID, code string) error
	BiometrieEinrichten(ctx context.Context, nutzerID string) error
	FaktorEntfernen(ctx context.Context, nutzerID, art string) error

	// **Die Einstellungen eines Menschen und seine Auskunft** — Werkstück 8
	// desselben Issues, Punkte 9, 10 und 11.
	//
	// **`Personendaten` gibt fertiges JSON.** Ihre Gestalt ist in `service`
	// entschieden; sie hier noch einmal zu erklären hieße, dieselbe Feldliste
	// an einer vierten Stelle zu führen.
	Nutzereinstellungen(ctx context.Context, nutzerID string) ([]Einstellungswahl, error)
	NutzereinstellungSetzen(ctx context.Context, nutzerID, schluessel, wert string) (
		[]Einstellungswahl, error)
	Personendaten(ctx context.Context, nutzerID string) (json.RawMessage, error)

	// **Das Profilbild** — Werkstück 8 desselben Issues.
	//
	// **Setzen und Entfernen, nicht Anzeigen.** Ein Daten-URI ist die
	// Darstellung einer Liste; wer den Dienst von einem Skript aus anspricht,
	// bekommt den Dateinamen im Nutzer und kann die Datei selbst lesen.
	NutzerbildSetzen(ctx context.Context, nutzerID, fassung, pfad string) (Nutzer, error)
	NutzerbildEntfernen(ctx context.Context, nutzerID, fassung string) (Nutzer, error)

	// **Das Abzeichen eines Profils** — Werkstück 9 desselben Issues.
	ProfilbadgeSetzen(ctx context.Context, profilID, fassung, pfad string) (Profil, error)
	ProfilbadgeEntfernen(ctx context.Context, profilID, fassung string) (Profil, error)

	// **Ändern und Löschen** — Werkstück 5 desselben Issues.
	NutzerAendern(ctx context.Context, angaben Nutzerangaben) (Nutzer, error)
	ProfilAendern(ctx context.Context, angaben Profilangaben) (Profil, error)
	NutzerLoeschen(ctx context.Context, nutzerID, name string) error
	ProfilLoeschen(ctx context.Context, profilID, name string) error
}

// Fremdtoken führt die Zugriffstoken technischer Profile — Werkstück 10
// desselben Issues.
//
// **Eine eigene Naht neben [Fremdnutzer]**, weil sie eine andere Frage
// beantwortet: jene sagt, **wer** hier arbeitet, diese, **womit sich ein
// Automat ausweist**. Und weil sie eine eigene Ablage hat.
type Fremdtoken interface {
	Token(ctx context.Context, profilID string) ([]Token, error)
	TokenAnlegen(ctx context.Context, angaben Tokenangaben) (Token, string, error)
	TokenAendern(ctx context.Context, angaben Tokenangaben) (Token, error)
	TokenLoeschen(ctx context.Context, tokenID string) error

	// **Rotation, Sammelwiderruf und Ablaufwahlen** — Werkstück O2 von
	// Phase 5.2 (#1426).
	TokenRotieren(ctx context.Context, tokenID string) (Token, string, error)
	TokenAlleWiderrufen(ctx context.Context, profilID string) (int64, error)
	Tokenablaufwahlen(ctx context.Context) ([]Ablaufwahl, error)
}

// Ablaufwahl ist ein benannter Vorschlag für den Ablauf eines Tokens.
type Ablaufwahl struct {
	Name   string `json:"name"`
	Ablauf string `json:"ablauf"`
}

// Konto ist ein Anbieterkonto.
type Konto struct {
	Anbieter    string `json:"anbieter"`
	Anzeigename string `json:"anzeigename"`
}

// Einstellungswahl ist eine Einstellung samt ihrer möglichen Werte.
//
// **Die Werte kommen mit** — dieselbe Begründung wie an jeder anderen Fläche:
// ohne sie stünden die erlaubten Werte beim Rufer noch einmal, und die zweite
// Liste liefe zurück.
type Einstellungswahl struct {
	Schluessel string   `json:"schluessel"`
	Wert       string   `json:"wert"`
	Werte      []string `json:"werte"`
}
