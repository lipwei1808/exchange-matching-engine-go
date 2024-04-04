GO ?= go
GOFMT ?= gofmt "-s"
GOFILES := $(shell find . -name "*.go")

CC = clang
CXX = clang++

CFLAGS := $(CFLAGS) -g -O3 -Wall -Wextra -pedantic -Werror -std=c18 -pthread
CXXFLAGS := $(CXXFLAGS) -g -O3 -Wall -Wextra -pedantic -Werror -std=c++20 -pthread

BUILD_DIR := build

all: engine client

engine:
	$(GO) build -o $(BUILD_DIR)/$@ $(GOFILES)

client: build/client.cpp.o
	$(LINK.cc) $^ $(LOADLIBES) $(LDLIBS) -o build/$@

.PHONY: clean
clean:
	rm -f *.o client engine

.PHONY: fmt engine
fmt:
	$(GOFMT) -w $(GOFILES)

# dependency handling
# https://make.mad-scientist.net/papers/advanced-auto-dependency-generation/#tldr

DEPDIR := .deps
DEPFLAGS = -MT $@ -MMD -MP -MF $(DEPDIR)/$<.d

COMPILE.cpp = $(CXX) $(DEPFLAGS) $(CXXFLAGS) $(CPPFLAGS) $(TARGET_ARCH) -c

build/%.cpp.o: client/%.cpp
build/%.cpp.o: client/%.cpp $(DEPDIR)/%.cpp.d | $(DEPDIR)
	$(COMPILE.cpp) $(OUTPUT_OPTION) $<


$(BUILD_DIR): ; @mkdir -p $@
$(DEPDIR): ; @mkdir -p $@

DEPFILES := $(SRCS:%=$(DEPDIR)/%.d) $(DEPDIR)/client.cpp.d
$(DEPFILES):

include $(wildcard $(DEPFILES))
