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
root_packages = [
    env.Go('bin/fmath.go'),
    env.Go('bin/stack.go'),
]

goray_packages = [
    env.Go('bin/goray/background.go'),
    env.Go('bin/goray/bound.go'),
    env.Go('bin/goray/camera.go'),
    env.Go('bin/goray/color.go'),
    env.Go('bin/goray/integrator.go'),
    env.Go('bin/goray/light.go'),
    env.Go('bin/goray/material.go'),
    env.Go('bin/goray/matrix.go'),
    env.Go('bin/goray/object.go'),
    env.Go('bin/goray/primitive.go'),
    env.Go('bin/goray/ray.go'),
    env.Go('bin/goray/render.go'),
    env.Go('bin/goray/scene.go'),
    env.Go('bin/goray/surface.go'),
    env.Go('bin/goray/vector.go'),
    env.Go('bin/goray/vmap.go'),
    env.Go('bin/goray/volume.go'),
]

packages = root_packages + goray_packages
Alias('lib', packages)

program = env.GoProgram('bin/run-goray', 'bin/main.go')

Default(packages + [program])

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
