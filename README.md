# Arch Linux Installation

**manul** can be installed on Arch Linux with the following:

```bash
git clone --branch pkg-archlinux git://github.com/kovetskiy/manul /tmp/manul
cd /tmp/manul
makepkg
pacman -U *.xz
```
