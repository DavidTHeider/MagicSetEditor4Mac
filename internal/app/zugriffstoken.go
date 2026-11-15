// **Zugriffstoken am Fenster und an den Flächen** — Werkstück 10 von Phase
// 4.12 (#1351),
// Punkt 13.
//
// # Der Klartext kommt genau einmal heraus
//
// **Beim Anlegen und nie wieder.** [EinTokenAntwort.Klartext] ist bei jeder
// anderen Antwort leer, und das ist keine Sparsamkeit: der Bestand trägt nur
// die gehütete Form, es gibt ihn danach nirgends mehr. Wer ihn verliert, legt
// ein neues an.
//
// **Deshalb sagt die Oberfläche es, bevor sie fragt.** Ein Feld, das man
// später nachschlagen kann, kopiert niemand — und dann fehlt es.
//
// # Was hier **nicht** steht
//
// **Kein Prüfen eines Tokens.** Das tut der REST-Zugang, und nur er: er ist
// die einzige Fläche, die auf einem Port lauscht und deshalb einen Gegenüber
// hat, der sich ausweisen muss. Das Fenster weiß bereits, wer davorsitzt.

package app

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Tokenfuehrer ist die Naht zur Tokenverwaltung.
type Tokenfuehrer interface {
	Alle(ctx context.Context, profilID int64) ([]domain.Zugriffstoken, error)
	Anlegen(ctx context.Context, neu domain.Zugriffstoken) (
		domain.Zugriffstoken, string, error)
	Aendern(ctx context.Context, geaendert domain.Zugriffstoken) (
		domain.Zugriffstoken, error)
	Loeschen(ctx context.Context, tokenID int64) error

	// ProfilDes sagt, welchem Profil ein Token gehört — die Rolle zählt dort
	// (#1817).
	ProfilDes(ctx context.Context, tokenID int64) (int64, error)

	// **Rotation, Sammelwiderruf und Ablaufwahlen** — Werkstück O2 von
	// Phase 5.2 (#1426).
	Rotieren(ctx context.Context, tokenID int64) (domain.Zugriffstoken, string, error)
	AlleWiderrufen(ctx context.Context, profilID int64) (int64, error)
	Ablaufwahlen() []domain.Ablaufwahl
}

// MitToken setzt die Naht.
func (a *App) MitToken(fuehrer Tokenfuehrer) {
	a.token = fuehrer
}

// TokenDTO ist ein Zugriffstoken, wie eine Fläche es sieht — **ohne sein
// Geheimnis**.
type TokenDTO struct {
	ID       string `json:"id"`
	ProfilID string `json:"profilID"`

	Bezeichnung string `json:"bezeichnung"`

	// Principal sagt, für wen das Token arbeitet — Werkstück O2 (#1426).
	Principal string `json:"principal"`

	// Scope ist `lesen` oder `schreiben`; `schreiben` schließt Lesen ein.
	Scope string `json:"scope"`

	// Ablauf ist ein Zeitpunkt in RFC 3339 — leer heißt „läuft nicht ab".
	Ablauf string `json:"ablauf"`

	// Angelegt ist der Zeitpunkt des Anlegens, ebenfalls RFC 3339.
	Angelegt string `json:"angelegt"`

	// LetzteNutzung ist RFC 3339 — leer heißt: noch nie angenommen
	// (Werkstück O2, #1426).
	LetzteNutzung string `json:"letzteNutzung"`

	// VorgaengerBis ist die Frist, bis zu der das vorige Geheimnis nach
	// einer Rotation noch gilt — leer heißt: es gibt keines.
	VorgaengerBis string `json:"vorgaengerBis"`

	// Abgelaufen sagt, ob [TokenDTO.Ablauf] **jetzt** überschritten ist.
	//
	// **Ausgerechnet und nicht der Fläche überlassen.** Sonst rechnete jede
	// der vier dieselbe Zeitspanne aus, und die Ergebnisse liefen an der
	// Zeitzone auseinander.
	Abgelaufen bool `json:"abgelaufen"`

	// Partie und Sitz nennen die Bindung an einen Sitz (#1819) — leer
	// heißt: ungebunden.
	Partie string `json:"partie"`
	Sitz   string `json:"sitz"`

	Fassung string `json:"fassung"`
}

