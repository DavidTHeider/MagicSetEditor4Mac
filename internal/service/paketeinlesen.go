package service

// **Einen `.mse-installer` einlesen** — FMT-02.AK-04
// (#1691).
//
// # Was aus MSE2 übernommen ist (ported)
//
// Welche Dateien zu welchem Paket gehören und wohin sie kommen —
// `PackageDirectory::actual_install` (`src/util/io/package_manager.cpp`,
// Zeilen 376–400, Fassung f9d9356 vom 2021-07-06); das Lesen selbst steht in
// `msepaket`.
//
// # Wo es anders ist (adapted)
//
// **Ein vorhandenes Paket wird übergangen und nicht ersetzt.** MSE2 bietet
// im Paketfenster an, es zu ersetzen, wenn der Installer eine neuere oder
// gleiche Fassung trägt (`InstallablePackage::determineStatus`,
// `src/data/installer.cpp`, Zeilen 288–312). Hier gilt dieselbe Regel wie
// beim Installieren und bei der Übernahme: wer eine Vorlage bearbeitet hat,
// verlöre sie sonst — wer ersetzen will, entfernt zuerst.
//
// **Und es gibt Grenzen**, die MSE2 nicht kennt: Einträge und ausgepackte
// Bytes, bevor ein fremdes Archiv den Arbeitsspeicher oder die Platte füllt.
// Die Zahlen sind die des Bezugs (4 GiB); die Zahl der Einträge trägt den
// ganzen `full-magic-pack` viermal (34.683 Dateien, gemessen am 2026-09-29).

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
)

// installerdeckel ist, wie viele Bytes ein Installer auspacken darf.
var installerdeckel int64 = 4 << 30

// installereintraege ist, wie viele Einträge ein Installer tragen darf.
var installereintraege = 1 << 17

// Einlesen legt die Pakete eines Installers ins Paketverzeichnis.
//
// **Ein Teilerfolg ist ein Erfolg mit Befunden** — dieselbe Gestalt wie
// [Paketverwalter.Uebernehmen].
func (v *Paketverwalter) Einlesen(ctx context.Context, quelle string) (*Uebernahme, error) {
	if err := v.ortPruefen(); err != nil {
		return nil, err
	}

	installer, err := installerOeffnen(quelle)
	if err != nil {
		return nil, err
	}

	defer func() { _ = installer.Schliessen() }() //nolint:errcheck // nur gelesen

	ergebnis := &Uebernahme{Uebernommen: []domain.Paket{}, Uebergangen: []string{}, Befunde: []domain.Befund{}}

	for _, befund := range installer.Befunde {
		ergebnis.Befunde = append(ergebnis.Befunde, domain.Befund{Datei: befund.Stelle.Datei, Meldung: befund.Meldung})
	}

	var vermerke []Vermerk

	rest := installerdeckel

	// **Ein Abbruch fällt zwischen zwei Pakete** (FMT-03.AK-04): was davor
	// ins Verzeichnis kam, ist ganz und wird unten vermerkt — ein Paket ohne
	// Herkunft wäre ein halbes.
	var abbruch error

	for nummer, paket := range installer.Pakete {
		if abbruch = Schritt(ctx, domain.Fortschritt{
			Erledigt: nummer, Gesamt: len(installer.Pakete), Einheit: domain.EinheitPaket,
		}); abbruch != nil {
			break
		}

		vermerk, benutzt, befund := v.einlesenEines(installer, paket.Name, rest)
		rest -= benutzt

		switch {
		case befund != nil:
			ergebnis.Befunde = append(ergebnis.Befunde, *befund)
		case vermerk == nil:
			ergebnis.Uebergangen = append(ergebnis.Uebergangen, paket.Name)
		default:
			vermerke = append(vermerke, *vermerk)
		}
	}

	// **Ab hier hält kein Abbruch mehr auf** — die Pakete liegen im
	// Verzeichnis; dieselbe Gestalt wie bei der Übernahme.
	ohneAbbruch := context.WithoutCancel(ctx)

	befunde, err := v.Vermerken(ohneAbbruch, vermerke...)
	ergebnis.Befunde = append(ergebnis.Befunde, befunde...)

	if err != nil {
		return ergebnis, err
	}

	ergebnis.Uebernommen = v.uebernommene(ohneAbbruch, vermerke)

	if abbruch != nil {
		return ergebnis, abbruch
	}

	Melde(ctx, domain.Fortschritt{
		Erledigt: len(installer.Pakete), Gesamt: len(installer.Pakete), Einheit: domain.EinheitPaket,
	})

	return ergebnis, nil
}

