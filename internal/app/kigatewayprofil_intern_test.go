package app

// **Ein Gateway auf diesem Rechner rechnet auf diesem Rechner** — Befund 3
// des Zweitblicks über Phase D, dieselbe Klasse wie der Proxybefund aus D06.
//
// [App.dienstadresse] las die eingestellte Adresse mit
// `context.Background()`; der Speicher klemmt auf das Profil des Kontexts,
// fand keines, und ein Lesefehler hieß „keine Adresse". Die Freigabe eines
// KI-Auftrags stufte damit ein Gateway auf der Schleife als Wolke ein — und
// beschrieb einen anderen Ort als den, an den der Auftrag ging.

import (
	"path/filepath"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

func TestEinGatewayAufDerSchleifeRechnetAufDiesemRechner(t *testing.T) {
	ctx := t.Context()

	db, fehler := storage.Oeffnen(ctx, filepath.Join(t.TempDir(), "bestand.db"))
	if fehler != nil {
		t.Fatalf("öffnen: %v", fehler)
	}

	t.Cleanup(func() { _ = db.Close() })

	if err := storage.Migriere(ctx, db, nil); err != nil {
		t.Fatalf("migrieren: %v", err)
	}

	imProfil, profilfehler := storage.MitGezeigtemProfil(ctx, db)
	if profilfehler != nil {
		t.Fatalf("das Profil: %v", profilfehler)
	}

	const gateway = "http://127.0.0.1:8080"

	if err := storage.NeueEinstellungen(db).Setzen(imProfil,
		domain.DienstAdresse(domain.Dienst("openai")), gateway); err != nil {
		t.Fatalf("die Adresse setzen: %v", err)
	}

	erwartet := domain.OrtFuerDienstadresse(gateway)
	if erwartet == domain.OrtFuerDienstadresse("") {
		t.Fatal("die Probe trennt nicht: ein Gateway auf der Schleife gilt wie keines")
	}

	if ort := zielOrt(t, PeerflaecheUeber(db), imProfil, "openai"); ort != erwartet {
		t.Errorf("der Ort ist %v — erwartet %v für %s", ort, erwartet, gateway)
	}
}
