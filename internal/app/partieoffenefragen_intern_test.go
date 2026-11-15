package app

// **Was am Tisch öffentlich ist, kommt ungekürzt durch die Naht** — die
// offenen Fragen und der Kampfstand.
//
// # Warum das einen eigenen Lauf braucht
//
// **Weil ein Feld, das an dieser Naht wegfällt, die Antwort vollständiger
// aussehen lässt und nicht falscher** (§3, die Naht zwischen den
// Prüfebenen). Genau das ist bei `M6` geschehen: der Kampfstand lag im
// Zugstand, **kein DTO trug ihn weiter**, und an keiner der vier Flächen
// stand, wer angreift — gefunden erst bei
// #1623.
//
// **Und eine offene Frage, die nicht ankommt, hält das Spiel an, ohne zu
// sagen warum**: CR 704.5j und CR 603.3 lassen beide die Partie stehen, bis
// jemand geantwortet hat.

import "testing"

// **Die Legendenwahl, die Auslöser und der Kampfstand kommen an.**
func TestDieOffenenFragenUndDerKampfstandKommenAn(t *testing.T) {
	t.Parallel()

	anwendung := NeueApp("")
	anwendung.db = bestandFuerDenRiegel(t)
	anwendung.profil.Store(1)
	anwendung.MitPartienaht(&mischpflichtigeLeitung{})

	antwort := NeueBindungen(anwendung).Partie(t.Context(), "p1", "sitz-1")
	if antwort.Fehler != nil {
		t.Fatalf("Partie: %+v", antwort.Fehler)
	}

	partie := antwort.Partie

	if len(partie.Legendenwahlen) != 1 ||
		partie.Legendenwahlen[0].Name != "Held" {
		t.Errorf("Legendenwahlen: %+v", partie.Legendenwahlen)
	}

	// **Die Reihung nennt den Sitz und die Fähigkeit, über die er
	// entscheidet** — CR 603.3b. Ohne die Kennung ließe sich der Befehl
	// `spieler.reihung` nicht stellen.
	if partie.Ausloeserreihung != "sitz-2" {
		t.Errorf("geordnet wird von %q", partie.Ausloeserreihung)
	}

	if len(partie.Ausgeloest) != 1 {
		t.Fatalf("Ausgeloest: %+v", partie.Ausgeloest)
	}

	einer := partie.Ausgeloest[0]
	if einer.Kennung != "sitz-2/monarch.zug/0" || einer.Art != "monarch.zug" ||
		einer.Beherrscher != "sitz-2" || einer.Wen != "sitz-1" {
		t.Errorf("Ausloeser: %+v", einer)
	}

	// **Und der Kampfstand** — CR 508.1a, CR 509.1a.
	if len(partie.Zug.Angriffe) != 1 ||
		partie.Zug.Angriffe[0].Kreatur != "sitz-1/baer/0" ||
		partie.Zug.Angriffe[0].Ziel != "sitz-2" {
		t.Errorf("Angriffe: %+v", partie.Zug.Angriffe)
	}

	if len(partie.Zug.Bloecke) != 1 ||
		partie.Zug.Bloecke[0].Kreatur != "sitz-2/wand/0" ||
		partie.Zug.Bloecke[0].Angreifer != "sitz-1/baer/0" {
		t.Errorf("Bloecke: %+v", partie.Zug.Bloecke)
	}
}
