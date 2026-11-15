package app

// **Die Naht zur Recherche** — Werkstück A16 von Phase 5.3
// (#1515).
//
// **Ein Lauf im Paket**, weil Ausgang und Profilquelle unexportiert sind: sie
// sind Verdrahtung und keine Fähigkeit — sie zu exportieren, nur damit ein
// Test sie sieht, machte aus einer inneren Naht eine Fläche.

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// eineProbekarte legt ein Set mit einer Karte an und gibt deren Kennung.
func eineProbekarte(t *testing.T, anwendung *App) int64 {
	t.Helper()

	ctx := imProbeprofil(t)

	setID, fehler := storage.NeuerSetSpeicher(anwendung.db).SetAnlegen(ctx,
		&domain.Set{Name: "Probe", Game: "magic", Style: "m15"}, nil)
	if fehler != nil {
		t.Fatalf("Set anlegen: %v", fehler)
	}

	karteID, kartenfehler := storage.NeuerKartenschreiber(anwendung.db).
		KarteAnlegen(ctx, setID)
	if kartenfehler != nil {
		t.Fatalf("Karte anlegen: %v", kartenfehler)
	}

	return karteID
}

// **Die Profilquelle gibt die Kennung aus dem Kontext.**
//
// **Der Fachkern liest sie nicht selbst** (§2): welche Kennung im Kontext
// steht, weiß `storage` — und ohne diesen Lauf stünde die Naht da, ohne je
// gerufen worden zu sein.
func TestDieProfilquelleGibtDieKennungAusDemKontext(t *testing.T) {
	t.Parallel()

	kennung, fehler := (rechercheprofil{}).Profil(imProbeprofil(t))
	if fehler != nil || kennung != 1 {
		t.Fatalf("das Profil ist %d (%v)", kennung, fehler)
	}
}

// **Ohne Profil im Kontext gibt sie einen Fehler und keine Null.**
//
// **Eine Null wäre ein Profil** — und zwar das, dem alles gehört, was keinem
// gehört.
func TestOhneProfilGibtDieProfilquelleEinenFehler(t *testing.T) {
	t.Parallel()

	if _, fehler := (rechercheprofil{}).Profil(t.Context()); fehler == nil {
		t.Fatal("ohne Profil kam kein Fehler")
	}
}

// **Der Ausgang liest die Karte und gibt ihre Kennungen.**
func TestDerAusgangLiestDieKarteUndGibtIhreKennungen(t *testing.T) {
	t.Parallel()

	anwendung := NeueApp("")
	anwendung.db = einProbebestand(t)

	karteID := eineProbekarte(t, anwendung)

	ausgang := kartenausgang{holen: anwendung.laufendeVerbindung}

	_, jetziger, kennungen, fehler := ausgang.Ausgangstext(
		imProbeprofil(t), karteID, "")
	if fehler != nil {
		t.Fatalf("Ausgangstext: %v", fehler)
	}

	// **Der gerechnete Stand kommt heraus**, auch wenn keiner hineinging —
	// nur deshalb kann eine Fläche ihn überhaupt nennen.
	if jetziger == "" {
		t.Fatal("der Ausgang nennt keinen Stand")
	}

	if len(kennungen) != 1 ||
		!strings.HasPrefix(kennungen[0], "eigenekarte:") {
		t.Fatalf("die Kennungen sind %+v", kennungen)
	}
}

// **Eine Karte, die es nicht gibt, ist ein Fehler.**
func TestEineFehlendeAusgangskarteIstEinFehler(t *testing.T) {
	t.Parallel()

	anwendung := NeueApp("")
	anwendung.db = einProbebestand(t)

	_, _, _, fehler := kartenausgang{holen: anwendung.laufendeVerbindung}.
		Ausgangstext(imProbeprofil(t), 4711, "")
	if fehler == nil {
		t.Fatal("die fehlende Karte kam durch")
	}
}

