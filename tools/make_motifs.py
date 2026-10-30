#!/usr/bin/env python3
"""Dessine les motifs de la bibliothèque intégrée du Courbeur de frise.
Draws the patterns of the Frieze Bender built-in library.

  pip install shapely
  python3 tools/make_motifs.py

Chaque frise est une tuile répétée 6 fois (l'outil la reconnaît et la duplique élément par élément).
Chaque ornement de coin est dessiné pointe en bas : le bas du dessin se place vers le coin.
Each frieze is one tile repeated 6 times; each corner ornament is drawn pointing down (toward the corner).
"""
import json, math, pathlib
from shapely.geometry import Point, Polygon, LineString, box, MultiPolygon
from shapely.ops import unary_union
from shapely import affinity

ROOT = pathlib.Path(__file__).resolve().parent.parent
OUT = ROOT / 'page' / 'motifs'
OUT.mkdir(parents=True, exist_ok=True)
NOTE = '© 2026 Christophe Tanniou, tous droits réservés : libre pour vos réalisations, revente du fichier interdite'


def ell(cx, cy, rx, ry, ang=0):
    e = affinity.scale(Point(0, 0).buffer(1, 64), rx, ry)
    return affinity.translate(affinity.rotate(e, ang, origin=(0, 0)), cx, cy)


def line(pts, w, cap='round', join='round'):
    return LineString(pts).buffer(w / 2, cap_style=cap, join_style=join, quad_segs=16)


def circ(cx, cy, r):
    return Point(cx, cy).buffer(r, 64)


def ring(cx, cy, r, w):
    return circ(cx, cy, r + w / 2).difference(circ(cx, cy, r - w / 2))


def path_of(g):
    g = g.simplify(0.012, preserve_topology=True)
    polys = [g] if isinstance(g, Polygon) else list(getattr(g, 'geoms', []))
    d = []
    for p in polys:
        if p.is_empty:
            continue
        for r in [p.exterior] + list(p.interiors):
            c = list(r.coords)[:-1]
            d.append('M' + 'L'.join(f'{x:.2f},{y:.2f}' for x, y in c) + 'Z')
    return ''.join(d)


def frieze(name, T, H, tile_fn, n=6):
    """tile_fn(x0) -> géométrie d'une tuile ; on y ajoute les voisines pour un raccord parfait, puis on coupe à la tuile."""
    paths = []
    for k in range(n):
        x0 = k * T
        g = unary_union([tile_fn(x0 + dx) for dx in (-T, 0, T)]).intersection(box(x0, -1, x0 + T, H + 1))
        g = g.buffer(0)
        paths.append(f'<path fill-rule="evenodd" d="{path_of(g)}"/>')
    write(name, n * T, H, paths)


def write(name, W, H, paths):
    svg = (f'<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 {W:g} {H:g}">\n<!-- {name} - Courbeur de frise - {NOTE} -->\n'
           '<g fill="#3a3a3a">\n' + '\n'.join(paths) + '\n</g>\n</svg>\n')
    (OUT / f'{name}.svg').write_text(svg, encoding='utf-8')


def corner(name, geom, W=30, H=30):
    write(name, W, H, [f'<path fill-rule="evenodd" d="{path_of(geom.buffer(0))}"/>'])


# ---------- frises ----------
def grecque(x):
    rails = unary_union([box(x, 0.5, x + 20, 2), box(x, 18, x + 20, 19.5)])
    key = line([(x + 3, 18.5), (x + 3, 5.2), (x + 16, 5.2), (x + 16, 14.8), (x + 8, 14.8), (x + 8, 10), (x + 12, 10)], 1.5, cap='flat', join='mitre')
    return unary_union([rails, key])
frieze('grecque', 20, 20, grecque)


def entrelacs(x):
    T = 24
    s1 = [(x + i * T / 60, 8 + 5 * math.sin(2 * math.pi * i / 60)) for i in range(61)]
    s2 = [(x + i * T / 60, 8 - 5 * math.sin(2 * math.pi * i / 60)) for i in range(61)]
    return unary_union([line(s1, 1.6), line(s2, 1.6), circ(x + 6, 8, 1.5), circ(x + 18, 8, 1.5)])
frieze('entrelacs', 24, 16, entrelacs)


