// **Die Werkzeuge, die den Arbeitsablauf schließen** — einfahren, sichern,
// zeichnen, den Bogen legen
// (#886).
//
// # Warum sie erlaubt sind
//
// [0040](../../docs/entscheidungen/0040-der-lebenszyklus-einer-karte-gehoert-dazu.md)
// sagt: kein Werkzeug überschreibt ein Set, keines ändert mehrere Karten auf
// einmal. Nach dem Wortlaut fielen alle drei darunter — ein Einfahren legt
// Hunderte Karten an.
//
// [0044](../../docs/entscheidungen/0044-was-ein-werkzeug-schreiben-darf.md)
// stellt den Maßstab richtig: **nicht die Menge, sondern die Umkehrbarkeit.**
// Ein Werkzeug schreibt nur, was es nicht überschreibt — ein neues Set, eine
// Datei an einem freien Ort.
//
// # Und warum die Kommandozeile es anders darf
//
// `mse zeichnen` überschreibt ein vorhandenes Ziel. Der Unterschied ist die
// **Aufsicht**, nicht die Fähigkeit: wer tippt, sieht seinen Aufruf; ein
// Sprachmodell ruft, während niemand hinsieht, und wiederholt nach einer
// Zeitüberschreitung. Der Grund steht ausführlich in 0044.

package mcp

import (
	"context"
	"encoding/json"
	"fmt"
)

// Einfahrt ist, was ein Einfahren hergibt.
//
// **Die Befunde gehören dazu** (§8): ein Set mit 300 lesbaren und zwei
// unlesbaren Karten wird eingefahren, nicht abgewiesen — und wer die zwei
// nicht sähe, hielte ein halbes Set für ein ganzes.
type Einfahrt struct {
	SetID   int64
	Karten  int
	Befunde []string
}

// MTGJSONEinfahrt ist, was ein Einlesen von MTG JSON hergibt.
//
// **Alle angelegten Kennungen und nicht eine** (#1291, Befund 1 des
// Zweitblicks): eine AllSets-Datei führt Hunderte Sets, jedes in einer eigenen
// Transaktion. Was vor einem Fehler durchkam, **steht im Bestand** — und wer
// nur den Fehler bekäme, ließe Sets zurück, von denen niemand weiß (§8).
type MTGJSONEinfahrt struct {
	// Angelegt sind die Kennungen der Sets, die entstanden sind.
	Angelegt []int64 `json:"angelegt"`

	// SchonOffen sagt, dass diese Quelle schon eingefahren war.
	SchonOffen bool `json:"schonOffen"`

	Befunde []string `json:"befunde"`
}

