// **Die Startseite** — herausgelöst, weil `seiten.ts` mit ihr über 400 Zeilen
// ginge (§2).
//
// **Ein Katalog darf sich auf mehrere Dateien verteilen**, solange er als
// **einer** ankommt — so wie `seitenkarten.ts`. Was gleich bleibt, ist die
// Zahl der Kataloge; an ihr hängt der Vertragstest gegen die Go-Seite (§9).
export const start = {
  ueberblick: {
    titel: 'Ton aperçu',
    mitProblemen: '{anzahl} avec des problèmes',
    geprueftVon: '{geprueft} sur {gesamt} vérifiés',
    mitBefund: '{anzahl} avec des constats',
    nichtGeprueft: 'pas entièrement vérifié',
    problemeGeprueft:
      'Problèmes vérifiés {wann} · {geprueft} sur {gesamt} cartes',
    problemeOffen: 'Problèmes : {offen} sets sur {gesamt} non vérifiés',
    pruefen: 'Vérifier maintenant',
    setsMitProblemen: '{sets} avec des problèmes',
    kartenMitBefund: '{karten} avec des constats dans {sets}',
    aeltesterPreis: 'prix le plus ancien {wann}',
    ohnePreis: '{anzahl} sans prix',
    mitFehlenden: '{anzahl} avec des cartes manquantes',
    decksMitFehlenden: '{decks} avec des cartes manquantes',
    alter: {
      gerade: 'à l’instant',
      minuten: 'il y a {minuten} min',
      stunden: 'il y a {stunden} h',
      tage: 'il y a {tage} jours',
      tag: 'il y a 1 jour',
    },
  },
  // Resume — work item U11 of phase 5.1 (#1422).
  // Attention and background tasks (#1422).
  aufmerksamkeit: {
    titel: 'À surveiller',
    nichts: 'Rien — {sets}, {decks} et {pakete} vérifiés.',
    durchgesehen: '{sets}, {decks} et {pakete} vérifiés.',
    arten: {
      setVorbehalte: 'Non interprétés : {anzahl}',
      deckFehlend: 'Manquants dans la collection : {anzahl}',
      paketBefund: 'Dans les scripts : {anzahl}',
    },
    paketeMitBefund: '{pakete} avec des constats dans le script',
    gruendeZeigen: 'Afficher les raisons',
    gruendeVerbergen: 'Masquer les raisons',
    setBefunde: '{karten} avec des constats',
    befunde: 'Non vérifiés :',
    alle: 'Afficher les {anzahl}',
    weniger: 'Afficher moins',
    gruende: {
      vorbehalt: '{0}, ligne {1}',
    },
    amOrt: '{ort} : {satz}',
    // The operations band (OPS02, #2348).
    betrieb: {
      arten: {
        platzKnapp: 'Plus que {frei} libres à côté de la bibliothèque',
        sicherungAlt: 'Dernière sauvegarde il y a {tage} jours',
        dienstBeendet:
          'Le service de connexion s’est arrêté avec le code {ausgang}',
        webprozessNeu: 'La vue web a été rechargée {anzahl} fois',
      },
      rat: {
        platzKnapp:
          'Sous environ 2 Go ou 5 % libres, chaque sauvegarde devient juste — fais de la place dans « Données et sauvegarde ».',
        sicherungAlt:
          'Plus de 14 jours — sauvegarde la bibliothèque dans « Données et sauvegarde ».',
        dienstBeendet:
          'Après une erreur, launchd le relance — le diagnostic dit pourquoi il s’est arrêté.',
        webprozessNeu:
          'WebKit avait arrêté son processus ; ce qui n’était pas enregistré dans un formulaire ouvert peut manquer — vérifie-le dans le diagnostic.',
      },
    },
  },
  auftraege: {
    titel: 'Tâches en arrière-plan',
    staende: {
      laeuft: 'en cours',
      fertig: 'terminé',
      gescheitert: 'échoué',
      abgebrochen: 'annulé',
    },
    arten: {
      preise: 'Récupérer les prix de la collection',
      sammeldaten: 'Charger les données en masse de Scryfall',
      fassung: 'Récupérer une nouvelle version',
      bogen: 'Composer une feuille d’impression',
      set: 'Importer un set',
      mtgjson: 'Importer du MTG JSON',
      vorlage: 'Exporter via un modèle',
      installer: 'Lire un installateur',
      weitergabe: 'Transmettre comme installateur',
      probleme: 'Vérifier les sets',
    },
    abbrechen: 'Annuler « {art} »',
    dauer: '{sekunden} s',
    laufend: { eine: 'tâche en cours', viele: 'tâches en cours' },
  },
  weiterarbeiten: {
    titel: 'Reprendre',
    weiterAn: 'Continuer avec {art} « {name} »',
    zuletztWann: 'Dernière action : {vorgang} le {zeit}',
    angeheftet: 'épinglé',
    angeheftetLeer: 'Rien d’épinglé — épingle une ligne pour la garder ici.',
    leer: 'C’est ici qu’apparaît ton travail récent. Pour l’instant, c’est vide — la première étape est un set ou un paquet de modèles.',
    handlungSets: 'Créer ou ouvrir un set',
    handlungPakete: 'Récupérer un paquet',
    anheften: 'Épingler',
    loesen: 'Désépingler',
    anheftenVon: 'Épingler « {name} »',
    loesenVon: 'Désépingler « {name} »',
    arten: { set: 'Set', deck: 'Deck', paket: 'Paquet' },
    vorgaenge: {
      geoeffnet: 'ouvert',
      eingefuehrt: 'importé',
      ausgegeben: 'exporté',
    },
  },
  titel: 'Accueil',
  ueberschrift: 'MagicSetEditor4Mac',
  was: 'Conçois, examine et imprime des cartes pour Magic et d’autres jeux — avec les modèles qu’utilise Magic Set Editor 2.',
  ersteSchritte: 'Premiers pas',
  einfuehrung: {
    willkommen:
      'Trois étapes et la première carte est prête. Tu peux rouvrir cette fenêtre à tout moment depuis les réglages.',
    wiederzeigen: 'Afficher de nouveau au démarrage',
    schliessen: 'C’est parti',
    bezeichnung: 'Introduction',
  },
  schritte: {
    setOeffnen: {
      titel: 'Ouvrir un set',
      erklaerung:
        'Ouvre un fichier .mse-set existant — ou importe un set depuis MTG JSON.',
      handlung: 'Aller aux sets',
    },
    kartenBearbeiten: {
      titel: 'Modifier des cartes',
      erklaerung:
        'Consulte et modifie les cartes du set ouvert ; l’aperçu montre chaque modification aussitôt.',
      handlung: 'Aller aux cartes',
      ersteinSet: 'Un set doit d’abord être ouvert.',
      stattdessen: 'Ouvrir d’abord un set',
    },
    paketeHolen: {
      titel: 'Récupérer des modèles',
      erklaerung:
        'Huit paquets de modèles sont déjà là. Tu veux d’autres jeux ou styles ? Récupère-les ici.',
      handlung: 'Aller aux paquets',
    },
  },
}
