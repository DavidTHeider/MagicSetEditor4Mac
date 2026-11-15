// Frame, navigation, language selection.
export const allgemein = {
  // **Count words** (#1218) — a number and its word in the matching form.
  zaehlwoerter: {
    karten: { eine: 'carte', viele: 'cartes' },

    // **Und die Seiten** (#1219):
    // der Druckbogen meldete »1 Karten auf 1 Seiten«.
    seiten: { eine: 'page', viele: 'pages' },

    // And the three objects of the overview (#1422).
    sets: { eine: 'set', viele: 'sets' },
    decks: { eine: 'deck', viele: 'decks' },
    pakete: { eine: 'paquet', viele: 'paquets' },
    eigenePakete: { eine: 'paquet personnel', viele: 'paquets personnels' },
    stile: { eine: 'style', viele: 'styles' },
    positionen: { eine: 'position', viele: 'positions' },
    exemplare: { eine: 'exemplaire', viele: 'exemplaires' },

    // **Und die Drucke einer Sichtfreigabe** (UXR-F18, #2033).
    drucke: { eine: 'impression', viele: 'impressions' },
  },
  bestandGesichert: 'La base de données a été sauvegardée dans',
  zahlenbereich: 'Un nombre entre {kleinstens} et {groesstens}.',
  zahlenAb: 'Un nombre à partir de {kleinstens}.',
  ganzeZahlBereich: 'Un nombre entier entre {kleinstens} et {groesstens}.',
  ganzeZahlAb: 'Un nombre entier à partir de {kleinstens}.',
  nurZiffern:
    'Chiffres uniquement, sans point ni virgule — le montant est exprimé dans la plus petite unité.',

  // **Ein Doppelklick im Finder, der nicht aufging** (#891).
  dateiNichtGeoeffnet: 'Le fichier {pfad} n’a pas pu être ouvert. {grund}',
  anwendungsname: 'MagicSetEditor4Mac',
  sprache: {
    bezeichnung: 'Langue',
    de: 'Deutsch',
    en: 'English',
    es: 'Español',
    fr: 'Français',
    it: 'Italiano',
    ja: '日本語',
    ru: 'Русский',
    zh: '简体中文',
  },
  // **Die fünf Arbeitsbereiche, der Einstieg und die Dienstzone** — UX-02
  // (#1674).
  //
  // **`bereiche` und `dienste` sind die zwei Überschriften der Leiste**, nicht
  // Bereiche selbst — sie sagen, was darunter steht.
  bereiche: {
    // **„Home" und nicht „Start"** — so heißt die Seite in dieser Sprache
    // (`en/seitenstart.ts`), und der Bereich trägt ihren Namen.
    start: 'Accueil',
    gestalten: 'Concevoir',
    sammeln: 'Collectionner',
    spielen: 'Jouer',
    handeln: 'Échanger',
    ressourcen: 'Ressources',
    programm: 'Application',

    ueberschrift: 'Espaces de travail',
  },

  navigation: {
    bezeichnung: 'Navigation principale',
    sprungZumInhalt: 'Aller au contenu',
    seitenleisteEinblenden: 'Afficher la barre latérale',
    seitenleisteAusblenden: 'Masquer la barre latérale',
    seitenleisteKuerzel: '⌥⌘S',
    weg: 'Tu es ici',
  },

  // When something does not draw — F01 (#2340).
  fehlergrenze: {
    bereichTitel: 'Cette zone n’a pas pu être affichée',
    bereichSatz:
      'Référence {kennung}. Le reste de la base de données est intact ; l’erreur figure dans le journal sous cette référence.',
    neuLaden: 'Recharger la zone',
    diagnoseOeffnen: 'Ouvrir le diagnostic',
    wurzelTitel: 'La fenêtre n’a pas pu être affichée',
    wurzelSatz:
      'Référence {kennung}. Ta base de données est intacte ; l’erreur figure dans le journal sous cette référence.',
    fensterNeuLaden: 'Recharger la fenêtre',
  },
  meldungen: {
    methodeFehlt: 'La méthode {name} n’est pas reliée.',
    vorschauLeer: 'L’aperçu est revenu sans image et sans erreur.',
    kontoLeer: 'Le compte est revenu sans utilisateur et sans erreur.',
  },
  // Das Band über dem Inhalt: was zu keiner Seite gehört (#180).
  meldungsband: {
    bereich: 'Messages de la fenêtre',
    setNichtGesichert: 'Le set « {name} » n’a pas été enregistré — {grund}',
    setNichtGesichertOhneName: 'Le set n’a pas été enregistré — {grund}',
    bestandNichtGesichert:
      'La base de données n’a pas été sauvegardée — {grund}',
    verlaufGescheitert: 'Impossible d’annuler ou de rétablir — {grund}',
    darstellungGescheitert: 'Impossible de réinitialiser l’apparence — {grund}',
  },
  // Die neue Fassung, wenn eine da ist (#1100).
  vertraeglichkeit: {
    vertraeglich: 'Cette version peut ouvrir ta bibliothèque.',
    unvertraeglich: 'Cette version ne peut pas ouvrir ta bibliothèque.',
    ungeklaert:
      'Cette version n’indique pas si elle peut ouvrir ta bibliothèque — elle est antérieure à cette vérification. Sauvegarde ta bibliothèque avant de changer de version, et restaure-la en cas de problème.',
    bereit: 'Le changement est prêt : sauvegarde effectuée et relue.',
    vorbereiten: 'Préparer le changement',
    hindernisse: {
      keinManifest: 'La version ne donne aucune information sur elle-même.',
      plattform: 'Elle a été conçue pour une autre machine.',
      wanderungFehlt:
        'Ta bibliothèque contient une modification qu’elle ne connaît pas.',
      summeAnders:
        'Une modification a un contenu différent dans cette version.',
      vorDerGrundlinie:
        'Ta bibliothèque est antérieure à sa ligne de base — ouvre-la d’abord avec une version intermédiaire.',
      platz: 'Il n’y a pas assez d’espace libre.',
      vorgaengeLaufen: 'Une opération est encore en cours.',
      keineSicherung: 'Aucune sauvegarde n’a été faite.',
      sicherungUnlesbar: 'La sauvegarde n’a pas pu être relue.',
      halter:
        'Un autre programme garde la bibliothèque ouverte — ferme-le avant le changement.',
    },
  },
  fassungsband: {
    bereich: 'Nouvelle version',
    neu: 'La version {fassung} est disponible.',
    seite: 'Ce qui a changé',
    laden: 'Télécharger',
    laeuft: 'Téléchargement…',
    fortschritt: '{gelesen} sur {ganz} téléchargés',
    fortschrittOhneLaenge: '{gelesen} téléchargés',
    geladen:
      'La version {fassung} se trouve dans {pfad}. Pour l’installer, ouvre-la dans le Finder et fais glisser l’application dans le dossier Applications.',
    weglegen: 'Fermer ce message',
  },
  wiederanlauf: {
    bereich: 'Dernière session',
    satz: 'La dernière session, commencée le {zeit}, s’est terminée de façon inattendue.',
    satzOhneZeit: 'La dernière session s’est terminée de façon inattendue.',
    mitSpur:
      'Une trace de plantage est disponible ; le lot de diagnostic l’inclut quand tu le crées.',
    ohneSpur:
      'Aucune trace de plantage : l’application a sans doute été fermée de l’extérieur.',
    ausblenden: 'Masquer',
  },

  // Der Arbeitsplatz eines Sets: seine Laschen und sein leerer Zustand (#288).
  arbeitsplatz: {
    leiste: {
      bezeichnung: 'Chiffres du set',
      befund: 'Avec constats',
      nichtGeprueft: 'non vérifié',
      nichtGelesen: 'non lu',
      geprueft: 'vérifié {wann}',
    },
    laschen: 'Sections du set',

    gruppen: {
      bauen: 'Créer des cartes',
    },
    schubladen: {
      eigenschaften: {
        knopf: 'Propriétés',
        titel: 'Propriétés du set',
        laschen: 'Sections des propriétés',
        hinweis: 'Ce qui s’applique à toutes les cartes de ce set.',
      },
      werkzeuge: {
        knopf: 'Outils',
        titel: 'Outils',
        laschen: 'Outils du set',
        hinweis: 'Statistiques, pack aléatoire et console pour ce set.',
      },
      schliessen: 'Fermer',
    },

    keinSet: 'Aucun set ouvert',
    keinSetErklaerung:
      'Choisis un set sous « Sets », crées-en un ou importes-en un — ses cartes, ses statistiques et sa feuille d’impression apparaîtront ici.',
    zuDenSets: 'Aller aux sets',
    zuDenPaketen: 'Aller aux paquets',
    zuDenKarten: 'Aller aux cartes',
    ohneKartenTitel: 'Aucune carte dans ce set pour l’instant',
    ohneKarten:
      'Cet onglet travaille avec les cartes du set. Les cartes naissent sous « Cartes » — créées, collées ou importées.',
    ohneNamen: 'Set ouvert',
  },
  palette: {
    bezeichnung: 'Palette de commandes',
    eingabe: 'Où aller ?',
    keineTreffer: 'Aucun résultat. Essayer un autre mot ?',
    oeffnen: 'Commandes',
    seiten: 'Pages',
    befehle: 'Commandes',
    befehlEingabe: 'Où aller — ou quelle commande ?',
    tun: 'Faire',
    finden: 'Trouver',
    karteErstellen: 'Créer une carte',
    setWaehlen: 'Créer ou ouvrir un set',
    deckSpielen: 'Partie rapide avec {name}',
    tauschenMit: 'Échanger des cartes avec {name}',
    sicherungErstellen: 'Créer une sauvegarde',
    schnittstellenEinrichten:
      'Configurer les interfaces (ligne de commande, MCP, REST)',
    woKarte: 'Concevoir · {set}',
    woSets: 'Concevoir · Sets',
    woPartie: 'Jouer · Commencer une partie',
    woHandel: 'Échanger · Correspondant',
    woEinstellungen: 'Réglages · {gruppe}',
    alsKarte: 'Carte · {set}',
    alsSet: 'Set · {karten}',
    alsDeck: 'Deck · {spiel}',
    alsGegenueber: 'Correspondant · confirmé',
    fuss: '↑↓ choisir · ↵ exécuter · échap fermer',
  },
} as const
