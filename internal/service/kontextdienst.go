package service

// **Das Kontextpaket** — Werkstück A16 von Phase 5.3
// (#1515).
//
// # Was der Dienst tut und was er nicht tut
//
// **Er wählt aus, entdoppelt, ordnet und kürzt** — nach einer versionierten
// Regel, und was dabei herauskommt, trägt Beleg-IDs, die **er** vergeben hat.
//
// **Er nimmt kein ganzes Set auf.** Das Issue sagt es scharf: „Rohes gesamtes
// Set/Deck/Profil wird nicht automatisch aufgenommen." Was hineinkommt, ist,
// was eine Recherche gefunden hat — und was davon ins Budget passt.
//
// **Und er ruft kein Modell.** Das Paket zusammenzustellen ist ein Lesevorgang;
// die Generierung ist der zweite, ausdrücklich ausgelöste Schritt (A6).
//
// # Warum das Budget am Zielmodell hängt
//
// **Weil es kein gemeinsames Maß gibt.** Was ein Modell als Token zählt,
// zählt das nächste anders; die Zählweise steht deshalb im Manifest, und ein
// anderes Ziel bekommt ein **neues** Paket ([domain.Kontextmanifest.PasstZu]).

import (
	"context"
	"strings"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Kontextwunsch ist der Auftrag an den Kontextdienst.
type Kontextwunsch struct {
	// Aufgabe ist, was ein Mensch getippt hat — sie wird **nicht** als Frage
	// an den Index verwendet, dafür steht [Kontextwunsch.Anfrage].
	Aufgabe string

	// Zielrevision bindet das Paket an den Stand der Zielkarte.
	Zielrevision string

	// Anfrage ist die Recherche, aus der die Belege kommen.
	Anfrage domain.Rechercheanfrage

	// Zielanbieter, Zielmodell und Laufzeit sind die Bindung, für die
	// budgetiert wird.
	Zielanbieter string
	Zielmodell   string
	Laufzeit     string

	// Kontextgrenze ist das Fenster des Zielmodells, Ausgabereserve der
	// Teil davon, der für die Antwort freibleibt.
	Kontextgrenze  int
	Ausgabereserve int

	// HoechstensBelege begrenzt die Zahl der Ausschnitte — null heißt: so
	// viele, wie ins Budget passen.
	HoechstensBelege int
}

// Kontextpaket ist, was ein Modell bekommt.
type Kontextpaket struct {
	Manifest domain.Kontextmanifest

	// Kontext ist der Text, der dem Modell mitgegeben wird — mit den
	// Belegkennungen davor.
	Kontext string

	// Ergebnis ist die Recherche, aus der die Belege stammen.
	Ergebnis domain.Rechercheergebnis
}

// Kontextdienst stellt Pakete zusammen.
type Kontextdienst struct {
	recherche *Recherchedienst
	profil    Profilquelle
	jetzt     func() time.Time

	// kennung vergibt die Paketkennungen — eingesetzt, damit ein Lauf sie
	// vorhersagen kann.
	kennung func() string
}

// NeuerKontextdienst baut ihn.
func NeuerKontextdienst(
	recherche *Recherchedienst, profil Profilquelle,
) *Kontextdienst {
	return &Kontextdienst{
		recherche: recherche, profil: profil, jetzt: time.Now,
		kennung: gewuerfelteKennung,
	}
}

// MitUhr setzt die Uhr.
func (d *Kontextdienst) MitUhr(jetzt func() time.Time) *Kontextdienst {
	d.jetzt = jetzt

	return d
}

// MitKennung setzt, wie eine Paketkennung entsteht.
func (d *Kontextdienst) MitKennung(kennung func() string) *Kontextdienst {
	d.kennung = kennung

	return d
}

// Zusammenstellen baut ein Paket.
//
// **Es ändert keine Karte** (Issue). Was hier geschieht, ist eine Suche und
// eine Auswahl; geschrieben wird nirgends.
func (d *Kontextdienst) Zusammenstellen(
	ctx context.Context, wunsch Kontextwunsch,
) (Kontextpaket, error) {
	if fehler := pruefeKontextwunsch(wunsch); fehler != nil {
		return Kontextpaket{}, fehler
	}

	profil, fehler := d.profil.Profil(ctx)
	if fehler != nil {
		return Kontextpaket{}, fehler
	}

	ergebnis, fehler := d.recherche.Recherchieren(ctx, wunsch.Anfrage, "")
	if fehler != nil {
		return Kontextpaket{}, fehler
	}

	jetzt := d.jetzt()

	manifest := domain.Kontextmanifest{
		PaketID:      d.kennung(),
		Aufgabe:      wunsch.Aufgabe,
		Zielrevision: wunsch.Zielrevision,
		Zielanbieter: wunsch.Zielanbieter,
		Zielmodell:   wunsch.Zielmodell,
		Laufzeit:     wunsch.Laufzeit,
		Zaehlweise:   Zaehlweise,

		Kontextgrenze:  wunsch.Kontextgrenze,
		Ausgabereserve: wunsch.Ausgabereserve,

		Indexfingerabdruck: ergebnis.Indexfingerabdruck,
		Raumabdruck:        ergebnis.Raumabdruck,
		Suchkonfiguration:  ergebnis.Suchkonfiguration,
		Korpusstaende:      ergebnis.Korpusstaende,
		Kontextpolicy:      domain.Kontextpolicyfassung,

		ProfilID: profil,
		Erstellt: jetzt,
		Laeuftab: jetzt.Add(domain.Kontextfrist),
	}

	manifest.Belege, manifest.Verbraucht = belegeAus(
		ergebnis.Treffer, wunsch, manifest.Verfuegbar())

	if fehler := manifest.Pruefe(); fehler != nil {
		return Kontextpaket{}, fehler
	}

	return Kontextpaket{
		Manifest: manifest, Kontext: kontexttext(manifest.Belege),
		Ergebnis: ergebnis,
	}, nil
}

// pruefeKontextwunsch weist ab, was kein Auftrag ist.
func pruefeKontextwunsch(wunsch Kontextwunsch) error {
	if strings.TrimSpace(wunsch.Aufgabe) == "" {
		return domain.MitGrund(domain.ErrValidation, nil,
			"ein Kontextpaket braucht seine Aufgabe — sie ist das, was ein "+
				"Mensch beantwortet haben will")
	}

	if strings.TrimSpace(wunsch.Zielmodell) == "" ||
		strings.TrimSpace(wunsch.Zielanbieter) == "" {
		return domain.MitGrund(domain.ErrValidation, nil,
			"ein Kontextpaket wird für ein bestimmtes Modell budgetiert — "+
				"ohne Anbieter und Modell gibt es keine Zählweise")
	}

	if wunsch.Kontextgrenze <= 0 ||
		wunsch.Ausgabereserve >= wunsch.Kontextgrenze {
		return domain.MitGrund(domain.ErrValidation, nil,
			"das Kontextfenster ist %d und die Ausgabereserve %d — dann "+
				"bleibt für den Kontext nichts",
			wunsch.Kontextgrenze, wunsch.Ausgabereserve)
	}

	return nil
}
