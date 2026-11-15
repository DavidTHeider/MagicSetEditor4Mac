package storage_test

// **Mehrere Karten in einem Zug** — Werkstück A13.1b von Phase 5.3
// (#1452).
//
// # Die Zusage, um die es hier geht
//
// Das Issue: „Gewählte Gruppe entweder vollständig mit Herkunft übernommen
// oder klarer Konflikt **ohne Teilbestand**."
//
// **Sie lässt sich nur hier prüfen.** `service` prüft vorher jeden Schritt
// und schickt **einen** Ruf los — dass darunter **eine** Transaktion liegt,
// sagt allein die Datenbank. Der Weg, es zu messen, ist derselbe wie bei
// A11: eine Herkunft, die sich nicht schreiben lässt, und danach die Frage,
// ob der Wert der **anderen** Karte dasteht.

import (
	"database/sql"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// kartenfeld liest einen Feldwert einer Karte.
//
// **Mit dem Feldnamen**, seit die Bildangaben dazukamen: sie hängen an einem
// Bildfeld, und ein Helfer, der nur »name« läse, prüfte dort das falsche.
//
// **Ein eigener Leser und nicht `feldwert` nebenan**: das ist der Wert eines
// **Schlüsselworts**, und zwei gleich benannte Helfer mit verschiedener
// Bedeutung sind dieselbe Verwechslung, die A13.2b im Sprachkatalog gefunden
// hat.
func kartenfeld(t *testing.T, db *sql.DB, karteID int64, feld string) string {
	t.Helper()

	karte, fehler := storage.NeuerSetLeser(db).KarteLesen(probekontext(t), karteID)
	if fehler != nil {
		t.Fatalf("die Karte %d lesen: %v", karteID, fehler)
	}

	wert, da := karte.Value(feld)
	if !da {
		t.Fatalf("die Karte %d trägt kein Feld %q", karteID, feld)
	}

	return wert
}

// **Zwei Karten, ein Zug — beide stehen danach da.**
func TestEineGruppeSchreibtBeideKarten(t *testing.T) {
	db := frischeDatenbank(t)
	eine := eineKarte(t, db, domain.Value{Field: "name", Text: "Flammenball"})
	andere := eineKarte(t, db, domain.Value{Field: "name", Text: "Eisregen"})

	schreiber := storage.NeuerKartenschreiber(db)

	if fehler := schreiber.GruppeSetzenMitHerkunft(probekontext(t),
		[]domain.Kartenschritt{
			{
				KarteID:    eine,
				Werte:      []domain.Value{{Field: "name", Text: "Blitzschlag"}},
				Herkunft:   "ki",
				Herkuenfte: []domain.KIFeldherkunft{eineFeldherkunft("name")},
			},
			{
				KarteID:    andere,
				Werte:      []domain.Value{{Field: "name", Text: "Frostnova"}},
				Herkunft:   "ki",
				Herkuenfte: []domain.KIFeldherkunft{eineFeldherkunft("name")},
			},
		}); fehler != nil {
		t.Fatalf("die Gruppe schreiben: %v", fehler)
	}

	for karte, erwartet := range map[int64]string{
		eine: "Blitzschlag", andere: "Frostnova",
	} {
		if wert := kartenfeld(t, db, karte, "name"); wert != erwartet {
			t.Errorf("Karte %d trägt %q, erwartet war %q", karte, wert, erwartet)
		}

		gelesen, fehler := storage.NeueFeldherkunft(db).Lesen(probekontext(t), karte)
		if fehler != nil {
			t.Fatalf("die Herkunft lesen: %v", fehler)
		}

		if len(gelesen) != 1 {
			t.Errorf("Karte %d trägt %d Herkunftszeilen", karte, len(gelesen))
		}
	}
}

// **Scheitert die zweite Karte, steht auch die erste nicht da.**
//
// # Wie hier gemessen wird
//
// **An einer Herkunft, die es nicht geben kann**: `eineFeldherkunft` mit
// einem Feldnamen, den die Karte nicht trägt, läuft durch — geprüft wird
// deshalb über den **Fremdschlüssel**, wie bei A11: eine Bildangabe zu einem
// Feld, für das keine Herkunft geschrieben wurde.
//
// **Der einfachere Weg ist ein leerer Schritt**: `Herkuenfte` ohne Eintrag
// weist [storage.Kartenschreiber.FelderSetzenMitHerkunft] ausdrücklich ab
// („eine Übernahme trägt ihre Herkunft — ohne sie ist sie keine"), und die
// Gruppe muss dieselbe Regel haben. Scheitert sie am **zweiten** Schritt,
// darf vom ersten nichts dastehen.
func TestEinFehlerImZweitenSchrittLaesstDenErstenNichtStehen(t *testing.T) {
	db := frischeDatenbank(t)
	eine := eineKarte(t, db, domain.Value{Field: "name", Text: "Flammenball"})
	andere := eineKarte(t, db, domain.Value{Field: "name", Text: "Eisregen"})

	schreiber := storage.NeuerKartenschreiber(db)

	fehler := schreiber.GruppeSetzenMitHerkunft(probekontext(t),
		[]domain.Kartenschritt{
			{
				KarteID:    eine,
				Werte:      []domain.Value{{Field: "name", Text: "Blitzschlag"}},
				Herkunft:   "ki",
				Herkuenfte: []domain.KIFeldherkunft{eineFeldherkunft("name")},
			},
			{
				KarteID:  andere,
				Werte:    []domain.Value{{Field: "name", Text: "Frostnova"}},
				Herkunft: "ki",
				// **Ohne Herkunft** — und damit ein Schritt, den die Ablage
				// abweist.
			},
		})
	if fehler == nil {
		t.Fatal("der Schritt ohne Herkunft kam durch")
	}

	// **Die erste Karte trägt noch ihren alten Wert.** Das ist die ganze
	// Zusage: kein Teilbestand.
	if wert := kartenfeld(t, db, eine, "name"); wert != "Flammenball" {
		t.Errorf("die erste Karte trägt %q — der Teilbestand steht da", wert)
	}

	if wert := kartenfeld(t, db, andere, "name"); wert != "Eisregen" {
		t.Errorf("die zweite Karte trägt %q", wert)
	}

	// **Und keine Herkunftszeile ist zurückgeblieben.**
	gelesen, lesefehler := storage.NeueFeldherkunft(db).Lesen(probekontext(t), eine)
	if lesefehler != nil {
		t.Fatalf("die Herkunft lesen: %v", lesefehler)
	}

	if len(gelesen) != 0 {
		t.Errorf("es stehen %d Herkunftszeilen da", len(gelesen))
	}
}

// **Eine leere Gruppe ändert nichts und ist kein Fehler.**
//
// **Dieselbe Regel wie bei den Werten** ([storage.Kartenschreiber] oben): wer
// nichts zu schreiben hat, bekommt dafür keine Meldung. Die Entscheidung, ob
// eine leere Gruppe überhaupt vorkommen darf, liegt in `service` — dort ist
// sie ein Fehler, und dort steht auch, warum.
func TestEineLeereGruppeAendertNichts(t *testing.T) {
	db := frischeDatenbank(t)
	karte := eineKarte(t, db, domain.Value{Field: "name", Text: "Flammenball"})

	if fehler := storage.NeuerKartenschreiber(db).GruppeSetzenMitHerkunft(
		probekontext(t), nil); fehler != nil {
		t.Fatalf("die leere Gruppe: %v", fehler)
	}

	if wert := kartenfeld(t, db, karte, "name"); wert != "Flammenball" {
		t.Errorf("die Karte trägt %q", wert)
	}
}

// **Ohne Profil im Kontext schreibt auch die Gruppe nichts.**
//
// **Dieselbe Regel wie beim einzelnen Weg** — die Herkunft steht je Profil,
// und eine Zeile unter einem geratenen Profil wäre schlimmer als keine. **Und
// die Prüfung liegt vor der Transaktion**: sonst stünde eine offene da, die
// niemand mehr schließt.
func TestOhneProfilSchreibtDieGruppeNichts(t *testing.T) {
	db := frischeDatenbank(t)
	karte := eineKarte(t, db, domain.Value{Field: "name", Text: "Flammenball"})

	if fehler := storage.NeuerKartenschreiber(db).GruppeSetzenMitHerkunft(
		t.Context(), []domain.Kartenschritt{{
			KarteID:    karte,
			Werte:      []domain.Value{{Field: "name", Text: "Blitzschlag"}},
			Herkunft:   "ki",
			Herkuenfte: []domain.KIFeldherkunft{eineFeldherkunft("name")},
		}}); fehler == nil {
		t.Fatal("ohne Profil wurde geschrieben")
	}

	if wert := kartenfeld(t, db, karte, "name"); wert != "Flammenball" {
		t.Errorf("die Karte trägt %q", wert)
	}
}

// **Eine scheiternde Herkunft nimmt die ganze Gruppe mit.**
//
// # Wie hier gemessen wird
//
// **Mit einer Herkunft, die die Datenbank abweist**: derselbe Feldname
// zweimal verletzt den Primärschlüssel (A11, `kifeldherkunft_test.go`). Die
// Werte **beider** Karten sind zu diesem Zeitpunkt schon geschrieben — läge
// keine umspannende Transaktion darunter, stünde die erste Karte geändert da.
func TestEineScheiterndeHerkunftNimmtDieGanzeGruppeMit(t *testing.T) {
	db := frischeDatenbank(t)
	eine := eineKarte(t, db, domain.Value{Field: "name", Text: "Flammenball"})
	andere := eineKarte(t, db, domain.Value{Field: "name", Text: "Eisregen"})

	fehler := storage.NeuerKartenschreiber(db).GruppeSetzenMitHerkunft(
		probekontext(t), []domain.Kartenschritt{
			{
				KarteID:    eine,
				Werte:      []domain.Value{{Field: "name", Text: "Blitzschlag"}},
				Herkunft:   "ki",
				Herkuenfte: []domain.KIFeldherkunft{eineFeldherkunft("name")},
			},
			{
				KarteID:  andere,
				Werte:    []domain.Value{{Field: "name", Text: "Frostnova"}},
				Herkunft: "ki",
				Herkuenfte: []domain.KIFeldherkunft{
					eineFeldherkunft("name"), eineFeldherkunft("name"),
				},
			},
		})
	if fehler == nil {
		t.Fatal("die doppelte Herkunft ging durch")
	}

	if wert := kartenfeld(t, db, eine, "name"); wert != "Flammenball" {
		t.Errorf("die erste Karte trägt %q — der Teilbestand steht da", wert)
	}
}

// **Die Bildangaben einer Gruppe gehen mit** — A12 (#1451) in der Gruppe.
//
// **In derselben Transaktion wie die Herkunft**, auf die ihr Fremdschlüssel
// zeigt: eine Angabe ohne ihre Herkunft wäre eine Angabe über nichts.
func TestEineGruppeSchreibtIhreBildangaben(t *testing.T) {
	db := frischeDatenbank(t)
	karte := eineKarte(t, db, domain.Value{Field: "bild", Text: "alt.png"})

	if fehler := storage.NeuerKartenschreiber(db).GruppeSetzenMitHerkunft(
		probekontext(t), []domain.Kartenschritt{{
			KarteID:    karte,
			Werte:      []domain.Value{{Field: "bild", Text: "neu.png"}},
			Herkunft:   "ki",
			Herkuenfte: []domain.KIFeldherkunft{eineFeldherkunft("bild")},
			Angaben: map[string][]domain.KIBildparameter{
				"bild": {{Name: "groesse", Wert: "1024x1024"}},
			},
		}}); fehler != nil {
		t.Fatalf("die Gruppe mit Bildangaben: %v", fehler)
	}

	if wert := kartenfeld(t, db, karte, "bild"); wert != "neu.png" {
		t.Errorf("die Karte trägt %q", wert)
	}
}

// **Eine Bildangabe zu einem Feld ohne Herkunft nimmt die Gruppe mit.**
//
// **Der Fremdschlüssel zeigt auf die Herkunftszeile**, und zu »farbe« gibt es
// keine — die Datenbank weist es ab, und danach steht nichts da.
func TestEineBildangabeOhneHerkunftBrichtDieGruppe(t *testing.T) {
	db := frischeDatenbank(t)
	karte := eineKarte(t, db, domain.Value{Field: "bild", Text: "alt.png"})

	fehler := storage.NeuerKartenschreiber(db).GruppeSetzenMitHerkunft(
		probekontext(t), []domain.Kartenschritt{{
			KarteID:    karte,
			Werte:      []domain.Value{{Field: "bild", Text: "neu.png"}},
			Herkunft:   "ki",
			Herkuenfte: []domain.KIFeldherkunft{eineFeldherkunft("bild")},
			Angaben: map[string][]domain.KIBildparameter{
				"farbe": {{Name: "groesse", Wert: "1024x1024"}},
			},
		}})
	if fehler == nil {
		t.Fatal("die Angabe ohne Herkunft ging durch")
	}

	if wert := kartenfeld(t, db, karte, "bild"); wert != "alt.png" {
		t.Errorf("die Karte trägt %q — der Teilbestand steht da", wert)
	}
}

// **Eine geschlossene Datenbank meldet sich, statt still nichts zu tun.**
//
// **Der Weg dorthin ist `BeginTx`** — er ist die erste Berührung mit der
// Datei, und ohne diesen Lauf bliebe seine Übersetzung ungeprüft.
func TestEineGeschlosseneDatenbankMeldetSichBeiDerGruppe(t *testing.T) {
	db := frischeDatenbank(t)
	karte := eineKarte(t, db, domain.Value{Field: "name", Text: "Flammenball"})

	schreiber := storage.NeuerKartenschreiber(db)

	if fehler := db.Close(); fehler != nil {
		t.Fatalf("schließen: %v", fehler)
	}

	if fehler := schreiber.GruppeSetzenMitHerkunft(probekontext(t),
		[]domain.Kartenschritt{{
			KarteID:    karte,
			Werte:      []domain.Value{{Field: "name", Text: "Blitzschlag"}},
			Herkunft:   "ki",
			Herkuenfte: []domain.KIFeldherkunft{eineFeldherkunft("name")},
		}}); fehler == nil {
		t.Fatal("die geschlossene Datenbank ging durch")
	}
}

// **Eine Karte, die es nicht gibt, bricht die Gruppe — beim Schreiben.**
//
// **Ein anderer Weg als die fehlende Herkunft**: der liegt vor der Schleife,
// dieser darin. Und er ist der, der sagt, ob die umspannende Transaktion auch
// dann trägt, wenn schon ein Schritt durch ist.
func TestEineUnbekannteKarteBrichtDieGruppeBeimSchreiben(t *testing.T) {
	db := frischeDatenbank(t)
	eine := eineKarte(t, db, domain.Value{Field: "name", Text: "Flammenball"})

	fehler := storage.NeuerKartenschreiber(db).GruppeSetzenMitHerkunft(
		probekontext(t), []domain.Kartenschritt{
			{
				KarteID:    eine,
				Werte:      []domain.Value{{Field: "name", Text: "Blitzschlag"}},
				Herkunft:   "ki",
				Herkuenfte: []domain.KIFeldherkunft{eineFeldherkunft("name")},
			},
			{
				KarteID:    999,
				Werte:      []domain.Value{{Field: "name", Text: "Frostnova"}},
				Herkunft:   "ki",
				Herkuenfte: []domain.KIFeldherkunft{eineFeldherkunft("name")},
			},
		})
	if fehler == nil {
		t.Fatal("die Karte 999 wurde geschrieben")
	}

	if wert := kartenfeld(t, db, eine, "name"); wert != "Flammenball" {
		t.Errorf("die erste Karte trägt %q — der Teilbestand steht da", wert)
	}
}

// **Eine Gruppe ist ein Schritt im Verlauf** (#2396): seit der Schreibweg
// nachzieht, wer ein Feld liest, trägt jede Karte mehrere Einträge — und ein
// ⌘Z, das nur den jüngsten nähme, ließe die Gruppe halb stehen, genau den
// Teilbestand, den A13.1b ausschließt.
func TestEinZurueckNimmtDieGanzeGruppe(t *testing.T) {
	db, setID, karten := dreiKarten(t)

	schritte := make([]domain.Kartenschritt, 0, 2)
	for _, karte := range karten[:2] {
		schritte = append(schritte, domain.Kartenschritt{
			KarteID: karte, SetID: setID, Herkunft: "ki",
			Werte: []domain.Value{
				{Field: "name", Text: "Blitzschlag"}, {Field: "rule text", Text: "Fliegend"},
			},
			Herkuenfte: []domain.KIFeldherkunft{eineFeldherkunft("name")},
		})
	}

	if fehler := storage.NeuerKartenschreiber(db).GruppeSetzenMitHerkunft(probekontext(t), schritte); fehler != nil {
		t.Fatalf("die Gruppe schreiben: %v", fehler)
	}

	schritt, da, fehler := storage.NeuerVerlaufsleser(db).Zuruecknehmen(probekontext(t), setID)
	if fehler != nil || !da {
		t.Fatalf("zurücknehmen: %v, da %v", fehler, da)
	}

	// **Der Schritt zählt Karten**, nicht Einträge — zwei.
	if schritt.Umfang != 2 || schritt.KarteID != karten[0] || schritt.Feld != "name" {
		t.Errorf("zurückgenommen: Karte %d, Feld %q, Umfang %d — erwartet Karte %d, name, zwei",
			schritt.KarteID, schritt.Feld, schritt.Umfang, karten[0])
	}

	for _, karte := range karten[:2] {
		gelesen, fehler := storage.NeuerSetLeser(db).KarteLesen(probekontext(t), karte)
		if fehler != nil {
			t.Fatalf("die Karte %d lesen: %v", karte, fehler)
		}

		if wert, da := gelesen.Value("rule text"); da {
			t.Errorf("Karte %d trägt nach einem ⌘Z noch %q", karte, wert)
		}
	}
}