// Ausfuhr ist, was die drei Werkzeuge des Arbeitsablaufs brauchen.
//
// Die Schnittstelle sagt, was `mcp` **braucht** — nicht, was `internal/app`
// anzubieten hat (§2). Erfüllt wird sie im Composition Root des Servers.
type Ausfuhr interface {
	// Einfahren liest ein Setpaket in den Bestand.
	Einfahren(ctx context.Context, pfad string) (Einfahrt, error)

	// Sichern schreibt ein Set an einen Ort, an dem nichts liegt.
	Sichern(ctx context.Context, setID int64, ziel string) (befunde []string, fehler error)

	// Zeichnen rastert eine Karte an einen Ort, an dem nichts liegt, und
	// gibt die Maße samt Befunden.
	//
	// **Die Befunde gehören dazu** (#1137): eine fehlende Schrift oder eine
	// Verknüpfung, die dieses Programm nicht kann, sind das, was ein Modell
	// wissen muss — die Datei entsteht trotzdem.
	Zeichnen(ctx context.Context, karteID int64, ziel, guete string) (
		breite, hoehe int, befunde []string, fehler error,
	)

	// Bogen legt die Karten eines Sets als Druckbogen an einen Ort, an dem
	// nichts liegt.
	//
	// **Ohne Papierwahl und ohne Auswahl** (#1291). Am Fenster stellt die
	// Seite beides; ein Modell, das einen Bogen zum Drucken braucht, braucht
	// A4 mit neun Karten in Druckauflösung — und wer etwas anderes will,
	// bekommt es an einer Fläche, an der jemand hinsieht. Erweitern geht
	// später, zurücknehmen nicht
	// ([0033](../../docs/entscheidungen/0033-was-ueber-fassungen-hinweg-zugesagt-wird.md)).
	Bogen(ctx context.Context, setID int64, ziel string) (
		seiten, karten int, befunde []string, fehler error,
	)

	// MTGJSONLesen fährt eine MTG-JSON-Quelle in den Bestand ein (#1291).
	//
	// **Datei oder Verzeichnis, ein Weg**: die zwei Bindungen am Fenster gibt
	// es nur, weil macOS zwei Dialoge kennt.
	//
	// **Nicht [Einfahrt], sondern [MTGJSONEinfahrt]** — und das ist Befund 1
	// des Zweitblicks über Phase 4.4. Eine AllSets-Datei legt **viele** Sets
	// an; `Einfahrt` trägt eine Kennung und eine Kartenzahl. Der erste Anlauf
	// setzte `Karten: len(Angelegt)` — ein Feld, das Karten heißt und Sets
	// zählte.
	MTGJSONLesen(ctx context.Context, pfad, spiel string) (MTGJSONEinfahrt, error)

	// MTGJSONSchreiben schreibt ein Set als MTG JSON an einen Ort, an dem
	// nichts liegt.
	MTGJSONSchreiben(ctx context.Context, setID int64, spiel, ziel string) (
		karten int, befunde []string, fehler error,
	)

	// MTGJSONSpiele nennt die Spiele, für die es eine Zuordnung gibt.
	//
	// **Ohne Kontext und ohne Fehler**: die Zuordnungen stehen im Programm.
	MTGJSONSpiele() []string
}

