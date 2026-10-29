# Courbeur de frise · Frieze Bender

[Français](#français) · [English](#english)

![Courbeur de frise : deux anneaux de frise sur une forme en arc brisé](screenshots/apercu-fr.png)

---

## Français

Le **Courbeur de frise** fait suivre un motif SVG (une frise) le long d'un bord de ta forme DXF ou SVG, puis envoie le résultat à **LightBurn** pour la découpe ou la gravure laser. LightBurn ne sait pas courber un motif le long d'une courbe : cet outil le fait, en gardant la taille et la qualité du motif.

**Essayer sans rien installer :** [version web](https://TON-COMPTE.github.io/courbeur-de-frise/), sans envoi direct à LightBurn ni dossiers d'échange.

### Ce qu'il fait

- Courbe le motif le long de n'importe quel bord, à l'intérieur ou à l'extérieur de la forme.
- Plusieurs frises indépendantes, chacune avec son bord, son motif, ses réglages et son calque LightBurn.
- **Répétition automatique** : le motif garde sa hauteur et se duplique pour remplir le bord, élément par élément.
- **Joindre les bordures** : les frises se raccordent bout à bout ; une frise de plus forme un anneau complet à l'intérieur.
- **Empilement** des frises au même endroit, **coins adoucis** dans les virages serrés, orientation horizontale ou verticale.
- **Fusion des formes** à l'export : un seul contour propre par calque, sans trait en double ni trou en gravure.
- Export **DXF** (coordonnées d'origine) et **SVG** en millimètres, ou **envoi direct à LightBurn**, posé sur la forme, sur le bon calque.
- Aperçu en temps réel, Précédent / Suivant (Ctrl+Z / Ctrl+Y), calcul réparti sur tous les cœurs du processeur.
- Interface en français et en anglais.

### Installation (Windows)

1. Télécharge la dernière version dans [Releases](../../releases) et décompresse-la dans un dossier fixe, par exemple `C:\Outils\Courbeur de frise\`.
2. Lance `CourbeurDeFrise.exe` : l'outil s'ouvre dans ton navigateur et une icône apparaît près de l'horloge (clic droit › Quitter pour l'arrêter).
3. Pour qu'il démarre avec Windows, sans ouvrir le navigateur : **Windows + R**, tape `shell:startup`, puis mets un raccourci vers `CourbeurDeFrise.exe` dans ce dossier.

> **Antivirus et SmartScreen.** Le `.exe` n'est pas encore signé, donc Windows ou certains antivirus peuvent l'avertir ou le bloquer. Tout le code est dans ce dépôt : tu peux le lire, et le compiler toi-même (voir plus bas). Si ton antivirus le bloque à tort, ajoute le dossier de l'outil à ses exclusions.

### Avec LightBurn (2.2 ou plus récent)

1. Dans LightBurn : **Web Tools › Manage User Tools › Add**, nom `Courbeur de frise`, adresse **`http://127.0.0.1:19580/lightburn`**. L'outil s'ouvre alors dans ton navigateur, et le menu de LightBurn reste libre.
2. Dans LightBurn, sélectionne ta forme puis **Fichier › Exporter** dans `Documents\Courbeur de frise\Forme` : elle apparaît toute seule dans l'outil.
3. Règle tes frises, puis **Envoyer à LightBurn** : la première fois, LightBurn demande d'autoriser « Courbeur de frise ».

L'envoi utilise l'[API REST de LightBurn](https://github.com/LightBurnSoftware/LightBurn-REST), encore marquée « Labs » (expérimentale) dans la version 2.2. Sans elle, les boutons **Télécharger DXF / SVG** fonctionnent toujours.

### Construire depuis les sources

Il faut [Python 3](https://www.python.org) et [Go 1.24](https://go.dev).

```
python3 build.py
cd launcher
GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -ldflags "-H windowsgui -s -w" -o ../dist/CourbeurDeFrise.exe .
```

`build.py` assemble la page (`page/courbeur.src.html` + Clipper + les exemples) dans `dist/courbeur.html` et `docs/index.html`. Le lanceur lit `courbeur.html` à côté de lui : pour mettre l'outil à jour, il suffit de remplacer ce fichier.

### Licence

Licence [MIT](LICENSE) © 2026 Christophe Tanniou. Composants tiers : voir [THIRD_PARTY_NOTICES.md](THIRD_PARTY_NOTICES.md). Les fichiers d'exemple (`examples/`) sont sous la même licence MIT.

LightBurn est une marque de LightBurn Software, LLC. Ce projet n'est ni affilié à LightBurn Software, ni approuvé par elle.

---

## English

**Frieze Bender** bends an SVG pattern (a frieze) along an edge of your DXF or SVG shape, then sends the result to **LightBurn** for laser cutting or engraving. LightBurn can't bend a pattern along a curve: this tool does, while keeping the pattern's size and quality.

**Try it without installing anything:** [web version](https://TON-COMPTE.github.io/courbeur-de-frise/), without direct sending to LightBurn or exchange folders.

![Frieze Bender: two rings of friezes on a pointed-arch shape](screenshots/apercu-en.png)

### What it does

- Bends the pattern along any edge, inside or outside the shape.
- Several independent friezes, each with its own edge, pattern, settings and LightBurn layer.
- **Automatic repeat**: the pattern keeps its height and is duplicated to fill the edge, element by element.
- **Join the edges**: friezes meet end to end; one more frieze makes a full ring inside.
- **Stacking** of friezes in the same place, **softened corners** in tight turns, horizontal or vertical orientation.
- **Shape merging** on export: one clean outline per layer, no double lines when cutting, no holes when engraving.
- **DXF** export (original coordinates) and **SVG** in millimetres, or **direct sending to LightBurn**, placed on the shape, on the right layer.
- Live preview, Undo / Redo (Ctrl+Z / Ctrl+Y), computation spread over all CPU cores.
- French and English interface.

### Installation (Windows)

1. Download the latest version from [Releases](../../releases) and unzip it into a fixed folder, for example `C:\Tools\Frieze Bender\`.
2. Run `CourbeurDeFrise.exe`: the tool opens in your browser and an icon appears next to the clock (right-click › Quit to stop it).
3. To start it with Windows, without opening the browser: **Windows + R**, type `shell:startup`, then put a shortcut to `CourbeurDeFrise.exe` in that folder.

> **Antivirus and SmartScreen.** The `.exe` isn't signed yet, so Windows or some antivirus software may warn about it or block it. All the code is in this repository: you can read it and build it yourself (see below). If your antivirus blocks it by mistake, add the tool's folder to its exclusions.

### With LightBurn (2.2 or later)

1. In LightBurn: **Web Tools › Manage User Tools › Add**, name `Frieze Bender`, address **`http://127.0.0.1:19580/lightburn`**. The tool then opens in your browser, and LightBurn's menu stays free.
2. In LightBurn, select your shape, then **File › Export** into `Documents\Frieze Bender\Shape`: it appears in the tool automatically.
3. Set up your friezes, then **Send to LightBurn**: the first time, LightBurn asks you to allow “Courbeur de frise”.

Sending uses the [LightBurn REST API](https://github.com/LightBurnSoftware/LightBurn-REST), still marked “Labs” (experimental) in version 2.2. Without it, the **Download DXF / SVG** buttons always work.

### Building from source

You need [Python 3](https://www.python.org) and [Go 1.24](https://go.dev).

```
python3 build.py
cd launcher
GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -ldflags "-H windowsgui -s -w" -o ../dist/CourbeurDeFrise.exe .
```

`build.py` assembles the page (`page/courbeur.src.html` + Clipper + the examples) into `dist/courbeur.html` and `docs/index.html`. The launcher reads `courbeur.html` next to itself: to update the tool, just replace that file.

### Licence

[MIT](LICENSE) licence © 2026 Christophe Tanniou. Third-party components: see [THIRD_PARTY_NOTICES.md](THIRD_PARTY_NOTICES.md). The example files (`examples/`) are under the same MIT licence.

LightBurn is a trademark of LightBurn Software, LLC. This project is not affiliated with or endorsed by LightBurn Software.