// einlesenEines packt ein Paket in einen Bauort, liest es und kopiert es ins
// Verzeichnis. Ohne Vermerk und ohne Befund lag es schon da.
//
// **Wie beim Installieren**: erst nachsehen, dann kopieren — und
// `kopierePaket` baut neben dem Ziel und benennt um; käme ihm jemand
// zuvor, scheiterte das Umbenennen, statt zwei Pakete zu mischen.
func (v *Paketverwalter) einlesenEines(installer *msepaket.Installer, name string, rest int64) (*Vermerk, int64, *domain.Befund) {
	ziel := filepath.Join(v.verwaltet, name)
	if _, err := os.Stat(ziel); err == nil {
		return nil, 0, nil
	}

	bau, ort, err := wegwerfpaket(v.bauort(), "mse4-einlesen-", name)
	if err != nil {
		return nil, 0, einlesebefund(name, "für das Paket fand sich kein Bauort", err)
	}

	defer func() { _ = os.RemoveAll(bau) }() //nolint:errcheck // Wegwerfverzeichnis

	benutzt, err := installer.Auspacken(name, ort, rest)
	if errors.Is(err, msepaket.ErrInstallerZuGross) {
		return nil, benutzt, einlesebefund(name,
			"der Installer ist größer als erlaubt — dieses Paket und alle danach blieben draußen", err)
	} else if err != nil {
		return nil, benutzt, einlesebefund(name, "das Paket ließ sich nicht auspacken", err)
	}

	// **Erst lesen, dann kopieren** — ein Paket, das sich nicht öffnen lässt,
	// soll gar nicht erst im Verzeichnis landen.
	if _, gelesen := lies(ort, name, nil); gelesen != nil {
		return nil, benutzt, gelesen
	}

	if err := kopierePaket(ort, ziel); err != nil {
		return nil, benutzt, einlesebefund(name, "das Paket ließ sich nicht ins Verzeichnis legen", err)
	}

	return &Vermerk{Ort: ziel, Herkunft: installer.Pfad}, benutzt, nil
}

// einlesebefund macht aus einem Fehler einen Befund — **der Satz für
// Menschen, der Grund fürs Protokoll** (0075): ein Pfad samt „permission
// denied" sagt dem, der ein Paket einliest, nichts.
func einlesebefund(name, satz string, grund error) *domain.Befund {
	slog.Warn("ein Paket aus einem Installer blieb draußen", "datei", name, "vorgang", satz, "grund", grund)

	return &domain.Befund{Datei: name, Meldung: satz}
}

// installerOeffnen prüft die Quelle und öffnet sie — **die Einträge werden
// gezählt, bevor `archive/zip` sie liest**: es legt je Kopf eine Struktur an,
// und die Zahl im Verzeichnisende schützt davor nicht.
func installerOeffnen(quelle string) (*msepaket.Installer, error) {
	angaben, err := os.Stat(quelle)
	if err != nil {
		return nil, errors.Join(
			domain.NewFieldError("quelle", domain.SchluesselValidierung,
				"%q ließ sich nicht lesen — prüfe, ob die Datei noch dort liegt.", quelle),
			domain.MitGrund(domain.ErrValidation, err, "den Installer ansehen"))
	}

	if angaben.IsDir() {
		return nil, domain.NewFieldError("quelle", domain.SchluesselValidierung,
			"%q ist ein Ordner und kein Installer.", quelle)
	}

	eintraege, _, unlesbar := archivEintraege(quelle, angaben.Size(), installereintraege+1)
	if unlesbar {
		return nil, domain.NewFieldError("quelle", domain.SchluesselValidierung,
			"%q ist kein lesbares Archiv — ein Installer ist ein ZIP.", quelle)
	}

	if len(eintraege) > installereintraege {
		return nil, domain.NewFieldError("quelle", domain.SchluesselValidierung,
			"%q trägt mehr als %d Einträge — so viele hat kein Vorlagenpaket.", quelle, installereintraege)
	}

	installer, err := msepaket.OeffneInstaller(quelle)
	if err != nil {
		return nil, errors.Join(
			domain.NewFieldError("quelle", domain.SchluesselValidierung,
				"%q ist kein Installer — ihm fehlt die Datei „installer“, die seine Pakete nennt.", quelle),
			domain.MitGrund(domain.ErrValidation, err, "den Installer öffnen"))
	}

	return installer, nil
}
