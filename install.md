Go toolchain installation and configuration
===========================================


Flux
----

* Connect to Flux (more information [here](http://arc-ts.umich.edu/flux-user-guide))

* Type `module load go`

* Use your editor to add the following line to your .bashrc file:

export GOPATH="${HOME}/go_projects"

* Type `source ~/.bashrc` or run `bash` to reload the bash
  configuration

* Type `go get golang.org/x/tools/cmd/goimports`

* Emacs configuration (optional):

    * Create a directory, e.g. `\home\uniqname\emacs`

    * Change to the directory that you just created and run:

        * wget https://raw.githubusercontent.com/dominikh/go-mode.el/master/go-rename.el

	* wget https://raw.githubusercontent.com/dominikh/go-mode.el/master/go-guru.el

        * wget https://raw.githubusercontent.com/dominikh/go-mode.el/master/go-mode-autoloads.el

        * wget https://raw.githubusercontent.com/dominikh/go-mode.el/master/go-mode.el

    * Open your Emacs configuration file, `emacs .emacs` and enter this at the bottom:

```
(add-to-list 'load-path "/home/uniqname/emacs")
(require 'go-mode-autoloads)
```