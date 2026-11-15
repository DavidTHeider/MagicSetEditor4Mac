// **Nutzer und Profile verwalten** — Werkstück 1 von Phase 4.12
// (#1351),
// nach Vermerk [0066](../../docs/entscheidungen/0066-nutzer-profile-und-die-vier-entscheidungen.md).
//
// # Was hier entschieden wird
//
// **Ob ein Name taugt** — die Prüfung steht in `domain`, gerufen wird sie
// hier (§2). Der `CHECK` in der Wanderung ist die zweite Reihe: er greift, wo
// niemand durch diesen Dienst geht, und er nennt kein Feld.
//
// # Was hier **nicht** steht
//
// **Keine Anmeldung und kein zweiter Faktor.** Wer gerade arbeitet, ist
// Werkstück 4; was ein zweiter Faktor ist, Werkstück 5.

package service

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Nutzerablage ist die Naht zum Bestand.
type Nutzerablage interface {
	Nutzer(ctx context.Context) ([]domain.Nutzer, error)
	Profile(ctx context.Context, nutzerID int64) ([]domain.Profil, error)
	NutzerAnlegen(ctx context.Context, neu domain.Nutzer) (domain.Nutzer, error)
	ProfilAnlegen(ctx context.Context, neu domain.Profil) (domain.Profil, error)

	// **Ändern und Löschen** — Werkstück 5 (#1351).
	NutzerAendern(ctx context.Context, geaendert domain.Nutzer) (domain.Nutzer, error)
	ProfilAendern(ctx context.Context, geaendert domain.Profil) (domain.Profil, error)
	NutzerLoeschen(ctx context.Context, nutzerID int64, name string,
		bindung ...func(profile []domain.Profil, umfaenge [][]domain.Betroffenes) error) error
	// ProfilLoeschen löscht ein Profil — und stellt den Zeiger im selben Zug.
	//
	// **`ausweich` ist das Profil, auf das der Zeiger danach steht**, oder
	// null, wenn er ohnehin woanders steht. Beides gehört in **eine**
	// Transaktion (Befund 1 des Zweitblicks über die Phasen 4.0 bis 4.13):
	// vorher gesetzt blieb er bewegt, wenn das Löschen danach abgewiesen
	// wurde.
	ProfilLoeschen(ctx context.Context, profilID int64, name string, ausweich int64,
		bindung ...func(profile []domain.Profil, umfaenge [][]domain.Betroffenes) error) error
}

// Nutzerverwaltung legt Nutzer und Profile an und gibt sie aus.
type Nutzerverwaltung struct {
	ablage Nutzerablage

	// zeiger sagt, in welchem Profil gerade gearbeitet wird — der
	// Profilwechsel steht in `profilwechsel.go`.
	zeiger Profilzeiger

	// freigaben sagt, wer in welches fremde Profil darf — Werkstück 11
	// (#1351). **`nil` heißt: es wird nicht gefragt**; der Grund steht bei
	// [Nutzerverwaltung.erlaubnisPruefen].
	freigaben *Freigabenverwaltung

	// faktor prüft den zweiten Faktor vor einem Wechsel — Werkstück 7
	// (#1351). `nil` heißt: es wird nicht gefragt.
	faktor Faktorpruefer

	// bilder und geheimnisse räumen weg, was **neben** dem Bestand an einem
	// Nutzer hängt — Befund 2 des Zweitblicks über die Phasen 4.0 bis 4.13
	// (2026-09-07).
	//
	// **`nil` heißt: es wird nicht weggeräumt.** Wer die Verwaltung ohne sie
	// baut — ein Doppel, eine Fläche, die nur liest —, löscht die Zeile und
	// sonst nichts; das ist der Stand von vor dem Befund und keine
	// Verschlechterung.
	bilder      Nutzerbildablage
	geheimnisse Nutzergeheimnisse
}

