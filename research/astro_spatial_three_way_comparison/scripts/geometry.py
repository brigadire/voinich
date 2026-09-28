"""Convex shape overlaps; raw geometries never modified."""
from __future__ import annotations
import math

FIELDS = ('bbox_x1','bbox_y1','bbox_x2','bbox_y2')

def box(row): return tuple(float(row[k]) for k in FIELDS)

def shape(row, kind='AABB', rotation=None, convention='actual'):
    x1,y1,x2,y2 = box(row)
    cx,cy = (x1+x2)/2,(y1+y2)/2
    w,h = x2-x1,y2-y1
    angle = float(rotation if rotation is not None else row.get('rotation') or row.get('orientation_angle') or 0)%180
    if kind == 'AABB': angle = 0.
    if convention == 'orientation_normalized_proxy' and kind == 'ROTATED_RECTANGLE' and min(angle,180-angle) > .01:
        w,h = max(w,h),min(w,h)
    rad = math.radians(angle)
    if kind == 'ELLIPSE_ENVELOPE_PROXY':
        local = [(w/2*math.cos(2*math.pi*i/256),h/2*math.sin(2*math.pi*i/256)) for i in range(256)]
    else:
        local = [(-w/2,-h/2),(w/2,-h/2),(w/2,h/2),(-w/2,h/2)]
    points = [(cx+x*math.cos(rad)-y*math.sin(rad), cy+x*math.sin(rad)+y*math.cos(rad)) for x,y in local]
    return {'cx':cx,'cy':cy,'width':w,'height':h,'angle':angle,'kind':kind,'convention':convention,'polygon':points}

def area(poly):
    return abs(sum(p[0]*q[1]-q[0]*p[1] for p,q in zip(poly,poly[1:]+poly[:1])))/2 if len(poly)>2 else 0.

def cross(a,b,p): return (b[0]-a[0])*(p[1]-a[1])-(b[1]-a[1])*(p[0]-a[0])

def intersection(subject, clip):
    result = subject
    for a,b in zip(clip,clip[1:]+clip[:1]):
        source = result
        result = []
        if not source: break
        previous = source[-1]
        prevdist = cross(a,b,previous)
        for current in source:
            dist = cross(a,b,current)
            if (dist >= -1e-8) != (prevdist >= -1e-8):
                denominator = prevdist-dist
                if abs(denominator)>1e-16:
                    t = prevdist/denominator
                    result.append((previous[0]+t*(current[0]-previous[0]),previous[1]+t*(current[1]-previous[1])))
            if dist >= -1e-8: result.append(current)
            previous,prevdist = current,dist
    return result

def iou(left, right):
    la,ra = area(left),area(right)
    common = area(intersection(left,right))
    den = la+ra-common
    return max(0.,min(1.,common/den)) if den>0 else 0.

def axial(a,b): return abs((a-b+90)%180-90)

def metrics(left,right,dimensions):
    dist = math.hypot(left['cx']-right['cx'],left['cy']-right['cy'])
    return {'iou':iou(left['polygon'],right['polygon']), 'center_distance_px':dist,
            'center_distance_panel_norm':dist/math.hypot(*dimensions),
            'center_distance_shape_norm':dist/math.hypot(right['width'],right['height']),
            'width_relative_error':abs(left['width']-right['width'])/right['width'],
            'height_relative_error':abs(left['height']-right['height'])/right['height'],
            'width_ratio':left['width']/right['width'], 'height_ratio':left['height']/right['height'],
            'axial_angle_error_deg':axial(left['angle'],right['angle']) if right['kind']!='AABB' else None}
