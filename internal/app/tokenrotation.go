// **Rotation, Sammelwiderruf und Ablaufwahlen** — Werkstück O2 von Phase 5.2
// (#1426),
// nach [Vermerk 0080](../../docs/entscheidungen/0080-maschinenzugaenge-und-widerrufbare-sitzungen.md).
//
// **Keine fachliche Entscheidung hier** (§2): übersetzt wird zwischen Wails
// und dem Dienst. Wie lange ein Vorgänger gilt und welche Fristen zur Wahl
// stehen, sagt `domain`.

package app

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// TokenwiderrufAntwort sagt, wie viele Token ein Sammelwiderruf gezogen hat.
type TokenwiderrufAntwort struct {
	Gezogen int64   `json:"gezogen"`
	Fehler  *Fehler `json:"fehler"`
}

// AblaufwahlDTO ist ein benannter Vorschlag für den Ablauf eines Tokens.
type AblaufwahlDTO struct {
	Name   string `json:"name"`
	Ablauf string `json:"ablauf"`
}

// AblaufwahlenAntwort nennt die Vorschläge, gerechnet aus der Uhr des
// Dienstes — damit alle vier Flächen dieselben Zeitpunkte nennen.
type AblaufwahlenAntwort struct {
	Wahlen []AblaufwahlDTO `json:"wahlen"`
	Fehler *Fehler         `json:"fehler"`
}

// TokenRotieren gibt einem Token ein neues Geheimnis — der Klartext steht nur
// in dieser einen Antwort, wie beim Anlegen.
func (a *App) TokenRotieren(ctx context.Context, tokenID string) EinTokenAntwort {
	ctx = a.imProfil(ctx)

	kennung, befund := benannteKennungAus("tokenID", tokenID)
	if befund != nil {
		return EinTokenAntwort{Fehler: befund}
	}

	// **Die Rolle an dem Profil, dem das Token gehört** (#1817) — die
	// Rotation gibt einen Klartext heraus, und der öffnet dieses Profil.
	if fehler := a.rolleAmToken(ctx, kennung); fehler != nil {
		return EinTokenAntwort{Fehler: Uebersetze(ctx, fehler)}
	}

	rotiert, klartext, fehler := a.token.Rotieren(ctx, kennung)
	if fehler != nil {
		return EinTokenAntwort{Fehler: Uebersetze(ctx, fehler)}
	}

	zeile := a.tokenzeile(rotiert)

	return EinTokenAntwort{Token: &zeile, Klartext: klartext}
}

// TokenAlleWiderrufen zieht jedes Token eines Profils zurück.
func (a *App) TokenAlleWiderrufen(ctx context.Context, profilID string) TokenwiderrufAntwort {
	ctx = a.imProfil(ctx)

	kennung, befund := benannteKennungAus("profilID", profilID)
	if befund != nil {
		return TokenwiderrufAntwort{Fehler: befund}
	}

	if fehler := a.zumSchreibenAm(ctx, kennung); fehler != nil {
		return TokenwiderrufAntwort{Fehler: Uebersetze(ctx, fehler)}
	}

	gezogen, fehler := a.token.AlleWiderrufen(ctx, kennung)
	if fehler != nil {
		return TokenwiderrufAntwort{Fehler: Uebersetze(ctx, fehler)}
	}

	return TokenwiderrufAntwort{Gezogen: gezogen}
}

// Tokenablaufwahlen nennt die Vorschläge für den Ablauf eines Tokens.
//
// **Ohne Profil im Kontext**: die Wahlen hängen an der Uhr, nicht am
// Bestand — deshalb steht hier kein `imProfil`, und der Wächter über die
// Bindungen weiß das.
func (a *App) Tokenablaufwahlen(context.Context) AblaufwahlenAntwort {
	return AblaufwahlenAntwort{Wahlen: ablaufwahlzeilen(a.token.Ablaufwahlen())}
}

func ablaufwahlzeilen(wahlen []domain.Ablaufwahl) []AblaufwahlDTO {
	zeilen := make([]AblaufwahlDTO, 0, len(wahlen))
	for _, wahl := range wahlen {
		zeilen = append(zeilen, AblaufwahlDTO{Name: wahl.Name, Ablauf: wahl.Ablauf})
	}

	return zeilen
}

// TokenRotieren rotiert ein Token vom Fenster aus.
func (b *Bindungen) TokenRotieren(ctx context.Context, tokenID string) EinTokenAntwort {
	return b.app(ctx).TokenRotieren(b.ruf(ctx), tokenID)
}

// TokenAlleWiderrufen zieht vom Fenster aus alle Token eines Profils zurück.
func (b *Bindungen) TokenAlleWiderrufen(ctx context.Context, profilID string) TokenwiderrufAntwort {
	return b.app(ctx).TokenAlleWiderrufen(b.ruf(ctx), profilID)
}

// Tokenablaufwahlen nennt dem Fenster die Ablaufwahlen.
func (b *Bindungen) Tokenablaufwahlen(ctx context.Context) AblaufwahlenAntwort {
	return b.app(ctx).Tokenablaufwahlen(b.ruf(ctx))
}

// TokenRotieren rotiert ein Token auch für die Flächen ohne Fenster.
func (f *Nutzerflaeche) TokenRotieren(ctx context.Context, tokenID string) EinTokenAntwort {
	return f.app.TokenRotieren(ctx, tokenID)
}

// TokenAlleWiderrufen zieht auch für die Flächen ohne Fenster alle Token zurück.
func (f *Nutzerflaeche) TokenAlleWiderrufen(ctx context.Context, profilID string) TokenwiderrufAntwort {
	return f.app.TokenAlleWiderrufen(ctx, profilID)
}

// Tokenablaufwahlen nennt auch den Flächen ohne Fenster die Ablaufwahlen.
func (f *Nutzerflaeche) Tokenablaufwahlen(ctx context.Context) AblaufwahlenAntwort {
	return f.app.Tokenablaufwahlen(ctx)
}
