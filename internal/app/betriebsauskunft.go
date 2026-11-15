package app

// **Was die Anwendung über sich selbst sagt** — OPS-07
// (#1724),
// Marke `new`.
//
// # Nichts, was nicht so ist
//
// **Ad-hoc signiert und nicht notarisiert** — so steht es da, weil es so
// ist ([0003](../../docs/entscheidungen/0003-keine-notarisierung.md)). Eine
// Übersicht, die „signiert" sagte und das Wort offenließe, verspräche, was
// Gatekeeper beim ersten Start widerlegt.
//
// # Die Orte als Tilde
//
// **Dieselbe Regel wie in der Diagnose**: der Ort ist die Auskunft, der Name
// des Benutzerverzeichnisses nicht.

import (
	"context"
	"os"
	"path/filepath"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// SignaturAdhoc ist die Signatur jedes Baus dieses Projekts.
const SignaturAdhoc = "adhoc"

// Die Namen der Orte.
const (
	OrtBestand          = "bestand"
	OrtPakete           = "pakete"
	OrtBilder           = "bilder"
	OrtZwischenspeicher = "zwischenspeicher"
	OrtProtokoll        = "protokoll"
	OrtProfile          = "leistungsprofile"
)

// BetriebsortDTO ist ein Ort, an dem die Anwendung etwas ablegt.
type BetriebsortDTO struct {
	Name string `json:"name"`
	Pfad string `json:"pfad"`
}

// BetriebsauskunftAntwort sagt, was installiert ist und was geprüft wurde.
type BetriebsauskunftAntwort struct {
	Fassung    string `json:"fassung"`
	Stand      string `json:"stand"`
	Gebaut     string `json:"gebaut"`
	Veraendert bool   `json:"veraendert"`

	// Migrationen und LetzteMigration sind die Fassung der Daten.
	Migrationen     int    `json:"migrationen"`
	LetzteMigration string `json:"letzteMigration"`

	Signatur    string `json:"signatur"`
	Notarisiert bool   `json:"notarisiert"`

	// Aktualisierung ist die Einstellung `aktualisierung` — `aus`, wo keine
	// gewählt ist.
	Aktualisierung string `json:"aktualisierung"`

	Orte       []BetriebsortDTO `json:"orte"`
	Pruefstand []PruefpunktDTO  `json:"pruefstand"`
	Fehler     *Fehler          `json:"fehler"`
}

// Betriebsauskunft stellt die Auskunft zusammen.
//
// **Ohne Fehlerweg**: jeder Teil hat eine Antwort für „weiß ich nicht" — kein
// Bestand heißt null Migrationen, eine unlesbare Einstellung heißt `aus`, ein
// Ort, der sich nicht bestimmen lässt, fehlt in der Liste.
func (a *App) Betriebsauskunft(ctx context.Context) BetriebsauskunftAntwort {
	ctx = a.imProfil(ctx)
	bau := DiesenBau()

	auskunft := BetriebsauskunftAntwort{
		Fassung: Fassung, Stand: bau.Stand, Gebaut: bau.Gebaut, Veraendert: bau.Veraendert,
		Signatur: SignaturAdhoc, Aktualisierung: domain.AktualisierungAus,
		Orte: a.betriebsorte(), Pruefstand: Pruefstand(),
	}

	if a.schemaquelle != nil {
		if angewendet, fehler := a.schemaquelle.AngewendeteMigrationen(ctx); fehler == nil && len(angewendet) > 0 {
			auskunft.Migrationen, auskunft.LetzteMigration = len(angewendet), angewendet[len(angewendet)-1]
		}
	}

	if a.einstellungen != nil {
		if werte, fehler := a.einstellungen.Alle(ctx); fehler == nil && werte[domain.EinstellungAktualisierung] != "" {
			auskunft.Aktualisierung = werte[domain.EinstellungAktualisierung]
		}
	}

	return auskunft
}

// betriebsorte zählt die Orte auf — was sich nicht bestimmen lässt, fehlt.
func (a *App) betriebsorte() []BetriebsortDTO {
	bestand := a.datenbankpfad
	if bestand == "" {
		bestand, _ = BestandsortFuerFlaechen() //nolint:errcheck // siehe oben: ein Ort, der fehlt, fehlt
	}

	zwischen, _ := os.UserCacheDir()   //nolint:errcheck // dito
	protokoll, _ := a.protokolldatei() //nolint:errcheck // dito
	pakete, _ := Paketverzeichnis()    //nolint:errcheck // dito
	bilder, _ := Bilderverzeichnis()   //nolint:errcheck // dito

	orte := make([]BetriebsortDTO, 0, 6)

	for _, ort := range []BetriebsortDTO{
		{Name: OrtBestand, Pfad: bestand},
		{Name: OrtPakete, Pfad: pakete},
		{Name: OrtBilder, Pfad: bilder},
		{Name: OrtZwischenspeicher, Pfad: filepath.Join(zwischen, "MagicSetEditor4Mac")},
		{Name: OrtProtokoll, Pfad: protokoll},
		{Name: OrtProfile, Pfad: filepath.Join(filepath.Dir(bestand), Profilordner)},
	} {
		if ort.Pfad != "" && filepath.IsAbs(ort.Pfad) {
			orte = append(orte, BetriebsortDTO{Name: ort.Name, Pfad: schwaerze(ort.Pfad)})
		}
	}

	return orte
}

// Betriebsauskunft ist die Fensterbindung.
func (b *Bindungen) Betriebsauskunft(ctx context.Context) BetriebsauskunftAntwort {
	return b.app(ctx).Betriebsauskunft(b.ruf(ctx))
}