// Nutzerbildablage nimmt weg, was als **Bild** neben dem Bestand liegt.
//
// **Eine Schnittstelle und nicht `*Bildablage`**: die Verwaltung braucht
// zwei Handlungen, und ein Lauf über das Löschen soll nicht ein
// Bilderverzeichnis mitbauen müssen (§2).
//
// # Zwei Namen, weil es zwei Bilder gibt
//
// **Das Abzeichen kam am 2026-09-07 dazu** — Nachlauf zum Zweitblick über die
// Phasen 4.0 bis 4.13, Befund 2. Die erste Fassung dieser Naht kannte nur das
// Nutzerfoto, und `profil-<id>.png` blieb liegen: **dieselbe Klasse, ein
// anderes Bild**. Der Befund lautete darauf zu Recht „nur teilweise
// geschlossen".
//
// **Beide liegen im selben Verzeichnis** (`nutzerbild.go`), und beide gehören
// zu jemandem: das Foto einem Menschen, das Abzeichen einem **Profil**. Wer
// einen Menschen löscht, löscht seine Profile mit — also auch deren
// Abzeichen.
//
// # Und zwei Hintergründe
//
// **Tisch- und Fensterbild** — Befund G4 des Zweitblicks über Phase 6.3
// (#1666). Sie liegen im selben Verzeichnis und gehören dem Menschen, nicht
// einem Profil (`tischbild.go`, `fensterbild.go`); hier fehlten sie, und
// beide überlebten das Löschen.
type Nutzerbildablage interface {
	NutzerbildWegnehmen(nutzerID int64) error
	TischbildWegnehmen(nutzerID int64) error
	FensterbildWegnehmen(nutzerID int64) error
	ProfilbadgeWegnehmen(profilID int64) error
}

// Nutzergeheimnisse nimmt das Geheimnis des zweiten Faktors weg.
//
// **Derselbe Zuschnitt wie oben**, und derselbe Grund: der Schlüsselbund ist
// eine Frage der Plattform, und ein Lauf über das Löschen fragt sie nicht.
type Nutzergeheimnisse interface {
	Nimm(ctx context.Context, nutzerID int64) error
}

// MitBeiwerk sagt der Verwaltung, wo Bild und Geheimnis liegen.
//
// **Gesetzt vom Composition Root**, wie die übrigen Nähte. Ohne sie löscht
// `NutzerLoeschen` die Zeile und lässt beides liegen — genau der Befund, den
// dieser Weg schließt.
func (v *Nutzerverwaltung) MitBeiwerk(bilder Nutzerbildablage, geheimnisse Nutzergeheimnisse) {
	v.bilder, v.geheimnisse = bilder, geheimnisse
}

// Faktorpruefer ist die Naht zum zweiten Faktor.
//
// **Eine Schnittstelle und nicht `*Zweiterfaktor`**, obwohl beides in diesem
// Paket liegt: die Verwaltung braucht **eine** Frage von ihm, und ein Lauf
// über den Wechsel soll nicht seine drei Nähte mitbauen müssen (§2).
type Faktorpruefer interface {
	// PruefeBeimWechsel und nicht `Pruefe`: der Nutzer darf den Faktor
	// **hier** abschalten und beim Anmelden nicht (Werkstück 8, #1351).
	PruefeBeimWechsel(ctx context.Context, nutzerID int64, nachweis string) error

	// VerlangtBeimWechsel sagt, ob ein Wechsel zu diesem Nutzer nach dem
	// Faktor fragte — **ohne zu fragen**: kein Sensor, kein Dialog. Daran
	// hängt, wer über einen anderen Menschen schreiben darf (ID-03, #1702).
	VerlangtBeimWechsel(ctx context.Context, nutzerID int64) (bool, error)
}

// NeueNutzerverwaltung baut sie über einer Ablage.
//
// **Eine Ablage und zwei Schnittstellen**: `storage.Nutzerablage` erfüllt
// beide, und ein zweites Argument wäre ein zweiter Weg, dieselbe Verbindung
// hereinzureichen.
func NeueNutzerverwaltung(ablage Nutzerablage) *Nutzerverwaltung {
	verwaltung := &Nutzerverwaltung{ablage: ablage}

	// **Wer den Zeiger nicht kann, bekommt keinen.** Ein Doppel in einem Test,
	// das nur die vier Anlagewege belegt, soll nicht daran scheitern — und
	// [Nutzerverwaltung.Wechseln] sagt es dann deutlich, statt auf einer Null
	// zu stolpern.
	if zeiger, kann := ablage.(Profilzeiger); kann {
		verwaltung.zeiger = zeiger
	}

	return verwaltung
}

