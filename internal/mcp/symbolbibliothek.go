// **Die Symbolbibliothek über MCP** — DOM-02
// (#1697).
//
// Einordnung nach §16: **erweiternd** — zwölf Werkzeuge kommen hinzu, drei
// lesend und neun schreibend. **Eine Datei ist ein Text**: ein MCP-Server
// steht an keiner Platte, die das Modell sieht (Vermerk 0140).

package mcp

import (
	"context"
	"encoding/json"
)

// Symbolbibliothek ist, was MCP für die Bibliothek braucht.
type Symbolbibliothek interface {
	Symbolbibliothek(ctx context.Context, suche string) (any, error)
	BibliothekssymbolText(ctx context.Context, id int64) (any, error)
	BibliothekssymbolAufnehmen(ctx context.Context, setID int64, name string) (any, error)
	BibliothekssymbolKopieren(ctx context.Context, id int64, name string) (any, error)
	BibliothekssymbolUmbenennen(ctx context.Context, id int64, name string) (any, error)
	BibliothekssymbolLoeschen(ctx context.Context, id int64, bestaetigt bool) (any, error)
	BibliothekssymbolEinsetzen(ctx context.Context, id, setID int64) (any, error)
	BibliothekssymbolAusText(ctx context.Context, inhalt, name string) (any, error)

	// Die Bearbeitung — DOM-02.AK-02, in `symbolbibliothekpflege.go`.
	BibliothekssymbolTeile(ctx context.Context, id int64) (any, error)
	BibliothekssymbolSichern(ctx context.Context, id int64, teile json.RawMessage) (any, error)
	BibliothekssymbolZuruecknehmen(ctx context.Context, id int64) (any, error)
	BibliothekssymbolNachziehen(ctx context.Context, id int64) (any, error)
}

// symbolbibliothekswerkzeuge gibt die lesenden und die schreibenden.
func symbolbibliothekswerkzeuge(naht Symbolbibliothek) (lesend, schreibend []Werkzeug) {
	if naht == nil {
		return nil, nil
	}

	return []Werkzeug{lesenSymbolbibliothek(naht), lesenBibliothekssymbol(naht), lesenSymbolteile(naht)},
		[]Werkzeug{
			schreibenSymbolAufnehmen(naht), schreibenSymbolDuplizieren(naht),
			schreibenSymbolUmbenennen(naht), schreibenSymbolLoeschen(naht),
			schreibenSymbolEinsetzen(naht), schreibenSymbolEinlesen(naht),
			schreibenSymbolSichern(naht), schreibenSymbolZuruecknehmen(naht),
			schreibenSymbolNachziehen(naht),
		}
}

// symbolkennung ist das Schema eines Rufs, der ein Symbol und einen Namen
// nennt.
func symbolkennungMitNamen(namenserklaerung string) map[string]any {
	return objekt(map[string]any{
		"id":   zahl("Die Kennung des Symbols aus »lesen.symbolbibliothek«."),
		"name": text(namenserklaerung),
	}, "id", "name")
}

// kennungUndName liest `id` und `name`.
func kennungUndName(roh json.RawMessage) (id int64, name string, fehler error) {
	var wahl struct {
		ID   int64  `json:"id"`
		Name string `json:"name"`
	}

	return wahl.ID, wahl.Name, lies(roh, &wahl)
}

func lesenSymbolbibliothek(naht Symbolbibliothek) Werkzeug {
	return Werkzeug{
		Name: "lesen.symbolbibliothek", Art: Lesend, Idempotent: true,
		Erklaerung: "Nennt die Symbole der Bibliothek dieses Profils — Symbole, die sich in " +
			"jedes Set einsetzen lassen —, je mit Kennung, Name, Abdruck (Prüfsumme der " +
			"Zeichnung) und den Sets, die es tragen. Ohne Vorschaubild; die Zeichnung " +
			"gibt »lesen.symbolbibliothek.symbol«. Symbolschriften stehen NICHT hier, " +
			"sondern unter den Paketen. ÄNDERT NICHTS.",
		Schema: objekt(map[string]any{
			"suche": text("Ein Teil des Namens, ohne Rücksicht auf Groß und Klein — leer für alle."),
		}),
		Ruf: func(ctx context.Context, roh json.RawMessage) (string, error) {
			var wahl struct {
				Suche string `json:"suche"`
			}

			if err := lies(roh, &wahl); err != nil {
				return "", err
			}

			return alsErgebnis(naht.Symbolbibliothek(ctx, wahl.Suche))
		},
	}
}

