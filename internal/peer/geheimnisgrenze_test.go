package peer_test

// **Was nach außen geht, trägt kein Geheimnis** — Werkstück N1 von Phase 5.4
// (#1454).
//
// # Warum das ein eigener Wächter ist
//
// **Weil die Abnahme es wörtlich verlangt**: „Alle Identitäts-/
// Handshake-Nachrichten und Exporte enthalten keine privaten Schlüssel,
// Kennwörter, Zugriffstoken oder lokale Profil-IDs."
//
// **Und weil ein Blick auf den Quelltext es nicht belegt.** Kein Feld heißt
// „privat", und trotzdem wäre ein Geheimnis nur ein falsch gesetztes Feld
// entfernt — die Identität **hält** eines, und sie unterschreibt damit. Dieser
// Lauf sieht die Bytes an, die wirklich hinausgehen.

import (
	"encoding/hex"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/peer"
)

// **Kein Schritt des Handschlags trägt den privaten Schlüssel.**
func TestKeinSchrittTraegtDenPrivatenSchluessel(t *testing.T) {
	t.Parallel()

	ich := eineIdentitaet(t)

	du, fehler := peer.Erzeugen(&festerZufall{stand: 99})
	if fehler != nil {
		t.Fatalf("Erzeugen: %v", fehler)
	}

	anbahnung, angebot, fehler := ich.Anbieten(peer.Wunsch{
		Zweck: domain.ZweckTeilen, Rolle: "gebend", Kanal: "kanal-1",
		Name: "Davids Rechner",
	}, &festerZufall{stand: 7})
	if fehler != nil {
		t.Fatalf("Anbieten: %v", fehler)
	}

	entgegennahme, annahme, fehler := du.Entgegennehmen(
		angebot, peer.Bereitschaft{Rolle: "nehmend", Kanal: "kanal-1"},
		&festerZufall{stand: 200}, eineListe(t))
	if fehler != nil {
		t.Fatalf("Entgegennehmen: %v", fehler)
	}

	meine, abschluss, fehler := anbahnung.Schliessen(annahme)
	if fehler != nil {
		t.Fatalf("Schliessen: %v", fehler)
	}

	if _, fehler := entgegennahme.Abschliessen(abschluss); fehler != nil {
		t.Fatalf("Abschliessen: %v", fehler)
	}

	// **Beide Geheimnisse, und beide in ihren zwei Gestalten** — als Hex,
	// wie [peer.Identitaet.Geheimnis] sie gibt, und als rohe Bytes, wie sie
	// im Speicher liegen. Wer nur die eine sucht, findet die andere nicht.
	hexformen := []string{ich.Geheimnis(), du.Geheimnis()}
	geheim := make([]string, 0, 2*len(hexformen))
	geheim = append(geheim, hexformen...)

	for _, hexform := range hexformen {
		roh, fehler := hex.DecodeString(hexform)
		if fehler != nil {
			t.Fatalf("Hex: %v", fehler)
		}

		geheim = append(geheim, string(roh))
	}

	for name, hinaus := range map[string]string{
		"Angebot":   string(angebot),
		"Annahme":   string(annahme),
		"Abschluss": string(abschluss),

		// **Und die Sitzung selbst** — sie reist durch `service` und `app`
		// und steht am Ende an einer Fläche.
		"Sitzung": meine.Kennung + meine.Fingerabdruck +
			string(meine.Gegenueber) + meine.SelbstgewaehlterName +
			meine.Kanal + meine.MeineRolle + meine.SeineRolle,
	} {
		for _, geheimnis := range geheim {
			if strings.Contains(hinaus, geheimnis) {
				t.Errorf("%s trägt ein Geheimnis", name)
			}
		}
	}
}

// **Und die Gegenprobe: der Wächter findet eines, wenn es da ist.**
//
// # Warum sie dazugehört
//
// **Weil ein Suchlauf, der nichts findet, zwei Lesarten hat** (§10): es ist
// nichts da — oder er sucht das Falsche. Hier wird das Geheimnis absichtlich
// in einen Rumpf gesetzt, und der Vergleich muss es sehen.
func TestDerWaechterFindetEinGeheimnisWennEsDaIst(t *testing.T) {
	t.Parallel()

	ich := eineIdentitaet(t)

	roh, fehler := ich.Verpacke(peer.Huelle{
		Protokoll: peer.ProtokollPaarung,
		Art:       "unsinn",
		ID:        "n-1",
		Sitzung:   "s-1",
		Absender:  string(ich.Kennung()),
		Rumpf:     []byte(`{"versehen":"` + ich.Geheimnis() + `"}`),
	})
	if fehler != nil {
		t.Fatalf("Verpacke: %v", fehler)
	}

	if !strings.Contains(string(roh), ich.Geheimnis()) {
		t.Fatal("der Vergleich sieht ein eingebautes Geheimnis nicht")
	}
}
