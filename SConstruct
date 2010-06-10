#!/usr/bin/env python

import os
import subprocess

test_build = ('test' in COMMAND_LINE_TARGETS)

# Set up environment
env = Environment(TOOLS=['default', 'go'], GOLIBPATH=['build'])
env.VariantDir('build', 'src')

# Version info
def get_bzr_path():
    for path in os.environ['PATH'].split(os.path.pathsep):
        path = os.path.join(path, 'bzr')
        if os.path.exists(path):
            return path
    return None

def generate_buildversion(env, target, source):
    template = """\
// This file is automatically generated.
// DO NOT EDIT.
package buildversion

const Source = "bzr"
const RevNo = "{revno}"
const RevID = "{revision_id}"
const BranchNickname = "{branch_nick}"
const CleanWC = {clean}
"""
    bzr_path = get_bzr_path()
    f = open(str(target[0]), 'w')
    try:
        if bzr_path:
            subprocess.call([bzr_path, 'version-info', '--custom', '--template', template], stdout=f)
        else:
            template = template.replace('{revno}', 'archive')
            template = template.replace('{revision_id}', 'archive')
            template = template.replace('{branch_nick}', 'archive')
            template = template.replace('{clean}', '1')
            template = template.replace('bzr', 'archive')
            f.write(template)
    finally:
        f.close()

version_file = File('build/buildversion.go')
Command(version_file, [], generate_buildversion)
AlwaysBuild(version_file)

build_info_packages = [
    env.Go(version_file),
]

# Main build
root_packages = [
    env.Go('build/goray/fmath.go'),
    env.Go('build/goray/logging.go'),
    env.Go('build/goray/time.go'),
    env.Go('build/goray/stack.go'),
]

core_packages = [
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
    'photon',
    'primitive',
    'ray',
    'render',
    'scene',
    'surface',
    'vector',
    'vmap',
    'volume',
    'version',
]
core_packages = [env.Go('build/goray/core/%s.go' % name) for name in core_packages]

std_packages = [
    'integrators/directlight',
    'integrators/trivial',
    'lights/point',
	'materials/debug',
    'objects/mesh',
    'primitives/sphere',
]
std_packages = [env.Go('build/goray/std/%s.go' % name) for name in std_packages]

packages = build_info_packages + root_packages + core_packages + std_packages
Alias('lib', packages)
Alias('core', core_packages)
Alias('std', std_packages)

program = env.GoProgram('bin/goray', 'build/main.go')

Default(packages + [program])
