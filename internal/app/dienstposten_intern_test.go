package app

// **Was ein Dienstposten zusagt** — Werkstück N6 von Phase 5.4
// (#1572).
//
// # Über eine eingesetzte Naht und nicht über ein Programm
//
// **Weil die Zusagen hier welche über den Posten sind** und nicht über
// `mse-rest`: dass nichts von selbst startet, dass ein zweiter Start abgewiesen
// wird, dass ein Befund bei den Adressen den Port nicht wieder zumacht. Dass
// ein echtes Programm wirklich lauscht, misst `internal/flaechen` — mit
// gebautem Programm und echter Anfrage.

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/leitung"
)

// probelauf ist eine Ansage, wie ein echter Dienst sie machte.
func probelauf(adresse, marke, abdruck string) dienstlauf {
	fertig := make(chan struct{})

	return dienstlauf{
		adresse:       adresse,
		marke:         marke,
		fingerabdruck: abdruck,
		beenden:       func() { close(fertig) },
		fertig:        fertig,
	}
}

// einePostenhalterin baut eine Anwendung mit eingesetztem Starter.
//
// **Der Zähler sagt, ob überhaupt gestartet wurde** — ohne ihn sähe eine
// Abweisung genauso aus wie ein Start, der stillschweigend nichts tat.
//
// **Je Ruf eine frische Ansage**, denn zwei Posten teilen sich keinen Kanal —
// ein gemeinsamer endete beim zweiten Beenden in einem `close` auf einem
// geschlossenen Kanal.
func einePostenhalterin(
	t *testing.T, adresse, marke, abdruck string, fehler error,
) (*App, *int, context.Context) {
	t.Helper()

	gestartet := 0
	anwendung := NeueApp("")

	anwendung.MitDienststarter(func(
		_ context.Context, _ domain.Dienstart, _ []domain.Bindung,
	) (dienstlauf, error) {
		gestartet++

		if fehler != nil {
			return dienstlauf{}, fehler
		}

		return probelauf(adresse, marke, abdruck), nil
	})

	anwendung.MitErreichbarkeit(func(
		_ context.Context, horchadresse string,
	) (leitung.Erreichbarkeit, error) {
		if strings.HasPrefix(horchadresse, "https://") {
			t.Errorf("gefragt wurde mit Schema: %q", horchadresse)
		}

		return leitung.Erreichbarkeit{
			Namen:    []string{"mein-mac.fritz.box:54591"},
			Adressen: []string{"192.168.178.51:54591"},
		}, nil
	})

	t.Cleanup(anwendung.diensteBeenden)

	return anwendung, &gestartet, context.Background()
}

// **Von selbst läuft keiner der zwei** (0115).
func TestVonSelbstLaeuftKeinDienst(t *testing.T) {
	t.Parallel()

	anwendung, _, _ := einePostenhalterin(t, "", "", "", nil)

	for _, art := range domain.Dienstarten() {
		stand := anwendung.dienststand(art)

		if stand.Laeuft || stand.Adresse != "" || stand.Marke != "" {
			t.Errorf("%s läuft ungefragt: %+v", art, stand)
		}

		// **Leere Listen und kein `nil`** — sonst steht im Fenster
		// `namen: null`, und jede Schleife darüber fällt um.
		if stand.Namen == nil || stand.Adressen == nil {
			t.Errorf("%s gibt nil statt einer leeren Liste: %+v", art, stand)
		}
	}
}

