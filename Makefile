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

SOURCES = $(shell go list -f '{{join .GoFiles " "}}') assets.go

bugzini: $(SOURCES) $(DEPS_DIR)/.stamp
	GOPATH=$(DEPS_DIR) go build

$(DEPS_DIR)/.stamp: $(DEPS_IN_DIR) $(LOCAL_DEPS_IN_DIR)
	@touch $(DEPS_DIR)/.stamp

$(DEPS_IN_DIR):
	@dep=$(subst $(DEPS_DIR)/src/,,$@); \
	echo "[DEP] $$dep"; GOPATH=$(DEPS_DIR) go get $$dep

$(LOCAL_DEPS_IN_DIR):
	@dep=$(subst $(DEPS_DIR)/src/,,$@); \
	echo "[DEP] $$dep"; rm -f "$@"; ln -s "../../$(notdir $@)" "$@"

$(DEPS_DIR)/bin/go-assets-builder:
	GOPATH=$(DEPS_DIR) go install github.com/jessevdk/go-assets-builder

assets.go: $(DEPS_DIR)/bin/go-assets-builder $(ASSETS)
	$(DEPS_DIR)/bin/go-assets-builder -o $@ $(ASSETS)

.PHONY:
