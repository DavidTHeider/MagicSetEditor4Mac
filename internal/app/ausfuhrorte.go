package app

import (
	"context"
	"os"
	"path/filepath"
	"strings"
)

// **Wohin eine Ausfuhr schreibt** —
// #732.
//
// Zwei Auskünfte, die vor dem Lauf feststehen müssen: das Verzeichnis für die
// Beigaben und der Ort des Sets. Beide hängen am Ziel, das jemand gewählt hat —
// deshalb stehen sie hier und nicht am Ausführer.

// beigabenname sagt, wo die Beigaben liegen würden — ohne das Verzeichnis
// anzulegen.
//
// **Der relative Name geht an die Vorlage, der Pfad an die Schreiber.** Ein
// absoluter Pfad in einer HTML-Datei zeigte auf den Rechner, auf dem
// exportiert wurde.
//
// **Ohne `create directory:` gibt es keines** — dann darf die Vorlage gar
// nicht schreiben (#732).
//
// **Der Weg des Probelaufs und des echten Laufs ist derselbe bis hierher**
// (#2015):
// nur das Anlegen kommt im echten Lauf dazu.
func beigabenname(ziel string, gewollt bool) (ort, name string) {
	if !gewollt {
		return filepath.Dir(ziel), ""
	}

	stamm := strings.TrimSuffix(filepath.Base(ziel), filepath.Ext(ziel))
	name = stamm + "-files"

	return filepath.Join(filepath.Dir(ziel), name), name
}

// beigabenort legt das Verzeichnis für die Beigaben an und nennt es.
//
// **Ohne `create directory:` entsteht keines**, und ein leeres Verzeichnis
// daneben wäre Müll.
func beigabenort(ziel string, gewollt bool) (ort, name string, fehler error) {
	ort, name = beigabenname(ziel, gewollt)

	if !gewollt {
		return ort, name, nil
	}

	if err := os.MkdirAll(ort, 0o750); err != nil {
		return "", "", err
	}

	return ort, name, nil
}

// setort gibt die Herkunftsdatei eines Sets — leer, wenn es keine gibt.
//
// **Ein Lesefehler ist kein Abbruch.** Was dann fehlt, ist das Setsymbol; die
// Ausfuhr selbst geht weiter, und das ist mehr als gar keine (§8).
func (a *App) setort(ctx context.Context, setID int64) string {
	koepfe, err := a.leser().Setkoepfe(ctx)
	if err != nil {
		return ""
	}

	for _, kopf := range koepfe {
		if kopf.ID == setID {
			return kopf.Origin
		}
	}

	return ""
}

// schreibeVorlage führt die Vorlage aus und legt das Ergebnis ab.
