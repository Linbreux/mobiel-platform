# Installation

## Requirements
If you're using Windows and having trouble building Pixel, please check this guide on the wiki.

PixelGL backend uses OpenGL to render graphics. Because of that, OpenGL development libraries are needed for compilation. The dependencies are same as for GLFW.

The OpenGL version used is OpenGL 3.3.

On macOS, you need Xcode or Command Line Tools for Xcode (xcode-select --install) for required headers and libraries.  
  
On Ubuntu/Debian-like Linux distributions, you need ```libgl1-mesa-dev``` and ```xorg-dev``` packages.  
```
sudo apt-get install libgl1-mesa-dev xorg-dev
```  

On CentOS/Fedora-like Linux distributions, you need libX11-devel libXcursor-devel libXrandr-devel libXinerama-devel mesa-libGL-devel libXi-devel libXxf86vm-devel packages.


See here for full details.
The combination of Go 1.8, macOS and latest XCode seems to be problematic as mentioned in issue #7. This issue is probably not related to Pixel. Upgrading to Go 1.8.1 fixes the issue.
