package service

// **Der Wirt fragt einen Gast nach dem Bild seiner Karte** — P9b
// (#2377),
// Vermerk 0198.
//
// # Warum über den Platz und die Post des Sitzes
//
// **Weil Fragen und Antworten zwei Wege gehen.** Hinaus geht der Wunsch über
// den Platz in der Runde — er hält die Leitung zu diesem Gast; zurück kommt
// die Antwort über den Wirt dieses Sitzes, der jede Nachricht des Gastes
// empfängt. Beide teilen eine [Bildpost], wie Gast und Empfangsschleife auf
// der anderen Seite.

import (
	"context"
	"fmt"
	"slices"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// MitBildpost setzt die Post, über die dieser Platz Bildantworten bekommt.
func (g *Gastplatz) MitBildpost(post *Bildpost) *Gastplatz {
	g.post = post

	return g
}

// Bildpost gibt die Post dieses Platzes — der Wirt seines Sitzes stellt ihr
// zu, was der Gast antwortet.
func (g *Gastplatz) Bildpost() *Bildpost { return g.post }

// MitBildpost setzt die Post, der dieser Wirt Bildantworten zustellt.
func (w *Partiewirt) MitBildpost(post *Bildpost) *Partiewirt {
	w.post = post

	return w
}

// Bild fragt den Gast an diesem Sitz nach einem Kartenbild.
func (r *Partierunde) Bild(
	ctx context.Context, sitz string, wunsch domain.Bildwunsch,
) (domain.Bildantwort, error) {
	r.schloss.Lock()
	platz := r.plaetze[sitz]
	r.schloss.Unlock()

	if platz == nil || platz.post == nil {
		return domain.Bildantwort{}, fmt.Errorf(
			"%w: an diesem Sitz sitzt gerade niemand, der das Bild zeichnet",
			domain.ErrConflict)
	}

	return platz.post.frage(ctx, wunsch, func(ctx context.Context, inhalt []byte) error {
		return platz.schicke(ctx, domain.ArtKontrolle, inhalt)
	}, "der Gast")
}

// bildNehmen stellt zu, was der Gast auf eine Bildfrage antwortet.
//
// **Eine späte Antwort reißt nichts ab** — wer nicht mehr wartet, hat die
// Frist verstreichen lassen. **Eine falsche Gestalt schon**: sie ist keine
// Antwort, sondern ein Fehler des Gastes.
func (w *Partiewirt) bildNehmen(roh []byte) error {
	antwort, fehler := liesBildantwort(roh)
	if fehler != nil {
		return fehler
	}

	w.post.liefere(antwort)

	return nil
}

// BildVomEigner fragt den Gast, dem das Objekt gehört — am Fenster des Wirts,
// das eine Karte aus einem mitgebrachten Deck zeigt (P9b, #2377).
//
// **Nur nach einem Objekt, das der Wirt selbst offen sieht**: die Sicht ist
// die seines Fensters (`betrachter`), und was darin verdeckt liegt, bekommt
// er auch über diesen Weg nicht zu sehen. **Und nur, wenn der Gast es
// zeichnet** (Befund 6) — sonst steht ein Grund in der Antwort.
func (r *Partierunde) BildVomEigner(
	ctx context.Context, leser Sichtleser, partie, betrachter string,
	wunsch domain.Bildwunsch,
) (domain.Bildantwort, error) {
	sicht, fehler := leser.Partie(ctx, partie, betrachter)
	if fehler != nil {
		return domain.Bildantwort{}, fehler
	}

	gesehen, da := sichtobjekt(sicht, wunsch.Objekt)
	if !da || gesehen.Geheim || gesehen.Bild == "" || !bildPasst(gesehen, wunsch) {
		return domain.Bildantwort{}, fmt.Errorf(
			"%w: dieses Objekt liegt nicht offen auf deinem Tisch", domain.ErrConflict)
	}

	if !r.zeichnetSelbst(ctx, leser, partie, gesehen.Eigner, wunsch.Objekt) {
		return domain.Bildantwort{Wunsch: wunsch, Grund: drittquelle}, nil
	}

	return r.Bild(ctx, gesehen.Eigner, wunsch)
}

// Sichtleser rechnet eine Sicht — der Wirt eines Sitzes hat einen, das
// Fenster des Wirts auch.
type Sichtleser interface {
	Partie(ctx context.Context, kennung, fuer string) (domain.Partiesicht, error)
}

// drittquelle ist der Grund, wenn ein Deck aus einem Bestand kommt, der nicht
// am Tisch sitzt.
const drittquelle = "dieses Bild zeichnet niemand am Tisch — das Deck " +
	"stammt aus einem Bestand, der nicht mitspielt"

// zeichnetSelbst sagt, ob der Gast am Sitz `sitz` das Objekt aus seinem
// **eigenen** Bestand zeichnet — Zweitblick über Phase 7, Befund 6.
//
// **Nur ihn zu fragen ist die Fähigkeitsprobe**: eine Zeile in seinem
// Bestand hat nur, wer sein Deck mitgebracht hat, und mitbringen kann erst
// ein Gast ab 0.9.0 (Vermerk 0199). Ein älterer an einem vorbelegten Sitz —
// mit einer Leihe aus einem dritten Bestand — bekäme sonst eine
// Kontrollart, die er nicht kennt, und seine Sitzung endete.
//
// **Und nur, wenn er es erklärt hat** (Nachlauf, R4): eine Zeile in seinem
// Bestand hat auch ein älterer Gast, dem der Wirt eine Leihe **von ihm**
// vorgelegt hat — Bestandseigentum ist kein Fähigkeitsnachweis.
func (r *Partierunde) zeichnetSelbst(
	ctx context.Context, leser Sichtleser, partie, sitz, objekt string,
) bool {
	peer := r.Peer(sitz)
	if peer == "" || !r.Zeichnet(sitz) {
		return false
	}

	sicht, fehler := leser.Partie(MitBestandssicht(ctx, peer), partie, sitz)
	if fehler != nil {
		return false
	}

	gesehen, da := sichtobjekt(sicht, objekt)

	return da && gesehen.KarteID != 0
}

// Kann trägt ein, was der Gast an einem Sitz erklärt hat (R4) — aus seinem
// Rückkehrwunsch. **Für den Platz, der gerade dort sitzt**: wer neu
// verbindet, bekommt einen neuen und erklärt es wieder.
func (r *Partierunde) Kann(sitz string, kann []string) {
	r.schloss.Lock()
	defer r.schloss.Unlock()

	if platz := r.plaetze[sitz]; platz != nil {
		platz.bildfaehig = slices.Contains(kann, domain.FaehigkeitBild)
	}
}

// Zeichnet sagt, ob der Gast an einem Sitz Bildwünsche beantwortet.
func (r *Partierunde) Zeichnet(sitz string) bool {
	r.schloss.Lock()
	defer r.schloss.Unlock()

	platz := r.plaetze[sitz]

	return platz != nil && platz.bildfaehig
}
