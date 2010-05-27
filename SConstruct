#!/usr/bin/env python

import os

test_build = ('test' in COMMAND_LINE_TARGETS)

# Set up environment
env = Environment(TOOLS=['default', 'go'], GOBUILDDIR='bin')
env.VariantDir('bin', 'src')

test_sources = []
def _add_test(source, test):
    if test_build:
        source.append(test)
    test_sources.append(test)

# Main build
fmath_package = env.Go('bin/fmath', 'bin/fmath.go')
stack_package = env.Go('bin/stack', 'bin/stack.go')

bound_package = env.Go('bin/goray/bound', 'bin/goray/bound.go')
camera_package = env.Go('bin/goray/camera', 'bin/goray/camera.go')
color_package = env.Go('bin/goray/color', 'bin/goray/color.go')
material_package = env.Go('bin/goray/material', 'bin/goray/material.go')
matrix_package = env.Go('bin/goray/matrix', 'bin/goray/matrix.go')
ray_package = env.Go('bin/goray/ray', 'bin/goray/ray.go')
vector_package = env.Go('bin/goray/vector', 'bin/goray/vector.go')

packages = [
    bound_package,
    camera_package,
    color_package,
    fmath_package,
    material_package,
    matrix_package,
    ray_package,
    stack_package,
    vector_package,
]

Default(packages)

# Testing
#testenv = env.Clone(ENV=os.environ)
#testenv.GoTests('bin/_gotest.go', test_sources)
#test_package = testenv.Go('bin/_gotest.go')
#test_harness = testenv.GoProgram('bin/_gotest', [test_package] + packages)
#AlwaysBuild(testenv.Alias('test', [test_harness], 'bin/_gotest'))
#if test_build:
    #AlwaysBuild('bin/_gotest.go')
    #AlwaysBuild(test_harness)
    #AlwaysBuild(packages)
    #testenv.Decider('make')
