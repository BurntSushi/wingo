BD=bindata
D=data
WINGOPKG=$(HOME)/go/me/pkg/linux_amd64/github.com/BurntSushi/wingo
BINDATA=$(BD)/wingo.png.go \
				$(BD)/close.png.go $(BD)/maximize.png.go $(BD)/minimize.png.go \
				$(BD)/DejaVuSans.ttf.go

bindata: $(BINDATA)

sushi-bindata: $(WINGOPKG)/bindata.a

gofmt:
	gofmt -w *.go wini/*.go
	colcheck *.go */*.go

cmd:
	go install github.com/BurntSushi/wingo/wingo-cmd

$(WINGOPKG)/bindata.a: $(BINDATA)
	(cd $(BD) ; go install)

$(BD)/%.png.go: $(D)/%.png
	go-bindata -f `python2 -c 'print "$*".title()'`Png \
		-i $(D)/$*.png -o $(BD)/$*.png.go -p bindata
	gofmt -w $(BD)/$*.png.go

$(BD)/%.ttf.go: $(D)/%.ttf
	go-bindata -f `python2 -c 'print "$*".title()'`Ttf \
		-i $(D)/$*.ttf -o $(BD)/$*.ttf.go -p bindata
	gofmt -w $(BD)/$*.ttf.go

loc:
	find ./ -name '*.go' -and -not -wholename './tests*' -and -not -wholename './bindata*' -print | sort | xargs wc -l

tags:
	find ./ \( -name '*.go' \
					   -and -not -wholename './tests/*' \
						 -and -not -wholename './bindata/*' \) \
			 -print0 \
	| xargs -0 gotags > TAGS