// **Ein überholter Stand wird abgewiesen — mit dem, was jetzt gilt.**
//
// **Hier und nicht im Dienst**: dort stünde eine Zahl gegen eine Zahl, hier
// steht sie gegen die Karte, die es gibt.
func TestEinUeberholterStandDerAusgangskarteWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	anwendung := NeueApp("")
	anwendung.db = einProbebestand(t)

	karteID := eineProbekarte(t, anwendung)

	_, _, _, fehler := kartenausgang{holen: anwendung.laufendeVerbindung}.
		Ausgangstext(imProbeprofil(t), karteID, "einstandvongestern")
	if !errors.Is(fehler, domain.ErrConflict) {
		t.Fatalf("der Fehler ist %v", fehler)
	}

	// **Der Satz nennt beide Stände** — sonst wüsste niemand, worauf er die
	// Frage neu stellen soll.
	if !strings.Contains(domain.SatzVon(fehler), "einstandvongestern") {
		t.Errorf("der Satz ist %q", domain.SatzVon(fehler))
	}
}

// **Der Ausgangstext setzt jeden Feldwert zusammen und lässt leere weg.**
//
// **Alle Felder und keine Auswahl**: welches Feld den Text einer Karte trägt,
// weiß das Paket und nicht dieses Programm (§11).
func TestDerAusgangstextSetztJedenFeldwertZusammen(t *testing.T) {
	t.Parallel()

	text := ausgangstextAus([]domain.Value{
		{Text: "Lightning Bolt"}, {Text: "   "}, {Text: "deals 3 damage"},
	})

	if text != "Lightning Bolt\ndeals 3 damage" {
		t.Fatalf("der Text ist %q", text)
	}
}

// **Die Verdrahtung hängt die Recherche ans Fenster — und den Index dazu.**
//
// **Sie wird hier ohne [App.retrievalVerdrahten] gerufen**, und das ist der
// Punkt: A16 baut keinen zweiten Retrievalkern, also braucht sie den ersten —
// und holt ihn sich, statt über einer leeren Schnittstelle zu hängen.
func TestDieVerdrahtungHaengtDieRechercheAnsFenster(t *testing.T) {
	t.Parallel()

	anwendung := NeueApp("")
	anwendung.db = einProbebestand(t)
	anwendung.rechercheVerdrahten()

	if anwendung.recherche == nil {
		t.Fatal("die Recherche hängt nicht am Fenster")
	}

	if anwendung.retrieval == nil {
		t.Fatal("der Suchindex hängt nicht am Fenster")
	}

	// **Ohne Index ist das eine Auskunft und kein Fehler** — und der Grund
	// steht dabei.
	antwort := anwendung.Recherchieren(imProbeprofil(t),
		string(domain.FallKarten), "bolt", 5, RecherchefilterDTO{}, "")
	if antwort.Fehler != nil {
		t.Fatalf("Recherchieren: %+v", antwort.Fehler)
	}

	if antwort.Leere != string(domain.LeereOhneIndex) || antwort.Grund == "" {
		t.Fatalf("die Antwort ist %+v", antwort)
	}
}

// **Und die Recherche sucht in genau dem Index, den `Indexsuche` bedient.**
//
// # Warum das ein eigener Lauf ist
//
// **Weil zwei Verwaltungen über demselben Index gleich aussehen.** Bis zum
// 2026-09-12 baute [RechercheUeber] eine zweite — und der Riegel gegen einen
// zweiten Aufbau **im selben Prozess** hätte dann zwei Hälften gehabt.
//
// **Zu sehen ist es an der Kennung des Zeigers**: er bindet den
// Indexfingerabdruck, und den nimmt die Recherche aus **ihrer** Verwaltung.
// Trägt er denselben Abdruck wie `Indexstand`, ist es dieselbe.
func TestDieRechercheSuchtInDemIndexDenIndexsucheBedient(t *testing.T) {
	t.Parallel()

	anwendung := NeueApp("")
	anwendung.db = einProbebestand(t)
	anwendung.rechercheVerdrahten()

	// **Ohne Index ist beides leer, und beides aus demselben Grund** — der
	// Grund ist der Beleg, dass es eine Verwaltung ist.
	stand := anwendung.Indexstand(imProbeprofil(t))
	antwort := anwendung.Recherchieren(imProbeprofil(t),
		string(domain.FallKarten), "bolt", 5, RecherchefilterDTO{}, "")

	if stand.Grund == "" || antwort.Grund != stand.Grund {
		t.Fatalf("der Index sagt %q, die Recherche %q", stand.Grund,
			antwort.Grund)
	}
}