// ablauf gibt die Werkzeuge, die den Arbeitsablauf schließen.
//
// **Keines sagt Idempotenz zu.** Ein zweiter Ruf legt eine Datei an, die es
// noch nicht gab, oder wird abgewiesen, weil sie schon da ist — beides ist
// eine zusätzliche Wirkung auf die Umgebung, und genau das ist der Wortlaut
// des Protokolls.
func ablauf(ausfuhr Ausfuhr) []Werkzeug {
	return []Werkzeug{
		{
			Name: "schreiben.set.einfahren", Art: Schreibend,
			Erklaerung: "ÄNDERT den Bestand: liest ein Setpaket (.mse-set) hinein und gibt " +
				"seine Kennung. Eine Datei, die schon eingefahren ist, wird nicht ein " +
				"zweites Mal angelegt — der Ruf gibt dann das vorhandene Set.",
			Schema: objekt(map[string]any{
				"pfad": text("Der Pfad zur .mse-set-Datei oder zum Verzeichnis."),
			}, "pfad"),
			Ruf: func(ctx context.Context, roh json.RawMessage) (string, error) {
				var argumente struct {
					Pfad string `json:"pfad"`
				}

				if err := lies(roh, &argumente); err != nil {
					return "", err
				}

				// **Der Teilerfolg reist mit dem Fehler** (#1291, Befund 1
				// des Nachlaufs): ein Archiv, das sich nach dem Einfahren
				// nicht schließen lässt, hinterlässt ein Set im Bestand.
				eingefahren, err := ausfuhr.Einfahren(ctx, argumente.Pfad)
				if err != nil {
					if eingefahren.SetID != 0 {
						return "", fmt.Errorf("%w — angelegt wurde Set %d",
							err, eingefahren.SetID)
					}

					return "", err
				}

				return alsJSON(map[string]any{
					"setID":   eingefahren.SetID,
					"karten":  eingefahren.Karten,
					"befunde": eingefahren.Befunde,
				})
			},
		},
		{
			Name: "schreiben.set.sichern", Art: Schreibend,
			Erklaerung: "ÄNDERT die Platte: schreibt ein Set als .mse-set an einen Ort, an " +
				"dem noch nichts liegt. Ein vorhandenes Ziel wird abgewiesen.",
			Schema: objekt(map[string]any{
				"setID": zahl("Die Kennung des Sets aus »lesen.sets«."),
				"ziel":  text("Der Pfad, an den geschrieben wird — dort darf nichts liegen."),
			}, "setID", "ziel"),
			Ruf: func(ctx context.Context, roh json.RawMessage) (string, error) {
				var argumente struct {
					SetID int64  `json:"setID"`
					Ziel  string `json:"ziel"`
				}

				if err := lies(roh, &argumente); err != nil {
					return "", err
				}

				if err := pruefeKennung("setID", argumente.SetID); err != nil {
					return "", err
				}

				// **Der Teilerfolg reist mit dem Fehler** (#1291, Befund 1
				// des Nachlaufs): scheitert nur das Setzen des Arbeitsorts,
				// **liegt die Datei schon am Ziel**.
				befunde, err := ausfuhr.Sichern(ctx, argumente.SetID, argumente.Ziel)
				if err != nil {
					return "", fmt.Errorf("%w — geschrieben wurde %s", err, argumente.Ziel)
				}

				return alsJSON(map[string]any{"pfad": argumente.Ziel, "befunde": befunde})
			},
		},
		{
			Name: "schreiben.karte.zeichnen", Art: Schreibend,
			Erklaerung: "ÄNDERT die Platte: rastert eine Karte als Bild (.png, .jpg oder " +
				".svg) an einen Ort, an dem noch nichts liegt. Ein vorhandenes Ziel wird " +
				"abgewiesen. guete wählt, wie sorgfältig gerastert wird — wahlfrei, " +
				"und ohne Angabe wie Magic Set Editor 2 selbst.",
			Schema: objekt(map[string]any{
				"karteID": zahl("Die Kennung der Karte aus »lesen.karten«."),
				"ziel": text("Der Pfad, an den geschrieben wird. Die Endung bestimmt " +
					"das Format: .png, .jpg oder .svg."),
				"guete": text("Wie sorgfältig gerastert wird: »fast« für eine schnelle " +
					"Ansicht, »medium« wie MSE2 zeichnet (die Vorgabe), »high« ohne " +
					"Flimmern an feinen Linien, »ultra« für den Druck in voller " +
					"Auflösung. Layout und Zeilenumbruch sind bei allen vier gleich."),
			}, "karteID", "ziel"),
			Ruf: func(ctx context.Context, roh json.RawMessage) (string, error) {
				var argumente struct {
					KarteID int64  `json:"karteID"`
					Ziel    string `json:"ziel"`
					Guete   string `json:"guete"`
				}

				if err := lies(roh, &argumente); err != nil {
					return "", err
				}

				if err := pruefeKennung("karteID", argumente.KarteID); err != nil {
					return "", err
				}

				breite, hoehe, befunde, err := ausfuhr.Zeichnen(ctx,
					argumente.KarteID, argumente.Ziel, argumente.Guete)
				if err != nil {
					return "", err
				}

				return alsJSON(map[string]any{
					"pfad": argumente.Ziel, "breite": breite, "hoehe": hoehe,
					"befunde": befunde,
				})
			},
		},
		{
			Name: "schreiben.set.bogen", Art: Schreibend,
			Erklaerung: "ÄNDERT die Platte: legt die Karten eines Sets als Druckbogen " +
				"(PDF, A4, neun Karten je Seite, 300 dpi, mit Schnittmarken) an einen " +
				"Ort, an dem noch nichts liegt. Ein vorhandenes Ziel wird abgewiesen.",
			Schema: objekt(map[string]any{
				"setID": zahl("Die Kennung des Sets aus »lesen.sets«."),
				"ziel":  text("Der Pfad zur PDF-Datei — dort darf nichts liegen."),
			}, "setID", "ziel"),
			Ruf: func(ctx context.Context, roh json.RawMessage) (string, error) {
				var argumente struct {
					SetID int64  `json:"setID"`
					Ziel  string `json:"ziel"`
				}

				if err := lies(roh, &argumente); err != nil {
					return "", err
				}

				if err := pruefeKennung("setID", argumente.SetID); err != nil {
					return "", err
				}

				seiten, karten, befunde, err := ausfuhr.Bogen(ctx, argumente.SetID, argumente.Ziel)
				if err != nil {
					return "", err
				}

				return alsJSON(map[string]any{
					"pfad": argumente.Ziel, "seiten": seiten, "karten": karten,
					"befunde": befunde,
				})
			},
		},
		{
			Name: "lesen.mtgjson.spiele", Art: Lesend,
			Erklaerung: "Nennt die Spiele, für die es eine Zuordnung von " +
				"MTG-JSON-Feldern gibt. Eines davon braucht »schreiben.mtgjson.lesen«.",
			Schema: objekt(map[string]any{}),
			Ruf: func(context.Context, json.RawMessage) (string, error) {
				return alsJSON(map[string]any{"spiele": ausfuhr.MTGJSONSpiele()})
			},
		},
		{
			Name: "schreiben.mtgjson.lesen", Art: Schreibend,
			Erklaerung: "ÄNDERT den Bestand: liest MTG JSON hinein — eine Datei oder ein " +
				"Verzeichnis einzelner Sets — und gibt die Kennungen der angelegten " +
				"Sets. Eine Quelle, die schon eingefahren ist, wird nicht ein zweites " +
				"Mal angelegt.",
			Schema: objekt(map[string]any{
				"pfad": text("Der Pfad zur .json-Datei oder zum Verzeichnis."),
				"spiel": text("Das Spiel, dessen Zuordnung gilt — eines aus " +
					"»lesen.mtgjson.spiele«."),
			}, "pfad", "spiel"),
			Ruf: func(ctx context.Context, roh json.RawMessage) (string, error) {
				var argumente struct {
					Pfad  string `json:"pfad"`
					Spiel string `json:"spiel"`
				}

				if err := lies(roh, &argumente); err != nil {
					return "", err
				}

				// **Der Teilerfolg reist mit dem Fehler** (#1291, Befund 1).
				// Wer hier `if err != nil { return "", err }` schriebe,
				// verlöre die Sets, die schon im Bestand stehen — und ein
				// Modell räumte sie nie wieder weg.
				eingefahren, err := ausfuhr.MTGJSONLesen(ctx, argumente.Pfad, argumente.Spiel)
				if err != nil {
					return "", fmt.Errorf("%w — angelegt wurden %v", err, eingefahren.Angelegt)
				}

				return alsJSON(eingefahren)
			},
		},
		{
			Name: "schreiben.mtgjson.schreiben", Art: Schreibend,
			Erklaerung: "ÄNDERT die Platte: schreibt ein Set als MTG JSON an einen Ort, " +
				"an dem noch nichts liegt. Ein vorhandenes Ziel wird abgewiesen.",
			Schema: objekt(map[string]any{
				"setID": zahl("Die Kennung des Sets aus »lesen.sets«."),
				"spiel": text("Das Spiel, dessen Zuordnung gilt — eines aus " +
					"»lesen.mtgjson.spiele«."),
				"ziel": text("Der Pfad zur .json-Datei — dort darf nichts liegen."),
			}, "setID", "spiel", "ziel"),
			Ruf: func(ctx context.Context, roh json.RawMessage) (string, error) {
				var argumente struct {
					SetID int64  `json:"setID"`
					Spiel string `json:"spiel"`
					Ziel  string `json:"ziel"`
				}

				if err := lies(roh, &argumente); err != nil {
					return "", err
				}

				if err := pruefeKennung("setID", argumente.SetID); err != nil {
					return "", err
				}

				karten, befunde, err := ausfuhr.MTGJSONSchreiben(
					ctx, argumente.SetID, argumente.Spiel, argumente.Ziel)
				if err != nil {
					return "", err
				}

				return alsJSON(map[string]any{
					"pfad": argumente.Ziel, "karten": karten, "befunde": befunde,
				})
			},
		},
	}
}
