package app

// **Ein Sitz und sein Profil** — Werkstück G8 von Phase 5.5
// (#1465).
//
// # Wogegen hier geprüft wird
//
//   - Ein Sitz **ohne** Profil erzeugt keinen Befund — und keinen Ruf an den
//     Bestand.
//   - Ein Sitz mit einem Profil, **das es gibt**, erzeugt auch keinen.
//   - Ein Sitz mit einem Profil, **das es nicht mehr gibt**, erzeugt einen —
//     und die Partie bleibt benutzbar.

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// profilprobe kennt die Profile eines Nutzers — oder klagt.
type profilprobe struct {
	profile []domain.Profil
	klagt   error

	gefragt int
}

func (p *profilprobe) ProfileDesBesitzersVon(context.Context, int64) (
	[]domain.Profil, error,
) {
	p.gefragt++

	if p.klagt != nil {
		return nil, p.klagt
	}

	return p.profile, nil
}

// arbeitsprobe sagt, wer arbeitet — und sonst nichts.
//
// **Ein eigenes Doppel und nicht `wechseldoppel`**: jenes steht in
// `package app_test`, und diese Läufe stehen **innen**, weil sie
// `profilbefunde` rufen — einen Weg, den die Fassade nicht nach außen gibt.
type arbeitsprobe struct{ klagt error }

func (a *arbeitsprobe) Wechseln(context.Context, int64, string) (
	domain.Profil, error,
) {
	return domain.Profil{}, nil
}

func (a *arbeitsprobe) Gezeigt(context.Context) (domain.Profil, error) {
	return domain.Profil{}, nil
}

func (a *arbeitsprobe) Arbeitsprofil(context.Context) (domain.Profil, error) {
	return domain.Profil{}, nil
}

func (a *arbeitsprobe) WerArbeitet(context.Context) (int64, error) {
	return 1, a.klagt
}

func (a *arbeitsprobe) Anmeldeprofil(context.Context, int64) (int64, error) { return 1, nil }

func (a *arbeitsprobe) NutzerWechseln(context.Context, int64, string) error {
	return nil
}

func (a *arbeitsprobe) RolleImGezeigten(context.Context) (domain.Rolle, error) {
	return domain.RolleVerwalter, nil
}

// eineProfilanwendung gibt eine Anwendung mit Kenner und Wechsler.
func eineProfilanwendung(t *testing.T, kenner *profilprobe) *App {
	t.Helper()

	return NeueApp("").MitProfilkenner(kenner)
}

func TestProfilbefunde(t *testing.T) {
	t.Parallel()

	faelle := map[string]struct {
		sitze   []SitzDTO
		profile []domain.Profil
		befunde int
		gefragt int
	}{
		"ohne Bindung wird nicht gefragt": {
			sitze: []SitzDTO{{Kennung: "anna"}, {Kennung: "bert"}},
		},
		"ein Profil, das es gibt": {
			sitze:   []SitzDTO{{Kennung: "anna", ProfilID: "2"}},
			profile: []domain.Profil{{ID: 2, Name: "Turnier"}},
			gefragt: 1,
		},
		"ein Profil, das es nicht mehr gibt": {
			sitze:   []SitzDTO{{Kennung: "anna", ProfilID: "9"}},
			profile: []domain.Profil{{ID: 2, Name: "Turnier"}},
			befunde: 1, gefragt: 1,
		},
		"zwei verschwundene": {
			sitze: []SitzDTO{
				{Kennung: "anna", ProfilID: "8"},
				{Kennung: "bert", ProfilID: "9"},
			},
			befunde: 2, gefragt: 1,
		},
	}

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			kenner := &profilprobe{profile: fall.profile}

			befunde := eineProfilanwendung(t, kenner).profilbefunde(
				riegelkontext(t), fall.sitze)

			switch {
			case befunde == nil:
				t.Fatal("die Liste ist nil statt leer")
			case len(befunde) != fall.befunde:
				t.Fatalf("erwartet %d Befunde, bekam %v",
					fall.befunde, befunde)
			case kenner.gefragt != fall.gefragt:
				t.Errorf("gefragt wurde %dmal", kenner.gefragt)
			}

			for _, befund := range befunde {
				if !strings.Contains(befund, "nicht mehr gibt") {
					t.Errorf("der Befund sagt nicht, was los ist: %q", befund)
				}
			}
		})
	}
}

