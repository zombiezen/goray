#!/usr/bin/env python
#
#   go.py
#   SCons Go Tools
#   
#   Copyright (c) 2010, Ross Light.
#   All rights reserved.
#
#   Redistribution and use in source and binary forms, with or without
#   modification, are permitted provided that the following conditions are met:
#
#       Redistributions of source code must retain the above copyright notice,
#       this list of conditions and the following disclaimer.
#
#       Redistributions in binary form must reproduce the above copyright
#       notice, this list of conditions and the following disclaimer in the
#       documentation and/or other materials provided with the distribution.
#
#       Neither the name of the SCons Go Tools nor the names of its contributors
#       may be used to endorse or promote products derived from this software
#       without specific prior written permission.
#
#   THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS"
#   AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
#   IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE
#   ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE
#   LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR
#   CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF
#   SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS
#   INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN
#   CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE)
#   ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE
#   POSSIBILITY OF SUCH DAMAGE.
#

import os
import posixpath
import re
import subprocess

from SCons import Util
from SCons.Action import Action
from SCons.Scanner import Scanner
from SCons.Builder import Builder

def _subdict(d, keys):
    result = {}
    for key in keys:
        try:
            result[key] = d[key]
        except KeyError:
            pass
    return result

def splitext(path):
    rightmost_sep = path.rfind(os.path.sep)
    try:
        dot = path.rfind(os.path.extsep, rightmost_sep + 1)
    except ValueError:
        return path, ''
    else:
        return path[:dot], path[dot:]

# PLATFORMS

_valid_platforms = frozenset((
    ('darwin', '386'),
    ('darwin', 'amd64'),
    ('freebsd', '386'),
    ('freebsd', 'amd64'),
    ('linux', '386'),
    ('linux', 'amd64'),
    ('linux', 'arm'),
    ('nacl', '386'),
    ('windows', '386'),
))
_archs = {'amd64': '6', '386': '8', 'arm': '5'}

def _get_platform_info(env, goos, goarch):
    info = {}
    if (goos, goarch) not in _valid_platforms:
        raise ValueError("Unrecognized platform: %s, %s" % (goos, goarch))
    info['archname'] = _archs[goarch]
    info['pkgroot'] = os.path.join(env['ENV']['GOROOT'], 'pkg', goos + '_' + goarch)
    info['gc'] = os.path.join(env['ENV']['GOBIN'], info['archname'] + 'g')
    info['ld'] = os.path.join(env['ENV']['GOBIN'], info['archname'] + 'l')
    info['as'] = os.path.join(env['ENV']['GOBIN'], info['archname'] + 'a')
    info['cc'] = os.path.join(env['ENV']['GOBIN'], info['archname'] + 'c')
    info['pack'] = os.path.join(env['ENV']['GOBIN'], 'gopack')
    return info

def _get_host_platform(env):
    newenv = env.Clone()
    newenv['ENV'].pop('GOOS', None)
    newenv['ENV'].pop('GOARCH', None)
    config = _parse_config(_run_goenv(newenv))
    return config['GOOS'], config['GOARCH']

# COMPILER

_package_pat = re.compile(r'package\s+(\w+)\s*;?', re.UNICODE)
_spec_pat = re.compile(r'\s*(\.|\w+)?\s*\"(.*?)\"\s*;?', re.UNICODE)
def _get_imports(node):
    source = node.get_text_contents()
    while source:
        source = source.lstrip()
        if source.startswith('//'):
            source = _after_token(source, '\n')
        elif source.startswith('/*'):
            source = _after_token(source, '*/')
        elif source.startswith('package') and source[7].isspace():
            m = _package_pat.match(source)
            if m:
                source = source[m.end():]
            else:
                return
        elif source.startswith('import') and not source[6].isalnum():
            source = source[6:].lstrip()
            if source.startswith('('):
                # Compound import
                source = source[1:]
                while True:
                    m = _spec_pat.match(source)
                    if m:
                        yield m.group(2)
                        source = source[m.end():]
                    else:
                        break
                source = source.lstrip()
                if not source.startswith(')'):
                    return
                source = source[1:]
            else:
                # Single import
                m = _spec_pat.match(source)
                if m:
                    yield m.group(2)
                    source = source[m.end():]
                else:
                    return
        else:
            # Once we see any other statement, the imports are done.
            return

def _after_token(s, tok):
    try:
        return s[s.index(tok) + len(tok):]
    except ValueError:
        return ''

