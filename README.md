### Setup
#### Kata framework
##### Requirements

- OS: centos 7
- Kata
   - Go 1.20.6, GLIBC_2.28
- Containerd
   - Go 1.12.17, GLIBC_2.28
- Others
   - python3, flask 2.3, bcc
##### Install basic framework 
```
tar xf cri-containerd-cni-1.6.13-linux-amd64.tar.gz -C /
tar xf kata-static-3.1.3-x86_64.tar.xz -C /
mv /opt/kata/ /opt/kata3/
ln -sf /opt/kata3/bin/containerd-shim-kata-v2 /usr/local/bin/containerd-shim-kata3-v2
```
##### Install configuration files
```
cd cnicmp
cp -r ./configs/containerd/ /etc/
cp -r ./configs/kata-containers /etc/
cp -r ./configs/net.d/ /etc/cni/
cp ./configs/crictl.yaml /etc/crictl.yaml
```
##### Apply changes
###### Containerd

- Download containerd source code and apply patch
```
wget https://codeload.github.com/containerd/containerd/zip/refs/heads/release/1.3
unzip containerd-release-1.3.zip
cd containerd-release-1.3
patch -p1 < cnicmp/patches/0001-containerd.patch
```

- Complie and install
```
export GO111MODULE=off
make 
cp ./bin/* /usr/local/bin/
```

