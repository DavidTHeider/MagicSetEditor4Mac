package service

// **Neuere Paketstände sichtbar und sicher aktualisieren** — G04
// (#2251).
//
// Ein bezogener Baum steht an einem **Commit** (§11), und der veraltet nicht
// von selbst — aber niemand erfuhr, dass es einen neueren gibt
// (`docs/PARITAET.md`, „Paketaktualisierungen anzeigen", bis hierher
// `unsupported`). MSE2 vergleicht eine Fassungsliste mit den installierten
// Paketen (`gui/update_checker.cpp`, Zeilen 27–80, Fassung f9d9356 vom
// 2021-07-06); dieses Programm fragt die Quelle selbst.
//
// # Drei Stände und ein Fehler, und keiner sieht aus wie der andere
//
// **Bezogen** ist der Commit, der auf der Platte liegt. **Geprüft** ist der,
// gegen den dieses Projekt gelaufen ist (`bezug.Quellen`). **Die Spitze** ist
// der, auf den der Hauptzweig der Quelle gerade zeigt — **als Commit**, nie
// als Zweig. Scheitert die Frage, ist das ein eigener Zustand mit seinem
// Grund, nicht „aktuell".
//
// # Was eine Aktualisierung vorher zeigt und wann sie fragt
//
// Welche Pakete sich ändern, welche Sets diese Pakete nennen — und sie
// verlangt eine **Bestätigung**, wo sie etwas nehmen könnte: wenn ein Paket,
// das ein Set nennt, Dateien verliert, wenn die Quelle gar nicht voraus ist,
// oder wenn die Gegenstelle nicht alle Dateien nennt.
//
// Portierungsmarke **adapted**: MSE2 bietet Aktualisierungen aus seiner Liste
// an und ersetzt ohne Aufbewahrung (`src/data/installer.cpp`); hier bleibt
// der bisherige Baum zum Zurückholen, eine Ebene tief
// (`paketaktualisierungstausch.go`).