// TokenangabenDTO ist, was eine Fläche an einem Token setzen darf.
type TokenangabenDTO struct {
	// ID ist beim Anlegen leer und beim Ändern die des Tokens.
	ID string `json:"id"`

	// ProfilID ist beim Anlegen nötig und wird beim Ändern nicht gelesen —
	// **ein Token wechselt sein Profil nicht.**
	ProfilID string `json:"profilID"`

	Bezeichnung string `json:"bezeichnung"`
	Principal   string `json:"principal"`
	Scope       string `json:"scope"`
	Ablauf      string `json:"ablauf"`

	// Partie und Sitz binden das Token beim Anlegen an einen Sitz (#1819);
	// beim Ändern werden sie nicht gelesen — **ein Token wechselt seinen Sitz
	// nicht**.
	Partie string `json:"partie"`
	Sitz   string `json:"sitz"`

	Fassung string `json:"fassung"`
}

// TokenAntwort trägt die Token eines Profils.
type TokenAntwort struct {
	Token  []TokenDTO `json:"token"`
	Fehler *Fehler    `json:"fehler"`
}

// EinTokenAntwort trägt ein einzelnes Token.
type EinTokenAntwort struct {
	Token *TokenDTO `json:"token"`

	// Klartext ist **nur beim Anlegen** gefüllt.
	Klartext string `json:"klartext"`

	Fehler *Fehler `json:"fehler"`
}

// Token gibt die Zugriffstoken eines Profils.
func (a *App) Token(ctx context.Context, profilID string) TokenAntwort {
	ctx = a.imProfil(ctx)

	kennung, befund := benannteKennungAus("profilID", profilID)
	if befund != nil {
		return TokenAntwort{Token: []TokenDTO{}, Fehler: befund}
	}

	// **Die Rolle am genannten Profil** (#1817) — auch fürs Lesen: wer Token
	// eines Profils sieht, sieht, wer dort von außen arbeitet.
	if fehler := a.rolleAm(ctx, kennung, domain.RolleMitarbeiter); fehler != nil {
		return TokenAntwort{Token: []TokenDTO{}, Fehler: Uebersetze(ctx, fehler)}
	}

	gefunden, fehler := a.token.Alle(ctx, kennung)
	if fehler != nil {
		// **Auch der Fehlerweg trägt eine leere Liste und kein `nil`.** Über
		// die JSON-Naht wird daraus sonst `null`, und das Fenster greift in
		// etwas, das es nicht gibt (§3, „zwischen den Prüfebenen").
		return TokenAntwort{Token: []TokenDTO{}, Fehler: Uebersetze(ctx, fehler)}
	}

	return TokenAntwort{Token: a.tokenzeilen(gefunden)}
}

// TokenAnlegen legt ein Zugriffstoken an und gibt seinen Klartext **einmal**.
func (a *App) TokenAnlegen(
	ctx context.Context, angaben TokenangabenDTO,
) EinTokenAntwort {
	ctx = a.imProfil(ctx)

	kennung, befund := benannteKennungAus("profilID", angaben.ProfilID)
	if befund != nil {
		return EinTokenAntwort{Fehler: befund}
	}

	if fehler := a.zumSchreibenAm(ctx, kennung); fehler != nil {
		return EinTokenAntwort{Fehler: Uebersetze(ctx, fehler)}
	}

	angelegt, klartext, fehler := a.token.Anlegen(ctx, domain.Zugriffstoken{
		ProfilID:    kennung,
		Bezeichnung: angaben.Bezeichnung,
		Principal:   angaben.Principal,
		Scope:       domain.Scope(angaben.Scope),
		Ablauf:      angaben.Ablauf,
		Sitzbindung: domain.Sitzbindung{Partie: angaben.Partie, Sitz: angaben.Sitz},
	})
	if fehler != nil {
		return EinTokenAntwort{Fehler: Uebersetze(ctx, fehler)}
	}

	zeile := a.tokenzeile(angelegt)

	return EinTokenAntwort{Token: &zeile, Klartext: klartext}
}

