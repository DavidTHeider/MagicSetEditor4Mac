// **Nutzer und Profile am Fenster und an den drei Flächen** — Werkstück 1 von
// Phase 4.12
// (#1351),
// nach Vermerk [0066](../../docs/entscheidungen/0066-nutzer-profile-und-die-vier-entscheidungen.md).

package app

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// Nutzerfuehrer ist die Naht zu `internal/service`.
type Nutzerfuehrer interface {
	Uebersicht(ctx context.Context) ([]service.Nutzerzeile, error)
	NutzerAnlegen(ctx context.Context, neu domain.Nutzer) (service.Nutzerzeile, error)
	ProfilAnlegen(ctx context.Context, neu domain.Profil) (domain.Profil, error)

	// **Ändern und Löschen** — Werkstück 5 (#1351).
	NutzerAendern(ctx context.Context, geaendert domain.Nutzer) (domain.Nutzer, error)
	ProfilAendern(ctx context.Context, geaendert domain.Profil) (domain.Profil, error)
	// **Die Bindung an eine Vorschau geht mit** — Befund 9 des Nachlaufs zum
	// Zweitblick über Phase 5.2. Sie ist wahlfrei: nur die Rücksetzung zeigt
	// vorher, was verschwindet, und nur sie reicht sie durch.
	NutzerLoeschen(ctx context.Context, nutzerID int64, name string,
		bindung ...func(profile []domain.Profil, umfaenge [][]domain.Betroffenes) error) error
	ProfilLoeschen(ctx context.Context, profilID int64, name string,
		bindung ...func(profile []domain.Profil, umfaenge [][]domain.Betroffenes) error) error

	// **Wer über wen schreiben darf** — ID-03 (#1702). Die Regel steht im
	// Dienst; gefragt wird an jeder Bindung, die einen Menschen oder ein
	// Profil nennt ([App.ueberNutzer]).
	UeberNutzer(ctx context.Context, ziel int64) error
	UeberProfil(ctx context.Context, profilID int64) error
}

// NutzerDTO ist ein Nutzer, wie das Fenster ihn sieht.
type NutzerDTO struct {
	ID   string `json:"id"`
	Name string `json:"name"`

	// Bild ist der Dateiname — leer heißt: keines.
	Bild string `json:"bild"`

	// **Woran die Seite die Kontoart erkennt** — Werkstück 6 (#1351).
	//
	// **Drei Felder und keine Art als Wort**: ein Nutzer kann eine Adresse
	// **und** eine Nummer haben, und eine Aufzählung müsste sich dann für
	// eine entscheiden.
	//
	// **Das Kennwort fehlt hier, und zwar mit Absicht.** Es geht herein und
	// nie zurück; `TestDasKennwortErreichtDieAntwortNie` hält das über die
	// JSON-Naht fest.
	Email   string `json:"email"`
	Telefon string `json:"telefon"`
	Anonym  bool   `json:"anonym"`

	// **Die Angaben aus Werkstück 8** (#1351) — alle freiwillig.
	Anrede       string `json:"anrede"`
	Namenszusatz string `json:"namenszusatz"`
	Vorname      string `json:"vorname"`
	Weiterenamen string `json:"weiterenamen"`
	Nachname     string `json:"nachname"`
	Slogan       string `json:"slogan"`
	Kurzebio     string `json:"kurzebio"`
	Langebio     string `json:"langebio"`

	// Fassung reist als Text, wie jede Kennung.
	Fassung string `json:"fassung"`

	Profile []ProfilDTO `json:"profile"`
}

// NutzerangabenDTO ist, was eine Fläche an einem Nutzer ändern darf.
//
// **Ein Verbund und keine elf Argumente**: die Felder gehören zusammen, und
// wer sie einzeln reicht, vertauscht irgendwann zwei. Dieselbe Bauform wie
// [ExemplarDTO] und [FormatDTO].
//
// **Was hier nicht steht, ändert man anderswo**: Adresse, Rufnummer und
// Kennwort gehören zur **Anmeldung** (Werkstück 6), die zweiten Faktoren zu
// ihrem eigenen Weg (Werkstück 7). Ein Formular, das alles zugleich ändert,
// hätte für jedes Feld eine andere Zusage.
type NutzerangabenDTO struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Bild string `json:"bild"`

	Anrede       string `json:"anrede"`
	Namenszusatz string `json:"namenszusatz"`
	Vorname      string `json:"vorname"`
	Weiterenamen string `json:"weiterenamen"`
	Nachname     string `json:"nachname"`
	Slogan       string `json:"slogan"`
	Kurzebio     string `json:"kurzebio"`
	Langebio     string `json:"langebio"`

	// Fassung ist die, die der Rufer gelesen hat.
	//
	// **Ohne sie ginge eine Änderung verloren** (§7, Regel 2 über die
	// Flächen hinweg): zwei gleichzeitige Schreibvorgänge läsen dieselbe
	// Fassung, und die zweite überschriebe die erste.
	Fassung string `json:"fassung"`
}

