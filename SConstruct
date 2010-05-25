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
color_package = env.Go('bin/goray/color', 'bin/goray/color.go')
vector_package = env.Go('bin/goray/vector', 'bin/goray/vector.go')

packages = [
    color_package,
    fmath_package,
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