// **Was sich nicht nachschlagen lässt, wird nicht behauptet.**
//
// **Aus „ich weiß es nicht" einen Befund zu machen hieße, eine
// Auffälligkeit zu behaupten, die niemand gesehen hat** (§10).
func TestEinKlagenderKennerBehauptetNichts(t *testing.T) {
	t.Parallel()

	kenner := &profilprobe{klagt: errors.New("der Bestand ist zu")}

	befunde := eineProfilanwendung(t, kenner).profilbefunde(
		riegelkontext(t), []SitzDTO{{Kennung: "anna", ProfilID: "9"}})

	if len(befunde) != 0 {
		t.Fatalf("erwartet keinen Befund, bekam %v", befunde)
	}
}

// **Und ohne eingesetzten Kenner ebenso.**
func TestOhneKennerGibtEsKeineBefunde(t *testing.T) {
	t.Parallel()

	befunde := NeueApp("").profilbefunde(
		t.Context(), []SitzDTO{{Kennung: "anna", ProfilID: "9"}})

	if befunde == nil || len(befunde) != 0 {
		t.Fatalf("erwartet eine leere Liste, bekam %v", befunde)
	}
}

// **Und wenn schon nicht feststeht, in welchem Profil der Ruf arbeitet, wird
// nichts behauptet** — und nicht einmal gefragt.
func TestOhneProfilImRufGibtEsKeineBefunde(t *testing.T) {
	t.Parallel()

	kenner := &profilprobe{}

	befunde := eineProfilanwendung(t, kenner).profilbefunde(
		t.Context(), []SitzDTO{{Kennung: "anna", ProfilID: "9"}})

	if len(befunde) != 0 || kenner.gefragt != 0 {
		t.Fatalf("erwartet keinen Befund und keine Frage, bekam %v nach %d",
			befunde, kenner.gefragt)
	}
}

// **Ein Sitzwunsch trägt sein Profil bis in die Partie** — und ein leeres
// Feld ist kein Fehler.
func TestEinSitzwunschTraegtSeinProfil(t *testing.T) {
	t.Parallel()

	db := bestandFuerDenRiegel(t)
	flaeche := PartieflaecheUeber(db)
	deckID := einDeckImBestand(t, db)

	antwort := flaeche.PartieAusWuenschen(riegelkontext(t), "magic", "anna",
		[]Sitzwunsch{
			{Kennung: "anna", Name: "Anna", DeckID: deckID, ProfilID: "1"},
			{Kennung: "bert", Name: "Bert", DeckID: deckID},
		})

	switch {
	case antwort.Fehler != nil:
		t.Fatalf("Fehler: %+v", antwort.Fehler)
	case antwort.Partie.Sitze[0].ProfilID != "1":
		t.Errorf("das Profil ist %q", antwort.Partie.Sitze[0].ProfilID)
	case antwort.Partie.Sitze[1].ProfilID != "":
		t.Errorf("der ungebundene Sitz trägt %q",
			antwort.Partie.Sitze[1].ProfilID)
	}
}

// **Eine krumme Profilkennung wird gemeldet.**
func TestEineKrummeProfilkennungWirdGemeldet(t *testing.T) {
	t.Parallel()

	db := bestandFuerDenRiegel(t)
	flaeche := PartieflaecheUeber(db)
	deckID := einDeckImBestand(t, db)

	antwort := flaeche.PartieAusWuenschen(riegelkontext(t), "magic", "anna",
		[]Sitzwunsch{
			{
				Kennung: "anna", Name: "Anna", DeckID: deckID,
				ProfilID: "keine Zahl",
			},
			{Kennung: "bert", Name: "Bert", DeckID: deckID},
		})

	if antwort.Fehler == nil {
		t.Fatal("die krumme Kennung kam durch")
	}
}

