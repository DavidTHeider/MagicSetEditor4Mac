package service_test

// **Eine Übersetzung von fremder Hand** — QA-06b
// (#1967),
// Fläche 30 aus `docs/FUZZZIELE.md`: [service.Sprachpakete.Einlesen].
//
// # Warum hier und nicht in `domain`
//
// **Die Datei wird im Dienst gelesen** — Größe, Gestalt, Format, dann die
// Angaben über [domain.Sprachpaket.Pruefe] und jeder Satz gegen die Vorlage.
// Ein Ziel über `domain` allein sähe nur die zweite Hälfte.
//
// # Was ein Fund ist — festgelegt **vor** dem ersten Lauf
//
//   - Panic
//   - zweimal dieselbe Datei, zwei Urteile oder zwei Pakete
//   - **abgewiesen ohne Feld oder ohne Schlüssel** — das Fenster zeigt einen
//     Befund an seinem Feld (§6, §8), und eine rohe Meldung von
//     `encoding/json` erreichte es sonst ungefiltert
//   - **abgewiesen und trotzdem geschrieben** — kein halber Stand
//   - angenommen, und das gespeicherte Paket besteht seine eigene Prüfung
//     nicht, oder ein Satz besteht die seines Pfades nicht
//   - **Einlesen∘Ausgeben ist nicht die Identität**: was eingelesen und mit
//     Lizenz wieder ausgegeben wird, kommt beim nächsten Einlesen als
//     dasselbe Paket an
//   - ein vorhandenes Paket, das ohne „ersetzen" still überschrieben wird
//
//	go test ./internal/service -run Fuzz -fuzz FuzzSprachpaketEinlesen -fuzztime 2m

