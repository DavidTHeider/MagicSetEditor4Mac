package storage_test

// **Woher ein Deck und woher ein Exemplar kam** — Werkstück 12 von Phase 4.12
// (#1351).
//
// **Ohne `t.Parallel()`**, wie jeder Lauf in diesem Paket: sie teilen eine
// Datenquelle im Arbeitsspeicher.
//
// **`deckherkunft_test.go` und nicht `herkunft_test.go`**: der Name war seit
// #1044
// vergeben — dort steht die Herkunft eines **Kartenfeldwerts** (`wert_herkunft`,
// Wanderung 0014). Zwei verschiedene Fragen, zwei Dateien.

import (
	"database/sql"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// einDeck legt eines an und gibt seine Kennung.
func einDeck(t *testing.T, db *sql.DB) int64 {
	t.Helper()

	angelegt, fehler := storage.NeueDeckablage(db).Anlegen(probekontext(t),
		domain.Deck{Name: "Nachtlauf", Spiel: "probe"})
	if fehler != nil {
		t.Fatalf("das Deck anlegen: %v", fehler)
	}

	return angelegt.ID
}

// **Eine Herkunft geht in den Bestand und kommt zurück.**
func TestEineDeckherkunftGehtInDenBestandUndZurueck(t *testing.T) {
	db := frischeDatenbank(t)
	ablage := storage.NeueHerkunftsablage(db)
	deck := einDeck(t, db)

	gesetzt := domain.Herkunft{
		Art: domain.HerkunftProfil, Quelle: "Ada", QuellProfil: 2,
	}

	if fehler := ablage.DeckherkunftSetzen(
		probekontext(t), deck, gesetzt, 41); fehler != nil {
		t.Fatalf("setzen: %v", fehler)
	}

	gelesen, fehler := ablage.Deckherkunft(probekontext(t), deck)
	if fehler != nil {
		t.Fatalf("lesen: %v", fehler)
	}

	if gelesen.Art != gesetzt.Art || gelesen.Quelle != gesetzt.Quelle ||
		gelesen.QuellProfil != gesetzt.QuellProfil {
		t.Errorf("zurück kam %+v", gelesen)
	}

	// **Der Zeitpunkt wird gesetzt und nicht mitgegeben** — er sagt, wann
	// **dieser** Bestand die Übernahme gesehen hat.
	if gelesen.Uebernommen == "" {
		t.Error("die Herkunft nennt keinen Zeitpunkt")
	}
}

// **Ein Deck ohne Herkunft ist hier entstanden — und das ist kein Fehler.**
//
// **Die wichtigste Zusage dieser Ablage.** Ein `ErrNotFound` zwänge jeden
// Aufrufer, ihn abzufangen und in „keine" zu übersetzen; wer es einmal
// vergäße, machte aus einem eigenen Deck einen Fehler.
func TestEinDeckOhneHerkunftGibtEineLeere(t *testing.T) {
	db := frischeDatenbank(t)

	gelesen, fehler := storage.NeueHerkunftsablage(db).Deckherkunft(
		probekontext(t), einDeck(t, db))
	if fehler != nil {
		t.Fatalf("lesen: %v", fehler)
	}

	if gelesen.Art != "" {
		t.Errorf("zurück kam %+v", gelesen)
	}
}

// **Eine zweite Übernahme schreibt die erste fort.**
//
// **Und legt keine zweite Zeile an**: ein Deck kommt von **einem** Ort. Wer es
// noch einmal übernimmt, hat es noch einmal von dort — oder von woanders, und
// dann gilt das Neue.
func TestEineZweiteUebernahmeSchreibtFort(t *testing.T) {
	db := frischeDatenbank(t)
	ablage := storage.NeueHerkunftsablage(db)
	deck := einDeck(t, db)

	if fehler := ablage.DeckherkunftSetzen(probekontext(t), deck, domain.Herkunft{
		Art: domain.HerkunftProfil, Quelle: "Ada", QuellProfil: 2,
	}, 41); fehler != nil {
		t.Fatalf("erste: %v", fehler)
	}

	if fehler := ablage.DeckherkunftSetzen(probekontext(t), deck, domain.Herkunft{
		Art: domain.HerkunftEinfuhr, Quelle: "nachtlauf.txt",
	}, 0); fehler != nil {
		t.Fatalf("zweite: %v", fehler)
	}

	gelesen, fehler := ablage.Deckherkunft(probekontext(t), deck)
	if fehler != nil {
		t.Fatalf("lesen: %v", fehler)
	}

	if gelesen.Art != domain.HerkunftEinfuhr || gelesen.QuellProfil != 0 {
		t.Errorf("zurück kam %+v", gelesen)
	}
}

// **Eine Herkunft, die in sich nicht stimmt, kommt nicht in den Bestand.**
func TestEineUnstimmigeHerkunftWirdAbgewiesen(t *testing.T) {
	db := frischeDatenbank(t)

	if fehler := storage.NeueHerkunftsablage(db).DeckherkunftSetzen(
		probekontext(t), einDeck(t, db),
		domain.Herkunft{Art: domain.HerkunftProfil, Quelle: "Ada"}, 0,
	); fehler == nil {
		t.Error("eine Profilherkunft ohne Profil kam durch")
	}
}

// **Und ein gelöschtes Deck nimmt seine Herkunft mit.**
//
// **`ON DELETE CASCADE` auf das eigene Ding** — die Gegenprobe zu der Zeile in
// der Wanderung. Eine Herkunft ohne das, wovon sie die Herkunft ist, würde
// wieder gültig, sobald jemand dieselbe Kennung bekommt.
func TestEinGeloeschtesDeckNimmtSeineHerkunftMit(t *testing.T) {
	db := frischeDatenbank(t)
	ablage := storage.NeueHerkunftsablage(db)
	deck := einDeck(t, db)

	if fehler := ablage.DeckherkunftSetzen(probekontext(t), deck, domain.Herkunft{
		Art: domain.HerkunftProfil, Quelle: "Ada", QuellProfil: 2,
	}, 41); fehler != nil {
		t.Fatalf("setzen: %v", fehler)
	}

	if fehler := storage.NeueDeckablage(db).Loeschen(probekontext(t), deck); fehler != nil {
		t.Fatalf("löschen: %v", fehler)
	}

	var zeilen int
	if err := db.QueryRowContext(t.Context(),
		`SELECT count(*) FROM deck_herkunft WHERE deck_id = ?`, deck,
	).Scan(&zeilen); err != nil {
		t.Fatalf("zählen: %v", err)
	}

	if zeilen != 0 {
		t.Errorf("es blieben %d Zeilen stehen", zeilen)
	}
}

// **Und sie reist mit dem Deck, ohne dass jemand danach fragt.**
//
// **Die Zusage, für die es den `LEFT JOIN` gibt** (§10): ein zweiter Ruf wäre
// eine zweite Stelle, an die jeder Aufrufer denken müsste — und wer es
// vergäße, zeigte ein übernommenes Deck als eigenes.
func TestDieHerkunftReistMitDemDeck(t *testing.T) {
	db := frischeDatenbank(t)
	deck := einDeck(t, db)

	if fehler := storage.NeueHerkunftsablage(db).DeckherkunftSetzen(
		probekontext(t), deck, domain.Herkunft{
			Art: domain.HerkunftProfil, Quelle: "Ada", QuellProfil: 2,
		}, 41); fehler != nil {
		t.Fatalf("setzen: %v", fehler)
	}

	gelesen, fehler := storage.NeueDeckablage(db).Deck(probekontext(t), deck)
	if fehler != nil {
		t.Fatalf("lesen: %v", fehler)
	}

	if gelesen.Herkunft.Art != domain.HerkunftProfil ||
		gelesen.Herkunft.Quelle != "Ada" {
		t.Errorf("das Deck nennt %+v", gelesen.Herkunft)
	}
}

// **Und ein Deck ohne Herkunft nennt die leere** — nicht `null`, nicht Unsinn.
func TestEinDeckOhneHerkunftNenntDieLeere(t *testing.T) {
	db := frischeDatenbank(t)

	gelesen, fehler := storage.NeueDeckablage(db).Deck(
		probekontext(t), einDeck(t, db))
	if fehler != nil {
		t.Fatalf("lesen: %v", fehler)
	}

	if gelesen.Herkunft.Art != "" || gelesen.Herkunft.Quelle != "" {
		t.Errorf("das Deck nennt %+v", gelesen.Herkunft)
	}
}

// **Dieselbe Zusage für ein Exemplar.**
func TestDieHerkunftReistMitDemExemplar(t *testing.T) {
	db := frischeDatenbank(t)
	speicher := storage.NeuerSammlungsspeicher(db)
	ctx := probekontext(t)

	exemplar := &domain.Exemplar{
		Printing: "a1b2c3d4-0000-4000-8000-00000000000a", Menge: 2,
		Zustand: domain.ZustandNearMint, Sprache: "en",
	}

	kennung, fehler := speicher.Anlegen(ctx, exemplar)
	if fehler != nil {
		t.Fatalf("anlegen: %v", fehler)
	}

	if setzfehler := storage.NeueHerkunftsablage(db).ExemplarherkunftSetzen(
		ctx, kennung, domain.Herkunft{
			Art: domain.HerkunftEinfuhr, Quelle: "sammlung.csv",
		}); setzfehler != nil {
		t.Fatalf("setzen: %v", setzfehler)
	}

	liste, fehler := speicher.Liste(ctx)
	if fehler != nil {
		t.Fatalf("lesen: %v", fehler)
	}

	if len(liste) != 1 || liste[0].Herkunft.Art != domain.HerkunftEinfuhr ||
		liste[0].Herkunft.Quelle != "sammlung.csv" {
		t.Fatalf("die Liste ist %+v", liste)
	}

	// **Und über den zweiten Weg ebenso** — `Fuer` liest dieselbe Tabelle mit
	// einer anderen Bedingung, und ein `JOIN` nur an einer der beiden Stellen
	// wäre die halbe Zusage (§10).
	fuer, lesefehler := speicher.Fuer(ctx, exemplar.Printing)
	if lesefehler != nil {
		t.Fatalf("Fuer: %v", lesefehler)
	}

	if len(fuer) != 1 || fuer[0].Herkunft.Quelle != "sammlung.csv" {
		t.Errorf("Fuer gab %+v", fuer)
	}
}

// **Die Herkunft eines Exemplars kommt einzeln zurück.**
//
// **Ein eigener Weg neben dem `LEFT JOIN`**: jener bringt sie mit der Liste,
// dieser beantwortet „woher kam **diese** Zeile" ohne die Liste. Beide werden
// gebraucht, und beide müssen dasselbe sagen.
func TestEineExemplarherkunftKommtEinzelnZurueck(t *testing.T) {
	db := frischeDatenbank(t)
	speicher := storage.NeuerSammlungsspeicher(db)
	ablage := storage.NeueHerkunftsablage(db)
	ctx := probekontext(t)

	kennung, fehler := speicher.Anlegen(ctx, &domain.Exemplar{
		Printing: "a1b2c3d4-0000-4000-8000-00000000000b", Menge: 1,
		Zustand: domain.ZustandNearMint, Sprache: "en",
	})
	if fehler != nil {
		t.Fatalf("anlegen: %v", fehler)
	}

	// **Ohne Zeile die leere** — dieselbe Zusage wie beim Deck.
	leer, lesefehler := ablage.Exemplarherkunft(ctx, kennung)
	if lesefehler != nil || leer.Art != "" {
		t.Fatalf("leer = %+v, Fehler %v", leer, lesefehler)
	}

	if setzfehler := ablage.ExemplarherkunftSetzen(ctx, kennung, domain.Herkunft{
		Art: domain.HerkunftEinfuhr, Quelle: "sammlung.csv",
	}); setzfehler != nil {
		t.Fatalf("setzen: %v", setzfehler)
	}

	gelesen, lesefehler2 := ablage.Exemplarherkunft(ctx, kennung)
	if lesefehler2 != nil {
		t.Fatalf("lesen: %v", lesefehler2)
	}

	if gelesen.Art != domain.HerkunftEinfuhr || gelesen.Quelle != "sammlung.csv" {
		t.Errorf("zurück kam %+v", gelesen)
	}
}

// **Und eine unstimmige Exemplarherkunft kommt nicht hinein.**
func TestEineUnstimmigeExemplarherkunftWirdAbgewiesen(t *testing.T) {
	db := frischeDatenbank(t)

	if fehler := storage.NeueHerkunftsablage(db).ExemplarherkunftSetzen(
		probekontext(t), 1,
		domain.Herkunft{Art: domain.HerkunftEinfuhr, Quelle: "x", QuellProfil: 2},
	); fehler == nil {
		t.Error("eine Einfuhr mit Profil kam durch")
	}
}

// **Ein geschlossener Bestand meldet sich an jedem der vier Wege.**
//
// **Jeder einzeln**, weil jeder seine eigene Übersetzung hat — eine
// vergessene gäbe eine rohe SQLite-Meldung an die Oberfläche (§8).
func TestJederHerkunftswegMeldetDenGeschlossenenBestand(t *testing.T) {
	db := frischeDatenbank(t)
	ablage := storage.NeueHerkunftsablage(db)
	ctx := probekontext(t)

	if fehler := db.Close(); fehler != nil {
		t.Fatalf("schließen: %v", fehler)
	}

	if _, fehler := ablage.Deckherkunft(ctx, 1); fehler == nil {
		t.Error("Deckherkunft schwieg")
	}

	if _, fehler := ablage.Exemplarherkunft(ctx, 1); fehler == nil {
		t.Error("Exemplarherkunft schwieg")
	}

	gueltig := domain.Herkunft{
		Art: domain.HerkunftProfil, Quelle: "Ada", QuellProfil: 2,
	}

	if fehler := ablage.DeckherkunftSetzen(ctx, 1, gueltig, 7); fehler == nil {
		t.Error("DeckherkunftSetzen schwieg")
	}

	if fehler := ablage.ExemplarherkunftSetzen(ctx, 1, gueltig); fehler == nil {
		t.Error("ExemplarherkunftSetzen schwieg")
	}
}

// **Die Entnahme liest über die Profilgrenze — und nur sie.**
//
// # Warum dieser Lauf eigens dasteht
//
// **Weil er die einzige Stelle prüft, die die Klemme aus Werkstück 2
// absichtlich überschreitet.** Was er belegt, ist beides: dass sie es **kann**
// (sonst wäre die Deckübernahme unmöglich) und dass die **normale** Abfrage es
// **nicht** kann (sonst wäre die Trennung keine).
func TestDieEntnahmeLiestUeberDieProfilgrenze(t *testing.T) {
	db := frischeDatenbank(t)

	angelegt, fehler := storage.NeueDeckablage(db).Anlegen(imProfil(t, 2),
		domain.Deck{Name: "Drueben", Spiel: "probe"})
	if fehler != nil {
		t.Fatalf("das fremde Deck anlegen: %v", fehler)
	}

	// **Aus Profil 1 heraus gelesen** — die gewöhnliche Abfrage findet es
	// nicht, die Entnahme schon.
	if _, lesefehler := storage.NeueDeckablage(db).Deck(
		imProfil(t, 1), angelegt.ID); lesefehler == nil {
		t.Error("die gewöhnliche Abfrage sah über die Profilgrenze")
	}

	gelesen, entnahmefehler := storage.NeueDeckentnahme(db).DeckAusProfil(
		imProfil(t, 1), 2, angelegt.ID)
	if entnahmefehler != nil {
		t.Fatalf("die Entnahme: %v", entnahmefehler)
	}

	if gelesen.Name != "Drueben" {
		t.Errorf("gelesen wurde %+v", gelesen)
	}
}