// ProfilDTO ist ein Profil, wie das Fenster es sieht.
type ProfilDTO struct {
	ID       string `json:"id"`
	NutzerID string `json:"nutzerID"`
	Name     string `json:"name"`

	// **Die freiwilligen Angaben** — Werkstück 9 (#1351), Punkte 10 bis 13.
	//
	// **`subslogan` und nicht `slogan`**: er steht **unter** dem des
	// Menschen und ersetzt ihn nicht.
	Subslogan         string `json:"subslogan"`
	Kurzebeschreibung string `json:"kurzebeschreibung"`
	Langebeschreibung string `json:"langebeschreibung"`

	// Badge ist der Dateiname des Abzeichens — leer heißt: keines.
	Badge string `json:"badge"`

	// **Ein technisches Profil und sein Ablauf** — Werkstück 10 (#1351),
	// Punkt 12.
	//
	// **`technisch` sagt, dass kein Mensch dahintersteht**: was dort entsteht,
	// entsteht nicht von jemandes Hand. Es gehört trotzdem einem Nutzer —
	// irgendwer muss es abschalten können.
	//
	// **`ablauf` ist ein Zeitpunkt und kein Löschauftrag** (RFC 3339, leer
	// heißt „läuft nicht ab"). Nichts verschwindet von selbst; was abgelaufen
	// ist, gilt nicht mehr, und was damit geschieht, entscheidet, wer es
	// liest.
	Technisch bool   `json:"technisch"`
	Ablauf    string `json:"ablauf"`

	// Fassung ist die, die der Rufer beim Ändern zurückgibt — Werkstück 8
	// (#1351).
	Fassung string `json:"fassung"`
}

// ProfilangabenDTO ist, was eine Fläche an einem Profil ändern darf.
//
// **Ein Verbund und keine sechs Argumente** — dieselbe Bauform und derselbe
// Grund wie bei [NutzerangabenDTO]: die Felder gehören zusammen, und wer sie
// einzeln reicht, vertauscht irgendwann zwei.
//
// **Die Nutzerkennung steht nicht darin.** Ein Profil wechselt seinen Nutzer
// nicht; wer sie mitschickte, könnte es.
type ProfilangabenDTO struct {
	ID   string `json:"id"`
	Name string `json:"name"`

	Subslogan         string `json:"subslogan"`
	Kurzebeschreibung string `json:"kurzebeschreibung"`
	Langebeschreibung string `json:"langebeschreibung"`
	Badge             string `json:"badge"`

	// **Technisch und Ablauf** — Werkstück 10 (#1351). Der Grund steht bei
	// [ProfilDTO] und wird hier nicht wiederholt.
	Technisch bool   `json:"technisch"`
	Ablauf    string `json:"ablauf"`

	// Fassung ist die, die der Rufer gelesen hat.
	Fassung string `json:"fassung"`
}

// NutzerAntwort trägt die Übersicht.
type NutzerAntwort struct {
	Nutzer []NutzerDTO `json:"nutzer"`
	Fehler *Fehler     `json:"fehler"`
}

// EinNutzerAntwort trägt einen einzelnen Nutzer.
type EinNutzerAntwort struct {
	Nutzer *NutzerDTO `json:"nutzer"`
	Fehler *Fehler    `json:"fehler"`
}

// EinProfilAntwort trägt ein einzelnes Profil.
type EinProfilAntwort struct {
	Profil *ProfilDTO `json:"profil"`
	Fehler *Fehler    `json:"fehler"`
}

// MitNutzern setzt die Naht.
func (a *App) MitNutzern(fuehrer Nutzerfuehrer) { a.nutzer = fuehrer }