// **Jenseits der Schleife steht alles im Stand, was ein Gegenüber braucht.**
//
// **Der Fingerabdruck gehört dazu**, und das ist die Zusage, die 0115 macht:
// das Zeugnis ist selbst unterschrieben, und wer sich verbindet, hat nichts
// anderes zu vergleichen. Ein Feld, das leer bleibt, wäre ein Kopf ohne Rumpf
// (§10).
func TestDerStandNenntAdresseMarkeUndFingerabdruck(t *testing.T) {
	t.Parallel()

	anwendung, gestartet, ctx := einePostenhalterin(t,
		"https://0.0.0.0:54591", "geheim", "AB:CD:EF", nil)

	stand, fehler := anwendung.dienstStarten(ctx, domain.DienstREST,
		domain.Bindung{Adresse: domain.JedeSchnittstelle})
	if fehler != nil {
		t.Fatalf("dienstStarten: %v", fehler)
	}

	if *gestartet != 1 {
		t.Errorf("das Programm wurde %dmal gestartet", *gestartet)
	}

	if !stand.Laeuft || stand.Adresse != "https://0.0.0.0:54591" ||
		stand.Marke != "geheim" || !stand.TLS {
		t.Errorf("der Stand ist unvollständig: %+v", stand)
	}

	if stand.Fingerabdruck != "AB:CD:EF" {
		t.Errorf("der Fingerabdruck fehlt: %q", stand.Fingerabdruck)
	}

	if len(stand.Namen) != 1 || len(stand.Adressen) != 1 {
		t.Errorf("die Liste zum Weitergeben fehlt: %+v", stand)
	}

	// **Und derselbe Stand kommt auch auf die zweite Frage** — er steht am
	// Posten und nicht in der Antwort des Starts.
	if nochmal := anwendung.dienststand(
		domain.DienstREST); !reflect.DeepEqual(nochmal, stand) {
		t.Errorf("dienststand gibt etwas anderes: %+v", nochmal)
	}
}

// **Auf der Schleife wird nicht nach Adressen gefragt** — es gibt keine.
func TestAufDerSchleifeBleibtDieListeLeer(t *testing.T) {
	t.Parallel()

	anwendung := NeueApp("")

	anwendung.MitDienststarter(func(
		_ context.Context, _ domain.Dienstart, _ []domain.Bindung,
	) (dienstlauf, error) {
		return probelauf("http://127.0.0.1:8123", "geheim", ""), nil
	})

	anwendung.MitErreichbarkeit(func(
		context.Context, string,
	) (leitung.Erreichbarkeit, error) {
		t.Error("auf der Schleife wurde nach Adressen gefragt")

		return leitung.Erreichbarkeit{}, nil
	})

	t.Cleanup(anwendung.diensteBeenden)

	stand, fehler := anwendung.dienstStarten(context.Background(),
		domain.DienstMCP, domain.Bindung{Adresse: domain.Schleifenadresse})
	if fehler != nil {
		t.Fatalf("dienstStarten: %v", fehler)
	}

	if stand.TLS {
		t.Errorf("auf der Schleife ist TLS angesagt: %+v", stand)
	}

	if len(stand.Namen) != 0 || len(stand.Adressen) != 0 {
		t.Errorf("die Liste zum Weitergeben ist gefüllt: %+v", stand)
	}
}

// **Ein Befund bei den Adressen hält den Dienst nicht auf.**
//
// **Die Gegenprobe zum Lauf darüber**: ohne sie stünde nur, dass eine Liste
// ankommt — nicht, dass ihr Ausbleiben den Port offen lässt.
func TestEinBefundBeiDenAdressenLaesstDenPortOffen(t *testing.T) {
	t.Parallel()

	anwendung := NeueApp("")

	anwendung.MitDienststarter(func(
		_ context.Context, _ domain.Dienstart, _ []domain.Bindung,
	) (dienstlauf, error) {
		return probelauf("https://0.0.0.0:54591", "geheim", "AB:CD"), nil
	})

	anwendung.MitErreichbarkeit(func(
		context.Context, string,
	) (leitung.Erreichbarkeit, error) {
		return leitung.Erreichbarkeit{}, errors.New("keine Schnittstellen")
	})

	t.Cleanup(anwendung.diensteBeenden)

	stand, fehler := anwendung.dienstStarten(context.Background(),
		domain.DienstREST, domain.Bindung{Adresse: domain.JedeSchnittstelle})
	if fehler != nil {
		t.Fatalf("dienstStarten: %v", fehler)
	}

	if !stand.Laeuft || stand.Marke != "geheim" {
		t.Errorf("der Befund hat den Posten umgeworfen: %+v", stand)
	}

	if len(stand.Namen) != 0 || len(stand.Adressen) != 0 {
		t.Errorf("woher kommt die Liste? %+v", stand)
	}
}

