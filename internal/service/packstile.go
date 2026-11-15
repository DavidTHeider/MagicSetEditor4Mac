package service

// **Ein Umgebungsbauer je Stilbogen** — Befund 1 des Zweitblicks über Phase 4.1
// (#1288).

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msescript"
)

// stilbauer hält je Stilbogen einen [umgebungsbauer].
//
// # Warum je Stilbogen und nicht je Set
//
// **Weil eine Karte ihren eigenen Stil tragen darf.** MSE2 hält seine Kontexte
// in einer Abbildung über den Stilbogen (`contexts.try_emplace(stylesheet)`)
// und wählt für jede Karte zuerst `stylesheetForP(card)`
// (`src/script/script_manager.cpp`, Zeilen 30 und 64, Fassung f9d9356 vom
// 2021-07-06). Ein Bauer für das ganze Set gäbe einer Karte mit eigenem Stil
// das Initskript eines fremden.
//
// # Warum die Abbildung über den Namen geht
//
// `Vorlage` ist ein Verbund mit Slices und damit nicht vergleichbar; was einen
// Stilbogen benennt, ist sein Name. Denselben Schlüssel benutzt
// [Stilspeicher.gemerkterKartenstil] eine Ebene tiefer — „ein Set hat wenige
// Stile und viele Karten".
//
// # Was er nicht tut
//
// **Er lebt nur so lange wie der Aufruf.** Ein Speicher über Aufrufe hinweg
// bräuchte eine Ungültigkeitsregel, und die ist eine eigene Frage; hier geht
// es darum, dass die Initskripte innerhalb **eines** Laufs einmal je Stil
// laufen.
type stilbauer struct {
	bereich *Packbereich
	setweit Vorlage
	bauer   map[string]*umgebungsbauer
}

// bauerJeStil baut die Verwaltung für einen Lauf.
func (p *Packbereich) bauerJeStil(_ context.Context, setweit Vorlage) *stilbauer {
	return &stilbauer{bereich: p, setweit: setweit, bauer: map[string]*umgebungsbauer{}}
}

// fuer gibt die Umgebung einer Karte — über dem Bauer ihres Stilbogens.
func (s *stilbauer) fuer(ctx context.Context, karte *domain.CardRow) (
	*msescript.Umgebung, []msescript.Eintrag, []domain.Befund,
) {
	vorlage := s.vorlageFuer(ctx, karte)

	name := stilname(vorlage.Stil)

	bauer, da := s.bauer[name]
	if !da {
		bauer = neuerUmgebungsbauer(ctx, vorlage,
			s.bereich.planer.schriften, s.bereich.planer.bilder, karte)
		s.bauer[name] = bauer
	}

	return bauer.fuerKarte(ctx, karte)
}

// vorlageFuer gibt den Stil dieser Karte — ihren eigenen, sonst den des Sets.
//
// **Der Rückfall ist die Setvorlage und nicht ein leerer Verbund.** Wo der
// Stilspeicher fehlt oder die Karte keinen eigenen Stil findet, gilt weiter,
// was der Lauf ohnehin gelesen hat; alles andere zöge aus nichts (§8).
func (s *stilbauer) vorlageFuer(ctx context.Context, karte *domain.CardRow) Vorlage {
	if s.bereich.planer == nil || s.bereich.planer.stile == nil {
		return s.setweit
	}

	vorlage, _, da := s.bereich.planer.stile.StilFuerKarte(ctx, karte)
	if !da {
		return s.setweit
	}

	return vorlage
}
