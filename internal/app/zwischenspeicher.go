package app

// **Begrenzte, beobachtbare und richtig geleerte Zwischenspeicher** — OPS-08
// (#1725).
//
// # Acht Speicher, ein Verzeichnis
//
// **Welche es gibt, woran ihr Schlüssel hängt und warum ihre Grenze so
// lautet**, steht in `docs/ZWISCHENSPEICHER.md`. Hier steht, wie sie sich
// zeigen und leeren lassen: je Speicher Ort, Treffer, Fehlgriffe, Einträge,
// Bytes und Grenze — gleich ob er auf der Platte liegt oder im Arbeitsspeicher
// dieses Prozesses.
//
// # Leeren ist Wegwerfen, nicht Neuaufbauen
//
// **Ein geleerter Speicher füllt sich beim nächsten Gebrauch von selbst** —
// das ist der kontrollierte Wiederaufbau: nichts wird auf Vorrat geholt, und
// was ohne Netz nicht wiederkommt, fehlt, bis das Netz wieder da ist. Darum
// sagt die Oberfläche es vor dem Knopf.

import (
	"context"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msescript"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// Die Namen der Speicher, wie die Auskunft sie führt.
const (
	SpeicherAntworten   = "scryfall.antworten"
	SpeicherBilder      = "scryfall.bilder"
	SpeicherSammeldaten = "scryfall.sammeldaten"
	SpeicherEbenen      = "render.ebenen"
	SpeicherPaketbilder = "pakete.bilder"
	SpeicherSkripte     = "skript.baeume"
	SpeicherVorschau    = "vorschau.bilder"
	SpeicherTexturen    = "tisch.texturen"
)

// Die zwei Orte eines Speichers.
const (
	OrtPlatte          = "platte"
	OrtArbeitsspeicher = "arbeitsspeicher"
)

// werkspeicherfelder sind die Zwischenspeicher des Zeichenwerks in [App] —
// `nil` ohne Zeichenwerk: dann gibt es sie nicht, und die Auskunft nennt sie
// nicht.
type werkspeicherfelder struct {
	speicherquellen []func() ZwischenspeicherDTO
	speicherleerer  []func()
}

// ZwischenspeicherDTO ist ein Speicher mit seinen Zahlen.
type ZwischenspeicherDTO struct {
	Name       string `json:"name"`
	Ort        string `json:"ort"`
	Treffer    int64  `json:"treffer"`
	Fehlgriffe int64  `json:"fehlgriffe"`
	Eintraege  int    `json:"eintraege"`
	Bytes      int64  `json:"bytes"`

	// GrenzeEintraege und GrenzeBytes sind die Grenzen — null heißt: diese
	// Grenze hat er nicht.
	GrenzeEintraege int   `json:"grenzeEintraege"`
	GrenzeBytes     int64 `json:"grenzeBytes"`
}

// ZwischenspeicherAntwort trägt die Speicher dieses Prozesses.
type ZwischenspeicherAntwort struct {
	Speicher []ZwischenspeicherDTO `json:"speicher"`
	Fehler   *Fehler               `json:"fehler"`
}

// scryfallantwort erkennt einen Eintrag des Antwortspeichers — die Prüfsumme
// seines Schlüssels (`internal/netz`, `Speicher.pfad`).
var scryfallantwort = regexp.MustCompile(`^[0-9a-f]{64}\.json$`)

// plattenspeicher gibt die drei Speicher auf der Platte.
func plattenspeicher(ort string) map[string]service.Plattenspeicher {
	return map[string]service.Plattenspeicher{
		SpeicherAntworten:   {Ordner: ort, Eintrag: scryfallantwort.MatchString},
		SpeicherBilder:      {Ordner: filepath.Join(ort, "bilder"), Eintrag: fertigeDatei},
		SpeicherSammeldaten: {Ordner: filepath.Join(ort, "sammeldaten"), Eintrag: fertigeDatei},
	}
}

// fertigeDatei erkennt, was kein Lauf mehr schreibt.
//
// **Nicht die Nachbardatei eines Bildes** (`.<name>.<zufall>`, aus
// [service.SchreibeAtomar]) und **nicht die `.teil` einer Sammeldatei**
// (`internal/scryfall`, `BulkHolen`): bis zum Zweitblick über Phase 6.4
// zählten beide mit, und ein Leeren nahm sie einem laufenden Schreiben unter
// den Füßen weg (Befund 14).
func fertigeDatei(name string) bool {
	return !strings.HasPrefix(name, ".") && !strings.HasSuffix(name, ".teil")
}

// Zwischenspeicher nennt jeden Speicher mit seinen Zahlen.
func (a *App) Zwischenspeicher(ctx context.Context) ZwischenspeicherAntwort {
	ctx = a.imProfil(ctx)

	speicher, fehler := a.zwischenspeicherstand(ctx)
	if fehler != nil {
		return ZwischenspeicherAntwort{Speicher: []ZwischenspeicherDTO{}, Fehler: Uebersetze(ctx, fehler)}
	}

	return ZwischenspeicherAntwort{Speicher: speicher}
}

// ZwischenspeicherLeeren leert alle und gibt, was vorher belegt war.
func (a *App) ZwischenspeicherLeeren(ctx context.Context) ZwischenspeicherAntwort {
	ctx = a.imProfil(ctx)

	vorher, fehler := a.zwischenspeicherstand(ctx)
	if fehler != nil {
		return ZwischenspeicherAntwort{Speicher: []ZwischenspeicherDTO{}, Fehler: Uebersetze(ctx, fehler)}
	}

	ort, _ := Scryfallverzeichnis() //nolint:errcheck // eben schon gelesen, ohne Fehler
	for _, platte := range plattenspeicher(ort) {
		if raeumbar(ort) {
			platte.Leeren()
		}
	}

	msescript.BaumspeicherLeeren()
	a.bildfach.Leere()

	for _, leeren := range a.speicherleerer {
		leeren()
	}

	return ZwischenspeicherAntwort{Speicher: vorher}
}

// zwischenspeicherstand stellt die Liste zusammen — die Platte zuerst.
func (a *App) zwischenspeicherstand(ctx context.Context) ([]ZwischenspeicherDTO, error) {
	ort, fehler := Scryfallverzeichnis()
	if fehler != nil {
		return nil, fehler
	}

	grenze := a.zwischenspeichergrenze(ctx)
	platten := plattenspeicher(ort)
	antworten := speicherzahlen()

	stand := make([]ZwischenspeicherDTO, 0, 8)

	for _, name := range []string{SpeicherAntworten, SpeicherBilder, SpeicherSammeldaten} {
		belegung := platten[name].Belegung()
		eintrag := ZwischenspeicherDTO{
			Name: name, Ort: OrtPlatte, Eintraege: belegung.Eintraege, Bytes: belegung.Bytes,
		}

		if name == SpeicherAntworten {
			eintrag.Treffer, eintrag.Fehlgriffe = antworten.Treffer, antworten.Geholt
		}

		if name != SpeicherSammeldaten {
			eintrag.GrenzeBytes = grenze
		}

		stand = append(stand, eintrag)
	}

	skripte := msescript.Baumzahlen()
	stand = append(stand,
		ZwischenspeicherDTO{
			Name: SpeicherSkripte, Ort: OrtArbeitsspeicher, Treffer: skripte.Treffer,
			Fehlgriffe: skripte.Fehlgriffe, Eintraege: skripte.Eintraege, GrenzeEintraege: skripte.GrenzeEintraege,
			Bytes: skripte.Bytes, GrenzeBytes: skripte.GrenzeBytes,
		},
		a.vorschauspeicher())

	for _, quelle := range a.speicherquellen {
		stand = append(stand, quelle())
	}

	return stand, nil
}

// zwischenspeichergrenze liest die Grenze aus den Einstellungen — ohne sie
// gilt die Vorgabe.
func (a *App) zwischenspeichergrenze(ctx context.Context) int64 {
	if a.einstellungen == nil {
		return domain.ZwischenspeichergrenzeAus("")
	}

	werte, fehler := a.einstellungen.Alle(ctx)
	if fehler != nil {
		return domain.ZwischenspeichergrenzeAus("")
	}

	return domain.ZwischenspeichergrenzeAus(werte[domain.ZwischenspeicherGrenze])
}

// zwischenspeicherPflegen hält die Grenze auf der Platte ein — beim Start.
func (a *App) zwischenspeicherPflegen(ctx context.Context) {
	pflegeZwischenspeicher(a.zwischenspeichergrenze(ctx))
}

// testlauf sagt, ob dieser Prozess ein Testlauf ist — erkannt wie beim
// Anmeldebund (`internal/schluesselbund`) an `os.Args[0]`.
var testlauf = strings.HasSuffix(os.Args[0], ".test")

// raeumbar sagt, ob an diesem Ort gelöscht werden darf — **aus einem
// Testlauf nur unter dessen Zwischenablage**. Viele Läufe starten die
// Anwendung, ohne das Zuhause umzulenken; beschnitte der Start dann, träfe er
// den echten Speicher dieses Rechners. Dieselbe Klasse wie der Umzug am
// 2026-09-26 ([umzugErlaubt]).
func raeumbar(ort string) bool {
	return !testlauf || strings.HasPrefix(ort, filepath.Clean(os.TempDir())+string(filepath.Separator))
}

// pflegeZwischenspeicher beschneidet die Speicher auf der Platte — für das
// Fenster und für die Server (`ZwischenspeicherPflegen`).
func pflegeZwischenspeicher(grenze int64) {
	ort, fehler := Scryfallverzeichnis()
	if fehler != nil || !raeumbar(ort) {
		return
	}

	platten := plattenspeicher(ort)

	platten[SpeicherAntworten].Beschneiden(grenze)
	platten[SpeicherBilder].Beschneiden(grenze)
}

// Zwischenspeicher ist die Fensterbindung.
func (b *Bindungen) Zwischenspeicher(ctx context.Context) ZwischenspeicherAntwort {
	return b.app(ctx).Zwischenspeicher(b.ruf(ctx))
}

// ZwischenspeicherLeeren ist die Fensterbindung.
func (b *Bindungen) ZwischenspeicherLeeren(ctx context.Context) ZwischenspeicherAntwort {
	return b.app(ctx).ZwischenspeicherLeeren(b.ruf(ctx))
}