// **Wer eine Partie anfängt, sitzt in seinem gezeigten Profil.**
//
// # Die zweite Zusage des Issues, und woran sie zu sehen ist
//
// „Das Fenster arbeitet in einem Profil (`GezeigtesProfil`); wer wechselt,
// bekommt die Sicht **dieses** Sitzes."
//
// **Der Wirt bleibt der erste Sitz** — er hält die Partie. Gezeigt wird
// trotzdem der zweite, weil das gezeigte Profil dort sitzt; ohne diese
// Trennung wäre der Lauf grün, sobald beide dieselbe Kennung tragen (§10).
//
// **Und ein Sitz ohne Profil bleibt erreichbar**: fällt die Bindung weg,
// steht wieder der Wirt da — das ist der zweite Fall darunter.
func TestDieSichtFolgtDemGezeigtenProfil(t *testing.T) {
	t.Parallel()

	faelle := map[string]struct {
		profilID string
		erwartet string
	}{
		"der zweite Sitz gehört dem gezeigten Profil": {
			profilID: "1", erwartet: "sitz-2",
		},
		"ohne Bindung bleibt es beim Wirt": {erwartet: "sitz-1"},
		"ein fremdes Profil sitzt hier nicht": {
			profilID: "7", erwartet: "sitz-1",
		},
	}

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			gebunden, deckID := eineSpielende(t)

			entwurf := einEntwurf(deckID)
			entwurf.Sitze[1].ProfilID = fall.profilID

			angelegt := gebunden.PartieAnlegen(t.Context(), entwurf)
			if angelegt.Fehler != nil {
				t.Fatalf("PartieAnlegen: %+v", angelegt.Fehler)
			}

			if angelegt.Partie.Wirt != "sitz-1" {
				t.Errorf("der Wirt ist %q", angelegt.Partie.Wirt)
			}

			if angelegt.Partie.Fuer != fall.erwartet {
				t.Errorf("gezeigt wird %q, erwartet %q",
					angelegt.Partie.Fuer, fall.erwartet)
			}
		})
	}
}

// **Und die Fläche ohne Fenster setzt denselben Menschen auf denselben Sitz.**
//
// # Warum dieser Lauf neben dem darüber steht
//
// **Weil das Fenster sein Profil woanders hält als die drei anderen.**
// [App.profil] ist ein Feld, das nur [App.Starten] füllt; Kommandozeile, MCP
// und REST stellen ihres in den **Wurzelkontext**
// ([storage.MitGezeigtemProfil]). Wer die Regel am Feld festmacht, hat sie an
// **einer** Fläche — und §16 sagt: dieselbe Fähigkeit, überall dieselbe
// Antwort.
//
// **Der zweite Sitz und nicht der erste**: der erste ist hier der Wirt, und
// dort wäre der Lauf grün, gleich woher die Kennung kommt (§10).
func TestAuchOhneFensterFolgtDieSichtDemProfil(t *testing.T) {
	t.Parallel()

	db := bestandFuerDenRiegel(t)
	deckID := einDeckImBestand(t, db)

	antwort := PartieflaecheUeber(db).PartieAusWuenschen(
		riegelkontext(t), "magic", "anna", []Sitzwunsch{
			{Kennung: "anna", Name: "Anna", DeckID: deckID},
			{Kennung: "bert", Name: "Bert", DeckID: deckID, ProfilID: "1"},
		})

	switch {
	case antwort.Fehler != nil:
		t.Fatalf("PartieAusWuenschen: %+v", antwort.Fehler)
	case antwort.Partie.Wirt != "anna":
		t.Errorf("der Wirt ist %q", antwort.Partie.Wirt)
	case antwort.Partie.Fuer != "bert":
		t.Errorf("gezeigt wird %q, erwartet \"bert\"", antwort.Partie.Fuer)
	}
}

// **Ein Ruf ohne Profil nennt niemanden — und sagt das.**
//
// **Der Fehlerweg von [App.PartieAnlegen]**: ohne Profil im Kontext lässt sich
// nicht sagen, wessen Sicht gemeint ist. Er steht hier, weil ihn im Betrieb
// keine der vier Flächen betritt — jede stattet ihren Kontext aus, bevor sie
// ruft (§3).
func TestEinePartieOhneProfilWirdGemeldet(t *testing.T) {
	t.Parallel()

	anwendung := NeueApp("")
	anwendung.MitProfilwechsel(&arbeitsprobe{})

	antwort := anwendung.PartieAnlegen(t.Context(), domain.Partie{})

	if antwort.Fehler == nil {
		t.Fatal("ein Ruf ohne Profil kam durch")
	}
}
