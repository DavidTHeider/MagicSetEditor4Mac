package service_test

import (
	"encoding/json"
	"errors"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// fertigesPaket legt ein Paket mit Lizenz und einem Satz an.
func fertigesPaket(t *testing.T, dienst *service.Sprachpakete) {
	t.Helper()

	if _, err := dienst.Anlegen(t.Context(), "pt-BR", "Português do Brasil", "en"); err != nil {
		t.Fatal(err)
	}

	lizenz, herkunft := "CC-BY-4.0", "Ana"

	if _, err := dienst.Aendern(t.Context(), "pt-BR",
		service.Sprachpaketangaben{Lizenz: &lizenz, Herkunft: &herkunft},
		map[string]string{"fehler.symbol": "O arquivo passa de {grenze}."}); err != nil {
		t.Fatal(err)
	}
}

// **Was hinausgeht: Herkunft, Lizenz und Stand — und sonst nichts** —
// LOC-02.AK-04 (#1688).
// Keine Einstellung, kein Profil, kein Geheimnis: die Datei hat genau diese
// Felder, und ein neues müsste hier eingetragen werden.
func TestEineAusgegebeneUebersetzungTraegtNurIhreAngaben(t *testing.T) {
	t.Parallel()

	ablage := neuesSprachpaketdoppel()
	dienst := sprachdienst(ablage)

	if _, err := dienst.Anlegen(t.Context(), "da", "Dansk", "en"); err != nil {
		t.Fatal(err)
	}

	// **Ohne Lizenz geht nichts hinaus** — wer es bekommt, wüsste nicht, was
	// er damit darf.
	_, _, err := dienst.Ausgeben(t.Context(), "da")
	feldbefund(t, err, "lizenz", domain.SchluesselSprachpaketOhneLizenz)

	fertigesPaket(t, dienst)

	name, inhalt, err := dienst.Ausgeben(t.Context(), "pt-BR")
	if err != nil || name != "pt-BR.sprachpaket.json" {
		t.Fatalf("ausgegeben: %q, %v", name, err)
	}

	var datei map[string]json.RawMessage
	if err := json.Unmarshal(inhalt, &datei); err != nil {
		t.Fatal(err)
	}

	if felder := slices.Sorted(maps.Keys(datei)); !slices.Equal(felder, []string{
		"ausgegeben", "basis", "format", "herkunft", "kennung", "lizenz", "name",
		"programm", "stand", "texte",
	}) {
		t.Errorf("die Datei trägt %v", felder)
	}

	for _, erwartet := range []string{
		`"format": "mse4mac/sprachpaket/1"`, `"programm": "MagicSetEditor4Mac 0.6.0"`,
		`"ausgegeben": "2026-09-29T23:00:00Z"`, `"lizenz": "CC-BY-4.0"`, `"stand": "entwurf"`,
		`"quelle": "Die Datei ist größer als {grenze}."`,
	} {
		if !strings.Contains(string(inhalt), erwartet) {
			t.Errorf("es fehlt %s in\n%s", erwartet, inhalt)
		}
	}

	if _, _, err := dienst.Ausgeben(t.Context(), "xx"); !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("ein fremdes Paket: %v", err)
	}
}

// **Eingelesen wird, was ausgegeben wurde — und nichts still ersetzt**
// (AK-03): ein Paket, das es schon gibt, braucht die ausdrückliche Wahl.
func TestEineUebersetzungKommtWiederHerein(t *testing.T) {
	t.Parallel()

	quelle := sprachdienst(neuesSprachpaketdoppel())
	fertigesPaket(t, quelle)

	_, inhalt, err := quelle.Ausgeben(t.Context(), "pt-BR")
	if err != nil {
		t.Fatal(err)
	}

	ablage := neuesSprachpaketdoppel()
	ziel := sprachdienst(ablage)

	auskunft, err := ziel.Einlesen(t.Context(), inhalt, false)
	if err != nil || auskunft.Paket.Herkunft != "Ana" || auskunft.Abgleich.Uebersetzt != 1 ||
		auskunft.Paket.Texte["fehler.symbol"].Text != "O arquivo passa de {grenze}." {
		t.Fatalf("eingelesen: %+v, %v", auskunft, err)
	}

	_, err = ziel.Einlesen(t.Context(), inhalt, false)
	feldbefund(t, err, "kennung", domain.SchluesselDoppelt)

	if _, err := ziel.Einlesen(t.Context(), inhalt, true); err != nil || ablage.geschrieben != 2 {
		t.Errorf("ersetzt: %v, %d Schreibvorgänge", err, ablage.geschrieben)
	}
}