def heart(cx, cy, s):
    pts = []
    for i in range(120):
        t = 2 * math.pi * i / 120
        X = 16 * math.sin(t) ** 3
        Y = 13 * math.cos(t) - 5 * math.cos(2 * t) - 2 * math.cos(3 * t) - math.cos(4 * t)
        pts.append((cx + X * s, cy - Y * s))
    return Polygon(pts)
frieze('coeurs', 16, 14, lambda x: unary_union([heart(x + 8, 7.2, 0.36), circ(x + 0, 7, 1.0)]))


def star(cx, cy, R, r, n=5, rot=-90):
    pts = []
    for i in range(2 * n):
        a = math.radians(rot + 180 * i / n)
        rr = R if i % 2 == 0 else r
        pts.append((cx + rr * math.cos(a), cy + rr * math.sin(a)))
    return Polygon(pts)
frieze('etoiles', 18, 16, lambda x: unary_union([star(x + 9, 8.6, 7.2, 2.9), circ(x, 8, 1.0)]))
frieze('chevrons', 12, 12, lambda x: line([(x, 9.5), (x + 6, 2.5), (x + 12, 9.5)], 2.2, cap='flat', join='mitre'))


def festons(x):
    arch = circ(x + 8, 10, 7.6).difference(circ(x + 8, 10, 5.8)).intersection(box(x - 1, 0, x + 17, 10))
    return unary_union([arch, box(x, 8.6, x + 16, 10), circ(x + 8, 6.4, 1.4)])
frieze('festons', 16, 10, festons)


def dentelle(x):
    g = unary_union([box(x, 8.5, x + 14, 12), circ(x + 7, 7.6, 4.6)])
    for h in [circ(x + 7, 7.3, 1.7), circ(x + 2.2, 10.3, 0.7), circ(x + 11.8, 10.3, 0.7), circ(x + 7, 2.9, 0.0001)]:
        g = g.difference(h)
    return g
frieze('dentelle', 14, 12, dentelle)


def losanges(x):
    outer = Polygon([(x + 7, 0.6), (x + 13.2, 6), (x + 7, 11.4), (x + 0.8, 6)])
    inner = Polygon([(x + 7, 2.6), (x + 11, 6), (x + 7, 9.4), (x + 3, 6)])
    dot = Polygon([(x + 7, 4.6), (x + 8.6, 6), (x + 7, 7.4), (x + 5.4, 6)])
    return unary_union([outer.difference(inner), dot, box(x, 5.5, x + 14, 6.5)])
frieze('losanges', 14, 12, losanges)
frieze('torsade', 8, 10, lambda x: ell(x + 4, 5, 6.2, 2.6, -55).difference(ell(x + 4, 5, 4.0, 0.7, -55)))


def epis(x):
    g = [box(x, 7.5, x + 18, 8.5)]
    for k, cx in enumerate([3, 9, 15]):
        g.append(ell(x + cx, 5.6, 3.0, 1.25, -35))
        g.append(ell(x + cx + 3, 10.4, 3.0, 1.25, 35))
    return unary_union(g)
frieze('epis', 18, 16, epis)


def fleurs(x):
    cx, cy = x + 10, 8
    petals = [ell(cx + 3.6 * math.cos(math.radians(a)), cy + 3.6 * math.sin(math.radians(a)), 3.0, 1.9, a) for a in range(-90, 270, 72)]
    f = unary_union(petals + [circ(cx, cy, 2.0)]).difference(circ(cx, cy, 1.0))
    return unary_union([f, circ(x, 8, 1.1), circ(x + 20 - 0.0, 8, 1.1)])
frieze('fleurs', 20, 16, fleurs)
frieze('pointilles', 10, 4, lambda x: unary_union([circ(x + 2.2, 2, 1.4), line([(x + 5.6, 2), (x + 8.6, 2)], 1.6)]))


# ---------- illusions d'optique ----------
def cubes3d(x):
    """cubes empilés (Necker / blocs basculants) : face gauche pleine, face droite hachurée, dessus vide"""
    T = 14.0
    sz = T / math.sqrt(3)
    cx, cy = x + T / 2, sz + 0.2
    w = T / 2
    top, ur, lr, bot, ll, ul, c = (cx, cy - sz), (cx + w, cy - sz / 2), (cx + w, cy + sz / 2), (cx, cy + sz), (cx - w, cy + sz / 2), (cx - w, cy - sz / 2), (cx, cy)
    left = Polygon([ul, c, bot, ll]).buffer(-0.28, join_style=2)
    right = Polygon([c, ur, lr, bot]).buffer(-0.28, join_style=2)
    bw = (w - 0.56) / 5
    stripes = unary_union([box(cx + 0.28 + (2 * k) * bw, -1, cx + 0.28 + (2 * k + 1) * bw, 40) for k in range(3)]).intersection(right)
    rim = Polygon([top, ur, c, ul]).exterior.buffer(0.25, join_style=2).intersection(Polygon([top, ur, c, ul]))
    return unary_union([left, stripes, rim])
