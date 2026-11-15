//go:build netzlauf || dienstdrift

// **Der Lauf gegen den echten Dienst** —
// Abnahmekriterium aus #888:
// „Ein Netzlauf hinter einem Build-Tag, wie `fremdpakete` — er läuft nicht in
// der Pipeline und zählt nicht zur Abdeckung."
//
// # Was er prüft, das kein anderer Test kann
//
// Die Vertragstests halten, dass dieses Programm die **dokumentierte** Gestalt
// liest. Sie können nicht sagen, ob Scryfall sie noch schickt — genau das ist
// die Drift, gegen die sie gebaut sind, und gegen sich selbst prüft ein
// Beispiel nicht.
//
// # Warum er nicht in der Pipeline läuft
//
// Drei Gründe, und jeder allein genügt:
//
//   - **Er befragt einen fremden Rechner.** Ein Testlauf, der das tut, ist
//     keiner (§3) — er scheitert, wenn jemand anderes ein Problem hat.
//   - **Er verbraucht ein Kontingent**, das anderen zusteht.
//   - **Er ist nicht wiederholbar**: die Antworten ändern sich.
//
// # Wie er gefahren wird
//
//	go test -tags netzlauf ./internal/scryfall/ -run TestGegenDenEchtenDienst -v
//
// **Und wann**: vor jeder Freigabe, und wenn ein Vertragstest ohne erkennbaren
// Grund bricht.
package scryfall_test

import (
	"strings"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/netz"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/scryfall"
)

// echterKlient baut einen Klienten gegen `api.scryfall.com`.
//
// **Mit einem eigenen Zwischenspeicher je Lauf**, damit der Lauf wirklich
// fragt: ein geteilter Speicher machte den zweiten Lauf zu einer Prüfung des
// Speichers.
func echterKlient(t *testing.T) *scryfall.Klient {
	t.Helper()

	return scryfall.NeuerKlient(scryfall.Adresse,
		netz.NeuerKlient(netz.UeblicheFristen()),
		netz.NeuerSpeicher(t.TempDir(), scryfall.Speicherfrist, time.Now),
		time.Now, "netzlauf")
}

// **Vier Wege, und jeder prüft seine Gestalt gegen den echten Dienst.**
//
// Geprüft wird **die Gestalt, nicht der Inhalt**: dass `Lightning Bolt` drei
// Schaden macht, ist Sache von Wizards und keine Zusage dieses Programms. Dass
// die Antwort ein `comment` trägt, ist eine.
func TestGegenDenEchtenDienst(t *testing.T) {
	klient := echterKlient(t)

	t.Run("eine Karte", func(t *testing.T) {
		karte, _, befunde, fehler := klient.KarteNachName(t.Context(), "Lightning Bolt")
		if fehler != nil {
			t.Fatalf("KarteNachName: %v", fehler)
		}

		if karte.Name == "" || karte.Adresse == "" {
			t.Errorf("Karte = %+v", karte)
		}

		// **Ein Befund ist kein Fehler, aber er gehört gesehen** — er heißt,
		// dass Scryfall einen Wert schickt, den dieses Programm nicht kennt.
		for _, befund := range befunde {
			t.Logf("Befund: %s", befund.Meldung)
		}

		if karte.Bilder.Normal == "" {
			t.Error("die Karte nennt kein normales Bild — der Bildweg bräche")
		}

		if karte.Kuenstler == "" {
			t.Error("die Karte nennt keinen Künstler — der art_crop würde abgewiesen")
		}
	})

	// **Ein bestimmter Druck über `/cards/:id`** (#1167).
	//
	// # Was er prüft, das der Lauf darüber nicht kann
	//
	// `/cards/named` gibt **irgendeinen** Druck; welchen, entscheidet
	// Scryfall. Der Verweis verlangt das Gegenteil: die Antwort muss den
	// erfragten Druck nennen, sonst weist der Klient sie ab. Ob Scryfall die
	// Kennung wirklich kanonisch klein zurückgibt, sagt nur dieser Lauf —
	// alles andere wäre eine Erinnerung, kein Prüfergebnis (§10).
	t.Run("ein bestimmter Druck", func(t *testing.T) {
		gefunden, _, _, fehler := klient.KarteNachName(t.Context(), "Lightning Bolt")
		if fehler != nil {
			t.Fatalf("KarteNachName: %v", fehler)
		}

		// **In Großschreibung erfragt.** Genau das ist die gemessene
		// Eigenheit: Scryfall nimmt sie an und antwortet klein.
		druck, _, befunde, druckfehler := klient.KarteNachKennung(
			t.Context(), strings.ToUpper(gefunden.Kennung))
		if druckfehler != nil {
			t.Fatalf("KarteNachKennung: %v", druckfehler)
		}

		for _, befund := range befunde {
			t.Logf("Befund: %s", befund.Meldung)
		}

		if druck.Kennung != gefunden.Kennung {
			t.Errorf("gefragt war %q, geantwortet hat %q", gefunden.Kennung, druck.Kennung)
		}

		if druck.Set == "" || druck.Nummer == "" {
			t.Errorf("der Druck nennt Set %q und Nummer %q", druck.Set, druck.Nummer)
		}
	})

	t.Run("die Regelhinweise", func(t *testing.T) {
		karte, _, _, fehler := klient.KarteNachName(t.Context(), "Derevi, Empyrial Tactician")
		if fehler != nil {
			t.Fatalf("KarteNachName: %v", fehler)
		}

		hinweise, _, befunde, rulingfehler := klient.Rulings(t.Context(), karte.Kennung)
		if rulingfehler != nil {
			t.Fatalf("Rulings: %v", rulingfehler)
		}

		if len(hinweise) == 0 {
			t.Fatal("keine Hinweise — die Dokumentation sagt, diese Karte hat welche")
		}

		for _, befund := range befunde {
			t.Logf("Befund: %s", befund.Meldung)
		}

		if hinweise[0].Text == "" || hinweise[0].Datum == "" {
			t.Errorf("erster Hinweis = %+v", hinweise[0])
		}
	})

	t.Run("eine Setangabe", func(t *testing.T) {
		satz, _, befunde, fehler := klient.Setangaben(t.Context(), "lea")
		if fehler != nil {
			t.Fatalf("Setangaben: %v", fehler)
		}

		for _, befund := range befunde {
			t.Logf("Befund: %s", befund.Meldung)
		}

		if satz.Code != "lea" || satz.Name == "" || satz.Karten == 0 {
			t.Errorf("Satz = %+v", satz)
		}
	})

	t.Run("die Symbologie", func(t *testing.T) {
		symbole, _, befunde, fehler := klient.Symbologie(t.Context())
		if fehler != nil {
			t.Fatalf("Symbologie: %v", fehler)
		}

		for _, befund := range befunde {
			t.Logf("Befund: %s", befund.Meldung)
		}

		if len(symbole) < 50 {
			t.Errorf("%d Symbole — Scryfall führt weit mehr", len(symbole))
		}

		// **`{W}` muss dabei sein**, und es muss den Manawert 1 tragen. Wer
		// hier nur die Zahl der Symbole prüfte, sähe eine Umbenennung nicht.
		gefunden := false

		for _, eines := range symbole {
			if eines.Zeichen != "{W}" {
				continue
			}

			gefunden = true

			if eines.Manawert == nil || *eines.Manawert != 1 || !eines.InManakosten {
				t.Errorf("{W} = %+v", eines)
			}
		}

		if !gefunden {
			t.Error("{W} fehlt in der Symbologie")
		}
	})
}