// **Eine Datei, die nicht passt, wird abgewiesen, bevor sie etwas ändert.**
func TestEineFremdeUebersetzungsdateiWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	gut := func(aendere func(map[string]any)) []byte {
		datei := map[string]any{
			"format": domain.Sprachpaketformat, "kennung": "pt-BR", "name": "Português",
			"basis": "en", "stand": "entwurf", "lizenz": "", "herkunft": "", "programm": "x",
			"ausgegeben": "2026-09-29T23:00:00Z",
			"texte": map[string]any{
				"allgemein.titel": map[string]string{"text": "Conjuntos", "quelle": "Sets"},
			},
		}
		aendere(datei)

		roh, _ := json.Marshal(datei) //nolint:errcheck // Zeichenketten und Karten

		return roh
	}

	faelle := []struct {
		name       string
		inhalt     []byte
		feld       string
		schluessel domain.Schluessel
	}{
		{"kein JSON", []byte("kein JSON"), "datei", domain.SchluesselValidierung},
		{"zu groß", []byte(strings.Repeat(" ", 4<<20+1)), "datei", domain.SchluesselValidierung},
		{
			"ein anderes Format", gut(func(d map[string]any) { d["format"] = "mse4mac/umschlag/1" }),
			"datei", domain.SchluesselValidierung,
		},
		{
			"aus einem neueren Programm", gut(func(d map[string]any) { d["format"] = "mse4mac/sprachpaket/2" }),
			"datei", domain.SchluesselSprachpaketNeuer,
		},
		{
			"ein fremdes Feld", gut(func(d map[string]any) { d["passwort"] = "geheim" }),
			"datei", domain.SchluesselValidierung,
		},
		{
			"eine falsche Kennung", gut(func(d map[string]any) { d["kennung"] = "pt_BR" }),
			"kennung", domain.SchluesselSprachkennungUngueltig,
		},
		{"ein Tag im Satz", gut(func(d map[string]any) {
			d["texte"] = map[string]any{"allgemein.titel": map[string]string{"text": "<b>x</b>", "quelle": "Sets"}}
		}), "allgemein.titel", domain.SchluesselAuszeichnungUnsicher},
		{"ein Tag unter einem unbekannten Pfad", gut(func(d map[string]any) {
			d["texte"] = map[string]any{"neu.satz": map[string]string{"text": "<script>", "quelle": "Neu"}}
		}), "neu.satz", domain.SchluesselAuszeichnungUnsicher},
		// **Ein Satz ohne Pfad gehört zu keiner Stelle** — gefunden von der
		// vollen Kampagne (C02, #2227, `FuzzSprachpaketEinlesen/b257223c79af9260`):
		// der Befund trug den leeren Pfad als Feld, und kein Feld hieß so.
		{"ein Satz ohne Pfad", gut(func(d map[string]any) {
			d["texte"] = map[string]any{"": map[string]string{"text": "", "quelle": ""}}
		}), "texte", domain.SchluesselValidierung},
		{"ein Pfad aus Leerraum", gut(func(d map[string]any) {
			d["texte"] = map[string]any{" ": map[string]string{"text": "Novo", "quelle": "Neu"}}
		}), "texte", domain.SchluesselValidierung},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			ablage := neuesSprachpaketdoppel()

			_, err := sprachdienst(ablage).Einlesen(t.Context(), fall.inhalt, false)
			feldbefund(t, err, fall.feld, fall.schluessel)

			if ablage.geschrieben != 0 {
				t.Error("geschrieben wurde trotzdem")
			}
		})
	}

	// **Ein Pfad, den diese Fassung nicht kennt, bleibt** — er stammt aus
	// einer neueren, und wer das Paket weitergibt, verliert ihn nicht.
	auskunft, err := sprachdienst(neuesSprachpaketdoppel()).Einlesen(t.Context(), gut(func(d map[string]any) {
		d["texte"] = map[string]any{"neu.satz": map[string]string{"text": "Novo", "quelle": "Neu"}}
	}), false)
	if err != nil || !slices.Equal(auskunft.Abgleich.Unbekannt, []string{"neu.satz"}) {
		t.Errorf("unbekannt: %+v, %v", auskunft.Abgleich, err)
	}
}

