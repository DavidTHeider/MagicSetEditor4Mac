package service_test

// **Der Lebenszyklus eines Index** — Werkstück A15 von Phase 5.3
// (#1514).
//
// # Wogegen hier geprüft wird
//
// **„Crash/Disk-full/Abbruch vor und nach Umschaltung erzeugen nie gemischte
// Vektoren oder verlorene Quelldaten."** Ein abgebrochener Aufbau hinterlässt
// eine verworfene Generation und nicht eine halbe aktive.
//
// **„Eine aktive Schreibinstanz je Index."** Der zweite Aufbau wartet nicht,
// er wird abgewiesen — Warten hieße, dass zwei Aufrufer nacheinander zwei
// Indizes bauen, von denen einer sofort veraltet ist.

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// schreiberdoppel hält, was geschrieben wurde.
type schreiberdoppel struct {
	mu sync.Mutex

	naechste    int64
	geschrieben []domain.Retrievaldokument
	vektoren    [][]float32
	aktiviert   []int64
	verworfen   []int64

	anlegefehler     error
	schreibfehler    error
	aktivfehler      error
	uebernahmefehler error

	// uebernommen hält, was kopiert wurde, und wovon nicht.
	uebernommen int
	ausser      []string

	// vorSchreiben läuft vor jedem Schreiben — für Abbruch mitten im Lauf.
	vorSchreiben func()
}