import (
	"errors"
	"reflect"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Was eingelesen ist, besteht seine Prüfung und kommt ausgegeben wieder
// so an.**
func FuzzSprachpaketEinlesen(f *testing.F) {
	ausgegeben := ausgegebeneUebersetzung(f)

	for _, saat := range []string{
		string(ausgegeben),
		`{"format":"mse4mac/sprachpaket/1","kennung":"da","name":"Dansk","basis":"en","stand":"entwurf",` +
			`"lizenz":"CC0-1.0","texte":{"allgemein.titel":{"text":"Sæt","quelle":"Sets"},` +
			`"neu.aus.spaeter":{"text":"**fed**","quelle":""}}}`,
		`{"format":"mse4mac/sprachpaket/2","kennung":"da"}`,
		`{"format":"mse4mac/sprachpaket/1","kennung":"de","name":"Deutsch","basis":"en","stand":"entwurf"}`,
		`{"format":"mse4mac/sprachpaket/1","kennung":"da","name":"Dansk","basis":"en","stand":"entwurf",` +
			`"texte":{"fehler.symbol":{"text":"Filen er over {andet}.","quelle":"x"}}}`,
		`{"format":"mse4mac/sprachpaket/1","unbekannt":1}`,
		`{}`, `[]`, ``,
	} {
		f.Add([]byte(saat), false)
	}

	f.Fuzz(func(t *testing.T, datei []byte, vorhanden bool) {
		ablage := neuesSprachpaketdoppel()
		dienst := sprachdienst(ablage)

		// **Ein Paket, das schon da ist** — dann darf ohne „ersetzen"
		// nichts darüberschreiben.
		if vorhanden {
			fertigesPaket(t, dienst)
		}

		vorher := ablage.geschrieben
		altesPaket, _ := ablage.Sprachpaket(t.Context(), "pt-BR") //nolint:errcheck // fehlt, wenn nicht vorhanden

		auskunft, fehler := dienst.Einlesen(t.Context(), datei, false)

		zweite := neuesSprachpaketdoppel()
		if vorhanden {
			fertigesPaket(t, sprachdienst(zweite))
		}

		_, zweiterFehler := sprachdienst(zweite).Einlesen(t.Context(), datei, false)

		if (fehler == nil) != (zweiterFehler == nil) || (fehler != nil && fehler.Error() != zweiterFehler.Error()) ||
			!reflect.DeepEqual(ablage.pakete, zweite.pakete) {
			t.Fatalf("dieselbe Datei, zwei Urteile: %v / %v", fehler, zweiterFehler)
		}

		if fehler != nil {
			spracheAbgewiesen(t, ablage, fehler, vorher, altesPaket)

			return
		}

		eingelesen := ablage.pakete[auskunft.Paket.Kennung]
		spracheAngenommen(t, eingelesen)
		spracheZurueckGelesen(t, dienst, eingelesen)
	})
}

// ausgegebeneUebersetzung ist eine Datei, wie der Dienst sie ausgibt.
func ausgegebeneUebersetzung(f *testing.F) []byte {
	f.Helper()

	dienst := sprachdienst(neuesSprachpaketdoppel())
	if _, fehler := dienst.Anlegen(f.Context(), "pt-BR", "Português do Brasil", "en"); fehler != nil {
		f.Fatal(fehler)
	}

	lizenz := "CC-BY-4.0"
	if _, fehler := dienst.Aendern(f.Context(), "pt-BR", service.Sprachpaketangaben{Lizenz: &lizenz},
		map[string]string{"fehler.symbol": "O arquivo passa de {grenze}."}); fehler != nil {
		f.Fatal(fehler)
	}

	_, inhalt, fehler := dienst.Ausgeben(f.Context(), "pt-BR")
	if fehler != nil {
		f.Fatal(fehler)
	}

	return inhalt
}

// spracheAbgewiesen hält die Zusagen über eine Ablehnung.
func spracheAbgewiesen(t *testing.T, ablage *sprachpaketdoppel, fehler error, vorher int, altesPaket domain.Sprachpaket) {
	t.Helper()

	var befund *domain.FieldError
	if !errors.As(fehler, &befund) || befund.Field == "" || befund.Key == "" {
		t.Errorf("abgewiesen ohne Feld oder Schlüssel: %v", fehler)
	}

	if ablage.geschrieben != vorher {
		t.Errorf("abgewiesen und trotzdem %d-mal geschrieben", ablage.geschrieben-vorher)
	}

	if heute, da := ablage.pakete["pt-BR"]; da && !reflect.DeepEqual(heute, altesPaket) {
		t.Error("ein vorhandenes Paket wurde ohne „ersetzen“ überschrieben")
	}
}

// spracheAngenommen hält das gespeicherte Paket gegen seine Prüfungen.
func spracheAngenommen(t *testing.T, paket domain.Sprachpaket) {
	t.Helper()

	if fehler := paket.Pruefe(); fehler != nil {
		t.Errorf("angenommen und ungültig: %v", fehler)
	}

	for pfad, text := range paket.Texte {
		pruefung := domain.PruefeFremdeUebersetzung(pfad, text.Text)
		if sprachvorlage.Quelle(pfad) != "" {
			pruefung = sprachvorlage.PruefeUebersetzung(pfad, text.Text)
		}

		if pruefung != nil {
			t.Errorf("%q angenommen: %v", pfad, pruefung)
		}
	}
}

// spracheZurueckGelesen gibt das Paket mit Lizenz aus und liest es wieder ein.
func spracheZurueckGelesen(t *testing.T, dienst *service.Sprachpakete, paket domain.Sprachpaket) {
	t.Helper()

	if paket.Lizenz == "" {
		return
	}

	_, inhalt, fehler := dienst.Ausgeben(t.Context(), paket.Kennung)
	if fehler != nil {
		t.Fatalf("angenommen und nicht wieder auszugeben: %v", fehler)
	}

	ablage := neuesSprachpaketdoppel()
	if _, fehler := sprachdienst(ablage).Einlesen(t.Context(), inhalt, false); fehler != nil {
		t.Fatalf("ausgegeben und nicht wieder einzulesen: %v", fehler)
	}

	zurueck := ablage.pakete[paket.Kennung]
	zurueck.Angelegt, zurueck.Geaendert = paket.Angelegt, paket.Geaendert

	if !reflect.DeepEqual(zurueck, paket) {
		t.Errorf("Einlesen∘Ausgeben ist nicht die Identität:\n%+v\n%+v", paket, zurueck)
	}
}
