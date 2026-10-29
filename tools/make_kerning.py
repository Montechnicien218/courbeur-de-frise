#!/usr/bin/env python3
"""Extrait l'approche (kerning) des polices de page/vendor/fonts/ dans kerning.json.
Extracts the kerning of the fonts in page/vendor/fonts/ into kerning.json.

  pip install fonttools
  python3 tools/make_kerning.py

opentype.js 1.3 ne lit pas les tables GPOS « extension » de la plupart des polices
récentes : sans ce fichier, les lettres seraient posées sans approche (« AV » trop écarté).
Les polices elles-mêmes ne sont pas modifiées.
opentype.js 1.3 cannot read the GPOS "extension" lookups used by most recent fonts;
without this file, letters would be set without kerning. The fonts themselves are unchanged.
"""
import json, pathlib
from fontTools.ttLib import TTFont

ROOT = pathlib.Path(__file__).resolve().parent.parent
FONTS = ROOT / 'page' / 'vendor' / 'fonts'
# lettres, chiffres et ponctuation courants en français, allemand, anglais
CHARS = [chr(c) for c in range(0x20, 0x7f)] + list('àâäáãåçéèêëíìîïñóòôöõúùûüÿýœæøÀÂÄÁÃÅÇÉÈÊËÍÌÎÏÑÓÒÔÖÕÚÙÛÜŸÝŒÆØß«»‘’“”‚„–—…€°·')


def pairs_of(font):
    cmap = font.getBestCmap()
    order = font.getGlyphOrder()
    gid = {n: i for i, n in enumerate(order)}
    wanted = {cmap[ord(c)] for c in CHARS if ord(c) in cmap}
    out = {}
    if 'GPOS' not in font:
        return out
    gpos = font['GPOS'].table
    idx = set()
    for fr in gpos.FeatureList.FeatureRecord:
        if fr.FeatureTag == 'kern':
            idx.update(fr.Feature.LookupListIndex)
    for li in sorted(idx):
        lk = gpos.LookupList.Lookup[li]
        subs = []
        for st in lk.SubTable:
            if lk.LookupType == 9:
                if st.ExtensionLookupType != 2:
                    continue
                st = st.ExtSubTable
            elif lk.LookupType != 2:
                continue
            subs.append(st)
        done = set()   # une paire trouvée dans une sous-table n'est plus cherchée dans les suivantes
        for st in subs:
            cov = st.Coverage.glyphs
            if st.Format == 1:
                for g1, ps in zip(cov, st.PairSet):
                    if g1 not in wanted:
                        continue
                    for r in ps.PairValueRecord:
                        g2 = r.SecondGlyph
                        if g2 not in wanted or (g1, g2) in done:
                            continue
                        done.add((g1, g2))
                        v = getattr(r.Value1, 'XAdvance', 0) if r.Value1 else 0
                        if v:
                            out[(g1, g2)] = out.get((g1, g2), 0) + v
            elif st.Format == 2:
                c1 = st.ClassDef1.classDefs if st.ClassDef1 else {}
                c2 = st.ClassDef2.classDefs if st.ClassDef2 else {}
                for g1 in cov:
                    if g1 not in wanted:
                        continue
                    row = st.Class1Record[c1.get(g1, 0)].Class2Record
                    for g2 in wanted:
                        if (g1, g2) in done:
                            continue
                        done.add((g1, g2))
                        rec = row[c2.get(g2, 0)]
                        v = getattr(rec.Value1, 'XAdvance', 0) if rec.Value1 else 0
                        if v:
                            out[(g1, g2)] = out.get((g1, g2), 0) + v
    return {(gid[a], gid[b]): v for (a, b), v in out.items() if v}


res = {}
for f in sorted(FONTS.glob('*.woff')):
    p = pairs_of(TTFont(str(f)))
    left = {}
    for (a, b), v in sorted(p.items()):
        left.setdefault(a, []).extend([b, v])
    res[f.name] = left   # { glyphe gauche : [glyphe droite, valeur, glyphe droite, valeur, …] }
    print(f.name, len(p), 'paires')
(FONTS / 'kerning.json').write_text(json.dumps(res, separators=(',', ':')), encoding='utf-8')
