apt-get install -y gcc-7 g++-7
apt-get install -y python3-pip
apt-get install -y pkg-config
python3 -m pip install pyelftools
python3 -m pip install meson ninja
wget http://fast.dpdk.org/rel/dpdk-21.11.4.tar.xz
rm -r dpdk-stable-21.11.4
tar -xvf dpdk-21.11.4.tar.xz
cd dpdk-stable-21.11.4
meson build
DESTDIR=/ ninja -C build install
ls /usr/local/lib/x86_64-linux-gnu/pkgconfig
export PKG_CONFIG_PATH=$PKG_CONFIG_PATH:/usr/local/lib/x86_64-linux-gnu/pkgconfig
ldconfig
python3 ./usertools/dpdk-hugepages.py -p 2M --setup 2G
python3 ./usertools/dpdk-hugepages.py -s
