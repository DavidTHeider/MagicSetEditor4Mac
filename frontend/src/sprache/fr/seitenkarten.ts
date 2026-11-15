// Die Texte der **Kartenseite** — die größte Seite der Anwendung.
//
// Aus `seiten.ts` herausgelöst, als die Datei über 400 Zeilen wuchs (§2).
// Der Schnitt folgt der Frage: dort stehen die Texte **aller** Seiten, hier
// die der einen, die am meisten trägt — und weiter wachsen wird
// (#674).

import { verlaufEn } from './seitenverlauf'
import { massenvorschauEn } from './seitenmassen'
import { autoersetzung } from './seitenautoersetzung'
import { schluesselwoerter } from './seitenschluesselwoerter'
import { werkstatt } from './seitenwerkstatt'
import { konsole } from './seitenkonsole'
import { symbolbibliothek } from './seitensymbolbibliothek'

export const karten = {
  ...werkstatt,
  autoersetzung,
  schluesselwoerter,
  symbolbibliothek,
  anlegen: 'Ajouter une carte',
  nachOben: 'Monter',
  nachUnten: 'Descendre',
  duplizieren: 'Dupliquer la carte',
  loeschen: 'Supprimer la carte',
  loeschenFrage: 'Supprimer cette carte ? ⌘Z permet de l’annuler.',
  loeschenJa: 'Supprimer',
  loeschenNein: 'Annuler',
  loeschenTrotzdem: 'Supprimer quand même',
  symbole: 'Insérer des symboles',
  hergeleitet: 'Déduit',
  hergeleitetWert: 'Déduit : {wert}',
  bildWaehlen: 'Choisir une image',
  stilfelder: 'Champs du style',
  stilfelderHilfe:
    'Ces champs viennent avec le style de la carte. Leur valeur ne vaut que pour ce style.',
  symbolZeichnen: 'Dessiner un symbole',
  kartenskripteTitel: 'Cartes à partir d’un script',
  kartenskripteErklaerung:
    'Des actions prêtes à l’emploi issues du paquet de jeu — un clic ajoute plusieurs cartes à la fois. ' +
    'Une seule annulation les retire toutes.',
  kartenskriptEine: 'Une carte ajoutée.',
  kartenskriptMehrere: '{anzahl} cartes ajoutées.',
  bildEinsetzen: 'Insérer l’image',
  pack: {
    ueberschrift: 'Pack aléatoire',
    erklaerung:
      'Un booster composé des cartes de ce set, selon les règles du paquet de jeu. La même graine donne le même tirage — c’est ainsi qu’un résultat peut être reproduit.',
    laedtGrund:
      'Chaque type de pack demande à chaque carte si elle y convient — cela prend un instant la première fois, puis c’est prêt.',
    art: 'Type de pack',
    same: 'Graine',
    sameZahl: 'La graine est un nombre entier de 0 à 18446744073709551615.',
    ziehen: 'Tirer',
    neuerSame: 'Nouvelle graine',
    gezogen: 'Tiré',
    keineArten:
      'Le jeu de ce set n’apporte aucun type de pack — ou son paquet n’est pas installé.',
    keineKarten: 'Ce type de pack n’a produit aucune carte.',
    artMitZahl: '{name} ({zahl} cartes)',
    nurBaustein: 'Composant — tiré par d’autres types de pack',
    vorbehalte: 'Types de pack dont le filtre n’a pas pu être évalué',
  },
  symbol: {
    ueberschrift: 'Symbole',
    erklaerung:
      'Le symbole de ce set — dessiné, pas chargé. Il apparaît aussitôt dans l’aperçu et est intégré au fichier lors de l’enregistrement du set. Chaque forme peut être placée et déplacée au clavier.',

    grundform: 'Ajouter une forme de base',

    verlauf: 'Historique du dessin',
    nachzeichnen: 'Vectoriser une image',
    groesser: 'Plus grand',
    nachzeichnenHilfe:
      'Un PNG, JPEG ou GIF à la silhouette nette devient des formes à points de Bézier que tu peux continuer à modifier ici.',
    zuruecknehmen: 'Annuler le geste',
    wiederherstellen: 'Rétablir le geste',
    kreis: 'Cercle',
    rechteck: 'Rectangle',
    dreieck: 'Triangle',
    stern: 'Étoile',

    teile: 'Éléments',
    keineTeile:
      'Rien n’est encore dessiné. Une forme de base est le point de départ — chaque point peut ensuite être modifié.',
    teilLoeschen: 'Supprimer l’élément',
    teilMitPunkten: '{name} — {zahl} points',
    teilMitKopien: '{name} — symétrie, {zahl} copies',
    gruppieren: 'Grouper',
    gruppierenHilfe:
      'Deux éléments sélectionnés ou plus forment un groupe. Maintiens ⌘ pour en sélectionner plusieurs.',
    aufloesen: 'Dissocier',
    nachOben: 'Monter',
    nachUnten: 'Descendre',
    nachObenFuer: 'Monter {name}',
    nachUntenFuer: 'Descendre {name}',
    teilLoeschenFuer: 'Supprimer {name}',
    teilMitKindern: '{name} — {zahl} éléments',
    teilKind: '{name}, dans un groupe',
    wahlHilfe:
      'Maintiens ⌘ pour sélectionner plusieurs éléments. L’ordre est l’ordre de dessin : ce qui figure en premier se trouve au-dessus.',

    zeichnung: 'Zone de dessin',
    punkt: 'Point {nummer} de {name}',
    griffDavor: 'Poignée avant le point {nummer}',
    griffDanach: 'Poignée après le point {nummer}',
    punktHinzufuegen: 'Insérer un point',
    punktLoeschen: 'Supprimer le point',

    tastatur: 'Au clavier',
    tastaturhilfe:
      'Tab passe d’un point à l’autre. Les flèches déplacent d’un centième, avec Maj d’un dixième. Entrée insère un nouveau point après le point sélectionné, Retour arrière le supprime.',

    kombination: 'Combiner',
    ueberdecken: 'Superposer',
    vereinen: 'Fusionner',
    abziehen: 'Soustraire',
    schneiden: 'Intersecter',
    unterschied: 'Différence',
    umranden: 'Contour',

    symmetrie: 'Symétrie',
    symmetrieAnlegen: 'Ajouter une symétrie',
    symmetrieAufloesen: 'Dissoudre la symétrie',
    symmetrieErklaerung:
      'Une symétrie multiplie ce qu’elle contient. Les copies apparaissent en filigrane sur le canevas ; seul l’original est modifié.',
    symmetrieHinweis: 'Sélectionne un élément pour en faire une symétrie.',
    symmetrieart: 'Type',
    drehung: 'Rotation',
    achse: 'Réflexion',
    kopien: 'Copies',
    kopienHinweis:
      'Une réflexion nécessite des paires — un nombre impair laisse une copie seule, et MSE arrondit à l’inférieur.',
    mitteX: 'Centre, horizontal',
    mitteY: 'Centre, vertical',
    griffX: 'Poignée, horizontale',
    griffY: 'Poignée, verticale',
    mitte: 'Centre de la symétrie',
    griff: 'Poignée de la symétrie',

    sichern: 'Placer dans le set',
    gesichert:
      'Enregistré sous {name}. Il est intégré au fichier lors de l’enregistrement du set.',
  },
  konsole,

  wortlisten: {
    ueberschrift: 'Choix',
    liste: 'Choisir dans {liste}',
    keine: 'aucun de ces choix',
    gruppe: '{name} (groupe)',
  },
  regeln: {
    ueberschrift: 'Rulings de Scryfall',
    holen: 'Récupérer les rulings',
    keine: 'Cette carte n’a aucun ruling.',
    offiziell: 'officiel',
    anmerkung: 'Remarque de Scryfall',
    anmerkungErklaerung:
      'Les remarques viennent de l’équipe Scryfall et expliquent souvent un format non officiel. Ce ne sont pas des règles.',
    quelle: 'Voir cette carte sur Scryfall',
    alter: 'État : datant de {stunden} heures',
    manaPruefen: 'Vérifier le coût de mana',
    manaGut: 'Le coût de mana est correct.',
    manaSchlecht: 'Inconnu : {symbole}',
    manaLeer: 'Cette carte n’a pas de coût de mana.',
  },
  abgleich: {
    ueberschrift: 'Données de carte de Scryfall',
    erklaerung: 'Montre ce qui changerait. Seul ce que tu coches est écrit.',
    holen: 'Récupérer une proposition',
    uebernehmen: 'Appliquer les éléments cochés',
    nichtsZuTun: 'Rien à modifier — la carte correspond à Scryfall.',
    alt: 'actuel',
    neu: 'nouveau',
    leer: '(vide)',
    alter: 'État : datant de {stunden} heures',
    frisch: 'État : tout juste récupéré',
    quelle: 'Voir cette carte sur Scryfall',
  },
  scryfall: {
    ueberschrift: 'Image de Scryfall',
    erklaerung:
      'Le nom de la carte détermine l’image. L’artiste est également inscrit dans la carte — Scryfall exige que son nom figure à côté de l’illustration.',
    name: 'Nom de la carte sur Scryfall',
    art: 'Type d’image',
    artKlein: 'Petite (aperçu)',
    artNormal: 'Carte entière',
    artAusschnitt: 'Illustration seule',
    holen: 'Récupérer l’image',
    ausschnittHinweis:
      'Le recadrage montre l’illustration seule. Il est refusé lorsque Scryfall ne nomme pas l’artiste.',
    quelle: 'Voir cette carte sur Scryfall',
  },
  scryfallkarte: {
    name: 'Nom de la carte sur Scryfall',
    platzhalter: 'p. ex. Lightning Bolt',
    mitBild: 'avec illustration',
    anlegen: 'Créer depuis Scryfall',
  },
  zuschnitt: {
    ueberschrift: 'Choisir le recadrage',
    erklaerung:
      'Le cadre indique quelle partie de l’image se retrouve sur la carte. Faites-le glisser pour le déplacer, fais glisser un coin pour le redimensionner.',
    quelle: 'Image source avec le recadrage choisi',
    vorschau: 'Aperçu à la taille du champ',
    links: 'Gauche',
    oben: 'Haut',
    breite: 'Largeur',
    hoehe: 'Hauteur',
    seitenverhaeltnis: 'Conserver les proportions du champ',
    uebernehmen: 'Insérer',
    abbrechen: 'Annuler',
    masse: '{breite} × {hoehe} pixels',
  },
  feldWaehlen: 'Modifier ce champ',
  leerTitel: 'Aucune carte pour l’instant',
  listeBezeichnung: 'Cartes du set',
  mehr: 'Plus',
  gruppeAnlage: 'Carte',
  gruppeAnsicht: 'Affichage',
  ausScryfallAufklapp: 'Créer depuis Scryfall …',
  ausScryfallErklaerung:
    'Au lieu d’une carte vierge : saisis un nom ou cherche sur Scryfall — les champs viendront alors de là.',
  dieseKarte: 'Cette carte',
  kartenmenue: 'Carte',
  auswahlLoesen: 'Afficher toutes les cartes',
  bezugAufklapp: 'Lier à une impression …',
  dieseKarteLeer: 'Ouvre une carte pour la modifier.',
  keineKarteTitel: 'Aucune carte sélectionnée',
  titel: 'Cartes',
  leer: 'Crées-en une ou colle des cartes copiées.',
  ohneName: 'Sans nom',
  ohneFelder: 'Cette carte n’a aucune valeur de champ.',
  nummer: 'Numéro',
  vorschau: 'Aperçu de la carte',
  keineKarte:
    'Choisis-en une dans la liste — ses champs et son aperçu s’afficheront ici.',
  bearbeiten: 'Modifier',

  notizen: {
    ueberschrift: 'Notes sur la carte',
    platzhalter: 'Pour toi seul — n’apparaît jamais sur la carte.',
  },

  // **What the game package does not declare** (#1271).
  buchfuehrung: 'Ce que l’application en sait',
  tiefe: 'Niveau de détail',
  tiefeNormal: 'Normal',
  tiefeErweitert: 'Avancé',
  tiefeExperte: 'Expert',
  tiefeErweitertTitel: 'Ce qu’un script remplit',
  tiefeErweitertHinweis:
    'Le paquet de jeu calcule ces champs — ils indiquent pourquoi quelque chose figure sur la carte. Ils ne peuvent pas être modifiés.',
  befundVerdeckt: 'Un constat se trouve sous « {tiefe} ».',
  befundSpringen: 'Y aller',
  buchfuehrungHinweis:
    'Ces informations n’appartiennent pas à la carte mais à sa gestion — quel modèle elle utilise et ce qui y est réglé. Ne les modifie que si tu sais pourquoi.',

  // **Note and marks together, collapsed** (#1271). They sat below the list
  // and took its height: four rows were left.
  arbeitsstand: 'Avancement sur cette carte',
  arbeitsstandHinweis:
    'Note et étiquettes — elles relèvent de ton travail sur la carte, pas de la carte. Un fichier que tu transmets ne les contient pas.',

  marken: {
    ueberschrift: 'Étiquettes',
    erklaerung:
      'Des mots pour ton propre travail — ils se trouvent à côté de la carte, pas dedans. Un fichier que tu transmets ne les contient pas.',
    platzhalter: 'p. ex. terminé',
    setzen: 'Ajouter une étiquette',
    entfernen: 'Retirer l’étiquette {marke}',
    keine: 'Aucune étiquette sur cette carte pour l’instant.',
    schonGesetzt:
      'La carte porte déjà « {marke} » — posée ailleurs entre-temps.',
    filter: 'Filtrer par étiquette',
    alle: 'Toutes',
  },

  scryfallsuche: {
    frage: 'Rechercher sur Scryfall',
    platzhalter: 'p. ex. Lightning Bolt — ou t:creature c:r cmc<=3',
    erklaerung:
      'Un nom de carte suffit. Si tu connais le langage de recherche de Scryfall, tu peux l’utiliser ici.',
    suchen: 'Rechercher',
    nichtsGefunden: 'Rien trouvé. Essayer un autre mot ?',
    mehr: 'Scryfall a plus de résultats. Affine la requête — rien n’est récupéré automatiquement.',
    alter: 'État : datant de {stunden} heures',
    frisch: 'État : tout juste récupéré',
    befund: 'Constat : {text}',
  },

  // **Der Verweis auf einen echten Druck** (#1167).
  verweis: {
    aufschrift: 'Lié à',
    keiner:
      'Cette carte n’est liée à aucune impression. La synchronisation avec Scryfall en choisira alors une au hasard — cherche ci-dessus et lie-la.',
    druck: 'de {setcode} {nummer}',
    gesetzt: '— lié le {wann}',
    setzen: 'Lier à {name} ({setcode} {nummer})',
    loesen: 'Dissocier',
  },
  // **Die Zwischenablage** (#1168).
  ablage: {
    kopieren: 'Copier ({zahl})',
    ausschneiden: 'Couper ({zahl})',
    einfuegen: 'Coller',
  },

  // **Die Spalten der Kartenliste** (#1168).
  listenspalten: {
    ueberschrift: 'Colonnes de la liste',
    jeSpiel: 'Ce choix s’applique à tous les sets du même jeu.',
    filtern: 'Rechercher une colonne',
    gezeigt: '{gezeigt} colonnes sur {alle}',
    ausgeblendet:
      '{anzahl} autres sont masquées — trouve-les avec le champ ci-dessus.',
    genauer: 'D’autres correspondent — affine la recherche pour les voir.',
    nachOben: 'déplacer plus à gauche',
    nachUnten: 'déplacer plus à droite',
    sortieren: 'Trier par',
    garNicht: 'Ordre du set',
    aufsteigend: 'Croissant',
    absteigend: 'Décroissant',
    ohneVorlage:
      'Le paquet de jeu est illisible. Les colonnes proviennent des cartes elles-mêmes — sans intitulés et sans l’ordre du paquet.',
    spalteAusblenden: 'Masquer la colonne',
  },

  verlauf: verlaufEn,
  massen: {
    ...massenvorschauEn,
    keinFeld: '— choisir un champ —',
    keinWert: '— choisir une valeur —',
    ueberschrift: 'Modification groupée',
    erklaerung:
      'Définit un champ sur toutes les cartes cochées — en une seule opération. Si quelque chose échoue, aucune carte n’est modifiée.',
    auswaehlen: 'Cocher pour la modification groupée',
    mehrereWaehlen: 'En choisir plusieurs',
    ausgewaehlt: 'cochées',
    keineAuswahl:
      'Aucune carte cochée. Coches-en plusieurs dans la liste pour les modifier ensemble.',
    feld: 'Champ',
    wert: 'Nouvelle valeur',
    anwenden: 'Appliquer',
    betroffen:
      'La modification a échoué pour ces cartes. Aucune carte n’a été modifiée :',
    gescheitert: 'échec',
    ohneWert: 'pas encore de valeur',
  },
  stilTitel: 'Style',
  stilWaehlen: 'Style de ce set',
  stilLeer:
    'Aucun style installé. Sans style, l’aperçu se rabat sur un plan de remplacement.',
  stilFrage: 'Passer à ce style ?',
  stilSpielwechsel: 'Le nouveau style appartient à un autre jeu.',
  stilUngedeutet:
    'Le nouveau jeu ne connaît pas ces champs. Leurs valeurs restent dans la base — elles ne seront plus dessinées :',
  stilVerloreneOptionen:
    'Le nouveau style n’a pas ces options de style. Leurs valeurs seront perdues :',
  stilUebernehmen: 'Changer de style',
  stilAbbrechen: 'Garder le style',
  stilKeiner:
    'Le set désigne un style qui n’est pas installé — choisis-en un parmi ceux installés.',
  stiloptionen: 'Options de style',
  stiloptionenSichtbar: '{zahl} sur {gesamt}',
  stiloptionenWeitere: 'Autres options ({zahl})',
  marktTitel: 'Où trouver cette carte',
  marktErklaerung:
    'Ces liens sont construits sur ton Mac — à partir du nom de la carte ou d’un identifiant du fournisseur. Rien ne part avant ton clic, dans le navigateur de ton choix. Aucune clé n’est nécessaire ; savoir si quelque chose y est en vente dépend du fournisseur.',
  marktGenau: 'mène exactement à cette carte',
  marktSuche: 'mène à une recherche',
  marktNurAnbieter: 'mène uniquement au fournisseur',
  umschlagAnnehmen: 'Carte depuis une enveloppe',
  umschlagAngenommen: 'La carte se trouve maintenant dans ce set.',
}