// Nutzer gibt alle Nutzer samt ihren Profilen.
func (a *App) Nutzer(ctx context.Context) NutzerAntwort {
	ctx = a.imProfil(ctx)

	zeilen, err := a.nutzer.Uebersicht(ctx)
	if err != nil {
		return NutzerAntwort{Fehler: Uebersetze(ctx, err)}
	}

	// **Nie `nil`, sondern eine leere Liste** (§3): `json:",omitempty"` ließe
	// sie sonst aus der Antwort verschwinden, und die Seite bliebe leer.
	gelesen := make([]NutzerDTO, 0, len(zeilen))

	for _, zeile := range zeilen {
		gelesen = append(gelesen, nutzerDTO(zeile))
	}

	return NutzerAntwort{Nutzer: gelesen}
}

// NutzerAnlegen legt einen Nutzer samt seinem ersten Profil an.
func (a *App) NutzerAnlegen(ctx context.Context, name, bild string) EinNutzerAntwort {
	ctx = a.imProfil(ctx)

	angelegt, err := a.nutzer.NutzerAnlegen(ctx, domain.Nutzer{Name: name, Bild: bild})
	if err != nil {
		return EinNutzerAntwort{Fehler: Uebersetze(ctx, err)}
	}

	// **Mit dem ersten Profil.** Es entsteht in derselben Transaktion, und
	// eine leere Liste hier behauptete, es gebe keines — wer danach
	// hinwechseln will, holte sonst die ganze Übersicht neu (§7, Regel 2).
	gelesen := nutzerDTO(angelegt)

	return EinNutzerAntwort{Nutzer: &gelesen}
}

// ProfilAnlegen legt ein weiteres Profil für einen Nutzer an.
func (a *App) ProfilAnlegen(ctx context.Context, nutzerID, name string) EinProfilAntwort {
	ctx = a.imProfil(ctx)

	gelesen, fehler := benannteKennungAus("nutzerID", nutzerID)
	if fehler != nil {
		return EinProfilAntwort{Fehler: fehler}
	}

	if fremd := a.ueberNutzer(ctx, gelesen); fremd != nil {
		return EinProfilAntwort{Fehler: Uebersetze(ctx, fremd)}
	}

	angelegt, anlegefehler := a.nutzer.ProfilAnlegen(
		ctx, domain.Profil{NutzerID: gelesen, Name: name})
	if anlegefehler != nil {
		return EinProfilAntwort{Fehler: Uebersetze(ctx, anlegefehler)}
	}

	profil := profilDTO(angelegt)

	return EinProfilAntwort{Profil: &profil}
}

// nutzerDTO übersetzt eine Zeile der Übersicht.
func nutzerDTO(zeile service.Nutzerzeile) NutzerDTO {
	profile := make([]ProfilDTO, 0, len(zeile.Profile))

	for _, eines := range zeile.Profile {
		profile = append(profile, profilDTO(eines))
	}

	return NutzerDTO{
		ID:      zahltext(zeile.Nutzer.ID),
		Name:    zeile.Nutzer.Name,
		Bild:    zeile.Nutzer.Bild,
		Email:   zeile.Nutzer.Email,
		Telefon: zeile.Nutzer.Telefon,
		Anonym:  zeile.Nutzer.Anonym,

		Anrede:       zeile.Nutzer.Anrede,
		Namenszusatz: zeile.Nutzer.Namenszusatz,
		Vorname:      zeile.Nutzer.Vorname,
		Weiterenamen: zeile.Nutzer.Weiterenamen,
		Nachname:     zeile.Nutzer.Nachname,
		Slogan:       zeile.Nutzer.Slogan,
		Kurzebio:     zeile.Nutzer.Kurzebio,
		Langebio:     zeile.Nutzer.Langebio,
		Fassung:      zahltext(zeile.Nutzer.Fassung),

		Profile: profile,
	}
}

// profilDTO übersetzt ein Profil.
func profilDTO(eines domain.Profil) ProfilDTO {
	return ProfilDTO{
		ID:       zahltext(eines.ID),
		NutzerID: zahltext(eines.NutzerID),
		Name:     eines.Name,

		Subslogan:         eines.Subslogan,
		Kurzebeschreibung: eines.Kurzebeschreibung,
		Langebeschreibung: eines.Langebeschreibung,
		Badge:             eines.Badge,
		Technisch:         eines.Technisch,
		Ablauf:            eines.Ablauf,

		Fassung: zahltext(eines.Fassung),
	}
}
