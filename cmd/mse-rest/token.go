// **Die Naht zu den Zugriffstoken und die Wache darüber** — Werkstück 10 von
// Phase 4.12
// (#1351),
// Punkt 13.
//
// **Nur die Verwaltung steht hier.** Das **Vorzeigen** — die Wache, die ein
// Token prüft und die Anfrage in sein Profil stellt — steht in
// `internal/app/tokenwache.go`: sie braucht `internal/service` und
// `internal/storage` zusammen, und ein Composition Root der Kommandozeile darf
// `service` nicht kennen (§2). `internal/app` ist die Schicht, die beides
// zusammenführen darf, und dieser Weg ist derselbe wie bei jeder anderen
// Fassade hier.

package main

import (
	"context"
	"database/sql"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/rest"
)

// fremdtoken bedient die Verwaltung über der Fassade.
type fremdtoken struct{ db *sql.DB }

func (f fremdtoken) Token(ctx context.Context, profilID string) ([]rest.Token, error) {
	antwort := app.NutzerflaecheKopflos(f.db).Token(ctx, profilID)
	if antwort.Fehler != nil {
		return nil, antwort.Fehler.AlsFehler()
	}

	zeilen := make([]rest.Token, 0, len(antwort.Token))
	for _, eines := range antwort.Token {
		zeilen = append(zeilen, alsTokenzeile(eines))
	}

	return zeilen, nil
}

func (f fremdtoken) TokenAnlegen(ctx context.Context, angaben rest.Tokenangaben) (
	rest.Token, string, error,
) {
	antwort := app.NutzerflaecheKopflos(f.db).TokenAnlegen(ctx, alsTokenangaben(angaben))
	if antwort.Fehler != nil {
		return rest.Token{}, "", antwort.Fehler.AlsFehler()
	}

	return alsTokenzeile(*antwort.Token), antwort.Klartext, nil
}

func (f fremdtoken) TokenAendern(ctx context.Context, angaben rest.Tokenangaben) (
	rest.Token, error,
) {
	antwort := app.NutzerflaecheKopflos(f.db).TokenAendern(ctx, alsTokenangaben(angaben))
	if antwort.Fehler != nil {
		return rest.Token{}, antwort.Fehler.AlsFehler()
	}

	return alsTokenzeile(*antwort.Token), nil
}

func (f fremdtoken) TokenLoeschen(ctx context.Context, tokenID string) error {
	antwort := app.NutzerflaecheKopflos(f.db).TokenLoeschen(ctx, tokenID)
	if antwort.Fehler != nil {
		return antwort.Fehler.AlsFehler()
	}

	return nil
}

// TokenRotieren gibt einem Token ein neues Geheimnis — Werkstück O2 von
// Phase 5.2 (#1426).
func (f fremdtoken) TokenRotieren(ctx context.Context, tokenID string) (rest.Token, string, error) {
	antwort := app.NutzerflaecheKopflos(f.db).TokenRotieren(ctx, tokenID)
	if antwort.Fehler != nil {
		return rest.Token{}, "", antwort.Fehler.AlsFehler()
	}

	return alsTokenzeile(*antwort.Token), antwort.Klartext, nil
}

// TokenAlleWiderrufen zieht alle Token eines Profils zurück.
func (f fremdtoken) TokenAlleWiderrufen(ctx context.Context, profilID string) (int64, error) {
	antwort := app.NutzerflaecheKopflos(f.db).TokenAlleWiderrufen(ctx, profilID)
	if antwort.Fehler != nil {
		return 0, antwort.Fehler.AlsFehler()
	}

	return antwort.Gezogen, nil
}

// Tokenablaufwahlen nennt die Vorschläge für den Ablauf.
func (f fremdtoken) Tokenablaufwahlen(ctx context.Context) ([]rest.Ablaufwahl, error) {
	antwort := app.NutzerflaecheKopflos(f.db).Tokenablaufwahlen(ctx)
	if antwort.Fehler != nil {
		return nil, antwort.Fehler.AlsFehler()
	}

	wahlen := make([]rest.Ablaufwahl, 0, len(antwort.Wahlen))
	for _, wahl := range antwort.Wahlen {
		wahlen = append(wahlen, rest.Ablaufwahl{Name: wahl.Name, Ablauf: wahl.Ablauf})
	}

	return wahlen, nil
}

func alsTokenangaben(angaben rest.Tokenangaben) app.TokenangabenDTO {
	return app.TokenangabenDTO{
		ID:          angaben.TokenID,
		ProfilID:    angaben.ProfilID,
		Bezeichnung: angaben.Bezeichnung,
		Principal:   angaben.Principal,
		Scope:       angaben.Scope,
		Ablauf:      angaben.Ablauf,
		Partie:      angaben.Partie,
		Sitz:        angaben.Sitz,
		Fassung:     angaben.Fassung,
	}
}

func alsTokenzeile(eines app.TokenDTO) rest.Token {
	return rest.Token{
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
		Partie:        eines.Partie,
		Sitz:          eines.Sitz,
		Fassung:       eines.Fassung,
	}
}
