
CGO_CFLAGS  := -I$(shell llvm-config --includedir)

CGO_LDFLAGS := -L$(shell pwd)/../go-clang/lib
CGO_LDFLAGS += -L$(shell llvm-config --libdir)
CGO_LDFLAGS += -L/usr/lib/x86_64-linux-gnu
CGO_LDFLAGS += -lclangExt -lclangTooling -lclangDriver -lclangFrontend -lclangParse
CGO_LDFLAGS += -lclangSema -lclangAnalysis -lclangEdit -lclangAST -lclangSerialization
CGO_LDFLAGS += -lclangLex -lclangBasic -lclang -lLLVM -lstdc++

export CC  := clang
export CXX := clang++

export CGO_CFLAGS
export CGO_LDFLAGS

install:
	go install ./...
install-dep:
	go get -d -u github.com/cwhliu/go-clang-v3.9/...