// **Und die Anwendung für die Flächen ohne Fenster kennt sie ebenso.**
//
// **Über dieselben Wege wie das Fenster** (§16): zwei Verdrahtungen gäben
// zwei Antworten auf dieselbe Frage.
func TestDieRechercheanwendungKenntDieRecherche(t *testing.T) {
	t.Parallel()

	antwort := RechercheAnwendungUeber(einProbebestand(t)).Recherchieren(
		imProbeprofil(t), string(domain.FallRegeln), "trample", 5,
		RecherchefilterDTO{}, "")

	if antwort.Fehler != nil {
		t.Fatalf("Recherchieren: %+v", antwort.Fehler)
	}

	if antwort.Leere != string(domain.LeereOhneIndex) {
		t.Fatalf("die Antwort ist %+v", antwort)
	}
}

// **Ein Kontextpaket ohne Index entsteht — und sagt, dass es leer ist.**
//
// # Warum es kein Fehler ist
//
// **Weil Zusammenstellen nicht Senden ist.** Der Dienst entscheidet nicht, ob
// etwas hinausgeht; er sagt, was hinausginge. Ein Fehler an dieser Stelle
// nähme dem Aufrufer den Grund weg: `leeresKontextpaket` trägt kein Ergebnis,
// und damit stünde am Fenster „es ging nicht" statt „es ist kein Suchindex
// aktiv".
//
// **Was es dafür geben muss, ist die Null in der Freigabe.** Wer ein Paket
// mit null Belegen sendet, bezahlt ein Modell dafür, aus nichts zu antworten.
func TestEinKontextpaketOhneIndexSagtDassEsLeerIst(t *testing.T) {
	t.Parallel()

	antwort := RechercheAnwendungUeber(einProbebestand(t)).
		KontextZusammenstellen(imProbeprofil(t), "schreib etwas", "rev",
			string(domain.FallRegeln), "trample", 5, ollamaanbieter,
			"gemma3", 4096, 512, 0, nil)

	if antwort.Fehler != nil {
		t.Fatalf("KontextZusammenstellen: %+v", antwort.Fehler)
	}

	if antwort.Freigabe.Belege != 0 || len(antwort.Belege) != 0 {
		t.Fatalf("das Paket trägt Belege ohne Index: %+v", antwort.Belege)
	}

	// **Der Grund steht im Ergebnis** — und nur deshalb ist die leere
	// Antwort brauchbar.
	if antwort.Ergebnis.Leere != string(domain.LeereOhneIndex) ||
		antwort.Ergebnis.Grund == "" {
		t.Fatalf("das Ergebnis ist %+v", antwort.Ergebnis)
	}
}

// **Ohne Einstellungen rechnet Ollama trotzdem auf diesem Rechner.**
//
// **Weil die Vorgabeadresse die lokale ist**: wer nichts eingestellt hat,
// meint das Ollama neben sich.
func TestOhneEinstellungenRechnetOllamaLokal(t *testing.T) {
	t.Parallel()

	if ort := zielOrt(t, NeueApp(""), t.Context(), ollamaanbieter); ort !=
		domain.KIOrtLokal {
		t.Fatalf("der Ort ist %q", ort)
	}
}

// festeKIAdresse antwortet mit genau einer Adresse.
type festeKIAdresse struct {
	Einstellungsfuehrer

	adresse string
}

func (f festeKIAdresse) Stand(context.Context) (map[string]service.Wahl, error) {
	return map[string]service.Wahl{
		domain.KIAdresse: {Wirksam: f.adresse},
	}, nil
}

// **Eine fremde Adresse macht auch Ollama zu einem Dienst in der Ferne.**
//
// **Das ist 0098 in die andere Richtung**: „Loopback allein beweist keine
// lokale Inferenz" — eine fremde Adresse beweist aber „nicht lokal".
func TestEineFremdeAdresseMachtOllamaZumFernenDienst(t *testing.T) {
	t.Parallel()

	anwendung := NeueApp("")
	anwendung.einstellungen = festeKIAdresse{adresse: "https://gpu.example.org"}

	if ort := zielOrt(t, anwendung, t.Context(), ollamaanbieter); ort !=
		domain.KIOrtCloud {
		t.Fatalf("der Ort ist %q", ort)
	}
}
