// The **collection** page — piece 2 and 3 of phase 3.4
// (#1165).
export const sammlung = {
  titel: 'Collection',
  erklaerung:
    'Ce que tu possèdes réellement — une ligne par impression, état, langue et finition. Cela n’a rien à voir avec tes sets : une carte conçue est un brouillon, un exemplaire est un morceau de carton.',

  aufnehmenOeffnen: 'Ajouter …',
  einfuhrOeffnen: 'Importer et exporter …',
  einfuhrTitel: 'Importer et exporter',
  leerTitel: 'Rien d’enregistré pour l’instant',
  leer: 'Un exemplaire a besoin de l’identifiant Scryfall de son impression — l’UUID figurant dans l’adresse d’une page de carte. Une liste entière s’importe d’un seul coup.',

  druck: 'Impression',
  menge: 'Quantité',
  zustand: 'État',
  sprache: 'Langue',
  ausfuehrung: 'Finition',
  foil: 'Foil',
  nonfoil: 'Non-foil',
  etched: 'Etched',
  erwerbsdatum: 'Acquis le',
  erwerbspreis: 'Payé',
  erwerbswaehrung: 'Devise',
  preisMehrdeutig:
    'Deux lectures possibles : {erste} ou {zweite}. Écris-le sans ambiguïté.',
  notiz: 'Note',

  // **Der lesbare Bezug und die Sicht darauf** (#1200).
  name: 'Nom de la carte',
  setcode: 'Set',
  nummer: 'Numéro',
  sicht: 'Vue de la collection',
  suche: 'Rechercher',
  suchePlatzhalter: 'nom, set, identifiant ou note',
  sortierung: 'Trier par',
  sortierBezeichnung: 'Carte',
  sortierMenge: 'Quantité',
  sortierErwerbsdatum: 'Acquis le',
  ausfuehrungAlle: 'Toutes les finitions',
  nurFoil: 'Foil uniquement',
  nurNonfoil: 'Non-foil uniquement',
  nurEtched: 'Etched uniquement',

  navigator: 'Vue et outils',
  inspector: 'Exemplaire',
  exemplarmenue: 'Exemplaire',
  inDecks: 'Dans ces decks',

  inspectorLeer:
    'Choisis un exemplaire dans la liste — la quantité, l’état, l’acquisition et le prix apparaissent alors ici, et c’est ici que tu le modifies.',
  inspectorAusserhalb:
    'Cet exemplaire ne figure plus dans la vue actuelle. Une modification concerne malgré tout cet exemplaire.',
  arbeitsplatz: 'Espace de travail de la collection',
  listeBezeichnung: 'Exemplaires',
  spalteWahl: 'Choix',
  spalteBezeichnung: 'Nom',
  spalteMenge: 'Quantité',
  spalteZustand: 'État',
  spalteSprache: 'Langue',
  spalteSet: 'Set',
  spalteLagerort: 'Emplacement',
  gruppeOhne: 'Non précisé',
  gruppeSumme: '{anzahl} exemplaires · {menge} pièces',
  kachelWaehlen: 'Sélectionner {bezeichnung}',

  ansicht: 'Affichage',
  ansichtTabelle: 'Tableau',
  ansichtGalerie: 'Galerie',
  ansichtGruppen: 'Groupes',
  gruppierung: 'Grouper par',
  gruppierungKeine: 'Ne pas grouper',
  gruppierungSet: 'Set',
  gruppierungLagerort: 'Emplacement',
  gruppierungZustand: 'État',
  gruppierungSprache: 'Langue',
  ohneTreffer: 'Aucun exemplaire ne correspond à la recherche et au filtre.',
  ohneTrefferZuruecksetzen: 'Réinitialiser la recherche et le filtre',
  gezeigt: '{gezeigt} sur {gesamt}',
  gekappt: '{gezeichnet} sur {passend} dessinés',
  mehrZeigen: 'Afficher plus',

  // **Die Listeneinfuhr** (#1292).
  listeTitel: 'Coller une liste',
  listeErklaerung:
    'Une carte par ligne, par exemple « 4 Lightning Bolt (M11) 149 ». Les lignes vides, « Deck », « Sideboard » et les lignes commençant par # ou // sont ignorées.',
  listeFeld: 'Liste de cartes',
  listePlatzhalter: '4 Lightning Bolt (M11) 149\n2x Counterspell',
  listeDeuten: 'Associer',
  listeEinlesen: 'Ajouter',
  listeOffen:
    '{offen} lignes sur {gesamt} demandent une décision — choisis l’impression que tu possèdes.',
  listeAlleEindeutig: 'Les {gesamt} lignes sont sans ambiguïté.',
  listeDruckWahl: 'Quelle impression ?',
  listeOhneWahl: '— pas encore décidé —',
  listeDrucke: '{drucke} impressions',
  listeBericht:
    '{angelegt} nouvelles, {erhoeht} augmentées, {stueck} exemplaires ajoutés.',
  listeNichtsGewaehlt:
    'Aucune ligne décidée — choisis au moins une impression.',
  listeAusgelassen: '{ausgelassen} lignes existantes ignorées.',
  listeKeinDruck: 'Aucune carte ne porte ce nom — vérifie l’orthographe.',
  listeZuViele:
    'Plus de {drucke} impressions portent ce nom — indique le set entre parenthèses, par exemple « (M11) », et il n’en reste qu’une.',
  listeOhneKarte: '{anzahl} lignes ne trouvent aucune carte et sont ignorées.',
  listeOhneKarteEins: 'Une ligne ne trouve aucune carte et est ignorée.',
  vorschauTitel: 'Avant l’import',
  vorschauErklaerung:
    'Certaines lignes correspondent à des lignes que ta collection contient déjà — même impression, même état, même langue, même finition et même intention. Choisis ce qu’il advient d’elles ; rien n’est écrit d’ici là.',
  vorschauZaehlung:
    '{neu} nouvelles · {zusammen} correspondent à des lignes existantes · {wiederholt} répètent une ligne de cet import',
  vorschauZusammen: '{bezeichnung}, {zustand} : {vorher} → {nachher}',
  vorschauModus: 'Lignes existantes',
  vorschauAddieren: 'Additionner les quantités',
  vorschauAuslassen: 'Ignorer les lignes existantes',
  vorschauEinlesen: 'Importer',
  vorschauAbbrechen: 'Ne pas importer',

  // **Die Ausfuhr** (#1292).
  ausfuhrTitel: 'Exporter et sauvegarder',
  ausfuhrErklaerung:
    '« CSV » contient toutes les informations et se relit intégralement. « Liste » est la forme collable pour un outil de construction de decks — elle ne contient que la quantité, la carte, l’impression et la finition.',
  ausfuhrAlsCSV: 'En CSV',
  ausfuhrAlsListe: 'En liste',
  ausfuhrErgebnis: 'Sortie',
  ausfuhrZurueck: 'Relire le CSV',
  ausfuhrLeer: 'Rien n’a encore été exporté.',

  // **Die Ordnung** (#1292, Vermerk 0057).
  lagerort: 'Emplacement de stockage',
  wunsch: 'Souhaité, non possédé',

  // **Die Ordnungssicht** (#1292).
  ordnungTitel: 'Organisation',
  ordnungLagerorte: 'Emplacements de stockage',
  ordnungDubletten: 'Doublons',
  ordnungWuensche: 'Liste de souhaits',
  ordnungBehalten: 'Garder',
  ordnungLeerLagerorte:
    'Aucun exemplaire pour l’instant — sans exemplaire, pas d’emplacement.',
  ordnungLeerDubletten: 'Aucune impression n’est là plus de {behalten} fois.',
  ordnungLeerWuensche:
    'Rien sur la liste de souhaits. Un exemplaire y arrive comme « Souhaité, non possédé » — ou depuis un deck via « Ajouter ce qui manque à la liste de souhaits ».',
  ordnungStueck: '{stueck} exemplaires sur {zeilen} lignes',
  ordnungOhneOrt: 'non précisé',
  ordnungUeberzaehlig: '{ueberzaehlig} en trop',

  // **Besitze ich diese Karte?** (#1292).
  besitzJa: 'Tu en possèdes {stueck}.',
  besitzNein: 'Tu ne possèdes pas cette impression.',
  besitzGewuenscht: '({gewuenscht} sur la liste de souhaits)',

  // **Die Massenvorgänge** (#1292).
  masseTitel: 'Actions groupées',
  masseWaehlen: 'Sélectionner',
  masseGewaehlt: '{gewaehlt} sur {gesamt} sélectionnés',
  masseLagerortSetzen: 'Définir l’emplacement de stockage sur',
  masseZustandSetzen: 'Définir l’état sur',
  masseAendern: 'Modifier la sélection',
  masseLoeschen: 'Supprimer la sélection',
  masseLoeschenFrage: 'Supprimer {anzahl} exemplaires ?',
  masseLoeschenErklaerung:
    'Ce qui disparaît ici, tu devras le saisir de nouveau : la quantité, l’état, l’emplacement de stockage et le prix d’acquisition n’existent nulle part ailleurs. La collection n’a pas d’historique comme un set.',
  masseLoeschenJa: 'Supprimer',
  masseBericht: '{geaendert} modifié(s).',
  masseZuruecknehmen: 'Annuler',
  masseZuruecknehmenVon: 'Annuler l’emplacement de stockage de {anzahl} lignes',
  masseZurueckgenommen: '{anzahl} annulé(s).',
  massePreise: 'Récupérer les prix de tous',
  massePreisschub: '{geholt} récupérés, {offen} restants',
  massePreiseFehlgeschlagen: '{anzahl} en échec',
  massePreiseSteht:
    'Ce lot n’a rien fait avancer — une nouvelle tentative donnerait le même résultat.',
  massePreiseFertig: 'Tous les prix sont là.',

  aufnehmen: 'Ajouter',
  aenderungSichern: 'Enregistrer',
  loeschen: 'Supprimer',
  loeschenFrage: 'Supprimer cet exemplaire ?',
  loeschenErklaerung:
    'La quantité, l’état, l’emplacement de stockage et le prix d’acquisition n’existent nulle part ailleurs — la collection n’a pas d’historique comme un set.',
  loeschenJa: 'Supprimer',
  abbrechen: 'Annuler',
  bearbeiten: 'Modifier',
  zumTausch: 'Vers l’échange',
  insDeck: 'Vers un deck',
  auswahlTitel: 'Sélection',
  auswahlInsDeck: '{anzahl} → deck',

  neuesTitel: 'Ajouter un exemplaire',
  weitereAngaben: 'Autres informations',

  dubletteFrage:
    '{bezeichnung} figure déjà dans la collection — {menge} exemplaire(s) dans le même état, la même langue et la même finition.',
  dubletteErklaerung:
    'Augmenter ajoute la nouvelle quantité ; les autres informations de la ligne existante sont conservées. Annuler ne change rien — ta saisie reste en place.',
  dubletteErhoehen: 'Augmenter à {summe}',
  dubletteZurZeile: 'Aller à la ligne existante',
  dubletteAbbrechen: 'Annuler',

  zustaende: {
    mint: 'Mint',
    nearMint: 'Near Mint',
    excellent: 'Excellent',
    good: 'Good',
    lightPlayed: 'Light Played',
    played: 'Played',
    poor: 'Poor',
  },

  preiseTitel: 'Prix',
  preiseHolen: 'Récupérer les prix',
  marktUndVerlauf: 'Offres et historique',
  // **The marketplace** — work item A4 of phase 5.3 (#1443).
  marktTitel: 'Offres de la place de marché',
  marktErklaerung:
    'Ce que des vendeurs individuels demandent pour cette impression — prix, quantité, état et langue. La consultation passe par le réseau. Rien n’est acheté.',
  marktHolen: 'Récupérer les offres',
  marktHolenBei: 'Récupérer les offres de {quelle}',
  marktLaeuft: 'Récupération en cours …',
  marktLeer: 'Aucune offre pour cette impression en ce moment.',
  marktZugeordnet:
    'Cette impression est maintenant associée à la version de carte {fassung} — la prochaine consultation la trouvera immédiatement.',
  marktStueck: '{menge} disponible(s)',
  marktBuendel: 'par lots de {buendel}',
  marktImUrlaub: 'n’expédie pas en ce moment',
  marktGradiert: 'gradée',
  marktSigniert: 'signée',
  marktVeraendert: 'altérée',
  marktBefunde: 'Toutes les offres n’ont pas pu être lues :',

  // **Die Begleitbefunde des Preisholens** (#1202).
  preisbefunde: 'Tous les prix n’ont pas pu être lus. Ces séries manquent :',
  preiseLaufen: 'Récupération en cours …',
  keinePreise: 'Aucun prix enregistré pour cette impression.',
  preisAlter: 'il y a {stunden} h · {quelle}',
  preisFrisch: 'à l’instant · {quelle}',
  preisVeraltet: 'périmé, {tage} jours · {quelle}',
  preisQuellstand: 'données de la source au {tag}',

  vorbehalt:
    'Les prix sont des estimations de Scryfall et de ses partenaires. Aucune garantie n’est donnée pour aucun d’entre eux ; c’est le prix en magasin qui fait foi.',

  verlauf: 'Historique',
  keinVerlauf:
    'Cette série ne compte qu’un seul relevé — un historique se constitue à mesure que tu récupères de nouveau les prix.',

  bulkTitel: 'Récupérer les prix d’une source entière',
  leiste: {
    bezeichnung: 'Chiffres de la collection',
    exemplare: 'Exemplaires',
    eintraege: '{zahl} entrées',
    wert: 'Valeur',
    keinPreis: 'aucun prix',
    ohnePreis: 'Sans prix',
    nichtGelesen: 'non lu',
  },
  drucksuche: {
    karte: 'Carte',
    hinweis:
      'Nom, éventuellement avec set et numéro — p. ex. Lightning Bolt (M11) 149',
    suchen: 'Chercher',
    drucke: 'Impressions',
    keinTreffer: 'Aucune carte de ce nom trouvée.',
    weitere:
      'Il existe d’autres impressions — set et numéro réduisent le choix.',
    treffer: '{name} · {set} ({code}) · N° {nummer}',
  },
  bulkErklaerung:
    'MTGJSON ne répond pas carte par carte, mais avec un fichier. « Aujourd’hui » apporte la journée en cours, « Quatre-vingt-dix jours » l’historique — et plus d’un gigaoctet.',
  bulkQuelle: 'Source',
  bulkUmfang: 'Étendue',
  bulkHeute: 'Aujourd’hui',
  bulkNeunzigTage: 'Quatre-vingt-dix jours',
  bulkHolen: 'Récupérer',
  bulkHolenBei: 'Récupérer les prix de {quelle}',
  bulkLaeuft: 'Récupération en cours …',
  bulkBilanz: '{staende} prix enregistrés pour {drucke} impressions.',
  bulkOhneDrucke:
    'Aucune impression de cette collection ne porte d’identifiant MTGJSON — rien n’a été demandé.',
  uebersichtTitel: 'Vue d’ensemble',
  uebersichtLeer: 'Aucun prix récupéré pour l’instant.',
  ohnePreis: 'sans prix : {zahl}',
  // **Ein Exemplar und nicht „1 Exemplare“.** Eine Zahl mit falscher
  // Form liest sich wie ein Fehler im Programm — gesehen im laufenden
  // Fenster.
  einExemplar: '{zahl} exemplaire',
  exemplare: '{zahl} exemplaires',

  keineUmrechnung:
    'Une somme par devise. Aucune conversion — cela demanderait un taux de change, que ce programme n’a pas.',
  // **A price history from another source** — A5, #1444.
  verlaufTitel: 'Historique des prix d’une autre source',
  verlaufErklaerung:
    'Récupère une série dans le temps et la stocke. La consultation passe par le réseau et consomme du quota ; l’état est affiché tel que la source le nomme.',
  verlaufHolen: 'Récupérer l’historique',
  verlaufHolenBei: 'Récupérer l’historique de {quelle}',
  verlaufLaeuft: 'Récupération en cours …',
  verlaufBilanz:
    '{staende} points de prix enregistrés — version de carte {fassung}.',
} as const