frieze('illusion-cubes', 14, 14 / math.sqrt(3) * 2 + 0.4, cubes3d)


def mur_cafe(x):
    """mur du café (Gregory) : rangées de carreaux décalées d'un demi-carreau ; les joints parallèles paraissent penchés"""
    s, m, g = 7.0, 0.7, 0.35
    rows = [0.0, s / 2, 0.0]
    g_ = []
    for r, sh in enumerate(rows):
        y0 = r * (s + m + 2 * g)
        for k in (-1, 0, 1):
            g_.append(box(x + sh + k * 2 * s, y0, x + sh + k * 2 * s + s, y0 + s))
        if r < len(rows) - 1:
            g_.append(box(x - 1, y0 + s + g, x + 2 * s + 1, y0 + s + g + m))
    return unary_union(g_)
frieze('illusion-cafe', 14, 3 * 7 + 2 * (0.7 + 0.7), mur_cafe)


def vagues_op(x):
    """vagues op-art : bandes dont l'épaisseur ondule, décalée d'une bande à l'autre : la frise semble bomber"""
    T, H, n = 24.0, 16.0, 7
    out = []
    for k in range(n):
        yc = (k + 0.5) * H / n
        ph = k * math.pi / (n - 1)
        top, bot = [], []
        for i in range(97):
            xx = x + T * i / 96
            t = 0.35 + 0.95 * (0.5 + 0.5 * math.sin(2 * math.pi * i / 96 + ph))
            top.append((xx, yc - t / 2))
            bot.append((xx, yc + t / 2))
        out.append(Polygon(top + bot[::-1]))
    return unary_union(out)
frieze('illusion-vagues', 24, 16, vagues_op)


def tunnel(x):
    """tunnel : carrés emboîtés dont le centre fuit vers la droite, effet de profondeur"""
    T = 16.0
    out = []
    sizes, thick, off = [15.4, 11.0, 6.6], [1.0, 0.85, 0.7], 0.5
    for j, (a, t) in enumerate(zip(sizes, thick)):
        cx, cy = x + T / 2 + j * off, T / 2
        outer = box(cx - a / 2, cy - a / 2, cx + a / 2, cy + a / 2)
        out.append(outer.difference(outer.buffer(-t, join_style=2)))
    cx = x + T / 2 + 3 * off
    out.append(box(cx - 1.1, T / 2 - 1.1, cx + 1.1, T / 2 + 1.1))
    return unary_union(out)
frieze('illusion-tunnel', 16, 16, tunnel)


def ruban(x):
    """ruban plié en zigzag : montées pleines, descentes en traits, le ruban paraît en relief"""
    T, H, b = 16.0, 14.0, 3.4
    y0, y1 = 1.5, H - 1.5 - b
    up = Polygon([(x, y1), (x + T / 2, y0), (x + T / 2, y0 + b), (x, y1 + b)])
    dn = Polygon([(x + T / 2, y0), (x + T, y1), (x + T, y1 + b), (x + T / 2, y0 + b)])
    lines = unary_union([line([(x + T / 2 - 1, y0 + d - 1 * (y1 - y0) / (T / 2)), (x + T + 1, y1 + d + 1 * (y1 - y0) / (T / 2))], 0.55, cap='flat') for d in (0.45, 1.7, 2.95)]).intersection(dn)
    return unary_union([up.buffer(-0.15, join_style=2), lines])
frieze('illusion-ruban', 16, 14, ruban)


def damier_bombe(x):
    """damier bombé : colonnes plus larges au milieu de chaque tuile, le damier semble gonfler"""
    T, rows, rh = 24.0, 3, 5.0
    n = 8
    w = [1 + 0.55 * math.sin(math.pi * (i + 0.5) / n) for i in range(n)]
    k = T / sum(w)
    xs = [0.0]
    for v in w:
        xs.append(xs[-1] + v * k)
    out = []
    for r in range(rows):
        for i in range(n):
            if (i + r) % 2 == 0:
                out.append(box(x + xs[i], r * rh, x + xs[i + 1], (r + 1) * rh))
    return unary_union(out)