def _go_scan_func(node, env, paths):
    package_paths = env['GO_LIBPATH'] + [env['GO_PKGROOT']]
    result = []
    for package_name in _get_imports(node):
        if package_name.startswith("./"):
            result.append(env.File(package_name + _go_object_suffix(env, [])))
            continue
        # Search for import
        package_dir, package_name = posixpath.split(package_name)
        subpaths = [posixpath.join(p, package_dir) for p in package_paths]
        # Check for a static library
        package = env.FindFile(
            package_name + os.path.extsep + 'a',
            subpaths,
        )
        if package is not None:
            result.append(package)
            continue
        # Check for a build result
        package = env.FindFile(
            package_name + os.path.extsep + env['GO_ARCHNAME'],
            subpaths,
        )
        if package is not None:
            result.append(package)
            continue
    return result

go_scanner = Scanner(function=_go_scan_func, skeys=['.go'])

def _gc_emitter(target, source, env):
    if env.get('GO_STRIPTESTS', False):
        return (target, [s for s in source if not str(s).endswith('_test.go')])
    else:
        return (target, source)

def _ld_scan_func(node, env, path):
    obj_suffix = os.path.extsep + env['GO_ARCHNAME']
    result = []
    for child in node.children():
        if str(child).endswith(obj_suffix) or str(child).endswith('.a'):
            result.append(child)
    return result

def _go_object_suffix(env, sources):
    return os.path.extsep + env['GO_ARCHNAME']

def _go_program_prefix(env, sources):
    return env['PROGPREFIX']

def _go_program_suffix(env, sources):
    return env['PROGSUFFIX']

go_compiler = Builder(
    action=Action('$GO_GCCOM', '$GO_GCCOMSTR'),
    emitter=_gc_emitter,
    suffix=_go_object_suffix,
    ensure_suffix=True,
    src_suffix='.go',
)
go_linker = Builder(
    action=Action('$GO_LDCOM', '$GO_LDCOMSTR'),
    prefix=_go_program_prefix,
    suffix=_go_program_suffix,
    src_builder=go_compiler,
    single_source=True,
    source_scanner=Scanner(function=_ld_scan_func, recursive=True),
)
go_assembler=Builder(
    action=Action('$GO_ACOM', '$GO_ACOMSTR'),
    suffix=_go_object_suffix,
    ensure_suffix=True,
    src_suffix='.s',
)
gopack = Builder(
    action=Action('$GO_PACKCOM', '$GO_PACKCOMSTR'),
    suffix='.a',
    ensure_suffix=True,
)

## CONFIGURATION

def _get_PATH(env):
    if isinstance(env['ENV']['PATH'], (list, tuple)):
        return list(env['ENV']['PATH'])
    else:
        return env['ENV']['PATH'].split(os.path.pathsep)

def _get_gobin():
    try:
        return os.environ['GOBIN']
    except KeyError:
        home = os.environ.get('HOME')
        if home:
            return os.path.join(os.environ['GOROOT'], 'bin')
        else:
            return None

def _parse_config(data):
    result = {}
    for line in data.splitlines():
        name, value = line.split('=', 1)
        if name.startswith('export '):
            name = name[len('export '):]
        result[name] = value
    return result

def _run_goenv(env):
    proc = subprocess.Popen(
        ['make', '--no-print-directory', '-f', 'Make.inc', 'go-env'],
        cwd=os.path.join(env['ENV']['GOROOT'], 'src'),
        stdin=subprocess.PIPE,
        stdout=subprocess.PIPE,
        stderr=subprocess.PIPE,
    )
    stdout, stderr = proc.communicate()
    return stdout

## TESTING

def _get_package_info(env, node):
    package_name = splitext(node.name)[0]
    # Find import path
    for path in env['GO_LIBPATH']:
        search_dir = env.Dir(path)
        if node.is_under(search_dir):
            return package_name, splitext(search_dir.rel_path(node))[0]
    # Try under launch directory as a last resort
    search_dir = env.Dir(env.GetLaunchDir())
    if node.is_under(search_dir):
        return package_name, "./" + splitext(search_dir.rel_path(node))[0]
    else:
        raise ValueError("Package %s not found in library path" % (package_name))

def _read_func_names(f):
    magic = '\tfunc "".'
    started = False
    pkg = ''
    for line in f:
        if started:
            if line.startswith(magic):
                yield (pkg, line[len(magic):line.index(' ', len(magic))])
            elif line.lstrip().startswith('package'):
                pkg = line.split()[1]
            elif line.startswith('$$'):
                # We only want the first section
                break
        elif line.startswith('$$'):
            started = True
    f.close()

