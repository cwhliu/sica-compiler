# SICA Compiler - Forge

## How to run the SICA compiler
### Install Go
```
apt-get install golang
```
Then set up environment variables for Go. Here I'm using ~/go as my _GOPATH_, you can use whatever directory you prefer. Go will put all files under this directory.
```
export GOPATH=~/go
export PATH=$PATH:$GOPATH/bin
```
### Install Clang && LLVM
```
apt-get install llvm-4.0 clang-4.0 libclang-4.0-dev
```
Make sure you can run clang and llvm-config without specifying their version (i.e. `llvm-config` instead of `llvm-config-4.0`), create links if necessary.
```
ln -s /usr/bin/llvm-config-4.0 /usr/bin/llvm-config
ln -s /usr/bin/clang-4.0 /usr/bin/clang
```
### Get source code
```
go get -d github.com/cwhliu/sica-compiler
```
This will get the source code and put them in _$GOPATH/src/github.com/cwhliu/sica-compiler_. If you decide to use `git clone` please clone into the same directory structure so that Go can find the files.
### Compile source code
```
cd $GOPATH/src/github.com/cwhliu/sica-compiler
make
```
An executable `sica-compiler` will be created at _$GOPATH/bin_.
### Prepare test data
```
cd $GOPATH/src/github.com/cwhliu/sica-compiler/testdata
tar zxvf testdata_atlas_default.tgz
```
### Run the compiler
```
sica-compiler testdata/atlas/default/torque_LeftStance_interior.cc
```

## Documentation
See [GoDoc](https://godoc.org/github.com/cwhliu/sica-compiler/forge) for detailed documentation
