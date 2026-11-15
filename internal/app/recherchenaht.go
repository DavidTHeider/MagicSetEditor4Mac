package app

// **Die Naht zur Recherche** — Werkstück A16 von Phase 5.3
// (#1515).
//
// # Sie hängt am Suchindex und nicht daneben
//
// **A16 baut keinen zweiten Retrievalkern.** Was hier verdrahtet wird, ist
// die [service.Retrievalverwaltung] aus A15 als Recherchequelle — dieselbe
// Verwaltung, die auch `Indexsuche` bedient.
//
// # Und die Ausgangskarte kommt aus dem Bestand
//
// **Mit ihrem Stand**, und der wird dort geprüft: eine Karte, die sich seit
// der Anfrage geändert hat, ist eine andere Frage.

import (
	"context"
	"database/sql"
	"strconv"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// rechercheprofil liest die Profilkennung aus dem Kontext.
//
// **Der Fachkern liest sie nicht selbst** (§2): welche Kennung im Kontext
// steht, weiß `storage`.
type rechercheprofil struct{}

func (rechercheprofil) Profil(ctx context.Context) (int64, error) {
	return storage.ProfilAus(ctx)
}

// kartenausgang liest die Ausgangskarte einer Ähnlichkeitsfrage.
type kartenausgang struct{ holen func() *sql.DB }

// Ausgangstext gibt den Text der Karte und die Kennungen ihrer Abschnitte.
//
// # Warum der Stand hier geprüft wird
//
// **Weil hier die Karte liegt.** Der Dienst reicht den Stand durch; wer ihn
// dort prüfte, prüfte eine Zahl gegen eine Zahl. Hier steht er gegen die
// Karte, die es gibt.
func (k kartenausgang) Ausgangstext(
	ctx context.Context, karteID int64, stand string,
) (text, jetziger string, kennungen []string, fehler error) {
	zeile, fehler := storage.NeuerSetLeser(k.holen()).KarteLesen(ctx, karteID)
	if fehler != nil {
		return "", "", nil, fehler
	}

	// **Derselbe gerechnete Stand wie in A6** ([domain.CardRow.Stand]): zwei
	// Arten, ihn zu bilden, gäben zwei Antworten auf dieselbe Frage.
	jetziger = zeile.Stand()

	if stand != "" && jetziger != stand {
		return "", "", nil, domain.MitGrund(domain.ErrConflict, nil,
			"die Ausgangskarte stand bei der Anfrage auf %q und steht jetzt "+
				"auf %q — Ähnliches zu einem Stand zu suchen, den es nicht "+
				"mehr gibt, gäbe eine Antwort auf eine andere Frage",
			stand, jetziger)
	}

	return ausgangstextAus(zeile.Values), jetziger,
		eigenkartenkennungen(karteID), nil
}

// ausgangstextAus setzt den Text zusammen, mit dem gesucht wird.
//
// **Alle Feldwerte und keine Auswahl.** Welches Feld den Text einer Karte
// trägt, weiß das **Paket** und nicht dieses Programm; eine hier getroffene
// Wahl wäre über jede fremde Vorlage falsch (§11).
// **Die Reihenfolge ist die der Zeile** und nicht sortiert: sie ist die des
// Pakets, und derselbe Text zweimal anders zusammengesetzt gäbe zwei Vektoren
// für dieselbe Karte.
func ausgangstextAus(werte []domain.Value) string {
	teile := make([]string, 0, len(werte))

	for _, wert := range werte {
		if gekuerzt := strings.TrimSpace(wert.Text); gekuerzt != "" {
			teile = append(teile, gekuerzt)
		}
	}

	return strings.Join(teile, "\n")
}

// eigenkartenkennungen gibt die Indexkennungen einer eigenen Karte.
//
// **Sie folgen der Gestalt aus A15** (`<system>:<kennung>:<abschnitt>`); der
// Stamm genügt, weil `domain.OhneAusgangskarte` über ihn ausschließt.
func eigenkartenkennungen(karteID int64) []string {
	return []string{"eigenekarte:" + strconv.FormatInt(karteID, 10) + ":0"}
}

// RechercheUeber baut die Recherche über einem **vorhandenen** Suchindex.
//
// # Warum die Verwaltung hereingereicht wird
//
// **Weil A16 keinen zweiten Retrievalkern baut.** Der Satz steht so in der
// Verdrahtung — und bis zum 2026-09-12 rief diese Funktion selbst
// [RetrievalUeber] und legte damit eine **zweite** [service.Retrievalverwaltung]
// über denselben Index.
//
// **Was das gekostet hätte, ist der Riegel.** Die Verwaltung hält den Riegel
// gegen einen zweiten Aufbau **im selben Prozess**; zwei Verwaltungen haben
// zwei Riegel, und der Schutz wäre die Zusage, die keiner einlöst. Heute baut
// über diese Naht niemand — genau das macht es zur Falle für den, der es
// morgen tut.
func RechercheUeber(
	quelle service.Recherchequelle, holen func() *sql.DB,
) *Recherchewerk {
	recherche := service.NeuerRecherchedienst(
		quelle, kartenausgang{holen: holen}, rechercheprofil{})

	return &Recherchewerk{
		recherche: recherche,
		kontext:   service.NeuerKontextdienst(recherche, rechercheprofil{}),
	}
}

// Recherchewerk führt Recherche und Kontext zusammen.
//
// **Zwei Dienste an einer Naht**, weil sie an einer Fläche als zwei Rufe
// erscheinen und derselben Frage gelten: was steht da, und was davon geht an
// ein Modell.
type Recherchewerk struct {
	recherche *service.Recherchedienst
	kontext   *service.Kontextdienst
}

// Recherchieren beantwortet eine Frage.
func (w *Recherchewerk) Recherchieren(
	ctx context.Context, anfrage domain.Rechercheanfrage, fortsetzung string,
) (domain.Rechercheergebnis, error) {
	return w.recherche.Recherchieren(ctx, anfrage, fortsetzung)
}

// Zusammenstellen baut ein Kontextpaket.
func (w *Recherchewerk) Zusammenstellen(
	ctx context.Context, wunsch service.Kontextwunsch,
) (service.Kontextpaket, error) {
	return w.kontext.Zusammenstellen(ctx, wunsch)
}

// rechercheVerdrahten hängt die Recherche ins Fenster.
//
// **Über die Verwaltung des Suchindex** — und sie hängt sie selbst ein, wenn
// noch keine da ist. Eine Reihenfolge, die in einer anderen Datei steht, ist
// eine Zusage ohne Wächter: wer sie dreht, bekäme eine Recherche über einer
// leeren Schnittstelle und den Sturz beim ersten Ruf.
func (a *App) rechercheVerdrahten() {
	if a.retrieval == nil {
		a.retrievalVerdrahten()
	}

	a.MitRecherche(RechercheUeber(a.retrieval, a.laufendeVerbindung))
}

// RechercheAnwendungUeber baut eine Anwendung, die recherchieren kann.
func RechercheAnwendungUeber(db *sql.DB) *App {
	anwendung := RetrievalAnwendungUeber(db)
	anwendung.rechercheVerdrahten()

	return anwendung
}
