# Journal des versions · Changelog

## 1.18.0 — 30.10.2026 — affichage optimisé · faster display
- Affichage en couches : bouger la souris ne redessine plus que les règles, les poignées et le curseur (environ 18 fois moins de travail par image) · Layered display: moving the mouse only redraws rulers, handles and cursor (about 18× less work per frame).
- Aperçu réaliste mis en mémoire : déplacer la vue glisse une image au lieu de tout recalculer (environ 25 fois moins de travail), rendu net dès que le zoom s'arrête · Realistic preview cached: panning slides an image instead of recomputing everything (about 25× less work), sharp render as soon as zooming stops.
- Quadrillage dessiné par le navigateur, en fond · Grid drawn by the browser as a background.
- Tracés allégés en vue éloignée, export toujours complet · Lighter paths when zoomed out, export always at full detail.
- Canevas des repères à faible latence · Low-latency overlay canvas.

## 1.17.1 — 30.10.2026 — tâche en cours · busy indicator
- Une roue tourne et le curseur passe en attente pendant les calculs, l'export, l'envoi à LightBurn et l'ouverture de fichiers · A spinner and a busy cursor show during computing, export, sending to LightBurn and opening files.

## 1.17.0 — 30.10.2026 — ergonomie · ergonomics
- Poignées du début et de la fin du bord à glisser directement sur le dessin · Start and end handles of the edge can be dragged on the drawing.
- Fichiers à déposer sur le dessin : forme, motif, projet ou police · Drop files onto the drawing: shape, pattern, project or font.
- Raccourcis clavier façon logiciel laser, liste avec la touche « ? » · Laser-software keyboard shortcuts, list with the "?" key.
- Éléments comme des calques : masquer (non exporté), verrouiller, dupliquer, glisser pour réordonner · Items like layers: hide (not exported), lock, duplicate, drag to reorder.
- Réglages rapides : molette, Maj/Alt + flèches, glisser sur le nom d'un réglage · Quick settings: wheel, Shift/Alt + arrows, drag on a setting's name.
- Panneau allégé : explications derrière un « i », réglages rares dans « Avancé » · Lighter panel: explanations behind an "i", rare settings under "Advanced".
- Galerie : recherche, favoris, récents, aperçu en direct au survol · Gallery: search, favourites, recent, live preview on hover.
- Enregistrement automatique et reprise du travail au démarrage · Autosave and resume at start-up.
- Règles en mm sur le dessin et position du curseur · mm rulers on the drawing and cursor position.

## 1.16.0 — 30.10.2026 — illusions d'optique · optical illusions
- 6 frises d'illusion d'optique dans la bibliothèque : cubes 3D, mur du café, vagues op-art, tunnel, ruban plié, damier bombé · 6 optical-illusion friezes in the library: 3D cubes, café wall, op-art waves, tunnel, folded ribbon, bulging checkerboard.

## 1.15.1 — 30.10.2026 — matériaux · materials
- Aperçu réaliste : deux nouveaux matériaux, contreplaqué et métal brossé · Realistic preview: two new materials, plywood and brushed metal.

## 1.15.0 — 30.10.2026 — nouvelle interface · new interface
- Interface façon logiciel laser : barre d'outils en haut, objets à gauche, dessin en grand au centre, propriétés à droite, barre d'état en bas · Laser-studio layout: toolbar on top, objects on the left, large drawing in the middle, properties on the right, status bar at the bottom.
- Thème sombre par défaut, thème clair en un clic (mémorisé) · Dark theme by default, light theme one click away (remembered).
- Sections repliables, zoom flottant sur le dessin, légende et messages intégrés · Collapsible sections, floating zoom on the drawing, built-in legend and messages.
- Export et envoi à LightBurn toujours à portée de main, en bas du panneau de droite · Export and send-to-LightBurn always at hand, at the bottom of the right panel.

## 1.14.2 — 30.10.2026 — plus rapide · faster
- Démarrage environ 6 fois plus rapide (2,8 s → 0,5 s) : les tracés SVG sont lus directement au lieu d'être mesurés point par point · About 6× faster start-up (2.8 s → 0.5 s): SVG paths are read directly instead of being measured point by point.
- Les motifs et les formes SVG s'ouvrent beaucoup plus vite, surtout les fichiers détaillés · SVG patterns and shapes open much faster, especially detailed files.
- Frappe d'un texte plus fluide : la liste des frises et la barre d'infos ne sont plus reconstruites à chaque touche · Smoother typing: the frieze list and info bar are no longer rebuilt on every keystroke.

