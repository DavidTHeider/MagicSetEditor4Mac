package app

// **Die Partienliste sagt, ob man dran ist** — UXR-F16
// (#2031).
//
// # Wogegen hier geprüft wird
//
// **Über einen echten Bestand und nicht über ein Doppel** (§3): ein Doppel
// der Leitung gäbe die Felder zurück, die der Lauf hineinschreibt, und prüfte
// damit nur die Abbildung. Hier läuft der Weg ganz — Entwurf mit gebundenem
// Sitz, Aufbau, Speicher, Dienst, Zeile, DTO. Was die Regel im Einzelnen
// entscheidet, prüft `internal/service/partiendran_test.go`; hier steht, dass
// **das gezeigte Profil** bei ihr ankommt.

import (
	"testing"
)

// **Der Sitz, den das gezeigte Profil belegt, ist der eigene — und er ist
// dran.**
//
// **Vom Aufbau an**: der Beginn steht aus (CR 103.5, Fassung 2026-09-25), und
// jeder Sitz hat die Starthand noch vor sich. Das Profil 1 ist das gezeigte,
// das die Anwendung in [eineSpielende] führt.
func TestDieListeNenntDenEigenenSitzUndObManDranIst(t *testing.T) {
	t.Parallel()

	gebunden, deckID := eineSpielende(t)

	entwurf := einEntwurf(deckID)
	entwurf.Sitze[1].ProfilID = "1"

	if angelegt := gebunden.PartieAnlegen(t.Context(), entwurf); angelegt.Fehler != nil {
		t.Fatalf("PartieAnlegen: %+v", angelegt.Fehler)
	}

	liste := gebunden.Partien(t.Context())
	if liste.Fehler != nil || len(liste.Partien) != 1 {
		t.Fatalf("Partien: %+v", liste)
	}

	if zeile := liste.Partien[0]; zeile.EigenerSitz != "sitz-2" || !zeile.Dran {
		t.Errorf("eigener Sitz %q, dran=%v", zeile.EigenerSitz, zeile.Dran)
	}
}

// **Ohne Bindung sitzt das gezeigte Profil am Sitz des Wirts** — UXR-F21
// (#2041),
// Vermerk 0164. Bis dahin stand hier „keiner ist gebunden" als ein Weg, auf
// dem niemand dran ist — und das war der Normalfall jedes Menschen mit einem
// Profil: die Marke erschien nie.
//
// **Der Wirt ist der erste Sitz**, wenn der Entwurf keinen nennt.
func TestOhneBindungSitztDasGezeigteProfilAmWirt(t *testing.T) {
	t.Parallel()

	gebunden, deckID := eineSpielende(t)

	if angelegt := gebunden.PartieAnlegen(t.Context(), einEntwurf(deckID)); angelegt.Fehler != nil {
		t.Fatalf("PartieAnlegen: %+v", angelegt.Fehler)
	}

	liste := gebunden.Partien(t.Context())
	if liste.Fehler != nil || len(liste.Partien) != 1 {
		t.Fatalf("Partien: %+v", liste)
	}

	if zeile := liste.Partien[0]; zeile.EigenerSitz != "sitz-1" || !zeile.Dran {
		t.Errorf("eigener Sitz %q, dran=%v", zeile.EigenerSitz, zeile.Dran)
	}
}

// **Wo das Profil nicht sitzt, bleibt beides leer** — der Sitz gehört einem
// anderen Profil. Der Beginn steht aus, und es ist trotzdem keiner dran, der
// hier fragt; **und die Vorgabe greift nicht**, denn gebunden ist schon einer.
func TestWoDasProfilNichtSitztSagtDieListeNichtDran(t *testing.T) {
	t.Parallel()

	gebunden, deckID := eineSpielende(t)

	entwurf := einEntwurf(deckID)
	entwurf.Sitze[1].ProfilID = "2"

	if angelegt := gebunden.PartieAnlegen(t.Context(), entwurf); angelegt.Fehler != nil {
		t.Fatalf("PartieAnlegen: %+v", angelegt.Fehler)
	}

	liste := gebunden.Partien(t.Context())
	if liste.Fehler != nil || len(liste.Partien) != 1 {
		t.Fatalf("Partien: %+v", liste)
	}

	if zeile := liste.Partien[0]; zeile.EigenerSitz != "" || zeile.Dran {
		t.Errorf("eigener Sitz %q, dran=%v", zeile.EigenerSitz, zeile.Dran)
	}
}

// **Eine angehaltene Partie bleibt die eigene, aber wartet auf niemanden.**
//
// **Über die Bindung des Fensters**, wie der Betrieb sie ruft: wer pausiert,
// ändert den Stand im Bestand, und die nächste Liste liest ihn von dort.
func TestEineAngehaltenePartieIstNichtDran(t *testing.T) {
	t.Parallel()

	gebunden, deckID := eineSpielende(t)

	entwurf := einEntwurf(deckID)
	entwurf.Sitze[0].ProfilID = "1"

	angelegt := gebunden.PartieAnlegen(t.Context(), entwurf)
	if angelegt.Fehler != nil {
		t.Fatalf("PartieAnlegen: %+v", angelegt.Fehler)
	}

	if liste := gebunden.Partien(t.Context()); !liste.Partien[0].Dran {
		t.Fatalf("vor dem Anhalten war niemand dran: %+v", liste.Partien[0])
	}

	if angehalten := gebunden.PartiePausieren(t.Context(), angelegt.Partie.ID, "sitz-1"); angehalten.Fehler != nil {
		t.Fatalf("PartiePausieren: %+v", angehalten.Fehler)
	}

	liste := gebunden.Partien(t.Context())

	if zeile := liste.Partien[0]; zeile.EigenerSitz != "sitz-1" || zeile.Dran {
		t.Errorf("eigener Sitz %q, dran=%v", zeile.EigenerSitz, zeile.Dran)
	}
}