func lesenBibliothekssymbol(naht Symbolbibliothek) Werkzeug {
	return Werkzeug{
		Name: "lesen.symbolbibliothek.symbol", Art: Lesend, Idempotent: true,
		Erklaerung: "Gibt ein Symbol der Bibliothek als Text einer .mse-symbol-Datei im Feld " +
			"»inhalt« — so, wie MSE sie liest. Zurück in eine Bibliothek kommt es mit " +
			"»schreiben.symbolbibliothek.einlesen«. ÄNDERT NICHTS.",
		Schema: objekt(map[string]any{
			"id": zahl("Die Kennung des Symbols aus »lesen.symbolbibliothek«."),
		}, "id"),
		Ruf: func(ctx context.Context, roh json.RawMessage) (string, error) {
			id, _, err := kennungUndName(roh)
			if err != nil {
				return "", err
			}

			return alsErgebnis(naht.BibliothekssymbolText(ctx, id))
		},
	}
}

func schreibenSymbolAufnehmen(naht Symbolbibliothek) Werkzeug {
	return Werkzeug{
		Name: "schreiben.symbolbibliothek.aufnehmen", Art: Schreibend,
		Erklaerung: "ÄNDERT den Bestand: legt das Symbol eines Sets als NEUES Symbol in die " +
			"Bibliothek. Ein Set ohne Symbol und ein Name, den es in der Bibliothek schon " +
			"gibt (ohne Rücksicht auf Groß und Klein), werden abgewiesen. Das Set bleibt, " +
			"wie es ist.",
		Schema: objekt(map[string]any{
			"setID": zahl("Die Kennung des Sets aus »lesen.sets«."),
			"name":  text("Der Name des Symbols in der Bibliothek, höchstens 80 Zeichen."),
		}, "setID", "name"),
		Ruf: func(ctx context.Context, roh json.RawMessage) (string, error) {
			var wahl struct {
				SetID int64  `json:"setID"`
				Name  string `json:"name"`
			}

			if err := lies(roh, &wahl); err != nil {
				return "", err
			}

			return alsErgebnis(naht.BibliothekssymbolAufnehmen(ctx, wahl.SetID, wahl.Name))
		},
	}
}

func schreibenSymbolDuplizieren(naht Symbolbibliothek) Werkzeug {
	return Werkzeug{
		Name: "schreiben.symbolbibliothek.duplizieren", Art: Schreibend,
		Erklaerung: "ÄNDERT den Bestand: legt ein Symbol der Bibliothek unter einem anderen " +
			"Namen NOCH EINMAL an — dieselbe Zeichnung, derselbe Abdruck.",
		Schema: symbolkennungMitNamen("Der Name der Kopie, höchstens 80 Zeichen."),
		Ruf: func(ctx context.Context, roh json.RawMessage) (string, error) {
			id, name, err := kennungUndName(roh)
			if err != nil {
				return "", err
			}

			return alsErgebnis(naht.BibliothekssymbolKopieren(ctx, id, name))
		},
	}
}

func schreibenSymbolUmbenennen(naht Symbolbibliothek) Werkzeug {
	return Werkzeug{
		Name: "schreiben.symbolbibliothek.umbenennen", Art: Schreibend,
		Erklaerung: "ÄNDERT den Bestand: gibt einem Symbol der Bibliothek einen anderen " +
			"Namen. Die Sets, die es tragen, bleiben, wie sie sind.",
		Schema: symbolkennungMitNamen("Der neue Name, höchstens 80 Zeichen."),
		Ruf: func(ctx context.Context, roh json.RawMessage) (string, error) {
			id, name, err := kennungUndName(roh)
			if err != nil {
				return "", err
			}

			return alsErgebnis(naht.BibliothekssymbolUmbenennen(ctx, id, name))
		},
	}
}

