#!/usr/bin/env python3

bl_addon_info = {
    'name': "Export goray Scene Format (.yaml)",
    'author': "Ross Light",
    'version': (0, 1),
    'blender': (2, 5, 5),
    'api': 33427,
    'location': "File > Export",
    'description': "Export to goray Scene Format (.yaml)",
    'category': "Import-Export",
}

import os

import bpy
from mathutils import *

indent = '  '

def write_file(path, scene):
    f = open(path, 'w')
    write_header(f)
    write_integrator(f)
    write_camera(f, scene)
    write_lights(f, scene)
    write_materials(f, scene)
    write_objects(f, scene)
    print('...', file=f)
    f.close()

def write_header(f):
    print("%YAML 1.2", file=f)
    print("%TAG !goray! tag:goray/", file=f)
    print("%TAG !std! tag:goray/std/", file=f)
    print("---", file=f)

def write_objects(f, scene):
    # TODO: Handle no objects
    print("objects:", file=f)
    for obj in scene.objects:
        if obj.type == 'MESH':
            write_mesh(f, obj)

def write_mesh(f, obj):
    print(indent + "- !std!objects/mesh", file=f)

    print(indent * 2 + "vertices:", file=f)
    for vert in obj.data.vertices:
        v = vert.co * obj.matrix_world
        print(indent * 3 + "- [%f, %f, %f]" % (v.x, v.y, v.z), file=f)

    print(indent * 2 + "faces:", file=f)
    for face in obj.data.faces:
        if len(face.vertices) == 3:
            # Triangle
            print(indent * 3 + "- vertices: [%d, %d, %d]" % (face.vertices[0], face.vertices[1], face.vertices[2]), file=f)
            print(indent * 3 + "  material: *mat%s" % (obj.data.materials[face.material_index].name), file=f)
        else:
            # Quad
            print(indent * 3 + "- vertices: [%d, %d, %d]" % (face.vertices[0], face.vertices[1], face.vertices[2]), file=f)
            print(indent * 3 + "  material: *mat%s" % (obj.data.materials[face.material_index].name), file=f)
            print(indent * 3 + "- vertices: [%d, %d, %d]" % (face.vertices[2], face.vertices[3], face.vertices[0]), file=f)
            print(indent * 3 + "  material: *mat%s" % (obj.data.materials[face.material_index].name), file=f)

def write_materials(f, scene):
    # Gather materials from objects in scene
    mats = set()
    for obj in scene.objects:
        if obj.type == 'MESH':
            mats |= set(obj.data.materials)
    # We don't technically have a materials section, but goray lets us get away with it.
    print("materials:", file=f)
    for material in mats:
        print(indent + "- &mat%s !std!materials/shinydiffuse" % (material.name), file=f)
        print(indent * 2 + "color: !goray!rgb [%f, %f, %f]" % (material.diffuse_color.r, material.diffuse_color.g, material.diffuse_color.b), file=f)
        print(indent * 2 + "diffuseReflect: %f" % (material.diffuse_intensity), file=f)
        print(indent * 2 + "mirrorColor: !goray!rgb [%f, %f, %f]" % (material.specular_color.r, material.specular_color.g, material.specular_color.b), file=f)
        print(indent * 2 + "specularReflect: %f" % (material.specular_intensity), file=f)
        print(indent * 2 + "transparency: %f" % (1 - material.alpha), file=f)
        print(indent * 2 + "translucency: %f" % (material.translucency), file=f)
        print(indent * 2 + "transmit: %f" % (material.raytrace_transparency.filter), file=f)

def write_lights(f, scene):
    # TODO: Handle no lights
    print("lights:", file=f)
    for obj in scene.objects:
        if obj.type != 'LAMP':
            continue
        if obj.data.type == 'POINT':
            write_point_light(f, obj)

def write_point_light(f, obj):
    print(indent + "- !std!lights/point", file=f)
    print(indent * 2 + "position: !goray!vec [%f, %f, %f]" % (obj.location.x, obj.location.y, obj.location.z), file=f)
    print(indent * 2 + "color: !goray!rgb [%f, %f, %f]" % (obj.data.color.r, obj.data.color.g, obj.data.color.b), file=f)
    print(indent * 2 + "intensity: %f" % (obj.data.energy), file=f)

def write_camera(f, scene):
    obj = scene.camera
    cam = obj.data
    if cam.type == 'ORTHO':
        print("camera: !std!cameras/ortho", file=f)
        print(indent + "scale: %f" % (cam.ortho_scale), file=f)
    elif cam.type == 'PERSP':
        print("camera: !std!cameras/perspective", file=f)
        f_aspect = 1.0
        fx = scene.render.resolution_x * scene.render.pixel_aspect_x
        fy = scene.render.resolution_y * scene.render.pixel_aspect_y
        if fx <= fy:
            f_aspect = fx / fy
        print(indent + "focalDistance: %f" % (cam.lens/(32 * f_aspect)), file=f)
    else:
        raise AssertionError("Unrecognized camera type")
    # Camera transform
    position = obj.matrix_world[3].xyz
    direction = obj.matrix_world[2].xyz
    up = position + obj.matrix_world[1].xyz
    look = position - direction
    print(indent + "position: !goray!vec [%f, %f, %f]" % (position.x, position.y, position.z), file=f)
    print(indent + "look: !goray!vec [%f, %f, %f]" % (look.x, look.y, look.z), file=f)
    print(indent + "up: !goray!vec [%f, %f, %f]" % (up.x, up.y, up.z), file=f)
    # Render parameters
    scale = scene.render.resolution_percentage / 100
    print(indent + "width: %d" % (scene.render.resolution_x * scale), file=f)
    print(indent + "height: %d" % (scene.render.resolution_y * scale), file=f)

def write_integrator(f):
    print("integrator: !std!integrators/directlight", file=f)
    print(indent + "transparentShadows: false", file=f)
    print(indent + "shadowDepth: 3", file=f)
    print(indent + "rayDepth: 10", file=f)

from bpy.props import *

class GorayExporter(bpy.types.Operator):
    """Export to the goray Scene Format (.yaml)"""

    bl_idname = 'export.goray'
    bl_label = "Export goray"

    filepath = StringProperty(subtype='FILE_PATH')

    def execute(self, context):
        write_file(self.filepath, context.scene)
        return {'FINISHED'}

    def invoke(self, context, event):
        context.window_manager.fileselect_add(self)
        return {'RUNNING_MODAL'}

def menu_func(self, context):
    default_path = os.path.splitext(bpy.data.filepath)[0] + os.path.extsep + 'yaml'
    op = self.layout.operator(GorayExporter.bl_idname, text="Goray (.yaml)")
    op.filepath = default_path

def register():
    bpy.types.INFO_MT_file_export.append(menu_func)

def unregister():
    bpy.types.INFO_MT_file_export.remove(menu_func)

if __name__ == '__main__':
    register()
