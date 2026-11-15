// **Die Naht zu den Zugriffstoken** — Werkstück 10 von Phase 4.12
// (#1351).
//
// **Dieselbe Form wie `nutzer.go`** und aus demselben Grund (§16): die
// Anwendung ist dieselbe, nur die Fläche ist eine andere.

package mcpnaehte

import (
	"context"
	"database/sql"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mcp"
)

// token bedient die Werkzeuge über einem Bestand.
type token struct{ db *sql.DB }

// Token nennt die Zugriffstoken eines Profils.
func (z token) Token(ctx context.Context, profilID string) ([]mcp.Tokenzeile, error) {
	antwort := app.NutzerflaecheKopflos(z.db).Token(ctx, profilID)
	if antwort.Fehler != nil {
		return nil, antwort.Fehler.AlsFehler()
	}

	zeilen := make([]mcp.Tokenzeile, 0, len(antwort.Token))
	for _, eines := range antwort.Token {
		zeilen = append(zeilen, alsTokenzeile(eines))
	}

	return zeilen, nil
}

// TokenAnlegen legt eines an und gibt seinen Klartext **einmal**.
func (z token) TokenAnlegen(ctx context.Context, angaben mcp.Tokenangaben) (
	mcp.Tokenzeile, string, error,
) {
	antwort := app.NutzerflaecheKopflos(z.db).TokenAnlegen(ctx, alsTokenangaben(angaben))
	if antwort.Fehler != nil {
		return mcp.Tokenzeile{}, "", antwort.Fehler.AlsFehler()
	}

	return alsTokenzeile(*antwort.Token), antwort.Klartext, nil
}

// TokenAendern schreibt Bezeichnung, Erlaubnis und Ablauf fort.
func (z token) TokenAendern(ctx context.Context, angaben mcp.Tokenangaben) (
	mcp.Tokenzeile, error,
) {
	antwort := app.NutzerflaecheKopflos(z.db).TokenAendern(ctx, alsTokenangaben(angaben))
	if antwort.Fehler != nil {
		return mcp.Tokenzeile{}, antwort.Fehler.AlsFehler()
	}

	return alsTokenzeile(*antwort.Token), nil
}

// TokenLoeschen zieht eines zurück.
func (z token) TokenLoeschen(ctx context.Context, tokenID string) error {
	antwort := app.NutzerflaecheKopflos(z.db).TokenLoeschen(ctx, tokenID)
	if antwort.Fehler != nil {
		return antwort.Fehler.AlsFehler()
	}

	return nil
}

// alsTokenangaben bildet die Gestalt des Werkzeugs auf die der Fassade ab.
// TokenAlleWiderrufen zieht alle Token eines Profils zurück — Werkstück O2
// von Phase 5.2 (#1426).
func (z token) TokenAlleWiderrufen(ctx context.Context, profilID string) (int64, error) {
	antwort := app.NutzerflaecheKopflos(z.db).TokenAlleWiderrufen(ctx, profilID)
	if antwort.Fehler != nil {
		return 0, antwort.Fehler.AlsFehler()
	}

	return antwort.Gezogen, nil
}

// Tokenablaufwahlen nennt die Vorschläge für den Ablauf.
func (z token) Tokenablaufwahlen(ctx context.Context) ([]mcp.Ablaufwahlzeile, error) {
	antwort := app.NutzerflaecheKopflos(z.db).Tokenablaufwahlen(ctx)
	if antwort.Fehler != nil {
		return nil, antwort.Fehler.AlsFehler()
	}

	zeilen := make([]mcp.Ablaufwahlzeile, 0, len(antwort.Wahlen))
	for _, wahl := range antwort.Wahlen {
		zeilen = append(zeilen, mcp.Ablaufwahlzeile{Name: wahl.Name, Ablauf: wahl.Ablauf})
	}

	return zeilen, nil
}

func alsTokenangaben(angaben mcp.Tokenangaben) app.TokenangabenDTO {
	return app.TokenangabenDTO{
		ID:          angaben.TokenID,
		ProfilID:    angaben.ProfilID,
		Bezeichnung: angaben.Bezeichnung,
		Principal:   angaben.Principal,
		Scope:       angaben.Scope,
		Ablauf:      angaben.Ablauf,
		Fassung:     angaben.Fassung,
	}
}

// alsTokenzeile bildet die Antwort der Fassade auf die des Werkzeugs ab.
func alsTokenzeile(eines app.TokenDTO) mcp.Tokenzeile {
	return mcp.Tokenzeile{
		ID:            eines.ID,
		ProfilID:      eines.ProfilID,
		Bezeichnung:   eines.Bezeichnung,
		Principal:     eines.Principal,
		Scope:         eines.Scope,
		Ablauf:        eines.Ablauf,
		Angelegt:      eines.Angelegt,
		Abgelaufen:    eines.Abgelaufen,
		LetzteNutzung: eines.LetzteNutzung,
		VorgaengerBis: eines.VorgaengerBis,
		Fassung:       eines.Fassung,
	}
}
