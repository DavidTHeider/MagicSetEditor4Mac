package storage_test

// **Der Index im Bestand** — Werkstück A15 von Phase 5.3
// (#1514).
//
// # Was hier gemessen wird
//
// **Dass ein Neubau nicht mitten im Lesen sichtbar wird.** Eine Generation
// entsteht, wird gefüllt, geprüft und dann **atomar** aktiv; wer währenddessen
// sucht, sieht den alten Stand oder gar nichts — nie eine Mischung.
//
// **Und dass fremde Texte nicht durchkommen.** Weder als Treffer, noch als
// Zahl, noch als Ausschnitt. Das Issue verlangt die Prüfung zweimal: „bereits
// in Kandidatenauswahl und erneut vor Ergebnis-/Snippet-/Kontextausgabe."

import (
	"errors"
	"slices"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// einIndexmanifest baut ein vollständiges.
func einIndexmanifest() domain.Indexmanifest {
	return domain.Indexmanifest{
		Schemafassung: 1, Zerteilfassung: "abschnitte-1", Indexfassung: 1,
		Modell: "all-minilm", Modellrevision: "1b226e2802db",
		Laufzeit: "ollama 0.33.3", Abstandsmass: domain.AbstandCosinus,
		Vertrag: domain.KIEinbettungsvertrag{
			Dimension: 3, Pooling: domain.KIPoolingMittel,
			Normalisiert: true, Zerteiler: "bert",
		},
		Korpusstaende: []domain.Korpusstand{
			{Quelle: domain.QuelleOracle, Revision: "2026-09-01"},
		},
		Suchkonfiguration: "rrf-60",
	}
}

// einIndexdokument baut ein öffentliches Oracle-Dokument.
func einIndexdokument(kennung, text string) domain.Retrievaldokument {
	return domain.Retrievaldokument{
		Kennung: kennung, Quelle: domain.QuelleOracle,
		Bezugsadresse: "https://beispiel.invalid/" + kennung,
		Spiel:         "magic", Sprache: "en", Titel: kennung, Text: text,
		Quellrevision: "2026-09-01", Quelldatum: "2026-09-01",
		Abgerufen:    time.Date(2026, 9, 12, 9, 0, 0, 0, time.UTC),
		Inhaltshash:  domain.Retrievalhash(text),
		Bis:          len(text),
		Sichtbarkeit: domain.SichtbarOeffentlich,
	}
}

// einEigenesDokument gehört einem Profil.
func einEigenesDokument(kennung, text string, profil int64) domain.Retrievaldokument {
	dokument := einIndexdokument(kennung, text)
	dokument.Quelle = domain.QuelleEigenerText
	dokument.Bezugsadresse = ""
	dokument.Sichtbarkeit = domain.SichtbarProfil
	dokument.ProfilID = profil

	return dokument
}

// mitIndex baut einen gefüllten, aktiven Index und gibt die Ablage.
func mitIndex(t *testing.T, dokumente ...domain.Retrievaldokument) *storage.Retrievalindex {
	t.Helper()

	index := storage.NeuerRetrievalindex(frischeDatenbank(t))

	generation, fehler := index.Anlegen(probekontext(t), einIndexmanifest())
	if fehler != nil {
		t.Fatalf("Anlegen: %v", fehler)
	}

	for i, dokument := range dokumente {
		vektor := []float32{float32(i + 1), 0, 0}

		if fehler := index.Schreiben(probekontext(t), generation,
			dokument, vektor); fehler != nil {
			t.Fatalf("Schreiben: %v", fehler)
		}
	}

	if fehler := index.Aktivieren(probekontext(t), generation); fehler != nil {
		t.Fatalf("Aktivieren: %v", fehler)
	}

	return index
}

// **Ein gebauter Index gibt seine Treffer — lexikalisch.**
func TestDerIndexFindetLexikalisch(t *testing.T) {
	index := mitIndex(t,
		einIndexdokument("bolt", "Lightning Bolt deals 3 damage to any target."),
		einIndexdokument("shock", "Shock deals 2 damage to any target."))

	treffer, fehler := index.Lexikalisch(probekontext(t), "lightning", 10)
	if fehler != nil {
		t.Fatalf("Lexikalisch: %v", fehler)
	}

	if len(treffer) != 1 || treffer[0].Dokument.Kennung != "bolt" {
		t.Fatalf("die Treffer sind %+v", treffer)
	}

	// **Der Ausschnitt zeigt die Stelle** — ein Treffer ohne sie erklärt
	// nichts.
	if treffer[0].Ausschnitt == "" {
		t.Error("der Treffer trägt keinen Ausschnitt")
	}
}

// **Und semantisch.**
func TestDerIndexFindetSemantisch(t *testing.T) {
	index := mitIndex(t,
		einIndexdokument("erster", "irgendein Text"),
		einIndexdokument("zweiter", "ein anderer Text"))

	// **Der zweite Vektor ist (2,0,0)** — dieselbe Richtung wie die Frage,
	// und bei Cosinus zählt allein die Richtung.
	treffer, fehler := index.Semantisch(probekontext(t), []float32{1, 0, 0}, 10)
	if fehler != nil {
		t.Fatalf("Semantisch: %v", fehler)
	}

	if len(treffer) != 2 {
		t.Fatalf("die Treffer sind %+v", treffer)
	}

	// **Beide zeigen in dieselbe Richtung** — der Abstand ist gleich, und die
	// Reihenfolge entscheidet der stabile Tie-Break über die Kennung.
	if treffer[0].Dokument.Kennung != "erster" {
		t.Fatalf("der erste Treffer ist %q", treffer[0].Dokument.Kennung)
	}
}

// **Ein fremdes Profil bekommt nichts — auch nicht die Zahl.**
//
// # Warum die Zahl mitzählt
//
// **Weil sie eine Auskunft ist.** Das Issue: „Fremde Treffer dürfen auch
// nicht in Anzahl, Cache, Debugdaten oder Vorschlägen auftauchen." Wer
// „3 Treffer, 1 davon darfst du sehen" sagt, hat zwei verraten.
func TestEinFremdesProfilBekommtNichts(t *testing.T) {
	index := mitIndex(t,
		einIndexdokument("oeffentlich", "Lightning Bolt deals damage."),
		einEigenesDokument("meins", "Lightning schlägt in mein Set ein.", 1),
		einEigenesDokument("fremd", "Lightning trifft ein fremdes Set.", 99))

	for name, suchen := range map[string]func() ([]domain.Retrievaltreffer, error){
		"lexikalisch": func() ([]domain.Retrievaltreffer, error) {
			return index.Lexikalisch(probekontext(t), "lightning", 10)
		},
		"semantisch": func() ([]domain.Retrievaltreffer, error) {
			return index.Semantisch(probekontext(t), []float32{1, 0, 0}, 10)
		},
	} {
		t.Run(name, func(t *testing.T) {
			treffer, fehler := suchen()
			if fehler != nil {
				t.Fatalf("suchen: %v", fehler)
			}

			if len(treffer) != 2 {
				t.Fatalf("%d Treffer, erwartet zwei: %+v", len(treffer), treffer)
			}

			for _, eines := range treffer {
				if eines.Dokument.Kennung == "fremd" {
					t.Fatalf("ein fremdes Dokument kam durch: %+v", eines)
				}
			}
		})
	}
}

// **Ein Neubau wird erst beim Aktivieren sichtbar.**
//
// **Das ist die Zusage aus dem Issue**: „Neuer Index entsteht separat und
// wird erst nach Prüfung atomar aktiviert; Leser halten konsistente
// Generation."
func TestEinNeubauWirdErstBeimAktivierenSichtbar(t *testing.T) {
	index := mitIndex(t, einIndexdokument("alt", "der alte Text"))

	neu, fehler := index.Anlegen(probekontext(t), einIndexmanifest())
	if fehler != nil {
		t.Fatalf("Anlegen: %v", fehler)
	}

	if geschrieben := index.Schreiben(probekontext(t), neu,
		einIndexdokument("neu", "der neue Text"), []float32{1, 0, 0}); geschrieben != nil {
		t.Fatalf("Schreiben: %v", geschrieben)
	}

	// **Noch sieht niemand den neuen Stand.**
	treffer, fehler := index.Lexikalisch(probekontext(t), "text", 10)
	if fehler != nil {
		t.Fatalf("Lexikalisch: %v", fehler)
	}

	if len(treffer) != 1 || treffer[0].Dokument.Kennung != "alt" {
		t.Fatalf("während des Baus sichtbar: %+v", treffer)
	}

	if aktiviert := index.Aktivieren(probekontext(t), neu); aktiviert != nil {
		t.Fatalf("Aktivieren: %v", aktiviert)
	}

	// **Und danach nur den neuen.**
	treffer, fehler = index.Lexikalisch(probekontext(t), "text", 10)
	if fehler != nil {
		t.Fatalf("Lexikalisch: %v", fehler)
	}

	if len(treffer) != 1 || treffer[0].Dokument.Kennung != "neu" {
		t.Fatalf("nach dem Umschalten sichtbar: %+v", treffer)
	}
}

// **Eine Sperre schlägt den letzten guten Index.**
//
// **Sofort und ohne Neubau** — das Issue: „aktuelle Autorisierung und
// Tombstones sperren alte Inhalte sofort, auch während Rebuild oder nach
// Restore."
func TestEineSperreSchlaegtDenIndex(t *testing.T) {
	index := mitIndex(t,
		einIndexdokument("bleibt", "Lightning bleibt."),
		einIndexdokument("geht", "Lightning geht."))

	if fehler := index.Sperren(probekontext(t), "geht", 0); fehler != nil {
		t.Fatalf("Sperren: %v", fehler)
	}

	treffer, fehler := index.Lexikalisch(probekontext(t), "lightning", 10)
	if fehler != nil {
		t.Fatalf("Lexikalisch: %v", fehler)
	}

	if len(treffer) != 1 || treffer[0].Dokument.Kennung != "bleibt" {
		t.Fatalf("die Treffer sind %+v", treffer)
	}

	// **Und sie überlebt einen Neubau.** Wer nur löschte, hätte den Text
	// beim nächsten Import wieder.
	neu, fehler := index.Anlegen(probekontext(t), einIndexmanifest())
	if fehler != nil {
		t.Fatalf("Anlegen: %v", fehler)
	}

	if geschrieben := index.Schreiben(probekontext(t), neu,
		einIndexdokument("geht", "Lightning geht."), []float32{1, 0, 0}); geschrieben != nil {
		t.Fatalf("Schreiben: %v", geschrieben)
	}

	if aktiviert := index.Aktivieren(probekontext(t), neu); aktiviert != nil {
		t.Fatalf("Aktivieren: %v", aktiviert)
	}

	treffer, fehler = index.Lexikalisch(probekontext(t), "lightning", 10)
	if fehler != nil {
		t.Fatalf("Lexikalisch: %v", fehler)
	}

	if len(treffer) != 0 {
		t.Fatalf("der gesperrte Text kam nach dem Neubau wieder: %+v", treffer)
	}
}

// **Ohne aktiven Index gibt es keine Treffer und keinen Fehler.**
//
// **Das ist eine Auskunft und kein Mangel**: wer nie gebaut hat, findet
// nichts, und das Fenster sagt es als leeren Zustand (§6).
func TestOhneAktivenIndexGibtEsKeineTreffer(t *testing.T) {
	index := storage.NeuerRetrievalindex(frischeDatenbank(t))

	treffer, fehler := index.Lexikalisch(probekontext(t), "irgendwas", 10)
	if fehler != nil {
		t.Fatalf("Lexikalisch: %v", fehler)
	}

	if len(treffer) != 0 {
		t.Fatalf("die Treffer sind %+v", treffer)
	}

	if _, fehler := index.Manifest(probekontext(t)); !errors.Is(
		fehler, domain.ErrNotFound) {
		t.Fatalf("das Manifest kam ohne Index: %v", fehler)
	}
}

// **Das Manifest des aktiven Index kommt heraus, wie es hineinging.**
func TestDasManifestKommtHerausWieEsHineinging(t *testing.T) {
	index := mitIndex(t, einIndexdokument("eines", "ein Text"))

	gelesen, fehler := index.Manifest(probekontext(t))
	if fehler != nil {
		t.Fatalf("Manifest: %v", fehler)
	}

	if !gelesen.PasstZu(einIndexmanifest()) {
		t.Fatalf("das Manifest ist %+v", gelesen)
	}

	if gelesen.Fingerabdruck() != einIndexmanifest().Fingerabdruck() {
		t.Fatalf("der Fingerabdruck ist %q", gelesen.Fingerabdruck())
	}
}

// **Ein unvollständiges Dokument kommt nicht in den Index.**
func TestEinUnvollstaendigesDokumentKommtNichtHinein(t *testing.T) {
	index := storage.NeuerRetrievalindex(frischeDatenbank(t))

	generation, fehler := index.Anlegen(probekontext(t), einIndexmanifest())
	if fehler != nil {
		t.Fatalf("Anlegen: %v", fehler)
	}

	kaputt := einIndexdokument("ohne", "")

	if fehler := index.Schreiben(probekontext(t), generation, kaputt,
		[]float32{1, 0, 0}); !errors.Is(fehler, domain.ErrValidation) {
		t.Fatalf("das leere Dokument kam durch: %v", fehler)
	}
}

// **Und ein Vektor der falschen Länge auch nicht.**
//
// **Das ist die Zusage aus dem Manifest, eine Ebene tiefer**: die Dimension
// steht dort, und ein Vektor, der sie nicht hat, gehört in einen anderen
// Raum.
func TestEinVektorDerFalschenLaengeKommtNichtHinein(t *testing.T) {
	index := storage.NeuerRetrievalindex(frischeDatenbank(t))

	generation, fehler := index.Anlegen(probekontext(t), einIndexmanifest())
	if fehler != nil {
		t.Fatalf("Anlegen: %v", fehler)
	}

	if fehler := index.Schreiben(probekontext(t), generation,
		einIndexdokument("eines", "ein Text"),
		[]float32{1, 0}); !errors.Is(fehler, domain.ErrValidation) {
		t.Fatalf("der kurze Vektor kam durch: %v", fehler)
	}
}

// **Die Merkmale überstehen den Weg durch den Bestand.**
//
// **Alle sechs, und einzeln geprüft.** Ein Feld, das unterwegs verloren geht,
// macht jeden Filter darüber still wirkungslos — die Suche bleibt grün und
// findet nichts (§10, „Ein Wächter, der die halbe Zusage hält").
func TestDieMerkmaleUeberstehenDenBestand(t *testing.T) {
	dokument := einIndexdokument("bolt", "Lightning Bolt deals 3 damage.")
	manawert := 1.0
	dokument.Merkmale = domain.Retrievalmerkmale{
		Farben: []string{"R"}, Farbidentitaet: []string{"R"},
		Typen: []string{"instant"}, Manawert: &manawert, Set: "lea",
		Formate:     map[string]string{"modern": "legal"},
		Formatstand: "2026-09-12",
	}

	treffer, fehler := mitIndex(t, dokument).Lexikalisch(
		probekontext(t), "lightning", 10)
	if fehler != nil {
		t.Fatalf("Lexikalisch: %v", fehler)
	}

	if len(treffer) != 1 {
		t.Fatalf("die Treffer sind %+v", treffer)
	}

	zurueck := treffer[0].Dokument.Merkmale

	if !slices.Equal(zurueck.Farben, []string{"R"}) ||
		!slices.Equal(zurueck.Farbidentitaet, []string{"R"}) ||
		!slices.Equal(zurueck.Typen, []string{"instant"}) ||
		zurueck.Manawert == nil || *zurueck.Manawert != 1 ||
		zurueck.Set != "lea" || zurueck.Formatstand != "2026-09-12" {
		t.Fatalf("zurück kam %+v", zurueck)
	}

	lage, stand, bekannt := zurueck.Format("modern")
	if !bekannt || lage != "legal" || stand != "2026-09-12" {
		t.Fatalf("%q, %q, %v", lage, stand, bekannt)
	}

	// **Und ein Dokument ohne Merkmale bekommt keine.**
	ohne, fehler := mitIndex(t, einIndexdokument("shock", "Shock deals 2.")).
		Lexikalisch(probekontext(t), "shock", 10)
	if fehler != nil {
		t.Fatalf("Lexikalisch: %v", fehler)
	}

	if len(ohne) != 1 || !ohne[0].Dokument.Merkmale.Leer() {
		t.Fatalf("zurück kam %+v", ohne)
	}
}
