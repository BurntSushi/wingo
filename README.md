Wingo is an X window manager written in pure Go. All of its dependencies, from 
communicating with X up to drawing text on windows, are also in Go. Wingo is 
mostly ICCCM and EWMH compliant (see COMPLIANCE).

If you have Go installed and configured on your machine, all you need to do is:
(For Archlinux users, Wingo is in the AUR.)

    go get github.com/BurntSushi/wingo

And in your $HOME/.xinitrc:

    exec wingo

Or if you're brave and are using a desktop environment, just run this to 
replace your current window manager: (Seriously though, save you work. Wingo is 
still very alpha.)

    wingo --replace


Help
====
You can find me in the #wingo IRC channel on FreeNode.


My triple head setup
====================

[![Triple head with Wingo](https://github.com/BurntSushi/wingo/wiki/screenshots/thumbs/triple-head.png)](https://github.com/BurntSushi/wingo/wiki/screenshots/triple-head.png)


Dude... why?
============
Wingo has two features which, when combined, set it apart from other window 
managers (maybe):

1) Support for both floating *and* tiling placement policies. Wingo can be used 
   as a regular floating (stacking) window manager, complete with decorations,
   maximization, sticky windows, and most other things you might find in a
   window manager. Wingo can also be switched into a tiling mode where window
   decorations disappear, and windows are automatically managed via tiling.

2) Workspaces per monitor. In a multi-head setup, each monitor can display its 
   own workspace---independent of the other monitors. This makes changing your
   view of windows in multi-head setups much easier than in most other window
   managers, which will only allow you to see one workspace stretched across 
   all of your monitors. Also, since placement policies like floating and 
   tiling affect workspaces, this makes it possible for one monitor to be 
   tiling while another is floating!

WARNING: The major drawback of using a workspaces per monitor model is that it 
violates an implicit assumption made by EWMH: that one and only one workspace 
may be viewable at any point in time. As a result, in multi-head setups, pagers 
and taskbars may operate in confusing ways. In a single head setup, they should 
continue to operate normally. Wingo provides prompts that allow you to 
add/remove workspaces and select clients that may alleviate the need for pagers 
or taskbars.


Configuration
=============
Wingo is extremely configurable. This includes binding any of a vast number of 
commands to key or mouse presses, theming your window decorations and setting 
up hooks that will fire depending upon a set of match conditions.

All configuration is done using an INI like file format with support for simple
variable substitution (which makes theming a bit simpler).
No XML. No recompiling. No scripting.

A fresh set of configuration files can be added to `$HOME/.config/wingo` with

    wingo --write-config

Each configuration file is heavily documented.

Configuring key/mouse bindings and hooks uses a central command system called 
Gribble. For example, one can add a workspace named "cromulent" with this
command:

    AddWorkspace "cromulent"

But that's not very flexible, right? It'd be nice if you could specify the name
of workspace on the fly... For this, simply use the "Input" command as an
argument to AddWorkspace, which shows a graphical prompt and allows you to type 
in a name:

    AddWorkspace (Input "Enter your workspace name:")

The text entered into the input box will be passed to the AddWorkspace command.

Please see the HOWTO-COMMANDS file for more info. We've barely scratched the 
surface.


Scripting Wingo
===============
So I lied earlier. You can kind of script Wingo by using its IPC mechanism. 
You'll need to make sure that wingo-cmd is installed:

    go get github.com/BurntSushi/wingo/wingo-cmd

While Wingo is running, you can send any command you like:

    wingo-cmd 'AddWorkspace "embiggen"'

Or perhaps you can't remember how to use the AddWorkspace command:

    wingo-cmd --usage AddWorkspace

Which will print the parameters, their types and a description of the command. 

Want to pipe some information to another program? No problem, since commands
can return stuff!

    wingo-cmd GetWorkspace

And you can even make commands repeat themselves every X milliseconds, which is 
ideal for use with something like dzen to show the name of the currently active 
window:

    wingo-cmd --poll 500 'GetClientName (GetActive)' | dzen2

