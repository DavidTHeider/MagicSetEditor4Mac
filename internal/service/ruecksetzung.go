package service

// **Zurücksetzen benennt genau, was betroffen ist** — Werkstück O4 von Phase
// 5.2 (#1428).
//
// # Die Reihenfolge ist dieselbe wie bei der Datenpflege
//
// Vorschau (**rein lesend**) → entscheiden → ausführen. Der Unterschied: hier
// gibt es keine Revision, sondern **den Namen**. Wer ein Profil löschen will,
// tippt seinen Namen ab; das ist die Bindung, und sie steht seit #1351 in
// `storage.Nutzerablage.ProfilLoeschen`.
//
// **Und für die ganze Anwendung gibt es einen Nachweis** — eine Rolle reicht
// dafür nicht, denn sie nähme anderen Eigentümern ihre Daten weg.

import (
	"context"
	"fmt"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Ruecksetzungsablage liest, was betroffen wäre, und nimmt Wahlen zurück.
//
// **Die Schnittstelle steht hier, weil dieser Anwendungsfall sie braucht**
// (§2) — erfüllt von `internal/storage`.
type Ruecksetzungsablage interface {
	Profilumfang(ctx context.Context, profilID int64) ([]domain.Betroffenes, error)
	OffeneRueckkehrmengen(ctx context.Context, profilID int64) (int, error)
	Nutzerzahl(ctx context.Context) (int, error)
	Profile(ctx context.Context, nutzerID int64) ([]domain.Profil, error)
	AlleProfile(ctx context.Context) ([]domain.Profil, error)
	FremdeZugaenge(ctx context.Context, nutzerID int64) (domain.Betroffenes, error)
	EinstellungenZuruecksetzen(ctx context.Context, schluessel []string) (int, error)
	NutzereinstellungenZuruecksetzen(ctx context.Context, nutzerID int64) (int, error)
}

// Ruecksetzungsloescher nimmt Profile und Menschen heraus.
//
// **Die vorhandenen Wege und keine zweiten** (#1428: „vorhandene Löschdienste
// erweitern"): sie räumen das Beiwerk mit, drehen den Profilzeiger um und
// melden, was liegen blieb.
// **Die Bindung geht mit** — Befund 9 des Nachlaufs zum Zweitblick über Phase
// 5.2. Sie wird im Schreibgeschäft gerufen, mit dem Umfang je betroffenem
// Profil; wer sie ablehnt, lässt nichts geschehen. Was hier ankommt, ist
// gezählt, **nicht** zusammengerechnet: das Zusammenzählen ist eine fachliche
// Regel und steht in [zusammenzaehlen].
type Ruecksetzungsloescher interface {
	ProfilLoeschen(ctx context.Context, profilID int64, name string,
		bindung ...func(profile []domain.Profil, umfaenge [][]domain.Betroffenes) error) error
	NutzerLoeschen(ctx context.Context, nutzerID int64, name string,
		bindung ...func(profile []domain.Profil, umfaenge [][]domain.Betroffenes) error) error
}

// Neuanfangsweg legt den Bestand beiseite und öffnet einen frischen.
type Neuanfangsweg func(ctx context.Context) (string, error)

// Ruecksetzungsziel sagt, worauf sich eine Rücksetzung bezieht.
type Ruecksetzungsziel struct {
	ProfilID int64
	NutzerID int64

	// Name ist die Bestätigung durch Abtippen — leer, wo nichts gelöscht wird.
	Name string
}

// Ruecksetzung führt den Weg.
type Ruecksetzung struct {
	ablage    Ruecksetzungsablage
	loescher  Ruecksetzungsloescher
	wache     Ruecksetzungswache
	neuanfang Neuanfangsweg

	// auftraege ist wahlweise — siehe [Ruecksetzung.MitAuftraegen].
	auftraege Auftragsschau
}

// NeueRuecksetzung baut sie.
//
// **Ohne `neuanfang` gibt es die ganze Anwendung nicht** (G09) — die Vorschau
// nennt dann [domain.RuecksetzungOhneTausch].
//
// **Die Wache steht im Bauplan und nicht in einem `MitWache`.** Eine
// Rücksetzung ohne sie wüsste nicht, wem etwas gehört — und eine Sicherung,
// die man vergessen kann, ist keine.
func NeueRuecksetzung(
	ablage Ruecksetzungsablage, loescher Ruecksetzungsloescher,
	wache Ruecksetzungswache, neuanfang Neuanfangsweg,
) *Ruecksetzung {
	return &Ruecksetzung{
		ablage: ablage, loescher: loescher, wache: wache, neuanfang: neuanfang,
	}
}

// Vorschau sagt, was eine Rücksetzung anfassen würde — **und ändert nichts**.
func (r *Ruecksetzung) Vorschau(
	ctx context.Context, art domain.Ruecksetzungsart, ziel Ruecksetzungsziel,
) (domain.Ruecksetzungsvorschau, error) {
	vorschau, _, fehler := r.vorschauUndZiel(ctx, art, ziel)

	return vorschau, fehler
}

// vorschauUndZiel gibt die Vorschau **und das gebundene Ziel**.
//
// **Zwei Rückgaben und kein zweiter Ruf.** Die Ausführung braucht dasselbe
// Ziel, das die Vorschau gezählt hat; es ein zweites Mal zu binden hieße, es
// zwischen Nachsehen und Wirken noch einmal zu lesen — und den Fehlerweg
// dieses zweiten Rufs könnte kein Test betreten (§3).
func (r *Ruecksetzung) vorschauUndZiel(
	ctx context.Context, art domain.Ruecksetzungsart, ziel Ruecksetzungsziel,
) (domain.Ruecksetzungsvorschau, Ruecksetzungsziel, error) {
	if !domain.IstRuecksetzungsart(string(art)) {
		return domain.Ruecksetzungsvorschau{}, ziel, fmt.Errorf(
			"%w: %q ist keine Rücksetzung, die es gibt", domain.ErrValidation, art)
	}

	vorschau := domain.Ruecksetzungsvorschau{
		Art:                art,
		Betroffen:          []domain.Betroffenes{},
		Hindernisse:        []domain.Hindernis{},
		NameNoetig:         art.BrauchtNamen(),
		NachweisNoetig:     art.BrauchtNachweis(),
		SicherungEmpfohlen: art.LoeschtDaten(),
	}

	ziel, profile, fremd, wachfehler := r.gebundenesZiel(ctx, art, ziel)
	if wachfehler != nil {
		return domain.Ruecksetzungsvorschau{}, ziel, wachfehler
	}

	// **Ein fremdes Ziel wird nicht gezählt.** Wer es nicht anfassen darf,
	// soll auch nicht erfahren, wie viele Karten darin liegen.
	if len(fremd) > 0 {
		vorschau.Hindernisse = fremd
		vorschau.Revision = domain.Ruecksetzungsmarke(art, vorschau.Betroffen)

		return vorschau, ziel, nil
	}

	vorschau.Hindernisse = append(vorschau.Hindernisse, r.laufende(art)...)

	// **Ohne Tausch keine ganze Anwendung** (G09, #2273): eine Fläche, die
	// ihren Bestand über ihre ganze Laufzeit hält, gibt keinen Neuanfang mit
	// — danach hielte jede ihrer Nähte eine geschlossene Verbindung.
	if art == domain.RuecksetzungAnwendung && r.neuanfang == nil {
		vorschau.Hindernisse = append(vorschau.Hindernisse,
			domain.Hindernis{Art: domain.RuecksetzungOhneTausch})
	}

	switch art {
	case domain.RuecksetzungLayout:
		vorschau.Betroffen = einstellungsgruppe(len(domain.Layouteinstellungen()),
			domain.Layouteinstellungen())
	case domain.RuecksetzungProfileinstellungen,
		domain.RuecksetzungNutzereinstellungen:
		gewaehlt, fehler := r.einstellungsumfang(ctx, art, ziel)
		if fehler != nil {
			return domain.Ruecksetzungsvorschau{}, ziel, fehler
		}

		vorschau.Betroffen = gewaehlt
	case domain.RuecksetzungProfil, domain.RuecksetzungNutzer,
		domain.RuecksetzungAnwendung:
		if fehler := r.umfangUndHindernisse(ctx, art, ziel.NutzerID, profile, &vorschau); fehler != nil {
			return domain.Ruecksetzungsvorschau{}, ziel, fehler
		}
	}

	vorschau.Revision = domain.Ruecksetzungsmarke(art, vorschau.Betroffen)

	return vorschau, ziel, nil
}

// einstellungsumfang zählt die Wahlen, die zurückgenommen würden.
func (r *Ruecksetzung) einstellungsumfang(
	ctx context.Context, art domain.Ruecksetzungsart, ziel Ruecksetzungsziel,
) ([]domain.Betroffenes, error) {
	if art == domain.RuecksetzungNutzereinstellungen {
		// **Die Zahl kommt aus dem Umfang des Profils nicht** — sie gehört dem
		// Menschen. Sie hier zu zählen hieße, eine zweite Abfrage für eine
		// Zahl zu bauen, die die Rücksetzung selbst zurückgibt; die Vorschau
		// nennt deshalb die **Schlüssel**, nicht ihre Zahl.
		return einstellungsgruppe(0, nil), nil
	}

	umfang, fehler := r.ablage.Profilumfang(ctx, ziel.ProfilID)
	if fehler != nil {
		return nil, fehler
	}

	for _, gruppe := range umfang {
		if gruppe.Was == domain.BetroffenEinstellungen {
			return []domain.Betroffenes{gruppe}, nil
		}
	}

	return []domain.Betroffenes{}, nil
}

// einstellungsgruppe baut die Gruppe aus einer Zahl und ihren Namen.
func einstellungsgruppe(anzahl int, namen []string) []domain.Betroffenes {
	if namen == nil {
		namen = []string{}
	}

	return []domain.Betroffenes{{
		Was: domain.BetroffenEinstellungen, Anzahl: anzahl, Namen: namen,
	}}
}

// umfangUndHindernisse füllt die Vorschau für die drei löschenden Arten.
//
// **Ein fremdes Profil wird mitgezählt und nicht benannt** — das kommt nur
// beim Neuanfang vor, der alle Profile des Bestands sieht (Befund 1 des
// Zweitblicks über Phase G, #2282). Seine Mengen gehen in die Summe, die beim
// Neuanfang ohnehin kein Profil nennt; seine offene Rückkehr sperrt, **ohne**
// seinen Namen zu verraten, und einmal statt je Profil.
func (r *Ruecksetzung) umfangUndHindernisse(
	ctx context.Context, art domain.Ruecksetzungsart, nutzerID int64,
	profile []domain.Profil, vorschau *domain.Ruecksetzungsvorschau,
) error {
	umfaenge := make([][]domain.Betroffenes, 0, len(profile))
	fremdGesperrt := false

	for _, profil := range profile {
		// **Fremd ist ein Profil nur beim Neuanfang** — die anderen Arten
		// sehen nur, was dem Menschen gehört (`gebundenesZiel`).
		fremd := art == domain.RuecksetzungAnwendung && profil.NutzerID != nutzerID

		umfang, umfangfehler := r.ablage.Profilumfang(ctx, profil.ID)
		if umfangfehler != nil {
			return umfangfehler
		}

		// **Ein fremdes Profil wird gezählt, nicht benannt** — auch seine
		// Sets, Decks und Token nicht (neuer Befund des Nachlaufs über Phase
		// G): die Namen gingen sonst in die Summe und stünden in der Vorschau.
		if fremd {
			umfang = ohneNamen(umfang)
		}

		umfaenge = append(umfaenge, umfang)

		// **Eine offene Rückkehr sperrt** (Vermerk 0088): über diese Mengen
		// steht nichts fest, und ein gelöschtes Profil beantwortet die Frage
		// nicht mehr — es macht sie unbeantwortbar.
		offen, rueckkehrfehler := r.ablage.OffeneRueckkehrmengen(ctx, profil.ID)
		if rueckkehrfehler != nil {
			return rueckkehrfehler
		}

		switch {
		case offen == 0:
		case !fremd:
			vorschau.Hindernisse = append(vorschau.Hindernisse, domain.Hindernis{
				Art: domain.RuecksetzungOffeneRueckkehr, Stelle: profil.Name,
			})
		case !fremdGesperrt:
			fremdGesperrt = true
			vorschau.Hindernisse = append(vorschau.Hindernisse,
				domain.Hindernis{Art: domain.RuecksetzungOffeneRueckkehr})
		}
	}

	if art == domain.RuecksetzungNutzer {
		// **Mit seinen Zugängen an fremden Profilen** (#2287): sie gehen per
		// Kaskade mit ihm. Gezählt und nicht benannt — und als eigener Umfang,
		// so wie das Schreibgeschäft ihn bindet (`storage.NutzerLoeschen`).
		fremd, zugangsfehler := r.ablage.FremdeZugaenge(ctx, nutzerID)
		if zugangsfehler != nil {
			return zugangsfehler
		}

		if fremd.Anzahl > 0 {
			umfaenge = append(umfaenge, []domain.Betroffenes{fremd})
		}

		anzahl, zahlfehler := r.ablage.Nutzerzahl(ctx)
		if zahlfehler != nil {
			return zahlfehler
		}

		if anzahl <= 1 {
			vorschau.Hindernisse = append(vorschau.Hindernisse,
				domain.Hindernis{Art: domain.RuecksetzungLetzterNutzer})
		}
	}

	vorschau.Betroffen = gesamtumfang(art, profile, umfaenge)

	return nil
}

// ohneNamen gibt einen Umfang mit denselben Zahlen und ohne Namen.
func ohneNamen(umfang []domain.Betroffenes) []domain.Betroffenes {
	gezaehlt := make([]domain.Betroffenes, 0, len(umfang))
	for _, gruppe := range umfang {
		gezaehlt = append(gezaehlt, domain.Betroffenes{
			Was: gruppe.Was, Anzahl: gruppe.Anzahl, Namen: []string{},
		})
	}

	return gezaehlt
}

// gesamtumfang legt die Umfänge übereinander und nennt die Profile dazu.
//
// **Eine Funktion und nicht zwei** — Befund 9 des Nachlaufs zum Zweitblick
// über Phase 5.2. Die Vorschau baut den Umfang, gegen den die Marke gerechnet
// wird; das Schreibgeschäft baut ihn ein zweites Mal, um zu prüfen, ob er noch
// derselbe ist. Zwei Fassungen davon liefen auseinander, sobald eine Gruppe
// dazukäme — und die Bindung wiese dann **jede** Löschung ab.
//
// **Die Profile stehen nur dort, wo sie selbst verschwinden.** Beim Neuanfang
// bleibt kein Profil übrig, und eine Zahl, die immer alle nennt, sagt nichts.
func gesamtumfang(
	art domain.Ruecksetzungsart, profile []domain.Profil,
	umfaenge [][]domain.Betroffenes,
) []domain.Betroffenes {
	gesamt := []domain.Betroffenes{}
	for _, umfang := range umfaenge {
		gesamt = zusammenzaehlen(gesamt, umfang)
	}

	if art != domain.RuecksetzungAnwendung {
		gesamt = append(gesamt, domain.Betroffenes{
			Was: domain.BetroffenProfile, Anzahl: len(profile), Namen: namenVon(profile),
		})
	}

	return gesamt
}

// zusammenzaehlen legt zwei Umfänge übereinander.
func zusammenzaehlen(bisher, dazu []domain.Betroffenes) []domain.Betroffenes {
	nachWas := map[string]int{}
	for stelle, gruppe := range bisher {
		nachWas[gruppe.Was] = stelle
	}

	for _, gruppe := range dazu {
		stelle, kennt := nachWas[gruppe.Was]
		if !kennt {
			bisher = append(bisher, gruppe)
			nachWas[gruppe.Was] = len(bisher) - 1

			continue
		}

		bisher[stelle].Anzahl += gruppe.Anzahl
		bisher[stelle].Namen = gekuerzt(
			append(bisher[stelle].Namen, gruppe.Namen...))
	}

	return bisher
}

// Ruecksetzungsbeispiele ist, wie viele Namen eine Gruppe nennt.
//
// **Dieselbe Zahl wie im Bestand** — dort ist sie die Grenze der Abfrage, hier
// die des Zusammenlegens. Zwei verschiedene wären zwei Zusagen.
const Ruecksetzungsbeispiele = 5

func gekuerzt(namen []string) []string {
	if len(namen) <= Ruecksetzungsbeispiele {
		return namen
	}

	return namen[:Ruecksetzungsbeispiele]
}

func namenVon(profile []domain.Profil) []string {
	namen := make([]string, 0, len(profile))
	for _, profil := range profile {
		if profil.Name != "" {
			namen = append(namen, profil.Name)
		}
	}

	return gekuerzt(namen)
}
