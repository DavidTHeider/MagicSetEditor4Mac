// **Der Profilwechsel** — Werkstück 3 von Phase 4.12
// (#1351),
// nach Vermerk [0066](../../docs/entscheidungen/0066-nutzer-profile-und-die-vier-entscheidungen.md).
//
// # Was hier entschieden wird
//
// **Ob es das Profil gibt** — und wenn nicht, ist es ein Nein und kein
// stiller Zeiger ins Leere. Ein solcher Zeiger machte **jede** Abfrage leer,
// und der Bestand sähe aus, als wäre er weg (§8).
//
// # Was hier **nicht** steht
//
// **Kein Rechnen am zweiten Faktor.** Er ist seit Werkstück 7 da und kommt
// genau hier an: ein Wechsel, den eine Prüfung abweist, findet nicht statt.
// **Was** ein Faktor ist, entscheidet [Zweiterfaktor]; hier steht nur, dass
// gefragt wird, bevor der Zeiger wandert.
//
// **Und kein Aufräumen der Oberfläche.** Dass ein Wechsel Formular, Rückfrage
// und geladene Nebenlisten mitnimmt (§7, Regel 1), ist eine Aussage über das
// Fenster und nicht über den Bestand — sie steht dort, wo der Zustand liegt.

package service

import (
	"context"
	"fmt"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Profilzeiger ist die Naht zum gezeigten Profil.
//
// **Eine eigene Schnittstelle neben [Nutzerablage]**, weil sie eine andere
// Frage beantwortet: jene sagt, **wer** es gibt, diese, **wer gerade
// arbeitet**. Die Verwaltung verlangt beide — das ist der Ort, an dem sie
// zusammenkommen (§2).
type Profilzeiger interface {
	Profilzeiger(ctx context.Context) (int64, error)
	ProfilzeigerSetzen(ctx context.Context, profilID int64) error

	// Arbeitsprofil sagt, in welchem Profil **dieser Ruf** arbeitet — was der
	// Kontext trägt (ein Token, ein `--profil`), sonst der Zeiger.
	//
	// **Dieselbe Frage wie in [Arbeitszeiger]**, und dieselbe Ablage
	// beantwortet sie; sie steht hier zusätzlich, weil die Nutzerverwaltung
	// keinen [Arbeitszeiger] hält. Zwei Schnittstellen mit derselben Methode
	// sind in Go dieselbe Zusage — eine gemeinsame dritte wäre ein Name mehr
	// und keine Aussage mehr.
	Arbeitsprofil(ctx context.Context) (int64, error)

	// Nutzerzeiger sagt, **wer** arbeitet — Werkstück 17 von Phase 4.12
	// (#1400).
	//
	// **Neben dem Profilzeiger und nicht in ihm.** Bis dahin gab es die Frage
	// nicht: „wer arbeitet" wurde aus dem gezeigten Profil abgeleitet, und wer
	// in ein fremdes wechselte, wurde dessen Eigner. Was das kostete, steht in
	// [0074](../../docs/entscheidungen/0074-wer-arbeitet-steht-fest.md).
	Nutzerzeiger(ctx context.Context) (int64, error)
	NutzerzeigerSetzen(ctx context.Context, nutzerID int64) error

	// ZeigerSetzen stellt **beide** Zeiger in einem Zug um.
	//
	// **Weil ein Nutzerwechsel beide bewegt** — Befund 1 des Zweitblicks über
	// die Phasen 4.0 bis 4.13 (2026-09-07). Nacheinander gesetzt, konnte das
	// zweite Schreiben scheitern, nachdem das erste durch war: dann arbeitete
	// ein anderer Mensch, während das Profil des alten gezeigt wurde. Der
	// Ruf meldete einen Fehler, und der Bestand stand halb um.
	ZeigerSetzen(ctx context.Context, nutzerID, profilID int64) error
}

// Wechseln stellt das gezeigte Profil um und gibt es zurück.
//
// **Das Profil und nicht nur ein „gut"**: der Rufer schreibt es sich an seine
// Fläche. Wer es danach erst holte, hätte zwei Rufe und zwei Stände, zwischen
// denen sich etwas ändern kann (§7, Regel 2).
// # Der zweite Faktor steht **vor** dem Zeiger
//
// **Weil ein Wechsel, den eine Prüfung abweist, nicht stattfinden darf** —
// und weil ein Zeiger, der schon gewandert ist, sich nicht zurücknehmen
// lässt, ohne dass jemand dazwischen etwas sieht.
//
// **Gefragt wird der Nutzer, zu dem das Ziel gehört**, und nicht der, der
// gerade arbeitet: wer in ein fremdes Profil wechselt, weist sich als dessen
// Eigentümer aus. Wer in sein eigenes zurückwechselt, weist sich als er
// selbst aus — dieselbe Frage, dieselbe Antwort.
//
// **Ein leerer Nachweis heißt „über den Sensor"**; wer keinen Faktor hat,
// wird gar nicht gefragt.
//
// # Und die Freigabe ist der **Ersatz** für diesen Nachweis
//
// **Werkstück 11 (#1351) fragt nicht zusätzlich, sondern stattdessen.** Wer
// den Code des Eigners kennt, kommt hinein — das ist die Lage seit Werkstück
// 7. Wer ihn nicht kennt, kommt hinein, wenn der Eigner ihn hereingelassen
// hat. Die Reihenfolge steht in [Nutzerverwaltung.zutrittPruefen], und der
// Grund dafür steht dort auch.
func (v *Nutzerverwaltung) Wechseln(ctx context.Context, profilID int64, nachweis string) (
	domain.Profil, error,
) {
	profil, fehler := v.profilSuchen(ctx, profilID)
	if fehler != nil {
		return domain.Profil{}, fehler
	}

	// **Vor dem Zutritt gelesen**, denn der Zutritt ändert ihn nicht — und
	// danach stünde er in einem Bestand, dessen Zeiger sich gerade bewegt.
	werArbeitet, lesefehler := v.WerArbeitet(ctx)
	if lesefehler != nil {
		return domain.Profil{}, lesefehler
	}

	if zutritt := v.Zutritt(ctx, profil, nachweis); zutritt != nil {
		return domain.Profil{}, zutritt
	}

	// **Ein Wechsel auf das gezeigte Profil schreibt nicht.** Ein Zeiger, der
	// schon steht, noch einmal gesetzt, wäre ein Bestandsschreiben ohne
	// Wirkung — und in der Beobachtbarkeit ein Ereignis, das nichts bedeutet.
	gezeigt, lesefehler := v.zeiger.Profilzeiger(ctx)
	if lesefehler != nil {
		return domain.Profil{}, lesefehler
	}

	if gezeigt == profilID {
		return profil, nil
	}

	// **Wer arbeitet, wird festgeschrieben, bevor der Zeiger sich bewegt** —
	// Werkstück 17 von Phase 4.12.
	//
	// # Warum genau hier und nicht bei jedem Wechsel
	//
	// **Weil der Rückfall sonst driftet.** Ohne Eintrag gilt der Eigner des
	// gezeigten Profils — das ist das Verhalten von vorher und für einen
	// Bestand mit einem einzigen Menschen genau richtig. Bewegt sich der
	// Zeiger aber in ein **fremdes** Profil, hieße derselbe Rückfall im
	// nächsten Augenblick „ich bin der Fremde".
	//
	// Festgeschrieben wird deshalb in dem Augenblick, in dem die Antwort
	// sonst überschriebe würde — und mit dem Wert, der **jetzt** gilt.
	if ziel := profil.NutzerID; ziel != werArbeitet {
		if merkfehler := v.zeiger.NutzerzeigerSetzen(ctx, werArbeitet); merkfehler != nil {
			return domain.Profil{}, merkfehler
		}
	}

	if setzfehler := v.zeiger.ProfilzeigerSetzen(ctx, profilID); setzfehler != nil {
		return domain.Profil{}, setzfehler
	}

	return profil, nil
}

// faktorPruefen fragt den zweiten Faktor — wenn es einen Prüfer gibt.
//
// # Warum ein `nil` hier durchlässt
//
// **Weil eine Verwaltung ohne Prüfer keine Zusage bricht, sondern keine
// macht.** Ein Doppel in einem Lauf, der nur den Wechsel meint, soll nicht an
// einer Naht scheitern, die zu seinem Fall nichts beiträgt — und der
// Composition Root setzt sie in **beiden** Wurzeln (`app.go` und
// `nutzernaht.go`), sodass im Betrieb keine fehlt.
//
// **Was das kostet, gehört gesagt**: ein vergessener Prüfer wäre still. Der
// Flächenlauf `internal/flaechen/zweiterfaktor_test.go` fährt deshalb gegen
// die **gebauten** Programme und nicht gegen ein Doppel (§10).
func (v *Nutzerverwaltung) faktorPruefen(
	ctx context.Context, nutzerID int64, nachweis string,
) error {
	if v.faktor == nil {
		return nil
	}

	return v.faktor.PruefeBeimWechsel(ctx, nutzerID, nachweis)
}

// Gezeigt sagt, in welchem Profil gerade gearbeitet wird.
func (v *Nutzerverwaltung) Gezeigt(ctx context.Context) (domain.Profil, error) {
	profilID, fehler := v.zeiger.Profilzeiger(ctx)
	if fehler != nil {
		return domain.Profil{}, fehler
	}

	return v.profilSuchen(ctx, profilID)
}

// Arbeitsprofil gibt das Profil, in dem **dieser Ruf** arbeitet.
//
// # Warum das nicht [Nutzerverwaltung.Gezeigt] ist
//
// **Weil ein Ruf mit Token nicht im gezeigten Profil arbeitet** — Werkstück
// A14 von Phase 5.3
// (#1453).
// Der Zeiger ist die Wahl eines **Menschen** am Fenster; das Arbeitsprofil ist
// die Wahl **dieses Aufrufs** — ein Token, ein `--profil`, ein Wechsel.
//
// **Bis A14 gab es die Frage nur für den Akteur** ([Autorisierung.Akteur]).
// `lesen.profil` nannte den Zeiger und versprach dabei: „Jede andere Antwort
// dieses Servers … gilt für DIESES Profil und kein anderes." Mit Token stimmte
// das nicht — die Sets kamen aus dem Tokenprofil, der Name aus dem Zeiger. Ein
// Kopf, der mehr verspricht als sein Rumpf hält (§10).
//
// **Beide bleiben nebeneinander stehen**, und das ist Absicht: `Gezeigt`
// beantwortet, was der Bestand als gezeigt führt — daran hängen der
// Nutzerwechsel, die Einstellungsübernahme und die Frage, ob der Bestand nur
// lesend zu öffnen ist. Diese hier beantwortet, wohin **dieser** Ruf schreibt.
func (v *Nutzerverwaltung) Arbeitsprofil(ctx context.Context) (domain.Profil, error) {
	profilID, fehler := v.zeiger.Arbeitsprofil(ctx)
	if fehler != nil {
		return domain.Profil{}, fehler
	}

	return v.profilSuchen(ctx, profilID)
}

// profilSuchen holt ein Profil über die Übersicht.
//
// **Über die Übersicht und nicht über eine eigene Abfrage**: sie liest ohnehin
// jedes Profil, und eine zweite Naht dafür wäre eine zweite Stelle, an der die
// Klemme aufs Profil stehen muss.
func (v *Nutzerverwaltung) profilSuchen(ctx context.Context, profilID int64) (
	domain.Profil, error,
) {
	zeilen, fehler := v.Uebersicht(ctx)
	if fehler != nil {
		return domain.Profil{}, fehler
	}

	for _, zeile := range zeilen {
		for _, profil := range zeile.Profile {
			if profil.ID == profilID {
				return profil, nil
			}
		}
	}

	return domain.Profil{}, fmt.Errorf(
		"%w: das Profil %d führt der Bestand nicht", domain.ErrNotFound, profilID)
}

// MitFreigaben setzt die Naht, die sagt, wer wohin darf.
//
// **Als eigener Ruf und nicht als Argument von [NeueNutzerverwaltung]** —
// dieselbe Bauform und derselbe Grund wie bei [Nutzerverwaltung.MitFaktor].
func (v *Nutzerverwaltung) MitFreigaben(
	freigaben *Freigabenverwaltung,
) *Nutzerverwaltung {
	v.freigaben = freigaben

	return v
}