// TokenAendern schreibt Bezeichnung, Erlaubnis und Ablauf fort.
//
// **Das Geheimnis bleibt dasselbe.** Wer ein Token umbenennt, entzieht damit
// niemandem den Zugang; wer ein neues Geheimnis will, legt ein neues Token an
// und nimmt das alte heraus.
func (a *App) TokenAendern(
	ctx context.Context, angaben TokenangabenDTO,
) EinTokenAntwort {
	ctx = a.imProfil(ctx)

	kennung, befund := benannteKennungAus("tokenID", angaben.ID)
	if befund != nil {
		return EinTokenAntwort{Fehler: befund}
	}

	fassung, fassungsbefund := benannteKennungAus("fassung", angaben.Fassung)
	if fassungsbefund != nil {
		return EinTokenAntwort{Fehler: fassungsbefund}
	}

	if fehler := a.rolleAmToken(ctx, kennung); fehler != nil {
		return EinTokenAntwort{Fehler: Uebersetze(ctx, fehler)}
	}

	geaendert, fehler := a.token.Aendern(ctx, domain.Zugriffstoken{
		ID:          kennung,
		Bezeichnung: angaben.Bezeichnung,
		Principal:   angaben.Principal,
		Scope:       domain.Scope(angaben.Scope),
		Ablauf:      angaben.Ablauf,
		Fassung:     fassung,
	})
	if fehler != nil {
		return EinTokenAntwort{Fehler: Uebersetze(ctx, fehler)}
	}

	zeile := a.tokenzeile(geaendert)

	return EinTokenAntwort{Token: &zeile}
}

// TokenLoeschen nimmt ein Zugriffstoken heraus.
func (a *App) TokenLoeschen(ctx context.Context, tokenID string) SchreibAntwort {
	ctx = a.imProfil(ctx)

	kennung, befund := benannteKennungAus("tokenID", tokenID)
	if befund != nil {
		return SchreibAntwort{Fehler: befund}
	}

	if fehler := a.rolleAmToken(ctx, kennung); fehler != nil {
		return SchreibAntwort{Fehler: Uebersetze(ctx, fehler)}
	}

	if fehler := a.token.Loeschen(ctx, kennung); fehler != nil {
		return SchreibAntwort{Fehler: Uebersetze(ctx, fehler)}
	}

	return SchreibAntwort{}
}

// tokenzeilen bildet eine Liste ab — **nie `nil`**, sonst stünde in der
// Antwort `null` statt einer leeren Liste (§3, „zwischen den Prüfebenen").
func (a *App) tokenzeilen(gefunden []domain.Zugriffstoken) []TokenDTO {
	zeilen := make([]TokenDTO, 0, len(gefunden))
	for _, eines := range gefunden {
		zeilen = append(zeilen, a.tokenzeile(eines))
	}

	return zeilen
}

// tokenzeile bildet eines ab.
func (a *App) tokenzeile(eines domain.Zugriffstoken) TokenDTO {
	return TokenDTO{
		ID:            zahltext(eines.ID),
		ProfilID:      zahltext(eines.ProfilID),
		Bezeichnung:   eines.Bezeichnung,
		Principal:     eines.Principal,
		Scope:         string(eines.Scope),
		Ablauf:        eines.Ablauf,
		Angelegt:      eines.Angelegt,
		LetzteNutzung: eines.LetzteNutzung,
		VorgaengerBis: eines.VorgaengerBis,
		Abgelaufen:    eines.Abgelaufen(a.jetzt()),
		Partie:        eines.Sitzbindung.Partie,
		Sitz:          eines.Sitzbindung.Sitz,
		Fassung:       zahltext(eines.Fassung),
	}
}