func (s *schreiberdoppel) Anlegen(_ context.Context, manifest domain.Indexmanifest) (
	int64, error,
) {
	if fehler := manifest.Pruefe(); fehler != nil {
		return 0, fehler
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.anlegefehler != nil {
		return 0, s.anlegefehler
	}

	s.naechste++

	return s.naechste, nil
}

func (s *schreiberdoppel) Schreiben(
	_ context.Context, _ int64, dokument domain.Retrievaldokument, vektor []float32,
) error {
	if s.vorSchreiben != nil {
		s.vorSchreiben()
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.schreibfehler != nil {
		return s.schreibfehler
	}

	s.geschrieben = append(s.geschrieben, dokument)
	s.vektoren = append(s.vektoren, vektor)

	return nil
}

func (s *schreiberdoppel) Aktivieren(_ context.Context, generation int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.aktivfehler != nil {
		return s.aktivfehler
	}

	s.aktiviert = append(s.aktiviert, generation)

	return nil
}

func (s *schreiberdoppel) Uebernehmen(
	_ context.Context, _ int64, ausser []string,
) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.uebernahmefehler != nil {
		return 0, s.uebernahmefehler
	}

	s.ausser = ausser

	return s.uebernommen, nil
}

func (s *schreiberdoppel) Verwerfen(_ context.Context, generation int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.verworfen = append(s.verworfen, generation)

	return nil
}

// korpusdoppel gibt Stapel, bis nichts mehr da ist.
type korpusdoppel struct {
	staende []domain.Korpusstand
	stapel  [][]domain.Retrievaldokument

	staendefehler error
	fehler        error

	gegeben int
}

func (k *korpusdoppel) Staende(context.Context) ([]domain.Korpusstand, error) {
	return k.staende, k.staendefehler
}

func (k *korpusdoppel) Stapel(context.Context) ([]domain.Retrievaldokument, error) {
	if k.fehler != nil {
		return nil, k.fehler
	}

	if k.gegeben >= len(k.stapel) {
		return nil, nil
	}

	k.gegeben++

	return k.stapel[k.gegeben-1], nil
}

// einKorpus baut eine Quelle mit zwei Stapeln.
func einKorpus() *korpusdoppel {
	return &korpusdoppel{
		staende: []domain.Korpusstand{
			{Quelle: domain.QuelleOracle, Revision: "2026-09-01"},
		},
		stapel: [][]domain.Retrievaldokument{
			{einRetrievaltreffer("bolt", 1, domain.QuelleOracle).Dokument},
			{einRetrievaltreffer("shock", 1, domain.QuelleOracle).Dokument},
		},
	}
}

// einAufbau baut ihn über den Doppeln.
func einAufbau(
	schreiber *schreiberdoppel, einbetter service.Einbetter,
) *aufbaupaar {
	return &aufbaupaar{
		aufbau: service.NeuerIndexaufbau(schreiber,
			service.NeuerHaushalt(domain.Vorgabegrenzen())),
		einbetter: einbetter,
	}
}

// aufbaupaar hält Aufbau und Einbetter zusammen — der Aufbau selbst kennt
// keinen.
type aufbaupaar struct {
	aufbau    *service.Indexaufbau
	einbetter service.Einbetter
}

func (a *aufbaupaar) Aufbauen(
	ctx context.Context, wunsch service.Aufbauwunsch, quelle service.Korpusquelle,
) (service.Aufbaubericht, error) {
	return a.aufbau.Aufbauen(ctx, wunsch,
		func(context.Context) (service.Einbetter, service.Korpusquelle,
			func(), error,
		) {
			return a.einbetter, quelle, func() {}, nil
		})
}

func (a *aufbaupaar) Stand() service.Aufbaustand { return a.aufbau.Stand() }

// einAufbauwunsch nennt die Fassungen.
func einAufbauwunsch() service.Aufbauwunsch {
	return service.Aufbauwunsch{
		Schemafassung: 1, Zerteilfassung: "abschnitte-1", Indexfassung: 1,
	}
}

// **Ein Aufbau schreibt alles und schaltet dann um.**
func TestEinAufbauSchreibtAllesUndSchaltetDannUm(t *testing.T) {
	t.Parallel()

	schreiber := &schreiberdoppel{}
	einbetter := &retrievaleinbetterdoppel{
		raum: einRetrievalmanifest().Raum(), vektor: []float32{1, 0, 0},
	}

	bericht, fehler := einAufbau(schreiber, einbetter).Aufbauen(
		t.Context(), einAufbauwunsch(), einKorpus())
	if fehler != nil {
		t.Fatalf("Aufbauen: %v", fehler)
	}

	if bericht.Geschrieben != 2 || bericht.Generation != 1 {
		t.Fatalf("der Bericht ist %+v", bericht)
	}

	if len(schreiber.aktiviert) != 1 || schreiber.aktiviert[0] != 1 {
		t.Fatalf("aktiviert wurde %v", schreiber.aktiviert)
	}

	if len(schreiber.verworfen) != 0 {
		t.Fatalf("verworfen wurde %v", schreiber.verworfen)
	}

	// **Eingebettet wurde als Dokument und nicht als Frage.**
	for _, gesehen := range einbetter.gesehen {
		if gesehen == "" {
			t.Fatal("ein leerer Text wurde eingebettet")
		}
	}

	if len(einbetter.gesehen) != 2 {
		t.Fatalf("eingebettet wurden %v", einbetter.gesehen)
	}
}

// **Das Manifest trägt den Raum des Einbetters und die Stände des Korpus.**
func TestDasManifestTraegtRaumUndStaende(t *testing.T) {
	t.Parallel()

	schreiber := &schreiberdoppel{}

	bericht, fehler := einAufbau(schreiber, &retrievaleinbetterdoppel{
		raum: einRetrievalmanifest().Raum(), vektor: []float32{1, 0, 0},
	}).Aufbauen(t.Context(), einAufbauwunsch(), einKorpus())
	if fehler != nil {
		t.Fatalf("Aufbauen: %v", fehler)
	}

	if !bericht.Manifest.Raum().PasstZu(einRetrievalmanifest().Raum()) {
		t.Errorf("der Raum ist %+v", bericht.Manifest.Raum())
	}

	if bericht.Manifest.Suchkonfiguration != service.Rangfusionsname {
		t.Errorf("die Suchkonfiguration ist %q",
			bericht.Manifest.Suchkonfiguration)
	}

	if len(bericht.Manifest.Korpusstaende) != 1 {
		t.Errorf("die Stände sind %+v", bericht.Manifest.Korpusstaende)
	}
}

// **Ein zweiter Aufbau wird abgewiesen, solange einer läuft.**
func TestEinZweiterAufbauWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	laeuft := make(chan struct{})
	weiter := make(chan struct{})

	schreiber := &schreiberdoppel{}
	aufbau := einAufbau(schreiber, &retrievaleinbetterdoppel{
		raum: einRetrievalmanifest().Raum(), vektor: []float32{1, 0, 0},
	})

	var einmal sync.Once

	schreiber.vorSchreiben = func() {
		einmal.Do(func() { close(laeuft); <-weiter })
	}

	fertig := make(chan error, 1)

	go func() {
		_, fehler := aufbau.Aufbauen(context.WithoutCancel(t.Context()),
			einAufbauwunsch(), einKorpus())
		fertig <- fehler
	}()

	<-laeuft

	if _, fehler := aufbau.Aufbauen(t.Context(), einAufbauwunsch(),
		einKorpus()); !errors.Is(fehler, domain.ErrConflict) {
		t.Errorf("der zweite Aufbau kam durch: %v", fehler)
	}

	// **Und der Stand sagt, dass einer läuft.**
	if stand := aufbau.Stand(); !stand.Laeuft {
		t.Errorf("der Stand ist %+v", stand)
	}

	close(weiter)

	if fehler := <-fertig; fehler != nil {
		t.Fatalf("der erste Aufbau: %v", fehler)
	}

	if stand := aufbau.Stand(); stand.Laeuft || stand.Geschrieben != 2 {
		t.Errorf("nach dem Lauf: %+v", stand)
	}
}
