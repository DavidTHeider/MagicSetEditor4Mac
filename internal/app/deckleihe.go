package app

// **Die digitale Deckleihe an der Fassade** — Werkstück G6 von Phase 5.5
// (#1463).
//
// # Warum eine Leihe als Umschlag reist
//
// **Weil es dafür schon einen Vertrag gibt** (N4,
// #1457),
// und die Abnahme verlangt ihn: „Datei-/LAN-/Internettransfer verwenden
// denselben N4-Vertrag." Der Umschlag prüft Größen, Pfade und Namen, **bevor**
// er auspackt, und er trägt eine Unterschrift über sein Manifest.
//
// **Zwei Unterschriften, und sie beantworten zwei Fragen.** Die des Umschlags
// sagt, **wer geschickt** hat; die der Leihe sagt, **wer verliehen** hat. Sie
// fallen meist zusammen und müssen es nicht — ein Dokument darf weitergereicht
// werden, und wer es weitergibt, erwirbt dadurch kein Recht daran.
//
// # Was hier ausdrücklich nicht geschieht
//
// **Nichts geht in die Sammlung.** Eine angenommene Leihe legt das Dokument
// und den Abzug ab und sonst nichts; das Issue sagt es: „die Leihe importiert
// nicht automatisch eine dauerhafte Kopie in die Sammlung." Wer den Abzug als
// Deck haben will, führt ihn ein — über den gewöhnlichen Weg, mit der
// Übernahmefreigabe, die dort steht.

