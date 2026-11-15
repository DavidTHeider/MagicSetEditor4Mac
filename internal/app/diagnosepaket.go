package app

// **Das Diagnosepaket: erst zeigen, dann geben** — Werkstück O10 von Phase 5.2
// (#1434).
//
// # Die Zusage, um derentwillen es das Werkstück gibt
//
// **Was gezeigt wurde, wird geschrieben — und sonst nichts.** [App.Diagnosepaket]
// stellt die Teile zusammen und **merkt sie sich**; [App.DiagnosepaketNach]
// schreibt genau diese. Zwischen Zeigen und Schreiben wächst das Protokoll
// weiter; ein zweites Zusammenstellen ergäbe ein anderes Paket als das
// gezeigte, und die Zustimmung gälte einem Inhalt, den niemand gesehen hat.
// Dieselbe Überlegung wie bei [App.DiagnoseSichern] — dort ist sie ein Befund
// des Zweitblicks über Phase 3.5.
//
// # Und nichts geht von selbst hinaus
//
// **Es gibt keinen Weg, der das Paket verschickt.** Es entsteht auf der
// Platte, dort, wo jemand hinzeigt; wer es weitergibt, tut das selbst.

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"runtime"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// Uebersichtsname ist der Name der Vorschau **im Paket**.
//
// **Sie liegt noch einmal darin**, obwohl sie vorher gezeigt wurde: wer das
// Paket später aufmacht — und das ist jemand anderes —, soll zuerst lesen, was
// darin ist, ohne jede Datei zu öffnen.
const Uebersichtsname = "uebersicht.json"

// Diagnosemanifest ist der Name des Manifests im Paket.
//
// **Nicht `sicherung.json`**: eine Sicherung spielt man zurück, ein
// Diagnosepaket verschickt man (`domain.Diagnosepaketendung`).
const Diagnosemanifest = "diagnose.json"

// dateinamen sagt, wie ein Teil im Archiv heißt.
var dateinamen = map[domain.Diagnoseteil]string{
	domain.TeilFassung:   "fassung.txt",
	domain.TeilBestand:   "bestand.txt",
	domain.TeilPakete:    "pakete.txt",
	domain.TeilMetriken:  "metriken.json",
	domain.TeilProtokoll: "protokoll.txt",
}

// PaketteilDTO ist ein Teil in der Vorschau.
type PaketteilDTO struct {
	Teil  string `json:"teil"`
	Datei string `json:"datei"`

	// Zeilen und Bytes sagen, wie groß der Teil ist — **bevor** jemand
	// zustimmt. „Ein Paket" ohne Zahl daneben ist keine Auskunft.
	Zeilen int `json:"zeilen"`
	Bytes  int `json:"bytes"`

	// Grund steht da, wenn der Teil **nicht** mitkommt — leer sonst.
	//
	// **Ein ausgelassener Teil wird genannt und nicht weggelassen** (Abnahme
	// aus dem Issue): „Quellen ohne Leserecht werden verständlich
	// ausgelassen" heißt, dass man sie vermisst und weiß, warum.
	Grund string `json:"grund"`
}

// DiagnosepaketAntwort ist die Vorschau.
type DiagnosepaketAntwort struct {
	Teile []PaketteilDTO `json:"teile"`

	// Bytes ist die Summe der Teile, Grenze die Obergrenze.
	Bytes  int `json:"bytes"`
	Grenze int `json:"grenze"`

	// Von und Bis sind der **Zeitraum** der aufgenommenen Protokollzeilen —
	// leer, wenn keine dabei sind. Das Issue verlangt ihn ausdrücklich: wer
	// zustimmt, soll wissen, über welche Zeit er Auskunft gibt.
	Von string `json:"von"`
	Bis string `json:"bis"`

	// Ohne nennt die Teile, die vor dem Sichern herausgenommen wurden —
	// leer in der Vorschau. **In der Übersicht des Archivs steht es**, damit
	// wer das Paket bekommt, sieht, dass etwas fehlt und was (OPS-04.AK-04,
	// #1721).
	Ohne []string `json:"ohne"`

	Fehler *Fehler `json:"fehler"`
}

// zeitmarke liest die Zeit aus einer Protokollzeile.
var zeitmarke = regexp.MustCompile(`(?m)^time=(\S+)`)

// Diagnosepaket stellt das Paket zusammen und zeigt, was darin steht.
func (a *App) Diagnosepaket(ctx context.Context) DiagnosepaketAntwort {
	ctx = a.imProfil(ctx)

	texte := DialogtexteFuer(a.sprache)
	teile := a.paketteile(ctx, texte)

	antwort := DiagnosepaketAntwort{
		Teile:  make([]PaketteilDTO, 0, len(teile)),
		Ohne:   []string{},
		Grenze: domain.Diagnosepaketgrenze,
	}

	for _, teil := range teile {
		antwort.Bytes += len(teil.inhalt)
		antwort.Teile = append(antwort.Teile, PaketteilDTO{
			Teil: string(teil.art), Datei: dateinamen[teil.art],
			Zeilen: zeilenzahl(teil.inhalt), Bytes: len(teil.inhalt),
			Grund: teil.grund,
		})

		if teil.art == domain.TeilProtokoll {
			antwort.Von, antwort.Bis = zeitraum(teil.inhalt)
		}
	}

	// **Die Grenze greift, bevor etwas geschrieben wird** (Abnahme): ein
	// Paket, das niemand mehr verschicken kann, soll gar nicht erst entstehen.
	if antwort.Bytes > domain.Diagnosepaketgrenze {
		antwort.Fehler = Uebersetze(ctx, domain.MitGrund(domain.ErrValidation,
			fmt.Errorf("%d von höchstens %d Byte", antwort.Bytes, domain.Diagnosepaketgrenze),
			"Das Diagnosepaket wäre zu groß."))

		return antwort
	}

	a.letzteVorschau, a.letzteTeile = &antwort, teile

	return antwort
}

