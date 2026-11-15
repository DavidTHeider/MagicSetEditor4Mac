package app_test

// **Zurücksetzen benennt genau, was betroffen ist** — Werkstück O4 von Phase
// 5.2 (#1428).

import (
	"context"
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// ruecksetzungsdoppel spielt den Dienst.
type ruecksetzungsdoppel struct {
	vorschau domain.Ruecksetzungsvorschau
	bilanz   service.Ruecksetzungsbilanz
	fehler   error

	gesehen  service.Ruecksetzungsziel
	art      domain.Ruecksetzungsart
	nachweis string
}

func (r *ruecksetzungsdoppel) Vorschau(
	_ context.Context, art domain.Ruecksetzungsart, ziel service.Ruecksetzungsziel,
) (domain.Ruecksetzungsvorschau, error) {
	r.art, r.gesehen = art, ziel

	return r.vorschau, r.fehler
}

func (r *ruecksetzungsdoppel) Zuruecksetzen(
	_ context.Context, art domain.Ruecksetzungsart, ziel service.Ruecksetzungsziel,
	nachweis, _ string,
) (service.Ruecksetzungsbilanz, error) {
	r.art, r.gesehen, r.nachweis = art, ziel, nachweis

	return r.bilanz, r.fehler
}

func ruecksetzungsapp(doppel *ruecksetzungsdoppel) *app.App {
	anwendung := app.NeueApp(t2Ort)
	anwendung.MitRuecksetzung(doppel)

	return anwendung
}

// t2Ort ist ein Bestandsort, den kein Test anfasst.
const t2Ort = "/nirgends/bestand.db"

// **Die sechs Arten kommen in der Reihe ihrer Schwere.**
func TestDieRuecksetzungsartenKommenInIhrerReihe(t *testing.T) {
	t.Parallel()

	arten := ruecksetzungsapp(&ruecksetzungsdoppel{}).Ruecksetzungsarten()
	if len(arten) != 6 || arten[0] != "layout" || arten[5] != "anwendung" {
		t.Fatalf("die Reihe ist %v", arten)
	}
}

// **Die Vorschau kommt vollständig durch** — samt gerechnetem `bereit`.
func TestDieVorschauKommtVollstaendigDurch(t *testing.T) {
	t.Parallel()

	doppel := &ruecksetzungsdoppel{vorschau: domain.Ruecksetzungsvorschau{
		Art: domain.RuecksetzungProfil,
		Betroffen: []domain.Betroffenes{
			{Was: domain.BetroffenSets, Anzahl: 3, Namen: []string{"Eins"}},
			// **Ohne Namen** — die leere Liste ist eine Aussage, `null` nicht.
			{Was: domain.BetroffenKarten, Anzahl: 107},
		},
		Hindernisse:        []domain.Hindernis{{Art: domain.RuecksetzungOffeneRueckkehr, Stelle: "Meins"}},
		SicherungEmpfohlen: true,
	}}

	antwort := ruecksetzungsapp(doppel).Ruecksetzungsvorschau(
		t.Context(), "profil", "12", "")
	if antwort.Fehler != nil {
		t.Fatalf("die Vorschau: %+v", antwort.Fehler)
	}

	if doppel.gesehen.ProfilID != 12 || doppel.art != domain.RuecksetzungProfil {
		t.Errorf("gefragt wurde nach %q und %+v", doppel.art, doppel.gesehen)
	}

	if antwort.Bereit {
		t.Error("mit Hindernis gilt die Vorschau als bereit")
	}

	if len(antwort.Betroffen) != 2 || antwort.Betroffen[1].Namen == nil {
		t.Errorf("die Gruppen sind %+v", antwort.Betroffen)
	}

	if !antwort.SicherungEmpfohlen || antwort.NachweisNoetig {
		t.Errorf("die Ansagen sind %+v", antwort)
	}

	if antwort.Hindernisse[0].Stelle != "Meins" {
		t.Errorf("die Hindernisse sind %+v", antwort.Hindernisse)
	}
}

// **Ein leeres Ziel heißt „meines"** — und wird nicht zur Kennung null.
func TestEinLeeresZielBleibtLeer(t *testing.T) {
	t.Parallel()

	doppel := &ruecksetzungsdoppel{}

	if antwort := ruecksetzungsapp(doppel).Ruecksetzungsvorschau(
		t.Context(), "layout", "", ""); antwort.Fehler != nil {
		t.Fatalf("die Vorschau: %+v", antwort.Fehler)
	}

	if doppel.gesehen.ProfilID != 0 || doppel.gesehen.NutzerID != 0 {
		t.Errorf("das Ziel ist %+v", doppel.gesehen)
	}
}

// **Eine unlesbare Kennung ist ein Fehler der Anfrage** — an beiden Feldern.
func TestEineUnlesbareKennungWirdBenannt(t *testing.T) {
	t.Parallel()

	for name, ruf := range map[string]func(*app.App) *app.Fehler{
		"das Profil": func(a *app.App) *app.Fehler {
			return a.Ruecksetzungsvorschau(context.Background(), "profil", "keine", "").Fehler
		},
		"der Nutzer": func(a *app.App) *app.Fehler {
			return a.Ruecksetzungsvorschau(context.Background(), "nutzer", "", "keiner").Fehler
		},
		"das Profil beim Ausführen": func(a *app.App) *app.Fehler {
			return a.Zuruecksetzen(context.Background(), "profil", "keine", "", "X", "", "").Fehler
		},
		"der Nutzer beim Ausführen": func(a *app.App) *app.Fehler {
			return a.Zuruecksetzen(context.Background(), "nutzer", "", "keiner", "X", "", "").Fehler
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			doppel := &ruecksetzungsdoppel{}

			befund := ruf(ruecksetzungsapp(doppel))
			if befund == nil {
				t.Fatal("die unlesbare Kennung kam durch")
			}

			if befund.Feld == "" {
				t.Errorf("der Fehler nennt kein Feld: %+v", befund)
			}

			if doppel.art != "" {
				t.Error("der Dienst wurde trotzdem gerufen")
			}
		})
	}
}