import (
	"context"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// LeiheDTO ist eine Leihe, wie das Fenster sie sieht.
type LeiheDTO struct {
	Kennung string `json:"kennung"`

	// Verleiher und Entleiher sind **Peerkennungen**.
	Verleiher string `json:"verleiher"`
	Entleiher string `json:"entleiher"`

	// Abdruck ist der Hash des Abzugs — **er sagt nichts über die Karten**,
	// er sagt, dass zwei Seiten dasselbe meinen.
	Abdruck string `json:"abdruck"`

	// Bindungsart ist `partie` oder `turnier`, Woran die Kennung — leer
	// heißt: die nächste dieser Art.
	Bindungsart string `json:"bindungsart"`
	Woran       string `json:"woran"`

	Rechte []string `json:"rechte"`

	Ausgestellt string `json:"ausgestellt"`
	Bis         string `json:"bis"`

	// Modus sagt, was mit einer laufenden Partie geschieht: `zuende` oder
	// `abbruch`.
	Modus string `json:"modus"`

	// Verliehen sagt, ob **dieses** Profil sie ausgestellt hat.
	Verliehen bool `json:"verliehen"`

	// Verbraucht ist die Partie, an der sie eingelöst wurde — leer heißt:
	// noch nicht.
	Verbraucht string `json:"verbraucht"`

	Widerrufen bool `json:"widerrufen"`

	// Karten ist die Zahl der Karten im Abzug — **eine Zahl und keine
	// Liste**: wer eine Leihe verwaltet, will wissen, wie groß sie ist.
	Karten int `json:"karten"`
}

// DeckleiheAntwort trägt die Leihen dieses Profils.
type DeckleiheAntwort struct {
	Leihen []LeiheDTO `json:"leihen"`

	// Pfad ist der abgelegte Umschlag — nur beim Anbieten.
	Pfad string `json:"pfad"`

	// Nachladefehler sagt, dass der Schritt **geschrieben** ist und nur die
	// Liste danach nicht zu lesen war (§7, Regel 3; dieselbe Klasse wie
	// [PeerkontakteAntwort.Nachladefehler], #1960) — dann ist `leihen` leer
	// und `fehler` ist es auch.
	Nachladefehler *Fehler `json:"nachladefehler,omitempty"`

	Fehler *Fehler `json:"fehler"`
}

// Deckleihen gibt alles, was dieses Profil an Leihen weiß.
func (a *App) Deckleihen(ctx context.Context) DeckleiheAntwort {
	ctx = a.imProfil(ctx)

	if fehler := a.darf(ctx, domain.WirkungLesen); fehler != nil {
		return leereLeihen(ctx, fehler)
	}

	staende, fehler := a.leihstelle().Leihen(ctx)
	if fehler != nil {
		return leereLeihen(ctx, fehler)
	}

	return DeckleiheAntwort{Leihen: leihenAus(staende)}
}

// DeckleiheAnbieten stellt eine Leihe aus und legt sie als Umschlag ab.
//
// **Die Dauer kommt in Stunden.** Eine Dauer und kein Zeitpunkt — der Grund
// steht bei [service.Leihwunsch.Dauer]; Stunden, weil das Fenster keine
// `time.Duration` kennt.
func (a *App) DeckleiheAnbieten(
	ctx context.Context, wunsch LeihwunschDTO, ziel string,
) DeckleiheAntwort {
	ctx = a.imProfil(ctx)

	// **`externSenden`, weil ein Dokument hinausgeht**, das einem fremden
	// Peer etwas erlaubt. Es verlässt diesen Rechner in dem Augenblick, in
	// dem es abgelegt wird — was danach damit geschieht, entscheidet niemand
	// hier mehr.
	if fehler := a.darf(ctx, domain.WirkungExternSenden); fehler != nil {
		return leereLeihen(ctx, fehler)
	}

	pfad, fehler := a.leiheAblegen(ctx, wunsch, ziel)
	if fehler != nil {
		return leereLeihen(ctx, fehler)
	}

	antwort := a.leihenNachDemSchreiben(ctx)
	antwort.Pfad = pfad

	return antwort
}

// DeckleiheAnnehmen liest einen Umschlag und nimmt die Leihe darin an.
func (a *App) DeckleiheAnnehmen(
	ctx context.Context, pfad string,
) DeckleiheAntwort {
	ctx = a.imProfil(ctx)

	// **`externEmpfangen`**: von hier an liegt ein fremdes Dokument in
	// diesem Bestand, und es erlaubt etwas.
	if fehler := a.darf(ctx, domain.WirkungExternEmpfangen); fehler != nil {
		return leereLeihen(ctx, fehler)
	}

	if fehler := a.leiheAnnehmen(ctx, pfad); fehler != nil {
		return leereLeihen(ctx, fehler)
	}

	return a.leihenNachDemSchreiben(ctx)
}

// DeckleiheWiderrufen nimmt eine ausgestellte Leihe zurück.
func (a *App) DeckleiheWiderrufen(
	ctx context.Context, kennung string,
) DeckleiheAntwort {
	ctx = a.imProfil(ctx)

	return a.leiheAendern(ctx, func(ctx context.Context) error {
		return a.leihstelle().Widerrufen(ctx, kennung)
	})
}

// DeckleiheZurueckgeben trägt ein, dass eine geliehene nicht mehr gilt.
func (a *App) DeckleiheZurueckgeben(
	ctx context.Context, kennung string,
) DeckleiheAntwort {
	ctx = a.imProfil(ctx)

	return a.leiheAendern(ctx, func(ctx context.Context) error {
		return a.leihstelle().Zurueckgeben(ctx, kennung)
	})
}

// leiheAendern fährt eine Änderung und gibt danach den Stand.
//
// **Das Profil setzt der Aufrufer** und nicht dieser Helfer: ein Wächter
// verlangt die Zeile in der **Bindung** selbst
// (`internal/werkzeuge/profilnaht_test.go`), und er hat recht — wer sie
// hinter einen Helfer schöbe, hätte sie beim nächsten Weg vergessen, ohne
// dass es auffiele.
func (a *App) leiheAendern(
	ctx context.Context, arbeit func(context.Context) error,
) DeckleiheAntwort {
	if fehler := a.darf(ctx, domain.WirkungSchreiben); fehler != nil {
		return leereLeihen(ctx, fehler)
	}

	if fehler := arbeit(ctx); fehler != nil {
		return leereLeihen(ctx, fehler)
	}

	return a.leihenNachDemSchreiben(ctx)
}

// leihenNachDemSchreiben gibt die Liste nach einem **geglückten** Schritt.
//
// **Geschrieben ist geschrieben** (§7, Regel 3): scheitert das Lesen danach,
// steht sein Fehler in [DeckleiheAntwort.Nachladefehler] und nicht in
// `fehler`. Wer `fehler` liest, hält den Schritt für ungeschehen und
// wiederholt ihn — gegen eine Leihe, die schon angenommen oder widerrufen
// ist. Dieselbe Klasse und dieselbe Form wie [App.amKontakt] (#1960).
func (a *App) leihenNachDemSchreiben(ctx context.Context) DeckleiheAntwort {
	antwort := a.Deckleihen(ctx)
	antwort.Nachladefehler, antwort.Fehler = antwort.Fehler, nil

	return antwort
}

// LeihwunschDTO ist, was jemand verleihen will.
type LeihwunschDTO struct {
	// Entleiher ist die **Peerkennung** des Gegenübers.
	Entleiher string `json:"entleiher"`

	DeckID string `json:"deckID"`

	Bindungsart string `json:"bindungsart"`
	Woran       string `json:"woran"`

	Rechte []string `json:"rechte"`

	// Stunden ist die Dauer ab jetzt.
	Stunden int `json:"stunden"`

	Modus string `json:"modus"`
}

// wunschAus macht aus dem DTO den Wunsch.
//
// **Unbekannte Wörter werden durchgereicht und nicht ausgesiebt.** Was eine
// Bindungsart, ein Recht und ein Modus sein können, entscheidet
// [domain.Deckleihe.Pruefe] — eine zweite Prüfung hier wäre eine zweite
// Antwort auf dieselbe Frage, und die hier wäre die ungeprüfte (§2).
//
// **Die Dauer ist die Ausnahme** (J05, #2269): ob eine Zahl von Stunden in
// eine Dauer passt, ist keine Frage der Leihe, sondern der Umrechnung, und
// die geschieht hier. Umgelaufen wurden 5.124.096 Stunden zu 25 Minuten.
func wunschAus(wunsch LeihwunschDTO, deckID int64) (service.Leihwunsch, error) {
	dauer, passt := domain.Zeitspanne(int64(wunsch.Stunden), time.Hour)
	if !passt {
		return service.Leihwunsch{}, domain.NewFieldError("stunden", domain.SchluesselValidierung,
			"Eine Leihe dauert eine ganze Zahl von Stunden ab null, und höchstens so lange, wie eine Dauer fasst.")
	}

	rechte := make([]domain.Leihrecht, 0, len(wunsch.Rechte))
	for _, recht := range wunsch.Rechte {
		rechte = append(rechte, domain.Leihrecht(recht))
	}

	return service.Leihwunsch{
		Entleiher: wunsch.Entleiher,
		DeckID:    deckID,
		Bindung: domain.Leihbindung{
			Art: domain.Leihart(wunsch.Bindungsart), Woran: wunsch.Woran,
		},
		Rechte: rechte,
		Dauer:  dauer,
		Modus:  domain.Leihmodus(wunsch.Modus),
	}, nil
}

// leereLeihen gibt eine Antwort mit leerer Liste und übersetztem Fehler.
func leereLeihen(ctx context.Context, fehler error) DeckleiheAntwort {
	return DeckleiheAntwort{
		Leihen: []LeiheDTO{}, Fehler: Uebersetze(ctx, fehler),
	}
}

// leihenAus bildet die Stände ab.
func leihenAus(staende []domain.Leihstand) []LeiheDTO {
	uebersetzt := make([]LeiheDTO, 0, len(staende))

	for _, stand := range staende {
		rechte := make([]string, 0, len(stand.Leihe.Rechte))
		for _, recht := range stand.Leihe.Rechte {
			rechte = append(rechte, string(recht))
		}

		uebersetzt = append(uebersetzt, LeiheDTO{
			Kennung:     stand.Leihe.Kennung,
			Verleiher:   stand.Leihe.Verleiher,
			Entleiher:   stand.Leihe.Entleiher,
			Abdruck:     stand.Leihe.Abdruck,
			Bindungsart: string(stand.Leihe.Bindung.Art),
			Woran:       stand.Leihe.Bindung.Woran,
			Rechte:      rechte,
			Ausgestellt: stand.Leihe.Ausgestellt.Format(time.RFC3339),
			Bis:         stand.Leihe.Bis.Format(time.RFC3339),
			Modus:       string(stand.Leihe.Modus),
			Verliehen:   stand.Verliehen,
			Verbraucht:  stand.Verbraucht,
			Widerrufen:  stand.Widerrufen,
			Karten:      stand.Abzug.Zaehle(),
		})
	}

	return uebersetzt
}
