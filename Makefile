DEPS_DIR = $(CURDIR)/.deps

DEPS =						\
	github.com/jessevdk/xmlrpc		\
	github.com/gorilla/mux			\
	github.com/jessevdk/go-assets

LOCAL_DEPS = 					\
	bugzilla

ASSETS =					\
	$(shell find assets/ -type f)

DEPS_IN_DIR = $(addprefix $(DEPS_DIR)/src/,$(DEPS))
LOCAL_DEPS_IN_DIR = $(addprefix $(DEPS_DIR)/src/,$(LOCAL_DEPS))

IKNOWMYGO :=

ifeq ($(IKNOWMYGO),)
check_go = $(shell test $$(go version | sed 's/.*go1\.\([0-9]*\).*/\1/g') -ge 1 && echo 1)

ifeq ($(check_go),)
$(error Could not find sufficiently recent installation of go. Please install go 1.1 or later first. If this check is borked, and you want to proceed anyway, please set IKNOWMYGO)
endif
endif

SOURCES = $(shell go list -f '{{join .GoFiles " "}}') assets.go

bugzini: $(SOURCES) $(DEPS_DIR)/.stamp
	@echo  "[build] $@"; \
	GOPATH=$(DEPS_DIR) go build

$(DEPS_DIR)/.stamp: $(DEPS_IN_DIR) $(LOCAL_DEPS_IN_DIR)
	@touch $(DEPS_DIR)/.stamp

$(DEPS_IN_DIR):
	@dep=$(subst $(DEPS_DIR)/src/,,$@); \
	echo "[DEP] $$dep"; GOPATH=$(DEPS_DIR) go get -d $$dep

$(LOCAL_DEPS_IN_DIR):
	@dep=$(subst $(DEPS_DIR)/src/,,$@); \
	echo "[DEP] $$dep"; rm -f "$@"; ln -s "../../$(notdir $@)" "$@"

$(DEPS_DIR)/bin/go-assets-builder:
	@echo "[GEN] go-assets-builder"; \
	GOPATH=$(DEPS_DIR) go get github.com/jessevdk/go-assets-builder

assets.go: $(DEPS_DIR)/bin/go-assets-builder $(ASSETS)
	@echo "[GEN] $@"; \
	$(DEPS_DIR)/bin/go-assets-builder -o $@ $(ASSETS)

.PHONY:
