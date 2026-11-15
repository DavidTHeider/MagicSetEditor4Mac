// **Nutzer und Profile über MCP** — Werkstück 1 von Phase 4.12
// (#1351).

package mcp

import (
	"context"
	"encoding/json"
	"slices"
)

// Nutzerzeile ist ein Nutzer samt seinen Profilen, wie MCP ihn ausgibt.
type Nutzerzeile struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Bild string `json:"bild"`

	// **Die Angaben aus Werkstück 8** (#1351) — alle freiwillig, dazu die
	// Fassung, die beim Ändern zurückzugeben ist.
	Anrede       string `json:"anrede"`
	Namenszusatz string `json:"namenszusatz"`
	Vorname      string `json:"vorname"`
	Weiterenamen string `json:"weiterenamen"`
	Nachname     string `json:"nachname"`
	Slogan       string `json:"slogan"`
	Kurzebio     string `json:"kurzebio"`
	Langebio     string `json:"langebio"`
	Fassung      string `json:"fassung"`

	// **Woran die Kontoart zu erkennen ist** — Werkstück 6 (#1351).
	//
	// **Das Kennwort steht hier nicht.** Es geht herein und nie zurück; was
	// ein Sprachmodell sieht, ist die Kennung und nicht der Nachweis.
	Email   string `json:"email"`
	Telefon string `json:"telefon"`
	Anonym  bool   `json:"anonym"`

	Profile []Profilzeile `json:"profile"`
}

// Profilzeile ist ein Profil.
type Profilzeile struct {
	ID       string `json:"id"`
	NutzerID string `json:"nutzerID"`
	Name     string `json:"name"`

	// **Die freiwilligen Angaben** — Werkstück 9 (#1351), Punkte 10 bis 13.
	Subslogan         string `json:"subslogan"`
	Kurzebeschreibung string `json:"kurzebeschreibung"`
	Langebeschreibung string `json:"langebeschreibung"`
	Badge             string `json:"badge"`

	// **Ein technisches Profil und sein Ablauf** — Werkstück 10 (#1351).
	Technisch bool   `json:"technisch"`
	Ablauf    string `json:"ablauf"`

	// Fassung ist die, die beim Ändern zurückzugeben ist — Werkstück 8
	// (#1351).
	Fassung string `json:"fassung"`
}

// Kontextzeile sagt, wer gerade in welchem Profil als was arbeitet —
// Werkstück O1 von Phase 5.2 (#1425).
type Kontextzeile struct {
	NutzerID  string `json:"nutzerID"`
	ProfilID  string `json:"profilID"`
	EignerID  string `json:"eignerID"`
	Rolle     string `json:"rolle"`
	NurLesend bool   `json:"nurLesend"`
	Modus     string `json:"modus"`
	Revision  int64  `json:"revision"`
}

// Nutzerangaben ist, was ein Werkzeug an einem Nutzer ändern darf.
//
// **Ein Verbund und keine elf Argumente** — Werkstück 8 (#1351).
type Nutzerangaben struct {
	NutzerID     string `json:"nutzerID"`
	Name         string `json:"name"`
	Bild         string `json:"bild"`
	Anrede       string `json:"anrede"`
	Namenszusatz string `json:"namenszusatz"`
	Vorname      string `json:"vorname"`
	Weiterenamen string `json:"weiterenamen"`
	Nachname     string `json:"nachname"`
	Slogan       string `json:"slogan"`
	Kurzebio     string `json:"kurzebio"`
	Langebio     string `json:"langebio"`
	Fassung      string `json:"fassung"`
}

