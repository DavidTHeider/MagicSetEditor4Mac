package app_test

// **Die freigegebenen Sichten an der Naht** — Werkstück `T3` von Phase 5.6
// (#1468).
//
// # Was hier gemessen wird
//
// **Dass die Naht überträgt, ohne zu entscheiden.** Was eine Freigabe zeigt,
// steht in `domain`; ob sie gilt, in `service`. Hier steht, dass die Angaben
// heil durchkommen — und dass ein Nein ein Nein bleibt, statt als leere Liste
// anzukommen.

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/peer"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// einGegenueber trägt ein bestätigtes Gegenüber ein, das handeln darf.
func einGegenueber(t *testing.T, flaeche *app.App) string {
	t.Helper()

	ctx := probekontext(t)

	// **Eine echte Peerkennung** — sie **ist** der öffentliche Schlüssel,
	// und die Naht prüft das. Ein erfundenes Wort käme nicht einmal in die
	// Kontaktliste.
	fremd, fehler := peer.Erzeugen(nil)
	if fehler != nil {
		t.Fatalf("Erzeugen: %v", fehler)
	}

	kennung := string(fremd.Kennung())

	antwort := flaeche.PeerkontaktAufnehmen(ctx, kennung, "Anna")
	if antwort.Fehler != nil {
		t.Fatalf("aufnehmen: %+v", antwort.Fehler)
	}

	abdruck := antwort.Kontakte[0].Fingerabdruck

	if bestaetigt := flaeche.PeerkontaktBestaetigen(
		ctx, kennung, abdruck); bestaetigt.Fehler != nil {
		t.Fatalf("bestätigen: %+v", bestaetigt.Fehler)
	}

	if erlaubt := flaeche.PeerzweckErlauben(
		ctx, kennung, "handel"); erlaubt.Fehler != nil {
		t.Fatalf("erlauben: %+v", erlaubt.Fehler)
	}

	return kennung
}

// **Eine Auswahlfreigabe kommt Feld für Feld zurück.**
func TestEineFreigabeKommtUeberDieNahtZurueck(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	flaeche := app.PeerflaecheUeber(frischerBestand(t))
	ctx := probekontext(t)
	kennung := einGegenueber(t, flaeche)

	antwort := flaeche.SammlungFreigeben(ctx, app.SammlungsfreigabeDTO{
		Empfaenger: kennung, Umfang: "auswahl", Zweck: "handel",
		Auswahl: []string{"abc", "def"}, Ablauf: "2026-12-24T18:00:00Z",
	})
	if antwort.Fehler != nil {
		t.Fatalf("freigeben: %+v", antwort.Fehler)
	}

	if len(antwort.Freigaben) != 1 {
		t.Fatalf("es stehen %d Freigaben da", len(antwort.Freigaben))
	}

	eine := antwort.Freigaben[0]
	if eine.Umfang != "auswahl" || len(eine.Auswahl) != 2 ||
		eine.Ablauf != "2026-12-24T18:00:00Z" || eine.Erteilt == "" {
		t.Errorf("zurück kam %+v", eine)
	}

	// **Und die Aufzählung kommt mit** — zwei Listen liefen auseinander.
	if len(antwort.Umfaenge) != len(domain.DieFreigabeumfaenge) {
		t.Errorf("die Umfänge sind %v", antwort.Umfaenge)
	}
}

// **Und ein Widerruf nimmt sie wieder weg.**
func TestEinWiderrufNimmtDieSichtUeberDieNahtWeg(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	flaeche := app.PeerflaecheUeber(frischerBestand(t))
	ctx := probekontext(t)
	kennung := einGegenueber(t, flaeche)

	if erteilt := flaeche.SammlungFreigeben(ctx, app.SammlungsfreigabeDTO{
		Empfaenger: kennung, Umfang: "sammlung", Zweck: "handel",
	}); erteilt.Fehler != nil {
		t.Fatalf("freigeben: %+v", erteilt.Fehler)
	}

	antwort := flaeche.SammlungsfreigabeWiderrufen(
		ctx, kennung, "handel", "sammlung")
	if antwort.Fehler != nil {
		t.Fatalf("widerrufen: %+v", antwort.Fehler)
	}

	if len(antwort.Freigaben) != 0 {
		t.Errorf("es stehen noch %+v da", antwort.Freigaben)
	}
}

// **Was der Bestand klagt, kommt als Fehler und nicht als leere Liste.**
//
// **Der Unterschied ist für den, der davorsitzt**: „nichts freigegeben" und
// „ich konnte nicht nachsehen" sind zwei Aussagen, und die zweite ist die,
// nach der man etwas tut (§8).
func TestWasDerBestandKlagtKommtAnDerNahtAn(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	db := frischerBestand(t)
	if fehler := db.Close(); fehler != nil {
		t.Fatalf("schließen: %v", fehler)
	}

	antwort := app.PeerflaecheUeber(db).Sammlungsfreigaben(
		probekontext(t), "bwer")

	if antwort.Fehler == nil {
		t.Error("der geschlossene Bestand kam ohne Fehler durch")
	}

	// **Und die Liste ist trotzdem leer statt null** — die Oberfläche zählt
	// sie, auch im Fehlerfall.
	if antwort.Freigaben == nil {
		t.Error("die Liste ist null")
	}
}

