package app

// **Was vor „Bereit" durchgesetzt wird** — Werkstück G4 von Phase 5.5
// (#1461).
//
// # Warum diese Auskunft eine eigene Bindung ist
//
// **Weil das Issue sie vor dem Start verlangt**: „Spieler sehen vor Start,
// welche Regeln automatisch durchgesetzt werden." Nach dem Anlegen wäre es zu
// spät — dann steht die Partie, und die Frage war, ob sie so stehen soll.
//
// **Rein lesend und ohne Partie**: sie rechnet über ein Regelprofil und eine
// gewünschte Strenge. Damit lässt sie sich stellen, bevor es etwas gibt,
// worüber man sie stellen könnte.

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// MatrixeintragDTO ist die Aussage über einen Regelbereich.
type MatrixeintragDTO struct {
	Bereich string `json:"bereich"`

	// Grad ist `unbekannt`, `manuell`, `begleitet` oder `automatisch`.
	Grad string `json:"grad"`

	// Quelle ist leer, wo nichts zugesagt ist — `CR 400 (Fassung …)`.
	Quelle string `json:"quelle"`

	// Grenzen nennt, was ausdrücklich **nicht** abgedeckt ist.
	Grenzen []string `json:"grenzen"`
}

// RegelfreigabeDTO ist die Antwort auf „darf es so losgehen?".
type RegelfreigabeDTO struct {
	Erlaubt bool `json:"erlaubt"`

	// Regelwerk und Fassung nennen das Profil, über das gerechnet wurde.
	Regelwerk string `json:"regelwerk"`
	Fassung   int    `json:"fassung"`

	// Wirksam ist das Datum, ab dem diese Fassung gilt.
	Wirksam string `json:"wirksam"`

	// Hoechste ist die strengste Stufe, die dieses Regelwerk trägt.
	Hoechste string `json:"hoechste"`

	// Matrix ist die Aussage über jeden der dreizehn Bereiche.
	Matrix []MatrixeintragDTO `json:"matrix"`

	// Luecken sind die Bereiche, die `streng` nicht tragen.
	Luecken []string `json:"luecken"`

	// Hindernisse sind die Gründe, warum es nicht geht — leer, wenn es geht.
	Hindernisse []string `json:"hindernisse"`

	// Modus ist der Bedienmodus, über den gerechnet wurde — PLAY-02.AK-01
	// (#1710).
	//
	// **Er steht in der Antwort**, damit ein Aufrufer merkt, dass er die
	// Vorgabe bekommen hat: wer nichts sagt, bekommt `anfaenger`.
	Modus string `json:"modus"`

	// Baum ordnet die Bereiche nach Ablauf und Mechanik (AK-02).
	Baum []RegelgruppeDTO `json:"baum"`

	// Widersprueche sind die Abwahlen, an denen noch etwas hängt.
	//
	// **Sie verhindern nichts** — sie erklären (AK-02).
	Widersprueche []RegelwiderspruchDTO `json:"widersprueche"`
}

// RegelgruppeDTO ist ein Ast des Regelbaums.
type RegelgruppeDTO struct {
	Kennung  string   `json:"kennung"`
	Bereiche []string `json:"bereiche"`
}

// RegelwiderspruchDTO ist eine Abwahl, an der noch etwas hängt.
type RegelwiderspruchDTO struct {
	// Bereich ist der abgewählte, Haengt der, der ihn braucht.
	Bereich string `json:"bereich"`
	Haengt  string `json:"haengt"`

	// Quelle belegt die Abhängigkeit — `CR 608 (Fassung …)`.
	Quelle string `json:"quelle"`
}

// RegelfreigabeAntwort trägt die Freigabe oder einen Fehler.
type RegelfreigabeAntwort struct {
	Freigabe RegelfreigabeDTO `json:"freigabe"`

	Fehler *Fehler `json:"fehler"`
}