// **Ein Fehler des Bestands kommt an** — beim Nachsehen, beim Anlegen und
// beim Ersetzen, und ein Satz, der nicht geschrieben wurde, gilt nicht.
func TestDerBestandSchweigtBeiEinerUebersetzung(t *testing.T) {
	t.Parallel()

	quelle := sprachdienst(neuesSprachpaketdoppel())
	fertigesPaket(t, quelle)

	_, inhalt, err := quelle.Ausgeben(t.Context(), "pt-BR")
	if err != nil {
		t.Fatal(err)
	}

	schweigt := errors.New("der Bestand schweigt")

	ablage := neuesSprachpaketdoppel()
	ablage.fehler = schweigt

	if _, err := sprachdienst(ablage).Einlesen(t.Context(), inhalt, false); !errors.Is(err, schweigt) {
		t.Errorf("nachsehen: %v", err)
	}

	ablage = neuesSprachpaketdoppel()
	ablage.schreibfehler = schweigt

	if _, err := sprachdienst(ablage).Einlesen(t.Context(), inhalt, false); !errors.Is(err, schweigt) {
		t.Errorf("anlegen: %v", err)
	}

	ablage = neuesSprachpaketdoppel()
	dienst := sprachdienst(ablage)
	fertigesPaket(t, dienst)
	ablage.schreibfehler = schweigt

	if _, err := dienst.Einlesen(t.Context(), inhalt, true); !errors.Is(err, schweigt) {
		t.Errorf("ersetzen: %v", err)
	}

	if _, err := dienst.Aendern(t.Context(), "pt-BR", service.Sprachpaketangaben{},
		map[string]string{"allgemein.titel": "Conjuntos"}); !errors.Is(err, schweigt) {
		t.Errorf("ändern: %v", err)
	}
}

// **Ausgegeben an einen Ort und von dort eingelesen** — die Wege der
// Kommandozeile und der Dateidialoge. Wer nur lesen darf, schreibt keine
// Datei (Vermerk 0078).
func TestEineUebersetzungGehtAufDiePlatteUndKommtWieder(t *testing.T) {
	t.Parallel()

	quelle := sprachdienst(neuesSprachpaketdoppel())
	fertigesPaket(t, quelle)

	ziel := filepath.Join(t.TempDir(), "pt.json")

	name, err := quelle.AusgebenNach(t.Context(), "pt-BR", ziel)
	if err != nil || name != "pt-BR.sprachpaket.json" {
		t.Fatalf("ausgegeben: %q, %v", name, err)
	}

	auskunft, err := sprachdienst(neuesSprachpaketdoppel()).EinlesenAus(t.Context(), ziel, false)
	if err != nil || auskunft.Paket.Kennung != "pt-BR" {
		t.Fatalf("eingelesen: %+v, %v", auskunft, err)
	}

	if _, err := quelle.AusgebenNach(t.Context(), "xx", ziel); !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("ein fremdes Paket: %v", err)
	}

	if _, err := quelle.EinlesenAus(t.Context(), filepath.Join(t.TempDir(), "fehlt.json"), false); !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("eine fehlende Datei: %v", err)
	}

	leser := sprachdienst(neuesSprachpaketdoppel()).MitPruefer(verweigerndeWache{})
	fertigesPaket(t, leser)

	andere := filepath.Join(t.TempDir(), "leser.json")
	if _, err := leser.AusgebenNach(t.Context(), "pt-BR", andere); !errors.Is(err, domain.ErrErlaubnis) {
		t.Errorf("ein Leser schrieb: %v", err)
	}

	if _, err := os.Stat(andere); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("die Datei liegt trotzdem da: %v", err)
	}
}
