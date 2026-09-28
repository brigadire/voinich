"""Independent synthetic morphology generator. Never imported by inference."""
import itertools
import random

ROLES = ('INITIAL', 'MEDIAL', 'FINAL')
POSITIVE_FAMILIES = ('affix', 'internal', 'mixed', 'composition', 'partial', 'distractors')
NULL_FAMILIES = ('independent_random', 'length_matched', 'unigram_preserving',
                 'bigram_preserving', 'shuffled_assignments', 'repeated_fragment',
                 'role_frequency', 'size_matched')
HARD_FAMILIES = tuple('HN'+str(i) for i in range(1, 11))


def row_bag(strings, rng):
    ids = rng.sample(range(10**7, 10**8), len(strings))
    rows = [{'id': str(ids[i]), 'surface': s} for i, s in enumerate(strings)]
    rng.shuffle(rows)
    return rows


def inventory(rng, variant, count=5):
    # Disjoint general Unicode symbol inventories; no engine-known alphabet.
    starts = ((0x61, 0x391), (0x410, 0x2100), (0x2400, 0x2500), (0x2800, 0x2C00))
    a, b = starts[variant % len(starts)]
    # Skip potentially non-NFC code points by retaining stable symbols only.
    import unicodedata
    def chars(start):
        return [chr(i) for i in range(start, start+200)
                if unicodedata.normalize('NFC', chr(i)) == chr(i)
                and unicodedata.category(chr(i))[0] not in 'CZM'][:100]
    source, target = chars(a), chars(b)
    rng.shuffle(source)
    rng.shuffle(target)
    rules = []
    for ri, role in enumerate(ROLES):
        for j in range(count):
            index = ri*count+j
            sl, tl = 2+index%2, 2+(index+1)%2
            rules.append((role, ''.join(source[3*index:3*index+sl]),
                          ''.join(target[3*index:3*index+tl])))
    return rules


def bigram_shuffle(s, rng):
    # Random Euler trail with fixed endpoints and identical directed edge multiset.
    edges = {}
    for i in range(len(s)-1):
        edges.setdefault(s[i], []).append(s[i+1])
    for out in edges.values():
        rng.shuffle(out)
    stack, path = [s[0]], []
    while stack:
        if edges.get(stack[-1]):
            stack.append(edges[stack[-1]].pop())
        else:
            path.append(stack.pop())
    return ''.join(reversed(path))


def generate(seed, family='mixed', size=29, alphabet=0):
    rng = random.Random(seed)
    rules = inventory(rng, alphabet)
    by_role = {r: [x for x in rules if x[0] == r] for r in ROLES}
    combinations = list(itertools.product(range(5), repeat=3))
    rng.shuffle(combinations)
    held_count = min(24, 125-size)
    splits, truth = {}, {'rules': [list(r) for r in rules], 'family': family, 'splits': {}}
    for split, combos in [('train', combinations[:size]), ('heldout', combinations[size:size+held_count])]:
        terms, labels, pairs, paths = [], [], [], []
        for index, combo in enumerate(combos):
            rr = [by_role[ROLES[j]][combo[j]] for j in range(3)]
            if family == 'composition' and index % 3 == 0:
                rr.insert(2, by_role['MEDIAL'][(combo[1]+1)%5])
            a, b = ''.join(r[1] for r in rr), ''.join(r[2] for r in rr)
            if family == 'partial' and index % 5 == 0:
                b = b[:len(rr[0][2])]+'☄'+b[len(rr[0][2]):]
            terms.append(a)
            labels.append(b)
            pairs.append([a, b])
            paths.append({'term': a, 'label': b, 'rules': [list(r) for r in rr]})
        if family == 'affix':
            # Heldout still has recurrent stems; frozen role model is shared.
            pass
        if family == 'distractors':
            terms.extend(('◉' if split=='train' else '◍')+str(i)+'◈' for i in range(max(1, size//10)))
            labels.extend(('◌' if split=='train' else '◎')+str(i)+'◊' for i in range(max(1, size//10)))
        if family in NULL_FAMILIES or family in HARD_FAMILIES:
            original = list(labels)
            if family in {'independent_random', 'length_matched', 'size_matched', 'HN1', 'HN2', 'HN10'}:
                # Large inventory makes pair-specific mappings non-recurrent.
                labels = [''.join(chr(0x3400+rng.randrange(1000)) for _ in range(len(x))) for x in labels]
                if family == 'independent_random':
                    terms = [''.join(chr(0x4400+rng.randrange(1000)) for _ in range(rng.randrange(4, 12))) for _ in terms]
            elif family == 'unigram_preserving':
                bag = list(''.join(labels))
                rng.shuffle(bag)
                labels, at = [], 0
                for x in original:
                    labels.append(''.join(bag[at:at+len(x)]))
                    at += len(x)
            elif family == 'bigram_preserving':
                labels = [bigram_shuffle(x, rng) for x in labels]
            elif family == 'shuffled_assignments':
                # Surfaces deliberately unchanged. Only the latent pairing changes.
                rng.shuffle(labels)
            elif family in {'repeated_fragment', 'HN5'}:
                labels = ['◇◇'+''.join(chr(0x3400+rng.randrange(1000)) for _ in range(max(2, len(x)-2))) for x in labels]
            elif family in {'role_frequency', 'HN6'}:
                shuffled = [rng.sample(range(len(labels)), len(labels)) for _ in ROLES]
                chunks = [[r[2] for r in p['rules']] for p in paths]
                labels = [''.join(chunks[shuffled[j][i]][j] for j in range(3)) for i in range(len(labels))]
                if family == 'HN6':
                    labels = [''.join(reversed([r[2] for r in p['rules']])) for p in paths]
            elif family == 'HN3':
                labels = [x[1:]+x[:1] for x in labels]
            elif family == 'HN4' and split == 'heldout':
                labels = [''.join(chr(0x3400+ord(c)%1000) for c in x) for x in labels]
            elif family == 'HN7':
                labels = [x if i%2 else ''.join(chr(0x3400+ord(c)%1000) for c in x) for i, x in enumerate(labels)]
            elif family == 'HN8':
                terms, labels = [terms[0]]*len(terms), [labels[0]]*len(labels)
            elif family == 'HN9':
                # A complete Cartesian product admits multiple bijections.
                ab = list(itertools.product('abc', 'def'))
                terms = [('uv' if split=='train' else 'wx')+a+b for a, b in ab]
                labels = [('yz' if split=='train' else 'st')+a+b for a, b in itertools.product('ghi', 'jkl')]
            pairs = [[terms[i], labels[i]] for i in range(min(len(terms), len(labels)))]
            paths = []
        splits[split] = {'terms': row_bag(terms, rng), 'labels': row_bag(labels, rng)}
        truth['splits'][split] = {'assignments': pairs, 'paths': paths}
    if family in NULL_FAMILIES or family in HARD_FAMILIES:
        truth['rules'] = []
    return splits, truth
