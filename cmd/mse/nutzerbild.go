// **Das Profilbild an der Kommandozeile** — Werkstück 8 von Phase 4.12
// (#1351).
//
// # Ein Befehl mit zwei Gestalten und nicht zwei Befehle
//
// **Weil es dieselbe Frage ist**: welches Bild hat dieser Mensch? Ein Pfad
// setzt eines, ein `--keines` nimmt es weg. Zwei Namen wären zwei Dinge zum
// Merken für einen Zustand mit zwei Werten.
//
// # Und ein Weg für zwei Befehle
//
// `nutzer-bild` und `profil-badge` haben dieselbe Gestalt (Werkstück 9) und
// standen bis ENG-03 (#1728)
// zweimal da — **mit demselben Fehler zweimal**: beide riefen das Setzen mit
// dem Pfad `--keines` auf, bevor sie wegnahmen. Lag eine Datei dieses Namens
// im Arbeitsverzeichnis, wurde sie gesetzt, die Fassung zählte hoch, und das
// Wegnehmen scheiterte an ihr. Die Wahl steht jetzt einmal, in [bildwechsel].

package main

import (
	"context"
	"database/sql"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// Keinbild ist die Angabe, die das Bild wegnimmt.
const Keinbild = "--keines"

// bildweg ist, was die zwei Bildbefehle unterscheidet — der Wortlaut ihrer
// Klage und die zwei Rufe an die Fläche.
type bildweg struct {
	klage     string
	setzen    func(ctx context.Context, flaeche *app.Nutzerflaeche, kennung, fassung, pfad string) bildantwort
	entfernen func(ctx context.Context, flaeche *app.Nutzerflaeche, kennung, fassung string) bildantwort
}

// bildantwort ist die Antwort beider Wege, so weit die Ausgabe sie braucht.
type bildantwort struct {
	roh           any
	fehler        *app.Fehler
	kennung, bild string
}

// nutzerbildbefehl setzt oder entfernt das Profilbild eines Nutzers.
func nutzerbildbefehl(ctx context.Context, bestandsort string, an *schreiber, rest []string) int {
	return bildwechsel(ctx, bestandsort, an, rest, bildweg{
		klage: "nutzer-bild braucht Kennung, Pfad und Fassung — " +
			"%s als Pfad nimmt das Bild weg.\n%s",
		setzen: func(ctx context.Context, flaeche *app.Nutzerflaeche, kennung, fassung, pfad string) bildantwort {
			return nutzerbild(flaeche.NutzerbildSetzen(ctx, kennung, fassung, pfad))
		},
		entfernen: func(ctx context.Context, flaeche *app.Nutzerflaeche, kennung, fassung string) bildantwort {
			return nutzerbild(flaeche.NutzerbildEntfernen(ctx, kennung, fassung))
		},
	})
}

// profilbadgebefehl setzt oder entfernt das Abzeichen eines Profils.
//
// **Dieselbe Gestalt wie `nutzer-bild`** — Werkstück 9 (#1351): ein Pfad
// setzt eines, `--keines` nimmt es weg. Ein Abzeichen gehört einem **Profil**
// und liegt über dem Bild des Menschen; deshalb ist die Kennung eine
// Profilkennung und die Fassung die **des Profils**.
func profilbadgebefehl(ctx context.Context, bestandsort string, an *schreiber, rest []string) int {
	return bildwechsel(ctx, bestandsort, an, rest, bildweg{
		klage: "profil-badge braucht Profilkennung, Pfad und Fassung — " +
			"%s als Pfad nimmt das Abzeichen weg.\n%s",
		setzen: func(ctx context.Context, flaeche *app.Nutzerflaeche, kennung, fassung, pfad string) bildantwort {
			return abzeichen(flaeche.ProfilbadgeSetzen(ctx, kennung, fassung, pfad))
		},
		entfernen: func(ctx context.Context, flaeche *app.Nutzerflaeche, kennung, fassung string) bildantwort {
			return abzeichen(flaeche.ProfilbadgeEntfernen(ctx, kennung, fassung))
		},
	})
}

// bildwechsel setzt ein Bild **oder** nimmt es weg — nie beides.
func bildwechsel(ctx context.Context, bestandsort string, an *schreiber, rest []string, weg bildweg) int {
	if len(rest) != 3 {
		an.klage(weg.klage, Keinbild, hilfe)

		return CodeBedienung
	}

	kennung, pfad, fassung := rest[0], rest[1], rest[2]

	return mitBestand(ctx, bestandsort, an, func(ctx context.Context, db *sql.DB) int {
		flaeche := app.NutzerflaecheUeber(db)

		var antwort bildantwort
		if pfad == Keinbild {
			antwort = weg.entfernen(ctx, flaeche, kennung, fassung)
		} else {
			antwort = weg.setzen(ctx, flaeche, kennung, fassung, pfad)
		}

		if antwort.fehler != nil {
			an.klageUeberFehler(antwort.fehler)

			return CodeFachlich
		}

		if an.alsJSON {
			return alsJSONSchreiben(an, antwort.roh)
		}

		// **Der Dateiname und nicht das Bild** — ein Daten-URI an der
		// Kommandozeile wäre eine Zeichenkette, die niemand ansieht.
		an.sag("%s\t%s\n", antwort.kennung, antwort.bild)

		return an.abschluss()
	})
}

// nutzerbild liest die Antwort am Bild eines Nutzers.
func nutzerbild(antwort app.EinNutzerAntwort) bildantwort {
	gelesen := bildantwort{roh: antwort, fehler: antwort.Fehler}
	if antwort.Nutzer != nil {
		gelesen.kennung, gelesen.bild = antwort.Nutzer.ID, antwort.Nutzer.Bild
	}

	return gelesen
}

// abzeichen liest die Antwort am Abzeichen eines Profils.
func abzeichen(antwort app.EinProfilAntwort) bildantwort {
	gelesen := bildantwort{roh: antwort, fehler: antwort.Fehler}
	if antwort.Profil != nil {
		gelesen.kennung, gelesen.bild = antwort.Profil.ID, antwort.Profil.Badge
	}

	return gelesen
}
