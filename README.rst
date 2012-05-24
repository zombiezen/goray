*********
  goray
*********

goray is very much beta quality software and not ready for production.  Many
features are missing, some things may not work right, and the behavior of the
program is not well documented.

That said, you must have `Go`_ 1.0.1 installed to build goray.

Download and install goray with::

    go get -u bitbucket.org/zombiezen/goray/goray

To run goray, provide a scene description file (a YAML file) as the sole
argument.  By default, goray will render to a file called ``goray.png`` in the
current directory.  For example::

    goray demos/suzanne.yaml

For further help::

    goray -help

.. _Go: http://golang.org/

License
=========

| Copyright © 2011 Ross Light
| Based on YafaRay: Copyright © 2005 Mathias Wein, Alejandro Conty, and Alfredo de Greef
| jQuery: Copyright © 2011 John Resig

goray comes with ABSOLUTELY NO WARRANTY.  goray is free software, and you are
welcome to redistribute it under the conditions of the GNU Lesser General
Public License v3, or (at your option) any later version.

jQuery License
----------------

Copyright © 2011 John Resig, http://jquery.com/

Permission is hereby granted, free of charge, to any person obtaining
a copy of this software and associated documentation files (the
"Software"), to deal in the Software without restriction, including
without limitation the rights to use, copy, modify, merge, publish,
distribute, sublicense, and/or sell copies of the Software, and to
permit persons to whom the Software is furnished to do so, subject to
the following conditions:

The above copyright notice and this permission notice shall be
included in all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND,
EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF
MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND
NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE
LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION
OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION
WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
