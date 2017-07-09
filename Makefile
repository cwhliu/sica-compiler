
export CC := clang
export CXX := clang++

pwd := $(shell pwd)

install:
	CGO_CFLAGS="-I`llvm-config --includedir`" CGO_LDFLAGS="-L$(pwd)/../go-clang/lib -L`llvm-config --libdir` -L/usr/lib/x86_64-linux-gnu -lclangExt -lclangTooling -lclangDriver -lclangFrontend -lclangParse -lclangSema -lclangAnalysis -lclangEdit -lclangAST -lclangSerialization -lclangLex -lclangBasic -lclang -lLLVM -lstdc++" go install ./...
install-dep:
	go get -d -u github.com/cwhliu/go-clang-v3.9/...