// **Der zweite Start wird abgewiesen, nicht doppelt gefahren.**
func TestDerZweiteStartIstEinKonflikt(t *testing.T) {
	t.Parallel()

	anwendung, gestartet, ctx := einePostenhalterin(t,
		"http://127.0.0.1:8123", "geheim", "", nil)

	if _, fehler := anwendung.dienstStarten(ctx, domain.DienstREST,
		domain.Bindung{Adresse: domain.Schleifenadresse}); fehler != nil {
		t.Fatalf("der erste Start: %v", fehler)
	}

	_, fehler := anwendung.dienstStarten(ctx, domain.DienstREST,
		domain.Bindung{Adresse: domain.Schleifenadresse})
	if !errors.Is(fehler, domain.ErrConflict) {
		t.Fatalf("der zweite Start gab %v", fehler)
	}

	if *gestartet != 1 {
		t.Errorf("das Programm lief %dmal", *gestartet)
	}

	// **Und der andere Dienst ist davon unberührt** — zwei Posten und nicht
	// einer mit zwei Namen.
	if _, fehler := anwendung.dienstStarten(ctx, domain.DienstMCP,
		domain.Bindung{Adresse: domain.Schleifenadresse}); fehler != nil {
		t.Errorf("MCP kam nicht durch: %v", fehler)
	}
}

// **Eine unmögliche Bindung fährt gar nichts.**
func TestEineUnmoeglicheBindungFaehrtNichts(t *testing.T) {
	t.Parallel()

	anwendung, gestartet, ctx := einePostenhalterin(t, "", "", "", nil)

	_, fehler := anwendung.dienstStarten(ctx, domain.DienstREST,
		domain.Bindung{Adresse: "localhost", Port: 8123})
	if !errors.Is(fehler, domain.ErrValidation) {
		t.Fatalf("ein Name als Adresse gab %v", fehler)
	}

	if *gestartet != 0 {
		t.Errorf("trotzdem gestartet: %dmal", *gestartet)
	}
}

// **Ein scheiterndes Programm kommt als Fehler und hinterlässt keinen Posten.**
func TestEinScheiterndesProgrammHinterlaesstKeinenPosten(t *testing.T) {
	t.Parallel()

	anwendung, _, ctx := einePostenhalterin(t, "", "", "",
		errors.New("mse-rest gibt es hier nicht"))

	if _, fehler := anwendung.dienstStarten(ctx, domain.DienstREST,
		domain.Bindung{Adresse: domain.Schleifenadresse}); fehler == nil {
		t.Fatal("der Fehlschlag kam nicht an")
	}

	if stand := anwendung.dienststand(domain.DienstREST); stand.Laeuft {
		t.Errorf("ein Posten steht trotzdem: %+v", stand)
	}
}

// **Beenden räumt wirklich auf — und ein zweites Mal tut nichts.**
func TestBeendenRaeumtDenPosten(t *testing.T) {
	t.Parallel()

	anwendung, _, ctx := einePostenhalterin(t,
		"https://0.0.0.0:54591", "geheim", "AB:CD", nil)

	if _, fehler := anwendung.dienstStarten(ctx, domain.DienstREST,
		domain.Bindung{Adresse: domain.JedeSchnittstelle}); fehler != nil {
		t.Fatalf("dienstStarten: %v", fehler)
	}

	anwendung.dienstBeenden(domain.DienstREST)

	stand := anwendung.dienststand(domain.DienstREST)
	if stand.Laeuft || stand.Adresse != "" || stand.Marke != "" ||
		stand.Fingerabdruck != "" || len(stand.Namen) != 0 {
		t.Errorf("etwas blieb stehen: %+v", stand)
	}

	// **Zweimal beenden ist kein Fehler** — und kein zweites `close`.
	anwendung.dienstBeenden(domain.DienstREST)
}

// **`ohneSchema` nimmt vorn weg, was [leitung.ErreichbarUnter] nicht kennt.**
func TestOhneSchema(t *testing.T) {
	t.Parallel()

	faelle := map[string]string{
		"https://0.0.0.0:54591": "0.0.0.0:54591",
		"http://127.0.0.1:8123": "127.0.0.1:8123",
		"[::]:41234":            "[::]:41234",
	}

	for gegeben, erwartet := range faelle {
		t.Run(gegeben, func(t *testing.T) {
			t.Parallel()

			if gekuerzt := ohneSchema(gegeben); gekuerzt != erwartet {
				t.Errorf("%q wurde zu %q statt zu %q",
					gegeben, gekuerzt, erwartet)
			}
		})
	}
}
