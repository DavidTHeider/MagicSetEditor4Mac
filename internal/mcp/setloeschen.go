// **Ein Set löschen** —
// #1218.
//
// # Warum ein Sprachmodell das dürfen soll
//
// Weil §16 es verlangt: was das Fenster kann, können die anderen Flächen — oder
// es steht begründet, warum nicht. Und die Begründung wäre hier schwach. Ein
// Modell, das gebeten wird „räum meine Testsets weg", hätte sonst keinen
// Schritt außer dem Rat, `sqlite3` zu öffnen.
//
// # Warum der Name die richtige Sperre ist
//
// Es ist der **einzige unumkehrbare** Ruf dieser Fläche, und eine Sperre, die
// nur aus einer Warnung im Erklärungstext besteht, ist keine. Der Name
// zwingt zum Nachsehen: wer ihn nennt, hat `lesen.sets` gerufen. Eine
// Kennung allein ließe sich raten — Namen nicht.
//
// **Der Bestand prüft ihn, nicht dieses Werkzeug.** Eine Prüfung hier wäre
// ein zweiter Stand und damit ein Wettlauf; sie steht dort, wo sie mit dem
// Löschen in einer Transaktion liegt.

package mcp

import (
	"context"
	"encoding/json"
)

// Setloescher ist, was dieses Werkzeug braucht.
type Setloescher interface {
	SetLoeschen(ctx context.Context, setID int64, name string) ([]string, error)
}

// Setloeschung ist die Antwort.
//
// **Die Befunde stehen darin, und sie sind kein Fehler** (§7, Regel 3): das
// Set ist weg, und was daneben liegen blieb — eingesetzte Bilder auf der
// Platte — steht als Meldung da. Ein Modell, das nur auf den Fehler sieht,
// erfährt trotzdem, dass alles ging.
type Setloeschung struct {
	Geloescht bool     `json:"geloescht"`
	Befunde   []string `json:"befunde"`
}

// ohneLeer macht aus `nil` eine leere Liste.
//
// **Damit `befunde` in der Antwort steht und nicht `null`.** Ein Gegenüber,
// das über die Länge geht, stolpert sonst über den Unterschied zwischen „keine
// Befunde" und „das Feld gibt es nicht".
func ohneLeer(texte []string) []string {
	if texte == nil {
		return []string{}
	}

	return texte
}

// setloeschwerkzeug gibt das schreibende Werkzeug.
func setloeschwerkzeug(loescher Setloescher) []Werkzeug {
	return []Werkzeug{
		{
			Name: "schreiben.set.loeschen", Art: Schreibend,
			Erklaerung: "ÄNDERT den Bestand UNUMKEHRBAR: nimmt ein GANZES Set " +
				"samt allen Karten, Werten und Schlüsselwörtern weg. Es gibt " +
				"KEINEN Papierkorb und KEIN Zurück. " +
				"name muss der HEUTIGE Name des Sets sein — genau, mit Groß- " +
				"und Kleinschreibung; stimmt er nicht, wird abgewiesen und " +
				"nichts gelöscht. Den Namen gibt »lesen.sets«. " +
				"Die SAMMLUNG bleibt: ein Exemplar hängt an einem Druck, nicht " +
				"an einem Set.",
			Schema: objekt(map[string]any{
				"setID": zahl("Die Kennung des Sets."),
				"name": text("Der heutige Name des Sets, genau. Er ist die " +
					"Bestätigung: stimmt er nicht, wird nichts gelöscht."),
			}, "setID", "name"),
			Ruf: func(ctx context.Context, roh json.RawMessage) (string, error) {
				var argumente struct {
					SetID int64  `json:"setID"`
					Name  string `json:"name"`
				}

				if err := lies(roh, &argumente); err != nil {
					return "", err
				}

				if pruefung := pruefeKennung("setID", argumente.SetID); pruefung != nil {
					return "", pruefung
				}

				befunde, fehler := loescher.SetLoeschen(ctx, argumente.SetID, argumente.Name)
				if fehler != nil {
					return "", fehler
				}

				return alsJSON(Setloeschung{Geloescht: true, Befunde: ohneLeer(befunde)})
			},
		},
	}
}
