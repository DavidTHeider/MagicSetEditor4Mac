package main

// **Einen Sitz an ein Profil binden** — UXR-F21
// (#2041),
// Folgebefund aus UXR-F16
// (#2031).
//
// # Warum ein eigener Schalter und kein fünfter Teil im Sitz
//
// **Weil der Sitz sonst für den Normalfall unlesbar würde.** Er steht als
// `<kennung>:<name>:<deck>[:<commander>]` da, und der Commander fehlt in acht
// von neun Spielweisen: ein Profil dahinter hieße `anna:Anna:7::3` — ein
// leerer Teil, der aussieht wie ein Tippfehler. **Ein benannter Teil lässt die
// bestehende Form, wie sie ist** (§16, verträglich), und heißt, was er ist.
//
// # Warum nicht `--profil=`
//
// **Weil der Name schon vergeben ist**: `--profil=` wählt das Profil, in dem
// der **ganze Befehl** arbeitet (`zerlege`), und ein zweiter Sinn für dasselbe
// Wort machte aus einem Tippfehler eine Partie in einem fremden Profil.
//
// # Wo geprüft wird
//
// **Hier nur, was die Form der Kommandozeile betrifft** — ob der Schalter
// Sitz und Profil nennt und ob der Sitz unter den Argumenten vorkommt. Ob das
// Profil eine Kennung ist und ob es sie gibt, entscheidet `internal/app` an
// **einer** Stelle für alle vier Flächen (§2, §16); ein Text, der keine
// Kennung ist, kommt deshalb als derselbe fachliche Fehler zurück wie am
// Dienst.

import (
	"slices"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// sitzprofilvorsatz leitet `--sitzprofil=<sitzkennung>:<profilkennung>` ein.
const sitzprofilvorsatz = "--sitzprofil="

// trenneSitzprofile gibt die Argumente ohne Sitzprofile und diese für sich.
//
// **In der Reihenfolge, in der sie dastanden**, damit eine Meldung über den
// zweiten Schalter nicht den ersten nennt.
func trenneSitzprofile(rest []string) (uebrig, profile []string) {
	uebrig = make([]string, 0, len(rest))

	for _, argument := range rest {
		if strings.HasPrefix(argument, sitzprofilvorsatz) {
			profile = append(profile, argument)

			continue
		}

		uebrig = append(uebrig, argument)
	}

	return uebrig, profile
}

// sitzprofileBinden trägt die Profile aus den Schaltern in die Wünsche ein.
//
// **Ein Sitz hat höchstens ein Profil.** Zwei Schalter für denselben Sitz
// wären zwei Antworten auf dieselbe Frage, und welche gälte, hinge von der
// Reihenfolge ab — das wird abgewiesen statt gewählt. **Dasselbe Profil an
// zwei Sitzen ist dagegen erlaubt**: wer gegen sich selbst spielt, hat zwei
// (`domain.Partie.SitzeVonProfil`).
func sitzprofileBinden(wuensche []app.Sitzwunsch, schalter []string) error {
	for _, schalt := range schalter {
		sitz, profil, gut := strings.Cut(
			strings.TrimPrefix(schalt, sitzprofilvorsatz), ":")
		if !gut || sitz == "" || profil == "" {
			return domain.NewFieldError("sitzprofil", "sitzprofilform",
				"Ein Sitzprofil steht als --sitzprofil=<sitzkennung>:<profil> "+
					"da — %q ist keines.", schalt)
		}

		stelle := slices.IndexFunc(wuensche, func(wunsch app.Sitzwunsch) bool {
			return wunsch.Kennung == sitz
		})
		if stelle < 0 {
			return domain.NewFieldError("sitzprofil", "sitzprofilsitz",
				"Der Sitz %q kommt unter den Sitzen nicht vor — %q bindet "+
					"keinen.", sitz, schalt)
		}

		if wuensche[stelle].ProfilID != "" {
			return domain.NewFieldError("sitzprofil", "sitzprofilzweimal",
				"Der Sitz %q hat schon ein Profil — %q wäre ein zweites.",
				sitz, schalt)
		}

		wuensche[stelle].ProfilID = profil
	}

	return nil
}
