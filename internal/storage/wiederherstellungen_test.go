package storage_test

// **Die Epoche einer Rückkehr im Bestand** — Werkstück O7 von Phase 5.2
// (#1431).

import (
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

func epoche() domain.Wiederherstellungsepoche {
	return domain.Wiederherstellungsepoche{
		Zurueckgespielt: time.Date(2026, 9, 10, 12, 0, 0, 0, time.UTC),
		Datenstand:      time.Date(2026, 9, 1, 8, 30, 0, 0, time.UTC),
		Quelle:          "bestand-2026-09-01.msesicherung",
		Klaerung:        domain.KlaerungOffen,
		Mengen: []domain.Sammlungsmenge{
			{Profil: 1, Schluessel: "scryfall:abc", Zustand: "nm", Sprache: "en", Menge: 3, Name: "Lightning Bolt"},
			{Profil: 1, Schluessel: "scryfall:abc", Zustand: "nm", Sprache: "en", Finish: domain.FinishFoil, Menge: 1, Name: "Lightning Bolt"},
		},
	}
}

// **Geschrieben und wiedergelesen — Zeitpunkt für Zeitpunkt.** Die Abnahme
// verlangt genau den wiederhergestellten Datenzeitpunkt; wer ihn beim Lesen
// verliert, hat die Zusage nicht.
func TestEineEpocheKommtSoZurueckWieSieHineinging(t *testing.T) {
	db := frischeDatenbank(t)
	buch := storage.NeuesWiederherstellungsbuch(db)

	kennung, fehler := buch.Vermerken(probekontext(t), epoche())
	if fehler != nil {
		t.Fatalf("die Epoche vermerken: %v", fehler)
	}

	if kennung == 0 {
		t.Fatal("die Epoche hat keine Kennung bekommen")
	}

	offen, lesefehler := buch.Offene(probekontext(t))
	if lesefehler != nil {
		t.Fatalf("die offenen Epochen lesen: %v", lesefehler)
	}

	if len(offen) != 1 {
		t.Fatalf("es stehen %d offene Epochen da, erwartet war eine", len(offen))
	}

	gelesen := offen[0]

	if !gelesen.Datenstand.Equal(epoche().Datenstand) {
		t.Errorf("der Datenstand ist %v, erwartet war %v", gelesen.Datenstand, epoche().Datenstand)
	}

	if !gelesen.Zurueckgespielt.Equal(epoche().Zurueckgespielt) {
		t.Errorf("die Rückkehr steht auf %v, erwartet war %v",
			gelesen.Zurueckgespielt, epoche().Zurueckgespielt)
	}

	if gelesen.Quelle != epoche().Quelle {
		t.Errorf("die Quelle ist %q", gelesen.Quelle)
	}

	if len(gelesen.Mengen) != 2 {
		t.Fatalf("es stehen %d Mengen da, erwartet waren zwei: %+v", len(gelesen.Mengen), gelesen.Mengen)
	}

	// **Foil und nicht-Foil bleiben zwei Zeilen.** Kämen sie zusammen zurück,
	// gäbe die Sperre eine Karte frei, die es so nicht gibt.
	sperre := domain.NeueHandelssperre(gelesen.Mengen)
	if summe := sperre.Menge(); summe != 4 {
		t.Errorf("die Sperre betrifft %d Stück, erwartet waren 4", summe)
	}
}

// **Eine quittierte Epoche steht nicht mehr unter den offenen — und ist
// trotzdem noch da.** Das Issue verlangt dauerhafte Nachweise; eine gelöschte
// Epoche wäre der Nachweis, den es nicht mehr gibt.
func TestEineQuittierteEpocheVerlaesstDieOffenenUndBleibtImBuch(t *testing.T) {
	db := frischeDatenbank(t)
	buch := storage.NeuesWiederherstellungsbuch(db)

	kennung, fehler := buch.Vermerken(probekontext(t), epoche())
	if fehler != nil {
		t.Fatalf("die Epoche vermerken: %v", fehler)
	}

	quittiert := time.Date(2026, 9, 10, 13, 0, 0, 0, time.UTC)
	if quittierfehler := buch.Quittieren(probekontext(t), kennung, quittiert); quittierfehler != nil {
		t.Fatalf("quittieren: %v", quittierfehler)
	}

	offen, lesefehler := buch.Offene(probekontext(t))
	if lesefehler != nil {
		t.Fatalf("die offenen Epochen lesen: %v", lesefehler)
	}

	if len(offen) != 0 {
		t.Errorf("nach der Quittung stehen %d offene Epochen da", len(offen))
	}

	alle, allefehler := buch.Alle(probekontext(t))
	if allefehler != nil {
		t.Fatalf("alle Epochen lesen: %v", allefehler)
	}

	if len(alle) != 1 {
		t.Fatalf("im Buch stehen %d Epochen, erwartet war eine", len(alle))
	}

	if alle[0].Klaerung != domain.KlaerungQuittiert {
		t.Errorf("die Epoche steht auf %q", alle[0].Klaerung)
	}

	if !alle[0].GeklaertAm.Equal(quittiert) {
		t.Errorf("der Zeitpunkt der Quittung ist %v, erwartet war %v",
			alle[0].GeklaertAm, quittiert)
	}
}

// **Eine Epoche, die es nicht gibt, lässt sich nicht quittieren** — und sagt
// das, statt still nichts zu tun.
func TestEineUnbekannteEpocheLaesstSichNichtQuittieren(t *testing.T) {
	db := frischeDatenbank(t)
	buch := storage.NeuesWiederherstellungsbuch(db)

	fehler := buch.Quittieren(probekontext(t), 4711, time.Now())
	if fehler == nil {
		t.Fatal("eine unbekannte Epoche ließ sich quittieren")
	}

	if !errors.Is(fehler, domain.ErrNotFound) {
		t.Errorf("der Fehler ist nicht ErrNotFound: %v", fehler)
	}
}

// **Die Mengen gehören dem Profil, in dem sie entstanden sind.** Ohne diese
// Trennung sperrte eine Rückkehr in einem Profil die Sammlung des anderen mit.
func TestDieMengenEinesProfilsSperrenNichtDasAndere(t *testing.T) {
	db := frischeDatenbank(t)
	buch := storage.NeuesWiederherstellungsbuch(db)

	if _, fehler := buch.Vermerken(probekontext(t), epoche()); fehler != nil {
		t.Fatalf("die Epoche vermerken: %v", fehler)
	}

	fremd, lesefehler := buch.Offene(storage.MitProfil(t.Context(), 99))
	if lesefehler != nil {
		t.Fatalf("die offenen Epochen lesen: %v", lesefehler)
	}

	if len(fremd) != 1 {
		t.Fatalf("das fremde Profil sieht %d Epochen, erwartet war eine", len(fremd))
	}

	// **Die Epoche selbst sieht jedes Profil** — der Bestand ist gemeinsam
	// zurückgegangen, und das zu verschweigen wäre die Unsicherheit, die
	// niemand sieht. **Die Mengen sieht nur, wem sie gehören.**
	if len(fremd[0].Mengen) != 0 {
		t.Errorf("das fremde Profil sieht %d Mengen: %+v", len(fremd[0].Mengen), fremd[0].Mengen)
	}
}

// **Die Zeitpunkte kommen auf die Millisekunde zurück, nicht feiner.**
//
// `Zeitstempel` schreibt drei Nachkommastellen — der Grund steht in
// `datenbank.go`. Der Lauf steht hier, damit niemand später eine Genauigkeit
// annimmt, die der Bestand nicht hält: eine Zusage, die eine Schicht weiter
// unten nicht gilt, ist schlimmer als eine engere.
func TestDieZeitpunkteKommenAufDieMillisekundeZurueck(t *testing.T) {
	db := frischeDatenbank(t)
	buch := storage.NeuesWiederherstellungsbuch(db)

	fein := epoche()
	fein.Datenstand = time.Date(2026, 9, 1, 8, 30, 0, 123456789, time.UTC)

	if _, fehler := buch.Vermerken(probekontext(t), fein); fehler != nil {
		t.Fatalf("die Epoche vermerken: %v", fehler)
	}

	offen, lesefehler := buch.Offene(probekontext(t))
	if lesefehler != nil {
		t.Fatalf("die offenen Epochen lesen: %v", lesefehler)
	}

	erwartet := time.Date(2026, 9, 1, 8, 30, 0, 123000000, time.UTC)
	if !offen[0].Datenstand.Equal(erwartet) {
		t.Errorf("der Datenstand ist %v, erwartet war %v", offen[0].Datenstand, erwartet)
	}
}

// **Ohne Profil wird vermerkt, aber nicht gelesen.**
//
// Der Unterschied ist der Kern: eine Rückkehr geschieht auch **beim Start**,
// bevor jemand ein Profil gewählt hat — und ausgerechnet die aus einem kaputten
// Bestand bekäme sonst keinen Nachweis. Was *wem* fehlt, ist dagegen eine Frage
// an eine Sammlung, und die gehört einem.
func TestOhneProfilWirdVermerktAberNichtGelesen(t *testing.T) {
	db := frischeDatenbank(t)
	buch := storage.NeuesWiederherstellungsbuch(db)

	kennung, fehler := buch.Vermerken(t.Context(), epoche())
	if fehler != nil {
		t.Fatalf("ohne Profil wurde nicht vermerkt: %v", fehler)
	}

	if _, lesefehler := buch.Offene(t.Context()); lesefehler == nil {
		t.Error("ohne Profil kam eine Liste zurück")
	}

	// **Und die Mengen stehen unter ihrem eigenen Profil**, nicht unter null.
	offene, profilfehler := buch.Offene(probekontext(t))
	if profilfehler != nil {
		t.Fatalf("die offenen Epochen lesen: %v", profilfehler)
	}

	if len(offene) != 1 || offene[0].ID != kennung || len(offene[0].Mengen) != 2 {
		t.Errorf("die Epoche steht als %+v da", offene)
	}
}

// **Und auf einer Datenbank, die zu ist, meldet jeder Weg das** — statt eine
// leere Liste zu geben, die wie ein Bestand ohne Rückkehren aussieht.
func TestAufEinerGeschlossenenDatenbankMeldetJederWegDenFehler(t *testing.T) {
	buch := storage.NeuesWiederherstellungsbuch(zugemachteDatenbank(t))

	if _, fehler := buch.Vermerken(probekontext(t), epoche()); fehler == nil {
		t.Error("Vermerken meldete nichts")
	}

	if _, fehler := buch.Offene(probekontext(t)); fehler == nil {
		t.Error("Offene meldete nichts")
	}

	if _, fehler := buch.Alle(probekontext(t)); fehler == nil {
		t.Error("Alle meldete nichts")
	}

	if fehler := buch.Quittieren(probekontext(t), 1, time.Now()); fehler == nil {
		t.Error("Quittieren meldete nichts")
	}
}

// **Fehlt die Mengentabelle, ist die Epoche keine halbe Auskunft.**
//
// Sie stünde sonst ohne ihre Mengen da — also mit der Aussage „etwas ist
// unklar" und ohne die, **was**. Das ist genau die Hälfte, die niemandem
// nützt (§10).
func TestOhneMengentabelleGibtEsKeineHalbeAuskunft(t *testing.T) {
	db := frischeDatenbank(t)
	buch := storage.NeuesWiederherstellungsbuch(db)

	if _, fehler := buch.Vermerken(probekontext(t), epoche()); fehler != nil {
		t.Fatalf("die Epoche vermerken: %v", fehler)
	}

	if _, fehler := db.ExecContext(probekontext(t),
		`DROP TABLE wiederherstellungsmengen`); fehler != nil {
		t.Fatalf("die Mengentabelle wegräumen: %v", fehler)
	}

	if _, fehler := buch.Offene(probekontext(t)); fehler == nil {
		t.Error("ohne Mengentabelle kam eine Epoche ohne ihre Mengen zurück")
	}
}

// **Ein unlesbarer Zeitpunkt ist ein Fehler und keine Nullzeit.** Eine
// Nullzeit sähe aus wie „1. Januar Jahr eins" und stünde so in der Oberfläche.
func TestEinUnlesbarerZeitpunktMeldetSichStattNullZuGeben(t *testing.T) {
	db := frischeDatenbank(t)
	buch := storage.NeuesWiederherstellungsbuch(db)

	if _, fehler := buch.Vermerken(probekontext(t), epoche()); fehler != nil {
		t.Fatalf("die Epoche vermerken: %v", fehler)
	}

	if _, fehler := db.ExecContext(probekontext(t),
		`UPDATE wiederherstellungen SET datenstand = 'irgendwann'`); fehler != nil {
		t.Fatalf("den Zeitpunkt verderben: %v", fehler)
	}

	if _, fehler := buch.Offene(probekontext(t)); fehler == nil {
		t.Error("ein unlesbarer Zeitpunkt kam als Nullzeit durch")
	}
}

// **Eine Menge ohne Stück wird abgewiesen** — vom Bestand selbst, und der
// Weg gibt den Fehler weiter, statt sie stillschweigend auszulassen.
func TestEineMengeOhneStueckWirdAbgewiesen(t *testing.T) {
	db := frischeDatenbank(t)
	buch := storage.NeuesWiederherstellungsbuch(db)

	leer := epoche()
	leer.Mengen = []domain.Sammlungsmenge{{Profil: 1, Schluessel: "scryfall:abc", Menge: 0}}

	if _, fehler := buch.Vermerken(probekontext(t), leer); fehler == nil {
		t.Error("eine Menge über null Stück kam durch")
	}
}

// zugemachteDatenbank gibt eine Verbindung, die schon zu ist.
//
// **Eine eigene und nicht die aus `fehlerpfade_test.go`**: die liegt im
// paketinternen Testpaket, und dieses hier ist das äußere.
func zugemachteDatenbank(t *testing.T) *sql.DB {
	t.Helper()

	db, fehler := storage.FuerLaeufe("file:" + t.TempDir() + "/zu.db")
	if fehler != nil {
		t.Fatalf("die Datenbank ließ sich nicht öffnen: %v", fehler)
	}

	if schliessfehler := db.Close(); schliessfehler != nil {
		t.Fatalf("die Datenbank ließ sich nicht schließen: %v", schliessfehler)
	}

	return db
}
