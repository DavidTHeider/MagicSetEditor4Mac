package service

// **Ein erster Vorgang entsteht nur, wo dieses Profil Partei ist** — Befund
// Z3 des Zweitblicks über Phase 6.8
// (#1669).
//
// # Warum nur beim ersten
//
// **Weil jede spätere Revision an einen hiesigen Stand anschließt**, und den
// gibt es nur, wo dieses Profil schon Partei war — die Parteien hält
// [domain.Handelsangebot.NimmtVon] fest. Neu entstehen kann ein Vorgang nur
// über die erste Revision, und dort prüfte bis hierher allein, dass der
// **Absender** Partei ist.

import (
	"context"
	"slices"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// MitAusweis lässt die Post wissen, wer sie ist.
//
// **Wahlfrei wie die Klärung** — die Anwendung setzt ihn
// (`internal/app/handelsleitungsnaht.go`), und ein Lauf an der echten Naht
// hält, dass sie es tut.
func (p *Handelspost) MitAusweis(stelle Ausweisstelle) *Handelspost {
	p.ausweis = stelle

	return p
}

// istPartei prüft, dass dieses Profil im Angebot steht.
func (p *Handelspost) istPartei(ctx context.Context, angebot domain.Handelsangebot) error {
	if p.ausweis == nil {
		return nil
	}

	ich, fehler := p.ausweis.Ausweis(ctx)
	if fehler != nil {
		return fehler
	}

	if !slices.Contains(angebot.Parteien, string(ich.Kennung())) {
		return domain.MitGrund(domain.ErrErlaubnis, nil,
			"Dieser Handel nennt dieses Profil nicht als Partei.")
	}

	return nil
}
