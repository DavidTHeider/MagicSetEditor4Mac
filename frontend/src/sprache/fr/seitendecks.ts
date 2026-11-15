// The **Decks** page — Werkstück 7 of phase 4.6
// (#1293).
export const decks = {
  titel: 'Decks',
  erklaerung:
    'Assemblages de tes propres cartes et d’impressions réelles. Un deck renvoie aux cartes au lieu de les copier — une correction dans le set d’origine l’atteint donc aussitôt.',

  leer: 'Aucun deck pour l’instant. « Créer un deck », au-dessus, en commence un. Un deck appartient à un jeu ; sans cela, il n’y aurait aucun format selon lequel le vérifier.',

  neu: 'Créer un deck',
  name: 'Nom',
  spiel: 'Jeu',
  notiz: 'Note',
  anlegen: 'Créer',
  ausSetTitel: 'Ou à partir d’un set',
  ausSetErklaerung:
    'Chaque carte du set devient un emplacement du deck principal — référencée, non copiée : si tu la corriges dans le set, elle est corrigée dans le deck.',
  ausSetSet: 'Set',
  ausSetMenge: 'Exemplaires par carte',
  ausSetAnlegen: 'Créer à partir du set',
  ausSetKeinSet:
    'Aucun set contenant des cartes pour l’instant — un deck issu d’un set en nécessite un.',

  hauptdeck: 'Deck principal',
  sideboard: 'Réserve',
  zaehlung: '{haupt} dans le deck principal, {side} dans la réserve',
  waehlen: 'Ouvrir',
  offenMarke: 'ouvert',
  umbenennen: 'Renommer',
  neuerName: 'Nouveau nom',
  umbenennenTitel: 'Renommer {name}',
  nameSichern: 'Enregistrer',
  umbenennenAbbrechen: 'Conserver le nom',
  duplizieren: 'Dupliquer',
  kopieTitel: 'Dupliquer {name}',
  kopieName: 'Nom de la copie',
  kopieVorschlag: '{name} (copie)',
  kopieAnlegen: 'Créer la copie',
  kopieAbbrechen: 'Ne pas dupliquer',
  aufWunschliste: 'Ajouter ce qui manque à la liste de souhaits',
  wunschlisteErgebnis:
    'Nouveaux sur la liste de souhaits : {neu} · déjà présents : {schon}',
  loeschen: 'Supprimer le deck',
  handlungenAn: 'Actions — {name}',
  handlungenKurz: '…',
  deckmenue: 'Deck',
  loeschenFrage: 'Supprimer ce deck ? Les cartes elles-mêmes restent.',
  loeschenJa: 'Supprimer',
  loeschenNein: 'Annuler',

  plaetze: 'Emplacements',
  plaetzeLeer: 'Aucune carte dans le deck pour l’instant.',
  bezug: 'Carte ou impression',
  menge: 'Quantité',
  bereich: 'Section',
  setWaehlen: 'Set',
  karteWaehlen: 'Carte',
  karteSuchen: 'Chercher une carte',
  karteKeine: 'Aucune carte ne correspond à « {suche} ».',
  setLeer:
    'Aucun set ouvert — les cartes de tes propres sets ne peuvent être choisies qu’une fois qu’il y en a un.',
  ohneSet: 'Choisir un set…',
  ohneKarte: 'Choisir une carte…',
  setOhneKarten: 'Ce set n’a pas encore de carte.',
  kurve: 'Courbe de mana',
  kurveHinweis:
    'Combien de cartes par valeur de mana. Sont comptées celles dont la valeur est connue : pour une impression Scryfall, celle qui y est indiquée, sinon celle calculée à partir du coût. Les terrains n’ont pas de coût et n’y figurent donc pas — pas plus qu’une carte personnelle dont le champ de coût est vide.',
  farben: 'Couleurs',
  farbeWeiss: 'Blanc',
  farbeBlau: 'Bleu',
  farbeSchwarz: 'Noir',
  farbeRot: 'Rouge',
  farbeGruen: 'Vert',
  farbenHinweis:
    'Une carte bicolore compte dans les deux couleurs — la somme est donc supérieure au deck.',
  karten: 'Cartes',
  gebietFormate: 'Formats',
  formateKnopf: 'Formats …',

  // **Taking a deck from someone else's profile** — Werkstück 12 of Phase
  // 4.12 (#1351).
  uebernahme: 'Prendre un deck d’un autre profil',
  auswahlLoesen: 'Afficher tous les decks',
  leiste: {
    bezeichnung: 'Chiffres du deck',
    sideboard: 'Réserve {zahl}',
    pruefung: 'Vérification',
    ohneFormat: 'sans format',
    nichtGeprueft: 'non vérifié',
    fehlt: 'Manque dans la collection',
    unbekannt: '{zahl} inconnues',
    nichtAbgeglichen: 'non comparé',
  },
  uebernahmeErklaerung:
    'Peut en récupérer quiconque peut passer dans l’autre profil. Si son propriétaire utilise un second facteur, il faut un partage sur ce profil — « lecteur » suffit. Ne sont pas repris : les emplacements qui pointent vers une carte propre à l’autre profil, et le format — ce sont des lignes de l’autre côté. Ce qui n’a pas été repris figure ensuite ici.',
  uebernahmeProfil: 'Identifiant de l’autre profil',
  uebernahmeDeck: 'Identifiant du deck',
  uebernahmeProfilwahl: 'De quel profil ?',
  uebernahmeProfilWaehlen: 'Choisir un profil…',
  uebernahmeProfilEintrag: '{nutzer} · {profil}',
  uebernahmeKeinProfil: 'Il n’y a pas d’autre profil sur cet ordinateur.',
  uebernahmeDeckwahl: 'Quel deck ?',
  uebernahmeDeckWaehlen: 'Choisir un deck…',
  uebernahmeKeinDeck: 'Il n’y a aucun deck dans ce profil.',
  uebernahmeFachleute: 'Indiquer soi-même les identifiants',
  uebernahmeFachleuteErklaerung:
    'Le profil et le deck peuvent aussi être désignés par leurs identifiants — tels que la ligne de commande et le service les présentent. L’autorisation est la même.',
  uebernahmeHolen: 'Le prendre',
  uebernahmeGeschrieben:
    '« {name} » figure maintenant parmi tes decks — {karten}.',
  uebernahmeNichtMit: 'Non repris :',

  gebietKarten: 'Cartes',
  gebietAuswertung: 'Analyse',
  gebietPruefung: 'Vérification',
  gebieteDeck: 'Zones de ce deck',
  legen: 'Ajouter au deck',
  nehmen: 'Retirer',
  druck: 'Impression (identifiant Scryfall)',
  anschlussOhneDeck:
    'La carte « {name} » attend — ouvre un deck ou crées-en un, elle apparaîtra alors dans son choix de cartes.',
  holenAus: 'Récupérer des cartes depuis',
  holenSammlung: 'Ma collection',
  holenSet: 'Un set',
  wegeTitel: 'Récupérer et transmettre',
  mitspielen: 'Jouer avec ce deck',
  anschlussWartet:
    'La carte « {name} » est venue de ta collection — choisis la quantité et la section, elle ira alors dans le deck.',

  format: 'Format',
  formatWaehlen: 'Choisir un format',
  ohneFormat: 'Aucun format — non vérifié',
  formatHinweis:
    'Un deck sans format est non vérifié, et non faux. Le format doit appartenir au jeu du deck.',

  pruefen: 'Vérifier selon le format',
  zulaessig: 'Légal.',
  unzulaessig: 'Non légal.',
  ungeprueftZahl: '{zahl} cartes ne sont pas vérifiées.',
  ungepruefterklaerung:
    'Non vérifié ne veut pas dire non légal : une carte que tu as conçue toi-même n’est répertoriée dans aucun format au monde. « Restreint » ne s’oppose pas non plus au deck — le nombre d’exemplaires autorisés d’une carte restreinte figure dans les règles de tournoi, et ce programme ne les a pas lues.',

  legal: 'légal',
  not_legal: 'non répertorié',
  restricted: 'restreint',
  banned: 'banni',
  ungeprüft: 'non vérifié',

  bestand: 'Ce que tu as',
  vorhanden: 'Possédé',
  gebraucht: 'Nécessaire',
  fehlend: 'Manquant',
  unbekannt: 'Inconnu',
  bestandVollstaendig: 'Tout est là.',
  bestandFehlt: '{zahl} cartes manquent.',

  // **Die Zuordnung** — Werkstück U9 von Phase 5.1 (#1420).
  gebietZuordnung: 'Attribution',
  zuordnungen: 'Besoin et attribution',
  zuordnungenLaden: 'Chargement…',
  zuordnungenLeer:
    'Aucune carte dans le deck pour l’instant. Ce qui s’affiche ici provient de ses emplacements.',
  zuordnungHinweis:
    'Une liste de deck ne réserve rien : deux decks contenant les mêmes trois cartes sont tous deux complets. Est attribué ce que tu mets expressément de côté pour ce deck — est libre ce qu’il reste ensuite.',
  zuordnungszahlen:
    '{vorhanden} sur {gebraucht} possédés · {frei} libres · {zugeordnet} attribués · {reserviert} réservés',
  zuordnungPlatz: 'Section',
  zuordnungExemplar: 'Exemplaire',
  zuordnungExemplarWaehlen: 'Choisir un exemplaire',
  zuordnungMenge: 'Quantité',
  zuordnen: 'Attribuer',
  zuordnungLoesen: 'Libérer',
  abweichend: 'autre impression',
  abweichendErklaerung:
    'Cet exemplaire est une autre impression que celle de l’emplacement. Choisi expressément ainsi — le même nom n’est pas la même carte.',
  fehlkarten: 'Cartes manquantes sous forme de liste',
  fehlkartenliste: 'Les cartes manquantes',
  fehlkartenKeine: 'Il ne manque rien.',

  typen: 'Types',
  kennzahlen: 'Chiffres clés',
  laender: 'Terrains',
  durchschnitt: 'Valeur de mana moyenne',
  keineZahl: '—',
  quellen: 'Sources de couleur',
  quellenHinweis:
    'Sont comptés les terrains ayant l’un des cinq types de terrain de base — ils produisent leur mana sans texte de règles (CR 305.6). Un terrain qui nomme sa couleur dans son texte n’est pas listé ici : l’interpréter serait deviner.',

  weniger: '−',
  wenigerVon: 'Un de moins',
  mehr: '+',
  mehrVon: 'Un de plus',
  insSideboard: 'Vers la réserve',
  insHauptdeck: 'Vers le deck principal',
  suche: 'Rechercher des cartes',
  deckbereich: 'Le deck',
  analyse: 'Analyse',
  analyseLeer:
    'L’analyse, la vérification et l’attribution du deck ouvert s’affichent ici.',
  stapelTitel: 'Depuis la collection',
  stapelWartet:
    '{name} depuis la collection, dans aucun deck pour l’instant : {liste}.',
  stapelLegen: 'Tout mettre dans le deck',
  stapelTeilweise: '{gelegt} sur {gesamt} sont dans le deck, le reste non :',
  stapelOhneDruck:
    '« {name} » n’a pas d’identifiant Scryfall — un deck n’accepte que des impressions qui en ont un.',
  stapelGrund: '« {name} » : {grund}',
  eintauschen: 'Échanger les {anzahl}',
  deckplatz: 'Atelier de deck',
  pruefungOhneFormat:
    'Sans format, rien n’est vérifié. Choisis un format pour le deck — d’ici là, il est non vérifié, et non faux.',
  pruefungLaeuft: 'Vérification…',
  postenUnzulaessig: 'non légal',
  postenUnzulaessigErklaerung:
    'Cette carte n’est pas autorisée dans le format choisi. L’onglet « Vérification » dit précisément pourquoi.',
  keinDeckOffen:
    'Aucun deck ouvert. Choisis-en un à gauche ou crées-en un nouveau.',

  formate: 'Formats',
  formateLeer:
    'Aucun format pour l’instant. Un format est ici un enregistrement et non une liste intégrée — il est créé, pas livré.',
  bauart: 'Construction',
  bauarten: 'Constructions documentées',
  bauartHinweis:
    'Les chiffres figurent dans les Comprehensive Rules. Ils constituent une suggestion ; chacun peut ensuite être modifié.',
  scryfallname: 'Nom Scryfall',
  scryfallnameHinweis:
    'La clé dans les legalities de Scryfall, par exemple « standard ». Vide signifie que Scryfall ne connaît pas ce format, et que chaque carte reste non vérifiée.',
  mindestgroesse: 'Au moins',
  hoechstgroesse: 'Au plus',
  sideboardgroesse: 'Réserve',
  sideboardOffen: 'Réserve sans limite',
  hoechstzahlGleich: 'Même carte',
  grenzenHinweis:
    'Zéro signifie aucune limite. Pour la réserve, zéro signifie aucune réserve.',
  formatNeu: 'Créer un format',
  formatSichern: 'Enregistrer le format',
  formatLoeschen: 'Supprimer le format',
  formatLoeschenFrage: 'Supprimer {name} ?',
  formatLoeschenJa: 'Supprimer',
  formatLoeschenHinweis:
    'Les decks qui l’utilisaient restent — ils sont ensuite non vérifiés.',
  umschlagTitel: 'Transmettre sous forme d’enveloppe',
  umschlagErklaerung:
    'Une enveloppe transporte un deck vers un autre ordinateur — signée, avec somme de contrôle et sans compte. Ce qui arrive devient un nouveau deck ; rien n’est écrasé. Le nom, le jeu et chaque emplacement avec une impression **sont transmis** ; la note, l’origine et les emplacements avec des cartes personnelles **restent ici** — leur identifiant ne vaut que dans cette bibliothèque.',
  umschlagSchnueren: 'Enregistrer le deck ouvert',
  umschlagAnnehmen: 'Deck depuis une enveloppe',
  umschlagOhneDeck: 'Ouvre d’abord un deck.',
  umschlagGeschnuert: 'Enregistré.',
  umschlagAngenommen: 'Importé',
  einfuhrTitel: 'Importer une liste de deck',
  einfuhrMenue: 'Importer une liste de deck …',
  einfuhrErklaerung:
    'Colle une liste de deck — une carte par ligne, « 4 Lightning Bolt (M11) 149 ». Les titres structurent la liste : ce qui suit « Sideboard » est la réserve. L’aperçu n’écrit rien ; seule l’importation place les lignes dans le deck ouvert, et elle n’en retire rien.',
  einfuhrFeld: 'Liste de deck',
  einfuhrVorschau: 'Aperçu',
  einfuhrUebernehmen: 'Importer la liste de deck',
  einfuhrLeer: 'Colle d’abord une liste.',
  einfuhrOffen: 'ouvert',
  einfuhrBilanz: 'Deck principal / réserve',
  einfuhrGelegt: 'Importé',
  einfuhrStueck: 'exemplaires',
  einfuhrNichtsOffen: 'Aucune ligne n’est sans ambiguïté — rien à placer.',
} as const