// Regelfreigabe sagt, ob eine Partie so anfangen darf.
//
// **Sie nennt auch dann die höchste Stufe und die Lücken, wenn sie ja sagt**
// — wer „begleitet" bekommt, soll sehen, dass „streng" nicht zur Wahl steht.
func (a *App) Regelfreigabe(
	ctx context.Context, spiel, strenge string, sitze int,
	modus string, abgewaehlt []string,
) RegelfreigabeAntwort {
	ctx = a.imProfil(ctx)

	if fehler := a.darf(ctx, domain.WirkungLesen); fehler != nil {
		return leereFreigabe(ctx, fehler)
	}

	profil := domain.Grundprofil(spiel)

	if modus == "" {
		modus = domain.ModusAnfaenger
	}

	freigabe, fehler := service.NeueRegelfreigabe().Pruefe(
		service.Freigabewunsch{
			Profil:  profil,
			Strenge: domain.Strenge(strenge),
			Sitze:   sitze,
			Modus:   modus,
			Wahl: domain.Regelwahl{
				Abgewaehlt: bereicheAus(abgewaehlt),
			},
		})
	if fehler != nil {
		return leereFreigabe(ctx, fehler)
	}

	return RegelfreigabeAntwort{Freigabe: RegelfreigabeDTO{
		Erlaubt:   freigabe.Erlaubt,
		Regelwerk: profil.Kennung,
		Fassung:   profil.Fassung,
		Wirksam:   profil.Wirksam,
		Hoechste:  string(freigabe.Hoechste),

		// **Die wirksame Matrix und nicht die des Profils** — wer die des
		// Profils zeigte, zeigte eine Zusage, die für diese Runde nicht
		// gilt (§10).
		Matrix:        alsMatrix(freigabe.Matrix),
		Luecken:       alsBereiche(freigabe.Luecken),
		Hindernisse:   freigabe.Hindernisse,
		Modus:         modus,
		Baum:          alsBaum(freigabe.Baum),
		Widersprueche: alsWidersprueche(freigabe.Widersprueche),
	}}
}

// alsMatrix macht aus der Matrix die Liste, die das Fenster zeigt.
//
// **In der Reihenfolge der Bereiche und nicht der Einträge** — eine Liste,
// deren Reihenfolge von der Eingabe abhängt, liest sich bei jedem Aufruf
// anders.
func alsMatrix(matrix domain.Regelmatrix) []MatrixeintragDTO {
	gezeigt := make([]MatrixeintragDTO, 0, len(domain.Regelbereiche()))

	for _, bereich := range domain.Regelbereiche() {
		dto := MatrixeintragDTO{
			Bereich: string(bereich),
			Grad:    string(matrix.Grad(bereich)),
			Grenzen: []string{},
		}

		if eintrag, da := matrix.Eintrag(bereich); da {
			if len(eintrag.Quelle.Nummern) > 0 {
				dto.Quelle = eintrag.Quelle.String()
			}

			dto.Grenzen = append(dto.Grenzen, eintrag.Grenzen...)
		}

		gezeigt = append(gezeigt, dto)
	}

	return gezeigt
}

// alsBereiche macht aus Bereichen Zeichenketten.
func alsBereiche(bereiche []domain.Regelbereich) []string {
	namen := make([]string, 0, len(bereiche))
	for _, bereich := range bereiche {
		namen = append(namen, string(bereich))
	}

	return namen
}

// leereFreigabe gibt eine Antwort mit leeren Listen und dem übersetzten
// Fehler — **dieselbe Zusage wie auf dem Erfolgsweg**.
func leereFreigabe(ctx context.Context, fehler error) RegelfreigabeAntwort {
	return RegelfreigabeAntwort{
		Freigabe: RegelfreigabeDTO{
			// **Auch Baum und Widersprüche tragen ihre leeren Listen**
			// (PLAY-02, #1710) — dieselbe Zusage wie auf dem Erfolgsweg.
			Baum:          []RegelgruppeDTO{},
			Widersprueche: []RegelwiderspruchDTO{},

			Matrix:      []MatrixeintragDTO{},
			Luecken:     []string{},
			Hindernisse: []string{},
		},
		Fehler: Uebersetze(ctx, fehler),
	}
}

// alsBaum bildet die Gruppen ab.
func alsBaum(gruppen []domain.Regelgruppe) []RegelgruppeDTO {
	abgebildet := make([]RegelgruppeDTO, 0, len(gruppen))

	for _, gruppe := range gruppen {
		abgebildet = append(abgebildet, RegelgruppeDTO{
			Kennung:  gruppe.Kennung,
			Bereiche: alsBereiche(gruppe.Bereiche),
		})
	}

	return abgebildet
}

// alsWidersprueche bildet die Widersprüche ab.
//
// **Eine leere Liste und kein `nil`** — „nichts widerspricht sich" ist eine
// Antwort, und sie sieht anders aus als keine.
func alsWidersprueche(
	widersprueche []domain.Regelwiderspruch,
) []RegelwiderspruchDTO {
	abgebildet := make([]RegelwiderspruchDTO, 0, len(widersprueche))

	for _, widerspruch := range widersprueche {
		abgebildet = append(abgebildet, RegelwiderspruchDTO{
			Bereich: string(widerspruch.Bereich),
			Haengt:  string(widerspruch.Haengt),
			Quelle:  widerspruch.Quelle.String(),
		})
	}

	return abgebildet
}