// Nutzerpflege ist, was der MCP-Server von den Nutzern braucht.
type Nutzerpflege interface {
	Nutzer(ctx context.Context) ([]Nutzerzeile, error)
	NutzerAnlegen(ctx context.Context, name, bild string) (Nutzerzeile, error)
	ProfilAnlegen(ctx context.Context, nutzerID, name string) (Profilzeile, error)

	// **Der Wechsel** — Werkstück 3 (#1351).
	GezeigtesProfil(ctx context.Context) (Profilzeile, error)
	ProfilWechseln(ctx context.Context, profilID, nachweis string) (Profilzeile, error)

	// **Wer arbeitet** — Werkstück 17 desselben Issues
	// (#1400).
	//
	// **Nicht aus `GezeigtesProfil` abzulesen**: wer in ein fremdes Profil
	// gewechselt ist, bleibt, wer er ist — und `GezeigtesProfil` nennt den
	// **Eigner** des Profils, in dem gearbeitet wird.
	WerArbeitet(ctx context.Context) (string, error)
	NutzerWechseln(ctx context.Context, nutzerID, nachweis string) (Profilzeile, error)

	// **Der Arbeitskontext** — Werkstück O1 von Phase 5.2
	// (#1425):
	// wer, wo, als was — in einem Zug, mit der Revision, an der ein Stand
	// von einem anderen zu unterscheiden ist.
	Arbeitskontext(ctx context.Context) (Kontextzeile, error)

	// **Die Anmeldung** — Werkstück 4 desselben Issues.
	//
	// **`Anmelden` fehlt hier mit Absicht.** Der Fluss öffnet einen Browser
	// und wartet, bis ein **Mensch** dort etwas tut; ein Sprachmodell hat
	// keinen Browser und keine Hände. Was es kann, ist nachsehen, wer
	// verbunden ist — und eine Verbindung lösen.
	Anbieterkonten(ctx context.Context, nutzerID string) ([]Kontozeile, error)
	KontoLoesen(ctx context.Context, nutzerID, anbieter string) error

	// **Das anonyme Konto** — Werkstück 6 desselben Issues.
	//
	// **Und nur dieses eine von dreien.** `KontoEintragen` und
	// `KontoAnmelden` fehlen hier mit Absicht: beide tragen ein **Kennwort**
	// im Aufruf, und ein Aufruf über MCP steht im Verlauf eines
	// Sprachmodells — auf dem Rechner, bei seinem Anbieter, in jedem
	// Mitschnitt. Was dieses Werkzeug kann, kostet kein Geheimnis.
	//
	// **Der zweite Grund ist ein Orakel**: `konto_anmelden` beantwortete die
	// Frage „stimmt dieses Kennwort?" beliebig oft.
	KontoAnonym(ctx context.Context) (Nutzerzeile, error)

	// **Der zweite Faktor** — Werkstück 7 desselben Issues.
	//
	// **Nur lesen und löschen.** Das Einrichten gäbe ein Geheimnis zurück,
	// und das stünde danach im Mitschnitt eines Sprachmodells; das
	// Bestätigen wäre ein Orakel. Siehe `faktoren.go`.
	Faktoren(ctx context.Context, nutzerID string) ([]Faktorzeile, error)
	FaktorEntfernen(ctx context.Context, nutzerID, art string) error

	// **Die Einstellungen, die dem Menschen gehören, und seine Auskunft** —
	// Werkstück 8 desselben Issues, Punkte 9, 10 und 11.
	//
	// **`Personendaten` gibt fertiges JSON.** Ihre Gestalt ist in `service`
	// entschieden, und dieses Paket darf `service` nicht kennen (§2); sie
	// hier noch einmal zu erklären hieße, dieselbe Feldliste an einer vierten
	// Stelle zu führen.
	Nutzereinstellungen(ctx context.Context, nutzerID string) (
		[]Nutzereinstellungszeile, error)
	NutzereinstellungSetzen(ctx context.Context, nutzerID, schluessel, wert string) (
		[]Nutzereinstellungszeile, error)
	Personendaten(ctx context.Context, nutzerID string) (string, error)

	// **Das Profilbild** — Werkstück 8 desselben Issues.
	//
	// **Setzen und Entfernen, nicht Anzeigen.** Ein Daten-URI im Mitschnitt
	// eines Sprachmodells ist eine Zeichenkette, die niemand ansieht —
	// dieselbe Begründung wie bei den Kartenbildern.
	NutzerbildSetzen(ctx context.Context, nutzerID, fassung, pfad string) (Nutzerzeile, error)
	NutzerbildEntfernen(ctx context.Context, nutzerID, fassung string) (Nutzerzeile, error)

	// **Das Abzeichen eines Profils** — Werkstück 9 desselben Issues,
	// Punkt 13. Es gehört dem **Profil** und liegt über dem Bild des
	// Menschen.
	ProfilbadgeSetzen(ctx context.Context, profilID, fassung, pfad string) (
		Profilzeile, error)
	ProfilbadgeEntfernen(ctx context.Context, profilID, fassung string) (
		Profilzeile, error)

	// **Ändern und Löschen** — Werkstück 5 desselben Issues.
	NutzerAendern(ctx context.Context, angaben Nutzerangaben) (Nutzerzeile, error)
	ProfilAendern(ctx context.Context, angaben Profilangaben) (Profilzeile, error)
	NutzerLoeschen(ctx context.Context, nutzerID, name string) error
	ProfilLoeschen(ctx context.Context, profilID, name string) error
}