- Or compile using docker
   - [cnicmp-external-link](#dataset) is the external dataset link
```
cd [cnicmp-external-link]/images/
docker load -i build2.zip
bash cnicmp/patches/build_scripts/build_containerd_centos8.sh 
```
###### Kata runtime

- Download kata source code and apply patch
```
wget https://codeload.github.com/kata-containers/kata-containers/zip/refs/heads/stable-3.0
unzip kata-containers-stable-3.0.zip
cd kata-containers-stable-3.0
patch -p1 < cnicmp/patches/0001-kata.patch
```

- Complie and install
```
make
cp ./containerd-shim-kata-v2 /opt/kata3/bin/
cp ./kata-runtime /opt/kata3/bin/
```

- Compile using docker
```
cd [cnicmp-external-link]/images/
docker load -i build-kata2.zip
bash cnicmp/patches/build_scripts/build_kata_centos8.sh
```
#### CNIs
##### ovs-veth: 

- go 1.20.6
```
cd ./ovs-veth
make build-plugin
cp ./ovs-cni/cmd/plugin/plugin /opt/cni/bin/ovs
chmod 777 /opt/cni/bin/ovs
```

- Compile using docker
```
cd cnicmp-external-link/images/
docker load -i build-ovs-cni.zip
bash ovs-veth/build_ovs_cni.sh
```
##### ovs-vhost

- go 1.12.17 (actually ovs-veth and vhost do not have strict requirements on the go version )
```
cd ./ovs-vhost
make build-plugin
cp ./ovs-vhost/cmd/plugin/plugin /opt/cni/bin/ovs-vhost
chmod 777 /opt/cni/bin/ovs-vhost
```

- Compile using docker
```
cd cnicmp-external-link/images/
docker load -i build2.zip
bash ovs-vhost/build_vhost_cni.sh
```
##### others: mac/ipvlan, mac/ipvtap

- go 1.21.0
```
cd ./cni-plugins
bash build_linux.sh
```
#### DPDK
```
tar xvf dpdk.tar.xz
DESTDIR=/ ninja -C build install
ls /usr/local/lib/x86_64-linux-gnu/pkgconfig (/usr/local/lib64/pkgconfig for centos)
export PKG_CONFIG_PATH=$PKG_CONFIG_PATH:/usr/local/lib/x86_64-linux-gnu/pkgconfig
(export PKG_CONFIG_PATH=$PKG_CONFIG_PATH:/usr/local/lib64/pkgconfig)
ldconfig
```
#### OVS
```
tar xvf openvswitch-3.1.2.tar.gz
./configure --with-dpdk=static
make -j
make install
bash ./all_scripts/setup/ovs_setup.sh  #start ovs
```
### Quick Start
#### DPDK&OVS start
```
python3 ./dpdk/usertools/dpdk-hugepages.py -p 2M --setup 4G
bash ./all_scripts/setup/ovs_setup.sh
ovs-vsctl show # check ovs works
```
#### Basic procedure

- Script path: _cnicmp/all_scripts/cnicmp/scripts/_
- Configure concurrency
```
vim ./time_test.conf
```

- Enable/Disable network
```
cd ../
bash toml_net.sh
bash toml.sh
```

- **Start lock-server (necessary step)**
```
cd ./scripts/ovs_test_scripts/
python3 lock_opt_server.py 400 400 400 400 400 400 400
```

- Manual test (need to disable rpc in the CNI config file，and launch http-server manually)
```
bash ./scripts/time_kata_test.sh (no network)
bash ./scripts/time_kata_test_net.sh (with network)
bash ./scripts/time_kata_test_net_poisson.sh (with network, Poisson invocation arrival)
```
### Paper experiment
#### Motivation

- no-net
```
cd cnicmp/all_scripts/cnicmp
bash toml.sh
cd ./scripts/ovs_test_scripts
python3 tenant-test-nonet.py ovs-ipvlan 400 1 0 0 kata
```
*tenant-test script
```
python3 tenant-test-nonet.py [cni name] [concurrency] [number of tenants] [number of security groups rule per pod] [0/1 if enable QoS] [security container type]
```

- kata
```
cd cnicmp/all_scripts/cnicmp/scripts/
bash toml_net.sh
cd ./ovs_test_scripts
python3 tenant-test.py ovs-ipvlan 400 1 0 0 kata
```

- rund
```
cd cnicmp/all_scripts/cnicmp/scripts/
bash toml_net.sh
cd ./ovs_test_scripts
python3 tenant-test.py ovs-ipvlan 400 1 0 0 rund
```
#### Overall concurrency

- use tenant-test script (cnicmp/all_scripts/cnicmp/scripts/ovs_test_scripts/)
```
python3 tenant-test.py [ovs/ovs-vhost/ovs-ipvlan/ovs-ipvtap/ovs-tc-routing ...] 400 1 0 0 kata
```
#### QoS

- use tenant-test script
```
python3 tenant-test.py [ovs/ovs-vhost/ovs-ipvlan/ovs-ipvtap/ovs-tc-routing ...] 400 1 0 1 kata
```
#### Security group rules

- use tenant-test script
```
python3 tenant-test.py [ovs/ovs-vhost/ovs-ipvlan/ovs-ipvtap/ovs-tc-routing ...] 400 1 [10/50/100] 0 kata
```
#### Multi-tenant

- tenant-test script
```
python3 tenant-test.py [ovs-ipvlan/ovs-ipvtap/ovs-tc-routing ...] 400 [1/50/100/200] [10/50/100] [0/1] kata
```
#### Data plane performance
##### Intra-node 

- script path：cnicmp/all_scripts/single_pod/dp-test.sh
- log path：cnicmp/all_scripts/single_pod/dataplane_logs/
```
bash dp-run.sh vhost(marker) 
```
##### Inter-node

- script path：cnicmp/all_scripts/single_pod/dp-test-inter.sh
- NIC configure
```
ovs-vsctl add-port br0 dpdk2 -- set Interface dpdk2 type=dpdk options:dpdk-devargs=0000:6b:00.0 # DPDK port BDF
ovs-vsctl add-port br0 ens65f0 [] # output NIC port
```

- Configure IP on two servers, and enable vswitch bridge arp broadcast
```
ovs-ofctl add-flow br0 priority=100,dl_type=0x0806,actions=FLOOD 
```

- Run the following command on two servers to start the test container
```
bash net_test_cni.sh
```

- Run the following command on one server
```
bash dp-test-inter.sh vhost(marker) 192.168.0.2(peer server ip) 10.0.0.1 (ip of the pod used as iperf3 server)
```
#### Pooling

- Containerd and kata recompile + reinstall
   - Search and turn on the _**DoUseCNIPool**_ flag in kata & containerd source code 
```
Kata:
  src/runtime/cnicmp/cnipool/pool.go
Containerd:
  cnicmp/cnipool/pool.go
```

   - Or simply reinstall containerd and kata using the provided binaries
```
/usr/local/bin/containerd.backup.pool
/opt/kata3/bin/containerd-shim-kata-v2.backup.pool
```

- Run tests using tenant-test-pool.py script
```
cd cnicmp/all_scripts/cnicmp/scripts/ovs_test_scripts/
python3 tenant-test-pool.py ovs-dev-pool ...
```
#### Bayes concurrency control
```
cd cnicmp/all_scripts/cnicmp/scripts/ovs_test_scripts/
nohup python3 lock_opt_run.py > loc-opt.loc 2>&1 & 
cat loc-opt.loc | grep cost #check results
```
#### Pooling + CC

- Containerd and kata recompile + reinstall as explained in **Pooling**
```
cd cnicmp/all_scripts/cnicmp/scripts/ovs_test_scripts/
nohup python3 lock_opt_run_pool.py > loc-opt-pool.loc 2>&1 & 
cat loc-opt-pool.loc | grep cost #check results
```
#### eBPF

- Script path：cnicmp/all_scripts/ebpf
- Rtnl lock time cost
```
cd cnicmp/all_scripts/ebpf
python3 rtnl_lock_timeline.py > log-rtnl-time
[then run the concurrent container launch]
python3 plot_specific.py ./log-rtnl-time   
```

- Number of rtnl calls
```
cd cnicmp/all_scripts/ebpf
python3 rtnl_lock_timeline_caller.py > log-rtnl-caller
python3 stat_caller.py ./log-rtnl-caller
```

- Spin lock sampling
```
python3 spin_lock_timeline.py > log-spin
python3 stat_spin.py ./log-spin
```
### Paper Plot
Script path: _cnicmp/plot/analysis/imc_plot_
#### Motivation

- Time cost
```
python3 motivation_network_overhead_camera.py
```

- Relative increase
```
python3 motivation_network_overhead_bar_camera.py
```
#### Overall performance

- Bar
```
python3 concurrency_overall_fixed_camera.py
```

- Breakdown
```
python3 gen_timeline.py # this command generates all the timeline breakdown figures in the paper
```

- Poisson invocation arrival breakdown
```
python3 pooling_results_breakdown_fixed_poisson.py 3
```
#### QoS and secury group rules

- QoS
```
python3 qos_study_inc_fixed.py 
```

- Security group rules
```
python3 acl_study_fixed.py
```
#### Number of tenants

- No QoS
```
python3 tenant_level_study_fixed.py
```

- With QoS
```
python3 tenant_level_study_qos_fixed.py 
```
#### Data plane performance

- Data plane
```
python3 throughput_overall_intra.py
python3 throughput_overall_inter.py
python3 latency_overall_intra.py
python3 latency_overall_inter.py
```

- Resource consumption
```
python3 resource_cpu_intra.py 
python3 resource_cpu_inter.py 
python3 resource_cpu_host.py 
python3 resource_cpu_host_bar.py 
```
#### Pipeline breakdown
```
python3 pipeline_acl_fixed2.py
```
#### Pooling

- No network breakdown
- Pooling breakdown

(* already generated by _python3 gen_timeline.py_ in overall performance)
#### Concurrency control

- Timeline breakdown (* already generated by _python3 gen_timeline.py_ in overall performance)
- CPU consumption
```
python3 cpu_consumption_opt_peak.py
python3 cpu_consumption_opt_total.py
```

- Generalization
```
python3 pipeline_optimize_extend_bar_distribution.py
```

- Optimization combination
```
python3 pipeline_optimize_combined_bar.py 
```
### Dataset, binaries and docker images<a id="dataset"></a>

- [cnicmp-external-link](https://drive.google.com/drive/folders/1xbJuY4iCGyyVqeaw20OhstbKl4NKzlY0?usp=sharing)
#### 

