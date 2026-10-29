#!/usr/bin/env python3
"""Assemble l'outil à partir des sources / Builds the tool from the sources.

  python3 build.py

Produit / produces:
  dist/courbeur.html      la page complète, à mettre à côté de CourbeurDeFrise.exe
  launcher/index.html     la même page, intégrée au .exe à la compilation (go build)
  docs/index.html         version web pour GitHub Pages (essai sans installation)
Ensuite / then:  cd launcher && GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -ldflags "-H windowsgui -s -w" -o ../dist/CourbeurDeFrise.exe .
"""
import base64, gzip, json, os, pathlib
ROOT = pathlib.Path(__file__).resolve().parent
read = lambda p: (ROOT / p).read_text(encoding='utf-8')
src = read('page/courbeur.src.html')
for key in ('__CLIPPER__', '__OPENTYPE__', '__FONTS__', '__SAMPLE_MOTIF__', '__SAMPLE_SHAPE__'):
    assert src.count(key) == 1, key
# polices intégrées : fichiers woff tels quels + approche (kerning) compressée
fdir = ROOT / 'page' / 'vendor' / 'fonts'
b64 = lambda b: base64.b64encode(b).decode('ascii')
fonts = {'files': {f.name: b64(f.read_bytes()) for f in sorted(fdir.glob('*.woff'))},
         'kern': b64(gzip.compress((fdir / 'kerning.json').read_bytes(), 9, mtime=0))}
page = (src.replace('__CLIPPER__', read('page/vendor/clipper.min.js'), 1)
           .replace('__OPENTYPE__', read('page/vendor/opentype.min.js'), 1)
           .replace('__FONTS__', json.dumps(fonts), 1)
           .replace('__SAMPLE_MOTIF__', json.dumps(read('examples/motif-tulipes.svg')), 1)
           .replace('__SAMPLE_SHAPE__', json.dumps(read('examples/forme-arc.dxf')), 1))
head = ('<!doctype html><html lang="fr"><head><meta charset="utf-8">'
        '<meta name="viewport" content="width=device-width,initial-scale=1">'
        '<link rel="icon" href="favicon.ico"><style>[hidden]{display:none!important}body{margin:0}</style></head><body>')
lic = '\n'.join([read('page/vendor/fonts/' + f.name) for f in sorted(fdir.glob('OFL-*.txt'))])
notice = ('<!-- Courbeur de frise : licence MIT. Composants tiers : Clipper (Boost), opentype.js (MIT), '
          'polices intégrées sous SIL Open Font License 1.1 :\n' + lic.replace('--', '- -') + '\n-->')
full = head + notice + page + '</body></html>\n'
(ROOT / 'dist').mkdir(exist_ok=True)
(ROOT / 'dist' / 'courbeur.html').write_text(full, encoding='utf-8')
(ROOT / 'launcher' / 'index.html').write_text(full.replace('href="favicon.ico"', 'href="/favicon.ico"'), encoding='utf-8')
(ROOT / 'docs').mkdir(exist_ok=True)
(ROOT / 'docs' / 'index.html').write_text(full, encoding='utf-8')
(ROOT / 'docs' / 'favicon.ico').write_bytes((ROOT / 'launcher' / 'icon.ico').read_bytes())
print('ok :', len(full) // 1024, 'Ko')