import (
	"context"
	"fmt"
	"regexp"
	"slices"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/bezug"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Quellauskunft fragt eine Bezugsquelle — die Naht zu `internal/bezug`, damit
// ein Lauf ohne Netz auskommt. [Paketbezieher] erfüllt sie.
type Quellauskunft interface {
	Spitze(ctx context.Context, adresse string) (bezug.Spitzenstand, error)
	Vergleich(ctx context.Context, adresse, von, bis string) (bezug.Vergleich, error)
}

// Bezugslage sagt, wie ein bezogener Stand zu seiner Quelle steht.
type Bezugslage string

const (
	// BezugUngefragt: niemand hat gefragt — der Normalfall ohne Netz.
	BezugUngefragt Bezugslage = "ungefragt"
	// BezugAktuell: die Spitze der Quelle ist der bezogene Stand.
	BezugAktuell Bezugslage = "aktuell"
	// BezugNeuer: die Quelle ist voraus.
	BezugNeuer Bezugslage = "neuer"
	// BezugAbweichend: die Quelle ist **nicht** voraus — zurück oder
	// verzweigt. Eine Aktualisierung wäre ein Schritt zur Seite.
	BezugAbweichend Bezugslage = "abweichend"
	// BezugUnbekannt: die Frage scheiterte; der Grund steht dabei.
	BezugUnbekannt Bezugslage = "unbekannt"
)

// Die Gründe, aus denen eine Aktualisierung eine Bestätigung verlangt.
const (
	GrundSetVerliert = "setVerliert"
	GrundNichtVoraus = "nichtVoraus"
	GrundGekuerzt    = "gekuerzt"
)

// Bezugsstand ist, was über einen bezogenen Baum bekannt ist.
type Bezugsstand struct {
	// Name ist das Verzeichnis unter dem Paketordner, Ort sein Pfad.
	Name, Ort string

	Adresse string

	// Stand ist der bezogene Commit, Geprueft der, gegen den dieses Projekt
	// gelaufen ist — leer, wo die Quelle nicht in der Liste steht.
	Stand, Geprueft string

	Pakete int

	// Vorige sagt, dass ein bisheriger Stand zum Zurückholen bereitliegt,
	// VorigerStand, welcher.
	Vorige       bool
	VorigerStand string

	Lage      Bezugslage
	Spitze    bezug.Spitzenstand
	Vorsprung int

	// Fehler ist der Grund, wenn Lage [BezugUnbekannt] ist.
	Fehler error
}

// Paketaenderung ist ein Paket, das eine Aktualisierung berührt — gezählt in
// Dateien.
type Paketaenderung struct {
	Paket                    string
	Neu, Geaendert, Entfernt int
}

// Aktualisierungsvorschau ist, was eine Aktualisierung täte.
type Aktualisierungsvorschau struct {
	Name, Adresse string

	// Von ist der bezogene Stand, Zu der Commit, auf den aktualisiert würde.
	Von, Zu string

	Lage                  string
	Vorsprung, Rueckstand int

	Pakete []Paketaenderung

	// Uebrige zählt geänderte Dateien, die zu keinem Paket gehören.
	Uebrige int

	Gekuerzt bool

	// Sets sind die Sets dieses Profils, die ein berührtes Paket nennen.
	Sets []string

	BestaetigungNoetig bool
	Gruende            []string

	// VorigeErsetzt sagt, dass ein schon aufgehobener Stand weicht.
	VorigeErsetzt bool
}

// Paketaktualisierung führt Katalog, Bezug und Quelle zusammen.
type Paketaktualisierung struct {
	verwalter *Paketverwalter
	bezieher  *Paketbezieher
	auskunft  Quellauskunft
}

// NeuePaketaktualisierung baut sie über Verwalter, Bezieher und Auskunft.
func NeuePaketaktualisierung(verwalter *Paketverwalter, bezieher *Paketbezieher, auskunft Quellauskunft) *Paketaktualisierung {
	return &Paketaktualisierung{verwalter: verwalter, bezieher: bezieher, auskunft: auskunft}
}

// commitgestalt ist die Gestalt eines Commits — dieselbe wie in
// `internal/bezug`.
var commitgestalt = regexp.MustCompile(`^[0-9a-f]{40}$`)

// Staende nennt jeden bezogenen Baum — und fragt seine Quelle nur, wenn
// `fragen` gesetzt ist (§16, Regel 5: ohne Netz arbeitet die Anwendung
// weiter, und gefragt wird, wenn jemand fragt).
func (a *Paketaktualisierung) Staende(ctx context.Context, fragen bool) ([]Bezugsstand, error) {
	pakete, err := a.verwalter.katalog.Pakete(ctx)
	if err != nil {
		return nil, err
	}

	staende, err := a.vermerkteBaeume(a.baeume(pakete))
	if err != nil {
		return nil, err
	}

	for i := range staende {
		staende[i].Vorige, staende[i].VorigerStand = a.vorige(staende[i].Name)

		if fragen {
			a.lageFragen(ctx, &staende[i])
		}
	}

	return staende, nil
}

// lageFragen fragt die Quelle und trägt ein, wie der Stand zu ihr steht.
func (a *Paketaktualisierung) lageFragen(ctx context.Context, stand *Bezugsstand) {
	spitze, err := a.auskunft.Spitze(ctx, stand.Adresse)
	if err != nil {
		stand.Lage, stand.Fehler = BezugUnbekannt, err

		return
	}

	stand.Spitze = spitze

	if spitze.Commit == stand.Stand {
		stand.Lage = BezugAktuell

		return
	}

	// **Verglichen und nicht nur verschieden**: ein bezogener Stand kann eine
	// Freigabe sein, die auf denselben Commit zeigt, und eine Quelle kann
	// zurückgesetzt worden sein.
	vergleich, err := a.auskunft.Vergleich(ctx, stand.Adresse, stand.Stand, spitze.Commit)
	if err != nil {
		stand.Lage, stand.Fehler = BezugUnbekannt, err

		return
	}

	stand.Vorsprung = vergleich.Vorsprung

	switch vergleich.Lage {
	case "identical":
		stand.Lage = BezugAktuell
	case "ahead":
		stand.Lage = BezugNeuer
	default:
		stand.Lage = BezugAbweichend
	}
}

// Vorschau sagt, was eine Aktualisierung auf `ziel` täte — vorher.
//
// **`ziel` ist ein Commit.** Ein Zweig wird abgewiesen, bevor jemand gefragt
// wird: er ist die Nachführung, die hier nicht geschieht.
func (a *Paketaktualisierung) Vorschau(ctx context.Context, name, ziel string) (Aktualisierungsvorschau, error) {
	stand, err := a.bezogen(ctx, name, ziel)
	if err != nil {
		return Aktualisierungsvorschau{}, err
	}

	vergleich, err := a.auskunft.Vergleich(ctx, stand.Adresse, stand.Stand, ziel)
	if err != nil {
		return Aktualisierungsvorschau{}, err
	}

	vorschau := Aktualisierungsvorschau{
		Name: name, Adresse: stand.Adresse, Von: stand.Stand, Zu: ziel,
		Lage: vergleich.Lage, Vorsprung: vergleich.Vorsprung, Rueckstand: vergleich.Rueckstand,
		Gekuerzt: vergleich.Gekuerzt, Sets: []string{}, Gruende: []string{}, VorigeErsetzt: stand.Vorige,
	}

	vorschau.Pakete, vorschau.Uebrige = paketaenderungen(vergleich.Dateien)

	nutzung, err := a.verwalter.nutzung(ctx)
	if err != nil {
		return Aktualisierungsvorschau{}, err
	}

	setVerliert := false

	for _, paket := range vorschau.Pakete {
		for _, set := range nutzung[paket.Paket] {
			if !slices.Contains(vorschau.Sets, set) {
				vorschau.Sets = append(vorschau.Sets, set)
			}

			setVerliert = setVerliert || paket.Entfernt > 0
		}
	}

	slices.Sort(vorschau.Sets)

	for grund, gilt := range map[string]bool{
		GrundNichtVoraus: vergleich.Lage != "ahead",
		GrundGekuerzt:    vergleich.Gekuerzt,
		GrundSetVerliert: setVerliert,
	} {
		if gilt {
			vorschau.Gruende = append(vorschau.Gruende, grund)
		}
	}

	slices.Sort(vorschau.Gruende)
	vorschau.BestaetigungNoetig = len(vorschau.Gruende) > 0

	return vorschau, nil
}

// bezogen sucht einen bezogenen Baum — nachdem das Ziel als Commit geprüft
// ist.
func (a *Paketaktualisierung) bezogen(ctx context.Context, name, ziel string) (Bezugsstand, error) {
	if !commitgestalt.MatchString(ziel) {
		return Bezugsstand{}, domain.NewFieldError("stand", domain.SchluesselValidierung,
			"%q ist kein Commit. Ein Zweig ist kein Stand: aktualisiert wird auf einen Commit, den jemand gesehen hat.", ziel)
	}

	staende, err := a.Staende(ctx, false)
	if err != nil {
		return Bezugsstand{}, err
	}

	for _, stand := range staende {
		if stand.Name == name {
			return stand, nil
		}
	}

	return Bezugsstand{}, fmt.Errorf("%w: kein bezogenes Paket heißt %q", domain.ErrNotFound, name)
}

// paketaenderungen ordnet geänderte Dateien ihren Paketen zu.
//
// **Ein Paket ist das Verzeichnis mit `.mse-` im Namen**, wo es im Pfad auch
// steht — die Packs tragen ihre Pakete unter `data/`, andere unmittelbar.
//
// **Eine Umbenennung über Paketgrenzen ist ein Verlust und ein Zugang**:
// dem alten Paket fehlt die Datei danach, dem neuen kommt sie hinzu. Zählte
// sie nur als Änderung des neuen, verlor ein Set auf dem alten etwas, ohne
// dass die Vorschau eine Bestätigung verlangte (Befund 3 des Zweitblicks
// über Phase G, zweiter Teil).
func paketaenderungen(dateien []bezug.Aenderung) (pakete []Paketaenderung, uebrige int) {
	nachPaket := map[string]*Paketaenderung{}

	zaehle := func(paket string) *Paketaenderung {
		aenderung, da := nachPaket[paket]
		if !da {
			aenderung = &Paketaenderung{Paket: paket}
			nachPaket[paket] = aenderung
		}

		return aenderung
	}

	for _, datei := range dateien {
		paket := paketVon(datei.Datei)

		if vorher := paketVon(datei.Vorher); datei.Art == "renamed" && datei.Vorher != "" && vorher != paket {
			if vorher != "" {
				zaehle(vorher).Entfernt++
			}

			if paket != "" {
				zaehle(paket).Neu++
			}

			continue
		}

		if paket == "" {
			uebrige++

			continue
		}

		switch aenderung := zaehle(paket); datei.Art {
		case "added":
			aenderung.Neu++
		case "removed":
			aenderung.Entfernt++
		default:
			aenderung.Geaendert++
		}
	}

	pakete = make([]Paketaenderung, 0, len(nachPaket))
	for _, aenderung := range nachPaket {
		pakete = append(pakete, *aenderung)
	}

	slices.SortFunc(pakete, func(x, y Paketaenderung) int { return strings.Compare(x.Paket, y.Paket) })

	return pakete, uebrige
}

// paketVon ist das Paket, in dem ein Pfad liegt — leer, wenn keines.
func paketVon(pfad string) string {
	for _, teil := range strings.Split(pfad, "/") {
		if strings.Contains(teil, ".mse-") {
			return teil
		}
	}

	return ""
}
