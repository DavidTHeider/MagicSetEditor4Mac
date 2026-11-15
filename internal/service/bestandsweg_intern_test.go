// **Die Vorrichtung für den Sicherungsweg**
// (#1175).
//
// # Warum sie keine Datenbank aufmacht
//
// Bis zum 2026-08-28 fuhren diese Tests durch `app.App`: `NeueApp`, `Starten`,
// `BestandSichern`. Das ging, weil der ganze Weg dort lag — und es band jeden
// Fall an eine echte SQLite-Datei, obwohl fast keiner sie brauchte.
//
// **Der Dienst nimmt den Lebenszyklus als Naht entgegen** (§2), und damit ist
// er hier eine Attrappe, die mitzählt. Was ein Test dadurch **gewinnt**: er
// sieht, ob geschlossen und wieder geöffnet wurde — vorher war das im Erfolg
// eines Aufrufs versteckt. Was er **verliert**, steht daneben: dass die
// Verbindung wirklich zumacht, prüft `internal/app` weiter an einer echten
// Datei (`sicherungrueckwege_intern_test.go`).

package service

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// zaehlenderLebenszyklus zählt Schließen und Öffnen mit.
type zaehlenderLebenszyklus struct {
	geschlossen int
	geoeffnet   int

	// schliessfehler und oeffnungsfehler stellen die Fehlerwege her.
	schliessfehler  error
	oeffnungsfehler error

	// beimSchliessen läuft, bevor das Schließen antwortet.
	//
	// **Der einzige Haken zwischen Auspacken und Tausch** — Befund 7 des
	// Nachlaufs zum Zweitblick über Phase 5.2 brauchte ihn: der Fehlerweg
	// „der Tausch scheitert" ist von außen nicht herzustellen, weil alles,
	// was das Verzeichnis unbrauchbar macht, schon das Auspacken trifft.
	beimSchliessen func()
}

// dateisicherer schreibt eine Datei mit festem Inhalt als „Bestand".
type dateisicherer struct {
	inhalt string
	fehler error

	// ziele sammelt, wohin gesichert wurde.
	ziele []string
}

func (d *dateisicherer) Sichern(_ context.Context, ziel string) error {
	d.ziele = append(d.ziele, ziel)

	if d.fehler != nil {
		return d.fehler
	}

	return os.WriteFile(ziel, []byte(d.inhalt), 0o600)
}

// merkbuch ist ein Rückkehrbuch im Arbeitsspeicher — Werkstück O7 von Phase
// 5.2 (#1431).
//
// **Jeder Weg, der zurückspielt, braucht eines.** Ohne Buch weist
// `Zurueckspielen` ab, und das ist Absicht: eine Rückkehr ohne Nachweis ist
// die eine, über die später niemand etwas sagen kann.
type merkbuch struct {
	epochen []domain.Wiederherstellungsepoche
	fehler  error
}

func (m *merkbuch) Vermerken(
	_ context.Context, epoche domain.Wiederherstellungsepoche,
) (int64, error) {
	if m.fehler != nil {
		return 0, m.fehler
	}

	epoche.ID = int64(len(m.epochen) + 1)
	m.epochen = append(m.epochen, epoche)

	return epoche.ID, nil
}

func (m *merkbuch) Offene(context.Context) ([]domain.Wiederherstellungsepoche, error) {
	if m.fehler != nil {
		return nil, m.fehler
	}

	offene := []domain.Wiederherstellungsepoche{}

	for _, epoche := range m.epochen {
		if epoche.Sperrt() {
			offene = append(offene, epoche)
		}
	}

	return offene, nil
}

func (m *merkbuch) Alle(context.Context) ([]domain.Wiederherstellungsepoche, error) {
	return m.epochen, m.fehler
}