def gotest(target, source, env):
    # Compile test information
    import_list = [[_get_package_info(env, snode)[1], False] for snode in source]
    tests = []
    benchmarks = []
    for i, snode in enumerate(source):
        proc = None
        # Start reading functions
        if str(snode).endswith('.a'):
            proc = subprocess.Popen([env['GO_PACK'], 'p', str(snode), '__.PKGDEF'], stdout=subprocess.PIPE)
            names = _read_func_names(proc.stdout)
        else:
            names = _read_func_names(open(str(snode)))
        # Handle names
        for (package, ident) in names:
            name = package + '.' + ident
            info = (i, ident, name)
            if ident.startswith('Test'):
                tests.append(info)
                import_list[i][1] = True # mark as used
            elif ident.startswith('Bench'):
                benchmarks.append(info)
                import_list[i][1] = True # mark as used
        # Wait on gopack subprocess
        if proc is not None:
            proc.wait()
            if proc.returncode != 0:
                return proc.returncode
    # Write out file
    f = open(str(target[0]), 'w')
    try:
        f.write("package main\n\n")
        # Imports
        f.write("import \"testing\"\n")
        f.write("import __regexp__ \"regexp\"\n")
        f.write("import (\n")
        for i, (import_path, used) in enumerate(import_list):
            if used:
                f.write("\tt%04d \"%s\"\n" % (i, import_path))
        f.write(")\n\n")
        # Test array
        f.write("var tests = []testing.InternalTest{\n")
        for pkg_num, ident, name in tests:
            f.write("\ttesting.InternalTest{\"%s\", t%04d.%s},\n" %
                (name, pkg_num, ident))
        f.write("}\n\n")
        # Benchmark array
        f.write("var benchmarks = []testing.InternalBenchmark{\n")
        for pkg_num, ident, name in benchmarks:
            f.write("\ttesting.InternalBenchmark{\"%s\", t%04d.%s},\n" %
                (name, pkg_num, ident))
        f.write("}\n\n")
        # Main function
        f.write("func main() {\n")
        f.write("\ttesting.Main(__regexp__.MatchString, tests, benchmarks)\n")
        f.write("}\n")
    finally:
        f.close()

go_tester = Builder(
    action=Action(gotest, '$GO_TESTCOMSTR'),
    suffix='.go',
    ensure_suffix=True,
    src_suffix=_go_object_suffix,
)

# API

def GoTarget(env, goos, goarch):
    config = _get_platform_info(env, goos, goarch)
    env['ENV']['GOOS'] = goos
    env['ENV']['GOARCH'] = goarch
    env['GO_GC'] = config['gc']
    env['GO_LD'] = config['ld']
    env['GO_A'] = config['as']
    env['GO_PACK'] = config['pack']
    env['GO_ARCHNAME'] = config['archname']
    env['GO_PKGROOT'] = config['pkgroot']

def generate(env):
    if 'HOME' not in env['ENV']:
        env['ENV']['HOME'] = os.environ['HOME']
    # Now set up the environment
    env.Append(ENV=_subdict(os.environ, ['GOROOT', 'GOBIN']))
    env['ENV'].setdefault('GOBIN', os.path.join(env['ENV']['GOROOT'], 'bin'))
    # Set up tools
    env.AddMethod(GoTarget, 'GoTarget')
    goos, goarch = _get_host_platform(env)
    env.GoTarget(goos, goarch)
    # Add builders and scanners
    env.Append(
        BUILDERS={
            'Go': go_compiler,
            'GoProgram': go_linker,
            'GoAssembly': go_assembler,
            'GoPack': gopack,
            'GoTest': go_tester,
        },
        SCANNERS=[go_scanner],
        GO_GCCOM='$GO_GC -o $TARGET ${_concat("-I ", GO_LIBPATH, "", __env__)} $GO_GCFLAGS $SOURCES',
        GO_LDCOM='$GO_LD -o $TARGET ${_concat("-L ", GO_LIBPATH, "", __env__)} $GO_LDFLAGS $SOURCE',
        GO_ACOM='$GO_A -o $TARGET $SOURCE',
        GO_PACKCOM='rm -f $TARGET ; $GO_PACK gcr $TARGET $SOURCES',
    )

def exists(env):
    return True