// **Ein unlesbarer Ablauf ist ein Feldbefund und kein stiller Nullwert.**
//
// **Sonst hieße ein Tippfehler „unbefristet"** — und das ist die weiteste
// Freigabe, die es gibt.
func TestEinUnlesbarerAblaufIstEinFeldbefund(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	antwort := app.PeerflaecheUeber(frischerBestand(t)).SammlungFreigeben(
		probekontext(t), app.SammlungsfreigabeDTO{
			Empfaenger: "bwer", Umfang: "sammlung", Zweck: "handel",
			Ablauf: "morgen",
		})

	if antwort.Fehler == nil || antwort.Fehler.Feld != "ablauf" {
		t.Errorf("es kam %+v", antwort.Fehler)
	}
}

// **Ein erfundener Umfang kommt nicht durch die Naht.**
func TestEinErfundenerUmfangKommtNichtDurch(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	antwort := app.PeerflaecheUeber(frischerBestand(t)).SammlungFreigeben(
		probekontext(t), app.SammlungsfreigabeDTO{
			Empfaenger: "bwer", Umfang: "alles", Zweck: "handel",
		})

	if antwort.Fehler == nil {
		t.Error("der erfundene Umfang kam durch")
	}
}

// **Ein Widerruf ins Leere ist ein Nein.**
func TestEinWiderrufInsLeereMeldetSichAnDerNaht(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	antwort := app.PeerflaecheUeber(frischerBestand(t)).
		SammlungsfreigabeWiderrufen(
			probekontext(t), "bwer", "handel", "sammlung")

	if antwort.Fehler == nil {
		t.Error("der Widerruf ins Leere kam durch")
	}
}

// **Die Vorschau zeigt, was ein Gegenüber sähe — und sonst nichts.**
func TestDieVorschauZeigtWasDasGegenueberSaehe(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	db := frischerBestand(t)
	flaeche := app.PeerflaecheUeber(db)
	ctx := probekontext(t)

	if _, fehler := storage.NeuerSammlungsspeicher(db).Anlegen(ctx,
		&domain.Exemplar{
			Printing: "abc-123", Menge: 2, Zustand: domain.ZustandPlayed,
			Sprache: "de", Finish: domain.FinishEtched,
			Erwerbspreis: 417, Notiz: "geheim", Lagerort: "Kiste 4",
		}); fehler != nil {
		t.Fatalf("Exemplar: %v", fehler)
	}

	kennung := einGegenueber(t, flaeche)

	// **Ohne Freigabe kommt ein Nein** — und keine leere Liste.
	if ohne := flaeche.Freigabesicht(
		ctx, kennung, "handel"); ohne.Fehler == nil {
		t.Error("die Vorschau kam ohne Freigabe durch")
	}

	if erteilt := flaeche.SammlungFreigeben(ctx, app.SammlungsfreigabeDTO{
		Empfaenger: kennung, Umfang: "sammlung", Zweck: "handel",
	}); erteilt.Fehler != nil {
		t.Fatalf("freigeben: %+v", erteilt.Fehler)
	}

	sicht := flaeche.Freigabesicht(ctx, kennung, "handel")
	if sicht.Fehler != nil {
		t.Fatalf("Vorschau: %+v", sicht.Fehler)
	}

	if len(sicht.Zeilen) != 1 {
		t.Fatalf("die Vorschau hat %d Zeilen", len(sicht.Zeilen))
	}

	zeile := sicht.Zeilen[0]
	if zeile.Menge != 2 || zeile.Finish != "etched" || zeile.Referenz == "" {
		t.Errorf("die Zeile ist %+v", zeile)
	}
}

// **Und die Liste eines Unbekannten ist leer, nicht null.**
//
// **Eine leere Liste und kein `nil`** (§3, „zwischen den Prüfebenen"): die
// Oberfläche zählt sie, und `null.length` ist ein Absturz.
func TestDieFreigabenEinesUnbekanntenSindLeer(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	antwort := app.PeerflaecheUeber(frischerBestand(t)).Sammlungsfreigaben(
		probekontext(t), "bniemand")

	if antwort.Fehler != nil {
		t.Fatalf("es kam %+v", antwort.Fehler)
	}

	if antwort.Freigaben == nil || len(antwort.Freigaben) != 0 {
		t.Errorf("die Liste ist %+v", antwort.Freigaben)
	}
}