frieze('illusion-damier', 24, 15, damier_bombe)

# ---------- ornements de coin (pointe en bas, vers le coin) ----------
corner('coin-fleuron', unary_union([
    ell(15, 13, 3.2, 9.5), ell(8.8, 18.5, 2.6, 7.5, -48), ell(21.2, 18.5, 2.6, 7.5, 48),
    ell(5.5, 10.5, 1.6, 4.2, -20), ell(24.5, 10.5, 1.6, 4.2, 20), circ(15, 26.5, 2.3)]).difference(ell(15, 13, 1.0, 5.0)))


def coin_feuilles():
    g = []
    for sgn in (-1, 1):
        a = math.radians(-90 + sgn * 45)
        ux, uy = math.cos(a), math.sin(a)
        base = (15, 27.5)
        g.append(line([base, (15 + ux * 18, 27.5 + uy * 18)], 1.2))
        for k, t in enumerate([4.5, 8.5, 12.5, 16.5]):
            px, py = 15 + ux * t, 27.5 + uy * t
            side = 1 if k % 2 == 0 else -1
            ang = math.degrees(a) + side * 50
            off = 2.6
            nx, ny = math.cos(math.radians(ang)), math.sin(math.radians(ang))
            g.append(ell(px + nx * off, py + ny * off, 3.0, 1.3, ang))
    g += [circ(15, 27.5, 2.0), circ(15, 22.5, 1.2)]
    return unary_union(g)
corner('coin-feuilles', coin_feuilles())


def coin_etoile():
    g = [star(15, 15, 13.5, 3.2, 4, -90), affinity.rotate(star(15, 15, 8.5, 2.4, 4, -90), 45, origin=(15, 15)), ring(15, 15, 5.2, 1.0)]
    return unary_union(g).difference(circ(15, 15, 1.4))
corner('coin-etoile', coin_etoile())


def spiral(cx, cy, r0, turns, sgn):
    pts = []
    for i in range(160):
        t = turns * 2 * math.pi * i / 159
        r = r0 * (1 - t / (turns * 2 * math.pi) * 0.82)
        pts.append((cx + sgn * r * math.cos(t), cy - r * math.sin(t)))
    return pts


def coin_volute():
    g = [line([(15, 28), (15, 20)], 1.6)]
    for s in (-1, 1):
        pts = [(15, 20), (15 + s * 4, 15)] + spiral(15 + s * 8.5, 11, 6, 1.6, s)
        g.append(line(pts, 1.5))
    g += [circ(15, 28, 1.8), ell(15, 15.5, 1.4, 3.2)]
    return unary_union(g)
corner('coin-volute', coin_volute())

LIB = {
    'frises': [('motif-feuilles', 'Guirlande de feuilles'), ('motif-perles', 'Perles'), ('motif-tulipes', 'Tulipes'),
               ('grecque', 'Grecque'), ('entrelacs', 'Entrelacs'), ('festons', 'Festons'), ('dentelle', 'Dentelle'),
               ('coeurs', 'Cœurs'), ('etoiles', 'Étoiles'), ('fleurs', 'Fleurs'), ('epis', 'Épis de blé'),
               ('losanges', 'Losanges'), ('torsade', 'Torsade'), ('chevrons', 'Chevrons'), ('pointilles', 'Pointillés')],
    'illusions': [('illusion-cubes', 'Cubes 3D'), ('illusion-cafe', 'Mur du café'), ('illusion-vagues', 'Vagues op-art'),
                  ('illusion-tunnel', 'Tunnel'), ('illusion-ruban', 'Ruban plié'), ('illusion-damier', 'Damier bombé')],
    'coins': [('coin-fleuron', 'Fleuron'), ('coin-feuilles', 'Branches de feuilles'), ('coin-volute', 'Volutes'), ('coin-etoile', 'Étoile')],
}
(OUT / 'index.json').write_text(json.dumps(LIB, ensure_ascii=False, indent=1), encoding='utf-8')
print('ok', len(list(OUT.glob('*.svg'))), 'motifs')