Finally, you can see a list of all commands, their parameters and their usage:
(even if Wingo isn't running)

    wingo-cmd --list-usage

(Wingo actually can provide enough information for ambitious hackers to script 
their own layouts in whatever programming language they like without ever
having to deal with X at all. Assuming it has support for connecting to unix
domain sockets. Or you could just use a shell with 'wingo-cmd' if you're into 
that kind of tomfoolery.)

Workspaces
==========
Having some set number of workspaces labeled 1, 2, 3, 4, ... is a thing of the 
past. While Wingo won't stop you from using such a simplistic model, it will 
egg you on to try something else: dynamic workspaces.

Dynamic workspaces takes advantage of two things: workspace names and 
adding/removing workspaces as you need them.

This is something that I find useful since I'm typically working on multiple 
projects, and my needs change as I work on them. For example, when working on 
Wingo, I might add the "wingo" workspace, along with the "xephyr" workspace and 
the "gribble" workspace. When I'm done, I can remove those and add other 
workspaces for my next project. Or I can leave those workspaces intact for 
when I come back to them later.

With Wingo, such a workflow is natural because you're no longer confined to 
"removing only the last workspace" or some other such nonsense. Plus, adding a 
workspace *requires* that you name it---so workspaces always carry some 
semantic meaning.

(N.B. I don't mean to imply that this model is new, just uncommon; particularly
      among floating window managers. I've personally taken the model from
      xmonad-contrib's DynamicWorkspaces module.)


Tiling layouts
==============
Right now, only simple tiling layouts are available. (Vertical and Horizontal.)
Mostly because those are the layouts that I primarily use. I'll be adding more 
as they are demanded.


Ummm... manual tiling?
======================
I'd actually love to add this to Wingo. It's slightly more complex than 
automatic tiling layouts, because it introduces the concept of containers, 
which is something that Wingo knows nothing about. Namely, a container can hold 
zero or more windows and an empty container may have focus.


Why doesn't Wingo have..?
=========================
Tags
----
Another popular workspace model (particularly among tiling window managers) is 
tagging a window with one or more workspaces.

Not only do I find this needlessly complex, but it doesn't really make sense in 
a model where more than one workspace can be visible in multi-head setups.

Shaded windows
--------------
This is in Openbox, but not Wingo. Honestly, I just never use it. I'm not 
really opposed to them, though.

Tabbed windows
--------------
The thought of programming the decorations for this scares me. This, like 
manual tiling, would also require that Wingo have a notion of containers (which 
it doesn't).

Compositing
-----------
Bandwidth allotment exceeded. Seriously.

If an ambitious person wanted to run with it, that's fine, but there are 
serious hurdles. The most pertinent one is mixing OpenGL with the pure X Go 
Binding. I am not sure how to do it.

One could use the X RENDER extension, but I think everyone hates that.

Wayland
-------
I have done a non-trivial amount of research into Wayland (but not a big 
amount) and there are serious hurdles to overcome before Go can work with the 
Wayland protocol in a practical way. Namely, while a pure Go binding could be 
written easily enough, it would be forced into software compositing---which 
could be too slow. In order to do hardware compositing, I think you need OpenGL 
(specifically, EGL), which links against the libwayland libraries. (Yeah, 
that's a recursive dependency. Wooho.)

Plus, in order to use Wayland, Wingo would need a compositing backend (along 
with every other non-compositing X11 window manager). This is also not an easy 
task.

Supposedly there are some ideas for plans floating around that would let
non-compositing X window managers to "plug into" the Wayland reference 
compositor (Weston). When this will be possible (or even *if* it will be 
possible with a window manager written in Go) remains to be seen.

If I am in err (and this is quite likely; my OpenGL knowledge is limited), 
please ping me.


Dependencies
============
You really should be using the 'go' tool to install Wingo, and therefore 
shouldn't care about dependencies. But I'll list them anyway---with many thanks 
to the authors (well, the ones that aren't me anyway).

* go                  http://golang.org
* graphics-go         http://code.google.com/p/graphics-go
* freetype-go         http://code.google.com/p/freetype-go
* ansi                http://github.com/str1ngs/ansi
* go-bindata          http://github.com/jteeuwen/go-bindata (build dependency)
* gribble             http://github.com/BurntSushi/gribble
* xgb                 http://github.com/BurntSushi/xgb
* xgbutil             http://github.com/BurntSushi/xgbutil
* xdg                 http://github.com/BurntSushi/xdg


Inspiration
===========
Wingo is *heavily* inspired by Openbox and Xmonad. Basically, Openbox's 
floating semantics (although most window managers are the same in this regard) 
and Xmonad's automatic tiling style plus its workspace model (workspaces per 
monitor). I've also adopted Xmonad's "greedy" workspace switching and embedded 
the concepts from the "DynamicWorkspaces" contrib library into the Gribble 
command system.


Go your own way
===============
Wingo is actually split up into *many* sub-packages. It is possible (but not 
necessarily likely) that you could pick out some of these sub-packages and use 
them in your own window manager. The packages of particular interest are 
probably the ones that do the most nitty gritty X stuff---especially relating 
to drawing windows. Here's a quick run down of those:

cursors
-------
Sets up some plain old X cursors. Not very interesting.

prompt
------
Provides several different kinds of prompt windows that can take user input. 
These should actually work in an existing window manager. (See the examples in 
the package directory.) Prompt requires both the 'render' and 'text' Wingo 
packages.

render
------
Renders some very basic shapes and gradients to X windows.

text
----
Renders text to windows. Also provides a special window type that can act as a 
text box for user input.

Others
------
The only other package worth mentioning is 'frame'. It's probably too 
monolithic to be used in another window manager (unless you really like Wingo's 
decorations), but it's possible that it could serve as a half-decent template 
for your own frames.

The rest of the sub-packages (excluding xclient and wm, since they are very 
Wingo specific) could also be used, particularly since only minimal Client 
interfaces are required. However, most of them aren't that complex and 
therefore probably aren't worth it. And the ones that do have some complexity 
(maybe 'heads' and 'workspace') aren't packages that I'm particularly proud of.

Also, if you're wanting to make a Go window manager, my xgbutil package 
(separate from Wingo) will be a big help. Feel free to ping me.


My past X work
==============
There's too much. The highlights are pytyle and Openbox Multihead.

For more: http://burntsushi.net/x11/