## 1.14 — 30.10.2026 — projet, coins, aperçu · project, corners, preview
- Mise à jour sans relancer le programme : quand courbeur.html est remplacé, un bandeau propose de recharger, en gardant le travail en cours · Update without restarting: when courbeur.html is replaced, a banner offers to reload, keeping the current work.
- Projet : enregistrer et rouvrir tout un travail (fichier .courbeur, polices comprises) · Project: save and reopen a whole job (.courbeur file, fonts included).
- Bibliothèque intégrée : 15 frises et 4 ornements de coin, à choisir en un clic · Built-in library: 15 friezes and 4 corner ornaments, one click away.
- Ornements de coin : un motif dans chaque coin de la forme, pointe vers le coin · Corner ornaments: a pattern in every corner of the shape.
- Polices monotrait (4) pour une gravure rapide, exportées en traits · 4 single-line fonts for fast engraving, exported as strokes.
- Textes libres : droits ou en cercle, n'importe où sur la forme, à glisser · Free text: straight or in a circle, anywhere on the shape, draggable.
- Aperçu réaliste (bois clair, noyer, ardoise, acrylique noir) et export d'une image PNG · Realistic preview (light wood, walnut, slate, black acrylic) and PNG export.

## 1.13.1 — 29.10.2026 — licence
- Nouvelle licence : tous droits réservés, utilisation gratuite (voir LICENSE). Les versions 1.12.0 à 1.13.0 restent sous licence MIT · New licence: all rights reserved, free to use (see LICENSE). Versions 1.12.0 to 1.13.0 remain under the MIT licence.
- Exemple : horloge murale (examples/horloge) · Example: wall clock.

## 1.13 — 29.10.2026 — textes · text
- Textes le long du bord : clic droit sur le dessin › « Ajouter un texte ici », ou « + Ajouter un texte » · Text along the edge: right-click › “Add a text here”, or “+ Add a text”.
- Hauteur des lettres, distance à la ligne suivie (mm), espacement, position le long du bord, rotation de 180° · Letter height, distance from the line (mm), spacing, position along the edge, 180° turn.
- Le texte se glisse directement sur le dessin et se remet à l'endroit tout seul · Drag the text on the drawing; it turns upright automatically.
- Lettres posées sur une ligne de base parallèle au bord : elles gardent leur largeur dans les virages et les creux · Letters sit on a baseline parallel to the edge and keep their width in bends.
- 7 polices intégrées (SIL OFL) et chargement de ta propre police .ttf / .otf / .woff · 7 built-in fonts (SIL OFL) plus your own .ttf / .otf / .woff.
- Lettres nettoyées et fusionnées à l'export, sur leur propre calque LightBurn · Letters cleaned and merged on export, on their own LightBurn layer.

## 1.12 — 29.10.2026 — première version publique · first public release
- Interface en français et en anglais, avec bascule FR / EN · French and English interface, with FR / EN switch.
- Lanceur bilingue selon la langue de Windows · Launcher follows the Windows display language.
- Motif et forme d'exemple originaux (`examples/`) · Original example pattern and shape.
- Code publié sous licence MIT · Code released under the MIT licence.

## 1.11
- Fusion des formes à l'export : un seul contour par calque (Clipper) · Shape merging on export: one outline per layer.
- Une frise jointe prend le calque de sa voisine · A joined frieze takes its neighbour's layer.

## 1.10
- Une frise ajoutée sur un contour déjà entouré forme un anneau intérieur complet · A frieze added on an already surrounded outline forms a full inner ring.
- « Joindre les bordures » ne joint qu'avec les frises du même anneau · "Join the edges" only joins friezes of the same ring.

## 1.9
- La frise jointe garde la hauteur de sa voisine · A joined frieze keeps its neighbour's height.
- Option « Adoucir les coins » · "Soften corners" option.

## 1.8
- Empilement automatique des frises au même endroit · Automatic stacking of friezes in the same place.

## 1.7
- Temps de réaction divisé par deux environ · Response time roughly halved.

## 1.6
- Orientation du motif horizontale ou verticale · Horizontal or vertical pattern orientation.

## 1.5
- Calcul réparti sur tous les cœurs du processeur · Computation spread over all CPU cores.

## 1.4
- Début et fin saisissables, bouton « Joindre les bordures » · Typed start and end, "Join the edges" button.

## 1.3
- Correction « Out of Memory », lanceur stable, page séparée (`courbeur.html`) · "Out of Memory" fix, stable launcher, separate page.

## 1.2
- Répétition automatique pour garder la taille du motif · Automatic repeat to keep the pattern size.

## 1.1
- Rendu accéléré par la carte graphique · GPU-accelerated drawing.

## 1.0
- Plusieurs frises, calques LightBurn, envoi direct à LightBurn, dossiers d'échange, bibliothèque de motifs, réglages mémorisés, icône dans la zone de notification · Multiple friezes, LightBurn layers, direct sending, exchange folders, pattern library, saved settings, tray icon.