// Profilangaben ist, was sich an einem Profil ändern lässt — Werkstück 9
// (#1351), Punkte 10 bis 13.
//
// **Ein Verbund und keine sechs Argumente**: dieselbe Bauform wie
// [Nutzerangaben] und aus demselben Grund.
type Profilangaben struct {
	ProfilID string `json:"profilID"`
	Name     string `json:"name"`

	Subslogan         string `json:"subslogan"`
	Kurzebeschreibung string `json:"kurzebeschreibung"`
	Langebeschreibung string `json:"langebeschreibung"`
	Badge             string `json:"badge"`

	// **Technisch und Ablauf** — Werkstück 10 (#1351).
	Technisch bool   `json:"technisch"`
	Ablauf    string `json:"ablauf"`

	Fassung string `json:"fassung"`
}

// Kontozeile ist ein Anbieterkonto, wie MCP es ausgibt.
type Kontozeile struct {
	Anbieter    string `json:"anbieter"`
	Anzeigename string `json:"anzeigename"`
}

// nutzerwerkzeuge gibt die lesenden und die schreibenden.
func nutzerwerkzeuge(pflege Nutzerpflege) (lesend, schreibend []Werkzeug) {
	faktorlesen, faktorloeschen := faktorwerkzeuge(pflege)
	einstellungslesen, einstellungsschreiben := nutzereinstellungswerkzeuge(pflege)

	lesend = slices.Concat(einstellungslesen, []Werkzeug{
		{
			Name: "lesen.nutzer", Art: Lesend,
			Erklaerung: "Nennt die Nutzer dieses Bestands samt ihren Profilen. Ein " +
				"Nutzer ist die Person, ein Profil eine Arbeitsumgebung dieser " +
				"Person — eigene Sets, eigene Sammlung, eigene Einstellungen.",
			Schema: objekt(map[string]any{}),
			Ruf: func(ctx context.Context, _ json.RawMessage) (string, error) {
				gelesen, err := pflege.Nutzer(ctx)
				if err != nil {
					return "", err
				}

				// **Nie `nil`** (§3): eine leere Liste ist eine Antwort,
				// `null` ist keine.
				if gelesen == nil {
					gelesen = []Nutzerzeile{}
				}

				return alsJSON(map[string]any{"nutzer": gelesen})
			},
		},
		{
			Name: "lesen.anbieterkonten", Art: Lesend,
			Erklaerung: "Nennt die Anbieterkonten eines Nutzers — GitHub, Google, " +
				"Microsoft oder X. Ein Konto sagt, WER jemand bei einem Anbieter " +
				"ist; was er hier tut, verlässt den Rechner nicht.",
			Schema: objekt(map[string]any{
				"nutzerID": text("Die Kennung des Nutzers aus »lesen.nutzer«."),
			}, "nutzerID"),
			Ruf: func(ctx context.Context, roh json.RawMessage) (string, error) {
				var argumente struct {
					NutzerID string `json:"nutzerID"`
				}

				if err := lies(roh, &argumente); err != nil {
					return "", err
				}

				gelesen, err := pflege.Anbieterkonten(ctx, argumente.NutzerID)
				if err != nil {
					return "", err
				}

				// **Nie `nil`** (§3): eine leere Liste ist eine Antwort.
				if gelesen == nil {
					gelesen = []Kontozeile{}
				}

				return alsJSON(map[string]any{"konten": gelesen})
			},
		},
		faktorlesen,
	})

	schreibend = append([]Werkzeug{
		{
			Name: "schreiben.nutzer.neu", Art: Schreibend,
			Erklaerung: "ÄNDERT den Bestand: legt einen NEUEN Nutzer an, samt seinem " +
				"ersten Profil. Der Name ist eindeutig und trägt keine Leerzeichen " +
				"am Rand — er ist das Einzige, woran sich zwei Nutzer unterscheiden " +
				"lassen. Ein vorhandener Nutzer wird NICHT überschrieben.",
			Schema: objekt(map[string]any{
				"name": text("Der Anzeigename. Eindeutig, ohne Leerzeichen am Rand."),
				"bild": text("Dateiname eines Bildes im Bildverzeichnis. Leer heißt: keines."),
			}, "name"),
			Ruf: func(ctx context.Context, roh json.RawMessage) (string, error) {
				var argumente struct {
					Name string `json:"name"`
					Bild string `json:"bild"`
				}

				if err := lies(roh, &argumente); err != nil {
					return "", err
				}

				angelegt, err := pflege.NutzerAnlegen(ctx, argumente.Name, argumente.Bild)
				if err != nil {
					return "", err
				}

				return alsJSON(map[string]any{"nutzer": angelegt})
			},
		},
		{
			Name: "schreiben.profil.neu", Art: Schreibend,
			Erklaerung: "ÄNDERT den Bestand: legt ein WEITERES Profil für einen " +
				"vorhandenen Nutzer an. Der Name ist je Nutzer eindeutig; zwei " +
				"Personen dürfen beide ein Profil »Turnier« führen.",
			Schema: objekt(map[string]any{
				"nutzerID": text("Die Kennung des Nutzers aus »lesen.nutzer«."),
				"name":     text("Der Name des Profils."),
			}, "nutzerID", "name"),
			Ruf: func(ctx context.Context, roh json.RawMessage) (string, error) {
				var argumente struct {
					NutzerID string `json:"nutzerID"`
					Name     string `json:"name"`
				}

				if err := lies(roh, &argumente); err != nil {
					return "", err
				}

				angelegt, err := pflege.ProfilAnlegen(ctx, argumente.NutzerID, argumente.Name)
				if err != nil {
					return "", err
				}

				return alsJSON(map[string]any{"profil": angelegt})
			},
		},
		{
			Name: "schreiben.konto.loesen", Art: Schreibend,
			Erklaerung: "ÄNDERT den Bestand: LÖST die Verbindung eines Nutzers zu " +
				"einem Anbieter. Der Nutzer bleibt, seine Daten bleiben — nur der " +
				"Weg über diesen Anbieter fällt weg. Wer sich danach wieder " +
				"anmeldet, ist derselbe Mensch.",
			Schema: objekt(map[string]any{
				"nutzerID": text("Die Kennung des Nutzers aus »lesen.nutzer«."),
				"anbieter": text("Der Anbieter aus »lesen.anbieterkonten«."),
			}, "nutzerID", "anbieter"),
			Ruf: func(ctx context.Context, roh json.RawMessage) (string, error) {
				var argumente struct {
					NutzerID string `json:"nutzerID"`
					Anbieter string `json:"anbieter"`
				}

				if err := lies(roh, &argumente); err != nil {
					return "", err
				}

				if err := pflege.KontoLoesen(
					ctx, argumente.NutzerID, argumente.Anbieter); err != nil {
					return "", err
				}

				return alsJSON(map[string]any{"geloest": argumente.Anbieter})
			},
		},
	}, append(append(append(pflegewerkzeuge(pflege), nutzerbildwerkzeuge(pflege)...),
		einstellungsschreiben...), kontowerkzeug(pflege), faktorloeschen)...)

	// **Und die Wechselwerkzeuge daneben** (Werkstück 17, #1400): wer arbeitet
	// und worin, in einer eigenen Datei.
	wechsellesend, wechselschreibend := wechselwerkzeuge(pflege)

	return append(lesend, wechsellesend...), append(schreibend, wechselschreibend...)
}
