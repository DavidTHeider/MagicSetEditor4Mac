package service

// **Die Liste der Handelsvorgänge und ihre Entwürfe** — TRADE-01
// (#1715).
//
// # Warum ein eigener Dienst neben der Handelsleitung
//
// **Weil es eine andere Frage ist.** Die Leitung führt **einen** Vorgang
// durch seine Zustände; das Verzeichnis beantwortet, **welche** es gibt, wo
// sie für mich stehen und was ich mit denen tun darf, die noch nicht
// hinausgegangen sind. Die Zustandsmatrix bleibt dabei unberührt: ein
// Entwurf wird bearbeitet, ein angebotener Vorgang weitergeführt — über eine
// neue Revision, nicht über ein Ändern (TRADE-01.AK-03).
//
// Portierungsmarke **new**: MSE2 handelt nicht (gesucht nach `trade`,
// `offer`, `exchange` in `/Users/david/mse2-quelle/src`).

import (
	"context"
	"slices"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Handelsverzeichnisablage ist, was Liste und Entwürfe von der Ablage
// verlangen.
type Handelsverzeichnisablage interface {
	Uebersicht(ctx context.Context) ([]domain.Handelsvorgangzeile, error)
	Vorgang(ctx context.Context, kennung string) (domain.Handelsangebot, error)
	Fortschreiben(ctx context.Context, vorher, danach domain.Handelsangebot) error

	// EntwurfErsetzen und EntwurfLoeschen treffen **nur** einen Entwurf —
	// die Bedingung steht in der Anweisung, denn zwei Prozesse schreiben.
	EntwurfErsetzen(ctx context.Context, angebot domain.Handelsangebot) error
	EntwurfLoeschen(ctx context.Context, vorgang string) error

	Archivieren(ctx context.Context, vorgang string, an bool) error
}

// Handelskontakte sind die Gegenüber, die dieses Profil kennt.
type Handelskontakte interface {
	Liste(ctx context.Context) ([]domain.Peerkontakt, error)
}

// Handelsmarken sind die Marken an den Vorgängen.
type Handelsmarken interface {
	JeObjekt(ctx context.Context, art string) (map[string][]string, error)
}

// Handelsverzeichnis führt die Liste und die Entwürfe.
type Handelsverzeichnis struct {
	ablage    Handelsverzeichnisablage
	kontakte  Handelskontakte
	marken    Handelsmarken
	kennungen func() string

	// schranke ist die Handelsrichtlinie (#2435, `handelsschranke.go`).
	schranke Handelsschranke
}

// NeuesHandelsverzeichnis baut es.
func NeuesHandelsverzeichnis(
	ablage Handelsverzeichnisablage, kontakte Handelskontakte, marken Handelsmarken,
) *Handelsverzeichnis {
	return &Handelsverzeichnis{
		ablage: ablage, kontakte: kontakte, marken: marken, kennungen: handelskennung, schranke: ohneEinstellung{},
	}
}

// MitHandelsschranke setzt die Handelsrichtlinie, die gilt (#2435).
func (v *Handelsverzeichnis) MitHandelsschranke(schranke Handelsschranke) *Handelsverzeichnis {
	v.schranke = schranke

	return v
}

// MitKennungen setzt die Vergabe neuer Kennungen — für den Test.
func (v *Handelsverzeichnis) MitKennungen(kennungen func() string) *Handelsverzeichnis {
	v.kennungen = kennungen

	return v
}

// Liste gibt die Vorgänge, an denen `ich` beteiligt bin — mit Lage,
// nächstem Schritt, den Namen der Gegenüber und den Marken, gefiltert und
// geordnet.
func (v *Handelsverzeichnis) Liste(
	ctx context.Context, ich string, filter domain.Handelsfilter,
) ([]domain.Handelsvorgangzeile, error) {
	if fehler := filter.Pruefe(); fehler != nil {
		return nil, fehler
	}

	zeilen, fehler := v.ablage.Uebersicht(ctx)
	if fehler != nil {
		return nil, fehler
	}

	namen, fehler := v.kontaktnamen(ctx)
	if fehler != nil {
		return nil, fehler
	}

	marken, fehler := v.marken.JeObjekt(ctx, string(domain.MarkenartHandel))
	if fehler != nil {
		return nil, fehler
	}

	eigene := make([]domain.Handelsvorgangzeile, 0, len(zeilen))

	for _, zeile := range zeilen {
		angebot := zeile.Angebot
		if !slices.Contains(angebot.Parteien, ich) {
			continue
		}

		zeile.Lage, zeile.Naechster = angebot.Lage(ich), angebot.Naechster(ich)
		zeile.Marken = marken[angebot.Vorgang]
		zeile.Gegenueber = nil

		for _, partei := range angebot.Parteien {
			if partei != ich {
				zeile.Gegenueber = append(zeile.Gegenueber, domain.Handelspartner{Kennung: partei, Name: namen[partei]})
			}
		}

		eigene = append(eigene, zeile)
	}

	return filter.Gesehen(eigene), nil
}

// kontaktnamen gibt je Kennung den Namen, unter dem ich das Gegenüber kenne
// — **den eigenen vor dem selbstgewählten**, wie in der Partnerwahl.
func (v *Handelsverzeichnis) kontaktnamen(ctx context.Context) (map[string]string, error) {
	kontakte, fehler := v.kontakte.Liste(ctx)
	if fehler != nil {
		return nil, fehler
	}

	namen := make(map[string]string, len(kontakte))

	for _, kontakt := range kontakte {
		namen[kontakt.Kennung] = kontakt.SelbstgewaehlterName
		if kontakt.EigenerName != "" {
			namen[kontakt.Kennung] = kontakt.EigenerName
		}
	}

	return namen, nil
}

// EntwurfAendern schreibt einen Entwurf neu.
//
// **Revision, Zustand, Initiator und Stimmen setzt der Dienst**: ein Entwurf
// bleibt Revision 1 dessen, der ihn anlegte, und hat keine Stimmen — was ein
// Rufer darüber sagt, zählt nicht.
func (v *Handelsverzeichnis) EntwurfAendern(
	ctx context.Context, wunsch domain.Handelsangebot,
) (domain.Handelsangebot, error) {
	heute, fehler := v.ablage.Vorgang(ctx, wunsch.Vorgang)
	if fehler != nil {
		return domain.Handelsangebot{}, fehler
	}

	if heute.Zustand != domain.HandelEntwurf {
		return domain.Handelsangebot{}, keinEntwurf()
	}

	wunsch.Vertragsfassung, wunsch.Revision, wunsch.Eltern = domain.Handelsvertragsfassung, 1, ""
	wunsch.Initiator, wunsch.Zustand, wunsch.Zustimmungen = heute.Initiator, domain.HandelEntwurf, nil

	if fehler := wunsch.Pruefe(); fehler != nil {
		return domain.Handelsangebot{}, fehler
	}

	if fehler := angebotErlaubt(ctx, v.schranke, wunsch); fehler != nil {
		return domain.Handelsangebot{}, fehler
	}

	if fehler := v.ablage.EntwurfErsetzen(ctx, wunsch); fehler != nil {
		return domain.Handelsangebot{}, fehler
	}

	return wunsch, nil
}

// keinEntwurf ist das Nein an einen Vorgang, der hinausgegangen ist.
func keinEntwurf() error {
	return domain.MitGrund(domain.ErrConflict, nil,
		"der Handel ist kein Entwurf mehr — er wird weitergeführt oder kopiert, nicht bearbeitet")
}

// Kopieren legt aus einem Vorgang einen neuen Entwurf an — mit mir als
// Initiator, ohne Stimmen und ohne Vorgeschichte (TRADE-03.AK-04: eine Kopie
// wiederholt keine alte Zustimmung).
func (v *Handelsverzeichnis) Kopieren(
	ctx context.Context, ich, vorgang string,
) (domain.Handelsangebot, error) {
	original, fehler := v.ablage.Vorgang(ctx, vorgang)
	if fehler != nil {
		return domain.Handelsangebot{}, fehler
	}

	if !slices.Contains(original.Parteien, ich) {
		return domain.Handelsangebot{}, domain.MitGrund(domain.ErrValidation, nil,
			"an diesem Handel ist dieses Profil nicht beteiligt")
	}

	parteien := []string{ich}
	for _, partei := range original.Parteien {
		if partei != ich {
			parteien = append(parteien, partei)
		}
	}

	kopie := domain.Handelsangebot{
		Vertragsfassung: domain.Handelsvertragsfassung, Vorgang: v.kennungen(), Revision: 1,
		Art: original.Art, Initiator: ich, Parteien: parteien, Zustand: domain.HandelEntwurf,
		Positionen: slices.Clone(original.Positionen), BewertungGeteilt: original.BewertungGeteilt,
	}

	if fehler := kopie.Pruefe(); fehler != nil {
		return domain.Handelsangebot{}, fehler
	}

	if fehler := angebotErlaubt(ctx, v.schranke, kopie); fehler != nil {
		return domain.Handelsangebot{}, fehler
	}

	// **Kollisionsfest angelegt** wie beim Eröffnen (Befund N2 des Nachlaufs
	// über Phase 6.8) — eine neue Kennung, die es schon gäbe, ersetzte sonst
	// einen fremden Vorgang.
	if fehler := v.ablage.Fortschreiben(ctx, domain.Handelsangebot{}, kopie); fehler != nil {
		return domain.Handelsangebot{}, fehler
	}

	return kopie, nil
}

// EntwurfLoeschen löscht einen Entwurf — was hinausgegangen ist, bleibt.
func (v *Handelsverzeichnis) EntwurfLoeschen(ctx context.Context, vorgang string) error {
	return v.ablage.EntwurfLoeschen(ctx, vorgang)
}

// Archivieren legt einen **beendeten** Vorgang ins Archiv oder holt einen
// heraus. Gelöscht wird nichts: ein abgeschlossener Handel ist ein Beleg.
func (v *Handelsverzeichnis) Archivieren(ctx context.Context, vorgang string, an bool) error {
	heute, fehler := v.ablage.Vorgang(ctx, vorgang)
	if fehler != nil {
		return fehler
	}

	if an && !domain.HandelEnde(heute.Zustand) {
		return domain.MitGrund(domain.ErrConflict, nil,
			"ins Archiv geht nur ein Handel, der zu Ende ist")
	}

	return v.ablage.Archivieren(ctx, vorgang, an)
}
