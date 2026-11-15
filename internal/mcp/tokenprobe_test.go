package mcp_test

// **Die Probe der Tokenpflege** — Werkstück 10 von Phase 4.12
// (#1351).
//
// **Sie hält ihren Stand**, weil die Idempotenzläufe zweimal gegen **denselben**
// Bestand rufen: eine Probe, die jedes Mal dasselbe erfindet, machte jedes
// Werkzeug idempotent (§10).

import (
	"context"
	"fmt"
	"strconv"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mcp"
)

// tokenprobe ist eine Tokenpflege im Arbeitsspeicher.
type tokenprobe struct {
	abgelegt map[string]mcp.Tokenzeile
	naechste int64

	// fehler ist der, den jeder Weg stattdessen gibt.
	fehler error
}

func neueTokenprobe() *tokenprobe {
	return &tokenprobe{
		abgelegt: map[string]mcp.Tokenzeile{
			"1": {
				ID: "1", ProfilID: "1", Bezeichnung: "Nachtlauf",
				Scope: "lesen", Angelegt: "2026-09-05T12:00:00Z", Fassung: "1",
			},
		},
		naechste: 2,
	}
}

func (p *tokenprobe) Token(_ context.Context, profilID string) ([]mcp.Tokenzeile, error) {
	if p.fehler != nil {
		return nil, p.fehler
	}

	zeilen := []mcp.Tokenzeile{}

	for kennung := int64(1); kennung < p.naechste; kennung++ {
		eines, da := p.abgelegt[strconv.FormatInt(kennung, 10)]
		if da && eines.ProfilID == profilID {
			zeilen = append(zeilen, eines)
		}
	}

	return zeilen, nil
}

func (p *tokenprobe) TokenAnlegen(_ context.Context, angaben mcp.Tokenangaben) (
	mcp.Tokenzeile, string, error,
) {
	if p.fehler != nil {
		return mcp.Tokenzeile{}, "", p.fehler
	}

	kennung := strconv.FormatInt(p.naechste, 10)
	p.naechste++

	neu := mcp.Tokenzeile{
		ID: kennung, ProfilID: angaben.ProfilID,
		Bezeichnung: angaben.Bezeichnung, Scope: angaben.Scope,
		Ablauf: angaben.Ablauf, Angelegt: "2026-09-05T12:00:00Z", Fassung: "1",
	}
	p.abgelegt[kennung] = neu

	// **Ein anderes Geheimnis bei jedem Ruf** — das ist die Eigenschaft, um
	// derentwillen dieses Werkzeug keine Idempotenz zusagt.
	return neu, "mse_" + kennung + "_GEHEIMNIS" + kennung, nil
}

func (p *tokenprobe) TokenAendern(_ context.Context, angaben mcp.Tokenangaben) (
	mcp.Tokenzeile, error,
) {
	if p.fehler != nil {
		return mcp.Tokenzeile{}, p.fehler
	}

	vorher, da := p.abgelegt[angaben.TokenID]
	if !da {
		return mcp.Tokenzeile{}, fmt.Errorf(
			"%w: dieses Token gibt es nicht", domain.ErrNotFound)
	}

	vorher.Bezeichnung = angaben.Bezeichnung
	vorher.Scope = angaben.Scope
	vorher.Ablauf = angaben.Ablauf
	p.abgelegt[angaben.TokenID] = vorher

	return vorher, nil
}

func (p *tokenprobe) TokenLoeschen(_ context.Context, tokenID string) error {
	if p.fehler != nil {
		return p.fehler
	}

	delete(p.abgelegt, tokenID)

	return nil
}

// TokenAlleWiderrufen zieht die Token eines Profils zurück — Werkstück O2.
func (p *tokenprobe) TokenAlleWiderrufen(_ context.Context, profilID string) (int64, error) {
	if p.fehler != nil {
		return 0, p.fehler
	}

	var gezogen int64

	for kennung, eines := range p.abgelegt {
		if eines.ProfilID == profilID {
			delete(p.abgelegt, kennung)
			gezogen++
		}
	}

	return gezogen, nil
}

// Tokenablaufwahlen nennt zwei feste Vorschläge.
func (p *tokenprobe) Tokenablaufwahlen(context.Context) ([]mcp.Ablaufwahlzeile, error) {
	if p.fehler != nil {
		return nil, p.fehler
	}

	return []mcp.Ablaufwahlzeile{{Name: "tag", Ablauf: "2026-09-06T12:00:00Z"}, {Name: "nie"}}, nil
}
