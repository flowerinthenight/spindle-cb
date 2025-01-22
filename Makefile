BASE = $(CURDIR)
MODULE = spindle-cb

.PHONY: all $(MODULE) install
all: $(MODULE)

$(MODULE):| $(BASE)
	@GO111MODULE=on go build -v -trimpath -o $(BASE)/bin/$@

$(BASE):
	@mkdir -p $(dir $@)

install:
	@GO111MODULE=on go install -v