// **Name und Nachweis reisen mit.**
func TestNameUndNachweisReisenMit(t *testing.T) {
	t.Parallel()

	doppel := &ruecksetzungsdoppel{bilanz: service.Ruecksetzungsbilanz{
		Art: domain.RuecksetzungAnwendung, Beiseite: "/alt/bestand.db",
	}}

	antwort := ruecksetzungsapp(doppel).Zuruecksetzen(
		t.Context(), "anwendung", "", "", "Alles weg", "nachweis-1", "")
	if antwort.Fehler != nil {
		t.Fatalf("zurücksetzen: %+v", antwort.Fehler)
	}

	if doppel.gesehen.Name != "Alles weg" || doppel.nachweis != "nachweis-1" {
		t.Errorf("gesehen wurden %+v und %q", doppel.gesehen, doppel.nachweis)
	}

	if antwort.Beiseite != "/alt/bestand.db" {
		t.Errorf("der alte Bestand liegt laut Antwort bei %q", antwort.Beiseite)
	}
}

// **Die Bilanz überlebt einen Fehler** (§7, Regel 3): was zurückgesetzt ist,
// bleibt es — und der Ort des alten Bestands ist die Auskunft, ohne die
// niemand ihn wiederfände.
func TestDieBilanzUeberlebtEinenFehler(t *testing.T) {
	t.Parallel()

	doppel := &ruecksetzungsdoppel{
		bilanz: service.Ruecksetzungsbilanz{Zurueckgesetzt: 4, Beiseite: "/alt.db"},
		fehler: errors.New("der Bestand ließ sich nicht wieder öffnen"),
	}

	antwort := ruecksetzungsapp(doppel).Zuruecksetzen(
		t.Context(), "anwendung", "", "", "", "nachweis", "")
	if antwort.Fehler == nil {
		t.Fatal("der Fehler verschwand")
	}

	if antwort.Zurueckgesetzt != 4 || antwort.Beiseite != "/alt.db" {
		t.Errorf("die Bilanz ist %+v", antwort)
	}
}

// **Ein Fehler der Vorschau kommt als Fehler und nicht als leere Liste.**
func TestEinFehlerDerVorschauKommtDurch(t *testing.T) {
	t.Parallel()

	doppel := &ruecksetzungsdoppel{fehler: domain.ErrNotFound}

	antwort := ruecksetzungsapp(doppel).Ruecksetzungsvorschau(t.Context(), "profil", "", "")
	if antwort.Fehler == nil {
		t.Fatal("der Fehler verschwand")
	}

	if antwort.Betroffen == nil || antwort.Hindernisse == nil {
		t.Errorf("die Listen sind null: %+v", antwort)
	}
}
