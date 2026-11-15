package service

// **Welche Bäume bezogen sind, und wo ihr voriger Stand liegt** — G04
// (#2251).
//
// Getrennt von `paketaktualisierung.go`, weil es eine eigene Frage
// beantwortet (§2): dort, wie ein Baum zu seiner Quelle steht und was eine
// Aktualisierung täte; hier, **welche** Bäume es gibt.

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/bezug"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// baeume fasst die Pakete nach dem Baum zusammen, aus dem sie bezogen wurden.
//
// **Ein Bezug ist ein Verzeichnis unter dem Paketordner mit Herkunft
// `adresse@stand`** (`app.herkunftstext`). Ein einzeln eingebundenes Paket
// liegt unmittelbar im Paketordner und ist keiner.
func (a *Paketaktualisierung) baeume(pakete []domain.Paket) []Bezugsstand {
	nachName := map[string]*Bezugsstand{}

	for _, paket := range pakete {
		teile, imOrdner := pfadteileUnter(a.verwalter.verwaltet, paket.Pfad)
		trenner := strings.LastIndex(paket.Herkunft, "@")

		if !imOrdner || len(teile) < 2 || trenner < 0 {
			continue
		}

		stand, da := nachName[teile[0]]
		if !da {
			stand = &Bezugsstand{
				Name: teile[0], Ort: filepath.Join(a.verwalter.verwaltet, teile[0]),
				Adresse: paket.Herkunft[:trenner], Stand: paket.Herkunft[trenner+1:],
				Geprueft: gepruefterStand(paket.Herkunft[:trenner]), Lage: BezugUngefragt,
			}
			nachName[teile[0]] = stand
		}

		stand.Pakete++
	}

	staende := make([]Bezugsstand, 0, len(nachName))
	for _, stand := range nachName {
		staende = append(staende, *stand)
	}

	slices.SortFunc(staende, func(x, y Bezugsstand) int { return strings.Compare(x.Name, y.Name) })

	return staende
}

// vermerkteBaeume liest den Vermerk jedes Baums — er ist, wie der Baum
// bezogen wurde, und gilt **vor** dem Katalog.
//
// **Ein Bezug ohne Paket ist trotzdem einer**: der erste Commit einer Quelle
// kann nur ihre README tragen. Der Katalog kennt ihn dann nicht — aber er
// liegt auf der Platte, und aktualisieren lässt er sich nur, wenn er hier
// steht. Gefunden am nativen Fenster (`g04bezuege`). **Ohne lesbaren Vermerk
// ist ein Verzeichnis kein Bezug**: es nennt keine Quelle.
//
// **Und einem bekannten Baum gibt er Adresse und Stand** (Befund 4 des
// Zweitblicks über Phase G, zweiter Teil): nach einem geglückten Tausch und
// einem gescheiterten Vermerken nannte der Katalog den alten Stand, und die
// nächste Vorschau verglich gegen ihn. Die Herkunft im Katalog bleibt der
// Rückfall für einen Baum von vor G04, der keinen Vermerk hat.
func (a *Paketaktualisierung) vermerkteBaeume(staende []Bezugsstand) ([]Bezugsstand, error) {
	eintraege, err := os.ReadDir(a.verwalter.verwaltet)
	if errors.Is(err, os.ErrNotExist) {
		return staende, nil
	}

	if err != nil {
		return nil, fmt.Errorf("den Paketordner lesen: %w", err)
	}

	for _, eintrag := range eintraege {
		// **Verborgen liegt ein Bezug, der gerade entsteht** (`beziehen`).
		if !eintrag.IsDir() || strings.HasPrefix(eintrag.Name(), ".") {
			continue
		}

		ort := filepath.Join(a.verwalter.verwaltet, eintrag.Name())

		vermerk, err := vorigerVermerk(ort)
		if err != nil {
			continue
		}

		stelle := slices.IndexFunc(staende, func(stand Bezugsstand) bool { return stand.Name == eintrag.Name() })
		if stelle < 0 {
			staende = append(staende, Bezugsstand{Name: eintrag.Name(), Ort: ort, Lage: BezugUngefragt})
			stelle = len(staende) - 1
		}

		staende[stelle].Adresse, staende[stelle].Stand = vermerk.Adresse, vermerk.Stand
		staende[stelle].Geprueft = gepruefterStand(vermerk.Adresse)
	}

	slices.SortFunc(staende, func(x, y Bezugsstand) int { return strings.Compare(x.Name, y.Name) })

	return staende, nil
}

// teileUnter zerlegt einen Pfad unter einer Wurzel in seine Teile.
func pfadteileUnter(wurzel, pfad string) ([]string, bool) {
	relativ, err := filepath.Rel(wurzel, pfad)
	if err != nil || relativ == ".." || strings.HasPrefix(relativ, ".."+string(filepath.Separator)) {
		return nil, false
	}

	return strings.Split(relativ, string(filepath.Separator)), true
}

// gepruefterStand nennt den Stand aus der Liste der Quellen — oder nichts.
func gepruefterStand(adresse string) string {
	for _, quelle := range bezug.Quellen() {
		if quelle.Adresse == strings.TrimSuffix(adresse, "/") {
			return quelle.Stand
		}
	}

	return ""
}

// vorige sagt, ob ein bisheriger Stand eines Baums bereitliegt, und welcher.
func (a *Paketaktualisierung) vorige(name string) (da bool, stand string) {
	if _, err := os.Stat(a.vorigeOrt(name)); err != nil {
		return false, ""
	}

	vermerk, err := vorigerVermerk(a.vorigeOrt(name))
	if err != nil {
		return true, ""
	}

	return true, vermerk.Stand
}

// vorigeOrt ist, wo der bisherige Stand eines Baums liegt — **neben** dem
// Paketordner und nicht darin, aus demselben Grund wie [Paketverwalter.vorigeOrt]:
// dort fände der Katalog jedes Paket darin ein zweites Mal.
func (a *Paketaktualisierung) vorigeOrt(name string) string {
	return filepath.Join(a.verwalter.verwaltet+"-vorige-bezuege", name)
}