func schreibenSymbolLoeschen(naht Symbolbibliothek) Werkzeug {
	return Werkzeug{
		Name: "schreiben.symbolbibliothek.loeschen", Art: Schreibend,
		Erklaerung: "ÄNDERT den Bestand: LÖSCHT ein Symbol aus der Bibliothek und nennt im Feld »sets« die Sets, " +
			"die es tragen. Tragen es welche, geht es NUR mit »bestaetigt« — die Absage " +
			"nennt dieselben Sets. Die Sets behalten ihre Kopie; verloren geht der Eintrag " +
			"der Bibliothek. Nicht umkehrbar.",
		Schema: objekt(map[string]any{
			"id":         zahl("Die Kennung des Symbols aus »lesen.symbolbibliothek«."),
			"bestaetigt": jaNein("Löscht auch ein Symbol, das Sets tragen."),
		}, "id"),
		Ruf: func(ctx context.Context, roh json.RawMessage) (string, error) {
			var wahl struct {
				ID         int64 `json:"id"`
				Bestaetigt bool  `json:"bestaetigt"`
			}

			if err := lies(roh, &wahl); err != nil {
				return "", err
			}

			return alsErgebnis(naht.BibliothekssymbolLoeschen(ctx, wahl.ID, wahl.Bestaetigt))
		},
	}
}

func schreibenSymbolEinsetzen(naht Symbolbibliothek) Werkzeug {
	return Werkzeug{
		Name: "schreiben.symbolbibliothek.einsetzen", Art: Schreibend,
		Erklaerung: "ÄNDERT ein Set: legt eine KOPIE des Symbols als .mse-symbol neben das Set " +
			"und macht sie zu dessen Setsymbol. Die Datei von vorher bleibt liegen. Die " +
			"Antwort nennt im Feld »datei« die neue Datei.",
		Schema: objekt(map[string]any{
			"id":    zahl("Die Kennung des Symbols aus »lesen.symbolbibliothek«."),
			"setID": zahl("Die Kennung des Sets aus »lesen.sets«."),
		}, "id", "setID"),
		Ruf: func(ctx context.Context, roh json.RawMessage) (string, error) {
			var wahl struct {
				ID    int64 `json:"id"`
				SetID int64 `json:"setID"`
			}

			if err := lies(roh, &wahl); err != nil {
				return "", err
			}

			return alsErgebnis(naht.BibliothekssymbolEinsetzen(ctx, wahl.ID, wahl.SetID))
		},
	}
}

func schreibenSymbolEinlesen(naht Symbolbibliothek) Werkzeug {
	return Werkzeug{
		Name: "schreiben.symbolbibliothek.einlesen", Art: Schreibend,
		Erklaerung: "ÄNDERT den Bestand: legt ein NEUES Symbol aus dem Text einer .mse-symbol-" +
			"Datei an, etwa aus »lesen.symbolbibliothek.symbol«. Höchstens 1 MiB, 4.096 " +
			"Teile, 65.536 Punkte und 32 Ebenen; was darüber liegt, wird mit der Grenze " +
			"abgewiesen.",
		Schema: objekt(map[string]any{
			"inhalt": text("Der Text der .mse-symbol-Datei."),
			"name":   text("Der Name des Symbols in der Bibliothek, höchstens 80 Zeichen."),
		}, "inhalt", "name"),
		Ruf: func(ctx context.Context, roh json.RawMessage) (string, error) {
			var wahl struct {
				Inhalt string `json:"inhalt"`
				Name   string `json:"name"`
			}

			if err := lies(roh, &wahl); err != nil {
				return "", err
			}

			return alsErgebnis(naht.BibliothekssymbolAusText(ctx, wahl.Inhalt, wahl.Name))
		},
	}
}

// alsErgebnis gibt eine Antwort der Naht als JSON — oder ihren Fehler.
func alsErgebnis(ergebnis any, fehler error) (string, error) {
	if fehler != nil {
		return "", fehler
	}

	return alsJSON(ergebnis)
}