// MitFaktor hängt den Prüfer des zweiten Faktors an.
//
// **Als eigener Ruf und nicht als Argument von [NeueNutzerverwaltung]**: der
// Prüfer braucht seinerseits eine Ablage, und beide über **einen** Aufruf zu
// verdrahten hieße, die Reihenfolge im Composition Root festzuschreiben.
func (v *Nutzerverwaltung) MitFaktor(pruefer Faktorpruefer) *Nutzerverwaltung {
	v.faktor = pruefer

	return v
}

// Nutzerzeile ist ein Nutzer samt seinen Profilen.
//
// **Beides zusammen und nicht zwei Rufe**: die Profilwahl zeigt sie
// nebeneinander, und wer sie einzeln holte, hätte für jeden Nutzer eine eigene
// Antwort — Stände, die auseinanderlaufen können (§7, Regel 2).
type Nutzerzeile struct {
	Nutzer  domain.Nutzer
	Profile []domain.Profil
}

// Uebersicht gibt alle Nutzer mit ihren Profilen.
func (v *Nutzerverwaltung) Uebersicht(ctx context.Context) ([]Nutzerzeile, error) {
	nutzer, err := v.ablage.Nutzer(ctx)
	if err != nil {
		return nil, err
	}

	zeilen := make([]Nutzerzeile, 0, len(nutzer))

	for _, einer := range nutzer {
		// **Ein Lesefehler reißt ab und wird nicht verschluckt.** Eine halbe
		// Übersicht sieht aus wie eine ganze: wer einen Nutzer ohne Profile
		// sieht, hält ihn für leer (§8).
		profile, lesefehler := v.ablage.Profile(ctx, einer.ID)
		if lesefehler != nil {
			return nil, lesefehler
		}

		zeilen = append(zeilen, Nutzerzeile{Nutzer: einer, Profile: profile})
	}

	return zeilen, nil
}

// NutzerAnlegen prüft und legt an — **samt dem ersten Profil**.
//
// # Warum die Antwort das Profil trägt
//
// **Weil der Speicher es in derselben Transaktion anlegt** und ein Nutzer
// ohne Arbeitsumgebung nicht vorkommt. Bis zum 2026-09-04 gab diese Antwort
// den Nutzer mit einer **leeren** Profilliste zurück — und die war nicht bloß
// unbequem, sondern falsch: sie behauptet, es gebe keines.
//
// Wer danach in das neue Profil wechseln wollte, musste die ganze Übersicht
// neu holen. Zwei Rufe, zwei Stände (§7, Regel 2).
func (v *Nutzerverwaltung) NutzerAnlegen(ctx context.Context, neu domain.Nutzer) (
	Nutzerzeile, error,
) {
	if err := neu.Validate(); err != nil {
		return Nutzerzeile{}, err
	}

	angelegt, fehler := v.ablage.NutzerAnlegen(ctx, neu)
	if fehler != nil {
		return Nutzerzeile{}, fehler
	}

	// **Ein Lesefehler reißt ab und wird nicht verschluckt** — dieselbe Regel
	// wie in [Nutzerverwaltung.Uebersicht]: eine leere Liste sähe aus wie eine
	// Auskunft und wäre eine Vermutung (§8).
	profile, lesefehler := v.ablage.Profile(ctx, angelegt.ID)
	if lesefehler != nil {
		return Nutzerzeile{}, lesefehler
	}

	return Nutzerzeile{Nutzer: angelegt, Profile: profile}, nil
}

// ProfilAnlegen prüft und legt an.
func (v *Nutzerverwaltung) ProfilAnlegen(ctx context.Context, neu domain.Profil) (
	domain.Profil, error,
) {
	if err := neu.Validate(); err != nil {
		return domain.Profil{}, err
	}

	return v.ablage.ProfilAnlegen(ctx, neu)
}
