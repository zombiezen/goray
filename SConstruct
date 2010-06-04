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
    'background',
    'bound',
    'camera',
    'color',
    'integrator',
    'kdtree',
    'light',
    'material',
    'matrix',
    'object',
    'partition',
    'primitive',
    'ray',
    'render',
    'scene',
    'surface',
    'vector',
    'vmap',
    'volume',
]
goray_packages = [env.Go('bin/goray/%s.go' % name) for name in goray_packages]

std_packages = [
    'integrators/trivial',
]
std_packages = [env.Go('bin/goray/std/%s.go' % name) for name in std_packages]

packages = root_packages + goray_packages + std_packages
Alias('lib', packages)
Alias('core', root_packages + goray_packages)
Alias('std', std_packages)

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
