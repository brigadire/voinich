#!/usr/bin/env python3
"""Optional future read-only runtime preflight. No CVAT writes, credentials or review."""
import argparse
import json
from urllib.parse import urlsplit
from urllib.request import urlopen

if __name__=='__main__':
    p=argparse.ArgumentParser();p.add_argument('--base-url',required=True);args=p.parse_args()
    parsed=urlsplit(args.base_url)
    if parsed.scheme not in {'http','https'} or not parsed.netloc or parsed.username or parsed.password:
        p.error('provide HTTP(S) base URL without embedded credentials')
    with urlopen(args.base_url.rstrip('/')+'/api/server/about',timeout=15) as response:data=json.load(response)
    print(json.dumps({'version':data.get('version','NOT_REPORTED'),'note':'Version only; verify UI lock/hide and empty image-XML roundtrip on this instance before review.'},ensure_ascii=False))
