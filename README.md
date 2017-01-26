# goray

goray is very much beta quality software and not ready for production.  Many
features are missing, some things may not work right, and the behavior of the
program is not well documented.

Download and install goray with:

    go get -u zombiezen.com/go/goray

To run goray, provide a scene description file (a YAML file) as the sole
argument.  By default, goray will render to a file called ``goray.png`` in the
current directory.  For example:

    goray demos/suzanne.yaml

For further help:

    goray -help

## License

Copyright © 2011 Ross Light
Based on YafaRay: Copyright © 2005 Mathias Wein, Alejandro Conty, and Alfredo de Greef

goray comes with ABSOLUTELY NO WARRANTY.  goray is free software, and you are
welcome to redistribute it under the conditions of the GNU Lesser General
Public License v3, or (at your option) any later version.
