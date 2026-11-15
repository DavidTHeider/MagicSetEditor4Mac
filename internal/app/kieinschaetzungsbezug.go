package app

// **Wonach gefragt wird, wenn jemand eine Einschätzung will** — Werkstück
// A13.2a von Phase 5.3
// (#1452),
// herausgeschnitten am 2026-09-13 mit A13.2b.
//
// **Der Schnitt folgt der Frage** (§2): hier steht, **was ein Ruf mitbringt
// und wohin er geht** — der Bezug, die Prüfung der Art, die Belege und der
// Recherchefall dahinter. Nebenan steht, **was zurückkommt**: die Antwort,
// ihr Schema und die Art, sie zu lesen.

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// KIEinschaetzen lässt ein Modell eine Karte beurteilen — mit Belegen.
//
// **Es schreibt nichts.** Weder an der Karte noch an der Sammlung noch
// irgendwo sonst; das Issue führt es als eigene Abnahme („Keine Regel-/
// Sammlungs-/Spielzustandsänderung beim Analysieren"), und ein Lauf hält es.
// **Der Nachweis ist die Freigabe eines Menschen** — dieselbe wie beim
// Vorschlag ([0078](../../docs/entscheidungen/0078-der-wirksame-akteur.md)).
// Ein Modell auf diesem Rechner braucht keine; ein Cloudanbieter schon, und
// ohne sie scheitert der Ruf an der Kostenwache. **Die drei Flächen ohne
// Fenster reichen die leere Zeichenkette** — dort stellt niemand eine aus.
func (a *App) KIEinschaetzen(
	ctx context.Context, bezug KIEinschaetzungsbezug,
) KIEinschaetzungAntwort {
	ctx = a.imProfil(ctx)

	// **Die Rolle vor dem Modell** — dieselbe Frage wie bei
	// [App.KIVorschlagen] (ID-03, #1702).
	if verweigert := a.darf(ctx, domain.WirkungSchreiben); verweigert != nil {
		return mitFehler(ctx, verweigert)
	}

	// **Die Art zuerst** — sie ist die billigste Prüfung, und eine erfundene
	// darf keinen Ruf auslösen, der Geld kostet.
	gewaehlt := domain.Einschaetzungsart(bezug.Art)
	if fehler := gewaehlt.Pruefe(); fehler != nil {
		return mitFehler(ctx, fehler)
	}

	profil, fehler := a.profilDiesesRufs(ctx)
	if fehler != nil {
		return mitFehler(ctx, fehler)
	}

	paket, fehler := a.hinweispaket(ctx, gewaehlt, bezug)
	if fehler != nil {
		return mitFehler(ctx, fehler)
	}

	werk, quelle, fehler := a.kiwerkzeug(ctx, bezug.Anbieter,
		kideckung{nachweis: bezug.Nachweis})
	if fehler != nil {
		return mitFehler(ctx, fehler)
	}

	return a.einschaetzungMit(ctx, einschaetzungslauf{
		art: gewaehlt, profil: profil, bezug: bezug, paket: paket,
	}, kiwerkzeugsatz{werk: werk, quelle: quelle})
}

// KIEinschaetzungsbezug ist, was für eine Einschätzung gebraucht wird.
//
// **Ein Verbund und keine sieben Stellungsparameter** (A13.2b, #1452): mit
// dem Format wären es acht, und wer zwei davon vertauscht, merkt es an keiner
// Stelle — der Typ ist derselbe. Dieselbe Begründung wie bei
// [App.MitKIAnbietern].
type KIEinschaetzungsbezug struct {
	// Art ist `regel` oder `balance`.
	Art string

	KarteID int64

	Anbieter string
	Alias    string

	// Frage geht **auch** an den Suchindex.
	Frage string

	// Format ist, worauf sich ein Vergleich bezieht — leer heißt: keines
	// genannt, und dann sagt die Antwort es (A13.2b).
	//
	// **Nur für `balance`.** Ein Regelhinweis gilt nach den Regeln und nicht
	// nach einem Format; die deterministische Formatprüfung bleibt ein
	// eigener Weg.
	Format string

	// Nachweis ist die Freigabe eines Menschen — leer heißt: keine.
	Nachweis string

	// Freigegeben nennt die **privaten** Korpora, die mitgesucht werden
	// dürfen — Werkstück A13.4 von Phase 5.3
	// (#1452).
	//
	// # Warum eine Balanceanalyse sie braucht
	//
	// **Weil sie über [domain.FallAehnlich] geht**, und der sucht in den
	// eigenen Karten. Bis A13.4 tat er es von selbst: wer eine Karte
	// vergleichen ließ, schickte dabei seine eigenen mit, ohne es zu sagen.
	// Jetzt bleibt der eigene Bestand außen, solange hier nichts steht — und
	// die Einschränkungen der Antwort sagen es.
	Freigegeben []string
}

// einschaetzungslauf ist, was ein Lauf mitbringt.
type einschaetzungslauf struct {
	art    domain.Einschaetzungsart
	profil int64
	bezug  KIEinschaetzungsbezug
	paket  service.Kontextpaket
}

// hinweispaket stellt die Belege zusammen — über den Dienst aus A16.
func (a *App) hinweispaket(
	ctx context.Context, art domain.Einschaetzungsart,
	bezug KIEinschaetzungsbezug,
) (service.Kontextpaket, error) {
	if a.recherche == nil {
		return service.Kontextpaket{}, fehlenderBestand()
	}

	return a.recherche.Zusammenstellen(ctx, service.Kontextwunsch{
		Aufgabe: bezug.Frage,
		Anfrage: domain.Rechercheanfrage{
			Fall: fallFuer(art), Frage: bezug.Frage,
			AusgangskarteID: bezug.KarteID,
			Grenze:          KIEinschaetzungstreffer,

			// **Die Freigabe reist mit** (A13.4, #1452): das Paket zeigt
			// danach, was wirklich darin steht.
			Freigegeben: freigegebeneQuellen(bezug.Freigegeben),
		},
		Zielanbieter: bezug.Anbieter, Zielmodell: bezug.Alias,
		Kontextgrenze:    KIEinschaetzungsfenster,
		Ausgabereserve:   KIEinschaetzungsreserve,
		HoechstensBelege: KIEinschaetzungsbelege,
	})
}

// fallFuer sagt, welche Recherche zu einer Art gehört.
//
// **Am Typ und nicht beim Aufrufer** — dieselbe Begründung wie bei
// [domain.Recherchefall.Quellen]: welche Korpora eine Frage liest, ist eine
// Eigenschaft der Frage.
func fallFuer(art domain.Einschaetzungsart) domain.Recherchefall {
	if art == domain.EinschaetzungBalance {
		return domain.FallAehnlich
	}

	return domain.FallRegelhinweis
}

// KIEinschaetzen ist die Fensterbindung.
func (b *Bindungen) KIEinschaetzen(
	ctx context.Context,
	bezug KIEinschaetzungsbezug,
) KIEinschaetzungAntwort {
	return b.app(ctx).KIEinschaetzen(b.ruf(ctx), bezug)
}
