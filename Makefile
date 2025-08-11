GO            ?= go
EXPORT_DIR    := ./loggo/export
OUT_DIR       := ./lightloggo/ffi
LIB  := lightloggo

# --- Tools ---
MINGW_CC    := x86_64-w64-mingw32-gcc
O64_CLANG   := $(OSXROOT)/bin/o64-clang
OA64_CLANG  := $(OSXROOT)/bin/oa64-clang
O_CLANGXX   := $(OSXROOT)/bin/o64-clang++
OA_CLANGXX  := $(OSXROOT)/bin/oa64-clang++

ifeq ($(GOOS),)
  UNAME_S := $(shell uname -s)
  ifeq ($(UNAME_S),Darwin)
    EXT := dylib
  else ifeq ($(UNAME_S),Linux)
    EXT := so
  else
    EXT := dll
  endif
else
  ifeq ($(GOOS),darwin)
    EXT := dylib
  else ifeq ($(GOOS),linux)
    EXT := so
  else ifeq ($(GOOS),windows)
    EXT := dll
  else
    $(error Unsupported GOOS '$(GOOS)')
  endif
endif

OUT_FILE := $(OUT_DIR)/$(LIB).$(EXT)
HDR_FILE := $(OUT_DIR)/$(LIB).h

.PHONY: deps deps-linux deps-windows
deps: deps-linux deps-windows

deps-linux:
	sudo dnf install -y git golang gcc make cmake llvm lld clang \
		tar xz cpio patch file \
		zlib-devel bzip2 bzip2-devel xz-devel \
		readline-devel sqlite sqlite-devel \
		openssl-devel tk-devel libffi-devel gdbm-devel libuuid-devel

deps-windows:
	sudo dnf install -y mingw64-gcc

.PHONY: build build-linux build-windows clean env

env:
	@echo "== go env =="
	$(GO) env GOOS GOARCH CGO_ENABLED CC CXX

build build-linux:
	@mkdir -p $(OUT_DIR)
	@(cd $(EXPORT_DIR) && \
	  GO111MODULE=on CGO_ENABLED=1 GOOS=linux GOARCH=amd64 \
	  $(GO) build -v -buildmode=c-shared \
	    -o ../../$(OUT_DIR)/$(LIB).so .)
	@echo ">> Built: $(OUT_DIR)/$(LIB).so"

build-windows:
	@mkdir -p $(OUT_DIR)
	@(cd $(EXPORT_DIR) && \
	  GO111MODULE=on CGO_ENABLED=1 GOOS=windows GOARCH=amd64 CC=x86_64-w64-mingw32-gcc \
	  $(GO) build -v -buildmode=c-shared \
	    -o ../../$(OUT_DIR)/$(LIB).dll .)
	@echo ">> Built: $(OUT_DIR)/$(LIB).dll"

.PHONY: package verify
package: verify
	@mkdir -p $(DIST_DIR)
	@# архивы по платформам
	@[ -f "$(OUT_DIR)/$(LIB).so" ]     && tar -C $(OUT_DIR) -cvf $(DIST_DIR)/$(LIB)-linux-amd64.tar $(LIB).so $(LIB).h     || true
	@[ -f "$(OUT_DIR)/$(LIB).dll" ]    && tar -C $(OUT_DIR) -cvf $(DIST_DIR)/$(LIB)-windows-amd64.tar $(LIB).dll $(LIB).h  || true
	@[ -f "$(OUT_DIR)/$(LIB).dylib" ]  && tar -C $(OUT_DIR) -cvf $(DIST_DIR)/$(LIB)-darwin.tar $(LIB).dylib $(LIB).h      || true
	@echo ">> Artifacts in $(DIST_DIR):"; ls -lh $(DIST_DIR) || true


verify:
	@echo "== file outputs =="
	@[ -f "$(OUT_DIR)/$(LIB).so" ]    && file $(OUT_DIR)/$(LIB).so    || true
	@[ -f "$(OUT_DIR)/$(LIB).dll" ]   && file $(OUT_DIR)/$(LIB).dll   || true
	@[ -f "$(OUT_DIR)/$(LIB).h" ]     && echo "Header: $(OUT_DIR)/$(LIB).h" || true


.PHONY: clean distclean
clean:
	rm -f $(OUT_DIR)/$(LIB).so $(OUT_DIR)/$(LIB).dll $(OUT_DIR)/$(LIB).dylib $(OUT_DIR)/$(LIB).h

distclean: clean
	rm -rf $(DIST_DIR)
