import os
import sys
import random

lines = ["cam 0 0 150 -50 -50 50 50 -50 50 -50 50 50 50 50 50"]

def construct_string(start, args):
    output = start + " "
    for arg in args:
        output += (str(arg) + " ")
    return output[:len(output)-1]

def gen_random_material():
    kar = random.uniform(0, 0.3)
    kag = random.uniform(0, 0.3)
    kab = random.uniform(0, 0.3)

    kdr = random.random()
    kdg = random.random()
    kdb = random.random()

    ksr = random.random()
    ksg = random.random()
    ksb = random.random()

    ksp = random.randint(2, 50)

    krr = random.uniform(0, 0.5)
    krg = random.uniform(0, 0.5)
    krb = random.uniform(0, 0.5)

    output = construct_string("mat", [kar, kag, kab, kdr, kdg, kdb, ksr, ksg, ksb, ksp, krr, krg, krb])
    lines.append(output)

def put_random_directional_light():
    x = random.randint(-200, 200)
    y = random.randint(-200, 200)
    z = random.randint(-50, 200)
    r = random.uniform(0.05, 0.8)
    g = random.uniform(0.05, 0.8)
    b = random.uniform(0.05, 0.8)
    output = construct_string("ltd", [x, y, z, r, g, b])
    lines.append(output)

def put_random_point_light():
    x = random.randint(-200, 200)
    y = random.randint(-200, 200)
    z = random.randint(-50, 50)
    #x = -200
    #y = 0
    #z = 100
    r = random.uniform(0.05, 0.8)
    g = random.uniform(0.05, 0.8)
    b = random.uniform(0.05, 0.8)
    output = construct_string("ltp", [x, y, z, r, g, b])
    lines.append(output)

def draw_random_sphere():
    radius = random.randint(5, 50)
    x = random.randint(-100, 100)
    y = random.randint(-100, 100)
    z = random.randint(-250, -50)
    output = construct_string("sph", [x, y, z, radius])
    lines.append(output)

def draw_random_triangle():
    ax = random.randint(-100, 100)
    ay = random.randint(-100, 100)
    az = random.randint(-100, 0)

    bx = random.randint(-100, 100)
    by = random.randint(-100, 100)
    bz = random.randint(-100, 0)

    cx = random.randint(-100, 100)
    cy = random.randint(-100, 100)
    cz = random.randint(-100, 0)

    output = construct_string("tri", [ax, ay, az, bx, by, bz, cx, cy, cz])
    lines.append(output)

def main():
    f = open("new_scene.txt", "w")
    for _ in range(1):
        put_random_directional_light()
    for _ in range(3):
        put_random_point_light()
    for _ in range(2):
        gen_random_material()
        draw_random_triangle()
    for _ in range(1):
        gen_random_material()
        #lines.append(construct_string("sph", [-30, 0, -50, 40]))
        #lines.append(construct_string("sph", [60, 0, -150, 60]))
        draw_random_sphere()
    for line in lines:
        f.write(line + "\n")

if __name__ == "__main__":
    main()
