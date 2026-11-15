package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"path/filepath"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/scryfall"
)

// Fremdbilder ist, was der Bildweg von einem fremden Dienst braucht.
//
// **Getrennt von [Fremdkarten]**, obwohl derselbe Klient beide erfüllt: ein
// Aufrufer, der nur Felder abgleicht, soll keine Bildfähigkeit binden müssen.
type Fremdbilder interface {
	Bild(ctx context.Context, karte *scryfall.Karte, art scryfall.Bildart) (
		scryfall.Bildfund, error)
}

// Bilduebernahme ist ein geholtes Bild, das als Datei bereitliegt.
//
// **Eine Datei und kein Byteblock**, und das ist die eigentliche Verdrahtung:
// `BildAusDatei` und `BildEinsetzen` nehmen einen **Pfad**. Wer hier Bytes
// zurückgäbe, hätte einen Weg gebaut, der vor der vorhandenen Naht endet —
// und Rasterung, Druckbogen und `.mse-set`-Ausfuhr sähen das Bild nie.
type Bilduebernahme struct {
	Pfad      string
	Art       scryfall.Bildart
	Kuenstler string

	// Nennung ist der Satz, der in derselben Oberfläche stehen muss.
	Nennung string

	// Adresse ist die Kartenseite bei Scryfall — die „source" der Bedingung.
	Adresse string

	// Alter sagt, wie alt die Kartenantwort war (§16, Regel 5).
	Alter time.Duration

	// Befunde ist, was die Kartenauskunft tolerant hingenommen hat.
	//
	// **Sie werden weitergereicht und nicht verschluckt** (§16: „nichts wird
	// still übernommen"). `KarteNachName` gibt sie, und der Feldabgleich
	// reicht dieselben Befunde durch.
	Befunde []domain.Befund
}

// Bildweg holt ein Kartenbild und legt es als Datei ab.
type Bildweg struct {
	haushaltswahl

	karten Fremdkarten
	bilder Fremdbilder
	ordner string
}

// NeuerBildweg baut den Bildweg.
func NeuerBildweg(karten Fremdkarten, bilder Fremdbilder, ordner string) *Bildweg {
	return &Bildweg{karten: karten, bilder: bilder, ordner: ordner}
}

// Holen sucht die Karte, holt ihr Bild und legt es ab.
//
// # Die Wache, und warum sie vor dem Schreiben steht
//
// **Ein Bild ohne Nennung wird nicht abgelegt.** Nicht erst nicht angezeigt —
// gar nicht erst geschrieben: eine Datei auf der Platte findet der Nächste
// ohne die Bedingung, die zu ihr gehört, und dann ist die Nennung verloren,
// ohne dass es jemand merkt.
//
// # Warum zwei Abrufe kein zweiter Netzweg sind
//
// Die Karte kommt aus demselben Zwischenspeicher wie beim Feldabgleich
// (24 Stunden, §16). Wer erst abgleicht und dann das Bild holt, fragt Scryfall
// **einmal** nach der Karte.
func (b *Bildweg) Holen(ctx context.Context, name string, art scryfall.Bildart) (
	*Bilduebernahme, error,
) {
	freigeben, platzfehler := b.platz(ctx, domain.RessourceBild)
	if platzfehler != nil {
		return nil, platzfehler
	}

	defer freigeben()

	karte, alter, befunde, fehler := b.karten.KarteNachName(ctx, name)
	if fehler != nil {
		return nil, fehler
	}

	// **Leer und nicht nil.** Eine Fläche, die zählt, unterscheidet sonst
	// „keine Befunde" von „das Feld fehlte" — und das JSON eine leere Liste
	// von `null`.
	if befunde == nil {
		befunde = []domain.Befund{}
	}

	fund, bildfehler := b.bilder.Bild(ctx, karte, art)
	if bildfehler != nil {
		return nil, bildfehler
	}

	if len(fund.Daten) == 0 {
		return nil, fmt.Errorf("%w: das Bild zu %q ist leer", domain.ErrValidation, name)
	}

	if fund.Nennung == "" {
		return nil, fmt.Errorf(
			"%w: zum Bild von %q fehlt die Nennung — ein Bild ohne Künstler und "+
				"Copyright darf die Anwendung nicht verlassen", domain.ErrValidation, name)
	}

	pfad, schreibfehler := b.ablegen(fund)
	if schreibfehler != nil {
		return nil, schreibfehler
	}

	return &Bilduebernahme{
		Pfad: pfad, Art: fund.Art, Kuenstler: fund.Kuenstler,
		Nennung: fund.Nennung, Adresse: fund.Adresse, Alter: alter,
		Befunde: befunde,
	}, nil
}

// ablegen schreibt die Bytes in eine Datei.
//
// **Der Name kommt aus den Bytes**, nicht aus dem Kartennamen: ein Kartenname
// enthält Schrägstriche (`Delver of Secrets // Insectile Aberration`), und
// derselbe Inhalt soll nicht zweimal liegen.
func (b *Bildweg) ablegen(fund scryfall.Bildfund) (string, error) {
	summe := sha256.Sum256(fund.Daten)
	pfad := filepath.Join(b.ordner, hex.EncodeToString(summe[:])+endung(fund.Daten))

	if fehler := schreibeAtomar(pfad, fund.Daten); fehler != nil {
		return "", fmt.Errorf("das Bild ablegen: %w", fehler)
	}

	return pfad, nil
}

// endung rät die Dateiendung aus den ersten Bytes.
//
// **Aus dem Inhalt und nicht aus der Adresse**: Scryfall liefert `art_crop`
// als JPEG und die `png`-Größe als PNG, aber eine Adresse kann eine
// Weiterleitung sein. Die Endung ist für den Menschen, der die Datei später
// sieht — der Leser der Datei prüft die Bytes.
func endung(daten []byte) string {
	if len(daten) >= 8 && string(daten[:8]) == "\x89PNG\r\n\x1a\n" {
		return ".png"
	}

	return ".jpg"
}