func (m *merkbuch) Quittieren(_ context.Context, kennung int64, wann time.Time) error {
	for stelle := range m.epochen {
		if m.epochen[stelle].ID == kennung {
			m.epochen[stelle].Klaerung = domain.KlaerungQuittiert
			m.epochen[stelle].GeklaertAm = wann

			return nil
		}
	}

	return domain.ErrNotFound
}

// wegbau ist der Aufbau eines Bestandswegs für einen Test.
type wegbau struct {
	weg    *Bestandsweg
	ort    string
	bilder string
	leben  *zaehlenderLebenszyklus
	quelle *dateisicherer
	buch   *merkbuch
}

// neuerWegbau stellt einen Bestandsweg über einem frischen Verzeichnis her.
//
// **Der Bestand ist eine Datei und keine Datenbank.** Was dieser Weg mit ihr
// tut, ist Umbenennen und Kopieren; was darin steht, entscheidet der Prüfer,
// und der ist eine Naht.
func neuerWegbau(t *testing.T) *wegbau {
	t.Helper()

	ort := t.TempDir()
	bestand := filepath.Join(ort, "bestand.db")

	if err := os.WriteFile(bestand, []byte("laufender bestand"), 0o600); err != nil {
		t.Fatalf("bestand anlegen: %v", err)
	}

	leben := &zaehlenderLebenszyklus{}
	quelle := &dateisicherer{inhalt: "gesicherter bestand"}
	bilder := filepath.Join(ort, "bilder")

	weg := NeuerBestandsweg(
		Sicherungsorte{
			Datenbank: bestand,
			Bilder:    func() (string, error) { return bilder, nil },
		},
		gepruefteSicherung(),
		func(context.Context, string) (func() error, error) {
			return func() error { return nil }, nil
		},
		quelle,
		func() error {
			leben.geschlossen++

			if leben.beimSchliessen != nil {
				leben.beimSchliessen()
			}

			return leben.schliessfehler
		},
		func(context.Context) error { leben.geoeffnet++; return leben.oeffnungsfehler },
		Pruefassung,
	)

	buch := &merkbuch{}
	weg.MitRuecksprungschau(nil, nil, buch)

	return &wegbau{
		weg: weg, ort: ort, bilder: bilder, leben: leben, quelle: quelle, buch: buch,
	}
}

// gepruefteSicherung winkt jede Sicherung durch.
func gepruefteSicherung() Sicherungspruefer {
	return func(context.Context, string) (Sicherungsbefund, error) {
		return Sicherungsbefund{AngelegtAm: "2026-08-23T00:00:00.000Z"}, nil
	}
}

// Pruefassung ist die Fassung, mit der die Tests dieses Pakets rechnen.
//
// **Sie steht hier und nicht als `app.Fassung`**: der Dienst bekommt die
// Fassung hineingereicht (§2), und ein Test, der die echte nähme, würde bei
// jedem Release rot — ohne dass sich am geprüften Verhalten etwas ändert.
const Pruefassung = "0.3.0"

// wegAuf baut einen Bestandsweg über einer **genannten** Bestandsdatei.
//
// Der Unterschied zu [neuerWegbau]: dort entsteht das Verzeichnis, hier ist
// es schon da — die Marken- und Auspacktests brauchen einen Ort, den sie
// vorher selbst gefüllt haben.
func wegAuf(t *testing.T, bestand string) *Bestandsweg {
	t.Helper()

	leben := &zaehlenderLebenszyklus{}

	weg := NeuerBestandsweg(
		Sicherungsorte{
			Datenbank: bestand,
			Bilder: func() (string, error) {
				return filepath.Join(filepath.Dir(bestand), "bilder"), nil
			},
		},
		gepruefteSicherung(),
		func(context.Context, string) (func() error, error) {
			return func() error { return nil }, nil
		},
		&dateisicherer{inhalt: "gesicherter bestand"},
		func() error { leben.geschlossen++; return nil },
		func(context.Context) error { leben.geoeffnet++; return nil },
		Pruefassung,
	)

	weg.MitRuecksprungschau(nil, nil, &merkbuch{})

	return weg
}
