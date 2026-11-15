// **Ein neues, leeres Set** —
// #891.
//
// # Warum es das bis dahin nicht gab
//
// In den Bestand kam ein Set nur über eine **Datei**: ein `.mse-set`, ein
// Beispielset, MTG JSON oder Scryfall. Wer bei null anfangen wollte, konnte es
// nicht.
//
// # Was MSE2 dafür verlangt
//
// Zwei Angaben: **Spiel und Stil**, in zwei Listen — die zweite gefiltert auf
// `<spiel>-*`. Beide sind **verlangt**:
//
//	void NewSetWindow::done() {
//	  if (!stylesheet_list->hasSelection()) return;
//	  StyleSheetP stylesheet = stylesheet_list->getSelection<StyleSheet>();
//	  set = make_intrusive<Set>(stylesheet);
//
// `src/gui/new_window.cpp`, Zeilen 91–95, Fassung f9d9356 vom 2021-07-06. Ohne
// gewählten Stil kehrt MSE **zurück, ohne ein Set anzulegen** — ein `Set`
// entsteht dort gar nicht anders als aus einem Stil. Vorgewählt sind das
// zuletzt benutzte Spiel (`settings.default_game`, Zeile 53) und dessen
// zuletzt benutzter Stil (`settings.gameSettingsFor(*game).default_stylesheet`,
// Zeile 66).
//
// # Und jede Fläche verlangt ihn
//
// **Nicht nur der Dialog.** MSE prüft den Stil auch in `Set::validate`, also
// am **Set** und nicht am Knopf; die Fundstelle steht bei [pruefeStil].
// [Importeur.Neu] ruft die Prüfung für Fenster, Kommandozeile, MCP und REST
// gleich. Ein Set ohne Stil gibt es im Bestand trotzdem: jedes importierte,
// dessen Stilpaket fehlt, ist eines.
//
// **Das Fenster verlangt ihn**, und zwar wie MSE: `SetNeu.tsx` wählt den ersten
// Stil des Spiels vor und bietet kein „ohne Stil" mehr an.

package service

import (
	"context"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Neu legt ein leeres Set an und gibt seine Kennung.
//
// **Ohne Karten und ohne Rohblöcke.** Ein neues Set hat keine Herkunft
// ([domain.Set.Origin] bleibt leer, und das heißt genau das: aus keiner
// Datei), also auch nichts, was ein Schreiber unverändert zurückgeben müsste.
//
// **Der Name steht als Setwert, nicht nur in der Zeile.** `set info: title:`
// ist das Feld, das ein geschriebenes `.mse-set` trägt; ohne es hieße das Set
// im Bestand so und in der Datei gar nicht.
func (i *Importeur) Neu(ctx context.Context, name, spiel, stil string) (int64, error) {
	set := &domain.Set{
		Name:  name,
		Game:  spiel,
		Style: stil,
		Values: []domain.Value{
			{Field: "title", Text: name},
		},
		Cards: []*domain.Card{},
	}

	if fehler := set.ValidateBasics(); fehler != nil {
		return 0, fehler
	}

	// **Der Name wird eigens geprüft.** `ValidateBasics` verlangt ein Spiel,
	// aber keinen Namen — ein Import darf ein namenloses Set anlegen, weil die
	// Datei den Namen trägt und nicht die Zeile. Wer hier eines **anlegt**,
	// nennt es: eine leere Zeile in der Liste ist nicht zu finden und nicht
	// wiederzufinden.
	if strings.TrimSpace(name) == "" {
		return 0, domain.NewFieldError("name", domain.SchluesselPflichtfeld,
			"Ein neues Set braucht einen Namen.")
	}

	if fehler := pruefeStil(spiel, stil); fehler != nil {
		return 0, fehler
	}

	return i.speicher.SetAnlegen(ctx, set, nil)
}

// pruefeStil verlangt einen Stil, und zwar einen des Spiels.
//
// **ported.** MSE prüft beides in `Set::validate`:
//
//	if (!stylesheet) {
//	  throw Error(_ERROR_("no stylesheet specified for the set"));
//	}
//	if (stylesheet->game != game) {
//	  throw Error(_ERROR_("stylesheet and set refer to different game"));
//	}
//
// `src/data/set.cpp`, Zeilen 143–149, Fassung f9d9356 vom 2021-07-06. Es ist
// keine Bedingung des Dialogs, sondern eine des **Sets**: ein `Set` ohne Stil
// besteht dort die Prüfung nicht, gleich woher es kommt.
//
// # Warum der Name das Spiel entscheidet und nicht das Paket
//
// MSE liest das Feld `game:` des Stilpakets. Hier steht der Name da und das
// Paket nicht: `Neu` kennt den Paketkatalog nicht, und ihn dafür
// hereinzuziehen hieße, dem Fachkern eine Paketverwaltung anzuhängen.
//
// **Gemessen am 2026-08-25 über alle 376 Stilpakete des vollen Packs**: bei
// **keinem einzigen** weicht das Namenspräfix vom erklärten `game:` ab. MSE
// selbst filtert seine Liste so (`game->name() + "-*"`,
// `src/gui/new_window.cpp:65`).
//
// **Was er nicht prüft**: ob es das Paket gibt. Ein richtig benannter, aber
// nicht installierter Stil gibt ein Set, das den Behelfsplan zeichnet — derselbe
// Zustand wie bei jedem importierten Set, dessen Stilpaket fehlt.
func pruefeStil(spiel, stil string) error {
	if strings.TrimSpace(stil) == "" {
		return domain.NewFieldError("stil", domain.SchluesselPflichtfeld,
			"Ein neues Set braucht einen Stil.")
	}

	if !strings.HasPrefix(stil, spiel+"-") {
		return domain.NewFieldError("stil", domain.SchluesselValidierung,
			"Der Stil %s gehört nicht zum Spiel %s.", stil, spiel)
	}

	return nil
}