// paketteil ist ein Teil samt seinem Text.
type paketteil struct {
	art    domain.Diagnoseteil
	inhalt string
	grund  string
}

// paketteile stellt die fünf Teile zusammen — **in der Reihenfolge aus
// `domain`**, damit Vorschau und Archiv dieselbe Ordnung haben.
func (a *App) paketteile(ctx context.Context, texte Dialogtexte) []paketteil {
	teile := make([]paketteil, 0, len(domain.Diagnoseteile()))

	for _, art := range domain.Diagnoseteile() {
		inhalt, grund := a.paketteil(ctx, art, texte)
		teile = append(teile, paketteil{art: art, inhalt: inhalt, grund: grund})
	}

	return teile
}

// paketteil gibt den Text eines Teils — oder den Grund, warum er fehlt.
func (a *App) paketteil(
	ctx context.Context, art domain.Diagnoseteil, texte Dialogtexte,
) (inhalt, grund string) {
	switch art {
	case domain.TeilFassung:
		bau := DiesenBau()

		return fmt.Sprintf("fassung\t%s\nplattform\t%s/%s\ngo\t%s\n"+
			"stand\t%s\ngebaut\t%s\nveraendert\t%t\nbindungsvertrag\t%s\n",
			Fassung, runtime.GOOS, runtime.GOARCH, runtime.Version(),
			bau.Stand, bau.Gebaut, bau.Veraendert, Vertragsabdruck()), ""

	case domain.TeilBestand:
		return a.bestandsteil(ctx, texte), ""

	case domain.TeilPakete:
		return redigiere(a.paketstaende(ctx, texte)), ""

	case domain.TeilMetriken:
		return metrikenAlsJSON(a.Metriken(ctx)), ""

	case domain.TeilProtokoll:
		// **Nur der Verwalter** (O9, #1433): das Protokoll gehört dem
		// Prozess und nicht dem Profil.
		if !a.darfDasProtokollSehen(ctx) {
			return "", texte.DiagnoseProtokollVerdeckt
		}

		pfad, fehler := a.protokolldatei()

		// **Und die Absturzspuren dahinter** (F04, #2343) — derselbe Kreis
		// darf sie sehen, und dieselbe Schwärzung geht darüber.
		return redigiere(protokollende(pfad, fehler, texte) + a.absturzspuren()), ""
	}

	// **Unerreichbar, solange die Aufzählung vollständig behandelt ist** — und
	// genau deshalb steht hier keine Ausrede, sondern ein Satz: ein neuer Teil
	// ohne Zweig fällt auf, statt still leer zu bleiben.
	return "", texte.DiagnoseTeilUnbekannt
}

// bestandsteil sagt, wo der Bestand liegt und in welchem Stand er ist —
// **ohne einen Inhalt daraus**.
func (a *App) bestandsteil(ctx context.Context, texte Dialogtexte) string {
	ort, fehler := BestandsortFuerFlaechen()
	if fehler != nil {
		ort = texte.DiagnoseNichtLesbar + ": " + fehler.Error()
	}

	return redigiere(fmt.Sprintf("%s\t%s\n%s\t%s\n%s\t%s\n",
		texte.DiagnoseBestand, ort,
		texte.DiagnoseSchema, a.schemastand(ctx, texte),
		texte.DiagnoseWaisen, a.waisenstand(ctx, texte)))
}

// metrikenAlsJSON schreibt den Zählstand.
//
// **Ohne Fehlerzweig**: [MetrikenAntwort] trägt Zeichenketten und Zahlen, und
// `json.Marshal` kann daran nicht scheitern — derselbe Grund wie bei
// [Fehler.MarshalJSON].
func metrikenAlsJSON(stand MetrikenAntwort) string {
	roh, _ := json.MarshalIndent(stand, "", "  ") //nolint:errcheck,errchkjson // siehe oben

	return string(roh) + "\n"
}

// alsArchivteile macht aus der Vorschau und den Teilen, was ins Archiv geht.
//
// **Die Übersicht zuerst**, und sie ist die Vorschau selbst: dieselbe
// Zeichenkette, die das Fenster gezeigt hat.
func alsArchivteile(antwort DiagnosepaketAntwort, teile []paketteil) []service.Paketteil {
	uebersicht, _ := json.MarshalIndent(antwort, "", "  ") //nolint:errcheck,errchkjson // wie oben

	archiv := make([]service.Paketteil, 0, len(teile)+1)
	archiv = append(archiv, service.Paketteil{
		Name: Uebersichtsname, Inhalt: append(uebersicht, '\n'),
	})

	for _, teil := range teile {
		archiv = append(archiv, service.Paketteil{
			Name: dateinamen[teil.art], Inhalt: []byte(teil.inhalt),
		})
	}

	return archiv
}

// zeilenzahl zählt die Zeilen eines Teils — null bei leerem Text.
func zeilenzahl(inhalt string) int {
	if inhalt == "" {
		return 0
	}

	return strings.Count(strings.TrimRight(inhalt, "\n"), "\n") + 1
}

// zeitraum gibt die erste und die letzte Zeitmarke der Protokollzeilen.
//
// **Aus den Zeilen und nicht aus der Uhr.** Wer zustimmt, soll wissen, über
// welche Zeit er Auskunft gibt — und das ist die Zeit **im Protokoll**, nicht
// die des Zusammenstellens.
func zeitraum(inhalt string) (von, bis string) {
	marken := zeitmarke.FindAllStringSubmatch(inhalt, -1)
	if len(marken) == 0 {
		return "", ""
	}

	return marken[0][1], marken[len(marken)-1][1]
}
